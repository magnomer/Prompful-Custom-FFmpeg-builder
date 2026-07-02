// Command promptfulx is the Promptful CLI (PromptfulX), the second standalone
// Promptful executable alongside the root GUI binary. Step 0 establishes this
// entry point only; commands are added in later steps of
// docs/internal/PlanCLI.md. The GUI keeps building via the wails CLI at the
// module root; this binary builds with `go build ./cmd/promptfulx`.
package main

import (
	"fmt"
	"os"
)

func main() { os.Exit(run(os.Args[1:])) }

// run dispatches CLI arguments and returns a process exit code. Exit codes
// follow docs/internal/PlanCLI.md (2 = invalid arguments).
func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 0
	}
	switch args[0] {
	case "-h", "--help", "help":
		printUsage()
		return 0
	default:
		fmt.Fprintln(os.Stderr, "promptfulx: unknown command:", args[0])
		fmt.Fprintln(os.Stderr, "run 'promptfulx --help' for usage")
		return 2
	}
}

func printUsage() {
	fmt.Println(`PromptfulX - Promptful CLI

Usage:
  promptfulx <command> [options]

Commands (planned):
  plan       Show the resolved build plan
  build      Build FFmpeg
  list       List known versions, libraries, or presets
  verify     Verify a built FFmpeg executable
  explain    Explain a version, library, or preset

Status: skeleton. See docs/internal/PlanCLI.md.`)
}
