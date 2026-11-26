package main

import (
	"context"
	"log"

	"github.com/compostiot/desktop/internal/cache"
	"github.com/compostiot/desktop/internal/client"
	"github.com/compostiot/desktop/internal/models"
	"github.com/compostiot/desktop/internal/secrets"
)

// App expone los métodos nativos a la interfaz React mediante el runtime de Wails (window.go.main.App).
type App struct {
	ctx   context.Context
	store *cache.Store
	svc   *client.Service
}

func NewApp() (*App, error) {
	dbPath, err := cache.DefaultDBPath()
	if err != nil {
		return nil, err
	}
	store, err := cache.Open(dbPath)
	if err != nil {
		return nil, err
	}
	log.Printf("[CompostIOT Desktop] Caché SQLite inicializado en %s", dbPath)
	return &App{
		store: store,
		svc:   client.NewService(store, secrets.NewKeyringStore()),
	}, nil
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) shutdown(ctx context.Context) {
	if a.store != nil {
		_ = a.store.Close()
	}
}

// --- Métodos enlazados a frontend (WailsDesktopBridge) ---

func (a *App) GetServers() ([]models.ServerProfile, error) {
	return a.svc.GetServers()
}

func (a *App) SaveServer(profile models.ServerProfile) (models.ServerProfile, error) {
	return a.svc.SaveServer(profile)
}

func (a *App) DeleteServer(id string) error {
	return a.svc.DeleteServer(id)
}

func (a *App) CheckHealth(serverURL string) (models.HealthInfo, error) {
	return a.svc.CheckHealth(serverURL)
}

func (a *App) Login(serverID, username, password string) (models.LoginResult, error) {
	return a.svc.Login(serverID, username, password)
}

func (a *App) RestoreSession(serverID string) (models.SessionRestoreResult, error) {
	return a.svc.RestoreSession(serverID)
}

func (a *App) Logout(serverID string) error {
	return a.svc.Logout(serverID)
}

func (a *App) SyncServer(serverID string) (models.PilesResponse, error) {
	return a.svc.SyncServer(serverID)
}

func (a *App) GetPiles(serverID, status string) (models.PilesResponse, error) {
	return a.svc.GetPiles(serverID, status)
}

func (a *App) GetReadings(serverID string, pileID int64, limit int) (models.ReadingsResponse, error) {
	return a.svc.GetReadings(serverID, pileID, limit)
}

func (a *App) CreatePile(serverID string, input models.PileInput) (models.Pile, error) {
	return a.svc.CreatePile(serverID, input)
}

func (a *App) UpdatePile(serverID string, pileID int64, input models.PileInput) (models.Pile, error) {
	return a.svc.UpdatePile(serverID, pileID, input)
}

func (a *App) DeletePile(serverID string, pileID int64) error {
	return a.svc.DeletePile(serverID, pileID)
}

func (a *App) AddManualReading(serverID string, pileID int64, input models.ManualReadingInput) (models.Reading, error) {
	return a.svc.AddManualReading(serverID, pileID, input)
}

func (a *App) ExportCSV(serverID string, pileID int64) (string, error) {
	return a.svc.ExportCSV(serverID, pileID)
}

func (a *App) GetUsers(serverID string) ([]models.User, error) {
	return a.svc.GetUsers(serverID)
}

func (a *App) CreateUser(serverID string, payload models.CreateUserInput) (models.User, error) {
	return a.svc.CreateUser(serverID, payload)
}

func (a *App) DeleteUser(serverID string, userID int64) error {
	return a.svc.DeleteUser(serverID, userID)
}

func (a *App) GetDevices(serverID string) ([]models.Device, error) {
	return a.svc.GetDevices(serverID)
}

func (a *App) CreateDevice(serverID, name string) (models.CreateDeviceResult, error) {
	return a.svc.CreateDevice(serverID, name)
}

func (a *App) DeleteDevice(serverID string, deviceID int64) error {
	return a.svc.DeleteDevice(serverID, deviceID)
}
