package scripting

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	lines, err := ConfigureScriptLines([]string{"--enable-liblensfun"})
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
	lines, err := ConfigureScriptLines([]string{"--enable-libsvtjpegxs"})
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
