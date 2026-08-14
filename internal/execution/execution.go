package execution

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"
	"time"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/hostexec"
	"promptfulcustomffmpegbuilder/internal/workspace"
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
	LScriptFFmpegConfigure    LScriptKind = "ffmpeg-configure"
	LFFmpegMakeScript         LScriptKind = "ffmpeg-make"
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

	retryDelay := LCommandInitialDelay
	for attemptNumber := 1; ; attemptNumber++ {
		transientFailureSeen, runErr := LCommandAttemptRun(LContext, commandPlan, scriptBytes, stdoutLogFile, stderrLogFile, emitProgress)
		if runErr == nil {
			return nil
		}
		// Never retry a cancelled run, a clearly non-transient failure, or once
		// the attempt budget is spent. Surface the real error in those cases.
		if LContext.Err() != nil || !transientFailureSeen || attemptNumber >= LCommandAttemptMax {
			return runErr
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
func LCommandAttemptRun(LContext context.Context, commandPlan LPlanCommand, scriptBytes []byte, stdoutLogFile, stderrLogFile *os.File, emitProgress LProgressFunc) (bool, error) {
	if emitProgress != nil {
		emitProgress("info", "Running approved command: "+filepath.Base(commandPlan.ExecutablePath))
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
	go LLogCommandCopy(stdoutPipe, stdoutLogFile, "info", emitProgress, &transientFailureSeen, &lastErrorLine, doneChannel)
	go LLogCommandCopy(stderrPipe, stderrLogFile, "warn", emitProgress, &transientFailureSeen, &lastErrorLine, doneChannel)
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

func LPlanCommandValidate(commandPlan LPlanCommand) error {
	if commandPlan.ExecutablePath == "" {
		return errors.New("approved command executable path is empty")
	}
	if commandPlan.WorkingDirectory == "" || commandPlan.WorkspaceDirectory == "" {
		return errors.New("approved command directories are empty")
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if len(commandPlan.AllowedExecutableBasenames) == 0 {
		return errors.New("approved command must include executable basename allowlist")
	}
	if !LProgramAllowedCheck(filepath.Base(commandPlan.ExecutablePath), commandPlan.AllowedExecutableBasenames) {
		return fmt.Errorf("approved command executable is not allowlisted: %s", filepath.Base(commandPlan.ExecutablePath))
	}
	if commandPlan.Msys2RootDirectory != "" {
		if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
		if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
	}
	if commandPlan.RunLAuditDirectoryGet != "" {
		if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.RunLAuditDirectoryGet); err != nil {
			return err
		}
		if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, filepath.Dir(commandPlan.RunLAuditDirectoryGet)); err != nil {
			return err
		}
	}
	if LTextShellCheck(commandPlan.ExecutablePath) {
		return errors.New("approved command executable contains shell metacharacters")
	}
	for _, argumentValue := range commandPlan.ArgumentValues {
		if strings.Contains(argumentValue, "\x00") {
			return errors.New("approved command argument contains null byte")
		}
	}
	if commandPlan.ApprovedScriptFilePath != "" {
		if commandPlan.LScriptKind == "" {
			return errors.New("approved script kind is empty")
		}
		if err := LHashScriptCheck(commandPlan); err != nil {
			return err
		}
	}
	return nil
}

func LHashScriptCheck(commandPlan LPlanCommand) error {
	_, _, err := LScriptStdinPrepare(commandPlan)
	return err
}

func LScriptStdinPrepare(commandPlan LPlanCommand) ([]byte, []string, error) {
	if commandPlan.ApprovedScriptSha256Hash == "" {
		return nil, nil, errors.New("approved script hash is empty")
	}
	if err := workspace.LPathWorkspaceCheck(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	if err := workspace.LPathRealCheck(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	updatedArgumentValues, foundScriptArgument := LStdinFlagReplace(commandPlan.ArgumentValues, commandPlan.ApprovedScriptFilePath)
	if !foundScriptArgument {
		return nil, nil, errors.New("approved script path is not present in command arguments")
	}
	file, err := os.Open(commandPlan.ApprovedScriptFilePath)
	if err != nil {
		return nil, nil, err
	}
	defer file.Close()
	scriptBytes, err := io.ReadAll(file)
	if err != nil {
		return nil, nil, err
	}
	hash := sha256.Sum256(scriptBytes)
	actualScriptSha256Hash := hex.EncodeToString(hash[:])
	if !strings.EqualFold(actualScriptSha256Hash, commandPlan.ApprovedScriptSha256Hash) {
		return nil, nil, errors.New("approved script hash does not match script content")
	}
	return scriptBytes, updatedArgumentValues, nil
}

func LEnvironmentMsysCreate(commandPlan LPlanCommand) []string {
	environmentByName := map[string]string{}
	environmentNameOrder := []string{}
	setEnvironmentValue := func(environmentName string, environmentValue string) {
		if _, exists := environmentByName[environmentName]; !exists {
			environmentNameOrder = append(environmentNameOrder, environmentName)
		}
		environmentByName[environmentName] = environmentValue
	}

	for _, inheritedEnvironmentValue := range os.Environ() {
		environmentName, environmentValue, found := strings.Cut(inheritedEnvironmentValue, "=")
		if !found || environmentName == "" {
			continue
		}
		setEnvironmentValue(environmentName, environmentValue)
	}

	setEnvironmentValue("MSYS2_PATH_TYPE", "minimal")
	setEnvironmentValue("CHERE_INVOKING", "1")

	if commandPlan.Msys2RootDirectory != "" && commandPlan.WindowsShellProfileName != "" {
		profileDirectoryName := strings.ToLower(commandPlan.WindowsShellProfileName)
		shellSystemName := strings.ToUpper(profileDirectoryName)
		existingPath := environmentByName["PATH"]
		msys2Path := filepath.Join(commandPlan.Msys2RootDirectory, profileDirectoryName, "bin") + string(os.PathListSeparator) + filepath.Join(commandPlan.Msys2RootDirectory, "usr", "bin")
		if existingPath != "" {
			msys2Path += string(os.PathListSeparator) + existingPath
		}

		msys2UnixPrefix := "/" + profileDirectoryName
		mingwPackagePrefix := "mingw-w64-" + profileDirectoryName + "-x86_64"
		mingwChost := "x86_64-w64-mingw32"
		if profileDirectoryName == "clangarm64" {
			mingwPackagePrefix = "mingw-w64-clang-aarch64"
			mingwChost = "aarch64-w64-mingw32"
		} else if profileDirectoryName == "clang64" {
			mingwPackagePrefix = "mingw-w64-clang-x86_64"
		} else if profileDirectoryName == "mingw64" {
			mingwPackagePrefix = "mingw-w64-x86_64"
		}

		setEnvironmentValue("MSYSTEM", shellSystemName)
		setEnvironmentValue("MSYSTEM_PREFIX", msys2UnixPrefix)
		setEnvironmentValue("MINGW_PREFIX", msys2UnixPrefix)
		setEnvironmentValue("MINGW_CHOST", mingwChost)
		setEnvironmentValue("MINGW_PACKAGE_PREFIX", mingwPackagePrefix)
		setEnvironmentValue("PKG_CONFIG_PATH", msys2UnixPrefix+"/lib/pkgconfig:"+msys2UnixPrefix+"/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig")
		setEnvironmentValue("PKG_CONFIG_LIBDIR", msys2UnixPrefix+"/lib/pkgconfig:"+msys2UnixPrefix+"/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig")
		setEnvironmentValue("PATH", msys2Path)
	}

	for environmentName, environmentValue := range commandPlan.EnvironmentVariables {
		if !LEnvironmentSafeCheck(environmentName, environmentValue) {
			continue
		}
		setEnvironmentValue(environmentName, environmentValue)
	}

	environment := make([]string, 0, len(environmentNameOrder))
	for _, environmentName := range environmentNameOrder {
		environment = append(environment, fmt.Sprintf("%s=%s", environmentName, environmentByName[environmentName]))
	}
	return environment
}

func LLogCommandOpen(workspaceDirectory string, runLAuditDirectoryGet string) (*os.File, *os.File, error) {
	if runLAuditDirectoryGet == "" {
		return nil, nil, nil
	}
	if err := workspace.LPathWorkspaceCheck(workspaceDirectory, runLAuditDirectoryGet); err != nil {
		return nil, nil, err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, filepath.Dir(runLAuditDirectoryGet)); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(runLAuditDirectoryGet, 0o755); err != nil {
		return nil, nil, err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, runLAuditDirectoryGet); err != nil {
		return nil, nil, err
	}
	stdoutLogFile, err := os.OpenFile(filepath.Join(runLAuditDirectoryGet, "stdout.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}
	stderrLogFile, err := os.OpenFile(filepath.Join(runLAuditDirectoryGet, "stderr.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		_ = stdoutLogFile.Close()
		return nil, nil, err
	}
	return stdoutLogFile, stderrLogFile, nil
}

func LEnvironmentSafeCheck(environmentName string, environmentValue string) bool {
	if environmentName == "" || strings.Contains(environmentName, "=") || strings.Contains(environmentName, "\x00") || strings.Contains(environmentValue, "\x00") {
		return false
	}
	for _, environmentNameRune := range environmentName {
		if !(environmentNameRune == '_' || environmentNameRune >= 'A' && environmentNameRune <= 'Z' || environmentNameRune >= 'a' && environmentNameRune <= 'z' || environmentNameRune >= '0' && environmentNameRune <= '9') {
			return false
		}
	}
	return true
}

func LTextShellCheck(value string) bool {
	return strings.ContainsAny(value, ";&|><`$\n\r")
}

func LProgramAllowedCheck(executableBasename string, allowedExecutableBasenames []string) bool {
	for _, allowedExecutableBasename := range allowedExecutableBasenames {
		if strings.EqualFold(executableBasename, allowedExecutableBasename) {
			return true
		}
	}
	return false
}

func LStdinFlagReplace(argumentValues []string, scriptFilePath string) ([]string, bool) {
	updatedArgumentValues := make([]string, len(argumentValues))
	copy(updatedArgumentValues, argumentValues)
	for index, argumentValue := range updatedArgumentValues {
		if argumentValue == scriptFilePath || argumentValue == filepath.ToSlash(scriptFilePath) {
			updatedArgumentValues[index] = "-s"
			return updatedArgumentValues, true
		}
	}
	return updatedArgumentValues, false
}

func LLogCommandCopy(pipeReader interface{ Read([]byte) (int, error) }, logFile *os.File, level string, emitProgress LProgressFunc, transientFailureSeen *atomic.Bool, lastErrorLine *atomic.Pointer[string], doneChannel chan<- struct{}) {
	defer func() { doneChannel <- struct{}{} }()
	scanner := bufio.NewScanner(pipeReader)
	for scanner.Scan() {
		line := scanner.Text()
		if logFile != nil {
			_, _ = logFile.WriteString(line + "\n")
		}
		if transientFailureSeen != nil && LLogNetworkCheck(line) {
			transientFailureSeen.Store(true)
		}
		classifiedLevel := LLogLineGet(level, line)
		if lastErrorLine != nil && classifiedLevel == "error" {
			capturedLine := strings.TrimSpace(line)
			if !LLogFailureGet(capturedLine) || lastErrorLine.Load() == nil {
				lastErrorLine.Store(&capturedLine)
			}
		}
		if emitProgress != nil {
			emitProgress(classifiedLevel, line)
		}
	}
}

// LErrorNetworkMarkers are substrings (matched case-insensitively)
// that signal a download or connection failed for a transient reason rather
// than a real build/install error: a stalled transfer, a dropped or refused
// connection, DNS failure, or a 5xx from a mirror. A line carrying any of these
// makes the whole command eligible for retry. Markers are kept specific so a
// genuine compile/link error is never mistaken for a network blip.
var LErrorNetworkMarkers = []string{
	"operation too slow",
	"failed retrieving file",
	"could not resolve host",
	"name or service not known",
	"temporary failure in name resolution",
	"connection timed out",
	"connection refused",
	"connection reset",
	"network is unreachable",
	"transfer closed",
	"timeout was reached",
	"failed to commit transaction (unexpected error)",
	"rpc failed",
	"early eof",
	"the remote end hung up unexpectedly",
	"gnutls recv error",
	"ssl_read",
	"unexpected disconnect",
}

// LLogNetworkCheck reports whether a streamed output line indicates
// a transient network failure that warrants retrying the command.
func LLogNetworkCheck(line string) bool {
	lower := strings.ToLower(line)
	for _, marker := range LErrorNetworkMarkers {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

// LCompilerEchoPattern matches the source-echo and caret lines GCC prints
// beneath a diagnostic, such as "297 |     memmove(...)" and "    | ^~~~". These
// arrive on stderr but are continuation context, not warnings of their own.
var LCompilerEchoPattern = regexp.MustCompile(`^\s*(?:\d+\s*)?\|`)

// LLogLineGet refines the severity of a streamed build-output line from its
// content. The raw pipe gives only a coarse default: every stderr line would
// otherwise be a "warn", burying genuine warnings under compiler notes,
// source-echo lines, "#pragma message" output, and pacman reinstall notices.
// This promotes real errors and demotes known-benign noise so the UI's warning
// level reflects lines that actually warrant attention. The full raw line is
// still written verbatim to the on-disk log regardless of level.
func LLogLineGet(defaultLevel string, line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return defaultLevel
	}
	lower := strings.ToLower(trimmed)

	// A failed retrieval of a repository database (.db/.files) during refresh is
	// non-fatal: pacman falls back to other repos and the cached database, and a
	// repo not needed by the selected profile (e.g. clang64 for a mingw64 build)
	// does not affect the install. Demote to warn so it does not read as a failure.
	// Failed *package* (.pkg) retrievals are not demoted and stay errors.
	if strings.Contains(lower, "failed retrieving file") &&
		(strings.Contains(lower, ".db") || strings.Contains(lower, ".files")) {
		return "warn"
	}

	// A make recipe prefixed with '-' prints "[Makefile:NN: target] Error 1
	// (ignored)" and keeps going; it is non-fatal continuation noise, not the
	// failure. Demote so it neither reads as an error nor is mistaken for the
	// cause attached to an exit status. Checked before the error block below.
	if strings.Contains(line, "(ignored)") && strings.Contains(line, "] Error ") {
		return "info"
	}

	// "strip: ... has no section(s)" is binutils refusing an intentionally-empty object:
	// x264/xavs2-style build systems assemble a 32-bit-only .asm (e.g. pixel-32.asm) even
	// in a 64-bit build, where every symbol is guarded out, yielding a 0-section object the
	// makefile then strips with an ignored ('-') rule. The build succeeds and the empty
	// object links harmlessly, so this strip message is benign tool noise, not a failure.
	if strings.Contains(lower, "strip") && strings.Contains(lower, "has no section") {
		return "info"
	}

	// Genuine failures: compiler/linker/configure/pacman errors and aborted make.
	if strings.Contains(lower, "error:") ||
		strings.Contains(lower, "argument list too long") ||
		strings.Contains(lower, "undefined reference") ||
		strings.HasPrefix(lower, "collect2:") ||
		strings.Contains(line, "] Error ") {
		return "error"
	}

	// Benign noise that is not a real warning. Demote to info so it does not
	// drown out actual warnings.
	if strings.Contains(line, "is up to date -- reinstalling") ||
		strings.Contains(line, "dependency cycle detected") ||
		strings.Contains(line, "will be installed before its") ||
		strings.Contains(lower, "note:") ||
		strings.Contains(line, "#pragma message") ||
		strings.HasPrefix(trimmed, "In file included from") ||
		strings.HasPrefix(trimmed, "In function") ||
		strings.HasPrefix(trimmed, "inlined from") ||
		strings.Contains(line, ": In function") ||
		LWarningThirdpartyCheck(lower) ||
		LCompilerEchoPattern.MatchString(line) {
		return "info"
	}

	// Genuine compiler/tool warnings stay (or are raised to) warn so they remain
	// visible even when they arrive on stdout.
	if strings.Contains(lower, "warning:") {
		return "warn"
	}

	return defaultLevel
}

// LWarningThirdpartyCheck matches compiler/assembler warnings that flood the log
// when building Internal-track libraries from their own upstream source (uavs3d, davs2,
// etc.) but are not actionable here: unused symbols, MSVC-only #pragma warning lines,
// macro/type redefinitions, and x264/nasm legacy macro-parameter warnings. They are
// demoted to info so genuine warnings stay visible. Deliberately NOT included:
// stringop-overflow and similar correctness warnings, which can flag real bugs and stay
// at warn. `lower` is the already-lowercased line.
func LWarningThirdpartyCheck(lower string) bool {
	for _, marker := range []string{
		"-wunused-variable",
		"-wunused-but-set-variable",
		"-wunused-function",
		"-wunknown-pragmas",
		"pp-macro-params-legacy",
		"ignoring '#pragma",
		"' redefined",
	} {
		if strings.Contains(lower, marker) {
			return true
		}
	}
	return false
}

func LLogFailureGet(line string) bool {
	trimmed := strings.TrimSpace(line)
	lower := strings.ToLower(trimmed)
	return strings.HasPrefix(lower, "make: ***") || strings.Contains(trimmed, "] Error ")
}
