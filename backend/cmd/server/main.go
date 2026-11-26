package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/compostiot/backend/internal/api"
	"github.com/compostiot/backend/internal/config"
	"github.com/compostiot/backend/internal/db"
)

func main() {
	cfg := config.Load()

	store, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("error inicializando SQLite (%s): %v", cfg.DBPath, err)
	}
	defer store.Close()

	if err := store.SeedDefaults(cfg.SeedDemo); err != nil {
		log.Fatalf("error ejecutando seed inicial: %v", err)
	}

	srv := api.NewServer(cfg, store)
	httpServer := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           srv.Handler(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("CompostIOT Backend v%s escuchando en http://localhost:%s (BD: %s)", api.ServiceVersion, cfg.Port, cfg.DBPath)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("error en servidor HTTP: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("Apagando CompostIOT Backend...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(ctx)
}
