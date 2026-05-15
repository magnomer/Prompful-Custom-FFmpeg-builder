package workspace

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestRealPathCheckRejectsSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink privileges vary on Windows runners")
	}
	workspaceDirectory := t.TempDir()
	outsideDirectory := t.TempDir()
	linkPath := filepath.Join(workspaceDirectory, "link")
	if err := os.Symlink(outsideDirectory, linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}
	candidatePath := filepath.Join(linkPath, "file.txt")
	if err := CheckRealPathInsideWorkspace(workspaceDirectory, candidatePath); err == nil {
		t.Fatal("expected symlink escape to be rejected")
	}
}

func TestRealPathCheckAllowsMissingChildInsideWorkspace(t *testing.T) {
	workspaceDirectory := t.TempDir()
	candidatePath := filepath.Join(workspaceDirectory, "nested", "file.txt")
	if err := CheckRealPathInsideWorkspace(workspaceDirectory, candidatePath); err != nil {
		t.Fatalf("expected missing child inside workspace to be allowed: %v", err)
	}
}
