package client

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/compostiot/desktop/internal/cache"
	"github.com/compostiot/desktop/internal/models"
	"github.com/compostiot/desktop/internal/secrets"
)

type APIError struct {
	Status  int
	Code    string
	Message string
}

func (e *APIError) Error() string {
	return e.Message
}

type Service struct {
	cache   *cache.Store
	secrets secrets.TokenStore
	http    *http.Client
}

func NewService(cacheStore *cache.Store, tokenStore secrets.TokenStore) *Service {
	return &Service{
		cache:   cacheStore,
		secrets: tokenStore,
		http: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func normalizeURL(raw string) string {
	trimmed := strings.TrimRight(strings.TrimSpace(raw), "/")
	if trimmed == "" {
		return "http://localhost:8080"
	}
	if !strings.HasPrefix(trimmed, "http://") && !strings.HasPrefix(trimmed, "https://") {
		return "http://" + trimmed
	}
	return trimmed
}

// --- Servidores ---

func (s *Service) GetServers() ([]models.ServerProfile, error) {
	return s.cache.ListServers()
}

func (s *Service) SaveServer(profile models.ServerProfile) (models.ServerProfile, error) {
	profile.Name = strings.TrimSpace(profile.Name)
	if profile.Name == "" {
		profile.Name = "Servidor CompostIOT"
	}
	profile.URL = normalizeURL(profile.URL)
	return s.cache.SaveServer(profile)
}

func (s *Service) DeleteServer(id string) error {
	_ = s.secrets.DeleteToken(id)
	return s.cache.DeleteServer(id)
}

func (s *Service) CheckHealth(serverURL string) (models.HealthInfo, error) {
	var out models.HealthInfo
	err := s.doJSON("", normalizeURL(serverURL), http.MethodGet, "/api/v1/health", nil, &out, false)
	return out, err
}

// --- Sesión y Keyring ---

func (s *Service) Login(serverID, username, password string) (models.LoginResult, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.LoginResult{}, fmt.Errorf("servidor no encontrado: %w", err)
	}

	var resp struct {
		AccessToken string      `json:"access_token"`
		User        models.User `json:"user"`
	}
	err = s.doJSON(serverID, srv.URL, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": username,
		"password": password,
	}, &resp, false)
	if err != nil {
		return models.LoginResult{}, err
	}

	// Guardar token Bearer de forma segura en el llavero del sistema operativo
	if err := s.secrets.SetToken(serverID, resp.AccessToken); err != nil {
		return models.LoginResult{}, fmt.Errorf("no se pudo guardar el token en el llavero del sistema: %w", err)
	}
	_ = s.cache.SaveCachedUser(serverID, resp.User)

	// Sincronización inicial al caché SQLite
	now := time.Now().UTC().Format(time.RFC3339)
	if syncRes, syncErr := s.syncSnapshot(srv); syncErr == nil && syncRes.LastSyncAt != nil {
		now = *syncRes.LastSyncAt
	}
	_ = s.cache.UpdateServerSync(serverID, resp.User.Username, now)

	return models.LoginResult{
		User:       resp.User,
		LastSyncAt: now,
	}, nil
}

func (s *Service) RestoreSession(serverID string) (models.SessionRestoreResult, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.SessionRestoreResult{}, err
	}
	if _, err := s.secrets.GetToken(serverID); err != nil {
		return models.SessionRestoreResult{}, errors.New("sin sesión guardada")
	}

	var lastSync *string
	if srv.LastSyncAt != "" {
		v := srv.LastSyncAt
		lastSync = &v
	}

	var user models.User
	err = s.doJSON(serverID, srv.URL, http.MethodGet, "/api/v1/auth/me", nil, &user, true)
	if err == nil {
		_ = s.cache.SaveCachedUser(serverID, user)
		return models.SessionRestoreResult{
			User:       user,
			Offline:    false,
			LastSyncAt: lastSync,
		}, nil
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		_ = s.secrets.DeleteToken(serverID)
		return models.SessionRestoreResult{}, err
	}

	// Fallo de red: restaurar sesión desde SQLite offline
	cachedUser, cacheErr := s.cache.GetCachedUser(serverID)
	if cacheErr != nil {
		return models.SessionRestoreResult{}, err
	}
	return models.SessionRestoreResult{
		User:       *cachedUser,
		Offline:    true,
		LastSyncAt: lastSync,
	}, nil
}

func (s *Service) Logout(serverID string) error {
	_ = s.secrets.DeleteToken(serverID)
	_ = s.cache.ClearCachedUser(serverID)
	return nil
}

// --- Sincronización y Caché SQLite Offline ---

func (s *Service) SyncServer(serverID string) (models.PilesResponse, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.PilesResponse{}, err
	}
	return s.syncSnapshot(srv)
}

func (s *Service) syncSnapshot(srv models.ServerProfile) (models.PilesResponse, error) {
	var snap struct {
		ServerTime string           `json:"server_time"`
		Piles      []models.Pile    `json:"piles"`
		Readings   []models.Reading `json:"readings"`
	}
	err := s.doJSON(srv.ID, srv.URL, http.MethodGet, "/api/v1/sync/snapshot", nil, &snap, true)
	if err != nil {
		var apiErr *APIError
		if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
			_ = s.secrets.DeleteToken(srv.ID)
			return models.PilesResponse{}, err
		}
		// Fallback offline desde SQLite
		piles, cErr := s.cache.ListCachedPiles(srv.ID, "all")
		if cErr != nil {
			return models.PilesResponse{}, err
		}
		var lastSync *string
		if srv.LastSyncAt != "" {
			v := srv.LastSyncAt
			lastSync = &v
		}
		return models.PilesResponse{Piles: piles, Offline: true, LastSyncAt: lastSync}, nil
	}

	_ = s.cache.ReplaceCachedPiles(srv.ID, snap.Piles)
	_ = s.cache.UpsertCachedReadings(srv.ID, snap.Readings)
	syncTime := snap.ServerTime
	if syncTime == "" {
		syncTime = time.Now().UTC().Format(time.RFC3339)
	}
	_ = s.cache.UpdateServerSync(srv.ID, "", syncTime)

	return models.PilesResponse{
		Piles:      snap.Piles,
		Offline:    false,
		LastSyncAt: &syncTime,
	}, nil
}

func (s *Service) GetPiles(serverID, status string) (models.PilesResponse, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.PilesResponse{}, err
	}
	if status == "" {
		status = "all"
	}

	// Si pide 'all', aprovechamos el snapshot completo para mantener el caché SQLite de lecturas al día
	if status == "all" {
		return s.syncSnapshot(srv)
	}

	var piles []models.Pile
	err = s.doJSON(serverID, srv.URL, http.MethodGet, "/api/v1/piles?status="+status, nil, &piles, true)
	if err == nil {
		now := time.Now().UTC().Format(time.RFC3339)
		_ = s.cache.UpdateServerSync(serverID, "", now)
		return models.PilesResponse{Piles: piles, Offline: false, LastSyncAt: &now}, nil
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		return models.PilesResponse{}, err
	}

	cached, cErr := s.cache.ListCachedPiles(serverID, status)
	if cErr != nil {
		return models.PilesResponse{}, err
	}
	var lastSync *string
	if srv.LastSyncAt != "" {
		v := srv.LastSyncAt
		lastSync = &v
	}
	return models.PilesResponse{Piles: cached, Offline: true, LastSyncAt: lastSync}, nil
}

func (s *Service) GetReadings(serverID string, pileID int64, limit int) (models.ReadingsResponse, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.ReadingsResponse{}, err
	}
	if limit <= 0 {
		limit = 300
	}

	var readings []models.Reading
	path := fmt.Sprintf("/api/v1/piles/%d/readings?limit=%d", pileID, limit)
	err = s.doJSON(serverID, srv.URL, http.MethodGet, path, nil, &readings, true)
	if err == nil {
		_ = s.cache.UpsertCachedReadings(serverID, readings)
		return models.ReadingsResponse{Readings: readings, Offline: false}, nil
	}

	var apiErr *APIError
	if errors.As(err, &apiErr) && apiErr.Status == http.StatusUnauthorized {
		return models.ReadingsResponse{}, err
	}

	cached, cErr := s.cache.ListCachedReadings(serverID, pileID, limit)
	if cErr != nil {
		return models.ReadingsResponse{}, err
	}
	return models.ReadingsResponse{Readings: cached, Offline: true}, nil
}

// --- Mutaciones (requieren conexión y rol admin) ---

func (s *Service) CreatePile(serverID string, input models.PileInput) (models.Pile, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.Pile{}, err
	}
	var out models.Pile
	err = s.doJSON(serverID, srv.URL, http.MethodPost, "/api/v1/piles", input, &out, true)
	return out, err
}

func (s *Service) UpdatePile(serverID string, pileID int64, input models.PileInput) (models.Pile, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.Pile{}, err
	}
	var out models.Pile
	err = s.doJSON(serverID, srv.URL, http.MethodPut, fmt.Sprintf("/api/v1/piles/%d", pileID), input, &out, true)
	return out, err
}

func (s *Service) DeletePile(serverID string, pileID int64) error {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return err
	}
	return s.doJSON(serverID, srv.URL, http.MethodDelete, fmt.Sprintf("/api/v1/piles/%d", pileID), nil, nil, true)
}

func (s *Service) AddManualReading(serverID string, pileID int64, input models.ManualReadingInput) (models.Reading, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.Reading{}, err
	}
	var resp struct {
		Reading models.Reading `json:"reading"`
	}
	err = s.doJSON(serverID, srv.URL, http.MethodPost, fmt.Sprintf("/api/v1/piles/%d/readings", pileID), input, &resp, true)
	if err == nil {
		_ = s.cache.UpsertCachedReadings(serverID, []models.Reading{resp.Reading})
	}
	return resp.Reading, err
}

func (s *Service) ExportCSV(serverID string, pileID int64) (string, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return "", err
	}

	csvText, err := s.doRawText(serverID, srv.URL, fmt.Sprintf("/api/v1/piles/%d/export?format=csv", pileID))
	if err == nil {
		return csvText, nil
	}

	// Fallback offline: exportar CSV desde caché SQLite
	readings, cErr := s.cache.ListCachedReadings(serverID, pileID, 5000)
	if cErr != nil || len(readings) == 0 {
		return "", err
	}
	var buf bytes.Buffer
	buf.WriteString("\xEF\xBB\xBF")
	cw := csv.NewWriter(&buf)
	_ = cw.Write([]string{"fecha_hora", "temperatura_c", "humedad_pct", "ph", "fase", "origen"})
	for _, r := range readings {
		phStr := ""
		if r.PH != nil {
			phStr = fmt.Sprintf("%.2f", *r.PH)
		}
		_ = cw.Write([]string{
			r.RecordedAt,
			fmt.Sprintf("%.1f", r.TemperatureC),
			fmt.Sprintf("%.1f", r.HumidityPct),
			phStr,
			r.Phase,
			"cache_offline",
		})
	}
	cw.Flush()
	return buf.String(), nil
}

// --- Administración ---

func (s *Service) GetUsers(serverID string) ([]models.User, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return nil, err
	}
	var out []models.User
	err = s.doJSON(serverID, srv.URL, http.MethodGet, "/api/v1/users", nil, &out, true)
	return out, err
}

func (s *Service) CreateUser(serverID string, payload models.CreateUserInput) (models.User, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.User{}, err
	}
	var out models.User
	err = s.doJSON(serverID, srv.URL, http.MethodPost, "/api/v1/users", payload, &out, true)
	return out, err
}

func (s *Service) DeleteUser(serverID string, userID int64) error {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return err
	}
	return s.doJSON(serverID, srv.URL, http.MethodDelete, fmt.Sprintf("/api/v1/users/%d", userID), nil, nil, true)
}

func (s *Service) GetDevices(serverID string) ([]models.Device, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return nil, err
	}
	var out []models.Device
	err = s.doJSON(serverID, srv.URL, http.MethodGet, "/api/v1/devices", nil, &out, true)
	return out, err
}

func (s *Service) CreateDevice(serverID, name string) (models.CreateDeviceResult, error) {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return models.CreateDeviceResult{}, err
	}
	var out models.CreateDeviceResult
	err = s.doJSON(serverID, srv.URL, http.MethodPost, "/api/v1/devices", map[string]string{"name": name}, &out, true)
	return out, err
}

func (s *Service) DeleteDevice(serverID string, deviceID int64) error {
	srv, err := s.cache.GetServer(serverID)
	if err != nil {
		return err
	}
	return s.doJSON(serverID, srv.URL, http.MethodDelete, fmt.Sprintf("/api/v1/devices/%d", deviceID), nil, nil, true)
}

// --- HTTP helpers ---

func (s *Service) buildRequest(serverID, baseURL, method, path string, body any, withAuth bool) (*http.Request, error) {
	var reader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, normalizeURL(baseURL)+path, reader)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if withAuth {
		token, err := s.secrets.GetToken(serverID)
		if err != nil {
			return nil, &APIError{Status: http.StatusUnauthorized, Code: "NO_TOKEN", Message: "Sesión no iniciada"}
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func (s *Service) doJSON(serverID, baseURL, method, path string, body any, dst any, withAuth bool) error {
	req, err := s.buildRequest(serverID, baseURL, method, path, body, withAuth)
	if err != nil {
		return err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode >= 400 {
		return parseAPIError(resp)
	}
	if dst != nil {
		return json.NewDecoder(resp.Body).Decode(dst)
	}
	return nil
}

func (s *Service) doRawText(serverID, baseURL, path string) (string, error) {
	req, err := s.buildRequest(serverID, baseURL, http.MethodGet, path, nil, true)
	if err != nil {
		return "", err
	}
	resp, err := s.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return "", parseAPIError(resp)
	}
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

func parseAPIError(resp *http.Response) error {
	var env struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&env)
	msg := env.Error.Message
	if msg == "" {
		msg = fmt.Sprintf("Error HTTP %d del servidor", resp.StatusCode)
	}
	return &APIError{
		Status:  resp.StatusCode,
		Code:    env.Error.Code,
		Message: msg,
	}
}
