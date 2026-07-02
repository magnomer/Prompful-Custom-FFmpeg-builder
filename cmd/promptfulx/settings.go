package main

import (
	"fmt"
	"strconv"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
)

// usageError carries a process exit code so the caller can distinguish a bad
// invocation (exit 2) from an unsupported version/library/preset value (exit 4),
// per docs/internal/PlanCLI.md §38.
type usageError struct {
	message string
	code    int
}

func (e usageError) Error() string { return e.message }

func badArgs(format string, args ...any) usageError {
	return usageError{message: fmt.Sprintf(format, args...), code: 2}
}

func unsupported(format string, args ...any) usageError {
	return usageError{message: fmt.Sprintf(format, args...), code: 4}
}

// cliBuildArgs holds the raw, pre-resolution CLI selections.
type cliBuildArgs struct {
	version   string
	preset    string
	noPreset  bool
	extended  bool
	enable    []string // FFmpeg-style flags, e.g. "--enable-libx264"
	disable   []string // FFmpeg-style flags, e.g. "--disable-libx264"
	workspace string
	jobs      int
	yes       bool
	noInput   bool
}

// argsParse scans build-shaped CLI arguments. Dynamic --enable-lib*/--disable-lib*
// flags rule out the standard flag package, so this hand-rolls the scan.
func argsParse(args []string) (cliBuildArgs, error) {
	parsed := cliBuildArgs{}
	index := 0
	takeValue := func(name, inline string, hasInline bool) (string, error) {
		if hasInline {
			return inline, nil
		}
		if index+1 >= len(args) {
			return "", badArgs("flag %s needs a value", name)
		}
		index++
		return args[index], nil
	}
	for ; index < len(args); index++ {
		token := args[index]
		name := token
		inline := ""
		hasInline := false
		if eq := strings.Index(token, "="); strings.HasPrefix(token, "--") && eq >= 0 {
			name = token[:eq]
			inline = token[eq+1:]
			hasInline = true
		}
		switch {
		case name == "--ffmpeg-version":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.version = value
		case name == "--preset":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.preset = value
		case name == "--workspace":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.workspace = value
		case name == "--jobs":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			jobs, convErr := strconv.Atoi(strings.TrimSpace(value))
			if convErr != nil || jobs < 0 {
				return parsed, badArgs("--jobs needs a non-negative integer, got %q", value)
			}
			parsed.jobs = jobs
		case name == "--extended":
			parsed.extended = true
		case name == "--no-preset":
			parsed.noPreset = true
		case name == "--yes":
			parsed.yes = true
		case name == "--no-input":
			parsed.noInput = true
		case strings.HasPrefix(name, "--enable-lib"):
			parsed.enable = append(parsed.enable, name)
		case strings.HasPrefix(name, "--disable-lib"):
			parsed.disable = append(parsed.disable, name)
		default:
			return parsed, badArgs("unknown flag: %s", name)
		}
	}
	return parsed, nil
}

// settingsResolve turns parsed CLI arguments into the planner's LSettingsFFmpeg,
// using the shared Step 2 resolvers. It leaves WindowsShellProfileName empty so
// the planner defaults it to ucrt64.
func settingsResolve(parsed cliBuildArgs) (planning.LSettingsFFmpeg, error) {
	if strings.TrimSpace(parsed.version) == "" {
		return planning.LSettingsFFmpeg{}, badArgs("--ffmpeg-version is required")
	}
	release, ok := planning.LReleaseVersionResolve(parsed.version)
	if !ok {
		return planning.LSettingsFFmpeg{}, unsupported("unsupported FFmpeg version: %s", parsed.version)
	}
	url := release.ArchiveUrl

	libraryIds := []string{}
	if parsed.preset != "" && !parsed.noPreset {
		ids, ok := planning.LPresetLibraryIdsResolve(url, "", parsed.preset, parsed.extended)
		if !ok {
			return planning.LSettingsFFmpeg{}, unsupported("unknown preset %q for FFmpeg %s", parsed.preset, release.Version)
		}
		libraryIds = ids
	}

	for _, flag := range parsed.enable {
		id, ok := planning.LLibraryFlagResolve(url, "", flag)
		if !ok {
			return planning.LSettingsFFmpeg{}, unsupported("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = idAppendUnique(libraryIds, id)
	}
	for _, flag := range parsed.disable {
		enableForm := strings.Replace(flag, "--disable-", "--enable-", 1)
		id, ok := planning.LLibraryFlagResolve(url, "", enableForm)
		if !ok {
			return planning.LSettingsFFmpeg{}, unsupported("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = idRemove(libraryIds, id)
	}

	return planning.LSettingsFFmpeg{
		WorkspaceDirectory:       strings.TrimSpace(parsed.workspace),
		FfmpegSourceArchiveUrl:   release.ArchiveUrl,
		FfmpegSourceSignatureUrl: release.SignatureUrl,
		SelectedLibraryIds:       libraryIds,
		ParallelJobCount:         parsed.jobs,
	}, nil
}

func idAppendUnique(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func idRemove(ids []string, id string) []string {
	filtered := ids[:0]
	for _, existing := range ids {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	return filtered
}
