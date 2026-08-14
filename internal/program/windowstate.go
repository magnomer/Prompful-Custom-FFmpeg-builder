package program

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const LWindowWidthDefault = 1200
const LWindowHeightDefault = 820
const LWindowMinimumWidth = 640
const LWindowMinimumHeight = 480

// LStateWindow is the persisted window geometry restored on the next launch.
type LStateWindow struct {
	Width       int  `json:"width"`
	Height      int  `json:"height"`
	X           int  `json:"x"`
	Y           int  `json:"y"`
	Maximised   bool `json:"maximised"`
	HasGeometry bool `json:"hasGeometry"`
}

func LPathWindowResolve() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDirectory, "PromptfulCustomFfmpegBuilder", "window-state.json"), nil
}

// LStateWindowLoad returns the saved geometry, or sane defaults when nothing
// valid is stored yet.
func LStateWindowLoad() LStateWindow {
	state := LStateWindow{Width: LWindowWidthDefault, Height: LWindowHeightDefault}
	filePath, err := LPathWindowResolve()
	if err != nil {
		return state
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return state
	}
	var savedState LStateWindow
	if err := json.Unmarshal(fileData, &savedState); err != nil {
		return state
	}
	if savedState.Width < LWindowMinimumWidth || savedState.Height < LWindowMinimumHeight {
		return state
	}
	savedState.HasGeometry = true
	return savedState
}

func LStateWindowSave(state LStateWindow) error {
	filePath, err := LPathWindowResolve()
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

// LWindowGeometryRestore applies the saved size, position, and maximised state
// once the Wails runtime context is available.
func (program *LProgram) LWindowGeometryRestore() {
	if program.LContext == nil || !program.LStateWindowStartup.HasGeometry {
		return
	}
	state := program.LStateWindowStartup
	wailsRuntime.WindowSetSize(program.LContext, state.Width, state.Height)
	wailsRuntime.WindowSetPosition(program.LContext, state.X, state.Y)
	if state.Maximised {
		wailsRuntime.WindowMaximise(program.LContext)
	}
}

// LWindowGeometrySave records the current window geometry so it can be
// restored on the next launch. When maximised, the last known normal bounds
// are kept so a later unmaximise restores a sane size.
func (program *LProgram) LWindowGeometrySave(LContext context.Context) {
	if LContext == nil {
		return
	}
	isMaximised := wailsRuntime.WindowIsMaximised(LContext)
	state := LStateWindow{Maximised: isMaximised, HasGeometry: true}
	if isMaximised {
		state.Width = program.LStateWindowStartup.Width
		state.Height = program.LStateWindowStartup.Height
		state.X = program.LStateWindowStartup.X
		state.Y = program.LStateWindowStartup.Y
		if state.Width < LWindowMinimumWidth || state.Height < LWindowMinimumHeight {
			state.Width = LWindowWidthDefault
			state.Height = LWindowHeightDefault
		}
	} else {
		width, height := wailsRuntime.WindowGetSize(LContext)
		x, y := wailsRuntime.WindowGetPosition(LContext)
		state.Width = width
		state.Height = height
		state.X = x
		state.Y = y
	}
	_ = LStateWindowSave(state)
}
