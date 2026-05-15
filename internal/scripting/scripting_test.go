package scripting

import (
	"os"
	"path/filepath"
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
