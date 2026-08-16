package program

import (
	"os"
	"path/filepath"
)

// LPathStateResolve is where the frontend's saved UI state (selected options,
// libraries, settings, active tab) is persisted between launches. WebView2
// localStorage is not reliably retained across restarts, so the state is kept
// in a file the backend owns instead.
func LPathStateResolve() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDirectory, "PromptfulCustomFfmpegBuilder", "ui-state.json"), nil
}

// LStateUiLoad returns the saved UI state as an opaque JSON string, or an empty
// string when nothing has been saved yet. The frontend owns the JSON shape.
func (program *LProgram) LStateUiLoad() (string, error) {
	filePath, err := LPathStateResolve()
	if err != nil {
		return "", err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	return string(fileData), nil
}

// LStateUiSave writes the frontend's UI state JSON to disk.
func (program *LProgram) LStateUiSave(stateJson string) error {
	filePath, err := LPathStateResolve()
	if err != nil {
		return err
	}
	return LStateFileAtomicWrite(filePath, []byte(stateJson), 0o644)
}
