//go:build windows

package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows"
)

// LWorkspaceOwner is an inter-process ownership lock over one workspace. The
// in-memory action slot only serializes actions inside a single process; two
// GUIs, two CLI commands, or a GUI plus a CLI can otherwise each believe they
// own the same workspace and corrupt each other's downloads, sources, toolchain,
// or artifacts. Holding an OS-exclusive handle on a per-workspace lock file
// closes that gap; Windows releases the handle when the process exits, so a crash
// never leaves a stale lock.
type LWorkspaceOwner struct {
	lWorkspaceOwnerHandle windows.Handle
}

// lWorkspaceLockName is the per-workspace lock file. It lives at the workspace
// root beside the created layout folders and is never a cleanup target.
const lWorkspaceLockName = ".promptful-workspace.lock"

// LWorkspaceOwnerClaim reserves the workspace for the current process. It fails
// when another process already owns the same workspace, so a workspace-mutating
// approved action can refuse to start over live state owned elsewhere.
func LWorkspaceOwnerClaim(workspaceDirectory string) (*LWorkspaceOwner, error) {
	if workspaceDirectory == "" {
		return nil, errors.New("workspace directory must not be empty")
	}
	if err := os.MkdirAll(workspaceDirectory, 0o755); err != nil {
		return nil, err
	}
	lockPath := filepath.Join(workspaceDirectory, lWorkspaceLockName)
	lockPathUtf16, err := windows.UTF16PtrFromString(lockPath)
	if err != nil {
		return nil, err
	}
	// dwShareMode 0 denies other processes any access, so a concurrent claim of
	// the same workspace fails immediately with ERROR_SHARING_VIOLATION.
	handle, err := windows.CreateFile(lockPathUtf16, windows.GENERIC_READ|windows.GENERIC_WRITE, 0, nil, windows.OPEN_ALWAYS, windows.FILE_ATTRIBUTE_NORMAL, 0)
	if err != nil {
		if errors.Is(err, windows.ERROR_SHARING_VIOLATION) {
			return nil, fmt.Errorf("workspace is in use by another Promptful process: %s", workspaceDirectory)
		}
		return nil, err
	}
	return &LWorkspaceOwner{lWorkspaceOwnerHandle: handle}, nil
}

// LWorkspaceOwnerClose releases the workspace so another process may claim it.
// It is nil-safe: an action that never acquired an owner can call it freely.
func (owner *LWorkspaceOwner) LWorkspaceOwnerClose() {
	if owner == nil {
		return
	}
	_ = windows.CloseHandle(owner.lWorkspaceOwnerHandle)
}
