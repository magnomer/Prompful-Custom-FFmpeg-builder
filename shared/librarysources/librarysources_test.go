package librarysources

import "testing"

func TestResolveExactFfmpegRelease(t *testing.T) {
	source, ok := ResolveLibrarySource("8.1.2", "vvenc")
	if !ok {
		t.Fatal("expected vvenc to resolve for the recorded 8.1.2 release")
	}
	if source.Version != "1.14.0" || source.Sha256 == "" || source.Url == "" {
		t.Fatalf("unexpected vvenc source: %#v", source)
	}
}

func TestResolveFallsBackToHighestReleaseForUnrecordedFfmpeg(t *testing.T) {
	// An unrecorded FFmpeg version must resolve via the highest recorded release, not fail.
	source, ok := ResolveLibrarySource("99.0.0", "lcevc-dec")
	if !ok {
		t.Fatal("expected fallback resolution for an unrecorded FFmpeg version")
	}
	if source.Version == "" {
		t.Fatalf("expected a fallback version, got %#v", source)
	}
	if key := ResolvedFfmpegReleaseKey("99.0.0"); key != "8.1.2" {
		t.Fatalf("expected fallback to the highest recorded release 8.1.2, got %q", key)
	}
}

func TestResolveUnknownLibraryReturnsFalse(t *testing.T) {
	if _, ok := ResolveLibrarySource("8.1.2", "does-not-exist"); ok {
		t.Fatal("expected an unknown library to not resolve")
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		left, right string
		want        int
		decidable   bool
	}{
		{"1.14.0", "1.6.1", 1, true},
		{"1.6.1", "1.6.1", 0, true},
		{"1.5.0", "1.6.1", -1, true},
		{"4.2.0", "4.0.0", 1, true},
		{"master", "1.6.1", 0, false},
		{"1.6.1", "", 0, false},
	}
	for _, testCase := range cases {
		got, decidable := CompareVersions(testCase.left, testCase.right)
		if decidable != testCase.decidable || (decidable && got != testCase.want) {
			t.Fatalf("CompareVersions(%q,%q) = (%d,%v), want (%d,%v)", testCase.left, testCase.right, got, decidable, testCase.want, testCase.decidable)
		}
	}
}
