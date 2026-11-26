package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/compostiot/backend/internal/auth"
)

type sseEvent struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}

type eventBroker struct {
	mu      sync.RWMutex
	clients map[chan []byte]struct{}
}

func newEventBroker() *eventBroker {
	return &eventBroker{
		clients: make(map[chan []byte]struct{}),
	}
}

func (b *eventBroker) subscribe() chan []byte {
	ch := make(chan []byte, 16)
	b.mu.Lock()
	b.clients[ch] = struct{}{}
	b.mu.Unlock()
	return ch
}

func (b *eventBroker) unsubscribe(ch chan []byte) {
	b.mu.Lock()
	if _, ok := b.clients[ch]; ok {
		delete(b.clients, ch)
		close(ch)
	}
	b.mu.Unlock()
}

func (b *eventBroker) publish(eventType string, payload any) {
	raw, err := json.Marshal(sseEvent{Type: eventType, Payload: payload})
	if err != nil {
		return
	}
	b.mu.RLock()
	defer b.mu.RUnlock()
	for ch := range b.clients {
		select {
		case ch <- raw:
		default:
			// Cliente lento: se descarta evento puntual sin bloquear ingesta
		}
	}
}

// handleEvents sirve un flujo SSE (Server-Sent Events) para actualizaciones instantáneas en Web/Wails/Móvil.
// Acepta token Bearer por cabecera Authorization o por parámetro ?token= (estándar para EventSource en navegador).
func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	tokenStr := ""
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		tokenStr = strings.TrimSpace(strings.TrimPrefix(h, "Bearer "))
	} else {
		tokenStr = strings.TrimSpace(r.URL.Query().Get("token"))
	}
	if tokenStr == "" {
		writeError(w, http.StatusUnauthorized, "UNAUTHORIZED", "Se requiere token para suscribirse a eventos en tiempo real")
		return
	}
	if _, err := auth.ParseToken(s.cfg.JWTSecret, tokenStr); err != nil {
		writeError(w, http.StatusUnauthorized, "INVALID_TOKEN", "Token inválido o expirado")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "SSE_UNSUPPORTED", "Streaming SSE no soportado")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no")

	ch := s.broker.subscribe()
	defer s.broker.unsubscribe(ch)

	// Confirmación inicial de conexión en vivo
	fmt.Fprintf(w, "event: connected\ndata: {\"status\":\"live\"}\n\n")
	flusher.Flush()

	keepalive := time.NewTicker(25 * time.Second)
	defer keepalive.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case <-keepalive.C:
			fmt.Fprintf(w, ": ping\n\n")
			flusher.Flush()
		case msg, open := <-ch:
			if !open {
				return
			}
			fmt.Fprintf(w, "event: message\ndata: %s\n\n", msg)
			flusher.Flush()
		}
	}
}
