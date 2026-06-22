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

		// Genuine warnings — stay warn.
		{"real stringop warning", "warn", "libavcodec/aacenc.c:901:35: warning: writing 1 byte into a region of size 0 [-Wstringop-overflow=]", "warn"},
		{"format truncation", "warn", "libavformat/dashenc.c:1449:65: warning: '-stream' directive output may be truncated [-Wformat-truncation=]", "warn"},

		// Errors — promoted to error.
		{"gcc error", "warn", "libavcodec/foo.c:10:5: error: 'x' undeclared (first use in this function)", "error"},
		{"linker error", "warn", "undefined reference to `foo'", "error"},
		{"make abort", "info", "make: *** [Makefile:120: libavcodec/foo.o] Error 1", "error"},

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
