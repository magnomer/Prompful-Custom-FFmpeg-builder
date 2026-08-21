//go:build windows

package hostexec

import (
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// LCommandTree owns a Windows Job Object that binds a spawned command and every
// descendant it creates. Windows does not terminate child processes when a
// parent is killed, so cancelling bash alone leaves pacman, make, and compiler
// children running; the job object closes that gap.
type LCommandTree struct {
	lCommandTreeHandle windows.Handle
}

// LCommandTreeCreate builds a kill-on-close job object and points command.Cancel
// at it, so context cancellation terminates the whole process tree instead of
// only the directly launched process. A nil result means job setup failed; the
// caller then keeps Go's default single-process cancellation.
func LCommandTreeCreate(command *exec.Cmd) *LCommandTree {
	if command == nil {
		return nil
	}
	handle, err := windows.CreateJobObject(nil, nil)
	if err != nil {
		return nil
	}
	limits := windows.JOBOBJECT_EXTENDED_LIMIT_INFORMATION{
		BasicLimitInformation: windows.JOBOBJECT_BASIC_LIMIT_INFORMATION{
			LimitFlags: windows.JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE,
		},
	}
	if _, err := windows.SetInformationJobObject(handle, windows.JobObjectExtendedLimitInformation, uintptr(unsafe.Pointer(&limits)), uint32(unsafe.Sizeof(limits))); err != nil {
		_ = windows.CloseHandle(handle)
		return nil
	}
	tree := &LCommandTree{lCommandTreeHandle: handle}
	command.Cancel = func() error {
		return windows.TerminateJobObject(tree.lCommandTreeHandle, 1)
	}
	return tree
}

// LCommandTreeMount binds the started process to the job so it and its
// descendants share the job's lifetime. Call once after command.Start; any
// child spawned afterwards is created inside the job.
func (tree *LCommandTree) LCommandTreeMount(command *exec.Cmd) {
	if tree == nil || command == nil || command.Process == nil {
		return
	}
	processHandle, err := windows.OpenProcess(windows.PROCESS_SET_QUOTA|windows.PROCESS_TERMINATE, false, uint32(command.Process.Pid))
	if err != nil {
		return
	}
	defer windows.CloseHandle(processHandle)
	_ = windows.AssignProcessToJobObject(tree.lCommandTreeHandle, processHandle)
}

// LCommandTreeClose releases the job handle after the command has exited.
// Because the job is kill-on-close, dropping the last handle also reaps any
// descendant that outlived its parent.
func (tree *LCommandTree) LCommandTreeClose() {
	if tree == nil {
		return
	}
	_ = windows.CloseHandle(tree.lCommandTreeHandle)
}
