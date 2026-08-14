package planning

import "testing"

// The catalog once shipped FFmpeg 9.0.1 preparation references that still pointed
// at versions/8.1.2/*.go source files. LSourcePathCheck must reject a work path
// whose version segment does not match the owning record's FFmpeg version.
func TestSourcePathVersionMismatchIsError(t *testing.T) {
	report := &LCatalogValidationReport{}
	file := LCatalogEmbeddedFile{PathName: "versions/9.0.1.json"}
	LSourcePathCheck(report, file, "versions/8.1.2/vvenc.go", "9.0.1")
	if report.LErrorCount() == 0 {
		t.Fatalf("expected an error for a version-mismatched source path, got %d issues", len(report.Issues))
	}
}

func TestSourcePathVersionMatchIsClean(t *testing.T) {
	report := &LCatalogValidationReport{}
	file := LCatalogEmbeddedFile{PathName: "versions/9.0.1.json"}
	LSourcePathCheck(report, file, "versions/9.0.1/vvenc.go", "9.0.1")
	if len(report.Issues) != 0 {
		t.Fatalf("expected no issues for a matching source path, got %d", len(report.Issues))
	}
}
