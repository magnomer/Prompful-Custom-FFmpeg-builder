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

	"customffmpegbuilder/internal/consent"
	"customffmpegbuilder/internal/workspace"
)

type ScriptKind string

const (
	PacmanInstallScript   ScriptKind = "pacman-install"
	FfmpegConfigureScript ScriptKind = "ffmpeg-configure"
	FfmpegMakeScript      ScriptKind = "ffmpeg-make"
)

type CommandPlan struct {
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
	ScriptKind                 ScriptKind        `json:"approvedScriptKindName"`
	ApprovedScriptFilePath     string            `json:"approvedScriptFilePath"`
	ApprovedScriptSha256Hash   string            `json:"approvedScriptSha256Hash"`
	RunLogDirectory            string            `json:"runLogDirectory"`
}

type ProgressFunc func(level string, message string)

func RunCommandWithConsent(ctx context.Context, userExternalCommandExecutionConsent consent.CommandExecutionConsent, commandPlan CommandPlan, emitProgress ProgressFunc) error {
	if err := consent.CheckConsent(userExternalCommandExecutionConsent.Consent, consent.ConsentKindExternalCommandExecution, commandPlan.ActionName, commandPlan.PlanHash); err != nil {
		return err
	}
	return executeCommand(ctx, commandPlan, emitProgress)
}

func RunPacmanWithConsent(ctx context.Context, userPacmanPackageInstallConsent consent.PacmanInstallConsent, commandPlan CommandPlan, emitProgress ProgressFunc) error {
	if err := consent.CheckConsent(userPacmanPackageInstallConsent.Consent, consent.ConsentKindPacmanPackageInstallation, commandPlan.ActionName, commandPlan.PlanHash); err != nil {
		return err
	}
	return executeCommand(ctx, commandPlan, emitProgress)
}

func executeCommand(ctx context.Context, commandPlan CommandPlan, emitProgress ProgressFunc) error {
	if err := ValidateCommandPlan(commandPlan); err != nil {
		return err
	}
	var stdinReader io.Reader
	if commandPlan.ApprovedScriptFilePath != "" {
		scriptBytes, updatedArgumentValues, err := prepareScriptForStdin(commandPlan)
		if err != nil {
			return err
		}
		stdinReader = bytes.NewReader(scriptBytes)
		commandPlan.ArgumentValues = updatedArgumentValues
	}
	if emitProgress != nil {
		emitProgress("info", "Running approved command: "+filepath.Base(commandPlan.ExecutablePath))
	}
	command := exec.CommandContext(ctx, commandPlan.ExecutablePath, commandPlan.ArgumentValues...)
	command.Dir = commandPlan.WorkingDirectory
	command.Env = createMsys2Env(commandPlan)
	if stdinReader != nil {
		command.Stdin = stdinReader
	}

	stdoutPipe, err := command.StdoutPipe()
	if err != nil {
		return err
	}
	stderrPipe, err := command.StderrPipe()
	if err != nil {
		return err
	}
	stdoutLogFile, stderrLogFile, err := openCommandLogs(commandPlan.WorkspaceDirectory, commandPlan.RunLogDirectory)
	if err != nil {
		return err
	}
	if stdoutLogFile != nil {
		defer stdoutLogFile.Close()
	}
	if stderrLogFile != nil {
		defer stderrLogFile.Close()
	}
	if err := command.Start(); err != nil {
		return err
	}
	doneChannel := make(chan struct{}, 2)
	go copyCommandOutput(stdoutPipe, stdoutLogFile, "info", emitProgress, doneChannel)
	go copyCommandOutput(stderrPipe, stderrLogFile, "warn", emitProgress, doneChannel)
	<-doneChannel
	<-doneChannel
	return command.Wait()
}

func ValidateCommandPlan(commandPlan CommandPlan) error {
	if commandPlan.ExecutablePath == "" {
		return errors.New("approved command executable path is empty")
	}
	if commandPlan.WorkingDirectory == "" || commandPlan.WorkspaceDirectory == "" {
		return errors.New("approved command directories are empty")
	}
	if err := workspace.CheckPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.WorkingDirectory); err != nil {
		return err
	}
	if err := workspace.CheckPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.ExecutablePath); err != nil {
		return err
	}
	if len(commandPlan.AllowedExecutableBasenames) == 0 {
		return errors.New("approved command must include executable basename allowlist")
	}
	if !isAllowedProgramName(filepath.Base(commandPlan.ExecutablePath), commandPlan.AllowedExecutableBasenames) {
		return fmt.Errorf("approved command executable is not allowlisted: %s", filepath.Base(commandPlan.ExecutablePath))
	}
	if commandPlan.Msys2RootDirectory != "" {
		if err := workspace.CheckPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
		if err := workspace.CheckRealPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.Msys2RootDirectory); err != nil {
			return err
		}
	}
	if commandPlan.RunLogDirectory != "" {
		if err := workspace.CheckPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.RunLogDirectory); err != nil {
			return err
		}
		if err := workspace.CheckRealPathInsideWorkspace(commandPlan.WorkspaceDirectory, filepath.Dir(commandPlan.RunLogDirectory)); err != nil {
			return err
		}
	}
	if hasShellMetacharacter(commandPlan.ExecutablePath) {
		return errors.New("approved command executable contains shell metacharacters")
	}
	for _, argumentValue := range commandPlan.ArgumentValues {
		if strings.Contains(argumentValue, "\x00") {
			return errors.New("approved command argument contains null byte")
		}
	}
	if commandPlan.ApprovedScriptFilePath != "" {
		if commandPlan.ScriptKind == "" {
			return errors.New("approved script kind is empty")
		}
		if err := checkScriptHash(commandPlan); err != nil {
			return err
		}
	}
	return nil
}

func checkScriptHash(commandPlan CommandPlan) error {
	_, _, err := prepareScriptForStdin(commandPlan)
	return err
}

func prepareScriptForStdin(commandPlan CommandPlan) ([]byte, []string, error) {
	if commandPlan.ApprovedScriptSha256Hash == "" {
		return nil, nil, errors.New("approved script hash is empty")
	}
	if err := workspace.CheckPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(commandPlan.WorkspaceDirectory, commandPlan.ApprovedScriptFilePath); err != nil {
		return nil, nil, err
	}
	updatedArgumentValues, foundScriptArgument := replaceScriptFileWithStdinFlag(commandPlan.ArgumentValues, commandPlan.ApprovedScriptFilePath)
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

func createMsys2Env(commandPlan CommandPlan) []string {
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
		if !isSafeEnvironmentVariable(environmentName, environmentValue) {
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

func openCommandLogs(workspaceDirectory string, runLogDirectory string) (*os.File, *os.File, error) {
	if runLogDirectory == "" {
		return nil, nil, nil
	}
	if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, runLogDirectory); err != nil {
		return nil, nil, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, filepath.Dir(runLogDirectory)); err != nil {
		return nil, nil, err
	}
	if err := os.MkdirAll(runLogDirectory, 0o755); err != nil {
		return nil, nil, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, runLogDirectory); err != nil {
		return nil, nil, err
	}
	stdoutLogFile, err := os.OpenFile(filepath.Join(runLogDirectory, "stdout.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, nil, err
	}
	stderrLogFile, err := os.OpenFile(filepath.Join(runLogDirectory, "stderr.log"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		_ = stdoutLogFile.Close()
		return nil, nil, err
	}
	return stdoutLogFile, stderrLogFile, nil
}

func isSafeEnvironmentVariable(environmentName string, environmentValue string) bool {
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

func hasShellMetacharacter(value string) bool {
	return strings.ContainsAny(value, ";&|><`$\n\r")
}

func isAllowedProgramName(executableBasename string, allowedExecutableBasenames []string) bool {
	for _, allowedExecutableBasename := range allowedExecutableBasenames {
		if strings.EqualFold(executableBasename, allowedExecutableBasename) {
			return true
		}
	}
	return false
}

func replaceScriptFileWithStdinFlag(argumentValues []string, scriptFilePath string) ([]string, bool) {
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

func copyCommandOutput(pipeReader interface{ Read([]byte) (int, error) }, logFile *os.File, level string, emitProgress ProgressFunc, doneChannel chan<- struct{}) {
	defer func() { doneChannel <- struct{}{} }()
	scanner := bufio.NewScanner(pipeReader)
	for scanner.Scan() {
		line := scanner.Text()
		if logFile != nil {
			_, _ = logFile.WriteString(line + "\n")
		}
		if emitProgress != nil {
			emitProgress(classifyLogLine(level, line), line)
		}
	}
}

// compilerSourceEchoRegex matches the source-echo and caret lines GCC prints
// beneath a diagnostic, such as "297 |     memmove(...)" and "    | ^~~~". These
// arrive on stderr but are continuation context, not warnings of their own.
var compilerSourceEchoRegex = regexp.MustCompile(`^\s*(?:\d+\s*)?\|`)

// classifyLogLine refines the severity of a streamed build-output line from its
// content. The raw pipe gives only a coarse default: every stderr line would
// otherwise be a "warn", burying genuine warnings under compiler notes,
// source-echo lines, "#pragma message" output, and pacman reinstall notices.
// This promotes real errors and demotes known-benign noise so the UI's warning
// level reflects lines that actually warrant attention. The full raw line is
// still written verbatim to the on-disk log regardless of level.
func classifyLogLine(defaultLevel string, line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return defaultLevel
	}
	lower := strings.ToLower(trimmed)

	// Genuine failures: compiler/linker/configure/pacman errors and aborted make.
	if strings.Contains(lower, "error:") ||
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
		compilerSourceEchoRegex.MatchString(line) {
		return "info"
	}

	// Genuine compiler/tool warnings stay (or are raised to) warn so they remain
	// visible even when they arrive on stdout.
	if strings.Contains(lower, "warning:") {
		return "warn"
	}

	return defaultLevel
}
