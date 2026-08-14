package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"promptfulcustomffmpegbuilder/internal/reporting"
)

// Compile-time checks that the CLI implementations satisfy the shared seams.
var (
	_ reporting.LReporter  = LReporterConsole{}
	_ reporting.LConfirmer = LConfirmerConsole{}
)

// LReporterConsole is the CLI reporting.LReporter. Status lines and non-error
// logs go to stdout; warn/error logs go to stderr. When quiet is set, only
// warn/error logs are printed.
type LReporterConsole struct {
	quiet bool
}

func (reporter LReporterConsole) LReporterStatusEmit(status string) {
	if reporter.quiet {
		return
	}
	fmt.Fprintf(os.Stdout, "[status] %s\n", status)
}

func (reporter LReporterConsole) LReporterLogEmit(level string, message string) {
	if level == "warn" || level == "error" {
		fmt.Fprintf(os.Stderr, "[%s] %s\n", level, message)
		return
	}
	if reporter.quiet {
		return
	}
	fmt.Fprintf(os.Stdout, "[%s] %s\n", level, message)
}

// LReporterStalledEmit is a no-op on the console: the tried addresses already
// reach the operator through the warn-level "downloadStalled" log line, which
// this reporter prints to stderr. The structured list exists for the GUI banner.
func (reporter LReporterConsole) LReporterStalledEmit(addresses []string) {}

// LConfirmerConsole is the CLI reporting.LConfirmer. It approves automatically
// when assumeYes is set (--yes), fails without prompting when noInput is set
// (--no-input), and otherwise reads a y/N answer from input.
type LConfirmerConsole struct {
	assumeYes bool
	noInput   bool
	input     io.Reader
	output    io.Writer
}

func (confirmer LConfirmerConsole) LConfirmerApprovalGet(actionName string, planHash string) (bool, error) {
	if confirmer.assumeYes {
		return true, nil
	}
	if confirmer.noInput {
		return false, errors.New("approval required but --no-input is set")
	}
	output := confirmer.output
	if output == nil {
		output = os.Stdout
	}
	input := confirmer.input
	if input == nil {
		input = os.Stdin
	}
	fmt.Fprintf(output, "Approve %s (plan %s)? [y/N] ", actionName, planHash)
	reader := bufio.NewReader(input)
	line, err := reader.ReadString('\n')
	if err != nil && line == "" {
		return false, err
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}
