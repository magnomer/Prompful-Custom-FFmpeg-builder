package main

import (
	"os"
	"os/signal"
	"syscall"

	"promptfulcustomffmpegbuilder/internal/program"
)

// lActionSignalCancel forwards a terminal interrupt (Ctrl-C) or termination
// signal to the running approved action's cancel path, so the inline CLI worker
// runs the same finalization as backend cancellation: failure cleanup, terminal
// status, and the final audit events, instead of the process dying mid-run. The
// returned stop function removes the handler; call it once the action returns.
func lActionSignalCancel(driver *program.LProgram) func() {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, os.Interrupt, syscall.SIGTERM)
	handlerDone := make(chan struct{})
	go func() {
		select {
		case <-signalChannel:
			driver.LActionApprovedCancel()
		case <-handlerDone:
		}
	}()
	return func() {
		signal.Stop(signalChannel)
		close(handlerDone)
	}
}
