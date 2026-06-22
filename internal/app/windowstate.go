package app

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const defaultWindowWidth = 1200
const defaultWindowHeight = 820
const minimumRememberedWindowWidth = 640
const minimumRememberedWindowHeight = 480

// windowState is the persisted window geometry restored on the next launch.
type windowState struct {
	Width       int  `json:"width"`
	Height      int  `json:"height"`
	X           int  `json:"x"`
	Y           int  `json:"y"`
	Maximised   bool `json:"maximised"`
	HasGeometry bool `json:"hasGeometry"`
}

func windowStateFilePath() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDirectory, "PromptfulCustomFFmpegBuilder", "window-state.json"), nil
}

// loadWindowState returns the saved geometry, or sane defaults when nothing
// valid is stored yet.
func loadWindowState() windowState {
	state := windowState{Width: defaultWindowWidth, Height: defaultWindowHeight}
	filePath, err := windowStateFilePath()
	if err != nil {
		return state
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return state
	}
	var savedState windowState
	if err := json.Unmarshal(fileData, &savedState); err != nil {
		return state
	}
	if savedState.Width < minimumRememberedWindowWidth || savedState.Height < minimumRememberedWindowHeight {
		return state
	}
	savedState.HasGeometry = true
	return savedState
}

func saveWindowState(state windowState) error {
	filePath, err := windowStateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	fileData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filePath, fileData, 0o644)
}

// restoreWindowGeometry applies the saved size, position, and maximised state
// once the Wails runtime context is available.
func (app *App) restoreWindowGeometry() {
	if app.ctx == nil || !app.startupWindowState.HasGeometry {
		return
	}
	state := app.startupWindowState
	wailsRuntime.WindowSetSize(app.ctx, state.Width, state.Height)
	wailsRuntime.WindowSetPosition(app.ctx, state.X, state.Y)
	if state.Maximised {
		wailsRuntime.WindowMaximise(app.ctx)
	}
}

// persistWindowGeometry records the current window geometry so it can be
// restored on the next launch. When maximised, the last known normal bounds
// are kept so a later unmaximise restores a sane size.
func (app *App) persistWindowGeometry(ctx context.Context) {
	if ctx == nil {
		return
	}
	isMaximised := wailsRuntime.WindowIsMaximised(ctx)
	state := windowState{Maximised: isMaximised, HasGeometry: true}
	if isMaximised {
		state.Width = app.startupWindowState.Width
		state.Height = app.startupWindowState.Height
		state.X = app.startupWindowState.X
		state.Y = app.startupWindowState.Y
		if state.Width < minimumRememberedWindowWidth || state.Height < minimumRememberedWindowHeight {
			state.Width = defaultWindowWidth
			state.Height = defaultWindowHeight
		}
	} else {
		width, height := wailsRuntime.WindowGetSize(ctx)
		x, y := wailsRuntime.WindowGetPosition(ctx)
		state.Width = width
		state.Height = height
		state.X = x
		state.Y = y
	}
	_ = saveWindowState(state)
}
