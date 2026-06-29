// Package releasesupport holds the explicit per-FFmpeg-release-line support manifest:
// which builder catalog libraries and configure options each maintained FFmpeg release line
// supports, with FFmpeg's pkg-config minimum where one is pinned. Each release line is
// authored standalone from that tag's configure script (not derived as a diff from another
// release), so adding or auditing a release is a self-contained data edit in
// ffmpeg-release-support.json, never a code change.
//
// Lookup is by release LINE key (major.minor, e.g. "8.0"), because a release line shares one
// configure surface across its patch releases. A line with no manifest entry returns found
// == false, and the planner then skips library/option gating for it (FFmpeg configure stays
// the backstop).
package releasesupport

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:embed ffmpeg-release-support.json
var releaseSupportFile []byte

// LibrarySupport is one supported library's per-release facts. MinVersion is FFmpeg's
// pkg-config minimum for that library in that release, or "" when configure pins none.
// Unavailable marks a library whose --enable switch exists in this release but that the
// package this builder can supply cannot satisfy (e.g. lensfun: FFmpeg gates it by the
// lf_db_create symbol that the available lensfun package lacks). Such a library is present
// in the map (so messaging can say "package too old", not "FFmpeg lacks the feature") but
// is treated as not available by LibraryAvailableFor and blocked by the planner.
// SourceBuild marks a library that IS available in this release line but whose MSYS2 native
// package cannot satisfy FFmpeg's required API, so the builder source-builds a compatible
// pinned version instead (resolved from library-sources.json) on the Internal track. Unlike
// Unavailable, the library is not blocked: the planner flips it Native -> Internal and drops
// its native package. Example: FFmpeg 4.4/5.0 probe libvmaf 1.x (compute_vmaf), removed at
// libvmaf 2.0, so the current 3.x package cannot satisfy them and no 1.x package exists.
type LibrarySupport struct {
	MinVersion  string `json:"minVersion,omitempty"`
	Unavailable bool   `json:"unavailable,omitempty"`
	SourceBuild bool   `json:"sourceBuild,omitempty"`
}

// ReleaseSupport is one FFmpeg release line's full support surface: the supported builder
// catalog library ids (keyed by id) and the supported builder configure-option ids.
type ReleaseSupport struct {
	Libraries map[string]LibrarySupport `json:"libraries"`
	Options   []string                  `json:"options"`
}

type manifest struct {
	SchemaVersion  int                       `json:"schemaVersion"`
	FfmpegReleases map[string]ReleaseSupport `json:"ffmpegReleases"`
}

var loadedManifest = mustLoadManifest()

func mustLoadManifest() manifest {
	var parsed manifest
	if err := json.Unmarshal(releaseSupportFile, &parsed); err != nil {
		panic(fmt.Sprintf("ffmpeg-release-support.json is invalid: %v", err))
	}
	if len(parsed.FfmpegReleases) == 0 {
		panic("ffmpeg-release-support.json declares no FFmpeg releases")
	}
	return parsed
}

// ForReleaseLine returns the support manifest for an FFmpeg release line key (major.minor,
// e.g. "8.0"). found is false when no manifest is recorded for that line.
func ForReleaseLine(releaseLineKey string) (ReleaseSupport, bool) {
	release, found := loadedManifest.FfmpegReleases[releaseLineKey]
	return release, found
}

// ResolveReleaseSupport returns the support manifest that governs a chosen FFmpeg version.
// An exact match on the version's release line is used when present. Otherwise a version NEWER
// than the newest recorded line falls back to that newest line, so a future FFmpeg the program
// does not yet record is gated against the latest recorded surface: a library marked
// unavailable on the latest line stays practically blocked on future versions until a new line
// is recorded that lists it differently (the future-proofing lever — never a global rule).
// A non-future unlisted line (a gap below the newest recorded line, or an unparseable/snapshot
// version) returns resolved=false, and the caller skips gating (FFmpeg configure is the
// backstop). All gating (UI availability, plan verification, pkg-config floors) resolves
// through here so the fallback behaves identically everywhere.
func ResolveReleaseSupport(version string) (ReleaseSupport, bool) {
	lineKey := ReleaseLineKey(version)
	if lineKey == "" {
		return ReleaseSupport{}, false
	}
	if release, found := loadedManifest.FfmpegReleases[lineKey]; found {
		return release, true
	}
	chosenMajor, chosenMinor, ok := splitLineKey(lineKey)
	if !ok {
		return ReleaseSupport{}, false
	}
	highestKey := highestRecordedLineKey()
	highestMajor, highestMinor, ok := splitLineKey(highestKey)
	if !ok {
		return ReleaseSupport{}, false
	}
	if chosenMajor > highestMajor || (chosenMajor == highestMajor && chosenMinor > highestMinor) {
		return loadedManifest.FfmpegReleases[highestKey], true
	}
	return ReleaseSupport{}, false
}

// highestRecordedLineKey returns the newest manifested release line key (e.g. "8.1") by numeric
// major.minor, or "" when the manifest records no parseable line.
func highestRecordedLineKey() string {
	best := ""
	bestMajor, bestMinor := -1, -1
	for lineKey := range loadedManifest.FfmpegReleases {
		major, minor, ok := splitLineKey(lineKey)
		if !ok {
			continue
		}
		if major > bestMajor || (major == bestMajor && minor > bestMinor) {
			bestMajor, bestMinor, best = major, minor, lineKey
		}
	}
	return best
}

// splitLineKey parses a "major.minor" line key into its two integer components.
func splitLineKey(lineKey string) (int, int, bool) {
	parts := strings.Split(lineKey, ".")
	if len(parts) != 2 {
		return 0, 0, false
	}
	major, majorErr := strconv.Atoi(parts[0])
	minor, minorErr := strconv.Atoi(parts[1])
	if majorErr != nil || minorErr != nil {
		return 0, 0, false
	}
	return major, minor, true
}

// LibrarySupportFor reports whether a library id is supported in a release line and, if so,
// its facts. supported is false when the line is manifested but the library is absent (its
// --enable switch does not exist in that release). Note: a supported library may still be
// Unavailable; use LibraryAvailableFor for the build-eligibility question.
func (release ReleaseSupport) LibrarySupportFor(libraryId string) (LibrarySupport, bool) {
	support, supported := release.Libraries[libraryId]
	return support, supported
}

// LibraryAvailableFor reports whether a library can actually be built in this release line:
// FFmpeg has the switch (present in the map) AND the package this builder can supply can
// satisfy it (not marked Unavailable). This is the build-eligibility predicate the UI and
// planner use; LibrarySupportFor answers the narrower "does FFmpeg have the switch" question.
func (release ReleaseSupport) LibraryAvailableFor(libraryId string) bool {
	support, supported := release.Libraries[libraryId]
	return supported && !support.Unavailable
}

// ReleaseLineKey returns the "major.minor" line key of a dotted-numeric version
// (e.g. "8.1.1" -> "8.1"), or "" when the version is not dotted-numeric with at least a
// major and minor component (e.g. a git snapshot). Manifest lookup is by line key because a
// release line shares one configure surface across its patch releases.
func ReleaseLineKey(version string) string {
	parts := strings.Split(strings.TrimSpace(version), ".")
	if len(parts) < 2 {
		return ""
	}
	for _, part := range parts[:2] {
		if _, err := strconv.Atoi(strings.TrimSpace(part)); err != nil {
			return ""
		}
	}
	return strings.TrimSpace(parts[0]) + "." + strings.TrimSpace(parts[1])
}

// OptionSupported reports whether a configure-option id is supported in a release line.
func (release ReleaseSupport) OptionSupported(optionId string) bool {
	for _, supportedOptionId := range release.Options {
		if supportedOptionId == optionId {
			return true
		}
	}
	return false
}
