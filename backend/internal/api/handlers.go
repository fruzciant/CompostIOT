package api

import (
	"database/sql"
	"encoding/csv"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/compostiot/backend/internal/auth"
	"github.com/compostiot/backend/internal/compost"
	"github.com/compostiot/backend/internal/config"
	"github.com/compostiot/backend/internal/db"
)

const ServiceVersion = "1.1.0"

type Server struct {
	cfg    config.Config
	store  *db.Store
	mux    *http.ServeMux
	broker *eventBroker
}

func NewServer(cfg config.Config, store *db.Store) *Server {
	s := &Server{
		cfg:    cfg,
		store:  store,
		mux:    http.NewServeMux(),
		broker: newEventBroker(),
	}
	s.registerRoutes()
	return s
}

func (s *Server) Handler() http.Handler {
	return withLogging(withCORS(s.mux))
}

func (s *Server) registerRoutes() {
	// Salud
	s.mux.HandleFunc("GET /api/v1/health", s.handleHealth)

	// Autenticación
	s.mux.HandleFunc("POST /api/v1/auth/login", s.handleLogin)
	s.mux.HandleFunc("GET /api/v1/auth/me", s.requireAuth(s.handleMe))
	s.mux.HandleFunc("POST /api/v1/auth/change-password", s.requireAuth(s.handleChangePassword))

	// Usuarios (solo admin)
	s.mux.HandleFunc("GET /api/v1/users", s.requireAdmin(s.handleListUsers))
	s.mux.HandleFunc("POST /api/v1/users", s.requireAdmin(s.handleCreateUser))
	s.mux.HandleFunc("PATCH /api/v1/users/{id}", s.requireAdmin(s.handleUpdateUser))
	s.mux.HandleFunc("DELETE /api/v1/users/{id}", s.requireAdmin(s.handleDeleteUser))

	// Pilas
	s.mux.HandleFunc("GET /api/v1/piles", s.requireAuth(s.handleListPiles))
	s.mux.HandleFunc("GET /api/v1/piles/{id}", s.requireAuth(s.handleGetPile))
	s.mux.HandleFunc("POST /api/v1/piles", s.requireAdmin(s.handleCreatePile))
	s.mux.HandleFunc("PUT /api/v1/piles/{id}", s.requireAdmin(s.handleUpdatePile))
	s.mux.HandleFunc("DELETE /api/v1/piles/{id}", s.requireAdmin(s.handleDeletePile))

	// Lecturas y exportación
	s.mux.HandleFunc("GET /api/v1/piles/{id}/readings", s.requireAuth(s.handleListReadings))
	s.mux.HandleFunc("POST /api/v1/piles/{id}/readings", s.requireAdmin(s.handleCreateManualReading))
	s.mux.HandleFunc("GET /api/v1/piles/{id}/export", s.requireAuth(s.handleExportCSV))

	// Ingesta Gateway ESP32
	s.mux.HandleFunc("POST /api/v1/ingest/readings", s.requireDeviceKey(s.handleIngestReading))

	// Gateways / dispositivos (solo admin)
	s.mux.HandleFunc("GET /api/v1/devices", s.requireAdmin(s.handleListDevices))
	s.mux.HandleFunc("POST /api/v1/devices", s.requireAdmin(s.handleCreateDevice))
	s.mux.HandleFunc("DELETE /api/v1/devices/{id}", s.requireAdmin(s.handleDeleteDevice))

	// Sincronización offline (Wails y Móvil) y flujo en tiempo real (SSE)
	s.mux.HandleFunc("GET /api/v1/sync/snapshot", s.requireAuth(s.handleSyncSnapshot))
	s.mux.HandleFunc("GET /api/v1/events", s.handleEvents)

	// Servir interfaz web compilada (modo todo-en-uno para Localhost y Nube)
	if s.cfg.WebDir != "" {
		s.mux.Handle("/", s.spaHandler(s.cfg.WebDir))
	}
}

func (s *Server) spaHandler(webDir string) http.Handler {
	fs := http.FileServer(http.Dir(webDir))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") {
			writeError(w, http.StatusNotFound, "NOT_FOUND", "Endpoint no encontrado")
			return
		}
		clean := filepath.Clean(strings.TrimPrefix(r.URL.Path, "/"))
		target := filepath.Join(webDir, clean)
		if info, err := os.Stat(target); err == nil && !info.IsDir() {
			fs.ServeHTTP(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(webDir, "index.html"))
	})
}

// --- DTOs ---

type UserDTO struct {
	ID        int64  `json:"id"`
	Username  string `json:"username"`
	FullName  string `json:"full_name"`
	Role      string `json:"role"`
	Active    bool   `json:"active"`
	CreatedAt string `json:"created_at"`
}

type ReadingDTO struct {
	ID           int64    `json:"id"`
	PileID       int64    `json:"pile_id"`
	TemperatureC float64  `json:"temperature_c"`
	HumidityPct  float64  `json:"humidity_pct"`
	PH           *float64 `json:"ph"`
	Phase        string   `json:"phase"`
	PowerSource  string   `json:"power_source"`
	BatteryV     *float64 `json:"battery_v"`
	RSSI         *int     `json:"rssi"`
	Seq          *int     `json:"seq"`
	RecordedAt   string   `json:"recorded_at"`
}

type PileDTO struct {
	ID            int64              `json:"id"`
	Name          string             `json:"name"`
	Location      string             `json:"location"`
	Substrate     string             `json:"substrate"`
	NodeID        string             `json:"node_id"`
	HasPHSensor   bool               `json:"has_ph_sensor"`
	Status        string             `json:"status"`
	Notes         string             `json:"notes"`
	StartedAt     string             `json:"started_at"`
	CreatedAt     string             `json:"created_at"`
	UpdatedAt     string             `json:"updated_at"`
	LatestReading *ReadingDTO        `json:"latest_reading"`
	Evaluation    compost.Evaluation `json:"evaluation"`
}

type DeviceDTO struct {
	ID         int64   `json:"id"`
	Name       string  `json:"name"`
	KeyPrefix  string  `json:"key_prefix"`
	Active     bool    `json:"active"`
	LastSeenAt *string `json:"last_seen_at"`
	CreatedAt  string  `json:"created_at"`
}

// --- Health ---

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":      "ok",
		"service":     "compostiot-backend",
		"version":     ServiceVersion,
		"server_name": s.cfg.ServerName,
		"server_time": time.Now().Format(time.RFC3339),
	})
}

// --- Auth ---

func (s *Server) handleLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		writeError(w, http.StatusBadRequest, "MISSING_FIELDS", "Debe ingresar usuario y contraseña")
		return
	}

	var u UserDTO
	var hash string
	var active int
	err := s.store.DB.QueryRow(`SELECT id, username, full_name, password_hash, role, active, created_at
		FROM users WHERE username = ?`, req.Username).
		Scan(&u.ID, &u.Username, &u.FullName, &hash, &u.Role, &active, &u.CreatedAt)
	if err != nil || active != 1 || !auth.CheckPassword(hash, req.Password) {
		writeError(w, http.StatusUnauthorized, "INVALID_CREDENTIALS", "Usuario o contraseña incorrectos")
		return
	}
	u.Active = active == 1

	token, err := auth.GenerateToken(s.cfg.JWTSecret, u.ID, u.Username, u.Role, s.cfg.TokenTTL)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "TOKEN_ERROR", "Error generando token de sesión")
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"access_token": token,
		"token_type":   "Bearer",
		"expires_in":   int(s.cfg.TokenTTL.Seconds()),
		"user":         u,
	})
}

func (s *Server) handleMe(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	var u UserDTO
	var active int
	err := s.store.DB.QueryRow(`SELECT id, username, full_name, role, active, created_at FROM users WHERE id = ?`, claims.UserID).
		Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &active, &u.CreatedAt)
	if err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
		return
	}
	u.Active = active == 1
	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleChangePassword(w http.ResponseWriter, r *http.Request) {
	claims := getClaims(r)
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	if len(req.NewPassword) < 6 {
		writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "La nueva contraseña debe tener al menos 6 caracteres")
		return
	}

	var hash string
	if err := s.store.DB.QueryRow(`SELECT password_hash FROM users WHERE id = ?`, claims.UserID).Scan(&hash); err != nil {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
		return
	}
	if !auth.CheckPassword(hash, req.CurrentPassword) {
		writeError(w, http.StatusUnauthorized, "INVALID_PASSWORD", "La contraseña actual es incorrecta")
		return
	}

	newHash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HASH_ERROR", "No se pudo actualizar la contraseña")
		return
	}
	if _, err := s.store.DB.Exec(`UPDATE users SET password_hash = ? WHERE id = ?`, newHash, claims.UserID); err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error guardando la nueva contraseña")
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "password_updated"})
}

// --- Users (admin) ---

func (s *Server) handleListUsers(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.DB.Query(`SELECT id, username, full_name, role, active, created_at FROM users ORDER BY id ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando usuarios")
		return
	}
	defer rows.Close()

	users := []UserDTO{}
	for rows.Next() {
		var u UserDTO
		var active int
		if err := rows.Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &active, &u.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo usuarios")
			return
		}
		u.Active = active == 1
		users = append(users, u)
	}
	writeJSON(w, http.StatusOK, users)
}

func (s *Server) handleCreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Username string `json:"username"`
		FullName string `json:"full_name"`
		Password string `json:"password"`
		Role     string `json:"role"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	req.Username = strings.TrimSpace(strings.ToLower(req.Username))
	req.FullName = strings.TrimSpace(req.FullName)
	if req.Role == "" {
		req.Role = auth.RoleAprendiz
	}
	if req.Username == "" || req.FullName == "" || len(req.Password) < 6 {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "Se requiere username, full_name y password (mínimo 6 caracteres)")
		return
	}
	if req.Role != auth.RoleAdmin && req.Role != auth.RoleAprendiz {
		writeError(w, http.StatusBadRequest, "INVALID_ROLE", "El rol debe ser 'admin' o 'aprendiz'")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "HASH_ERROR", "Error procesando contraseña")
		return
	}
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := s.store.DB.Exec(`INSERT INTO users (username, full_name, password_hash, role, active, created_at)
		VALUES (?, ?, ?, ?, 1, ?)`, req.Username, req.FullName, hash, req.Role, nowStr)
	if err != nil {
		if isUniqueConstraint(err) {
			writeError(w, http.StatusConflict, "USERNAME_EXISTS", "Ya existe un usuario con ese nombre")
			return
		}
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error creando usuario")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, UserDTO{
		ID:        id,
		Username:  req.Username,
		FullName:  req.FullName,
		Role:      req.Role,
		Active:    true,
		CreatedAt: nowStr,
	})
}

func (s *Server) handleUpdateUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
		return
	}
	var req struct {
		FullName *string `json:"full_name"`
		Role     *string `json:"role"`
		Active   *bool   `json:"active"`
		Password *string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}

	claims := getClaims(r)
	if claims.UserID == id && req.Active != nil && !*req.Active {
		writeError(w, http.StatusBadRequest, "SELF_DISABLE", "No puedes desactivar tu propia cuenta")
		return
	}

	var u UserDTO
	var active int
	err = s.store.DB.QueryRow(`SELECT id, username, full_name, role, active, created_at FROM users WHERE id = ?`, id).
		Scan(&u.ID, &u.Username, &u.FullName, &u.Role, &active, &u.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando usuario")
		return
	}
	u.Active = active == 1

	if req.FullName != nil && strings.TrimSpace(*req.FullName) != "" {
		u.FullName = strings.TrimSpace(*req.FullName)
	}
	if req.Role != nil {
		if *req.Role != auth.RoleAdmin && *req.Role != auth.RoleAprendiz {
			writeError(w, http.StatusBadRequest, "INVALID_ROLE", "Rol inválido")
			return
		}
		u.Role = *req.Role
	}
	if req.Active != nil {
		u.Active = *req.Active
	}

	activeInt := 0
	if u.Active {
		activeInt = 1
	}

	if req.Password != nil && *req.Password != "" {
		if len(*req.Password) < 6 {
			writeError(w, http.StatusBadRequest, "WEAK_PASSWORD", "La contraseña debe tener al menos 6 caracteres")
			return
		}
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "HASH_ERROR", "Error procesando contraseña")
			return
		}
		_, err = s.store.DB.Exec(`UPDATE users SET full_name = ?, role = ?, active = ?, password_hash = ? WHERE id = ?`,
			u.FullName, u.Role, activeInt, hash, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error actualizando usuario")
			return
		}
	} else {
		_, err = s.store.DB.Exec(`UPDATE users SET full_name = ?, role = ?, active = ? WHERE id = ?`,
			u.FullName, u.Role, activeInt, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error actualizando usuario")
			return
		}
	}

	writeJSON(w, http.StatusOK, u)
}

func (s *Server) handleDeleteUser(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de usuario inválido")
		return
	}
	if getClaims(r).UserID == id {
		writeError(w, http.StatusBadRequest, "SELF_DELETE", "No puedes eliminar tu propia cuenta")
		return
	}
	res, err := s.store.DB.Exec(`DELETE FROM users WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando usuario")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "USER_NOT_FOUND", "Usuario no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Piles ---

func (s *Server) handleListPiles(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")
	if statusFilter == "" {
		statusFilter = "active"
	}

	query := `SELECT id, name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at FROM piles`
	var args []any
	if statusFilter == "active" || statusFilter == "archived" {
		query += ` WHERE status = ?`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY CASE WHEN status = 'active' THEN 0 ELSE 1 END, id ASC`

	rows, err := s.store.DB.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando pilas")
		return
	}
	defer rows.Close()

	var rawPiles []PileDTO
	for rows.Next() {
		p, err := scanPile(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo pilas")
			return
		}
		rawPiles = append(rawPiles, p)
	}
	rows.Close()

	result := make([]PileDTO, 0, len(rawPiles))
	now := time.Now().UTC()
	for _, p := range rawPiles {
		enriched, err := s.enrichPile(p, now)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error evaluando pilas")
			return
		}
		result = append(result, enriched)
	}
	writeJSON(w, http.StatusOK, result)
}

func (s *Server) handleGetPile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	p, err := s.getPileByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando pila")
		return
	}
	enriched, err := s.enrichPile(p, time.Now().UTC())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error evaluando pila")
		return
	}
	writeJSON(w, http.StatusOK, enriched)
}

func (s *Server) handleCreatePile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Location    string `json:"location"`
		Substrate   string `json:"substrate"`
		NodeID      string `json:"node_id"`
		HasPHSensor bool   `json:"has_ph_sensor"`
		StartedAt   string `json:"started_at"`
		Notes       string `json:"notes"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.NodeID = strings.ToUpper(strings.TrimSpace(req.NodeID))
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El nombre de la pila es obligatorio")
		return
	}
	if req.NodeID == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El identificador de nodo (node_id) es obligatorio")
		return
	}

	now := time.Now().UTC()
	startedAt := now
	if strings.TrimSpace(req.StartedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.StartedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "started_at debe estar en formato ISO 8601 (RFC3339)")
			return
		}
		startedAt = parsed.UTC()
	}

	hasPH := 0
	if req.HasPHSensor {
		hasPH = 1
	}
	nowStr := now.Format(time.RFC3339)
	startStr := startedAt.Format(time.RFC3339)

	res, err := s.store.DB.Exec(`INSERT INTO piles (name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, 'active', ?, ?, ?, ?)`,
		req.Name, strings.TrimSpace(req.Location), strings.TrimSpace(req.Substrate),
		req.NodeID, hasPH, strings.TrimSpace(req.Notes), startStr, nowStr, nowStr)
	if err != nil {
		if isUniqueConstraint(err) {
			writeError(w, http.StatusConflict, "NODE_IN_USE", "Ya existe una pila activa asociada a ese node_id")
			return
		}
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error creando pila")
		return
	}

	id, _ := res.LastInsertId()
	p, err := s.getPileByID(id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error recuperando pila creada")
		return
	}
	enriched, _ := s.enrichPile(p, now)
	s.broker.publish("pile_changed", map[string]any{"action": "created", "pile_id": id})
	writeJSON(w, http.StatusCreated, enriched)
}

func (s *Server) handleUpdatePile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	existing, err := s.getPileByID(id)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando pila")
		return
	}

	var req struct {
		Name        *string `json:"name"`
		Location    *string `json:"location"`
		Substrate   *string `json:"substrate"`
		NodeID      *string `json:"node_id"`
		HasPHSensor *bool   `json:"has_ph_sensor"`
		Status      *string `json:"status"`
		Notes       *string `json:"notes"`
		StartedAt   *string `json:"started_at"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}

	if req.Name != nil && strings.TrimSpace(*req.Name) != "" {
		existing.Name = strings.TrimSpace(*req.Name)
	}
	if req.Location != nil {
		existing.Location = strings.TrimSpace(*req.Location)
	}
	if req.Substrate != nil {
		existing.Substrate = strings.TrimSpace(*req.Substrate)
	}
	if req.NodeID != nil && strings.TrimSpace(*req.NodeID) != "" {
		existing.NodeID = strings.ToUpper(strings.TrimSpace(*req.NodeID))
	}
	if req.HasPHSensor != nil {
		existing.HasPHSensor = *req.HasPHSensor
	}
	if req.Status != nil {
		if *req.Status != "active" && *req.Status != "archived" {
			writeError(w, http.StatusBadRequest, "INVALID_STATUS", "status debe ser 'active' o 'archived'")
			return
		}
		existing.Status = *req.Status
	}
	if req.Notes != nil {
		existing.Notes = strings.TrimSpace(*req.Notes)
	}
	if req.StartedAt != nil && strings.TrimSpace(*req.StartedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, *req.StartedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "started_at inválido")
			return
		}
		existing.StartedAt = parsed.UTC().Format(time.RFC3339)
	}

	hasPH := 0
	if existing.HasPHSensor {
		hasPH = 1
	}
	now := time.Now().UTC()
	existing.UpdatedAt = now.Format(time.RFC3339)

	_, err = s.store.DB.Exec(`UPDATE piles
		SET name = ?, location = ?, substrate = ?, node_id = ?, has_ph_sensor = ?, status = ?, notes = ?, started_at = ?, updated_at = ?
		WHERE id = ?`,
		existing.Name, existing.Location, existing.Substrate, existing.NodeID,
		hasPH, existing.Status, existing.Notes, existing.StartedAt, existing.UpdatedAt, id)
	if err != nil {
		if isUniqueConstraint(err) {
			writeError(w, http.StatusConflict, "NODE_IN_USE", "Otra pila activa ya utiliza ese node_id")
			return
		}
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error actualizando pila")
		return
	}

	enriched, err := s.enrichPile(existing, now)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error evaluando pila")
		return
	}
	writeJSON(w, http.StatusOK, enriched)
}

func (s *Server) handleDeletePile(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	res, err := s.store.DB.Exec(`DELETE FROM piles WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error eliminando pila")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Readings & Ingest ---

func (s *Server) handleListReadings(w http.ResponseWriter, r *http.Request) {
	pileID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	if _, err := s.getPileByID(pileID); errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	}

	limit := 500
	if rawLimit := r.URL.Query().Get("limit"); rawLimit != "" {
		if n, err := strconv.Atoi(rawLimit); err == nil && n > 0 && n <= 5000 {
			limit = n
		}
	}

	where := []string{"pile_id = ?"}
	args := []any{pileID}

	if from := r.URL.Query().Get("from"); from != "" {
		if t, err := time.Parse(time.RFC3339, from); err == nil {
			where = append(where, "recorded_at >= ?")
			args = append(args, t.UTC().Format(time.RFC3339))
		}
	}
	if to := r.URL.Query().Get("to"); to != "" {
		if t, err := time.Parse(time.RFC3339, to); err == nil {
			where = append(where, "recorded_at <= ?")
			args = append(args, t.UTC().Format(time.RFC3339))
		}
	}

	// Seleccionar las últimas `limit` lecturas y devolverlas en orden cronológico ascendente
	query := fmt.Sprintf(`SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
		FROM (
			SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
			FROM readings WHERE %s ORDER BY recorded_at DESC, id DESC LIMIT ?
		) ORDER BY recorded_at ASC, id ASC`, strings.Join(where, " AND "))
	args = append(args, limit)

	rows, err := s.store.DB.Query(query, args...)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando lecturas")
		return
	}
	defer rows.Close()

	readings := []ReadingDTO{}
	for rows.Next() {
		rd, err := scanReading(rows)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo registros")
			return
		}
		readings = append(readings, rd)
	}
	writeJSON(w, http.StatusOK, readings)
}

type readingInput struct {
	NodeID       string   `json:"node_id"`
	TemperatureC *float64 `json:"temperature_c"`
	HumidityPct  *float64 `json:"humidity_pct"`
	PH           *float64 `json:"ph"`
	PowerSource  string   `json:"power_source"`
	BatteryV     *float64 `json:"battery_v"`
	RSSI         *int     `json:"rssi"`
	Seq          *int     `json:"seq"`
	RecordedAt   string   `json:"recorded_at"`
}

func (s *Server) handleCreateManualReading(w http.ResponseWriter, r *http.Request) {
	pileID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	p, err := s.getPileByID(pileID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando pila")
		return
	}

	var req readingInput
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	s.saveReadingForPile(w, p, req)
}

func (s *Server) handleIngestReading(w http.ResponseWriter, r *http.Request) {
	var req readingInput
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	req.NodeID = strings.ToUpper(strings.TrimSpace(req.NodeID))
	if req.NodeID == "" {
		writeError(w, http.StatusBadRequest, "MISSING_NODE_ID", "El campo node_id es obligatorio")
		return
	}

	// Actualizar last_seen_at del gateway
	deviceID := getDeviceID(r)
	_, _ = s.store.DB.Exec(`UPDATE devices SET last_seen_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), deviceID)

	// Buscar pila activa asociada al node_id
	row := s.store.DB.QueryRow(`SELECT id, name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at
		FROM piles WHERE node_id = ? AND status = 'active'`, req.NodeID)
	p, err := scanPileRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "NODE_UNASSIGNED", fmt.Sprintf("No hay ninguna pila activa asignada al nodo %s", req.NodeID))
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error buscando pila para el nodo")
		return
	}

	s.saveReadingForPile(w, p, req)
}

func (s *Server) saveReadingForPile(w http.ResponseWriter, p PileDTO, req readingInput) {
	if req.TemperatureC == nil || req.HumidityPct == nil {
		writeError(w, http.StatusBadRequest, "MISSING_VARIABLES", "temperature_c y humidity_pct son obligatorias")
		return
	}
	if *req.TemperatureC < -10.0 || *req.TemperatureC > 110.0 {
		writeError(w, http.StatusBadRequest, "INVALID_TEMPERATURE", "temperature_c fuera de rango válido (-10 a 110 °C)")
		return
	}
	if *req.HumidityPct < 0.0 || *req.HumidityPct > 100.0 {
		writeError(w, http.StatusBadRequest, "INVALID_HUMIDITY", "humidity_pct fuera de rango válido (0 a 100 %)")
		return
	}
	if req.PH != nil && (*req.PH < 0.0 || *req.PH > 14.0) {
		writeError(w, http.StatusBadRequest, "INVALID_PH", "ph fuera de rango válido (0 a 14)")
		return
	}

	powerSource := req.PowerSource
	if powerSource != "battery" {
		powerSource = "mains"
	}

	recordedAt := time.Now().UTC()
	if strings.TrimSpace(req.RecordedAt) != "" {
		parsed, err := time.Parse(time.RFC3339, req.RecordedAt)
		if err != nil {
			writeError(w, http.StatusBadRequest, "INVALID_DATE", "recorded_at debe estar en formato ISO 8601 (RFC3339)")
			return
		}
		recordedAt = parsed.UTC()
	}

	// Evaluar fase incluyendo la nueva medición
	samples, err := s.getSamplesForPile(p.ID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error obteniendo historial de la pila")
		return
	}
	samples = append(samples, compost.ReadingSample{
		TemperatureC: *req.TemperatureC,
		HumidityPct:  *req.HumidityPct,
		PH:           req.PH,
		RecordedAt:   recordedAt,
	})
	startedAt, _ := time.Parse(time.RFC3339, p.StartedAt)
	eval := compost.Evaluate(startedAt, samples, p.Status, recordedAt)

	recStr := recordedAt.Format(time.RFC3339)
	res, err := s.store.DB.Exec(`INSERT INTO readings (pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.ID, *req.TemperatureC, *req.HumidityPct, req.PH, string(eval.Phase.Code),
		powerSource, req.BatteryV, req.RSSI, req.Seq, recStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error guardando lectura")
		return
	}
	id, _ := res.LastInsertId()

	// Si llega pH por primera vez, marcar automáticamente que la pila dispone de pH
	if req.PH != nil && !p.HasPHSensor {
		_, _ = s.store.DB.Exec(`UPDATE piles SET has_ph_sensor = 1, updated_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), p.ID)
	} else {
		_, _ = s.store.DB.Exec(`UPDATE piles SET updated_at = ? WHERE id = ?`, time.Now().UTC().Format(time.RFC3339), p.ID)
	}

	saved := ReadingDTO{
		ID:           id,
		PileID:       p.ID,
		TemperatureC: *req.TemperatureC,
		HumidityPct:  *req.HumidityPct,
		PH:           req.PH,
		Phase:        string(eval.Phase.Code),
		PowerSource:  powerSource,
		BatteryV:     req.BatteryV,
		RSSI:         req.RSSI,
		Seq:          req.Seq,
		RecordedAt:   recStr,
	}

	s.broker.publish("reading", map[string]any{
		"pile_id":    p.ID,
		"node_id":    p.NodeID,
		"reading":    saved,
		"evaluation": eval,
	})

	writeJSON(w, http.StatusCreated, map[string]any{
		"reading":    saved,
		"evaluation": eval,
	})
}

func (s *Server) handleExportCSV(w http.ResponseWriter, r *http.Request) {
	pileID, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de pila inválido")
		return
	}
	p, err := s.getPileByID(pileID)
	if errors.Is(err, sql.ErrNoRows) {
		writeError(w, http.StatusNotFound, "PILE_NOT_FOUND", "Pila no encontrada")
		return
	} else if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando pila")
		return
	}

	rows, err := s.store.DB.Query(`SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
		FROM readings WHERE pile_id = ? ORDER BY recorded_at ASC, id ASC`, pileID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error exportando lecturas")
		return
	}
	defer rows.Close()

	filename := fmt.Sprintf("compostiot-pila-%d-%s.csv", p.ID, strings.ToLower(p.NodeID))
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filename))
	// UTF-8 BOM para compatibilidad directa con Excel
	_, _ = w.Write([]byte("\xEF\xBB\xBF"))

	cw := csv.NewWriter(w)
	_ = cw.Write([]string{
		"fecha_hora", "pila", "nodo", "temperatura_c", "humedad_pct", "ph", "fase", "fuente_energia", "bateria_v", "rssi_dbm",
	})
	for rows.Next() {
		rd, err := scanReading(rows)
		if err != nil {
			continue
		}
		phStr := ""
		if rd.PH != nil {
			phStr = fmt.Sprintf("%.2f", *rd.PH)
		}
		battStr := ""
		if rd.BatteryV != nil {
			battStr = fmt.Sprintf("%.2f", *rd.BatteryV)
		}
		rssiStr := ""
		if rd.RSSI != nil {
			rssiStr = strconv.Itoa(*rd.RSSI)
		}
		_ = cw.Write([]string{
			rd.RecordedAt,
			p.Name,
			p.NodeID,
			fmt.Sprintf("%.1f", rd.TemperatureC),
			fmt.Sprintf("%.1f", rd.HumidityPct),
			phStr,
			compost.InfoForPhase(compost.PhaseCode(rd.Phase)).Label,
			rd.PowerSource,
			battStr,
			rssiStr,
		})
	}
	cw.Flush()
}

// --- Devices / Gateways ---

func (s *Server) handleListDevices(w http.ResponseWriter, r *http.Request) {
	rows, err := s.store.DB.Query(`SELECT id, name, key_prefix, active, last_seen_at, created_at FROM devices ORDER BY id ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error consultando gateways")
		return
	}
	defer rows.Close()

	devices := []DeviceDTO{}
	for rows.Next() {
		var d DeviceDTO
		var active int
		var lastSeen sql.NullString
		if err := rows.Scan(&d.ID, &d.Name, &d.KeyPrefix, &active, &lastSeen, &d.CreatedAt); err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error leyendo gateways")
			return
		}
		d.Active = active == 1
		if lastSeen.Valid {
			d.LastSeenAt = &lastSeen.String
		}
		devices = append(devices, d)
	}
	writeJSON(w, http.StatusOK, devices)
}

func (s *Server) handleCreateDevice(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_JSON", "Cuerpo JSON inválido")
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		writeError(w, http.StatusBadRequest, "VALIDATION_ERROR", "El nombre del gateway es obligatorio")
		return
	}

	plain, prefix, hash, err := auth.GenerateAPIKey()
	if err != nil {
		writeError(w, http.StatusInternalServerError, "KEY_ERROR", "Error generando clave de gateway")
		return
	}
	nowStr := time.Now().UTC().Format(time.RFC3339)
	res, err := s.store.DB.Exec(`INSERT INTO devices (name, key_prefix, key_hash, active, created_at) VALUES (?, ?, ?, 1, ?)`,
		req.Name, prefix, hash, nowStr)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error registrando gateway")
		return
	}
	id, _ := res.LastInsertId()
	writeJSON(w, http.StatusCreated, map[string]any{
		"device": DeviceDTO{
			ID:        id,
			Name:      req.Name,
			KeyPrefix: prefix,
			Active:    true,
			CreatedAt: nowStr,
		},
		"api_key": plain, // Se muestra una única vez
	})
}

func (s *Server) handleDeleteDevice(w http.ResponseWriter, r *http.Request) {
	id, err := parseID(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusBadRequest, "INVALID_ID", "ID de gateway inválido")
		return
	}
	res, err := s.store.DB.Exec(`DELETE FROM devices WHERE id = ?`, id)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error revocando gateway")
		return
	}
	if n, _ := res.RowsAffected(); n == 0 {
		writeError(w, http.StatusNotFound, "DEVICE_NOT_FOUND", "Gateway no encontrado")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Sync Snapshot (Wails & Mobile Offline Cache) ---

func (s *Server) handleSyncSnapshot(w http.ResponseWriter, r *http.Request) {
	now := time.Now().UTC()
	sinceStr := strings.TrimSpace(r.URL.Query().Get("since"))

	rows, err := s.store.DB.Query(`SELECT id, name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at FROM piles ORDER BY id ASC`)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error sincronizando pilas")
		return
	}
	var rawPiles []PileDTO
	for rows.Next() {
		if p, err := scanPile(rows); err == nil {
			rawPiles = append(rawPiles, p)
		}
	}
	rows.Close()

	piles := make([]PileDTO, 0, len(rawPiles))
	for _, p := range rawPiles {
		enriched, err := s.enrichPile(p, now)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error evaluando pilas para sincronización")
			return
		}
		piles = append(piles, enriched)
	}

	var readingRows *sql.Rows
	if sinceStr != "" {
		if t, err := time.Parse(time.RFC3339, sinceStr); err == nil {
			readingRows, err = s.store.DB.Query(`SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
				FROM readings WHERE recorded_at >= ? ORDER BY recorded_at ASC, id ASC LIMIT 5000`, t.UTC().Format(time.RFC3339))
		}
	}
	if readingRows == nil {
		readingRows, err = s.store.DB.Query(`SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
			FROM readings ORDER BY recorded_at DESC, id DESC LIMIT 2000`)
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "DB_ERROR", "Error sincronizando lecturas")
		return
	}
	defer readingRows.Close()

	readings := []ReadingDTO{}
	for readingRows.Next() {
		if rd, err := scanReading(readingRows); err == nil {
			readings = append(readings, rd)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"server_name": s.cfg.ServerName,
		"server_time": now.Format(time.RFC3339),
		"piles":       piles,
		"readings":    readings,
	})
}

// --- Helpers internos ---

func (s *Server) getPileByID(id int64) (PileDTO, error) {
	row := s.store.DB.QueryRow(`SELECT id, name, location, substrate, node_id, has_ph_sensor, status, notes, started_at, created_at, updated_at
		FROM piles WHERE id = ?`, id)
	return scanPileRow(row)
}

func (s *Server) getSamplesForPile(pileID int64) ([]compost.ReadingSample, error) {
	rows, err := s.store.DB.Query(`SELECT temperature_c, humidity_pct, ph, recorded_at FROM readings WHERE pile_id = ? ORDER BY recorded_at ASC, id ASC`, pileID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var samples []compost.ReadingSample
	for rows.Next() {
		var temp, hum float64
		var ph sql.NullFloat64
		var recStr string
		if err := rows.Scan(&temp, &hum, &ph, &recStr); err != nil {
			return nil, err
		}
		ts, _ := time.Parse(time.RFC3339, recStr)
		var phPtr *float64
		if ph.Valid {
			v := ph.Float64
			phPtr = &v
		}
		samples = append(samples, compost.ReadingSample{
			TemperatureC: temp,
			HumidityPct:  hum,
			PH:           phPtr,
			RecordedAt:   ts,
		})
	}
	return samples, nil
}

func (s *Server) enrichPile(p PileDTO, now time.Time) (PileDTO, error) {
	samples, err := s.getSamplesForPile(p.ID)
	if err != nil {
		return p, err
	}
	startedAt, _ := time.Parse(time.RFC3339, p.StartedAt)
	p.Evaluation = compost.Evaluate(startedAt, samples, p.Status, now)

	row := s.store.DB.QueryRow(`SELECT id, pile_id, temperature_c, humidity_pct, ph, phase, power_source, battery_v, rssi, seq, recorded_at
		FROM readings WHERE pile_id = ? ORDER BY recorded_at DESC, id DESC LIMIT 1`, p.ID)
	rd, err := scanReadingRow(row)
	if err == nil {
		p.LatestReading = &rd
	}
	return p, nil
}

func scanPile(rows *sql.Rows) (PileDTO, error) {
	var p PileDTO
	var hasPH int
	err := rows.Scan(&p.ID, &p.Name, &p.Location, &p.Substrate, &p.NodeID, &hasPH, &p.Status, &p.Notes, &p.StartedAt, &p.CreatedAt, &p.UpdatedAt)
	p.HasPHSensor = hasPH == 1
	return p, err
}

func scanPileRow(row *sql.Row) (PileDTO, error) {
	var p PileDTO
	var hasPH int
	err := row.Scan(&p.ID, &p.Name, &p.Location, &p.Substrate, &p.NodeID, &hasPH, &p.Status, &p.Notes, &p.StartedAt, &p.CreatedAt, &p.UpdatedAt)
	p.HasPHSensor = hasPH == 1
	return p, err
}

func scanReading(rows *sql.Rows) (ReadingDTO, error) {
	var rd ReadingDTO
	var ph, batt sql.NullFloat64
	var rssi, seq sql.NullInt64
	err := rows.Scan(&rd.ID, &rd.PileID, &rd.TemperatureC, &rd.HumidityPct, &ph, &rd.Phase, &rd.PowerSource, &batt, &rssi, &seq, &rd.RecordedAt)
	if ph.Valid {
		v := ph.Float64
		rd.PH = &v
	}
	if batt.Valid {
		v := batt.Float64
		rd.BatteryV = &v
	}
	if rssi.Valid {
		v := int(rssi.Int64)
		rd.RSSI = &v
	}
	if seq.Valid {
		v := int(seq.Int64)
		rd.Seq = &v
	}
	return rd, err
}

func scanReadingRow(row *sql.Row) (ReadingDTO, error) {
	var rd ReadingDTO
	var ph, batt sql.NullFloat64
	var rssi, seq sql.NullInt64
	err := row.Scan(&rd.ID, &rd.PileID, &rd.TemperatureC, &rd.HumidityPct, &ph, &rd.Phase, &rd.PowerSource, &batt, &rssi, &seq, &rd.RecordedAt)
	if ph.Valid {
		v := ph.Float64
		rd.PH = &v
	}
	if batt.Valid {
		v := batt.Float64
		rd.BatteryV = &v
	}
	if rssi.Valid {
		v := int(rssi.Int64)
		rd.RSSI = &v
	}
	if seq.Valid {
		v := int(seq.Int64)
		rd.Seq = &v
	}
	return rd, err
}

func parseID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || id <= 0 {
		return 0, errors.New("invalid id")
	}
	return id, nil
}

func isUniqueConstraint(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "unique constraint")
}
