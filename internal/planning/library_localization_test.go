package planning

import (
	"testing"

	"promptfulcustomffmpegbuilder/shared/localization"
)

// TestCatalogLibraryLocalizationCoverage proves the display prose for every catalog
// library lives in the localization files. The Go LibraryChoice intentionally no longer
// carries plain/technical explanation text (the fields are "" by design); the frontend
// reads catalog.libraries.<id>.<field> directly, with the empty Go field only as a
// never-hit fallback. If a library is ever added without its localization keys, this test
// fails instead of silently shipping a blank explanation in the UI.
func TestCatalogLibraryLocalizationCoverage(t *testing.T) {
	profiles := []string{"ucrt64", "mingw64", "clang64"}
	fields := []string{"displayName", "categoryName", "plainExplanation", "technicalExplanation"}
	seen := map[string]bool{}
	for _, profile := range profiles {
		for _, library := range LibraryCatalogForShellProfile(profile) {
			if seen[library.LibraryId] {
				continue
			}
			seen[library.LibraryId] = true
			if library.PlainExplanation != "" || library.TechnicalExplanation != "" {
				t.Errorf("%q: explanation prose must live only in localization, Go fields must stay empty", library.LibraryId)
			}
			for _, field := range fields {
				key := "catalog.libraries." + library.LibraryId + "." + field
				if got := localization.Localize(key, nil); got == "" || got == key {
					t.Errorf("missing localization for library %q (key %s)", library.LibraryId, key)
				}
			}
		}
	}
}
