package execution

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync/atomic"
	"time"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/hostexec"
	"promptfulcustomffmpegbuilder/localization"
)

// Transient network failures (a stalled download, a dropped connection, an
// unresolved host) abort an otherwise-healthy command. Because pacman resumes
// from its package cache, make resumes from existing object files, configure is
// idempotent, and source clones are individually guarded, re-running the whole
// command is safe. These constants bound how many times and how long apart a
// command is retried before its failure is treated as real.
const (
	LCommandAttemptMax         = 10
	LCommandInitialDelay       = 5 * time.Second
	LCommandRetryBackoffFactor = 2
	LCommandMaximumDelay       = 60 * time.Second
)

type LScriptKind string

const (
	LPacmanInstallScript      LScriptKind = "pacman-install"
	LScriptFfmpegConfigure    LScriptKind = "ffmpeg-configure"
	LFfmpegMakeScript         LScriptKind = "ffmpeg-make"
	LScriptLibraryPreparation LScriptKind = "library-preparation"
)

type LPlanCommand struct {
	ActionName                 string            `json:"actionName"`
	PlanHash                   string            `json:"planHash"`
	ExecutablePath             string            `json:"executablePath"`
	ArgumentValues             []string          `json:"argumentValues"`
	WorkingDirectory           string            `json:"workingDirectory"`
	WorkspaceDirectory         string            `json:"workspaceDirectory"`
	Msys2RootDirectory         string            `json:"msys2RootDirectory"`
	WindowsShellProfileName    string            `json:"windowsShellProfileName"`
	EnvironmentVariables       map[string]string `json:"environmentVariables"`
	AllowedExecutableBasenames []string          `json:"allowedExecutableBasenames"`
	LScriptKind                LScriptKind       `json:"approvedScriptKindName"`
	ApprovedScriptFilePath     string            `json:"approvedScriptFilePath"`
	ApprovedScriptSha256Hash   string            `json:"approvedScriptSha256Hash"`
	RunLAuditDirectoryGet      string            `json:"runLAuditDirectoryGet"`
}

type LProgressFunc func(level string, message string)

func LCommandConsentRun(LContext context.Context, userExternalLConsentCommand consent.LConsentCommand, commandPlan LPlanCommand, emitProgress LProgressFunc) error {
	if err := consent.LConsentCheck(userExternalLConsentCommand.LConsent, consent.LConsentKindCommand, commandPlan.ActionName, commandPlan.PlanHash); err != nil {
		return err
	}
	return LCommandRun(LContext, commandPlan, emitProgress)
}

func LCommandPacmanRun(LContext context.Context, userPacmanPackageInstallLConsent consent.LConsentPacman, commandPlan LPlanCommand, emitProgress LProgressFunc) error {
	if err := consent.LConsentCheck(userPacmanPackageInstallLConsent.LConsent, consent.LConsentKindPacman, commandPlan.ActionName, commandPlan.PlanHash); err != nil {
		return err
	}
	return LCommandRun(LContext, commandPlan, emitProgress)
}

func LCommandRun(LContext context.Context, commandPlan LPlanCommand, emitProgress LProgressFunc) error {
	if err := LPlanCommandValidate(commandPlan); err != nil {
		return err
	}
	var scriptBytes []byte
	if commandPlan.ApprovedScriptFilePath != "" {
		preparedScriptBytes, updatedArgumentValues, err := LScriptStdinPrepare(commandPlan)
		if err != nil {
			return err
		}
		scriptBytes = preparedScriptBytes
		commandPlan.ArgumentValues = updatedArgumentValues
	}
	stdoutLogFile, stderrLogFile, err := LLogCommandOpen(commandPlan.WorkspaceDirectory, commandPlan.RunLAuditDirectoryGet)
	if err != nil {
		return err
	}
	if stdoutLogFile != nil {
		defer stdoutLogFile.Close()
	}
	if stderrLogFile != nil {
		defer stderrLogFile.Close()
	}

	addressCollector := &LNetworkAddressCollector{}
	retryDelay := LCommandInitialDelay
	for attemptNumber := 1; ; attemptNumber++ {
		transientFailureSeen, runErr := LCommandAttemptRun(LContext, commandPlan, scriptBytes, stdoutLogFile, stderrLogFile, attemptNumber, LCommandAttemptMax, emitProgress, addressCollector)
		if runErr == nil {
			return nil
		}
		// A cancelled run is never a stall: surface its real error so the run maps
		// to a cancelled/failed state, not the retryable stalled state.
		if LContext.Err() != nil {
			return runErr
		}
		// A clearly non-transient failure (a genuine build/link/config error) is a
		// real failure regardless of the attempt budget.
		if !transientFailureSeen {
			return runErr
		}
		// Transient failure that outlasted the whole retry budget: halt in the
		// retryable stalled state and record the addresses that were tried.
		if attemptNumber >= LCommandAttemptMax {
			return LNetworkStalledCreate(runErr, addressCollector)
		}
		if emitProgress != nil {
			emitProgress("warn", fmt.Sprintf("Transient network failure detected (attempt %d of %d): %v. Retrying in %s...", attemptNumber, LCommandAttemptMax, runErr, retryDelay))
		}
		select {
		case <-LContext.Done():
			return runErr
		case <-time.After(retryDelay):
		}
		if retryDelay *= LCommandRetryBackoffFactor; retryDelay > LCommandMaximumDelay {
			retryDelay = LCommandMaximumDelay
		}
	}
}

// LCommandAttemptRun executes the planned command exactly once. It reports
// whether any streamed line looked like a transient network failure so the
// caller can decide to retry. Fresh pipes and a fresh stdin LReader are built
// per attempt because both are single-use.
func LCommandAttemptRun(LContext context.Context, commandPlan LPlanCommand, scriptBytes []byte, stdoutLogFile, stderrLogFile *os.File, attemptNumber int, attemptMax int, emitProgress LProgressFunc, addressCollector *LNetworkAddressCollector) (bool, error) {
	if emitProgress != nil {
		if attemptNumber > 1 {
			// Announce the current attempt before it runs so a user watching a slow
			// retry sees progress, not silence, until it succeeds or fails. warn
			// renders orange to mark it as a retry.
			emitProgress("warn", localization.LLocaleTextGet("run.log.commandRetryAttempt", map[string]string{
				"n":   fmt.Sprintf("%d", attemptNumber),
				"max": fmt.Sprintf("%d", attemptMax),
			}))
		} else {
			emitProgress("info", "Running approved command: "+filepath.Base(commandPlan.ExecutablePath))
		}
	}
	command := exec.CommandContext(LContext, commandPlan.ExecutablePath, commandPlan.ArgumentValues...)
	command.Dir = commandPlan.WorkingDirectory
	command.Env = LEnvironmentMsysCreate(commandPlan)
	hostexec.LCommandWindowHide(command)
	if scriptBytes != nil {
		command.Stdin = bytes.NewReader(scriptBytes)
	}

	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return false, err
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return false, err
	}
	if err := command.Start(); err != nil {
		return false, err
	}
	var transientFailureSeen atomic.Bool
	var lastErrorLine atomic.Pointer[string]
	doneChannel := make(chan struct{}, 2)
	go LLogCommandCopy(stdoutPipe, stdoutLogFile, "info", emitProgress, &transientFailureSeen, &lastErrorLine, addressCollector, doneChannel)
	go LLogCommandCopy(stderrPipe, stderrLogFile, "warn", emitProgress, &transientFailureSeen, &lastErrorLine, addressCollector, doneChannel)
	<-doneChannel
	<-doneChannel
	waitErr := command.Wait()
	// Turn an opaque "exit status 1" into something diagnosable by attaching the
	// last line that classified as an error (the compiler/configure/pacman line
	// that actually caused the failure). The full log is still on disk.
	if waitErr != nil {
		if errorLine := lastErrorLine.Load(); errorLine != nil {
			waitErr = fmt.Errorf("%w: %s", waitErr, *errorLine)
		}
	}
	return transientFailureSeen.Load(), waitErr
}
