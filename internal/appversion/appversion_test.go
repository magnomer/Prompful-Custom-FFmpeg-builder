package appversion

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestVersionMatchesRepositoryManifests(t *testing.T) {
	t.Parallel()

	root := filepath.Join("..", "..")
	versionData := struct {
		Current string `json:"current-version"`
	}{}
	readJSON(t, filepath.Join(root, "version.json"), &versionData)

	wailsData := struct {
		Version string `json:"version"`
	}{}
	readJSON(t, filepath.Join(root, "wails.json"), &wailsData)

	if Version != versionData.Current {
		t.Fatalf("compiled version %q does not match version.json %q", Version, versionData.Current)
	}
	if Version != wailsData.Version {
		t.Fatalf("compiled version %q does not match wails.json %q", Version, wailsData.Version)
	}
}

func readJSON(t *testing.T, path string, target any) {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(data, target); err != nil {
		t.Fatalf("decode %s: %v", path, err)
	}
}
