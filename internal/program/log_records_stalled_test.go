package program

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// LRecordEventsWrite writes the given audit events as a security-events.jsonl file
// inside a fresh record directory and returns that directory.
func LRecordEventsWrite(t *testing.T, events []LAuditLocalEvent) string {
	t.Helper()
	recordDirectory := t.TempDir()
	file, err := os.Create(filepath.Join(recordDirectory, "security-events.jsonl"))
	if err != nil {
		t.Fatalf("could not create audit log: %v", err)
	}
	defer file.Close()
	for _, event := range events {
		eventBytes, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("could not marshal event: %v", err)
		}
		if _, err := file.Write(append(eventBytes, '\n')); err != nil {
			t.Fatalf("could not write event: %v", err)
		}
	}
	return recordDirectory
}

func TestLRecordStalledStatusIsWarnNotError(t *testing.T) {
	recordDirectory := LRecordEventsWrite(t, []LAuditLocalEvent{
		{LAuditEventName: "action-started", Level: "info", Message: "started"},
		{LAuditEventName: "action-stalled", Level: "warn", Message: "network stalled: tried repo.msys2.org"},
	})

	record, ok := LRecordLocalRead(recordDirectory, "20260814T000000Z", true)
	if !ok {
		t.Fatal("expected the record to be read")
	}
	if record.Status != "stalled" {
		t.Fatalf("expected status stalled, got %q", record.Status)
	}
	if record.WarnCount != 1 {
		t.Fatalf("expected WarnCount 1, got %d", record.WarnCount)
	}
	if record.ErrorCount != 0 {
		t.Fatalf("expected ErrorCount 0 on a stalled run, got %d", record.ErrorCount)
	}
}

func TestLRecordStalledOverridesEarlierErrorLine(t *testing.T) {
	// A stalled pacman run streams an error-classified retrieval line before the
	// terminal stall event. The terminal action-stalled must still win.
	recordDirectory := LRecordEventsWrite(t, []LAuditLocalEvent{
		{LAuditEventName: "log", Level: "error", Message: "error: failed retrieving file 'x.pkg' from repo.msys2.org"},
		{LAuditEventName: "action-stalled", Level: "warn", Message: "network stalled"},
	})

	record, _ := LRecordLocalRead(recordDirectory, "20260814T000001Z", true)
	if record.Status != "stalled" {
		t.Fatalf("expected status stalled to override an earlier error, got %q", record.Status)
	}
}

func TestLRecordGenuineFailureStaysFailed(t *testing.T) {
	recordDirectory := LRecordEventsWrite(t, []LAuditLocalEvent{
		{LAuditEventName: "log", Level: "error", Message: "compile error"},
		{LAuditEventName: "action-failed", Level: "error", Message: "FFmpeg build failed"},
	})

	record, _ := LRecordLocalRead(recordDirectory, "20260814T000002Z", true)
	if record.Status != "failed" {
		t.Fatalf("expected status failed, got %q", record.Status)
	}
}
