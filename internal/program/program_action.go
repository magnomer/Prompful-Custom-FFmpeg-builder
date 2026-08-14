package program

import (
	"context"
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
	if program.LActionCancelFunction != nil {
		return "", nil, errors.New("an approved action is already running")
	}
	LContextAction, LActionCancelFunction := context.WithCancel(context.Background())
	program.LContextAction = LContextAction
	program.LActionCancelFunction = LActionCancelFunction
	LRunId := time.Now().UTC().Format("20060102T150405Z")
	return LRunId, LContextAction, nil
}

func (program *LProgram) LActionApprovedFinish(status string) {
	program.LMutexAction.Lock()
	program.LActionCancelFunction = nil
	program.LContextAction = nil
	program.LMutexAction.Unlock()
	program.LStatusEmit(status)
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
