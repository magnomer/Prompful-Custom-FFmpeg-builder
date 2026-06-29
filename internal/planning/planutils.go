package planning

import (
	"regexp"
	"sort"
	"strings"

	"promptfulcustomffmpegbuilder/shared/librarysources"
	"promptfulcustomffmpegbuilder/shared/releasesupport"
)

// annotateLibraryVersionCompatibility sets VersionCompatibility on each library in place,
// per the chosen FFmpeg release's support manifest: whether the release supports the library
// and its pkg-config minimum. Covers every track (a non-native library such as libmpeghdec is
// equally version-gated). When the release line is not manifested, or the version is a
// snapshot, libraries are left unannotated. The annotation is informational for the UI; the
// build-blocking decision is made in appendFfmpegVersionWarnings.
func annotateLibraryVersionCompatibility(libraries []LibraryChoice, ffmpegVersion string) {
	release, manifested := releasesupport.ResolveReleaseSupport(ffmpegVersion)
	if !manifested {
		return
	}
	for index := range libraries {
		if libraries[index].Locked {
			continue // included FFmpeg components (ffmpeg.exe, libavcodec, ...) have no --enable switch
		}
		support, supported := release.LibrarySupportFor(libraries[index].LibraryId)
		libraries[index].VersionCompatibility = &LibraryVersionCompatibility{
			Supported:  supported,
			Available:  supported && !support.Unavailable,
			MinVersion: support.MinVersion,
		}
	}
}

// applyVersionDependentTrack flips a Native catalog library to the Internal source-build track
// for FFmpeg release lines whose support manifest marks it sourceBuild — releases whose required
// API the MSYS2 native package cannot satisfy, so the builder builds a compatible pinned version
// from source (resolved from library-sources.json) instead. The native package names are cleared
// so the unsatisfiable package is not installed. A release that does not mark the library leaves
// it on its catalog track, where the native package is the right one. Mutates in place; call once
// per library slice before the track slices, package list, and prep partition are derived. The
// library still presents as available (it is buildable, just from source); only Unavailable blocks.
func applyVersionDependentTrack(libraries []LibraryChoice, ffmpegVersion string) {
	release, manifested := releasesupport.ResolveReleaseSupport(ffmpegVersion)
	if !manifested {
		return
	}
	for index := range libraries {
		if libraries[index].TrackName != LibraryTrackNative {
			continue
		}
		if support, supported := release.LibrarySupportFor(libraries[index].LibraryId); supported && support.SourceBuild {
			libraries[index].TrackName = LibraryTrackInternal
			// Empty (not nil): PackageNames has no omitempty json tag, so nil would serialize
			// to JSON null and crash the picker, which calls packageNames.join()/.length on
			// every catalog row. An empty slice serializes to [] like the static-catalog
			// internal rows, and len()==0 still means "no native package to install".
			libraries[index].PackageNames = []string{}
		}
	}
}

// LibraryCatalogForFfmpegSource returns the shell-profile catalog annotated with FFmpeg
// release support facts for every selectable library. It is the lightweight picker-facing
// path: unlike PlanFfmpegBuild, it does not create a review session and it covers the full
// catalog rather than only the selected libraries.
func LibraryCatalogForFfmpegSource(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []LibraryChoice {
	libraries := LibraryCatalogForShellProfile(windowsShellProfileName)
	ffmpegVersion := ffmpegVersionFromArchiveUrl(ffmpegSourceArchiveUrl)
	annotateLibraryVersionCompatibility(libraries, ffmpegVersion)
	// Flip any library the chosen release marks sourceBuild from Native to Internal so the
	// picker shows its "Source build" track badge for exactly the FFmpeg versions where it is
	// source-built (e.g. libplacebo on 5.1, vmaf on 4.4), matching the plan. A release that
	// does not mark it leaves it Native, so the badge is version-specific, not global.
	applyVersionDependentTrack(libraries, ffmpegVersion)
	return libraries
}

// appendFfmpegVersionWarnings adds the FFmpeg-version plan warnings and reports whether any
// of them blocks the build. When the chosen release line is manifested in the per-release
// support data, each selected library or configure option the release does not support is
// blocked early with a clear message (configure would otherwise fail late and cryptically) —
// the same treatment as an unknown library id or an unprepared track. An un-manifested line
// is not gated here (FFmpeg configure stays the backstop). It always appends the non-blocking
// release-line advisory (see appendFfmpegReleaseAdvisory).
func appendFfmpegVersionWarnings(warnings []PlanWarning, ffmpegVersion string, libraries []LibraryChoice, selectedConfigureOptionIds []string) ([]PlanWarning, bool) {
	blocked := false
	if release, manifested := releasesupport.ResolveReleaseSupport(ffmpegVersion); manifested {
		seenLibrary := map[string]bool{}
		for _, library := range libraries {
			if library.Locked || seenLibrary[library.LibraryId] {
				continue // included FFmpeg components have no --enable switch to gate
			}
			seenLibrary[library.LibraryId] = true
			support, supported := release.LibrarySupportFor(library.LibraryId)
			if !supported {
				warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.libraryUnsupportedForVersion",
					library.DisplayName+" is not supported by FFmpeg "+ffmpegVersion+". Remove it, or choose an FFmpeg release that provides it.",
					map[string]string{"library": library.DisplayName, "version": ffmpegVersion}))
				blocked = true
			} else if support.Unavailable {
				warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.libraryPackageUnavailableForVersion",
					library.DisplayName+" cannot be built for FFmpeg "+ffmpegVersion+": the available package is too old for what this FFmpeg release requires. Remove it, or choose an FFmpeg release it can satisfy.",
					map[string]string{"library": library.DisplayName, "version": ffmpegVersion}))
				blocked = true
			}
		}
		seenOption := map[string]bool{}
		for _, optionId := range selectedConfigureOptionIds {
			if optionId == "" || seenOption[optionId] {
				continue
			}
			seenOption[optionId] = true
			if !release.OptionSupported(optionId) {
				warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.optionUnsupportedForVersion",
					"The selected build option "+optionId+" is not supported by FFmpeg "+ffmpegVersion+". Remove it, or choose an FFmpeg release that provides it.",
					map[string]string{"option": optionId, "version": ffmpegVersion}))
				blocked = true
			}
		}
	}
	warnings = appendFfmpegReleaseAdvisory(warnings, ffmpegVersion)
	return warnings, blocked
}

// appendFfmpegReleaseAdvisory adds a non-blocking advisory about the chosen FFmpeg release:
//   - Supported line, older patch than recommended (e.g. 8.1.1 on the 8.1 line): advise the
//     recommended patch (8.1.2), which carries that line's security/critical fixes.
//   - Unmaintained line (e.g. 6.0, 4.2), and not newer than the newest known release: advise
//     the supported set.
//
// A snapshot/unparseable version, an exact recommended release, a newer patch on a supported
// line, or a version newer than the newest known release draws no advisory.
func appendFfmpegReleaseAdvisory(warnings []PlanWarning, ffmpegVersion string) []PlanWarning {
	lineKey := ffmpegReleaseLineKey(ffmpegVersion)
	if lineKey == "" {
		return warnings
	}
	if recommended, supported := supportedFfmpegReleases[lineKey]; supported {
		if comparison, decidable := librarysources.CompareVersions(ffmpegVersion, recommended); decidable && comparison < 0 {
			warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.ffmpegPatchOutdated",
				"FFmpeg "+ffmpegVersion+" is an older patch of the "+lineKey+" line. Use the recommended "+recommended+", which carries that line's security and critical fixes.",
				map[string]string{"version": ffmpegVersion, "line": lineKey, "recommended": recommended}))
		}
		return warnings
	}
	if comparison, decidable := librarysources.CompareVersions(ffmpegVersion, highestRecommendedFfmpegRelease()); decidable && comparison > 0 {
		return warnings
	}
	supportedList := strings.Join(recommendedFfmpegReleasesDescending(), ", ")
	warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.ffmpegReleaseUnsupported",
		"FFmpeg "+ffmpegVersion+" is not a supported release line. The vouched-for versions are: "+supportedList+".",
		map[string]string{"version": ffmpegVersion, "supported": supportedList}))
	return warnings
}

// ffmpegArchiveVersionPattern extracts the dotted-numeric release from an FFmpeg source
// archive filename, e.g. ".../ffmpeg-8.1.2.tar.xz" -> "8.1.2".
var ffmpegArchiveVersionPattern = regexp.MustCompile(`ffmpeg-(\d+(?:\.\d+){0,2})`)

// ffmpegVersionFromArchiveUrl derives the FFmpeg version from its source archive URL so
// the library version layer can be resolved for that release. Returns "" when the URL
// carries no recognizable version (e.g. a git snapshot), which makes the resolver fall
// back to the highest recorded release.
func ffmpegVersionFromArchiveUrl(archiveUrl string) string {
	match := ffmpegArchiveVersionPattern.FindStringSubmatch(archiveUrl)
	if match == nil {
		return ""
	}
	return match[1]
}

// FfmpegVersionFromArchiveUrl is the exported form used by callers outside this package (the
// configure-script builder) to derive the FFmpeg version from a plan's source archive URL so
// the build-time pkg-config floors can be resolved per release. Returns "" for a snapshot URL.
func FfmpegVersionFromArchiveUrl(archiveUrl string) string {
	return ffmpegVersionFromArchiveUrl(archiveUrl)
}

func selectConfigureOptions(selectedOptionIds []string) ([]ConfigureOptionChoice, []string) {
	catalog := ConfigureOptionCatalog()
	catalogById := map[string]ConfigureOptionChoice{}
	for _, option := range catalog {
		catalogById[option.OptionId] = option
	}
	selectedOptions := []ConfigureOptionChoice{}
	unknownOptionIds := []string{}
	seen := map[string]bool{}
	for _, selectedOptionId := range selectedOptionIds {
		if selectedOptionId == "" || seen[selectedOptionId] {
			continue
		}
		seen[selectedOptionId] = true
		option, found := catalogById[selectedOptionId]
		if !found {
			unknownOptionIds = append(unknownOptionIds, selectedOptionId)
			continue
		}
		selectedOptions = append(selectedOptions, option)
	}
	return selectedOptions, unknownOptionIds
}

func uniqueFlagsFromConfigureOptions(options []ConfigureOptionChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, option := range options {
		for _, flag := range option.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
}

func selectLibraries(windowsShellProfileName string, selectedLibraryIds []string) ([]LibraryChoice, []string) {
	catalog := LibraryCatalogForShellProfile(windowsShellProfileName)
	catalogById := map[string]LibraryChoice{}
	for _, library := range catalog {
		catalogById[library.LibraryId] = library
	}
	selectedLibraries := []LibraryChoice{}
	unknownLibraryIds := []string{}
	seen := map[string]bool{}
	for _, selectedLibraryId := range selectedLibraryIds {
		if selectedLibraryId == "" || seen[selectedLibraryId] {
			continue
		}
		seen[selectedLibraryId] = true
		library, found := catalogById[selectedLibraryId]
		if !found {
			unknownLibraryIds = append(unknownLibraryIds, selectedLibraryId)
			continue
		}
		selectedLibraries = append(selectedLibraries, library)
	}
	return selectedLibraries, unknownLibraryIds
}

// librariesForConfigureFlags returns catalog entries whose configure flags overlap
// with the given list, excluding any already in skip. Used to resolve ExtraConfigureFlags
// back to their MSYS2 packages.
func librariesForConfigureFlags(windowsShellProfileName string, flags []string, skip []LibraryChoice) []LibraryChoice {
	flagSet := map[string]bool{}
	for _, f := range flags {
		flagSet[f] = true
	}
	skipIds := map[string]bool{}
	for _, lib := range skip {
		skipIds[lib.LibraryId] = true
	}
	result := []LibraryChoice{}
	seen := map[string]bool{}
	for _, lib := range LibraryCatalogForShellProfile(windowsShellProfileName) {
		if skipIds[lib.LibraryId] || seen[lib.LibraryId] {
			continue
		}
		for _, f := range lib.ConfigureFlags {
			if flagSet[f] {
				seen[lib.LibraryId] = true
				result = append(result, lib)
				break
			}
		}
	}
	return result
}

func uniquePackagesFromLibraries(libraries []LibraryChoice) []string {
	packages := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		if library.TrackName != LibraryTrackNative {
			continue
		}
		for _, packageName := range library.PackageNames {
			if !seen[packageName] {
				packages = append(packages, packageName)
				seen[packageName] = true
			}
		}
	}
	sort.Strings(packages)
	return packages
}

func librariesForTrack(libraries []LibraryChoice, trackName LibraryTrackName) []LibraryChoice {
	trackedLibraries := []LibraryChoice{}
	for _, library := range libraries {
		if library.TrackName == trackName {
			trackedLibraries = append(trackedLibraries, library)
		}
	}
	return trackedLibraries
}

func groupLibrariesByTrack(libraries []LibraryChoice) []TrackedLibrarySelection {
	return []TrackedLibrarySelection{
		{TrackName: LibraryTrackNative, Libraries: librariesForTrack(libraries, LibraryTrackNative)},
		{TrackName: LibraryTrackInternal, Libraries: librariesForTrack(libraries, LibraryTrackInternal)},
		{TrackName: LibraryTrackExternal, Libraries: librariesForTrack(libraries, LibraryTrackExternal)},
	}
}

func uniqueFlagsFromLibraries(libraries []LibraryChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		for _, flag := range library.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
}

func mergeUniqueStrings(first []string, second []string) []string {
	merged := []string{}
	seen := map[string]bool{}
	for _, value := range append(first, second...) {
		if value == "" || seen[value] {
			continue
		}
		merged = append(merged, value)
		seen[value] = true
	}
	return merged
}

// windowsHardwareBaseConfigureFlags are configure flags forced on for every build because
// this builder targets Windows only. VAAPI and VDPAU are Linux/X11 hardware-acceleration
// APIs that do not exist on Windows, but FFmpeg's configure auto-detects VAAPI whenever a
// libva is visible to pkg-config. libva can arrive transitively — e.g. selecting opencv
// pulls the MSYS2 ffmpeg package, which depends on libva — and the auto-enabled VAAPI then
// breaks the Intel QSV build: libavcodec/qsv_internal.h includes <va/va_drm.h> under
// CONFIG_VAAPI, a Linux-only header, so qsv*.o fail to compile. Disabling both unconditionally
// keeps accidental host libraries from changing the build and is harmless on Windows.
func windowsHardwareBaseConfigureFlags() []string {
	return []string{"--disable-vaapi", "--disable-vdpau"}
}

func addLicenseFlags(configureFlags []string, licenseProfileName string, libraries []LibraryChoice) []string {
	needsGpl := false
	needsNonfree := licenseProfileName == "nonfree-local"
	needsVersion3 := false
	for _, library := range libraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
		if libraryRequiresVersion3(library.LibraryId) {
			needsVersion3 = true
		}
	}
	if licenseProfileName == "gpl-local" {
		needsGpl = true
	}
	if needsGpl {
		configureFlags = mergeUniqueStrings([]string{"--enable-gpl"}, configureFlags)
	}
	if needsVersion3 {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-version3"})
	}
	if needsNonfree {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-nonfree"})
	}
	return configureFlags
}

func libraryRequiresVersion3(libraryId string) bool {
	switch libraryId {
	case "opencore-amr", "vo-amrwbenc", "lensfun", "aribb24":
		return true
	default:
		return false
	}
}

func deriveLicenseProfileName(selectedLibraries []LibraryChoice, configureFlags []string) string {
	needsGpl := false
	needsNonfree := false
	for _, library := range selectedLibraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
	}
	for _, configureFlag := range configureFlags {
		switch configureFlag {
		case "--enable-nonfree":
			needsNonfree = true
		case "--enable-gpl":
			needsGpl = true
		}
	}
	if needsNonfree {
		return "nonfree-local"
	}
	if needsGpl {
		return "gpl-local"
	}
	return "lgpl-local"
}

func selectedLibrariesRequireVersion3(selectedLibraries []LibraryChoice) bool {
	for _, library := range selectedLibraries {
		if libraryRequiresVersion3(library.LibraryId) {
			return true
		}
	}
	return false
}
