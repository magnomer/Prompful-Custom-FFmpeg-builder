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
	case "list":
		return cmdList(args[1:])
	case "plan":
		return cmdPlan(args[1:])
	case "build":
		return cmdBuild(args[1:])
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

Commands:
  list       List known versions, presets, or libraries
  plan       Show the resolved build plan (no build is run)
  build      Build FFmpeg (workspace toolchain must be prepared first)
  verify     Verify a built FFmpeg executable    (not implemented yet)
  explain    Explain a version, library, or preset (not implemented yet)

Plan / build options:
  --ffmpeg-version X     required; a supported release (e.g. 8.1.2)
  --preset P             start from a preset (e.g. full, minimal)
  --extended             use the preset's extended library set
  --enable-libNAME       add an FFmpeg library (e.g. --enable-libx264)
  --disable-libNAME      remove a library from the preset
  --workspace DIR        build workspace directory (absolute)
  --jobs N               parallel build jobs
  --yes                  accept the confirmation automatically (build)
  --no-input             never prompt; fail if input would be needed (build)

Examples:
  promptfulx list versions
  promptfulx list libraries --ffmpeg-version 8.1.2
  promptfulx plan --ffmpeg-version 8.1.2 --preset full --disable-liboapv
  promptfulx build --ffmpeg-version 8.1.2 --preset full --workspace D:\Work --yes

See docs/internal/PlanCLI.md.`)
}
