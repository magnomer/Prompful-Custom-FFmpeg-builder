package program

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"promptfulcustomffmpegbuilder/internal/workspace"
)

func LDirectoryChildFind(parentDirectory string) (string, error) {
	entries, err := os.ReadDir(parentDirectory)
	if err != nil {
		return "", err
	}
	childDirectories := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			childDirectories = append(childDirectories, filepath.Join(parentDirectory, entry.Name()))
		}
	}
	if len(childDirectories) != 1 {
		return "", fmt.Errorf("expected exactly one extracted source directory, found %d", len(childDirectories))
	}
	return childDirectories[0], nil
}

func LFileCopy(workspaceDirectory string, sourcePath string, destinationPath string) error {
	if err := workspace.LPathRealCheck(workspaceDirectory, sourcePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, filepath.Dir(destinationPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, destinationPath); err != nil {
		return err
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destinationFile, sourceFile)
	closeErr := destinationFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func LFileSizeRead(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func LHashFileCreate(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func LFileExistCheck(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func LTextErrorTrim(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 600 {
		return value[:600] + "..."
	}
	return value
}
