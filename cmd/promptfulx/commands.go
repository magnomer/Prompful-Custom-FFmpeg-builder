package main

import (
	"fmt"
	"os"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
)

// exitFor maps an error to a process exit code: a usageError carries its own
// code, anything else is a general failure (1).
func exitFor(err error) int {
	var usage usageError
	if e, ok := err.(usageError); ok {
		usage = e
		fmt.Fprintln(os.Stderr, "promptfulx:", usage.message)
		return usage.code
	}
	fmt.Fprintln(os.Stderr, "promptfulx:", err.Error())
	return 1
}

// cmdPlan resolves CLI arguments into a build plan and prints it. It never
// modifies the build environment.
func cmdPlan(args []string) int {
	parsed, err := argsParse(args)
	if err != nil {
		return exitFor(err)
	}
	settings, err := settingsResolve(parsed)
	if err != nil {
		return exitFor(err)
	}
	plan, err := planning.LPlanFFmpegCreate(settings)
	if err != nil {
		return exitFor(unsupported("could not resolve plan: %v", err))
	}
	printPlan(plan)
	if !plan.IsExecutable {
		return 4 // resolved, but blocked by an unsupported combination
	}
	return 0
}

func printPlan(plan planning.LPlanFFmpeg) {
	version := plan.CompatibilityFfmpegVersion
	if version == "" {
		version = plan.RequestedFfmpegVersion
	}
	fmt.Println("FFmpeg build plan")
	fmt.Println("  version:   ", version)
	fmt.Println("  source:    ", plan.FfmpegSourceArchiveUrl)
	if plan.WorkspaceDirectory != "" {
		fmt.Println("  workspace: ", plan.WorkspaceDirectory)
	}
	fmt.Println("  jobs:      ", plan.ParallelJobCount)
	fmt.Println("  executable:", plan.IsExecutable)

	fmt.Printf("\nSelected libraries (%d):\n", len(plan.SelectedLibraries))
	for _, library := range plan.SelectedLibraries {
		fmt.Printf("  %-20s %s\n", library.LibraryId, library.DisplayName)
	}

	if len(plan.ConfigureFlags) > 0 {
		fmt.Printf("\nConfigure command:\n  ./configure \\\n")
		for index, flag := range plan.ConfigureFlags {
			suffix := " \\"
			if index == len(plan.ConfigureFlags)-1 {
				suffix = ""
			}
			fmt.Printf("    %s%s\n", flag, suffix)
		}
	}

	if len(plan.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(plan.Warnings))
		for _, warning := range plan.Warnings {
			fmt.Printf("  [%s] %s\n", warning.LRiskLevel, warning.Message)
		}
	}
}

// cmdList prints embedded catalog data: versions, presets, or libraries.
func cmdList(args []string) int {
	if len(args) == 0 {
		return exitFor(badArgs("list needs a target: versions | presets | libraries"))
	}
	target := args[0]
	rest := args[1:]
	switch target {
	case "versions":
		return listVersions()
	case "presets":
		return listPresets()
	case "libraries":
		return listLibraries(rest)
	default:
		return exitFor(badArgs("unknown list target %q (want versions | presets | libraries)", target))
	}
}

func listVersions() int {
	for _, release := range planning.LReleaseSupportedListGet() {
		fmt.Printf("%-8s %s\n", release.Version, release.Codename)
	}
	return 0
}

func listPresets() int {
	// Preset IDs are stable across releases; resolve against the latest one.
	release, ok := latestRelease()
	if !ok {
		return exitFor(fmt.Errorf("no supported FFmpeg releases are embedded"))
	}
	for _, preset := range planning.LCatalogPresetSourceBuildResolved(release.ArchiveUrl, "") {
		if preset.Hidden {
			continue
		}
		extendedNote := ""
		if len(preset.ExtendedLibraryIds) > 0 {
			extendedNote = "  (--extended available)"
		}
		fmt.Printf("%-14s %d libraries%s\n", preset.PresetId, len(preset.LibraryIds), extendedNote)
	}
	return 0
}

func listLibraries(args []string) int {
	parsed, err := argsParse(args)
	if err != nil {
		return exitFor(err)
	}
	if strings.TrimSpace(parsed.version) == "" {
		return exitFor(badArgs("list libraries requires --ffmpeg-version (availability varies per release)"))
	}
	release, ok := planning.LReleaseVersionResolve(parsed.version)
	if !ok {
		return exitFor(unsupported("unsupported FFmpeg version: %s", parsed.version))
	}
	for _, library := range planning.LCatalogSourceBuildResolved(release.ArchiveUrl, "") {
		flag := ""
		if len(library.ConfigureFlags) > 0 {
			flag = library.ConfigureFlags[0]
		}
		fmt.Printf("%-20s %-26s %s\n", library.LibraryId, flag, library.DisplayName)
	}
	return 0
}

func latestRelease() (planning.LReleaseChoice, bool) {
	// LReleaseSupportedListGet returns releases newest-first, so index 0 is latest.
	releases := planning.LReleaseSupportedListGet()
	if len(releases) == 0 {
		return planning.LReleaseChoice{}, false
	}
	return releases[0], true
}
