package program

import (
	"context"
	"encoding/json"
	"fmt"
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
	if state.Width < LWindowMinimumWidth || state.Height < LWindowMinimumHeight {
		return fmt.Errorf("window dimensions %dx%d are below the minimum %dx%d", state.Width, state.Height, LWindowMinimumWidth, LWindowMinimumHeight)
	}
	filePath, err := LPathWindowResolve()
	if err != nil {
		return err
	}
	fileData, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return LStateFileAtomicWrite(filePath, fileData, 0o644)
}

// LWindowGeometryRestore applies the saved size, position, and maximised state
// once the Wails runtime context is available.
func (program *LProgram) LWindowGeometryRestore() {
	if program.LContext == nil || !program.LStateWindowStartup.HasGeometry {
		return
	}
	state := program.LStateWindowStartup
	hasPrimaryScreen := false
	if screens, err := wailsRuntime.ScreenGetAll(program.LContext); err == nil {
		for _, screen := range screens {
			if !screen.IsPrimary {
				continue
			}
			state = LWindowGeometryNormalize(state, screen.Size.Width, screen.Size.Height)
			hasPrimaryScreen = true
			break
		}
	}
	if !hasPrimaryScreen {
		state = LWindowGeometryNormalize(state, 0, 0)
	}
	wailsRuntime.WindowSetSize(program.LContext, state.Width, state.Height)
	if state.HasGeometry {
		wailsRuntime.WindowSetPosition(program.LContext, state.X, state.Y)
	} else {
		wailsRuntime.WindowCenter(program.LContext)
	}
	if state.Maximised {
		wailsRuntime.WindowMaximise(program.LContext)
	}
}

// LWindowGeometrySave records the current window geometry so it can be
// restored on the next launch. Wails cannot report normal bounds while the
// window is maximised, so that case stores safe defaults instead of stale data.
func (program *LProgram) LWindowGeometrySave(LContext context.Context) error {
	if LContext == nil {
		return nil
	}
	isMaximised := wailsRuntime.WindowIsMaximised(LContext)
	state := LStateWindow{Maximised: isMaximised, HasGeometry: true}
	if isMaximised {
		// Wails exposes the maximised rectangle here, not the latest normal
		// rectangle. Persist safe normal bounds instead of falsely reusing the
		// process-startup bounds, which may be stale after a move or resize.
		state.Width = LWindowWidthDefault
		state.Height = LWindowHeightDefault
	} else {
		width, height := wailsRuntime.WindowGetSize(LContext)
		x, y := wailsRuntime.WindowGetPosition(LContext)
		state.Width = width
		state.Height = height
		state.X = x
		state.Y = y
	}
	return LStateWindowSave(state)
}

// LWindowGeometryNormalize keeps restored dimensions usable on the primary
// display. Wails does not expose monitor origins, so coordinates that cannot be
// proven visible are centered rather than applied.
func LWindowGeometryNormalize(state LStateWindow, screenWidth int, screenHeight int) LStateWindow {
	if state.Width < LWindowMinimumWidth || state.Height < LWindowMinimumHeight {
		state.Width = LWindowWidthDefault
		state.Height = LWindowHeightDefault
		state.HasGeometry = false
	}
	if screenWidth <= 0 || screenHeight <= 0 {
		state.Width = LWindowWidthDefault
		state.Height = LWindowHeightDefault
		state.HasGeometry = false
		return state
	}
	if state.Width > screenWidth {
		state.Width = screenWidth
	}
	if state.Height > screenHeight {
		state.Height = screenHeight
	}
	const visibleEdge = 80
	if state.X < 0 || state.Y < 0 || state.X+visibleEdge > screenWidth || state.Y+visibleEdge > screenHeight {
		state.HasGeometry = false
	}
	return state
}
