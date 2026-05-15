package execution

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestPrepareScriptForStdinReplacesPathWithStdinFlag(t *testing.T) {
	workspaceDirectory := t.TempDir()
	executablePath := filepath.Join(workspaceDirectory, "toolchains", "msys64", "usr", "bin", "bash.exe")
	workingDirectory := filepath.Join(workspaceDirectory, "toolchains", "msys64")
	scriptPath := filepath.Join(workspaceDirectory, "build", "scripts", "run.sh")
	for _, directoryPath := range []string{filepath.Dir(executablePath), workingDirectory, filepath.Dir(scriptPath)} {
		if err := os.MkdirAll(directoryPath, 0o755); err != nil {
			t.Fatalf("mkdir: %v", err)
		}
	}
	if err := os.WriteFile(executablePath, []byte("placeholder"), 0o755); err != nil {
		t.Fatalf("write exe: %v", err)
	}
	scriptBytes := []byte("#!/usr/bin/env bash\necho safe\n")
	if err := os.WriteFile(scriptPath, scriptBytes, 0o700); err != nil {
		t.Fatalf("write script: %v", err)
	}
	scriptHash := sha256.Sum256(scriptBytes)
	spec := CommandPlan{
		ExecutablePath:             executablePath,
		ArgumentValues:             []string{filepath.ToSlash(scriptPath)},
		WorkingDirectory:           workingDirectory,
		WorkspaceDirectory:         workspaceDirectory,
		AllowedExecutableBasenames: []string{"bash.exe"},
		ScriptKind:                 FfmpegMakeScript,
		ApprovedScriptFilePath:     scriptPath,
		ApprovedScriptSha256Hash:   hex.EncodeToString(scriptHash[:]),
	}
	readScriptBytes, updatedArguments, err := prepareScriptForStdin(spec)
	if err != nil {
		t.Fatalf("prepare script: %v", err)
	}
	if string(readScriptBytes) != string(scriptBytes) {
		t.Fatal("script bytes changed")
	}
	if !reflect.DeepEqual(updatedArguments, []string{"-s"}) {
		t.Fatalf("unexpected arguments: %#v", updatedArguments)
	}
}
