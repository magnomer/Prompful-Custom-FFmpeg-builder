package main

import (
	"embed"

	backendapp "customffmpegbuilder/internal/app"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	application := backendapp.New()
	width, height := backendapp.InitialWindowSize(application)

	err := wails.Run(&options.App{
		Title:  backendapp.Localize("app.brand", nil),
		Width:  width,
		Height: height,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 23, B: 34, A: 1},
		OnStartup:        application.Startup,
		OnBeforeClose:    application.BeforeClose,
		OnShutdown:       application.Shutdown,
		Bind: []interface{}{
			application,
		},
	})
	if err != nil {
		println(backendapp.Localize("startup.failure", nil), err.Error())
	}
}
