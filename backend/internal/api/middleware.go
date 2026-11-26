package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/compostiot/backend/internal/auth"
)

type ctxKey string

const (
	ctxClaimsKey ctxKey = "auth_claims"
	ctxDeviceID  ctxKey = "device_id"
)

type errorEnvelope struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	var env errorEnvelope
	env.Error.Code = code
	env.Error.Message = message
	writeJSON(w, status, env)
}

func decodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(http.MaxBytesReader(nil, r.Body, 1<<20))
	return dec.Decode(dst)
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" {
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		} else {
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Device-Key")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s (%s)", r.Method, r.URL.Path, time.Since(start).Round(time.Millisecond))
	})
}

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Se requiere token Bearer en cabecera Authorization")
			return
		}
		tokenStr := strings.TrimSpace(strings.TrimPrefix(header, "Bearer "))
		claims, err := auth.ParseToken(s.cfg.JWTSecret, tokenStr)
		if err != nil {
			writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido o expirado")
			return
		}

		// Verificar que el usuario siga activo en SQLite
		var active int
		var role string
		err = s.store.DB.QueryRow(`SELECT active, role FROM users WHERE id = ?`, claims.UserID).Scan(&active, &role)
		if err != nil || active != 1 {
			writeError(w, http.StatusUnauthorized, "USER_INACTIVE", "La cuenta no se encuentra activa")
			return
		}
		claims.Role = role

		ctx := context.WithValue(r.Context(), ctxClaimsKey, claims)
		next(w, r.WithContext(ctx))
	}
}

func (s *Server) requireAdmin(next http.HandlerFunc) http.HandlerFunc {
	return s.requireAuth(func(w http.ResponseWriter, r *http.Request) {
		claims := getClaims(r)
		if claims == nil || claims.Role != auth.RoleAdmin {
			writeError(w, http.StatusForbidden, "FORBIDDEN", "Esta acción requiere rol de Instructor/Administrador")
			return
		}
		next(w, r)
	})
}

func (s *Server) requireDeviceKey(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		apiKey := strings.TrimSpace(r.Header.Get("X-Device-Key"))
		if apiKey == "" {
			writeError(w, http.StatusUnauthorized, "MISSING_DEVICE_KEY", "Se requiere cabecera X-Device-Key")
			return
		}
		hash := auth.HashAPIKey(apiKey)
		var deviceID int64
		var active int
		err := s.store.DB.QueryRow(`SELECT id, active FROM devices WHERE key_hash = ?`, hash).Scan(&deviceID, &active)
		if err != nil || active != 1 {
			writeError(w, http.StatusUnauthorized, "INVALID_DEVICE_KEY", "Clave de dispositivo inválida o revocada")
			return
		}
		ctx := context.WithValue(r.Context(), ctxDeviceID, deviceID)
		next(w, r.WithContext(ctx))
	}
}

func getClaims(r *http.Request) *auth.Claims {
	v, _ := r.Context().Value(ctxClaimsKey).(*auth.Claims)
	return v
}

func getDeviceID(r *http.Request) int64 {
	v, _ := r.Context().Value(ctxDeviceID).(int64)
	return v
}
