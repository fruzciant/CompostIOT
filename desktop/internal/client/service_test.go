package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/compostiot/desktop/internal/cache"
	"github.com/compostiot/desktop/internal/client"
	"github.com/compostiot/desktop/internal/models"
	"github.com/compostiot/desktop/internal/secrets"
)

func TestMultiServerKeyringAndOfflineSQLiteCache(t *testing.T) {
	const expectedToken = "jwt-bearer-token-demo-123"
	var gotAuthHeader string

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/v1/health", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(models.HealthInfo{
			Status: "ok", Service: "compostiot-backend", Version: "1.0.0", ServerName: "SENA Lab",
		})
	})
	mux.HandleFunc("POST /api/v1/auth/login", func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"access_token": expectedToken,
			"user": models.User{
				ID: 1, Username: "instructor", FullName: "Instructor SENA", Role: "admin", Active: true,
			},
		})
	})
	mux.HandleFunc("GET /api/v1/auth/me", func(w http.ResponseWriter, r *http.Request) {
		gotAuthHeader = r.Header.Get("Authorization")
		if gotAuthHeader != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(models.User{
			ID: 1, Username: "instructor", FullName: "Instructor SENA", Role: "admin", Active: true,
		})
	})
	mux.HandleFunc("GET /api/v1/sync/snapshot", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+expectedToken {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"server_time": "2026-11-26T15:00:00Z",
			"piles": []models.Pile{
				{
					ID: 1, Name: "Pila #1 SENA", NodeID: "NODO-01", Status: "active",
					Evaluation: models.Evaluation{
						Phase: models.PhaseInfo{Code: "termofila", Label: "Termófila"},
					},
				},
			},
			"readings": []models.Reading{
				{ID: 10, PileID: 1, TemperatureC: 58.2, HumidityPct: 54.0, Phase: "termofila", RecordedAt: "2026-11-26T14:55:00Z"},
				{ID: 11, PileID: 1, TemperatureC: 59.1, HumidityPct: 53.5, Phase: "termofila", RecordedAt: "2026-11-26T15:00:00Z"},
			},
		})
	})

	ts := httptest.NewServer(mux)

	store, err := cache.Open(":memory:")
	if err != nil {
		t.Fatalf("cache.Open: %v", err)
	}
	defer store.Close()

	tokenStore := secrets.NewMemoryStore()
	svc := client.NewService(store, tokenStore)

	// 1. Agregar un segundo servidor (multi-servidor)
	srv, err := svc.SaveServer(models.ServerProfile{
		Name: "Servidor Nube SENA",
		URL:  ts.URL,
	})
	if err != nil {
		t.Fatalf("SaveServer: %v", err)
	}
	servers, _ := svc.GetServers()
	if len(servers) != 2 {
		t.Fatalf("esperaba 2 servidores (default + nuevo), obtuvo %d", len(servers))
	}

	// 2. Probar health check
	health, err := svc.CheckHealth(srv.URL)
	if err != nil || health.ServerName != "SENA Lab" {
		t.Fatalf("CheckHealth falló: %+v err=%v", health, err)
	}

	// 3. Iniciar sesión -> guarda Bearer en llavero y sincroniza SQLite
	loginRes, err := svc.Login(srv.ID, "instructor", "compost2026")
	if err != nil || loginRes.User.Username != "instructor" {
		t.Fatalf("Login falló: %v", err)
	}
	storedTok, err := tokenStore.GetToken(srv.ID)
	if err != nil || storedTok != expectedToken {
		t.Fatalf("token en llavero inesperado: %q err=%v", storedTok, err)
	}

	// 4. Restaurar sesión online envía cabecera Bearer
	restoredOnline, err := svc.RestoreSession(srv.ID)
	if err != nil || restoredOnline.Offline || gotAuthHeader != "Bearer "+expectedToken {
		t.Fatalf("RestoreSession online falló: res=%+v auth=%q err=%v", restoredOnline, gotAuthHeader, err)
	}

	// 5. Apagar el servidor para simular campo SIN INTERNET -> verificar caché SQLite offline!
	ts.Close()

	restoredOffline, err := svc.RestoreSession(srv.ID)
	if err != nil || !restoredOffline.Offline || restoredOffline.User.Username != "instructor" {
		t.Fatalf("RestoreSession offline falló: res=%+v err=%v", restoredOffline, err)
	}

	pilesOffline, err := svc.GetPiles(srv.ID, "all")
	if err != nil || !pilesOffline.Offline || len(pilesOffline.Piles) != 1 {
		t.Fatalf("GetPiles offline falló: res=%+v err=%v", pilesOffline, err)
	}

	readingsOffline, err := svc.GetReadings(srv.ID, 1, 50)
	if err != nil || !readingsOffline.Offline || len(readingsOffline.Readings) != 2 {
		t.Fatalf("GetReadings offline falló: res=%+v err=%v", readingsOffline, err)
	}

	csvOffline, err := svc.ExportCSV(srv.ID, 1)
	if err != nil || !strings.Contains(csvOffline, "59.1") {
		t.Fatalf("ExportCSV offline falló: err=%v csv=%q", err, csvOffline)
	}
}
