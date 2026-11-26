package cache

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/compostiot/desktop/internal/models"
	_ "modernc.org/sqlite"
)

type Store struct {
	db *sql.DB
}

func DefaultDBPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "./compostiot-desktop-cache.db", nil
	}
	dir := filepath.Join(cfgDir, "CompostIOT")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	return filepath.Join(dir, "cache.db"), nil
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
	conn.SetMaxOpenConns(1)
	if path == ":memory:" {
		_, _ = conn.Exec("PRAGMA foreign_keys = ON;")
	}

	s := &Store{db: conn}
	if err := s.migrate(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := s.ensureDefaultServer(); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS servers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	url TEXT NOT NULL,
	username TEXT NOT NULL DEFAULT '',
	last_sync_at TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS cached_users (
	server_id TEXT PRIMARY KEY,
	user_json TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cached_piles (
	server_id TEXT NOT NULL,
	pile_id INTEGER NOT NULL,
	status TEXT NOT NULL,
	pile_json TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	PRIMARY KEY (server_id, pile_id),
	FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS cached_readings (
	server_id TEXT NOT NULL,
	pile_id INTEGER NOT NULL,
	reading_id INTEGER NOT NULL,
	recorded_at TEXT NOT NULL,
	reading_json TEXT NOT NULL,
	PRIMARY KEY (server_id, reading_id),
	FOREIGN KEY (server_id) REFERENCES servers(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_cached_readings_lookup
	ON cached_readings(server_id, pile_id, recorded_at);
`
	_, err := s.db.Exec(schema)
	return err
}

func (s *Store) ensureDefaultServer() error {
	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM servers`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		now := time.Now().UTC().Format(time.RFC3339)
		_, err := s.db.Exec(`INSERT INTO servers (id, name, url, username, last_sync_at, created_at)
			VALUES ('local-default', 'Servidor local (Localhost)', 'http://localhost:8080', '', '', ?)`, now)
		return err
	}
	return nil
}

// --- Servers ---

func (s *Store) ListServers() ([]models.ServerProfile, error) {
	rows, err := s.db.Query(`SELECT id, name, url, username, last_sync_at FROM servers ORDER BY created_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := []models.ServerProfile{}
	for rows.Next() {
		var p models.ServerProfile
		if err := rows.Scan(&p.ID, &p.Name, &p.URL, &p.Username, &p.LastSyncAt); err != nil {
			return nil, err
		}
		list = append(list, p)
	}
	return list, nil
}

func (s *Store) GetServer(id string) (models.ServerProfile, error) {
	var p models.ServerProfile
	err := s.db.QueryRow(`SELECT id, name, url, username, last_sync_at FROM servers WHERE id = ?`, id).
		Scan(&p.ID, &p.Name, &p.URL, &p.Username, &p.LastSyncAt)
	return p, err
}

func (s *Store) SaveServer(p models.ServerProfile) (models.ServerProfile, error) {
	if p.ID == "" {
		p.ID = fmt.Sprintf("srv_%d", time.Now().UnixNano())
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := s.db.Exec(`INSERT INTO servers (id, name, url, username, last_sync_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET name = excluded.name, url = excluded.url,
			username = CASE WHEN excluded.username != '' THEN excluded.username ELSE servers.username END,
			last_sync_at = CASE WHEN excluded.last_sync_at != '' THEN excluded.last_sync_at ELSE servers.last_sync_at END`,
		p.ID, p.Name, p.URL, p.Username, p.LastSyncAt, now)
	return p, err
}

func (s *Store) UpdateServerSync(serverID, username, lastSyncAt string) error {
	_, err := s.db.Exec(`UPDATE servers
		SET username = CASE WHEN ? != '' THEN ? ELSE username END,
			last_sync_at = CASE WHEN ? != '' THEN ? ELSE last_sync_at END
		WHERE id = ?`, username, username, lastSyncAt, lastSyncAt, serverID)
	return err
}

func (s *Store) DeleteServer(id string) error {
	if _, err := s.db.Exec(`DELETE FROM servers WHERE id = ?`, id); err != nil {
		return err
	}
	return s.ensureDefaultServer()
}

// --- Cached User ---

func (s *Store) SaveCachedUser(serverID string, user models.User) error {
	b, err := json.Marshal(user)
	if err != nil {
		return err
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err = s.db.Exec(`INSERT INTO cached_users (server_id, user_json, updated_at)
		VALUES (?, ?, ?)
		ON CONFLICT(server_id) DO UPDATE SET user_json = excluded.user_json, updated_at = excluded.updated_at`,
		serverID, string(b), now)
	return err
}

func (s *Store) GetCachedUser(serverID string) (*models.User, error) {
	var raw string
	err := s.db.QueryRow(`SELECT user_json FROM cached_users WHERE server_id = ?`, serverID).Scan(&raw)
	if err != nil {
		return nil, err
	}
	var u models.User
	if err := json.Unmarshal([]byte(raw), &u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (s *Store) ClearCachedUser(serverID string) error {
	_, err := s.db.Exec(`DELETE FROM cached_users WHERE server_id = ?`, serverID)
	return err
}

// --- Cached Piles ---

func (s *Store) ReplaceCachedPiles(serverID string, piles []models.Pile) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`DELETE FROM cached_piles WHERE server_id = ?`, serverID); err != nil {
		return err
	}
	stmt, err := tx.Prepare(`INSERT INTO cached_piles (server_id, pile_id, status, pile_json, updated_at) VALUES (?, ?, ?, ?, ?)`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, p := range piles {
		b, err := json.Marshal(p)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(serverID, p.ID, p.Status, string(b), p.UpdatedAt); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListCachedPiles(serverID, status string) ([]models.Pile, error) {
	query := `SELECT pile_json FROM cached_piles WHERE server_id = ?`
	args := []any{serverID}
	if status == "active" || status == "archived" {
		query += ` AND status = ?`
		args = append(args, status)
	}
	query += ` ORDER BY pile_id ASC`

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	piles := []models.Pile{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var p models.Pile
		if err := json.Unmarshal([]byte(raw), &p); err == nil {
			piles = append(piles, p)
		}
	}
	return piles, nil
}

// --- Cached Readings ---

func (s *Store) UpsertCachedReadings(serverID string, readings []models.Reading) error {
	if len(readings) == 0 {
		return nil
	}
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	stmt, err := tx.Prepare(`INSERT INTO cached_readings (server_id, pile_id, reading_id, recorded_at, reading_json)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(server_id, reading_id) DO UPDATE SET reading_json = excluded.reading_json, recorded_at = excluded.recorded_at`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, r := range readings {
		b, err := json.Marshal(r)
		if err != nil {
			return err
		}
		if _, err := stmt.Exec(serverID, r.PileID, r.ID, r.RecordedAt, string(b)); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Store) ListCachedReadings(serverID string, pileID int64, limit int) ([]models.Reading, error) {
	if limit <= 0 {
		limit = 300
	}
	rows, err := s.db.Query(`SELECT reading_json FROM (
		SELECT reading_json, recorded_at, reading_id
		FROM cached_readings
		WHERE server_id = ? AND pile_id = ?
		ORDER BY recorded_at DESC, reading_id DESC
		LIMIT ?
	) ORDER BY recorded_at ASC, reading_id ASC`, serverID, pileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	readings := []models.Reading{}
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, err
		}
		var r models.Reading
		if err := json.Unmarshal([]byte(raw), &r); err == nil {
			readings = append(readings, r)
		}
	}
	return readings, nil
}
