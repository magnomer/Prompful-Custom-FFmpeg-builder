package main

import (
	"fmt"
	"strconv"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
)

// LErrorUsage carries a process exit code so the caller can distinguish a bad
// invocation (exit 2) from an unsupported version/library/preset value (exit 4),
// per docs/internal/PlanCLI.md §38.
type LErrorUsage struct {
	message string
	code    int
}

func (e LErrorUsage) Error() string { return e.message }

func LErrorArgumentCreate(format string, args ...any) LErrorUsage {
	return LErrorUsage{message: fmt.Sprintf(format, args...), code: 2}
}

func LErrorSupportCreate(format string, args ...any) LErrorUsage {
	return LErrorUsage{message: fmt.Sprintf(format, args...), code: 4}
}

// LArgumentBuild holds the raw, pre-resolution CLI selections.
type LArgumentBuild struct {
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

	// setup-only MSYS2 source overrides (empty = use embedded defaults)
	msys2URL          string
	msys2Sha256       string
	msys2SignatureURL string
}

// LArgumentParse scans build-shaped CLI arguments. Dynamic --enable-lib*/--disable-lib*
// flags rule out the standard flag package, so this hand-rolls the scan.
func LArgumentParse(args []string) (LArgumentBuild, error) {
	parsed := LArgumentBuild{}
	index := 0
	takeValue := func(name, inline string, hasInline bool) (string, error) {
		if hasInline {
			return inline, nil
		}
		if index+1 >= len(args) {
			return "", LErrorArgumentCreate("flag %s needs a value", name)
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
				return parsed, LErrorArgumentCreate("--jobs needs a non-negative integer, got %q", value)
			}
			parsed.jobs = jobs
		case name == "--msys2-url":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2URL = value
		case name == "--msys2-sha256":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2Sha256 = value
		case name == "--msys2-signature-url":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2SignatureURL = value
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
			return parsed, LErrorArgumentCreate("unknown flag: %s", name)
		}
	}
	return parsed, nil
}

// LSettingsFFmpegResolve turns parsed CLI arguments into the planner's LSettingsFFmpeg,
// using the shared Step 2 resolvers. It leaves WindowsShellProfileName empty so
// the planner defaults it to ucrt64.
func LSettingsFFmpegResolve(parsed LArgumentBuild) (planning.LSettingsFFmpeg, error) {
	if strings.TrimSpace(parsed.version) == "" {
		return planning.LSettingsFFmpeg{}, LErrorArgumentCreate("--ffmpeg-version is required")
	}
	release, ok := planning.LReleaseVersionResolve(parsed.version)
	if !ok {
		return planning.LSettingsFFmpeg{}, LErrorSupportCreate("unsupported FFmpeg version: %s", parsed.version)
	}
	url := release.ArchiveUrl

	libraryIds := []string{}
	if parsed.preset != "" && !parsed.noPreset {
		ids, ok := planning.LPresetIdentifiersResolve(url, "", parsed.preset, parsed.extended)
		if !ok {
			return planning.LSettingsFFmpeg{}, LErrorSupportCreate("unknown preset %q for FFmpeg %s", parsed.preset, release.Version)
		}
		libraryIds = ids
	}

	for _, flag := range parsed.enable {
		id, ok := planning.LLibraryFlagResolve(url, "", flag)
		if !ok {
			return planning.LSettingsFFmpeg{}, LErrorSupportCreate("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = LIdentifierAppend(libraryIds, id)
	}
	for _, flag := range parsed.disable {
		enableForm := strings.Replace(flag, "--disable-", "--enable-", 1)
		id, ok := planning.LLibraryFlagResolve(url, "", enableForm)
		if !ok {
			return planning.LSettingsFFmpeg{}, LErrorSupportCreate("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = LIdentifierRemove(libraryIds, id)
	}

	return planning.LSettingsFFmpeg{
		WorkspaceDirectory:       strings.TrimSpace(parsed.workspace),
		FfmpegSourceArchiveUrl:   release.ArchiveUrl,
		FfmpegSourceSignatureUrl: release.SignatureUrl,
		SelectedLibraryIds:       libraryIds,
		ParallelJobCount:         parsed.jobs,
	}, nil
}

// LSettingsToolchainResolve builds the MSYS2 toolchain settings for `setup`,
// starting from the embedded defaults (archive URL, signature URL, package set,
// ucrt64 profile) and applying any --msys2-* overrides.
func LSettingsToolchainResolve(parsed LArgumentBuild) (planning.LSettingsToolchain, error) {
	if strings.TrimSpace(parsed.workspace) == "" {
		return planning.LSettingsToolchain{}, LErrorArgumentCreate("--workspace is required")
	}
	settings := planning.LSettingsBuildCreate()
	settings.WorkspaceDirectory = strings.TrimSpace(parsed.workspace)
	if parsed.msys2URL != "" {
		settings.Msys2ArchiveUrl = parsed.msys2URL
	}
	if parsed.msys2Sha256 != "" {
		settings.Msys2ArchiveSha256Hash = parsed.msys2Sha256
	}
	if parsed.msys2SignatureURL != "" {
		settings.Msys2ArchiveSignatureUrl = parsed.msys2SignatureURL
	}
	return settings, nil
}

func LIdentifierAppend(ids []string, id string) []string {
	for _, existing := range ids {
		if existing == id {
			return ids
		}
	}
	return append(ids, id)
}

func LIdentifierRemove(ids []string, id string) []string {
	filtered := ids[:0]
	for _, existing := range ids {
		if existing != id {
			filtered = append(filtered, existing)
		}
	}
	return filtered
}
