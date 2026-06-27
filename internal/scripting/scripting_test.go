package scripting

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestInternalCMakePrivatePrefixIsolatesLibrary(t *testing.T) {
	lines, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId:                "libtls",
		DisplayName:              "libtls",
		BuildSystem:              "cmake",
		CMakeOptions:             []string{"-DBUILD_SHARED_LIBS=OFF"},
		PkgConfigName:            "libtls",
		PkgConfigLibsLine:        "${libdir}/libtls.a ${libdir}/libssl.a ${libdir}/libcrypto.a -lws2_32 -lbcrypt -lntdll",
		PrivatePrefixInstall:     true,
		VerifyHeaderRelativePath: "tls.h",
		VerifyLibStem:            "tls",
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	// install_prefix must be the per-library private dir, wiped and created before install.
	if !strings.Contains(joined, `install_prefix="${profile_prefix}/`+PrivateLibraryInstallSubdir+`/libtls"; rm -rf "${install_prefix}"; mkdir -p "${install_prefix}"`) {
		t.Fatalf("expected private install_prefix setup, got:\n%s", joined)
	}
	// CMake must install into the private prefix (windows form derived from install_prefix).
	if !strings.Contains(joined, `install_prefix_win="$(cygpath -m "${install_prefix}")"`) {
		t.Fatalf("expected cmake to install into the private prefix, got:\n%s", joined)
	}
	// Libs override binds absolute private archives; Requires lines stripped.
	if !strings.Contains(joined, `Libs: ${libdir}/libtls.a ${libdir}/libssl.a ${libdir}/libcrypto.a`) {
		t.Fatalf("expected absolute-archive Libs override, got:\n%s", joined)
	}
	if !strings.Contains(joined, `sed -i -E '/^Requires(\.private)?:/d'`) {
		t.Fatalf("expected Requires/Requires.private strip, got:\n%s", joined)
	}
	// Verification must check the private prefix, not the shared one.
	if !strings.Contains(joined, `installed_header="${install_prefix}/include/tls.h"`) {
		t.Fatalf("expected verification under install_prefix, got:\n%s", joined)
	}
	if strings.Contains(joined, `${profile_prefix}/lib/pkgconfig/libtls.pc`) {
		t.Fatalf("private install must not touch the shared-prefix pkgconfig, got:\n%s", joined)
	}
}

func TestPrivateLibraryPkgConfigDir(t *testing.T) {
	got := PrivateLibraryPkgConfigDir("/ucrt64", "libtls")
	want := "/ucrt64/opt/customffmpeg/libtls/lib/pkgconfig"
	if got != want {
		t.Fatalf("PrivateLibraryPkgConfigDir = %q, want %q", got, want)
	}
}

func TestConfigureScriptLinesPrependsPrivatePkgConfigDirs(t *testing.T) {
	lines, err := ConfigureScriptLines([]string{"--enable-libtls"}, []string{"/ucrt64/opt/customffmpeg/libtls/lib/pkgconfig"})
	if err != nil {
		t.Fatalf("ConfigureScriptLines: %v", err)
	}
	joined := strings.Join(lines, "\n")
	if !strings.Contains(joined, `export PKG_CONFIG_PATH="/ucrt64/opt/customffmpeg/libtls/lib/pkgconfig:${profile_prefix}/lib/pkgconfig`) {
		t.Fatalf("expected private pkgconfig dir prepended to PKG_CONFIG_PATH, got:\n%s", joined)
	}
	if _, err := ConfigureScriptLines([]string{"--enable-libtls"}, []string{"/ucrt64/bad dir/pkgconfig"}); err == nil {
		t.Fatal("expected an unsafe private pkgconfig dir to be rejected")
	}
}

func TestPreparationScriptsAreValidBash(t *testing.T) {
	bashPath, lookErr := exec.LookPath("bash")
	if lookErr != nil {
		t.Skip("bash not available to syntax-check generated scripts")
	}
	// On Windows, LookPath may resolve to the WSL bash.exe stub rather than a real
	// POSIX bash. Probe it; skip unless it actually behaves like bash.
	if output, err := exec.Command(bashPath, "-c", "echo ok").Output(); err != nil || strings.TrimSpace(string(output)) != "ok" {
		t.Skip("no usable bash on PATH to syntax-check generated scripts")
	}
	internalSpec := LibraryBuildSpec{LibraryId: "uavs3d", DisplayName: "libuavs3d", VerifyHeaderRelativePath: "uavs3d.h", VerifyLibStem: "uavs3d"}
	externalSpec := LibraryBuildSpec{LibraryId: "tensorflow", DisplayName: "TensorFlow", ImportIncludeSubdir: "include", ImportLibSubdir: "lib", VerifyHeaderRelativePath: "tensorflow/c/c_api.h", VerifyLibStem: "tensorflow"}

	internalLines, err := InternalLibrarySourceBuildScriptLines(internalSpec)
	if err != nil {
		t.Fatalf("internal script generation: %v", err)
	}
	externalLines, err := ExternalLibraryImportScriptLines(externalSpec)
	if err != nil {
		t.Fatalf("external script generation: %v", err)
	}

	for name, scriptLines := range map[string][]string{"internal": internalLines, "external": externalLines} {
		scriptPath := filepath.Join(t.TempDir(), name+".sh")
		if err := os.WriteFile(scriptPath, []byte(strings.Join(scriptLines, "\n")+"\n"), 0o700); err != nil {
			t.Fatalf("%s write: %v", name, err)
		}
		if output, err := exec.Command(bashPath, "-n", scriptPath).CombinedOutput(); err != nil {
			t.Fatalf("%s script failed bash syntax check: %v\n%s", name, err, output)
		}
	}
}

func TestInternalBuildSystemDispatch(t *testing.T) {
	base := LibraryBuildSpec{LibraryId: "x", DisplayName: "X", VerifyHeaderRelativePath: "x.h", VerifyLibStem: "x"}
	for _, buildSystem := range []string{"", "cmake"} {
		spec := base
		spec.BuildSystem = buildSystem
		if _, err := InternalLibrarySourceBuildScriptLines(spec); err != nil {
			t.Fatalf("cmake build system %q should generate, got %v", buildSystem, err)
		}
	}
	for _, buildSystem := range []string{"autotools", "meson"} {
		spec := base
		spec.BuildSystem = buildSystem
		if _, err := InternalLibrarySourceBuildScriptLines(spec); err == nil {
			t.Fatalf("build system %q should not yet generate a script", buildSystem)
		}
	}
}

func TestInternalMakeBuildInstallsHeaderAndStaticLib(t *testing.T) {
	lines, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId:                "quirc",
		DisplayName:              "libquirc",
		BuildSystem:              "make",
		MakeBuildTargets:         []string{"libquirc.a"},
		MakeVariables:            []string{"SDL_CFLAGS=", "SDL_LIBS="},
		MakeInstallHeaderFiles:   []string{"lib/quirc.h"},
		MakeStaticLibFile:        "libquirc.a",
		VerifyHeaderRelativePath: "quirc.h",
		VerifyLibStem:            "quirc",
	})
	if err != nil {
		t.Fatalf("make build system should generate, got %v", err)
	}
	joined := strings.Join(lines, "\n")
	for _, expected := range []string{
		`make -j"$(nproc)" 'SDL_CFLAGS=' 'SDL_LIBS=' 'libquirc.a' 2>&1`,
		`built_header="${src_dir}/lib/quirc.h"`,
		`cp "${built_header}" "${profile_prefix}/include/"`,
		`built_lib="${src_dir}/libquirc.a"`,
		`cp "${built_lib}" "${profile_prefix}/lib/"`,
		`installed_header="${install_prefix}/include/quirc.h"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected generated make script to contain %q, got:\n%s", expected, joined)
		}
	}
}

func TestInternalMakeBuildRequiresInstallArtifacts(t *testing.T) {
	if _, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId: "quirc", BuildSystem: "make", VerifyHeaderRelativePath: "quirc.h",
		MakeStaticLibFile: "libquirc.a",
	}); err == nil {
		t.Fatal("expected make build without install header files to be rejected")
	}
	if _, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId: "quirc", BuildSystem: "make", VerifyHeaderRelativePath: "quirc.h",
		MakeInstallHeaderFiles: []string{"lib/quirc.h"},
	}); err == nil {
		t.Fatal("expected make build without a static lib file to be rejected")
	}
}

func TestPreparationScriptRejectsUnsafeFields(t *testing.T) {
	if _, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{LibraryId: "x", VerifyHeaderRelativePath: "../escape.h"}); err == nil {
		t.Fatal("expected path traversal in verify header to be rejected")
	}
	if _, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{LibraryId: "x", VerifyHeaderRelativePath: "ok.h", CMakeOptions: []string{"-DX=$(rm -rf /)"}}); err == nil {
		t.Fatal("expected unsafe cmake option to be rejected")
	}
	if _, err := ExternalLibraryImportScriptLines(LibraryBuildSpec{LibraryId: "x", VerifyHeaderRelativePath: "ok.h", ImportIncludeSubdir: "include", ImportLibSubdir: "; rm -rf /"}); err == nil {
		t.Fatal("expected unsafe import subdir to be rejected")
	}
}

func TestWriteScriptFileReplacesDifferentExistingContent(t *testing.T) {
	workspaceDirectory := t.TempDir()
	scriptPath := filepath.Join(workspaceDirectory, "build", "scripts", "run.sh")
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(scriptPath, []byte("stale previous run\n"), 0o700); err != nil {
		t.Fatalf("write existing: %v", err)
	}
	writtenScript, err := WriteScriptFile(ScriptFilePlan{WorkspaceDirectory: workspaceDirectory, ScriptFilePath: scriptPath, ScriptLines: []string{"#!/usr/bin/env bash", "echo safe"}})
	if err != nil {
		t.Fatalf("expected stale script to be replaced: %v", err)
	}
	if writtenScript.ScriptSha256Hash == "" {
		t.Fatal("expected script hash")
	}
	content, err := os.ReadFile(scriptPath)
	if err != nil {
		t.Fatalf("read written script: %v", err)
	}
	if string(content) != "#!/usr/bin/env bash\necho safe\n" {
		t.Fatalf("unexpected script content: %q", string(content))
	}
}

func TestWriteScriptFileRejectsSymlinkPath(t *testing.T) {
	workspaceDirectory := t.TempDir()
	scriptPath := filepath.Join(workspaceDirectory, "build", "scripts", "run.sh")
	if err := os.MkdirAll(filepath.Dir(scriptPath), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	targetPath := filepath.Join(workspaceDirectory, "target.sh")
	if err := os.WriteFile(targetPath, []byte("target\n"), 0o700); err != nil {
		t.Fatalf("write target: %v", err)
	}
	if err := os.Symlink(targetPath, scriptPath); err != nil {
		t.Skipf("symlink not available on this platform/configuration: %v", err)
	}
	_, err := WriteScriptFile(ScriptFilePlan{WorkspaceDirectory: workspaceDirectory, ScriptFilePath: scriptPath, ScriptLines: []string{"#!/usr/bin/env bash", "echo safe"}})
	if err == nil {
		t.Fatal("expected symlink script path to be rejected")
	}
}

func TestConfigureScriptLinesSkipsLensfunWhenFfmpegApiIsIncompatible(t *testing.T) {
	lines, err := ConfigureScriptLines([]string{"--enable-liblensfun"}, nil)
	if err != nil {
		t.Fatalf("ConfigureScriptLines: %v", err)
	}
	joined := strings.Join(lines, "\n")
	for _, expected := range []string{
		"try_enable_lensfun",
		"lensfun_ffmpeg_api_probe",
		"lensfun is hidden from automatic presets for now. Backend support is left for future compatibility.",
		"remove_configure_flag --enable-liblensfun",
		"lensfun pkg-config diagnostic skipped because --enable-liblensfun was disabled after compatibility checks.",
		"----- BEGIN ffbuild/config.log tail -----",
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected generated script to contain %q, got:\n%s", expected, joined)
		}
	}
	if strings.Contains(joined, "apply_lensfun_compatibility_patch") || strings.Contains(joined, "lf_db_create/lf_db_new") {
		t.Fatalf("lensfun fallback must not patch only part of the API mismatch, got:\n%s", joined)
	}
}

func TestConfigureScriptLinesTriesAndSkipsSvtJpegxsWhenIncompatible(t *testing.T) {
	lines, err := ConfigureScriptLines([]string{"--enable-libsvtjpegxs"}, nil)
	if err != nil {
		t.Fatalf("ConfigureScriptLines: %v", err)
	}
	joined := strings.Join(lines, "\n")
	for _, expected := range []string{
		"try_enable_svt_jpeg_xs",
		"Trying MSYS2/package-provided SvtJpegxs first.",
		"Trying official upstream SVT-JPEG-XS source as a fallback.",
		"SVT JPEG XS is hidden from the UI for now. Backend support is left for future compatibility.",
		"remove_configure_flag --enable-libsvtjpegxs",
		"SVT JPEG XS pkg-config diagnostic skipped because --enable-libsvtjpegxs was disabled after compatibility checks.",
		`./configure "${configure_flags[@]}"`,
	} {
		if !strings.Contains(joined, expected) {
			t.Fatalf("expected generated script to contain %q, got:\n%s", expected, joined)
		}
	}
	if strings.Contains(joined, "Patching SvtJpegxs.pc Version") {
		t.Fatalf("SVT JPEG XS fallback must not fake the pkg-config version, got:\n%s", joined)
	}
}

func TestInternalCMakeBuildAppendsPkgConfigLibsForStaticLinkOrder(t *testing.T) {
	lines, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId:                "lcevc-dec",
		DisplayName:              "liblcevc-dec",
		BuildSystem:              "cmake",
		CMakeOptions:             []string{"-DBUILD_SHARED_LIBS=OFF"},
		PkgConfigName:            "lcevc_dec",
		PkgConfigAppendLibs:      []string{"stdc++", "m"},
		VerifyHeaderRelativePath: "LCEVC/lcevc_dec.h",
		VerifyLibStem:            "lcevc_dec_api",
	})
	if err != nil {
		t.Fatal(err)
	}
	joined := strings.Join(lines, "\n")
	// The fixup must target the right .pc and rewrite the Libs line.
	if !strings.Contains(joined, `pc_file="${install_prefix}/lib/pkgconfig/lcevc_dec.pc"`) {
		t.Fatalf("expected the lcevc_dec.pc path in the fixup, got:\n%s", joined)
	}
	if !strings.Contains(joined, `s/ -lstdc\+\+( |$)/\1/g`) || !strings.Contains(joined, `s/$/ -lstdc++ -lm/`) {
		t.Fatalf("expected strip-then-append sed for stdc++/m, got:\n%s", joined)
	}
	// The patch must run after install and before verification so the verified install is
	// the patched one.
	installIdx := strings.Index(joined, `cmake --install`)
	patchIdx := strings.Index(joined, `sed -i -E`)
	verifyIdx := strings.LastIndex(joined, "lcevc_dec_api")
	if !(installIdx < patchIdx && patchIdx < verifyIdx) {
		t.Fatalf("expected order install < patch < verify, got install=%d patch=%d verify=%d", installIdx, patchIdx, verifyIdx)
	}

	// Run the actual sed program against a sample upstream static Libs line and confirm the
	// runtime libs end up exactly once, at the end, after the component archives.
	bashPath, lookErr := exec.LookPath("bash")
	if lookErr != nil {
		t.Skip("bash not available")
	}
	if out, err := exec.Command(bashPath, "-c", "echo ok").Output(); err != nil || strings.TrimSpace(string(out)) != "ok" {
		t.Skip("no usable bash on PATH")
	}
	pc := filepath.Join(t.TempDir(), "lcevc_dec.pc")
	original := "Libs: -L\"/ucrt64/lib\" -lstdc++ -lm -llcevc_dec_api -llcevc_dec_pipeline_cpu -llcevc_dec_common\n"
	if err := os.WriteFile(pc, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	var sedLine string
	for _, line := range lines {
		if strings.HasPrefix(line, "sed -i -E ") {
			sedLine = strings.Replace(line, `"${pc_file}"`, "'"+pc+"'", 1)
			break
		}
	}
	if sedLine == "" {
		t.Fatal("could not find the sed line in the generated script")
	}
	if out, err := exec.Command(bashPath, "-c", sedLine).CombinedOutput(); err != nil {
		t.Fatalf("sed failed: %v\n%s", err, out)
	}
	patched, err := os.ReadFile(pc)
	if err != nil {
		t.Fatal(err)
	}
	got := strings.TrimSpace(string(patched))
	want := `Libs: -L"/ucrt64/lib" -llcevc_dec_api -llcevc_dec_pipeline_cpu -llcevc_dec_common -lstdc++ -lm`
	if got != want {
		t.Fatalf("patched Libs line wrong:\n got: %s\nwant: %s", got, want)
	}
}

func TestInternalCMakeBuildNoPkgConfigFixupWhenUnset(t *testing.T) {
	lines, err := InternalLibrarySourceBuildScriptLines(LibraryBuildSpec{
		LibraryId:                "uavs3d",
		DisplayName:              "libuavs3d",
		BuildSystem:              "cmake",
		VerifyHeaderRelativePath: "uavs3d.h",
		VerifyLibStem:            "uavs3d",
	})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(strings.Join(lines, "\n"), "pkgconfig/") {
		t.Fatalf("expected no .pc fixup when the recipe declares none")
	}
}
