package main

import (
	"embed"

	backendprogram "promptfulcustomffmpegbuilder/internal/program"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
)

//go:embed all:frontend/dist
var frontendAssets embed.FS

func main() {
	program := backendprogram.LProgramCreate()
	width, height := backendprogram.LWindowInitialRead(program)

	err := wails.Run(&options.App{
		Title:  backendprogram.LLocaleTextGet("app.brand", nil),
		Width:  width,
		Height: height,
		AssetServer: &assetserver.Options{
			Assets: frontendAssets,
		},
		BackgroundColour: &options.RGBA{R: 18, G: 23, B: 34, A: 1},
		OnStartup:        program.LProgramStart,
		OnBeforeClose:    program.LWindowCloseCheck,
		OnShutdown:       program.LProgramStop,
		Bind: []interface{}{
			program,
		},
	})
	if err != nil {
		println(backendprogram.LLocaleTextGet("startup.failure", nil), err.Error())
	}
}
