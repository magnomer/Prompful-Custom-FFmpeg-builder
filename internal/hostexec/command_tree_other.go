//go:build !windows

package hostexec

import "os/exec"

// LCommandTree is a no-op outside Windows, where cancelling a command already
// signals its process group. It exists so callers share one call shape on every
// platform without build-tag guards.
type LCommandTree struct{}

// LCommandTreeCreate returns nil on non-Windows platforms; the methods below are
// nil-safe so callers need no guard.
func LCommandTreeCreate(command *exec.Cmd) *LCommandTree {
	_ = command
	return nil
}

// LCommandTreeMount is a no-op outside Windows.
func (tree *LCommandTree) LCommandTreeMount(command *exec.Cmd) { _ = command }

// LCommandTreeClose is a no-op outside Windows.
func (tree *LCommandTree) LCommandTreeClose() {}
