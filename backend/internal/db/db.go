package db

import (
	"database/sql"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/compostiot/backend/internal/auth"
	"github.com/compostiot/backend/internal/compost"
	_ "modernc.org/sqlite"
)

type Store struct {
	DB *sql.DB
}

func Open(path string) (*Store, error) {
	dsn := path
	if path != ":memory:" {
		dsn = fmt.Sprintf("file:%s?_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)", path)
	}
	conn, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, err
	}
	conn.SetMaxOpenConns(1) // SQLite serializa escrituras limpiamente

	if path == ":memory:" {
		if _, err := conn.Exec("PRAGMA foreign_keys = ON;"); err != nil {
			_ = conn.Close()
			return nil, err
		}
	}

	s := &Store{DB: conn}
	if err := s.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.DB.Close()
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	username TEXT NOT NULL UNIQUE,
	full_name TEXT NOT NULL,
	password_hash TEXT NOT NULL,
	role TEXT NOT NULL CHECK (role IN ('admin', 'aprendiz')),
	active INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS piles (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	location TEXT NOT NULL DEFAULT '',
	substrate TEXT NOT NULL DEFAULT '',
	node_id TEXT NOT NULL,
	has_ph_sensor INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived')),
	notes TEXT NOT NULL DEFAULT '',
	started_at TEXT NOT NULL,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_piles_active_node
	ON piles(node_id) WHERE status = 'active' AND node_id != '';

CREATE TABLE IF NOT EXISTS readings (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	pile_id INTEGER NOT NULL,
	temperature_c REAL NOT NULL,
	humidity_pct REAL NOT NULL,
	ph REAL,
	phase TEXT NOT NULL,
	power_source TEXT NOT NULL DEFAULT 'mains',
	battery_v REAL,
	rssi INTEGER,
	seq INTEGER,
	recorded_at TEXT NOT NULL,
	FOREIGN KEY (pile_id) REFERENCES piles(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_readings_pile_time
	ON readings(pile_id, recorded_at);

CREATE TABLE IF NOT EXISTS devices (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	key_prefix TEXT NOT NULL,
	key_hash TEXT NOT NULL UNIQUE,
	active INTEGER NOT NULL DEFAULT 1,
	last_seen_at TEXT,
	created_at TEXT NOT NULL
);
`
	_, err := s.DB.Exec(schema)
	return err
}

// SeedDefaults crea usuarios iniciales si la BD está vacía y, si seedDemo es true, carga pilas y curvas realistas.
func (s *Store) SeedDefaults(seedDemo bool) error {
	var userCount int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&userCount); err != nil {
		return err
	}
	nowStr := time.Now().UTC().Format(time.RFC3339)

	if userCount == 0 {
		adminHash, err := auth.HashPassword("compost2026")
		if err != nil {
			return err
		}
		aprendizHash, err := auth.HashPassword("aprendiz2026")
		if err != nil {
			return err
		}
		_, err = s.DB.Exec(`INSERT INTO users (username, full_name, password_hash, role, active, created_at) VALUES
			('instructor', 'Instructor Ambiental SENA', ?, 'admin', 1, ?),
			('aprendiz', 'Aprendiz Tecnólogo SENA', ?, 'aprendiz', 1, ?)`,
			adminHash, nowStr, aprendizHash, nowStr)
		if err != nil {
			return err
		}
		log.Println("[SEED] Usuarios creados: instructor / compost2026 (admin) | aprendiz / aprendiz2026 (aprendiz)")
	}

	if !seedDemo {
		return nil
	}

	var devCount int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM devices`).Scan(&devCount); err != nil {
		return err
	}
	if devCount == 0 {
		const demoKey = "ciot_demo_gateway_sena_2026"
		_, err := s.DB.Exec(`INSERT INTO devices (name, key_prefix, key_hash, active, last_seen_at, created_at)
			VALUES ('Gateway TTGO Laboratorio Demo', ?, ?, 1, ?, ?)`,
			demoKey[:12], auth.HashAPIKey(demoKey), nowStr, nowStr)
		if err != nil {
			return err
		}
		log.Printf("[SEED] Gateway demo creado. X-Device-Key: %s", demoKey)
	}

	var pileCount int
	if err := s.DB.QueryRow(`SELECT COUNT(*) FROM piles`).Scan(&pileCount); err != nil {
		return err
	}
	if pileCount == 0 {
		if err := s.seedDemoPiles(); err != nil {
			return err
		}
		log.Println("[SEED] 3 pilas de compostaje de demostración cargadas con histórico.")
	}
	return nil
}

func (s *Store) seedDemoPiles() error {
	now := time.Now().UTC()
	type demoPile struct {
		name      string
		location  string
		substrate string
		nodeID    string
		hasPH     int
		daysAgo   int
		notes     string
		curve     func(step, total int) (temp, hum float64, ph *float64)
	}

	piles := []demoPile{
		{
			name:      "Pila #1 — Residuos Cafetería y Poda",
			location:  "Invernadero Norte — Lote A",
			substrate: "Restos vegetales + hojarasca + aserrín (C/N ~30:1)",
			nodeID:    "NODO-01",
			hasPH:     1,
			daysAgo:   14,
			notes:     "Pila en plena fase termófila activa. Volteo cada 4 días.",
			curve: func(step, total int) (float64, float64, *float64) {
				p := float64(step) / float64(total-1)
				temp := 27.0 + 34.0*math.Sin(p*math.Pi*0.65)
				hum := 59.0 - 5.0*p + 1.5*math.Sin(float64(step)*0.6)
				phVal := math.Round((6.2+1.1*p)*100) / 100
				return math.Round(temp*10) / 10, math.Round(hum*10) / 10, &phVal
			},
		},
		{
			name:      "Pila #2 — Estiércol Bovino y Viruta",
			location:  "Zona Productiva — Lote B",
			substrate: "Estiércol bovino + viruta de madera + pasto picado",
			nodeID:    "NODO-02",
			hasPH:     0,
			daysAgo:   32,
			notes:     "Superó higienización termófila (pico 62 °C). En enfriamiento controlado.",
			curve: func(step, total int) (float64, float64, *float64) {
				p := float64(step) / float64(total-1)
				var temp float64
				if p < 0.45 {
					temp = 26.0 + 37.0*(p/0.45)
				} else {
					temp = 63.0 - 26.0*((p-0.45)/0.55)
				}
				hum := 56.0 - 7.0*p
				return math.Round(temp*10) / 10, math.Round(hum*10) / 10, nil
			},
		},
		{
			name:      "Pila #3 — Lote Experimental Nuevo",
			location:  "Aula Práctica Ambiental",
			substrate: "Residuos de cosecha + cascarilla de café",
			nodeID:    "NODO-03",
			hasPH:     0,
			daysAgo:   2,
			notes:     "Recién armada por grupo de aprendices ficha 2901.",
			curve: func(step, total int) (float64, float64, *float64) {
				p := float64(step) / float64(total-1)
				temp := 24.5 + 14.0*p
				hum := 57.0 - 2.0*p
				return math.Round(temp*10) / 10, math.Round(hum*10) / 10, nil
			},
		},
	}

	for _, dp := range piles {
		started := now.Add(-time.Duration(dp.daysAgo) * 24 * time.Hour)
		res, err := s.DB.Exec(`INSERT INTO piles (name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, 'active', ?, ?, ?, ?)`,
			dp.name, dp.location, dp.substrate, dp.nodeID, dp.hasPH, dp.notes,
			started.Format(time.RFC3339), started.Format(time.RFC3339), now.Format(time.RFC3339))
		if err != nil {
			return err
		}
		pileID, err := res.LastInsertId()
		if err != nil {
			return err
		}

		const steps = 28
		var history []compost.ReadingSample
		for i := 0; i < steps; i++ {
			temp, hum, ph := dp.curve(i, steps)
			// Distribuir lecturas desde started hasta hace 5 minutos
			span := now.Add(-5 * time.Minute).Sub(started)
			ts := started.Add(time.Duration(float64(span) * float64(i) / float64(steps-1)))
			history = append(history, compost.ReadingSample{
				TemperatureC: temp,
				HumidityPct:  hum,
				PH:           ph,
				RecordedAt:   ts,
			})
			ev := compost.Evaluate(started, history, "active", ts)
			_, err := s.DB.Exec(`INSERT INTO readings (pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at)
				VALUES (?, ?, ?, ?, ?, 'mains', NULL, ?, ?, ?)`,
				pileID, temp, hum, ph, string(ev.Phase.Code), -74-(i%8), i+1, ts.Format(time.RFC3339))
			if err != nil {
				return err
			}
		}
	}
	return nil
}
