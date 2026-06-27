package app

import (
	"os"
	"path/filepath"
	"testing"

	"promptfulcustomffmpegbuilder/internal/planning"
)

const sampleFfmpegConfigure = `# excerpt
enabled libvvenc          && require_pkg_config libvvenc "libvvenc >= 1.6.1" "vvenc/vvenc.h" vvenc_get_version
enabled libuavs3d         && require_pkg_config libuavs3d "uavs3d >= 1.1.41" uavs3d.h uavs3d_decode
`

func TestRequiredPkgConfigMinVersion(t *testing.T) {
	if version, ok := requiredPkgConfigMinVersion(sampleFfmpegConfigure, "libvvenc"); !ok || version != "1.6.1" {
		t.Fatalf("libvvenc: got (%q,%v), want (1.6.1,true)", version, ok)
	}
	if version, ok := requiredPkgConfigMinVersion(sampleFfmpegConfigure, "uavs3d"); !ok || version != "1.1.41" {
		t.Fatalf("uavs3d: got (%q,%v), want (1.1.41,true)", version, ok)
	}
	if _, ok := requiredPkgConfigMinVersion(sampleFfmpegConfigure, "libnotpresent"); ok {
		t.Fatal("expected an absent module to not be found")
	}
}

func writeConfigure(t *testing.T, contents string) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "configure"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func planWith(prep planning.LibraryPreparation) planning.FfmpegBuildPlan {
	return planning.FfmpegBuildPlan{LibraryPreparations: []planning.LibraryPreparation{prep}}
}

func TestValidateLibraryVersionsBlocksTooOldPin(t *testing.T) {
	dir := writeConfigure(t, `require_pkg_config libvvenc "libvvenc >= 2.0.0" "vvenc/vvenc.h" vvenc_get_version`)
	app := &App{}
	plan := planWith(planning.LibraryPreparation{DisplayName: "vvenc", PkgConfigName: "libvvenc", Version: "1.14.0"})
	if err := app.validateLibraryVersionsAgainstFfmpeg(plan, dir, func(string, string) {}); err == nil {
		t.Fatal("expected a pin older than FFmpeg's required minimum to fail the preflight")
	}
}

func TestValidateLibraryVersionsAllowsSatisfiedPin(t *testing.T) {
	dir := writeConfigure(t, `require_pkg_config libvvenc "libvvenc >= 1.6.1" "vvenc/vvenc.h" vvenc_get_version`)
	app := &App{}
	plan := planWith(planning.LibraryPreparation{DisplayName: "vvenc", PkgConfigName: "libvvenc", Version: "1.14.0"})
	if err := app.validateLibraryVersionsAgainstFfmpeg(plan, dir, func(string, string) {}); err != nil {
		t.Fatalf("expected a satisfied pin to pass, got %v", err)
	}
}

func TestValidateLibraryVersionsSkipsUndecidablePin(t *testing.T) {
	// A moving "master" pin cannot be compared to a numeric minimum; the preflight must
	// skip it rather than block.
	dir := writeConfigure(t, `require_pkg_config libuavs3d "uavs3d >= 1.1.41" uavs3d.h uavs3d_decode`)
	app := &App{}
	plan := planWith(planning.LibraryPreparation{DisplayName: "libuavs3d", PkgConfigName: "uavs3d", Version: "master"})
	if err := app.validateLibraryVersionsAgainstFfmpeg(plan, dir, func(string, string) {}); err != nil {
		t.Fatalf("expected an undecidable pin to be skipped, got %v", err)
	}
}

func TestValidateLibraryVersionsSkipsWhenConfigureMissing(t *testing.T) {
	app := &App{}
	plan := planWith(planning.LibraryPreparation{DisplayName: "vvenc", PkgConfigName: "libvvenc", Version: "1.14.0"})
	if err := app.validateLibraryVersionsAgainstFfmpeg(plan, t.TempDir(), func(string, string) {}); err != nil {
		t.Fatalf("expected missing configure to skip, not fail, got %v", err)
	}
}
