package planning

import "strings"

// This file holds the argument-to-settings resolvers shared by the GUI and the
// PromptfulX CLI. They translate human-facing selections (a version string, a
// preset name, an FFmpeg-style configure flag) into the raw inputs the planner
// consumes (source archive URL, library IDs). The planner remains the single
// authority on compatibility; these resolvers only look up embedded catalog
// data, they do not decide what is buildable.

// LReleaseVersionResolve finds the supported FFmpeg release for an exact version
// string (for example "8.1.2"). The returned choice carries the archive and
// signature URLs the planner needs. The boolean is false when the version is not
// a supported release.
func LReleaseVersionResolve(version string) (LReleaseChoice, bool) {
	version = strings.TrimSpace(version)
	if version == "" {
		return LReleaseChoice{}, false
	}
	for _, release := range LReleaseSupportedListGet() {
		if release.Version == version {
			return release, true
		}
	}
	return LReleaseChoice{}, false
}

// LPresetLibraryIdsResolve returns the library IDs a preset selects for the
// FFmpeg release identified by ffmpegSourceArchiveUrl. When extended is true it
// returns the preset's extended set (a superset), falling back to the normal set
// when the preset defines no extended libraries. The boolean is false when the
// preset ID is unknown for that release. The result is a fresh copy the caller
// may mutate (for example to apply --enable/--disable overrides).
func LPresetLibraryIdsResolve(ffmpegSourceArchiveUrl string, windowsShellProfileName string, presetId string, extended bool) ([]string, bool) {
	presetId = strings.TrimSpace(presetId)
	if presetId == "" {
		return nil, false
	}
	for _, preset := range LCatalogPresetSourceBuildResolved(ffmpegSourceArchiveUrl, windowsShellProfileName) {
		if preset.PresetId != presetId {
			continue
		}
		if extended && len(preset.ExtendedLibraryIds) > 0 {
			return append([]string{}, preset.ExtendedLibraryIds...), true
		}
		return append([]string{}, preset.LibraryIds...), true
	}
	return nil, false
}

// LLibraryFlagResolve maps an FFmpeg-style configure enable flag (for example
// "--enable-libx264") to Promptful's internal library ID for the FFmpeg release
// identified by ffmpegSourceArchiveUrl. Callers normalize a --disable-lib* flag
// to its --enable-lib* form before lookup, since both refer to the same library.
// The boolean is false when no catalog library owns the flag for that release.
func LLibraryFlagResolve(ffmpegSourceArchiveUrl string, windowsShellProfileName string, configureEnableFlag string) (string, bool) {
	configureEnableFlag = strings.TrimSpace(configureEnableFlag)
	if configureEnableFlag == "" {
		return "", false
	}
	for _, library := range LCatalogSourceBuildResolved(ffmpegSourceArchiveUrl, windowsShellProfileName) {
		for _, flag := range library.ConfigureFlags {
			if flag == configureEnableFlag {
				return library.LibraryId, true
			}
		}
	}
	return "", false
}
