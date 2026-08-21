package program

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"promptfulcustomffmpegbuilder/internal/audit"
	"promptfulcustomffmpegbuilder/internal/download"
)

func (program *LProgram) LActionApprovedCancel() bool {
	program.LMutexAction.Lock()
	defer program.LMutexAction.Unlock()
	if program.LActionCancelFunction == nil {
		return false
	}
	program.LActionCancelFunction()
	program.LLogEmit("warn", LLocaleTextGetInternal("logs.system.cancellationRequested", nil))
	return true
}

func (program *LProgram) LActionApprovedStart() (string, context.Context, error) {
	program.LMutexAction.Lock()
	defer program.LMutexAction.Unlock()
	if program.lProgramStopping {
		return "", nil, errors.New("program is shutting down")
	}
	if program.LActionCancelFunction != nil || program.lActionDone != nil {
		return "", nil, errors.New("an approved action is already running")
	}
	LContextAction, LActionCancelFunction := context.WithCancel(context.Background())
	program.LContextAction = LContextAction
	program.LActionCancelFunction = LActionCancelFunction
	program.lActionDone = make(chan struct{})
	// The UTC-second timestamp alone is not unique: one short action can finish
	// and another start within the same second, and the run id names the audit
	// directory, the FFmpeg source root, and the report file. A random suffix
	// keeps sequential runs' artifacts distinct while the leading timestamp stays
	// parseable for display (see LRecordLocalRead).
	LRunId := time.Now().UTC().Format("20060102T150405Z") + "-" + lRunTokenGet()
	return LRunId, LContextAction, nil
}

// lRunTokenGet returns a short random hex token used to keep run ids unique
// beyond one-second resolution. On the practically impossible chance the random
// source fails, it falls back to the nanosecond fraction of the current time.
func lRunTokenGet() string {
	token := make([]byte, 3)
	if _, err := rand.Read(token); err != nil {
		return time.Now().UTC().Format("000000000")
	}
	return hex.EncodeToString(token)
}

func (program *LProgram) LActionApprovedFinish(status string) {
	program.LMutexAction.Lock()
	if program.LActionCancelFunction == nil || program.lActionDone == nil {
		program.LMutexAction.Unlock()
		return
	}
	done := program.lActionDone
	program.LActionCancelFunction = nil
	program.LContextAction = nil
	program.LMutexAction.Unlock()
	program.LStatusEmit(status)
	if done != nil {
		close(done)
		program.LMutexAction.Lock()
		if program.lActionDone == done {
			program.lActionDone = nil
		}
		program.LMutexAction.Unlock()
	}
}

// lActionApprovedStop prevents new work, cancels the active action, and waits
// until its deferred cleanup, audit writes, and final status notification end.
func (program *LProgram) lActionApprovedStop() {
	program.LMutexAction.Lock()
	program.lProgramStopping = true
	cancel := program.LActionCancelFunction
	done := program.lActionDone
	if cancel != nil {
		cancel()
	}
	program.LMutexAction.Unlock()
	if cancel != nil {
		program.LLogEmit("warn", LLocaleTextGetInternal("logs.system.cancellationRequested", nil))
	}
	if done != nil {
		<-done
	}
}

func (program *LProgram) LAuditProgressCreate(auditWriter *audit.LAuditWriter, actionName string, planHash string) func(string, string) {
	return func(level string, message string) {
		_ = auditWriter.LAuditEventWrite("log", actionName, planHash, level, message)
		program.LLogEmit(level, message)
	}
}

func (program *LProgram) LStatusEmit(status string) {
	if program.LReporter != nil {
		program.LReporter.LReporterStatusEmit(status)
	}
}

func (program *LProgram) LLogEmit(level string, message string) {
	if program.LReporter != nil {
		program.LReporter.LReporterLogEmit(level, message)
	}
}

func (program *LProgram) lStalledEmit(addresses []string) {
	if program.LReporter != nil {
		program.LReporter.LReporterStalledEmit(addresses)
	}
}

func (program *LProgram) LStatusFailureEmit(message string, err error) {
	program.LLogEmit("error", message+": "+err.Error())
	program.LStatusEmit("failed")
}

func (program *LProgram) LErrorLocalizedEmit(messageKey string, fallback string, err error) {
	message := LLocaleTextGetInternal(messageKey, nil)
	if message == messageKey {
		message = fallback
	}
	program.LStatusFailureEmit(message, err)
}

func LPolicyHashResolve(expectedSha256Hash string) download.LPolicyFile {
	if expectedSha256Hash == "" {
		return download.LPolicyFileOverwrite
	}
	return download.LHashReusePolicy
}
