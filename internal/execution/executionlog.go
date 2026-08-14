package execution

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"promptfulcustomffmpegbuilder/internal/workspace"
	"promptfulcustomffmpegbuilder/localization"
)

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

func LLogCommandCopy(pipeReader interface{ Read([]byte) (int, error) }, logFile *os.File, level string, emitProgress LProgressFunc, transientFailureSeen *atomic.Bool, lastErrorLine *atomic.Pointer[string], addressCollector *LNetworkAddressCollector, doneChannel chan<- struct{}) {
	defer func() { doneChannel <- struct{}{} }()
	scanner := bufio.NewScanner(pipeReader)
	// Build/link output lines can far exceed the default 64 KB scan limit (a linker
	// invocation echoing hundreds of object files, for example). Grow the buffer so
	// Scan does not stop early on ErrTooLong and silently drop the rest of the pipe,
	// which would truncate the on-disk log and could hide a later error or transient-
	// network marker that drives classification and retry.
	scanner.Buffer(make([]byte, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if logFile != nil {
			_, _ = logFile.WriteString(line + "\n")
		}
		if transientFailureSeen != nil && LLogNetworkCheck(line) {
			transientFailureSeen.Store(true)
			addressCollector.LNetworkHostAdd(LNetworkHostParse(line))
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
	// A read error (including an over-limit line beyond the grown buffer) ends the
	// loop with output still on the pipe. Surface it so the truncation is visible
	// rather than mistaken for clean end-of-stream.
	if err := scanner.Err(); err != nil && emitProgress != nil {
		emitProgress("warn", localization.LLocaleTextGet("run.log.commandLogTruncated", map[string]string{"message": err.Error()}))
	}
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
