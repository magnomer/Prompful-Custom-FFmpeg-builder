package planning

import (
	"testing"

	"promptfulcustomffmpegbuilder/shared/releasesupport"
)

func planWarningKeys(plan FfmpegBuildPlan) map[string]bool {
	keys := map[string]bool{}
	for _, warning := range plan.Warnings {
		keys[warning.MessageKey] = true
	}
	return keys
}

func validVersionedSettings(version string, libraryIds []string) FfmpegBuildSettings {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-" + version + ".tar.xz"
	settings.FfmpegSourceSignatureUrl = settings.FfmpegSourceArchiveUrl + ".asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = libraryIds
	return settings
}

// knownUnreleasedCatalogLibraries are catalog rows whose --enable switch exists only in
// FFmpeg master, so no released (and therefore no manifested) version supports them. They
// are intentionally absent from every release manifest; the planner blocks selecting them.
var knownUnreleasedCatalogLibraries = map[string]bool{
	"onnxruntime": true, // --enable-libonnxruntime is master-only as of FFmpeg 8.1
}

func TestReleaseSupportManifestLookup(t *testing.T) {
	for _, manifestedLine := range []string{"8.1", "8.0", "7.1", "7.0", "6.1", "5.1", "4.4"} {
		if _, found := releasesupport.ForReleaseLine(manifestedLine); !found {
			t.Errorf("expected a %s release-support manifest", manifestedLine)
		}
	}
	if _, found := releasesupport.ForReleaseLine("3.4"); found {
		t.Fatalf("did not expect a manifest for an unsupported old line (3.4)")
	}
}

// TestReleaseLineLibraryDeltas pins the configure-verified per-line support boundaries.
func TestReleaseLineLibraryDeltas(t *testing.T) {
	support81, _ := releasesupport.ForReleaseLine("8.1")
	support80, _ := releasesupport.ForReleaseLine("8.0")
	support71, _ := releasesupport.ForReleaseLine("7.1")
	supported := func(r releasesupport.ReleaseSupport, id string) bool {
		_, ok := r.LibrarySupportFor(id)
		return ok
	}
	// 8.1-only (verified absent from n8.0 / n7.1 configure).
	for _, id := range []string{"cairo", "opencolorio", "svtjpegxs", "mpeghdec", "xeveb", "xevdb"} {
		if !supported(support81, id) {
			t.Errorf("8.1 must support %q", id)
		}
		if supported(support80, id) {
			t.Errorf("8.0 must not support %q", id)
		}
		if supported(support71, id) {
			t.Errorf("7.1 must not support %q", id)
		}
	}
	// Present in 8.0 but dropped in 7.1 (oapv, whisper added in 8.0).
	for _, id := range []string{"oapv", "whisper"} {
		if !supported(support80, id) {
			t.Errorf("8.0 must support %q", id)
		}
		if supported(support71, id) {
			t.Errorf("7.1 must not support %q", id)
		}
	}
	// Present in 7.1 but dropped in 7.0 (lc3, lcevc-dec, vvenc added in 7.1).
	support70, _ := releasesupport.ForReleaseLine("7.0")
	for _, id := range []string{"lc3", "lcevc-dec", "vvenc"} {
		if !supported(support71, id) {
			t.Errorf("7.1 must support %q", id)
		}
		if supported(support70, id) {
			t.Errorf("7.0 must not support %q", id)
		}
	}
	// EVC (xeve/xevd) and oneVPL (libvpl) reach back to 7.0.
	for _, id := range []string{"xeve", "xevd", "libvpl"} {
		if !supported(support70, id) {
			t.Errorf("7.0 must support %q", id)
		}
	}
	// Present in 7.0 but dropped in 6.1 (EVC, dvd*, qrencode/quirc, torch added in 7.0).
	support61, _ := releasesupport.ForReleaseLine("6.1")
	for _, id := range []string{"xeve", "xevd", "dvdnav", "dvdread", "qrencode", "quirc", "torch"} {
		if !supported(support70, id) {
			t.Errorf("7.0 must support %q", id)
		}
		if supported(support61, id) {
			t.Errorf("6.1 must not support %q", id)
		}
	}
	// Long-standing libraries remain in 6.1.
	for _, id := range []string{"x264", "x265", "dav1d", "aom", "svt-av1", "libvpx"} {
		if !supported(support61, id) {
			t.Errorf("6.1 must support %q", id)
		}
	}
	// Present in 6.1 but dropped in 5.1 (aribcaption, harfbuzz added 6.1; oneVPL/libvpl 6.0+).
	support51, _ := releasesupport.ForReleaseLine("5.1")
	for _, id := range []string{"aribcaption", "harfbuzz", "libvpl"} {
		if !supported(support61, id) {
			t.Errorf("6.1 must support %q", id)
		}
		if supported(support51, id) {
			t.Errorf("5.1 must not support %q", id)
		}
	}
	// Present in 5.1 but dropped in 4.4 (lcms2, libjxl, libplacebo, shaderc).
	support44, _ := releasesupport.ForReleaseLine("4.4")
	for _, id := range []string{"lcms2", "libjxl", "libplacebo", "shaderc"} {
		if !supported(support51, id) {
			t.Errorf("5.1 must support %q", id)
		}
		if supported(support44, id) {
			t.Errorf("4.4 must not support %q", id)
		}
	}
	// Configure-verified against the doc, which wrongly listed these as removed in 4.4.
	for _, id := range []string{"openvino", "uavs3d", "glslang"} {
		if !supported(support44, id) {
			t.Errorf("4.4 must support %q (present in n4.4 configure; doc was wrong)", id)
		}
	}
}

// TestManifestMinVersionsAreConfigureVerified pins the per-line minimum-version divergences
// extracted from each tag's configure (the comparison doc was wrong about these).
func TestManifestMinVersionsAreConfigureVerified(t *testing.T) {
	cases := []struct {
		line, id, wantMin string
	}{
		{"8.1", "dav1d", "1.0.0"}, {"8.0", "dav1d", "0.5.0"}, {"6.1", "dav1d", "0.5.0"},
		{"8.1", "aom", "2.0.0"}, {"7.0", "aom", "1.0.0"}, {"6.1", "aom", "1.0.0"},
		{"8.0", "libplacebo", "5.229.0"}, {"7.1", "libplacebo", "4.192.0"},
		{"8.1", "lcevc-dec", "4.0.0"}, {"8.0", "lcevc-dec", "2.0.0"},
		{"8.1", "xeve", "0.5.1"}, {"7.0", "xeve", "0.4.3"},
		{"8.1", "kvazaar", "2.0.0"}, {"5.1", "kvazaar", "0.8.1"},
		{"8.1", "avisynthplus", "3.7.3"}, {"5.1", "avisynthplus", "3.7.1"},
		{"8.1", "vmaf", "2.0.0"}, {"4.4", "vmaf", "1.5.2"},
		{"8.1", "svt-av1", "0.9.0"}, {"4.4", "svt-av1", "0.8.4"},
		{"4.4", "avisynthplus", ""}, // n4.4 checks avisynth headers only, no version floor
	}
	for _, c := range cases {
		release, _ := releasesupport.ForReleaseLine(c.line)
		support, supported := release.LibrarySupportFor(c.id)
		if !supported {
			t.Errorf("%s must support %q", c.line, c.id)
			continue
		}
		if support.MinVersion != c.wantMin {
			t.Errorf("%s %s minVersion = %q, want %q", c.line, c.id, support.MinVersion, c.wantMin)
		}
	}
}

func TestReleaseLineKeyHelper(t *testing.T) {
	if got := releasesupport.ReleaseLineKey("8.1.2"); got != "8.1" {
		t.Fatalf("ReleaseLineKey(\"8.1.2\") = %q, want \"8.1\"", got)
	}
	if got := releasesupport.ReleaseLineKey("snapshot"); got != "" {
		t.Fatalf("ReleaseLineKey(snapshot) = %q, want \"\"", got)
	}
}

// lensfun is present in every manifested line (FFmpeg has the switch) but marked unavailable
// (the package this builder can supply lacks the lf_db_create symbol FFmpeg gates on), so it
// is supported-but-not-available on every release.
func TestLensfunSupportedButUnavailableEveryLine(t *testing.T) {
	for _, line := range []string{"8.1", "8.0", "7.1", "7.0", "6.1", "5.1", "4.4"} {
		release, found := releasesupport.ForReleaseLine(line)
		if !found {
			t.Fatalf("expected a %s manifest", line)
		}
		if _, supported := release.LibrarySupportFor("lensfun"); !supported {
			t.Errorf("%s: lensfun should stay present/supported in the manifest", line)
		}
		if release.LibraryAvailableFor("lensfun") {
			t.Errorf("%s: lensfun must be unavailable", line)
		}
		if !release.LibraryAvailableFor("x264") {
			t.Errorf("%s: x264 must remain available", line)
		}
	}
}

// A manifest-unavailable library blocks the plan with the package-unavailable key (distinct
// from the unsupported key) and is annotated Supported-but-not-Available for the UI.
func TestPlanBlocksManifestUnavailableLibrary(t *testing.T) {
	settings := validVersionedSettings("8.1.2", []string{"lensfun"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected lensfun (manifest-unavailable) to block the plan on 8.1.2")
	}
	if !planWarningKeys(plan)["plan.warnings.libraryPackageUnavailableForVersion"] {
		t.Fatalf("expected libraryPackageUnavailableForVersion warning, got %#v", plan.Warnings)
	}
	for _, library := range plan.SelectedLibraries {
		if library.LibraryId != "lensfun" {
			continue
		}
		if library.VersionCompatibility == nil || !library.VersionCompatibility.Supported || library.VersionCompatibility.Available {
			t.Fatalf("expected lensfun annotated supported-but-unavailable, got %#v", library.VersionCompatibility)
		}
		return
	}
	t.Fatalf("lensfun not found in selected libraries")
}

// A future FFmpeg version the program does not yet record falls back to the latest recorded
// line (8.1). lensfun, unavailable on 8.1, therefore stays blocked on a future version through
// the version mechanism (not a global rule); a normal library stays buildable.
func TestFutureVersionFallsBackToLatestRecordedLine(t *testing.T) {
	blockedPlan, err := PlanFfmpegBuild(validVersionedSettings("8.2.0", []string{"lensfun"}))
	if err != nil {
		t.Fatal(err)
	}
	if blockedPlan.IsExecutable {
		t.Fatalf("expected lensfun to stay blocked on a future FFmpeg via the latest recorded line")
	}
	if !planWarningKeys(blockedPlan)["plan.warnings.libraryPackageUnavailableForVersion"] {
		t.Fatalf("expected libraryPackageUnavailableForVersion on 8.2.0, got %#v", blockedPlan.Warnings)
	}
	okPlan, err := PlanFfmpegBuild(validVersionedSettings("9.0.0", []string{"x264"}))
	if err != nil {
		t.Fatal(err)
	}
	if !okPlan.IsExecutable {
		t.Fatalf("expected x264 to stay buildable on a future FFmpeg, warnings: %#v", okPlan.Warnings)
	}
}

// ResolveReleaseSupport: future -> latest recorded; gap/old unlisted -> not resolved (no gating).
func TestResolveReleaseSupportFallback(t *testing.T) {
	future, resolved := releasesupport.ResolveReleaseSupport("8.2.0")
	if !resolved {
		t.Fatalf("future version 8.2.0 should resolve to the latest recorded line")
	}
	if future.LibraryAvailableFor("lensfun") {
		t.Fatalf("future version must inherit latest line's lensfun unavailability")
	}
	if _, ok := releasesupport.ResolveReleaseSupport("6.0.0"); ok {
		t.Fatalf("a gap line below the newest recorded line must not resolve (no gating)")
	}
	if _, ok := releasesupport.ResolveReleaseSupport("snapshot"); ok {
		t.Fatalf("a snapshot version must not resolve")
	}
}

func TestMasterOnlyLibraryBlockedOnReleasedVersions(t *testing.T) {
	settings := validVersionedSettings("8.1.2", []string{"onnxruntime"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected onnxruntime (master-only switch) to be blocked on FFmpeg 8.1.2")
	}
	if !planWarningKeys(plan)["plan.warnings.libraryUnsupportedForVersion"] {
		t.Fatalf("expected libraryUnsupportedForVersion for onnxruntime, got %#v", plan.Warnings)
	}
}

func libraryByIdInPlan(libraries []LibraryChoice, libraryId string) (LibraryChoice, bool) {
	for _, library := range libraries {
		if library.LibraryId == libraryId {
			return library, true
		}
	}
	return LibraryChoice{}, false
}

func TestVmafSourceBuiltOnFourFourNativeElsewhere(t *testing.T) {
	// FFmpeg 4.4 probes the libvmaf 1.x compute_vmaf API that the MSYS2 3.x package cannot
	// satisfy, so the manifest marks vmaf sourceBuild and the planner flips it to an Internal
	// source build (pinned libvmaf 1.5.2) with its native package dropped.
	plan44, err := PlanFfmpegBuild(validVersionedSettings("4.4.8", []string{"vmaf"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 4.4.8: %v", err)
	}
	vmaf44, found := libraryByIdInPlan(plan44.SelectedLibraries, "vmaf")
	if !found {
		t.Fatal("vmaf missing from 4.4.8 selected libraries")
	}
	if vmaf44.TrackName != LibraryTrackInternal {
		t.Fatalf("expected vmaf on Internal track for 4.4.8, got %q", vmaf44.TrackName)
	}
	if len(vmaf44.PackageNames) != 0 {
		t.Fatalf("expected vmaf native package dropped for 4.4.8, got %v", vmaf44.PackageNames)
	}
	var vmafPrep *LibraryPreparation
	for index := range plan44.LibraryPreparations {
		if plan44.LibraryPreparations[index].LibraryId == "vmaf" {
			vmafPrep = &plan44.LibraryPreparations[index]
		}
	}
	if vmafPrep == nil {
		t.Fatal("expected a vmaf preparation recipe for 4.4.8")
	}
	if vmafPrep.Version != "1.5.2" {
		t.Fatalf("expected libvmaf 1.5.2 source pin, got %q", vmafPrep.Version)
	}
	if vmafPrep.BuildSystem != BuildSystemMeson {
		t.Fatalf("expected meson build system for vmaf, got %q", vmafPrep.BuildSystem)
	}
	if !plan44.IsExecutable {
		t.Fatalf("4.4.8 vmaf plan should be executable, warnings: %#v", plan44.Warnings)
	}

	// FFmpeg 5.1 probes the libvmaf 2.x vmaf_init API the native 3.x package provides, so vmaf
	// stays Native and is installed as an MSYS2 package, never source-built.
	plan51, err := PlanFfmpegBuild(validVersionedSettings("5.1.9", []string{"vmaf"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 5.1.9: %v", err)
	}
	vmaf51, found := libraryByIdInPlan(plan51.SelectedLibraries, "vmaf")
	if !found {
		t.Fatal("vmaf missing from 5.1.9 selected libraries")
	}
	if vmaf51.TrackName != LibraryTrackNative {
		t.Fatalf("expected vmaf on Native track for 5.1.9, got %q", vmaf51.TrackName)
	}
	if len(vmaf51.PackageNames) == 0 {
		t.Fatal("expected vmaf to keep its native MSYS2 package on 5.1.9")
	}
	for _, prep := range plan51.LibraryPreparations {
		if prep.LibraryId == "vmaf" {
			t.Fatal("vmaf must not be source-built on 5.1.9")
		}
	}
}

func TestLibplaceboSourceBuiltOnFiveOneNativeElsewhere(t *testing.T) {
	// FFmpeg 5.1's vf_libplacebo.c references pl_peak_detect_params.overshoot_margin and
	// pl_render_params.force_icc_lut unconditionally; libplacebo removed those fields in its
	// current (API 7.x) release, so the MSYS2 native package makes 5.1 fail to compile
	// libavfilter/vf_libplacebo.c. The 5.1 line marks libplacebo sourceBuild and the planner
	// flips it to an Internal source build (pinned libplacebo 4.192.0, API 192, both fields
	// present) with its native package dropped.
	plan51, err := PlanFfmpegBuild(validVersionedSettings("5.1.9", []string{"libplacebo"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 5.1.9: %v", err)
	}
	libplacebo51, found := libraryByIdInPlan(plan51.SelectedLibraries, "libplacebo")
	if !found {
		t.Fatal("libplacebo missing from 5.1.9 selected libraries")
	}
	if libplacebo51.TrackName != LibraryTrackInternal {
		t.Fatalf("expected libplacebo on Internal track for 5.1.9, got %q", libplacebo51.TrackName)
	}
	if len(libplacebo51.PackageNames) != 0 {
		t.Fatalf("expected libplacebo native packages dropped for 5.1.9, got %v", libplacebo51.PackageNames)
	}
	var libplaceboPrep *LibraryPreparation
	for index := range plan51.LibraryPreparations {
		if plan51.LibraryPreparations[index].LibraryId == "libplacebo" {
			libplaceboPrep = &plan51.LibraryPreparations[index]
		}
	}
	if libplaceboPrep == nil {
		t.Fatal("expected a libplacebo preparation recipe for 5.1.9")
	}
	if libplaceboPrep.Version != "4.192.0" {
		t.Fatalf("expected libplacebo 4.192.0 source pin, got %q", libplaceboPrep.Version)
	}
	if libplaceboPrep.BuildSystem != BuildSystemMeson {
		t.Fatalf("expected meson build system for libplacebo, got %q", libplaceboPrep.BuildSystem)
	}
	if !plan51.IsExecutable {
		t.Fatalf("5.1.9 libplacebo plan should be executable, warnings: %#v", plan51.Warnings)
	}

	// FFmpeg 6.1+ guards both fields behind PL_API_VER, so libplacebo stays Native (the current
	// MSYS2 package compiles fine) and is never source-built.
	plan71, err := PlanFfmpegBuild(validVersionedSettings("7.1.5", []string{"libplacebo"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 7.1.5: %v", err)
	}
	libplacebo71, found := libraryByIdInPlan(plan71.SelectedLibraries, "libplacebo")
	if !found {
		t.Fatal("libplacebo missing from 7.1.5 selected libraries")
	}
	if libplacebo71.TrackName != LibraryTrackNative {
		t.Fatalf("expected libplacebo on Native track for 7.1.5, got %q", libplacebo71.TrackName)
	}
	if len(libplacebo71.PackageNames) == 0 {
		t.Fatal("expected libplacebo to keep its native MSYS2 packages on 7.1.5")
	}
	for _, prep := range plan71.LibraryPreparations {
		if prep.LibraryId == "libplacebo" {
			t.Fatal("libplacebo must not be source-built on 7.1.5")
		}
	}
}

func TestXeveSourceBuiltOnSevenZeroNativeElsewhere(t *testing.T) {
	// FFmpeg 7.0's libavcodec/libxeve.c assigns a scalar to param->fps, but XEVE 0.5 made fps an
	// XEVE_RATIONAL struct, so the MSYS2 native XEVE (0.5.x) makes 7.0 fail to compile libxeve.c
	// ("incompatible types ... 'XEVE_RATIONAL' ... 'long int'"). The 7.0 line marks xeve
	// sourceBuild and the planner flips it to an Internal source build (pinned XEVE 0.4.3, scalar
	// fps) with its native package dropped.
	plan70, err := PlanFfmpegBuild(validVersionedSettings("7.0.3", []string{"xeve"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 7.0.3: %v", err)
	}
	xeve70, found := libraryByIdInPlan(plan70.SelectedLibraries, "xeve")
	if !found {
		t.Fatal("xeve missing from 7.0.3 selected libraries")
	}
	if xeve70.TrackName != LibraryTrackInternal {
		t.Fatalf("expected xeve on Internal track for 7.0.3, got %q", xeve70.TrackName)
	}
	if len(xeve70.PackageNames) != 0 {
		t.Fatalf("expected xeve native package dropped for 7.0.3, got %v", xeve70.PackageNames)
	}
	var xevePrep *LibraryPreparation
	for index := range plan70.LibraryPreparations {
		if plan70.LibraryPreparations[index].LibraryId == "xeve" {
			xevePrep = &plan70.LibraryPreparations[index]
		}
	}
	if xevePrep == nil {
		t.Fatal("expected a xeve preparation recipe for 7.0.3")
	}
	if xevePrep.Version != "0.4.3" {
		t.Fatalf("expected XEVE 0.4.3 source pin, got %q", xevePrep.Version)
	}
	if xevePrep.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected cmake build system for xeve, got %q", xevePrep.BuildSystem)
	}
	if !plan70.IsExecutable {
		t.Fatalf("7.0.3 xeve plan should be executable, warnings: %#v", plan70.Warnings)
	}

	// FFmpeg 7.1+ adapted libxeve.c to the XEVE 0.5 API the native package provides, so xeve
	// stays Native (the MSYS2 package compiles fine) and is never source-built.
	plan71, err := PlanFfmpegBuild(validVersionedSettings("7.1.5", []string{"xeve"}))
	if err != nil {
		t.Fatalf("PlanFfmpegBuild 7.1.5: %v", err)
	}
	xeve71, found := libraryByIdInPlan(plan71.SelectedLibraries, "xeve")
	if !found {
		t.Fatal("xeve missing from 7.1.5 selected libraries")
	}
	if xeve71.TrackName != LibraryTrackNative {
		t.Fatalf("expected xeve on Native track for 7.1.5, got %q", xeve71.TrackName)
	}
	if len(xeve71.PackageNames) == 0 {
		t.Fatal("expected xeve to keep its native MSYS2 package on 7.1.5")
	}
	for _, prep := range plan71.LibraryPreparations {
		if prep.LibraryId == "xeve" {
			t.Fatal("xeve must not be source-built on 7.1.5")
		}
	}
}

func TestPlanBlocksLibraryUnsupportedByVersion(t *testing.T) {
	// xeveb is a native library whose --enable-libxeveb switch does not exist before 8.1.
	settings := validVersionedSettings("8.0.3", []string{"xeveb"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected plan to be blocked when xeveb is selected on FFmpeg 8.0.3")
	}
	if !planWarningKeys(plan)["plan.warnings.libraryUnsupportedForVersion"] {
		t.Fatalf("expected libraryUnsupportedForVersion warning, got %#v", plan.Warnings)
	}
}

func TestPlanAllowsSameLibraryOnSupportingVersion(t *testing.T) {
	settings := validVersionedSettings("8.1.2", []string{"xeveb"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsExecutable {
		t.Fatalf("expected xeveb to be allowed on FFmpeg 8.1.2, warnings: %#v", plan.Warnings)
	}
	if planWarningKeys(plan)["plan.warnings.libraryUnsupportedForVersion"] {
		t.Fatalf("did not expect libraryUnsupportedForVersion on 8.1.2, got %#v", plan.Warnings)
	}
}

func TestPlanAllowsSupportedLibraryOnEightZero(t *testing.T) {
	settings := validVersionedSettings("8.0.3", []string{"x264"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsExecutable {
		t.Fatalf("expected x264 to be allowed on FFmpeg 8.0.3, warnings: %#v", plan.Warnings)
	}
}

func TestPlanAnnotatesUnsupportedLibrary(t *testing.T) {
	settings := validVersionedSettings("8.0.3", []string{"xeveb"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	for _, library := range plan.SelectedLibraries {
		if library.LibraryId != "xeveb" {
			continue
		}
		if library.VersionCompatibility == nil || library.VersionCompatibility.Supported {
			t.Fatalf("expected xeveb annotated unsupported on 8.0.3, got %#v", library.VersionCompatibility)
		}
		return
	}
	t.Fatalf("xeveb not found in selected libraries")
}

func TestLibraryCatalogForFfmpegSourceAnnotatesFullCatalog(t *testing.T) {
	catalog := LibraryCatalogForFfmpegSource("https://ffmpeg.org/releases/ffmpeg-8.0.3.tar.xz", "ucrt64")
	foundUnsupported := false
	foundSupported := false
	for _, library := range catalog {
		switch library.LibraryId {
		case "xeveb":
			if library.VersionCompatibility == nil || library.VersionCompatibility.Supported {
				t.Fatalf("expected xeveb annotated unsupported on 8.0.3, got %#v", library.VersionCompatibility)
			}
			foundUnsupported = true
		case "x264":
			if library.VersionCompatibility == nil || !library.VersionCompatibility.Supported {
				t.Fatalf("expected x264 annotated supported on 8.0.3, got %#v", library.VersionCompatibility)
			}
			foundSupported = true
		}
	}
	if !foundUnsupported || !foundSupported {
		t.Fatalf("expected catalog to include xeveb and x264, found xeveb=%v x264=%v", foundUnsupported, foundSupported)
	}
}

func TestCatalogShowsSourceBuildTrackForVersion(t *testing.T) {
	// The picker catalog must flip libplacebo to the Internal track on 5.1 (where it is
	// source-built) so the UI renders its "Source build" badge, but leave it Native on 7.1.
	catalog51 := LibraryCatalogForFfmpegSource("https://ffmpeg.org/releases/ffmpeg-5.1.9.tar.xz", "ucrt64")
	libplacebo51, found := libraryByIdInPlan(catalog51, "libplacebo")
	if !found {
		t.Fatal("libplacebo missing from 5.1.9 catalog")
	}
	if libplacebo51.TrackName != LibraryTrackInternal {
		t.Fatalf("expected libplacebo Internal track in 5.1.9 catalog, got %q", libplacebo51.TrackName)
	}
	catalog71 := LibraryCatalogForFfmpegSource("https://ffmpeg.org/releases/ffmpeg-7.1.5.tar.xz", "ucrt64")
	libplacebo71, found := libraryByIdInPlan(catalog71, "libplacebo")
	if !found {
		t.Fatal("libplacebo missing from 7.1.5 catalog")
	}
	if libplacebo71.TrackName != LibraryTrackNative {
		t.Fatalf("expected libplacebo Native track in 7.1.5 catalog, got %q", libplacebo71.TrackName)
	}
}

func TestManifestCoversEverySelectableCatalogLibrary(t *testing.T) {
	support81, _ := releasesupport.ForReleaseLine("8.1")
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		if library.Locked {
			continue // included/locked rows (ffmpeg.exe, libavcodec, ...) have no --enable switch
		}
		_, supported := support81.LibrarySupportFor(library.LibraryId)
		if knownUnreleasedCatalogLibraries[library.LibraryId] {
			if supported {
				t.Errorf("master-only library %q must NOT appear in the 8.1 manifest", library.LibraryId)
			}
			continue
		}
		if !supported {
			t.Errorf("catalog library %q is missing from the 8.1 support manifest", library.LibraryId)
		}
	}
}

func TestManifestLibraryIdsExistInCatalog(t *testing.T) {
	catalogIds := map[string]bool{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogIds[library.LibraryId] = true
	}
	for _, releaseLine := range []string{"8.1", "8.0", "7.1", "7.0", "6.1", "5.1", "4.4"} {
		support, _ := releasesupport.ForReleaseLine(releaseLine)
		for libraryId := range support.Libraries {
			if !catalogIds[libraryId] {
				t.Errorf("%s manifest references unknown catalog library id %q", releaseLine, libraryId)
			}
		}
	}
}

func TestManifestRecordsMinVersion(t *testing.T) {
	support81, _ := releasesupport.ForReleaseLine("8.1")
	support, _ := support81.LibrarySupportFor("dav1d")
	if support.MinVersion != "1.0.0" {
		t.Fatalf("expected dav1d minVersion 1.0.0 in 8.1 manifest, got %q", support.MinVersion)
	}
}

func TestPlanAdvisesRecommendedPatchOnSameLine(t *testing.T) {
	settings := validVersionedSettings("8.1.1", []string{"x264"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsExecutable {
		t.Fatalf("expected an older patch on a supported line to remain executable")
	}
	if !planWarningKeys(plan)["plan.warnings.ffmpegPatchOutdated"] {
		t.Fatalf("expected ffmpegPatchOutdated advisory recommending 8.1.2, got %#v", plan.Warnings)
	}
}

func TestPlanWarnsUnsupportedReleaseLine(t *testing.T) {
	settings := validVersionedSettings("6.0", []string{"x264"})
	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsExecutable {
		t.Fatalf("expected an unsupported release line to warn but remain executable")
	}
	if !planWarningKeys(plan)["plan.warnings.ffmpegReleaseUnsupported"] {
		t.Fatalf("expected ffmpegReleaseUnsupported advisory, got %#v", plan.Warnings)
	}
}

func TestHighestRecommendedFfmpegRelease(t *testing.T) {
	if got := highestRecommendedFfmpegRelease(); got != "8.1.2" {
		t.Fatalf("highestRecommendedFfmpegRelease() = %q, want 8.1.2", got)
	}
}

func TestEveryRecommendedReleaseMatchesItsLineKey(t *testing.T) {
	for lineKey, recommended := range supportedFfmpegReleases {
		if got := ffmpegReleaseLineKey(recommended); got != lineKey {
			t.Fatalf("recommended %q maps to line key %q, want %q", recommended, got, lineKey)
		}
	}
}
