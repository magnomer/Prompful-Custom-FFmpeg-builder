package planning

import "testing"

func stringsContain(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func TestLReleaseVersionResolve(t *testing.T) {
	release, ok := LReleaseVersionResolve("8.1.2")
	if !ok {
		t.Fatalf("expected 8.1.2 to be a supported release")
	}
	if release.Version != "8.1.2" {
		t.Fatalf("version = %q, want 8.1.2", release.Version)
	}
	if release.ArchiveUrl == "" || release.SignatureUrl == "" {
		t.Fatalf("release URLs empty: archive=%q signature=%q", release.ArchiveUrl, release.SignatureUrl)
	}
	if _, ok := LReleaseVersionResolve("0.0.0"); ok {
		t.Fatalf("expected 0.0.0 to be unresolved")
	}
	if _, ok := LReleaseVersionResolve("  "); ok {
		t.Fatalf("expected blank version to be unresolved")
	}
}

func TestLPresetLibraryIdsResolve(t *testing.T) {
	release, ok := LReleaseVersionResolve("8.1.2")
	if !ok {
		t.Fatalf("setup: 8.1.2 must resolve")
	}
	url := release.ArchiveUrl

	normal, ok := LPresetLibraryIdsResolve(url, "", "full", false)
	if !ok {
		t.Fatalf("expected preset 'full' to resolve")
	}
	if !stringsContain(normal, "x264") {
		t.Fatalf("normal 'full' preset missing x264: %v", normal)
	}
	// vvenc is an extended-only library for the full preset.
	if stringsContain(normal, "vvenc") {
		t.Fatalf("normal 'full' unexpectedly contains extended-only vvenc")
	}

	extended, ok := LPresetLibraryIdsResolve(url, "", "full", true)
	if !ok {
		t.Fatalf("expected extended 'full' to resolve")
	}
	if !stringsContain(extended, "vvenc") {
		t.Fatalf("extended 'full' missing vvenc: %v", extended)
	}
	if len(extended) <= len(normal) {
		t.Fatalf("extended set (%d) should be a superset of normal (%d)", len(extended), len(normal))
	}

	if _, ok := LPresetLibraryIdsResolve(url, "", "no-such-preset", false); ok {
		t.Fatalf("expected unknown preset to be unresolved")
	}

	// Returned slice must be an independent copy.
	normal[0] = "mutated"
	fresh, _ := LPresetLibraryIdsResolve(url, "", "full", false)
	if fresh[0] == "mutated" {
		t.Fatalf("resolver returned a shared slice; mutation leaked")
	}
}

func TestLLibraryFlagResolve(t *testing.T) {
	release, ok := LReleaseVersionResolve("8.1.2")
	if !ok {
		t.Fatalf("setup: 8.1.2 must resolve")
	}
	url := release.ArchiveUrl

	id, ok := LLibraryFlagResolve(url, "", "--enable-libx264")
	if !ok {
		t.Fatalf("expected --enable-libx264 to resolve")
	}
	if id != "x264" {
		t.Fatalf("libraryId = %q, want x264", id)
	}

	if _, ok := LLibraryFlagResolve(url, "", "--enable-libnonexistent"); ok {
		t.Fatalf("expected unknown flag to be unresolved")
	}
	if _, ok := LLibraryFlagResolve(url, "", "  "); ok {
		t.Fatalf("expected blank flag to be unresolved")
	}
}
