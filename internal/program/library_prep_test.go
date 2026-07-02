package program

import (
	"strings"
	"testing"

	"promptfulcustomffmpegbuilder/internal/planning"
	version519 "promptfulcustomffmpegbuilder/versions/5.1.9"
	"promptfulcustomffmpegbuilder/versions/shared"
)

func TestLibjxl519PreparationPromotesStaticDefineToPkgConfigCflags(t *testing.T) {
	recipePlan := shared.NewLibraryPreparationPlan("5.1.9", "libjxl", "versions/5.1.9/libjxl.go")
	version519.LLibraryLibjxlPrepare(recipePlan)

	preparation := planning.LPreparationFromVersionLibraryPlan(*recipePlan, planning.LLibrarySourcePin{
		Version: "0.8.2",
		Url:     "https://example.test/libjxl.tar.gz",
		Sha256:  strings.Repeat("0", 64),
		Format:  "tar.gz",
		Host:    "example.test",
	})

	scriptLines, err := LScriptPreparationBuild(preparation)
	if err != nil {
		t.Fatalf("build libjxl preparation script: %v", err)
	}
	script := strings.Join(scriptLines, "\n")
	if !strings.Contains(script, "-DJXL_STATIC_DEFINE") {
		t.Fatalf("libjxl preparation script did not append JXL static define:\n%s", script)
	}
	if !strings.Contains(script, "Patched pkg-config Cflags") {
		t.Fatalf("libjxl preparation script did not patch pkg-config Cflags:\n%s", script)
	}
	if !strings.Contains(script, "-L${libdir} -ljxl") {
		t.Fatalf("libjxl preparation script did not keep libjxl in linker flag form:\n%s", script)
	}
	if strings.Contains(script, "${libdir}/libjxl.a") {
		t.Fatalf("libjxl preparation script uses absolute archive path that FFmpeg configure orders before test object:\n%s", script)
	}
	if !strings.Contains(script, "Overriding libjxl_threads.pc Libs line") {
		t.Fatalf("libjxl preparation script did not patch libjxl_threads.pc:\n%s", script)
	}
	if !strings.Contains(script, "-L${libdir} -ljxl_threads -lstdc++ -pthread") {
		t.Fatalf("libjxl preparation script did not add the C++ runtime to libjxl_threads.pc:\n%s", script)
	}
}
