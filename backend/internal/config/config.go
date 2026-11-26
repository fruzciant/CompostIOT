package config

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port       string
	DBPath     string
	JWTSecret  string
	ServerName string
	WebDir     string
	TokenTTL   time.Duration
	SeedDemo   bool
}

func Load() Config {
	cfg := Config{
		Port:       envOrDefault("COMPOST_PORT", "8080"),
		DBPath:     envOrDefault("COMPOST_DB_PATH", "./compostiot.db"),
		JWTSecret:  os.Getenv("COMPOST_JWT_SECRET"),
		ServerName: envOrDefault("COMPOST_SERVER_NAME", "CompostIOT Localhost (SENA)"),
		WebDir:     envOrDefault("COMPOST_WEB_DIR", autoDetectWebDir()),
		TokenTTL:   24 * time.Hour,
		SeedDemo:   envBool("COMPOST_SEED_DEMO", false),
	}

	if ttlHours := os.Getenv("COMPOST_TOKEN_TTL_HOURS"); ttlHours != "" {
		if h, err := strconv.Atoi(ttlHours); err == nil && h > 0 {
			cfg.TokenTTL = time.Duration(h) * time.Hour
		}
	}

	flag.StringVar(&cfg.Port, "port", cfg.Port, "Puerto HTTP del servidor")
	flag.StringVar(&cfg.DBPath, "db", cfg.DBPath, "Ruta del archivo SQLite")
	flag.StringVar(&cfg.ServerName, "server-name", cfg.ServerName, "Nombre visible del servidor")
	flag.StringVar(&cfg.WebDir, "web-dir", cfg.WebDir, "Carpeta estática del frontend web (web/dist) para servir todo en un solo puerto")
	flag.BoolVar(&cfg.SeedDemo, "seed-demo", cfg.SeedDemo, "Cargar pilas y lecturas de demostración si la BD está vacía")
	flag.Parse()

	if cfg.JWTSecret == "" {
		// Clave persistente por defecto para entorno local si el usuario no definió COMPOST_JWT_SECRET.
		// En producción/nube se debe configurar COMPOST_JWT_SECRET explícitamente.
		cfg.JWTSecret = envOrDefault("COMPOST_FALLBACK_SECRET", "compostiot-local-dev-secret-change-in-prod")
	}

	return cfg
}

func RandomSecret() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

func autoDetectWebDir() string {
	for _, candidate := range []string{"./web-dist", "../web/dist"} {
		if info, err := os.Stat(candidate + "/index.html"); err == nil && !info.IsDir() {
			return candidate
		}
	}
	return ""
}
