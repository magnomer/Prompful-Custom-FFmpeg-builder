package program

import "testing"

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

func TestLocaleNormalize(t *testing.T) {
	if got := LLocaleNormalize(" ko\n"); got != "ko" {
		t.Fatalf("got %q, want ko", got)
	}
	if got := LLocaleNormalize("unsupported"); got != "en" {
		t.Fatalf("got %q, want en", got)
	}
}
