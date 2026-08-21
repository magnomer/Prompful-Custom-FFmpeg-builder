//go:build !windows

package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/sys/unix"
)

// LWorkspaceOwner is an inter-process ownership lock over one workspace. See the
// Windows build for the full rationale; here an advisory flock on a per-workspace
// lock file provides the same single-owner guarantee, released by the kernel when
// the process exits so a crash never leaves a stale lock.
type LWorkspaceOwner struct {
	lWorkspaceOwnerFile *os.File
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
	lockFile, err := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0o644)
	if err != nil {
		return nil, err
	}
	if err := unix.Flock(int(lockFile.Fd()), unix.LOCK_EX|unix.LOCK_NB); err != nil {
		_ = lockFile.Close()
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, fmt.Errorf("workspace is in use by another Promptful process: %s", workspaceDirectory)
		}
		return nil, err
	}
	return &LWorkspaceOwner{lWorkspaceOwnerFile: lockFile}, nil
}

// LWorkspaceOwnerClose releases the workspace so another process may claim it.
// It is nil-safe: an action that never acquired an owner can call it freely.
func (owner *LWorkspaceOwner) LWorkspaceOwnerClose() {
	if owner == nil || owner.lWorkspaceOwnerFile == nil {
		return
	}
	_ = unix.Flock(int(owner.lWorkspaceOwnerFile.Fd()), unix.LOCK_UN)
	_ = owner.lWorkspaceOwnerFile.Close()
}
