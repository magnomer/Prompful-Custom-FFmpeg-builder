package program

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLResultTargetValidationUsesDisplayedPaths(t *testing.T) {
	workspaceDirectory := t.TempDir()
	artifactsDirectory := filepath.Join(workspaceDirectory, "FFmpeg", "8.1.2")
	if err := os.MkdirAll(artifactsDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	reportPath := filepath.Join(artifactsDirectory, "build-report-first.json")
	if err := os.WriteFile(reportPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := LResultDirectoryValidate(workspaceDirectory, artifactsDirectory); err != nil {
		t.Fatalf("expected displayed artifact directory to be accepted: %v", err)
	}
	if err := LResultReportValidate(workspaceDirectory, reportPath); err != nil {
		t.Fatalf("expected displayed report to be accepted: %v", err)
	}

	outsideArtifactDirectory := filepath.Join(workspaceDirectory, "logs")
	if err := os.MkdirAll(outsideArtifactDirectory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := LResultDirectoryValidate(workspaceDirectory, outsideArtifactDirectory); err == nil {
		t.Fatal("expected a directory outside FFmpeg to be rejected")
	}
	outsideReportPath := filepath.Join(outsideArtifactDirectory, "build-report-other.json")
	if err := os.WriteFile(outsideReportPath, []byte("{}"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := LResultReportValidate(workspaceDirectory, outsideReportPath); err == nil {
		t.Fatal("expected a report outside FFmpeg to be rejected")
	}
}

func TestLLogRunIdValidateRejectsDotSegments(t *testing.T) {
	for _, runID := range []string{".", "..", "child/record", `child\record`, "record\x00id"} {
		if err := LLogRunIdValidate(runID); err == nil {
			t.Errorf("expected run id %q to be rejected", runID)
		}
	}
	if err := LLogRunIdValidate("20260817T031500Z"); err != nil {
		t.Fatalf("expected a normal run id to be accepted: %v", err)
	}
}

func TestLRecordLocalReadKeepsDirectoryRunId(t *testing.T) {
	recordDirectory := LRecordEventsWrite(t, []LAuditLocalEvent{
		{RunId: "different-record", LAuditEventName: "action-started", Level: "info", Message: "started"},
	})
	record, ok := LRecordLocalRead(recordDirectory, "directory-record", false)
	if !ok {
		t.Fatal("expected the record to be read")
	}
	if record.RunId != "directory-record" {
		t.Fatalf("expected directory run id to remain authoritative, got %q", record.RunId)
	}
}
