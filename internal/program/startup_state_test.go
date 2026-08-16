package program

import (
	"os"
	"path/filepath"
	"testing"
)

func TestWindowGeometryNormalizeRejectsUnusableState(t *testing.T) {
	tests := []struct {
		name  string
		state LStateWindow
	}{
		{name: "too small", state: LStateWindow{Width: 10, Height: 10, X: 20, Y: 20, HasGeometry: true}},
		{name: "negative position", state: LStateWindow{Width: 900, Height: 700, X: -5000, Y: 20, HasGeometry: true}},
		{name: "off right edge", state: LStateWindow{Width: 900, Height: 700, X: 1900, Y: 20, HasGeometry: true}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			normalized := LWindowGeometryNormalize(test.state, 1920, 1080)
			if normalized.HasGeometry {
				t.Fatalf("expected unusable geometry to be centered: %+v", normalized)
			}
			if normalized.Width < LWindowMinimumWidth || normalized.Height < LWindowMinimumHeight {
				t.Fatalf("expected usable dimensions: %+v", normalized)
			}
		})
	}
}

func TestWindowGeometryNormalizeKeepsVisibleState(t *testing.T) {
	state := LStateWindow{Width: 1200, Height: 820, X: 100, Y: 100, HasGeometry: true}
	normalized := LWindowGeometryNormalize(state, 1920, 1080)
	if normalized != state {
		t.Fatalf("visible geometry changed: got %+v want %+v", normalized, state)
	}
}

func TestWindowGeometryNormalizeCentersWhenScreenIsUnavailable(t *testing.T) {
	state := LStateWindow{Width: 9000, Height: 7000, X: 8000, Y: 6000, HasGeometry: true}
	normalized := LWindowGeometryNormalize(state, 0, 0)
	if normalized.HasGeometry {
		t.Fatalf("expected unknown-screen geometry to be centered: %+v", normalized)
	}
	if normalized.Width != LWindowWidthDefault || normalized.Height != LWindowHeightDefault {
		t.Fatalf("got dimensions %dx%d, want defaults %dx%d", normalized.Width, normalized.Height, LWindowWidthDefault, LWindowHeightDefault)
	}
}

func TestStateFileAtomicWriteReplacesCompleteFile(t *testing.T) {
	directory := t.TempDir()
	filePath := filepath.Join(directory, "ui-state.json")
	if err := os.WriteFile(filePath, []byte(`{"old":true}`), 0o644); err != nil {
		t.Fatal(err)
	}
	want := []byte(`{"new":true}`)
	if err := LStateFileAtomicWrite(filePath, want, 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filePath)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != string(want) {
		t.Fatalf("got %q, want %q", got, want)
	}
	temporaryFiles, err := filepath.Glob(filepath.Join(directory, ".ui-state.json.tmp-*"))
	if err != nil {
		t.Fatal(err)
	}
	if len(temporaryFiles) != 0 {
		t.Fatalf("temporary files remain after commit: %v", temporaryFiles)
	}
}

func TestLocaleNormalize(t *testing.T) {
	if got := LLocaleNormalize(" ko\n"); got != "ko" {
		t.Fatalf("got %q, want ko", got)
	}
	if got := LLocaleNormalize("unsupported"); got != "en" {
		t.Fatalf("got %q, want en", got)
	}
}
