package program

import (
	"io/fs"
	"os"
	"path/filepath"
)

// LStateFileAtomicWrite replaces a state file only after the complete new value
// has been written and flushed to a temporary file in the same directory.
func LStateFileAtomicWrite(filePath string, data []byte, permissions fs.FileMode) error {
	directory := filepath.Dir(filePath)
	if err := os.MkdirAll(directory, 0o755); err != nil {
		return err
	}
	temporaryFile, err := os.CreateTemp(directory, "."+filepath.Base(filePath)+".tmp-*")
	if err != nil {
		return err
	}
	temporaryPath := temporaryFile.Name()
	committed := false
	defer func() {
		_ = temporaryFile.Close()
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporaryFile.Chmod(permissions); err != nil {
		return err
	}
	if _, err := temporaryFile.Write(data); err != nil {
		return err
	}
	if err := temporaryFile.Sync(); err != nil {
		return err
	}
	if err := temporaryFile.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, filePath); err != nil {
		return err
	}
	committed = true
	return nil
}
