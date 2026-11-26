package api_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/compostiot/backend/internal/api"
	"github.com/compostiot/backend/internal/config"
	"github.com/compostiot/backend/internal/db"
)

func setupTestServer(t *testing.T) *httptest.Server {
	t.Helper()
	store, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("error abriendo sqlite en memoria: %v", err)
	}
	if err := store.SeedDefaults(false); err != nil {
		t.Fatalf("error sembrando defaults: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })

	cfg := config.Config{
		Port:       "0",
		JWTSecret:  "test-secret-key",
		ServerName: "CompostIOT Test",
		TokenTTL:   time.Hour,
	}
	srv := api.NewServer(cfg, store)
	ts := httptest.NewServer(srv.Handler())
	t.Cleanup(ts.Close)
	return ts
}

func doJSON(t *testing.T, method, url, bearer, deviceKey string, body any) *http.Response {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req, err := http.NewRequest(method, url, &buf)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if bearer != "" {
		req.Header.Set("Authorization", "Bearer "+bearer)
	}
	if deviceKey != "" {
		req.Header.Set("X-Device-Key", deviceKey)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("http do %s %s: %v", method, url, err)
	}
	return resp
}

func login(t *testing.T, baseURL, username, password string) string {
	t.Helper()
	resp := doJSON(t, http.MethodPost, baseURL+"/api/v1/auth/login", "", "", map[string]string{
		"username": username,
		"password": password,
	})
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s falló con status %d", username, resp.StatusCode)
	}
	var out struct {
		AccessToken string `json:"access_token"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || out.AccessToken == "" {
		t.Fatalf("token inválido: %v", err)
	}
	return out.AccessToken
}

func TestFullWorkflowEndToEnd(t *testing.T) {
	ts := setupTestServer(t)

	// 1. Health check público
	resp := doJSON(t, http.MethodGet, ts.URL+"/api/v1/health", "", "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("health status = %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 2. Login admin y aprendiz
	adminToken := login(t, ts.URL, "instructor", "compost2026")
	aprendizToken := login(t, ts.URL, "aprendiz", "aprendiz2026")

	// 3. Aprendiz NO puede crear pilas (403)
	resp = doJSON(t, http.MethodPost, ts.URL+"/api/v1/piles", aprendizToken, "", map[string]any{
		"name":    "Pila No Permitida",
		"node_id": "NODO-99",
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("aprendiz debía recibir 403 al crear pila, obtuvo %d", resp.StatusCode)
	}
	resp.Body.Close()

	// 4. Instructor crea una pila
	resp = doJSON(t, http.MethodPost, ts.URL+"/api/v1/piles", adminToken, "", map[string]any{
		"name":          "Pila SENA #1",
		"location":      "Lote Experimental",
		"substrate":     "Residuos vegetales + aserrín",
		"node_id":       "NODO-01",
		"has_ph_sensor": false,
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("crear pila falló con %d", resp.StatusCode)
	}
	var pile api.PileDTO
	_ = json.NewDecoder(resp.Body).Decode(&pile)
	resp.Body.Close()
	if pile.Evaluation.Phase.Code != "sin_datos" {
		t.Fatalf("pila nueva debía estar en sin_datos, obtuvo %s", pile.Evaluation.Phase.Code)
	}

	// 5. Instructor crea una clave de Gateway
	resp = doJSON(t, http.MethodPost, ts.URL+"/api/v1/devices", adminToken, "", map[string]string{
		"name": "TTGO Gateway Aula",
	})
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("crear gateway falló con %d", resp.StatusCode)
	}
	var devOut struct {
		APIKey string `json:"api_key"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&devOut)
	resp.Body.Close()
	if !strings.HasPrefix(devOut.APIKey, "ciot_") {
		t.Fatalf("api_key inesperada: %q", devOut.APIKey)
	}

	// 6. Gateway envía secuencia de lecturas hasta alcanzar fase termófila
	temps := []float64{30.5, 48.0, 56.4}
	var lastPhase string
	for i, temp := range temps {
		resp = doJSON(t, http.MethodPost, ts.URL+"/api/v1/ingest/readings", "", devOut.APIKey, map[string]any{
			"node_id":       "NODO-01",
			"temperature_c": temp,
			"humidity_pct":  54.0,
			"ph":            nil,
			"power_source":  "mains",
			"rssi":          -76,
			"seq":           i + 1,
		})
		if resp.StatusCode != http.StatusCreated {
			t.Fatalf("ingest lectura %d falló con %d", i, resp.StatusCode)
		}
		var ingOut struct {
			Evaluation struct {
				Phase struct {
					Code string `json:"code"`
				} `json:"phase"`
			} `json:"evaluation"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&ingOut)
		resp.Body.Close()
		lastPhase = ingOut.Evaluation.Phase.Code
	}
	if lastPhase != "termofila" {
		t.Fatalf("esperaba fase termofila tras lecturas 30.5, 48.0, 56.4; obtuvo %s", lastPhase)
	}

	// 7. Aprendiz consulta pilas, histórico, CSV y sync snapshot
	resp = doJSON(t, http.MethodGet, ts.URL+"/api/v1/piles", aprendizToken, "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("listar pilas aprendiz falló con %d", resp.StatusCode)
	}
	var piles []api.PileDTO
	_ = json.NewDecoder(resp.Body).Decode(&piles)
	resp.Body.Close()
	if len(piles) != 1 || piles[0].LatestReading == nil || piles[0].LatestReading.TemperatureC != 56.4 {
		t.Fatalf("datos de pila inesperados: %+v", piles)
	}

	// Exportación CSV
	resp = doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/piles/%d/export?format=csv", ts.URL, pile.ID), aprendizToken, "", nil)
	if resp.StatusCode != http.StatusOK || !strings.Contains(resp.Header.Get("Content-Type"), "text/csv") {
		t.Fatalf("export CSV falló: status=%d ct=%s", resp.StatusCode, resp.Header.Get("Content-Type"))
	}
	resp.Body.Close()

	// Sync snapshot para Wails/Móvil
	resp = doJSON(t, http.MethodGet, ts.URL+"/api/v1/sync/snapshot", aprendizToken, "", nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("sync snapshot falló: %d", resp.StatusCode)
	}
	var snap struct {
		Piles    []api.PileDTO    `json:"piles"`
		Readings []api.ReadingDTO `json:"readings"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&snap)
	resp.Body.Close()
	if len(snap.Piles) != 1 || len(snap.Readings) != 3 {
		t.Fatalf("snapshot inesperado: piles=%d readings=%d", len(snap.Piles), len(snap.Readings))
	}
}
