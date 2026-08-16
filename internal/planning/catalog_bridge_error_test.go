package planning

import "testing"

func TestCatalogResolveDistinguishesEmptySourceFromInvalidSource(t *testing.T) {
	if choices, err := LCatalogLibraryResolve("", "ucrt64"); err != nil || len(choices) != 0 {
		t.Fatalf("empty source should be a legitimate empty catalog: choices=%d err=%v", len(choices), err)
	}
	if _, err := LCatalogLibraryResolve("not-an-ffmpeg-url", "ucrt64"); err == nil {
		t.Fatal("invalid non-empty source should return an error")
	}
	if presets, err := LCatalogPresetResolve("", "ucrt64"); err != nil || len(presets) != 0 {
		t.Fatalf("empty source should be a legitimate empty preset catalog: presets=%d err=%v", len(presets), err)
	}
	if _, err := LCatalogPresetResolve("not-an-ffmpeg-url", "ucrt64"); err == nil {
		t.Fatal("invalid non-empty source should return an error")
	}
}
