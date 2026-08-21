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

	// Canonical flag tokens seen, in order, for per-command scope validation.
	// Dynamic library flags collapse to the "--enable-lib"/"--disable-lib" sentinels.
	seenFlags []string
}

// Flag scope groups. Each command accepts only the flags its resolver reads;
// LArgumentScopeCheck rejects any other recognized flag instead of ignoring it.
var (
	lArgumentFfmpegFlags    = []string{"--ffmpeg-version", "--preset", "--extended", "--no-preset", "--enable-lib", "--disable-lib", "--jobs"}
	lArgumentWorkspaceFlags = []string{"--workspace"}
	lArgumentMsys2Flags     = []string{"--msys2-url", "--msys2-sha256", "--msys2-signature-url"}
	lArgumentConfirmFlags   = []string{"--yes", "--no-input"}
)

// LArgumentScopeCheck rejects any parsed flag not in the command's allowed set,
// so a flag meant for another command fails loudly instead of being discarded.
func LArgumentScopeCheck(parsed LArgumentBuild, allowed ...[]string) error {
	set := map[string]bool{}
	for _, group := range allowed {
		for _, name := range group {
			set[name] = true
		}
	}
	for _, name := range parsed.seenFlags {
		if !set[name] {
			return LErrorArgumentCreate("flag %s is not valid for this command", name)
		}
	}
	return nil
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
		// A following --flag is a missing value, not the value itself; consuming it
		// would hide the mistake and mis-report a later error.
		if next := args[index+1]; strings.HasPrefix(next, "--") {
			return "", LErrorArgumentCreate("flag %s needs a value, but %s looks like another flag", name, next)
		}
		index++
		return args[index], nil
	}
	rejectInline := func(name string, hasInline bool) error {
		if hasInline {
			return LErrorArgumentCreate("flag %s does not take a value", name)
		}
		return nil
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
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--preset":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.preset = value
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--workspace":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.workspace = value
			parsed.seenFlags = append(parsed.seenFlags, name)
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
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--msys2-url":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2URL = value
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--msys2-sha256":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2Sha256 = value
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--msys2-signature-url":
			value, err := takeValue(name, inline, hasInline)
			if err != nil {
				return parsed, err
			}
			parsed.msys2SignatureURL = value
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--extended":
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.extended = true
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--no-preset":
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.noPreset = true
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--yes":
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.yes = true
			parsed.seenFlags = append(parsed.seenFlags, name)
		case name == "--no-input":
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.noInput = true
			parsed.seenFlags = append(parsed.seenFlags, name)
		case strings.HasPrefix(name, "--enable-lib"):
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.enable = append(parsed.enable, name)
			parsed.seenFlags = append(parsed.seenFlags, "--enable-lib")
		case strings.HasPrefix(name, "--disable-lib"):
			if err := rejectInline(name, hasInline); err != nil {
				return parsed, err
			}
			parsed.disable = append(parsed.disable, name)
			parsed.seenFlags = append(parsed.seenFlags, "--disable-lib")
		default:
			return parsed, LErrorArgumentCreate("unknown flag: %s", name)
		}
	}
	return parsed, nil
}

// LSettingsFfmpegResolve turns parsed CLI arguments into the planner's LSettingsFfmpeg,
// using the shared Step 2 resolvers. It leaves WindowsShellProfileName empty so
// the planner defaults it to ucrt64.
func LSettingsFfmpegResolve(parsed LArgumentBuild) (planning.LSettingsFfmpeg, error) {
	if strings.TrimSpace(parsed.version) == "" {
		return planning.LSettingsFfmpeg{}, LErrorArgumentCreate("--ffmpeg-version is required")
	}
	release, ok := planning.LReleaseVersionResolve(parsed.version)
	if !ok {
		return planning.LSettingsFfmpeg{}, LErrorSupportCreate("unsupported FFmpeg version: %s", parsed.version)
	}
	url := release.ArchiveUrl

	libraryIds := []string{}
	if parsed.preset != "" && !parsed.noPreset {
		ids, ok := planning.LPresetIdentifiersResolve(url, "", parsed.preset, parsed.extended)
		if !ok {
			return planning.LSettingsFfmpeg{}, LErrorSupportCreate("unknown preset %q for FFmpeg %s", parsed.preset, release.Version)
		}
		libraryIds = ids
	}

	for _, flag := range parsed.enable {
		id, ok := planning.LLibraryFlagResolve(url, "", flag)
		if !ok {
			return planning.LSettingsFfmpeg{}, LErrorSupportCreate("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = LIdentifierAppend(libraryIds, id)
	}
	for _, flag := range parsed.disable {
		enableForm := strings.Replace(flag, "--disable-", "--enable-", 1)
		id, ok := planning.LLibraryFlagResolve(url, "", enableForm)
		if !ok {
			return planning.LSettingsFfmpeg{}, LErrorSupportCreate("unknown library flag %q for FFmpeg %s", flag, release.Version)
		}
		libraryIds = LIdentifierRemove(libraryIds, id)
	}

	return planning.LSettingsFfmpeg{
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
		// A custom archive invalidates the default repo.msys2.org .sig; drop it so
		// LSettingsBuildClean re-derives archive+".sig" unless overridden below.
		settings.Msys2ArchiveSignatureUrl = ""
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
