//go:build !windows

package hostexec

import "os/exec"

// LCommandWindowHide is a no-op on non-Windows platforms, where spawning a
// console subprocess does not allocate a visible window. It exists so callers
// can invoke the same helper on every platform without build-tag guards.
func LCommandWindowHide(command *exec.Cmd) {
	_ = command
}
