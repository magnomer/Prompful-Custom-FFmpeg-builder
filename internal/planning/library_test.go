package planning

import "testing"

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
