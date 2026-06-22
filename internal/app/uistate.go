package app

import (
	"os"
	"path/filepath"
)

// uiStateFilePath is where the frontend's saved UI state (selected options,
// libraries, settings, active tab) is persisted between launches. WebView2
// localStorage is not reliably retained across restarts, so the state is kept
// in a file the backend owns instead.
func uiStateFilePath() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDirectory, "PromptfulCustomFFmpegBuilder", "ui-state.json"), nil
}

// LoadUiState returns the saved UI state as an opaque JSON string, or an empty
// string when nothing has been saved yet. The frontend owns the JSON shape.
func (app *App) LoadUiState() (string, error) {
	filePath, err := uiStateFilePath()
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

// SaveUiState writes the frontend's UI state JSON to disk.
func (app *App) SaveUiState(stateJson string) error {
	filePath, err := uiStateFilePath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(filePath, []byte(stateJson), 0o644)
}
