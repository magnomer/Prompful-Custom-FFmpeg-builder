package program

import (
	"bufio"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/hostexec"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

// LLibraryProbeComponents maps a flagless library (no --enable-* flag to find in
// the configuration line) to the ffmpeg component names that prove it is present.
// Verified by running the built ffmpeg and listing its decoders/encoders.
var LLibraryProbeComponents = map[string][]string{
	"png": {"png", "apng"},
}

// LVerificationLibrary reports whether one planned library is present in the built
// ffmpeg, verified either by its configure flag or by probing for a component.
type LVerificationLibrary struct {
	LibraryId     string   `json:"libraryId"`
	DisplayName   string   `json:"displayName"`
	Method        string   `json:"method"` // "flag" | "component" | "builtin"
	ExpectedFlags []string `json:"expectedFlags"`
	MissingFlags  []string `json:"missingFlags"`
	Components    []string `json:"components"`
	Status        string   `json:"status"` // "ok" | "missing" | "builtin"
}

// LVerificationState is the Method A result: compare the build plan's libraries
// against the configuration string the built ffmpeg.exe reports for itself.
type LVerificationState struct {
	FfmpegPath            string                 `json:"ffmpegPath"`
	FfmpegVersion         string                 `json:"ffmpegVersion"`
	Libraries             []LVerificationLibrary `json:"libraries"`
	UnexpectedEnableFlags []string               `json:"unexpectedEnableFlags"`
	OkCount               int                    `json:"okCount"`
	TotalCount            int                    `json:"totalCount"`
	Overall               string                 `json:"overall"` // "ok" | "issues" | "unverifiable"
	Message               string                 `json:"message"`
	VerifiedAt            string                 `json:"verifiedAt"`
}

// LVerificationBuildRun runs the built ffmpeg.exe, reads the configure flags it was
// compiled with, and checks each planned library's enable flag is present.
func (program *LProgram) LVerificationBuildRun(workspaceDirectory string) (LVerificationState, error) {
	if workspaceDirectory == "" {
		return LVerificationState{}, errors.New("workspace directory is required")
	}
	layout := LArtifactLayoutFind(workspaceDirectory)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.ArtifactsDirectory); err != nil {
		return LVerificationState{}, err
	}
	ffmpegPath := filepath.Join(layout.ArtifactsDirectory, "ffmpeg.exe")
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, ffmpegPath); err != nil {
		return LVerificationState{}, err
	}

	verification := LVerificationState{
		FfmpegPath:            ffmpegPath,
		Libraries:             []LVerificationLibrary{},
		UnexpectedEnableFlags: []string{},
		VerifiedAt:            time.Now().UTC().Format(time.RFC3339),
	}

	if info, err := os.Stat(ffmpegPath); err != nil || info.IsDir() {
		verification.Overall = "unverifiable"
		verification.Message = "ffmpeg.exe was not found in the build artifacts."
		return verification, nil
	}

	versionOutput, err := LFFmpegVersionRun(ffmpegPath)
	if err != nil {
		verification.Overall = "unverifiable"
		verification.Message = "Could not run ffmpeg.exe to read its build configuration: " + err.Error()
		return verification, nil
	}
	verification.FfmpegVersion = LVersionFFmpegParse(versionOutput)
	configuredFlags := LFlagConfiguredParse(versionOutput)
	if len(configuredFlags) == 0 {
		verification.Overall = "unverifiable"
		verification.Message = "ffmpeg.exe did not report a configuration line to verify against."
		return verification, nil
	}

	_, report, reportErr := LReportLatestRead(layout)
	if reportErr != nil {
		verification.Overall = "unverifiable"
		verification.Message = "No build report was found to compare against."
		return verification, nil
	}

	// Component names are only probed (an extra ffmpeg run) if a flagless library
	// actually needs them.
	var componentNames map[string]bool
	componentNamesLoaded := false
	ensureComponentNames := func() map[string]bool {
		if !componentNamesLoaded {
			componentNames = LFFmpegComponentGet(ffmpegPath)
			componentNamesLoaded = true
		}
		return componentNames
	}

	plannedFlags := map[string]bool{}
	for _, library := range report.SelectedLibraries {
		if library.DisplayName == "" {
			continue
		}
		expectedFlags := []string{}
		for _, flag := range library.ConfigureFlags {
			flag = strings.TrimSpace(flag)
			if !strings.HasPrefix(flag, "--enable-") {
				continue
			}
			expectedFlags = append(expectedFlags, flag)
			plannedFlags[flag] = true
		}
		libraryVerification := LVerificationLibrary{
			LibraryId:     library.LibraryId,
			DisplayName:   library.DisplayName,
			ExpectedFlags: expectedFlags,
			MissingFlags:  []string{},
			Components:    []string{},
		}
		probeComponents := LLibraryProbeComponents[library.LibraryId]

		switch {
		case len(expectedFlags) > 0:
			libraryVerification.Method = "flag"
			for _, flag := range expectedFlags {
				if !configuredFlags[flag] {
					libraryVerification.MissingFlags = append(libraryVerification.MissingFlags, flag)
				}
			}
			if len(libraryVerification.MissingFlags) == 0 {
				libraryVerification.Status = "ok"
				verification.OkCount++
			} else {
				libraryVerification.Status = "missing"
			}
			verification.TotalCount++
		case len(probeComponents) > 0:
			libraryVerification.Method = "component"
			libraryVerification.Components = probeComponents
			availableComponents := ensureComponentNames()
			found := false
			for _, component := range probeComponents {
				if availableComponents[strings.ToLower(component)] {
					found = true
					break
				}
			}
			if found {
				libraryVerification.Status = "ok"
				verification.OkCount++
			} else {
				libraryVerification.Status = "missing"
			}
			verification.TotalCount++
		default:
			// No enable flag and no probe mapping: these are FFmpeg's built-in
			// components (ffmpeg.exe, libavcodec, native codecs, ...) that ship in
			// every build, so they are reported as present rather than unknown.
			libraryVerification.Method = "builtin"
			libraryVerification.Status = "builtin"
		}
		verification.Libraries = append(verification.Libraries, libraryVerification)
	}

	// "Extra" = a known library catalog library's enable flag present in the binary but
	// not planned. Restricting to library catalog flags avoids flagging the many core
	// --enable-* options (gpl, vulkan, sdl2, ...) that are not libraries.
	knownLibraryFlags := LCatalogLibraryRead(verification.FfmpegVersion)
	for flag := range configuredFlags {
		if knownLibraryFlags[flag] && !plannedFlags[flag] {
			verification.UnexpectedEnableFlags = append(verification.UnexpectedEnableFlags, flag)
		}
	}
	sort.Strings(verification.UnexpectedEnableFlags)

	verification.Overall = "ok"
	for _, library := range verification.Libraries {
		if library.Status == "missing" {
			verification.Overall = "issues"
			break
		}
	}
	return verification, nil
}

// LCatalogLibraryRead returns every --enable-* flag known to belong to a
// library catalog library for the verified FFmpeg version. It intentionally has
// no library/version fallback: an empty or unsupported verified FFmpeg version
// returns no known library flags instead of substituting another release.
func LCatalogLibraryRead(ffmpegVersion string) map[string]bool {
	flags := map[string]bool{}
	versionName := strings.TrimSpace(ffmpegVersion)
	if versionName == "" {
		return flags
	}
	resolvedPlan, err := planning.LCatalogEmbeddedResolve(planning.LCatalogResolutionSettings{
		FfmpegVersion:           versionName,
		WindowsShellProfileName: planning.LSettingsFFmpegCreate().WindowsShellProfileName,
	})
	if err != nil {
		return flags
	}
	for _, library := range resolvedPlan.VisibleLibraries {
		LLibraryFlagsAdd(flags, library.ConfigureFlags)
	}
	for _, library := range resolvedPlan.UnsupportedLibraries {
		LLibraryFlagsAdd(flags, library.ConfigureFlags)
	}
	return flags
}

func LLibraryFlagsAdd(flags map[string]bool, configureFlags []string) {
	for _, flag := range configureFlags {
		flag = strings.TrimSpace(flag)
		if strings.HasPrefix(flag, "--enable-") {
			flags[flag] = true
		}
	}
}

// LFFmpegComponentGet runs the built ffmpeg and collects the names of every
// decoder/encoder/demuxer/muxer it reports, so flagless libraries can be checked
// by the component they provide (e.g. the "png" decoder).
func LFFmpegComponentGet(ffmpegPath string) map[string]bool {
	names := map[string]bool{}
	for _, listArg := range []string{"-decoders", "-encoders", "-demuxers", "-muxers"} {
		output, err := LFFmpegListRun(ffmpegPath, listArg)
		if err != nil {
			continue
		}
		LNameComponentAdd(names, output)
	}
	return names
}

func LFFmpegListRun(ffmpegPath string, listArg string) (string, error) {
	LContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(LContext, ffmpegPath, "-hide_banner", listArg)
	hostexec.LCommandWindowHide(command)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func LNameComponentAdd(set map[string]bool, listOutput string) {
	scanner := bufio.NewScanner(strings.NewReader(listOutput))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	pastHeader := false
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !pastHeader {
			// The capability columns are followed by a "------" separator line;
			// real entries begin after it.
			if strings.HasPrefix(line, "---") {
				pastHeader = true
			}
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 {
			set[strings.ToLower(fields[1])] = true
		}
	}
}

func LFFmpegVersionRun(ffmpegPath string) (string, error) {
	LContext, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(LContext, ffmpegPath, "-hide_banner", "-version")
	hostexec.LCommandWindowHide(command)
	output, err := command.CombinedOutput()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func LFlagConfiguredParse(versionOutput string) map[string]bool {
	flags := map[string]bool{}
	scanner := bufio.NewScanner(strings.NewReader(versionOutput))
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "configuration:") {
			continue
		}
		for _, token := range strings.Fields(strings.TrimPrefix(line, "configuration:")) {
			if strings.HasPrefix(token, "--") {
				flags[token] = true
			}
		}
	}
	return flags
}

func LVersionFFmpegParse(versionOutput string) string {
	scanner := bufio.NewScanner(strings.NewReader(versionOutput))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "ffmpeg version ") {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2]
			}
		}
	}
	return ""
}
