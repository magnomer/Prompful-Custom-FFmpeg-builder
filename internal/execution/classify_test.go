package execution

import "testing"

func TestClassifyLogLine(t *testing.T) {
	cases := []struct {
		name    string
		dflt    string
		line    string
		wantLvl string
	}{
		// Benign pacman noise — demoted from warn to info.
		{"pacman reinstall", "warn", "warning: mingw-w64-ucrt-x86_64-aom-3.14.1-1 is up to date -- reinstalling", "info"},
		{"dependency cycle", "warn", "warning: dependency cycle detected:", "info"},
		{"cycle continuation", "warn", "warning: mingw-w64-ucrt-x86_64-harfbuzz will be installed before its mingw-w64-ucrt-x86_64-freetype dependency", "info"},

		// GCC continuation context — demoted to info.
		{"note offset", "warn", "libavcodec/aacenc.h:81:13: note: at offset 8 into destination object 'group_len' of size 8", "info"},
		{"pragma message", "warn", "D:/.../CL/cl_version.h:22:9: note: '#pragma message: cl_version.h: CL_TARGET_OPENCL_VERSION is not defined. Defaulting to 300 (OpenCL 3.0)'", "info"},
		{"in file included", "warn", "In file included from D:/tempc/.../CL/cl.h:20,", "info"},
		{"in function header", "warn", "libavfilter/af_afftfilt.c: In function 'filter_channel':", "info"},
		{"inlined from", "warn", "inlined from 'generate_kernel' at libavfilter/af_firequalizer.c:691:13:", "info"},
		{"source echo numbered", "warn", "297 |         memmove(buf, buf + s->hop_size, window_size * sizeof(float));", "info"},
		{"source echo caret", "warn", "    | ^~~~~~~~~~~~~~~~", "info"},

		// Benign third-party prep-build warnings (uavs3d/davs2 upstream) — demoted to info.
		{"cdecl redefined", "warn", "D:/.../uavs3d-master/source/decoder/uavs3d.h:52:9: warning: '__cdecl' redefined", "info"},
		{"unused variable", "warn", "D:/.../davs2-1.7/source/common/scantab.h:435:23: warning: 'tab_scan_cg' defined but not used [-Wunused-variable]", "info"},
		{"unused but set", "warn", "D:/.../davs2-1.7/source/common/intra.cc:551:12: warning: variable 'dst_base' set but not used [-Wunused-but-set-variable=]", "info"},
		{"unknown pragma", "warn", "D:/.../davs2-1.7/source/common/decoder.cc:54: warning: ignoring '#pragma warning ' [-Wunknown-pragmas]", "info"},
		{"nasm legacy macro", "warn", "D:/.../davs2-1.7/source/common/x86/cpu-a.asm:39: warning: dropping trailing empty parameter in call to multi-line macro `DEFINE_ARGS_INTERNAL' [-w+pp-macro-params-legacy]", "info"},

		// Genuine warnings — stay warn. stringop-overflow is NOT demoted: it can flag real bugs.
		{"real stringop warning", "warn", "libavcodec/aacenc.c:901:35: warning: writing 1 byte into a region of size 0 [-Wstringop-overflow=]", "warn"},
		{"upstream stringop stays warn", "warn", "D:/.../uavs3d-master/source/decore/intra_pred.c:1932:23: warning: writing 16 bytes into a region of size 15 [-Wstringop-overflow=]", "warn"},
		{"format truncation", "warn", "libavformat/dashenc.c:1449:65: warning: '-stream' directive output may be truncated [-Wformat-truncation=]", "warn"},

		// Errors — promoted to error.
		{"gcc error", "warn", "libavcodec/foo.c:10:5: error: 'x' undeclared (first use in this function)", "error"},
		{"linker error", "warn", "undefined reference to `foo'", "error"},
		{"make abort", "info", "make: *** [Makefile:120: libavcodec/foo.o] Error 1", "error"},
		{"make ignored error", "warn", "make: [Makefile:245: common/x86/pixel-32.o] Error 1 (ignored)", "info"},
		{"strip empty object", "warn", "D:/tempc/toolchains/msys2-ucrt64/ucrt64/bin/strip.exe: error: the input file 'common/x86/pixel-32.o' has no sections", "info"},

		// Plain progress — keep default.
		{"plain stdout info", "info", "CC      libavcodec/aacenc.o", "info"},
		{"plain stderr default", "warn", ":: Synchronizing package databases...", "warn"},
		{"empty line", "warn", "   ", "warn"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := classifyLogLine(c.dflt, c.line); got != c.wantLvl {
				t.Errorf("classifyLogLine(%q, %q) = %q, want %q", c.dflt, c.line, got, c.wantLvl)
			}
		})
	}
}

func TestIsTransientNetworkFailureLine(t *testing.T) {
	cases := []struct {
		name string
		line string
		want bool
	}{
		// The exact failures that aborted a setup run.
		{"pacman slow", "error: failed retrieving file 'heimdal-libs-7.8.0-5-x86_64.pkg.tar.zst' from repo.msys2.org : Operation too slow. Less than 1 bytes/sec transferred the last 10 seconds", true},
		{"pacman sig slow", "error: failed retrieving file 'bison-3.8.2-5-x86_64.pkg.tar.zst.sig' from repo.msys2.org : Operation too slow. Less than 1 bytes/sec transferred the last 10 seconds", true},
		{"pacman commit abort", "error: failed to commit transaction (unexpected error)", true},

		// Other transient network failures across the program (git/configure/make).
		{"dns failure", "fatal: unable to access 'https://github.com/...': Could not resolve host: github.com", true},
		{"git rpc", "error: RPC failed; curl 18 transfer closed with outstanding read data remaining", true},
		{"git early eof", "fatal: early EOF", true},
		{"conn refused", "curl: (7) Failed to connect to repo.msys2.org port 443: Connection refused", true},

		// Real build/install errors must NOT be treated as transient.
		{"compile error", "libavcodec/foo.c:10:5: error: 'x' undeclared (first use in this function)", false},
		{"linker error", "undefined reference to `foo'", false},
		{"make abort", "make: *** [Makefile:120: libavcodec/foo.o] Error 1", false},
		{"pacman conflict", "error: failed to commit transaction (conflicting files)", false},
		{"plain progress", "CC      libavcodec/aacenc.o", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := isTransientNetworkFailureLine(c.line); got != c.want {
				t.Errorf("isTransientNetworkFailureLine(%q) = %v, want %v", c.line, got, c.want)
			}
		})
	}
}
