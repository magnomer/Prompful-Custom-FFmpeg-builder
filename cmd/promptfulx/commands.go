package main

import (
	"fmt"
	"os"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
)

// LCommandExitGet maps an error to a process exit code: a LErrorUsage carries its own
// code, anything else is a general failure (1).
func LCommandExitGet(err error) int {
	var usage LErrorUsage
	if e, ok := err.(LErrorUsage); ok {
		usage = e
		fmt.Fprintln(os.Stderr, "promptfulx:", usage.message)
		return usage.code
	}
	fmt.Fprintln(os.Stderr, "promptfulx:", err.Error())
	return 1
}

// LCommandPlanRun resolves CLI arguments into a build plan and prints it. It never
// modifies the build environment.
func LCommandPlanRun(args []string) int {
	parsed, err := LArgumentParse(args)
	if err != nil {
		return LCommandExitGet(err)
	}
	if err := LArgumentScopeCheck(parsed, lArgumentFfmpegFlags, lArgumentWorkspaceFlags); err != nil {
		return LCommandExitGet(err)
	}
	settings, err := LSettingsFfmpegResolve(parsed)
	if err != nil {
		return LCommandExitGet(err)
	}
	plan, err := planning.LPlanFfmpegCreate(settings)
	if err != nil {
		return LCommandExitGet(LErrorSupportCreate("could not resolve plan: %v", err))
	}
	LCommandPlanPrint(plan)
	if !plan.IsExecutable {
		return 4 // resolved, but blocked by an unsupported combination
	}
	return 0
}

func LCommandPlanPrint(plan planning.LPlanFfmpeg) {
	version := plan.CompatibilityFfmpegVersion
	if version == "" {
		version = plan.RequestedFfmpegVersion
	}
	fmt.Println("FFmpeg build plan")
	fmt.Println("  version:   ", version)
	fmt.Println("  source:    ", plan.FfmpegSourceArchiveUrl)
	if plan.FfmpegSourceSignatureUrl != "" {
		fmt.Println("  signature: ", plan.FfmpegSourceSignatureUrl)
	}
	if plan.FfmpegSourceSha256Hash != "" {
		fmt.Println("  sha256:    ", plan.FfmpegSourceSha256Hash)
	}
	if plan.WorkspaceDirectory != "" {
		fmt.Println("  workspace: ", plan.WorkspaceDirectory)
	}
	if plan.WindowsShellProfileName != "" {
		fmt.Println("  profile:   ", plan.WindowsShellProfileName)
	}
	if plan.LicenseProfileName != "" {
		fmt.Println("  license:   ", plan.LicenseProfileName)
	}
	fmt.Println("  jobs:      ", plan.ParallelJobCount)
	fmt.Println("  executable:", plan.IsExecutable)
	// The final confirmation prompt approves this digest; show it here so the
	// reviewed plan and the confirmed hash can be compared.
	fmt.Println("  plan hash: ", plan.PlanHash)

	fmt.Printf("\nSelected libraries (%d):\n", len(plan.SelectedLibraries))
	for _, library := range plan.SelectedLibraries {
		fmt.Printf("  %-20s %s\n", library.LibraryId, library.DisplayName)
	}

	if len(plan.SelectedConfigureOptions) > 0 {
		fmt.Printf("\nSelected options (%d):\n", len(plan.SelectedConfigureOptions))
		for _, option := range plan.SelectedConfigureOptions {
			fmt.Printf("  %-20s %s\n", option.OptionId, option.DisplayName)
		}
	}

	// Non-native libraries build from their own source archive; each URL/version is
	// part of what Build fetches and what the plan hash covers.
	if len(plan.LPreparationCatalog) > 0 {
		fmt.Printf("\nLibrary preparations (%d):\n", len(plan.LPreparationCatalog))
		for _, preparation := range plan.LPreparationCatalog {
			fmt.Printf("  %-20s %-12s %s\n", preparation.LibraryId, preparation.Version, preparation.ArchiveUrl)
		}
	}

	if len(plan.RequiredMsys2PackageNames) > 0 {
		fmt.Printf("\nRequired MSYS2 packages (%d):\n", len(plan.RequiredMsys2PackageNames))
		for _, packageName := range plan.RequiredMsys2PackageNames {
			fmt.Println("  -", packageName)
		}
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

	if len(plan.Operations) > 0 {
		fmt.Printf("\nOperations (%d):\n", len(plan.Operations))
		for _, operation := range plan.Operations {
			fmt.Printf("  - %s: %s\n", operation.OperationName, operation.Summary)
		}
	}

	lCommandFidelityPrint(plan)

	if len(plan.Warnings) > 0 {
		fmt.Printf("\nWarnings (%d):\n", len(plan.Warnings))
		for _, warning := range plan.Warnings {
			fmt.Printf("  [%s] %s\n", warning.LRiskLevel, warning.Message)
		}
	}
}

// lCommandFidelityPrint discloses configure flags whose future-compatibility check
// can drop them at build time, so the printed/hashed command is not always what runs.
func lCommandFidelityPrint(plan planning.LPlanFfmpeg) {
	mutableFlags := map[string]string{
		"--enable-libsvtjpegxs": "may fetch and build upstream https://github.com/OpenVisualCloud/SVT-JPEG-XS.git (not in this plan) if the package is missing/incompatible, else removed",
		"--enable-liblensfun":   "removed if lensfun fails its compatibility probe",
		"--enable-vapoursynth":  "removed if VapourSynth fails its compatibility probe",
	}
	notes := make([][2]string, 0, len(mutableFlags))
	for _, flag := range plan.ConfigureFlags {
		if note, ok := mutableFlags[flag]; ok {
			notes = append(notes, [2]string{flag, note})
		}
	}
	if len(notes) == 0 {
		return
	}
	fmt.Printf("\nRuntime fidelity notes (%d):\n", len(notes))
	for _, note := range notes {
		fmt.Printf("  %s: %s\n", note[0], note[1])
	}
}

// LCommandListRun prints embedded catalog data: versions, presets, or libraries.
func LCommandListRun(args []string) int {
	if len(args) == 0 {
		return LCommandExitGet(LErrorArgumentCreate("list needs a target: versions | presets | libraries"))
	}
	target := args[0]
	rest := args[1:]
	switch target {
	case "versions":
		if len(rest) > 0 {
			return LCommandExitGet(LErrorArgumentCreate("list versions takes no arguments, got %q", rest[0]))
		}
		return LCommandVersionList()
	case "presets":
		if len(rest) > 0 {
			return LCommandExitGet(LErrorArgumentCreate("list presets takes no arguments, got %q", rest[0]))
		}
		return LCommandPresetList()
	case "libraries":
		return LCommandLibraryList(rest)
	default:
		return LCommandExitGet(LErrorArgumentCreate("unknown list target %q (want versions | presets | libraries)", target))
	}
}

func LCommandVersionList() int {
	for _, release := range planning.LReleaseSupportedGet() {
		fmt.Printf("%-8s %s\n", release.Version, release.Codename)
	}
	return 0
}

func LCommandPresetList() int {
	// Preset IDs are stable across releases; resolve against the latest one.
	release, ok := LReleaseLatestGet()
	if !ok {
		return LCommandExitGet(fmt.Errorf("no supported FFmpeg releases are embedded"))
	}
	for _, preset := range planning.LCatalogPresetGet(release.ArchiveUrl, "") {
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

func LCommandLibraryList(args []string) int {
	parsed, err := LArgumentParse(args)
	if err != nil {
		return LCommandExitGet(err)
	}
	if err := LArgumentScopeCheck(parsed, []string{"--ffmpeg-version"}); err != nil {
		return LCommandExitGet(err)
	}
	if strings.TrimSpace(parsed.version) == "" {
		return LCommandExitGet(LErrorArgumentCreate("list libraries requires --ffmpeg-version (availability varies per release)"))
	}
	release, ok := planning.LReleaseVersionResolve(parsed.version)
	if !ok {
		return LCommandExitGet(LErrorSupportCreate("unsupported FFmpeg version: %s", parsed.version))
	}
	for _, library := range planning.LCatalogLibraryGet(release.ArchiveUrl, "") {
		flag := ""
		if len(library.ConfigureFlags) > 0 {
			flag = library.ConfigureFlags[0]
		}
		fmt.Printf("%-20s %-26s %s\n", library.LibraryId, flag, library.DisplayName)
	}
	return 0
}

func LReleaseLatestGet() (planning.LReleaseChoice, bool) {
	// LReleaseSupportedGet returns releases newest-first, so index 0 is latest.
	releases := planning.LReleaseSupportedGet()
	if len(releases) == 0 {
		return planning.LReleaseChoice{}, false
	}
	return releases[0], true
}
