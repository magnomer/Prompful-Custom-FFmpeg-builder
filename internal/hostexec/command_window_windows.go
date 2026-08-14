//go:build windows

// Package hostexec centralizes host process-launch adjustments that differ by
// operating system. Its single responsibility is preparing an *exec.Cmd so a
// GUI application does not flash a console window when it spawns a console
// subprocess (bash, pacman, ffmpeg, and similar toolchain commands).
package hostexec

import (
	"os/exec"
	"syscall"
)

// createNoWindow is the Windows process-creation flag that prevents a new
// console window from being allocated for a spawned console program. It is
// defined here because older syscall packages do not export the constant.
const createNoWindow = 0x08000000

// LCommandWindowHide suppresses the console window Windows would otherwise
// allocate for a spawned console subprocess. It must be called after the
// command is constructed and before it is started. A nil command is ignored so
// callers need not guard the call site.
func LCommandWindowHide(command *exec.Cmd) {
	if command == nil {
		return
	}
	if command.SysProcAttr == nil {
		command.SysProcAttr = &syscall.SysProcAttr{}
	}
	command.SysProcAttr.HideWindow = true
	command.SysProcAttr.CreationFlags |= createNoWindow
}
