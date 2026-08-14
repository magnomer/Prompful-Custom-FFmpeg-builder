package planning

import "testing"

// LCatalogTestLoad loads the embedded catalog resolver or fails the test.
func LCatalogTestLoad(t *testing.T) LCatalogResolver {
	t.Helper()
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		t.Fatalf("load embedded catalog resolver: %v", err)
	}
	return resolver
}

// LLibraryFlagsGet resolves one library for an FFmpeg version and returns
// its configure flags, bypassing selection conflict normalization so a removed
// flag is observed directly on the library record.
func LLibraryFlagsGet(t *testing.T, resolver LCatalogResolver, ffmpegVersion string, libraryId string) []string {
	t.Helper()
	resolvedLibrary, err := resolver.LLibraryResolve(ffmpegVersion, libraryId, LShellProfileDefault)
	if err != nil {
		t.Fatalf("resolve library %q for FFmpeg %q: %v", libraryId, ffmpegVersion, err)
	}
	return resolvedLibrary.ConfigureFlags
}

// Section 1 — release resolution.
func TestLRelease901Supported(t *testing.T) {
	var choice901 LReleaseChoice
	found := false
	for _, choice := range LReleaseSupportedGet() {
		if choice.Version == "9.0.1" {
			choice901 = choice
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("LReleaseSupportedGet() does not include 9.0.1")
	}
	if choice901.Codename != "Lei" {
		t.Fatalf("9.0.1 codename = %q, want Lei", choice901.Codename)
	}
	wantArchive := "https://www.ffmpeg.org/releases/ffmpeg-9.0.1.tar.xz"
	if choice901.ArchiveUrl != wantArchive {
		t.Fatalf("9.0.1 archive = %q, want %q", choice901.ArchiveUrl, wantArchive)
	}
	if choice901.SignatureUrl != wantArchive+".asc" {
		t.Fatalf("9.0.1 signature = %q, want archive + .asc", choice901.SignatureUrl)
	}
}

func TestLRelease901OrdersFirst(t *testing.T) {
	recommended := LReleaseRecommendList()
	if len(recommended) == 0 {
		t.Fatalf("LReleaseRecommendList() is empty")
	}
	if recommended[0] != "9.0.1" {
		t.Fatalf("newest release = %q, want 9.0.1 ordered first", recommended[0])
	}
	for _, version := range recommended[1:] {
		comparison, ok := LVersionCompare("9.0.1", version)
		if !ok {
			t.Fatalf("LVersionCompare(9.0.1, %q) not comparable", version)
		}
		if comparison <= 0 {
			t.Fatalf("LVersionCompare(9.0.1, %q) = %d, want 9.0.1 newer", version, comparison)
		}
	}
	if highest := LReleaseHighestGet(); highest != "9.0.1" {
		t.Fatalf("LReleaseHighestGet() = %q, want 9.0.1", highest)
	}
}

// Section 2 — library catalog completeness.
func TestLLibraryCatalog901Complete(t *testing.T) {
	resolver := LCatalogTestLoad(t)
	if len(resolver.LibraryRecords) == 0 {
		t.Fatalf("no library records loaded")
	}
	for libraryId, libraryRecord := range resolver.LibraryRecords {
		versionObject, exists := LVersionRecordRead(libraryRecord, "9.0.1")
		if !exists {
			t.Errorf("library %q has no 9.0.1 block (exact-match resolver has no fallback)", libraryId)
			continue
		}
		if got := LCatalogFieldGet(versionObject, "ffmpegVersion"); got != "9.0.1" {
			t.Errorf("library %q 9.0.1 block ffmpegVersion = %q, want 9.0.1", libraryId, got)
		}
		if got := LCatalogFieldGet(versionObject, "libraryId"); got != libraryId {
			t.Errorf("library %q 9.0.1 block libraryId = %q, want %q", libraryId, got, libraryId)
		}
	}
}

// Section 3 — preset completeness.
func TestLPreset901Complete(t *testing.T) {
	resolver := LCatalogTestLoad(t)
	if len(resolver.PresetRecords) == 0 {
		t.Fatalf("no preset records loaded")
	}
	for presetId, presetRecord := range resolver.PresetRecords {
		versionObject, exists := LPresetRecordRead(presetRecord, "9.0.1")
		if !exists {
			t.Errorf("preset %q has no 9.0.1 block", presetId)
			continue
		}
		if len(LArrayFieldGet(versionObject, "libraryIds")) == 0 {
			t.Errorf("preset %q 9.0.1 block has no libraryIds", presetId)
		}
		modeNames := []string{"normal"}
		if len(LArrayFieldGet(versionObject, "extendedLibraryIds")) > 0 {
			modeNames = append(modeNames, "extended")
		}
		for _, modeName := range modeNames {
			resolvedPlan, err := resolver.LVersionResolve(LCatalogResolutionSettings{
				FfmpegVersion:  "9.0.1",
				PresetId:       presetId,
				PresetModeName: modeName,
			})
			if err != nil {
				t.Errorf("resolve preset %q (%s) for 9.0.1: %v", presetId, modeName, err)
				continue
			}
			if len(resolvedPlan.NormalizedLibraryIds) == 0 {
				t.Errorf("preset %q (%s) resolved no 9.0.1 libraries", presetId, modeName)
			}
			for _, warning := range resolvedPlan.Warnings {
				t.Errorf("preset %q (%s) for 9.0.1 dropped a library: %s", presetId, modeName, warning.Message)
			}
		}
	}
}

// Section 4 — removed configure flags (shaderc / glslang).
func TestLConfigureFlags901ShadercGlslangRemoved(t *testing.T) {
	resolver := LCatalogTestLoad(t)

	flags901 := append(
		LLibraryFlagsGet(t, resolver, "9.0.1", "shaderc"),
		LLibraryFlagsGet(t, resolver, "9.0.1", "glslang")...,
	)
	if LStringsContainCheck(flags901, "--enable-libshaderc") {
		t.Fatalf("9.0.1 flags contain removed --enable-libshaderc: %v", flags901)
	}
	if LStringsContainCheck(flags901, "--enable-libglslang") {
		t.Fatalf("9.0.1 flags contain removed --enable-libglslang: %v", flags901)
	}

	flags812 := append(
		LLibraryFlagsGet(t, resolver, "8.1.2", "shaderc"),
		LLibraryFlagsGet(t, resolver, "8.1.2", "glslang")...,
	)
	if !LStringsContainCheck(flags812, "--enable-libshaderc") {
		t.Fatalf("8.1.2 flags missing --enable-libshaderc (regression): %v", flags812)
	}
	if !LStringsContainCheck(flags812, "--enable-libglslang") {
		t.Fatalf("8.1.2 flags missing --enable-libglslang (regression): %v", flags812)
	}
}

// Section 5 — ONNX Runtime flag emission and shell-profile availability.
func TestLOnnxRuntime901Flag(t *testing.T) {
	resolver := LCatalogTestLoad(t)

	flags901 := LLibraryFlagsGet(t, resolver, "9.0.1", "onnxruntime")
	if !LStringsContainCheck(flags901, "--enable-libonnxruntime") {
		t.Fatalf("9.0.1 onnxruntime missing --enable-libonnxruntime: %v", flags901)
	}
	flags812 := LLibraryFlagsGet(t, resolver, "8.1.2", "onnxruntime")
	if LStringsContainCheck(flags812, "--enable-libonnxruntime") {
		t.Fatalf("8.1.2 onnxruntime unexpectedly emits --enable-libonnxruntime: %v", flags812)
	}
}

func TestLOnnxRuntime901ShellProfiles(t *testing.T) {
	resolver := LCatalogTestLoad(t)
	for _, shellProfileName := range []string{"ucrt64", "clang64"} {
		resolvedLibrary, err := resolver.LLibraryResolve("9.0.1", "onnxruntime", shellProfileName)
		if err != nil {
			t.Fatalf("resolve onnxruntime for %q: %v", shellProfileName, err)
		}
		if resolvedLibrary.SupportState == LLibrarySupportUnavailable {
			t.Fatalf("onnxruntime should be available for %q on 9.0.1", shellProfileName)
		}
		if len(resolvedLibrary.PackageNames) == 0 {
			t.Fatalf("onnxruntime resolved no package for %q on 9.0.1", shellProfileName)
		}
	}
	unavailable, err := resolver.LLibraryResolve("9.0.1", "onnxruntime", "mingw64")
	if err != nil {
		t.Fatalf("resolve onnxruntime for mingw64: %v", err)
	}
	if unavailable.SupportState != LLibrarySupportUnavailable {
		t.Fatalf("onnxruntime mingw64 state = %q, want unavailable", unavailable.SupportState)
	}
	if len(unavailable.PackageNames) != 0 {
		t.Fatalf("onnxruntime mingw64 should expose no package, got %v", unavailable.PackageNames)
	}
}

// Section 6 — libplacebo minimum pkg-config version.
func TestLLibplacebo901Minimum(t *testing.T) {
	resolver := LCatalogTestLoad(t)
	cases := map[string]string{"9.0.1": "7.351.0", "8.1.2": "5.229.0"}
	for ffmpegVersion, wantMinimum := range cases {
		resolvedLibrary, err := resolver.LLibraryResolve(ffmpegVersion, "libplacebo", LShellProfileDefault)
		if err != nil {
			t.Fatalf("resolve libplacebo for %q: %v", ffmpegVersion, err)
		}
		if resolvedLibrary.VersionCompatibility == nil {
			t.Fatalf("libplacebo %q has no version compatibility", ffmpegVersion)
		}
		if got := resolvedLibrary.VersionCompatibility.MinVersion; got != wantMinimum {
			t.Fatalf("libplacebo %q minimum = %q, want %q", ffmpegVersion, got, wantMinimum)
		}
	}
}

// Section 8 — work registry.
func TestLWorkRegistry901Resolves(t *testing.T) {
	registry, err := LWorkRegistryLoad()
	if err != nil {
		t.Fatalf("load work registry with 9.0.1 catalog: %v", err)
	}
	if _, exists := registry.LWorkLibraryResolve("9.0.1", "davs2"); !exists {
		t.Fatalf("work registry did not resolve 9.0.1 davs2 preparation work")
	}
}
