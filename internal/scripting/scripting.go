package scripting

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"promptfulcustomffmpegbuilder/internal/workspace"
)

type LPlanScript struct {
	WorkspaceDirectory string   `json:"workspaceDirectory"`
	ScriptFilePath     string   `json:"scriptFilePath"`
	ScriptLines        []string `json:"scriptLines"`
}

type LScriptFile struct {
	ScriptFilePath   string `json:"scriptFilePath"`
	ScriptSha256Hash string `json:"scriptSha256Hash"`
}

func LScriptFileWrite(scriptFilePlan LPlanScript) (LScriptFile, error) {
	if scriptFilePlan.WorkspaceDirectory == "" || scriptFilePlan.ScriptFilePath == "" {
		return LScriptFile{}, errors.New("approved shell script paths must not be empty")
	}
	if err := workspace.LPathWorkspaceCheck(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return LScriptFile{}, err
	}
	if err := workspace.LPathRealCheck(scriptFilePlan.WorkspaceDirectory, filepath.Dir(scriptFilePlan.ScriptFilePath)); err != nil {
		return LScriptFile{}, err
	}
	if len(scriptFilePlan.ScriptLines) == 0 {
		return LScriptFile{}, errors.New("approved shell script has no lines")
	}
	scriptDirectory := filepath.Dir(scriptFilePlan.ScriptFilePath)
	if err := os.MkdirAll(scriptDirectory, 0o755); err != nil {
		return LScriptFile{}, err
	}
	if err := workspace.LPathRealCheck(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return LScriptFile{}, err
	}
	scriptText := strings.Join(scriptFilePlan.ScriptLines, "\n") + "\n"
	scriptHash := sha256.Sum256([]byte(scriptText))
	scriptHashString := hex.EncodeToString(scriptHash[:])
	if existingInfo, err := os.Lstat(scriptFilePlan.ScriptFilePath); err == nil {
		if existingInfo.Mode()&os.ModeSymlink != 0 {
			return LScriptFile{}, errors.New("approved shell script path is a symlink")
		}
		if existingInfo.IsDir() {
			return LScriptFile{}, errors.New("approved shell script path is a directory")
		}
		existingBytes, readErr := os.ReadFile(scriptFilePlan.ScriptFilePath)
		if readErr != nil {
			return LScriptFile{}, readErr
		}
		existingHash := sha256.Sum256(existingBytes)
		if strings.EqualFold(hex.EncodeToString(existingHash[:]), scriptHashString) {
			return LScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
		}
		if removeErr := os.Remove(scriptFilePlan.ScriptFilePath); removeErr != nil {
			return LScriptFile{}, fmt.Errorf("remove stale approved shell script: %w", removeErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return LScriptFile{}, err
	}
	if err := workspace.LPathRealCheck(scriptFilePlan.WorkspaceDirectory, scriptDirectory); err != nil {
		return LScriptFile{}, err
	}
	outputFile, err := os.OpenFile(scriptFilePlan.ScriptFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return LScriptFile{}, err
	}
	_, writeErr := outputFile.Write([]byte(scriptText))
	closeErr := outputFile.Close()
	if writeErr != nil {
		return LScriptFile{}, writeErr
	}
	if closeErr != nil {
		return LScriptFile{}, closeErr
	}
	return LScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
}

// LShellTextQuote wraps a value in single quotes for safe embedding in a generated shell
// script, escaping any embedded single quote. Shared by every script generator in this package.
func LShellTextQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
