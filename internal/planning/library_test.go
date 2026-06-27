package planning

import (
	"strings"
	"testing"
)

func TestPlanFfmpegBuildShowsLibrariesSeparately(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.LicenseProfileName = "gpl-local"
	settings.SelectedLibraryIds = []string{"x264", "opus"}
	settings.ExtraConfigureFlags = []string{"--disable-doc"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.SelectedLibraries) < 2 {
		t.Fatalf("expected selected libraries to be visible in plan, got %d", len(plan.SelectedLibraries))
	}
	if !librarySliceContains(plan.SelectedLibraries, "x264") || !librarySliceContains(plan.SelectedLibraries, "opus") {
		t.Fatalf("expected x264 and opus in selected libraries: %#v", plan.SelectedLibraries)
	}
	if !stringSliceContains(plan.GeneratedConfigureFlags, "--enable-libx264") {
		t.Fatalf("expected generated x264 flag in plan: %#v", plan.GeneratedConfigureFlags)
	}
	if !stringSliceContains(plan.RequiredMsys2PackageNames, "mingw-w64-ucrt-x86_64-libx264") {
		t.Fatalf("expected generated x264 package in plan: %#v", plan.RequiredMsys2PackageNames)
	}
	if !stringSliceContains(plan.ConfigureFlags, "--enable-gpl") {
		t.Fatalf("expected GPL flag in final configure flags: %#v", plan.ConfigureFlags)
	}
}

func TestOnnxRuntimeOmittedForMingw64Only(t *testing.T) {
	hasLibrary := func(profile string, libraryId string) bool {
		for _, library := range LibraryCatalogForShellProfile(profile) {
			if library.LibraryId == libraryId {
				return true
			}
		}
		return false
	}
	if !hasLibrary("ucrt64", "onnxruntime") {
		t.Fatalf("expected onnxruntime in ucrt64 catalog")
	}
	if !hasLibrary("clang64", "onnxruntime") {
		t.Fatalf("expected onnxruntime in clang64 catalog")
	}
	if hasLibrary("mingw64", "onnxruntime") {
		t.Fatalf("expected onnxruntime to be omitted from mingw64 catalog")
	}
	// The newly added always-available libraries must appear on every profile.
	for _, profile := range []string{"ucrt64", "mingw64", "clang64"} {
		for _, libraryId := range []string{"kvazaar", "dvdnav", "mbedtls"} {
			if !hasLibrary(profile, libraryId) {
				t.Fatalf("expected %s in %s catalog", libraryId, profile)
			}
		}
	}
}

func TestMsys2RootDirectoryIsPerProfile(t *testing.T) {
	workspace := `C:\ws`
	ucrt := Msys2RootDirectoryForProfile(workspace, "ucrt64")
	mingw := Msys2RootDirectoryForProfile(workspace, "mingw64")
	clang := Msys2RootDirectoryForProfile(workspace, "clang64")
	if ucrt == mingw || ucrt == clang || mingw == clang {
		t.Fatalf("expected distinct per-profile roots, got ucrt=%q mingw=%q clang=%q", ucrt, mingw, clang)
	}
	if !strings.HasSuffix(ucrt, "msys2-ucrt64") {
		t.Fatalf("expected ucrt64 root to end with msys2-ucrt64, got %q", ucrt)
	}
	// Empty profile falls back to ucrt64 so callers without a profile still resolve.
	if Msys2RootDirectoryForProfile(workspace, "") != ucrt {
		t.Fatalf("expected empty profile to fall back to ucrt64 root")
	}
}

func TestDefaultToolchainPackagesFollowShellProfile(t *testing.T) {
	cases := map[string]string{
		"ucrt64":  "mingw-w64-ucrt-x86_64-gcc",
		"mingw64": "mingw-w64-x86_64-gcc",
		"clang64": "mingw-w64-clang-x86_64-gcc",
	}
	for profile, wantPackage := range cases {
		packages := defaultMsys2PackageNames(profile)
		if !stringSliceContains(packages, wantPackage) {
			t.Fatalf("profile %q: expected toolchain package %q, got %#v", profile, wantPackage, packages)
		}
		// Unprefixed MSYS packages stay profile-independent.
		if !stringSliceContains(packages, "base-devel") {
			t.Fatalf("profile %q: expected base-devel to remain, got %#v", profile, packages)
		}
		// A non-ucrt profile must not leak the ucrt toolchain.
		if profile != "ucrt64" && stringSliceContains(packages, "mingw-w64-ucrt-x86_64-gcc") {
			t.Fatalf("profile %q: leaked ucrt64 toolchain package: %#v", profile, packages)
		}
	}
}

func TestDerivedLicenseBoundaryAllowsGplLibrary(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.LicenseProfileName = "lgpl-local"
	settings.SelectedLibraryIds = []string{"x264"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.LicenseProfileName != "gpl-local" {
		t.Fatalf("expected GPL boundary to be derived, got %s", plan.LicenseProfileName)
	}
}

func stringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func librarySliceContains(values []LibraryChoice, targetId string) bool {
	for _, value := range values {
		if value.LibraryId == targetId {
			return true
		}
	}
	return false
}

func TestVersion3LibraryAddsVersion3Flag(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"opencore-amr"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !stringSliceContains(plan.ConfigureFlags, "--enable-libopencore-amrnb") {
		t.Fatalf("expected OpenCORE AMR NB flag in final configure flags: %#v", plan.ConfigureFlags)
	}
	if !stringSliceContains(plan.ConfigureFlags, "--enable-version3") {
		t.Fatalf("expected --enable-version3 for OpenCORE AMR: %#v", plan.ConfigureFlags)
	}
}

func TestOpenSSLAndGnuTLSCannotBothBeEnabled(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"openssl", "gnutls"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected OpenSSL and GnuTLS together to be blocked")
	}
	if !stringSliceContains(plan.ConfigureFlags, "--enable-openssl") || !stringSliceContains(plan.ConfigureFlags, "--enable-gnutls") {
		t.Fatalf("expected both TLS flags in the blocked review plan: %#v", plan.ConfigureFlags)
	}
}

func TestShadercAndGlslangCannotBothBeEnabled(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"shaderc", "glslang"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected shaderc and glslang together to be blocked")
	}
	if !stringSliceContains(plan.ConfigureFlags, "--enable-libshaderc") || !stringSliceContains(plan.ConfigureFlags, "--enable-libglslang") {
		t.Fatalf("expected both shader compiler flags in the blocked review plan: %#v", plan.ConfigureFlags)
	}
}

func TestKnownLibraryPackageNamesMatchMSYS2Ucrt64(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"vmaf", "soxr"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !stringSliceContains(plan.RequiredMsys2PackageNames, "mingw-w64-ucrt-x86_64-vmaf") {
		t.Fatalf("expected MSYS2 package mingw-w64-ucrt-x86_64-vmaf, got %#v", plan.RequiredMsys2PackageNames)
	}
	if stringSliceContains(plan.RequiredMsys2PackageNames, "mingw-w64-ucrt-x86_64-libvmaf") {
		t.Fatalf("unexpected nonexistent MSYS2 package mingw-w64-ucrt-x86_64-libvmaf in %#v", plan.RequiredMsys2PackageNames)
	}
	if !stringSliceContains(plan.RequiredMsys2PackageNames, "mingw-w64-ucrt-x86_64-libsoxr") {
		t.Fatalf("expected MSYS2 package mingw-w64-ucrt-x86_64-libsoxr, got %#v", plan.RequiredMsys2PackageNames)
	}
	if stringSliceContains(plan.RequiredMsys2PackageNames, "mingw-w64-ucrt-x86_64-soxr") {
		t.Fatalf("unexpected nonexistent MSYS2 package mingw-w64-ucrt-x86_64-soxr in %#v", plan.RequiredMsys2PackageNames)
	}
}

func TestExistingCatalogLibrariesRemainNativeTrack(t *testing.T) {
	newMissingLibraryIds := map[string]bool{
		"avisynthplus": true,
		"davs2":        true,
		"uavs3d":       true,
		"lcevc-dec":    true,
		"mpeghdec":     true,
		"openvino":     true,
		"tensorflow":   true,
		"torch":        true,
		"quirc":        true,
		"klvanc":       true,
		"smbclient":    true,
		"libtls":       true,
		"vvenc":        true,
		"xavs2":        true,
		"libmfx":       true,
		"pocketsphinx": true,
		"dc1394":       true,
		"decklink":     true,
		"cuda-nvcc":    true,
	}
	for _, profile := range []string{"ucrt64", "mingw64", "clang64"} {
		for _, library := range LibraryCatalogForShellProfile(profile) {
			if newMissingLibraryIds[library.LibraryId] {
				continue
			}
			if library.TrackName != LibraryTrackNative {
				t.Fatalf("profile %s library %s: expected existing catalog entry to remain native, got %s", profile, library.LibraryId, library.TrackName)
			}
		}
	}
}

func TestMissingLibrariesAppearWithNonNativeTracks(t *testing.T) {
	expectedTracks := map[string]LibraryTrackName{
		"avisynthplus": LibraryTrackInternal,
		"davs2":        LibraryTrackInternal,
		"uavs3d":       LibraryTrackInternal,
		"lcevc-dec":    LibraryTrackInternal,
		"mpeghdec":     LibraryTrackInternal,
		"openvino":     LibraryTrackExternal,
		"tensorflow":   LibraryTrackExternal,
		"torch":        LibraryTrackExternal,
		"quirc":        LibraryTrackInternal,
		"klvanc":       LibraryTrackInternal,
		"smbclient":    LibraryTrackExternal,
		"libtls":       LibraryTrackInternal,
		"vvenc":        LibraryTrackInternal,
		"xavs2":        LibraryTrackInternal,
		"libmfx":       LibraryTrackExternal,
		"pocketsphinx": LibraryTrackInternal,
		"dc1394":       LibraryTrackInternal,
		"decklink":     LibraryTrackExternal,
		"cuda-nvcc":    LibraryTrackExternal,
	}
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	for libraryId, expectedTrack := range expectedTracks {
		library, exists := catalogById[libraryId]
		if !exists {
			t.Fatalf("expected missing-library catalog entry %s", libraryId)
		}
		if library.TrackName != expectedTrack {
			t.Fatalf("library %s: expected track %s, got %s", libraryId, expectedTrack, library.TrackName)
		}
		if len(library.PackageNames) != 0 {
			t.Fatalf("library %s: expected no Native MSYS2 packages, got %#v", libraryId, library.PackageNames)
		}
	}
}

func TestPlanFfmpegBuildGroupsSelectedLibrariesByTrack(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"x264", "opus"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.SelectedNativeLibraries) != len(plan.SelectedLibraries) {
		t.Fatalf("expected all preexisting selected libraries to be native, got native=%d total=%d", len(plan.SelectedNativeLibraries), len(plan.SelectedLibraries))
	}
	if len(plan.SelectedInternalLibraries) != 0 {
		t.Fatalf("expected no internal libraries for preexisting selection, got %#v", plan.SelectedInternalLibraries)
	}
	if len(plan.SelectedExternalLibraries) != 0 {
		t.Fatalf("expected no external libraries for preexisting selection, got %#v", plan.SelectedExternalLibraries)
	}
	if len(plan.SelectedLibrariesByTrack) != 3 {
		t.Fatalf("expected three track groups, got %#v", plan.SelectedLibrariesByTrack)
	}
}

func TestEvcProfileBindingsAreMutuallyExclusive(t *testing.T) {
	// FFmpeg configure rejects enabling both full- and baseline-profile EVC bindings.
	decoderWarnings, decoderBlocked := validateConfigureFlagConflicts([]string{"--enable-libxevd", "--enable-libxevdb"})
	if !decoderBlocked {
		t.Fatal("expected libxevd + libxevdb to block the plan")
	}
	if !hasWarningKey(decoderWarnings, "plan.warnings.evcDecoderConflict") {
		t.Fatalf("expected evcDecoderConflict warning, got %#v", decoderWarnings)
	}
	encoderWarnings, encoderBlocked := validateConfigureFlagConflicts([]string{"--enable-libxeve", "--enable-libxeveb"})
	if !encoderBlocked {
		t.Fatal("expected libxeve + libxeveb to block the plan")
	}
	if !hasWarningKey(encoderWarnings, "plan.warnings.evcEncoderConflict") {
		t.Fatalf("expected evcEncoderConflict warning, got %#v", encoderWarnings)
	}
	// Either binding alone must not block.
	if _, blocked := validateConfigureFlagConflicts([]string{"--enable-libxevd", "--enable-libxeve"}); blocked {
		t.Fatal("expected full-profile-only EVC bindings to be allowed")
	}
}

func hasWarningKey(warnings []PlanWarning, key string) bool {
	for _, warning := range warnings {
		if warning.MessageKey == key {
			return true
		}
	}
	return false
}

func TestPackageResolutionOnlyUsesNativeTrack(t *testing.T) {
	libraries := []LibraryChoice{
		trackedLibraryChoice(LibraryTrackNative, "native-test", "Native Test", "Test", []string{"--enable-native-test"}, []string{"mingw-w64-ucrt-x86_64-native-test"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "internal-test", "Internal Test", "Test", []string{"--enable-internal-test"}, []string{"must-not-install-internal"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackExternal, "external-test", "External Test", "Test", []string{"--enable-external-test"}, []string{"must-not-install-external"}, "lgpl"),
	}
	packages := uniquePackagesFromLibraries(libraries)
	if !stringSliceContains(packages, "mingw-w64-ucrt-x86_64-native-test") {
		t.Fatalf("expected native package, got %#v", packages)
	}
	if stringSliceContains(packages, "must-not-install-internal") || stringSliceContains(packages, "must-not-install-external") {
		t.Fatalf("expected internal/external tracks to be excluded from pacman packages, got %#v", packages)
	}
}

func TestUnpreparedNonNativeLibrariesBlockExecution(t *testing.T) {
	internalLibrary := trackedLibraryChoice(LibraryTrackInternal, "internal-test", "Internal Test", "Test", []string{"--enable-internal-test"}, nil, "lgpl")
	externalLibrary := trackedLibraryChoice(LibraryTrackExternal, "external-test", "External Test", "Test", []string{"--enable-external-test"}, nil, "lgpl")

	warnings, blocked := appendUnpreparedTrackWarnings(nil, []LibraryChoice{internalLibrary, externalLibrary})
	if !blocked {
		t.Fatalf("expected non-native libraries without a prep recipe to block execution")
	}
	if len(warnings) != 2 {
		t.Fatalf("expected one warning per blocked non-native track, got %#v", warnings)
	}
	if warnings[0].RiskLevel != RiskLevelBlocked || warnings[1].RiskLevel != RiskLevelBlocked {
		t.Fatalf("expected blocked warnings for unprepared non-native libraries, got %#v", warnings)
	}
}

func TestPartitionSplitsPreparableFromBlocked(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	// uavs3d has an implemented internal recipe; smbclient does not yet.
	preparable, blocked := partitionNonNativeLibraries([]LibraryChoice{catalogById["uavs3d"], catalogById["smbclient"]}, "8.1.2")
	if len(preparable) != 1 || preparable[0].LibraryId != "uavs3d" {
		t.Fatalf("expected uavs3d to be preparable, got %#v", preparable)
	}
	if len(blocked) != 1 || blocked[0].LibraryId != "smbclient" {
		t.Fatalf("expected smbclient to remain blocked, got %#v", blocked)
	}
}

func TestPlanFfmpegBuildPreparesLibraryWithRecipe(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"uavs3d"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.IsExecutable {
		t.Fatalf("expected a library with an implemented prep recipe to keep the plan executable: %#v", plan.Warnings)
	}
	if len(plan.LibraryPreparations) != 1 || plan.LibraryPreparations[0].LibraryId != "uavs3d" {
		t.Fatalf("expected uavs3d preparation in the plan, got %#v", plan.LibraryPreparations)
	}
	preparation := plan.LibraryPreparations[0]
	if preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected uavs3d build system cmake, got %q", preparation.BuildSystem)
	}
	if preparation.ArchiveSha256Hash == "" || preparation.ArchiveUrl == "" {
		t.Fatalf("expected version layer to populate archive url+hash, got %#v", preparation)
	}
	if !operationSliceContains(plan.Operations, "prepare-internal-libraries") {
		t.Fatalf("expected internal prep operation when a preparable internal library is selected: %#v", plan.Operations)
	}
}

func TestAvisynthPlusRecipeIsHeaderOnlyCMake(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	preparation, exists := preparationForLibrary(catalogById["avisynthplus"], "8.1.2")
	if !exists {
		t.Fatal("expected avisynthplus to have an implemented recipe")
	}
	if preparation.Method != PreparationMethodInternalSource || preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected internal cmake build, got method=%q buildSystem=%q", preparation.Method, preparation.BuildSystem)
	}
	if !stringSliceContains(preparation.CMakeOptions, "-DHEADERS_ONLY=ON") {
		t.Fatalf("expected HEADERS_ONLY cmake option, got %#v", preparation.CMakeOptions)
	}
	if !stringSliceContains(preparation.CMakeBuildTargets, "VersionGen") {
		t.Fatalf("expected VersionGen build target, got %#v", preparation.CMakeBuildTargets)
	}
	if preparation.VerifyLibStem != "" {
		t.Fatalf("expected no link library for header-only avisynth, got %q", preparation.VerifyLibStem)
	}
	if preparation.VerifyHeaderRelativePath != "avisynth/avisynth_c.h" {
		t.Fatalf("expected FFmpeg-required header path, got %q", preparation.VerifyHeaderRelativePath)
	}
}

func TestDavs2RecipeIsConfigureMakeGpl(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	davs2 := catalogById["davs2"]
	if davs2.LicenseEffectName != "gpl" {
		t.Fatalf("expected libdavs2 to be GPL (FFmpeg requires --enable-gpl), got %q", davs2.LicenseEffectName)
	}
	preparation, exists := preparationForLibrary(davs2, "8.1.2")
	if !exists {
		t.Fatal("expected davs2 to have an implemented recipe")
	}
	if preparation.BuildSystem != BuildSystemConfigureMake {
		t.Fatalf("expected configure-make build system, got %q", preparation.BuildSystem)
	}
	if preparation.ConfigureSubdir != "build/linux" {
		t.Fatalf("expected configure subdir build/linux, got %q", preparation.ConfigureSubdir)
	}
	if !stringSliceContains(preparation.MakeInstallTargets, "install-lib-static") {
		t.Fatalf("expected install-lib-static target, got %#v", preparation.MakeInstallTargets)
	}
	if preparation.VerifyLibStem != "davs2" || preparation.VerifyHeaderRelativePath != "davs2.h" {
		t.Fatalf("unexpected verify fields: %#v", preparation)
	}
}

func TestXavs2RecipeIsConfigureMakeStatic(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	xavs2 := catalogById["xavs2"]
	if xavs2.LicenseEffectName != "gpl" {
		t.Fatalf("expected xavs2 to be GPL (FFmpeg requires --enable-gpl), got %q", xavs2.LicenseEffectName)
	}
	preparation, exists := preparationForLibrary(xavs2, "8.1.2")
	if !exists {
		t.Fatal("expected xavs2 to have an implemented recipe")
	}
	if preparation.BuildSystem != BuildSystemConfigureMake {
		t.Fatalf("expected configure-make build system, got %q", preparation.BuildSystem)
	}
	if preparation.ConfigureSubdir != "build/linux" {
		t.Fatalf("expected configure subdir build/linux, got %q", preparation.ConfigureSubdir)
	}
	if !stringSliceContains(preparation.MakeInstallTargets, "install-lib-static") {
		t.Fatalf("expected install-lib-static target, got %#v", preparation.MakeInstallTargets)
	}
	if preparation.VerifyLibStem != "xavs2" || preparation.VerifyHeaderRelativePath != "xavs2.h" {
		t.Fatalf("unexpected verify fields: %#v", preparation)
	}
}

func TestLcevcDecRecipeIsStaticCMakeWithPythonBuildDep(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	lcevc := catalogById["lcevc-dec"]
	if lcevc.LicenseEffectName != "lgpl" {
		t.Fatalf("expected LCEVCdec to stay LGPL (BSD-licensed upstream, no --enable-gpl), got %q", lcevc.LicenseEffectName)
	}
	preparation, exists := preparationForLibrary(lcevc, "8.1.2")
	if !exists {
		t.Fatal("expected lcevc-dec to have an implemented recipe")
	}
	if preparation.Method != PreparationMethodInternalSource || preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected internal cmake build, got method=%q buildSystem=%q", preparation.Method, preparation.BuildSystem)
	}
	if !stringSliceContains(preparation.CMakeOptions, "-DBUILD_SHARED_LIBS=OFF") {
		t.Fatalf("expected a static build so the .pc links every component archive, got %#v", preparation.CMakeOptions)
	}
	if preparation.VerifyHeaderRelativePath != "LCEVC/lcevc_dec.h" || preparation.VerifyLibStem != "lcevc_dec_api" {
		t.Fatalf("expected FFmpeg-required header + lcevc_dec_api stem, got %#v", preparation)
	}
	// The recipe stores a profile-independent suffix; flatten leaves it unprefixed.
	if !stringSliceContains(preparation.BuildDependencyPackages, "python") {
		t.Fatalf("expected python build dependency (version_files.py needs Python3), got %#v", preparation.BuildDependencyPackages)
	}
	// Upstream's static .pc lists the C++/math runtime before its own archives; the recipe
	// repairs the link order by appending them to the lcevc_dec.pc Libs line.
	if preparation.PkgConfigName != "lcevc_dec" {
		t.Fatalf("expected pkg-config module lcevc_dec, got %q", preparation.PkgConfigName)
	}
	if !stringSliceContains(preparation.PkgConfigAppendLibs, "stdc++") || !stringSliceContains(preparation.PkgConfigAppendLibs, "m") {
		t.Fatalf("expected stdc++ and m appended to the .pc for static GNU link order, got %#v", preparation.PkgConfigAppendLibs)
	}
}

func TestLibtlsRecipeUsesPrivatePrefixIsolation(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	preparation, exists := preparationForLibrary(catalogById["libtls"], "8.1.2")
	if !exists {
		t.Fatal("expected libtls to have an implemented recipe")
	}
	if preparation.Method != PreparationMethodInternalSource || preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected internal cmake build, got method=%q buildSystem=%q", preparation.Method, preparation.BuildSystem)
	}
	// LibreSSL's archives collide with the openssl package in the shared prefix, so libtls
	// must install privately and bind its own archives by absolute ${libdir} path.
	if !preparation.PrivatePrefixInstall {
		t.Fatal("expected libtls to use private-prefix isolation")
	}
	if preparation.PkgConfigName != "libtls" {
		t.Fatalf("expected pkg-config module libtls, got %q", preparation.PkgConfigName)
	}
	for _, want := range []string{"${libdir}/libtls.a", "${libdir}/libssl.a", "${libdir}/libcrypto.a"} {
		if !strings.Contains(preparation.PkgConfigLibsLine, want) {
			t.Fatalf("expected absolute archive %q in Libs line, got %q", want, preparation.PkgConfigLibsLine)
		}
	}
}

func TestVvencRecipeIsStaticLibraryOnlyCMake(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	vvenc := catalogById["vvenc"]
	if vvenc.LicenseEffectName != "lgpl" {
		t.Fatalf("expected vvenc to stay LGPL (BSD-licensed upstream, no --enable-gpl), got %q", vvenc.LicenseEffectName)
	}
	preparation, exists := preparationForLibrary(vvenc, "8.1.2")
	if !exists {
		t.Fatal("expected vvenc to have an implemented recipe")
	}
	if preparation.Method != PreparationMethodInternalSource || preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected internal cmake build, got method=%q buildSystem=%q", preparation.Method, preparation.BuildSystem)
	}
	if !stringSliceContains(preparation.CMakeOptions, "-DBUILD_SHARED_LIBS=OFF") || !stringSliceContains(preparation.CMakeOptions, "-DVVENC_LIBRARY_ONLY=ON") {
		t.Fatalf("expected static, library-only cmake options, got %#v", preparation.CMakeOptions)
	}
	if preparation.VerifyHeaderRelativePath != "vvenc/vvenc.h" || preparation.VerifyLibStem != "vvenc" {
		t.Fatalf("expected FFmpeg-required header + vvenc stem, got %#v", preparation)
	}
	if preparation.PkgConfigName != "libvvenc" {
		t.Fatalf("expected pkg-config module libvvenc (FFmpeg checks libvvenc >= 1.6.1), got %q", preparation.PkgConfigName)
	}
	// vvenc needs no build dependencies and no .pc fixup (runtime sits in Libs.private).
	if len(preparation.BuildDependencyPackages) != 0 {
		t.Fatalf("expected no build dependencies, got %#v", preparation.BuildDependencyPackages)
	}
	if len(preparation.PkgConfigAppendLibs) != 0 {
		t.Fatalf("expected no .pc fixup, got %#v", preparation.PkgConfigAppendLibs)
	}
	if preparation.Version != "1.14.0" {
		t.Fatalf("expected resolved version 1.14.0 from library-sources.json, got %q", preparation.Version)
	}
}

func TestMpeghdecRecipeIsNonfreeStaticCMake(t *testing.T) {
	catalogById := map[string]LibraryChoice{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		catalogById[library.LibraryId] = library
	}
	mpeghdec := catalogById["mpeghdec"]
	// FFmpeg lists libmpeghdec under EXTERNAL_LIBRARY_NONFREE_LIST, so the catalog row must
	// carry --enable-nonfree and surface a nonfree license effect.
	if mpeghdec.LicenseEffectName != "nonfree" {
		t.Fatalf("expected mpeghdec to be nonfree (FFmpeg requires --enable-nonfree), got %q", mpeghdec.LicenseEffectName)
	}
	if !stringSliceContains(mpeghdec.ConfigureFlags, "--enable-libmpeghdec") || !stringSliceContains(mpeghdec.ConfigureFlags, "--enable-nonfree") {
		t.Fatalf("expected --enable-libmpeghdec and --enable-nonfree, got %#v", mpeghdec.ConfigureFlags)
	}
	preparation, exists := preparationForLibrary(mpeghdec, "8.1.2")
	if !exists {
		t.Fatal("expected mpeghdec to have an implemented recipe")
	}
	if preparation.Method != PreparationMethodInternalSource || preparation.BuildSystem != BuildSystemCMake {
		t.Fatalf("expected internal cmake build, got method=%q buildSystem=%q", preparation.Method, preparation.BuildSystem)
	}
	if !stringSliceContains(preparation.CMakeOptions, "-DBUILD_SHARED_LIBS=OFF") || !stringSliceContains(preparation.CMakeOptions, "-Dmpeghdec_BUILD_BINARIES=OFF") {
		t.Fatalf("expected static, binaries-off cmake options, got %#v", preparation.CMakeOptions)
	}
	if preparation.VerifyHeaderRelativePath != "mpeghdec/mpeghdecoder.h" || preparation.VerifyLibStem != "mpeghdec" {
		t.Fatalf("expected FFmpeg-required header + mpeghdec stem, got %#v", preparation)
	}
	if preparation.PkgConfigName != "mpeghdec" {
		t.Fatalf("expected pkg-config module mpeghdec (FFmpeg checks mpeghdec >= 3.0.0), got %q", preparation.PkgConfigName)
	}
	// The C++ static mpeghdec.pc lists only "-lmpeghdec -lm"; the recipe appends the C++
	// runtime so FFmpeg's C link probe resolves std::/operator new.
	if !stringSliceContains(preparation.PkgConfigAppendLibs, "stdc++") {
		t.Fatalf("expected stdc++ appended to the .pc for static GNU link order, got %#v", preparation.PkgConfigAppendLibs)
	}
	if preparation.Version != "r3.0.3" {
		t.Fatalf("expected resolved version r3.0.3 from library-sources.json, got %q", preparation.Version)
	}
	// Two source patches: (1) disable the libFDK SHORT fMin/fMax overloads that duplicate
	// FIXP_SGL(=short) on Windows/LLP64; (2) add -DMPEGHDEC_STATIC=1 to the .pc Cflags so
	// the static archive's symbols are not referenced as dllimport (__imp_...).
	if len(preparation.SourcePatches) != 2 {
		t.Fatalf("expected two source patches (libFDK overloads + static .pc Cflags), got %#v", preparation.SourcePatches)
	}
	patchByFile := map[string]LibrarySourcePatch{}
	for _, sourcePatch := range preparation.SourcePatches {
		patchByFile[sourcePatch.File] = sourcePatch
	}
	if patchByFile["src/libFDK/include/common_fix.h"].Find != "#if !defined(_MSC_VER) && defined(__x86_64__)" {
		t.Fatalf("missing or wrong libFDK overload patch: %#v", preparation.SourcePatches)
	}
	if pcPatch := patchByFile["mpeghdec.pc.in"]; !strings.Contains(pcPatch.Replace, "-DMPEGHDEC_STATIC=1") {
		t.Fatalf("missing or wrong static .pc Cflags patch: %#v", preparation.SourcePatches)
	}
}

func TestFfmpegVersionFromArchiveUrl(t *testing.T) {
	cases := map[string]string{
		"https://ffmpeg.org/releases/ffmpeg-8.1.2.tar.xz":     "8.1.2",
		"https://ffmpeg.org/releases/ffmpeg-7.1.tar.gz":       "7.1",
		"https://ffmpeg.org/releases/ffmpeg-9.tar.xz":         "9",
		"https://ffmpeg.org/releases/ffmpeg-snapshot.tar.bz2": "",
		"": "",
	}
	for url, want := range cases {
		if got := ffmpegVersionFromArchiveUrl(url); got != want {
			t.Fatalf("ffmpegVersionFromArchiveUrl(%q) = %q, want %q", url, got, want)
		}
	}
}

func TestLcevcDecBuildDependenciesArePrefixedPerProfile(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.WindowsShellProfileName = "ucrt64"
	settings.SelectedLibraryIds = []string{"lcevc-dec"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.LibraryPreparations) != 1 {
		t.Fatalf("expected exactly the lcevc-dec preparation, got %#v", plan.LibraryPreparations)
	}
	if !stringSliceContains(plan.LibraryPreparations[0].BuildDependencyPackages, "mingw-w64-ucrt-x86_64-python") {
		t.Fatalf("expected the python build dependency prefixed for the ucrt64 profile, got %#v", plan.LibraryPreparations[0].BuildDependencyPackages)
	}
}

func TestPlanFfmpegBuildBlocksLibraryWithoutRecipe(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	settings.SelectedLibraryIds = []string{"smbclient"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected a non-native library without a prep recipe to block the plan")
	}
	if len(plan.LibraryPreparations) != 0 {
		t.Fatalf("expected no preparations for a library without a recipe, got %#v", plan.LibraryPreparations)
	}
}

func TestExtraFlagNonNativeLibraryIsGated(t *testing.T) {
	settings := DefaultFfmpegBuildSettings()
	settings.WorkspaceDirectory = `C:\CustomFFmpegBuilder\workspace`
	settings.FfmpegSourceArchiveUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz"
	settings.FfmpegSourceSignatureUrl = "https://ffmpeg.org/releases/ffmpeg-test.tar.xz.asc"
	settings.FfmpegSourceSha256Hash = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	// smbclient is non-native with no recipe; reaching it via the extra-flags box
	// must still block, so the track gate cannot be bypassed by a raw configure flag.
	settings.ExtraConfigureFlags = []string{"--enable-libsmbclient"}

	plan, err := PlanFfmpegBuild(settings)
	if err != nil {
		t.Fatal(err)
	}
	if plan.IsExecutable {
		t.Fatalf("expected a non-native library added via extra flags to block the plan")
	}
}

func TestFfmpegBuildOperationsIncludeNonNativePreparationStepsOnlyWhenNeeded(t *testing.T) {
	nativeOperations := ffmpegBuildOperations(false, false)
	if operationSliceContains(nativeOperations, "prepare-internal-libraries") || operationSliceContains(nativeOperations, "prepare-external-libraries") || operationSliceContains(nativeOperations, "verify-prepared-libraries") {
		t.Fatalf("expected native-only operations to stay unchanged, got %#v", nativeOperations)
	}

	mixedOperations := ffmpegBuildOperations(true, true)
	for _, operationName := range []string{"prepare-internal-libraries", "prepare-external-libraries", "verify-prepared-libraries"} {
		if !operationSliceContains(mixedOperations, operationName) {
			t.Fatalf("expected operation %s in mixed-track operations: %#v", operationName, mixedOperations)
		}
	}
}

func operationSliceContains(values []PlanOperation, targetName string) bool {
	for _, value := range values {
		if value.OperationName == targetName {
			return true
		}
	}
	return false
}
