package scripting

import (
	"strings"
	"testing"
)

// LConfigureScriptTestCreate builds the configure script for one flag set and
// FFmpeg version, joins it to a single string, and fails on error.
func LConfigureScriptTestCreate(t *testing.T, configureFlags []string, ffmpegVersion string) string {
	t.Helper()
	scriptLines, err := LConfigureScriptCreate(configureFlags, nil, ffmpegVersion)
	if err != nil {
		t.Fatalf("build configure script for %q: %v", ffmpegVersion, err)
	}
	return strings.Join(scriptLines, "\n")
}

// Section 5 — ONNX Runtime nested include dir joins extra-cflags (Job 06).
func TestLConfigureScript901OnnxRuntimeCflags(t *testing.T) {
	withOnnx := LConfigureScriptTestCreate(t, []string{"--enable-libonnxruntime"}, "9.0.1")
	if !strings.Contains(withOnnx, `--extra-cflags=-I${onnx_incdir}`) {
		t.Fatalf("onnxruntime script missing --extra-cflags include line:\n%s", withOnnx)
	}
	if !strings.Contains(withOnnx, "/include/onnxruntime") {
		t.Fatalf("onnxruntime script missing include/onnxruntime dir:\n%s", withOnnx)
	}

	withoutOnnx := LConfigureScriptTestCreate(t, []string{"--enable-libx264"}, "9.0.1")
	if strings.Contains(withoutOnnx, "/include/onnxruntime") {
		t.Fatalf("script without onnxruntime unexpectedly references include/onnxruntime:\n%s", withoutOnnx)
	}
}

// Section 7 — AMF header preflight is scoped to FFmpeg 9+.
func TestLConfigureScriptAmfPreflightGated(t *testing.T) {
	script901 := LConfigureScriptTestCreate(t, []string{"--enable-amf"}, "9.0.1")
	if !strings.Contains(script901, "amf_version_header") {
		t.Fatalf("9.0.1 --enable-amf script missing AMF header preflight:\n%s", script901)
	}
	if !strings.Contains(script901, "FFmpeg 9 requires AMF >= 1.5.2.0 for --enable-amf") {
		t.Fatalf("9.0.1 AMF preflight missing version diagnostic:\n%s", script901)
	}
	script812 := LConfigureScriptTestCreate(t, []string{"--enable-amf"}, "8.1.2")
	if strings.Contains(script812, "amf_version_header") {
		t.Fatalf("8.1.2 --enable-amf must not include the FFmpeg 9 AMF preflight:\n%s", script812)
	}
}

// Section 7 — NVENC (ffnvcodec) header preflight is scoped to FFmpeg 9+.
func TestLConfigureScriptNvencPreflightGated(t *testing.T) {
	script901 := LConfigureScriptTestCreate(t, []string{"--enable-ffnvcodec"}, "9.0.1")
	if !strings.Contains(script901, "nvenc_version_header") {
		t.Fatalf("9.0.1 --enable-ffnvcodec script missing NVENC header preflight:\n%s", script901)
	}
	if !strings.Contains(script901, "FFmpeg 9 requires NVENC SDK >= 11.1") {
		t.Fatalf("9.0.1 NVENC preflight missing version diagnostic:\n%s", script901)
	}
	script812 := LConfigureScriptTestCreate(t, []string{"--enable-ffnvcodec"}, "8.1.2")
	if strings.Contains(script812, "nvenc_version_header") {
		t.Fatalf("8.1.2 --enable-ffnvcodec must not include the FFmpeg 9 NVENC preflight:\n%s", script812)
	}
}
