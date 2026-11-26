package main

import (
	"embed"
	"log"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	app, err := NewApp()
	if err != nil {
		log.Fatalf("error iniciando CompostIOT Desktop: %v", err)
	}

	err = wails.Run(&options.App{
		Title:     "CompostIOT · Monitor de Compostaje SENA",
		Width:     1240,
		Height:    820,
		MinWidth:  420,
		MinHeight: 640,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 248, G: 250, B: 249, A: 1},
		OnStartup:        app.startup,
		OnShutdown:       app.shutdown,
		Bind: []any{
			app,
		},
	})
	if err != nil {
		log.Fatalf("error ejecutando Wails: %v", err)
	}
}
