// Package librarysources holds the per-FFmpeg-release download pins for source-built
// and imported libraries, decoupled from Go code so supporting a new FFmpeg release or
// re-pinning a library is a data edit (library-sources.json), not a code change.
//
// Resolution for a given FFmpeg version: an exact release-key match wins; otherwise the
// highest recorded FFmpeg release is used as a fallback (its pins are the newest the
// project has curated, so they are the best guess for an even-newer, unrecorded FFmpeg).
package librarysources

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
)

//go:embed library-sources.json
var librarySourcesFile []byte

// LibrarySource is one library's download pin: the only data that changes between
// releases. The build/verify logic lives in Go, keyed by the same library id.
type LibrarySource struct {
	Version           string   `json:"version"`
	Url               string   `json:"url"`
	Sha256            string   `json:"sha256"`
	Format            string   `json:"format"`
	Host              string   `json:"host"`
	ExtraCMakeOptions []string `json:"extraCMakeOptions,omitempty"`
}

type ffmpegRelease struct {
	Libraries map[string]LibrarySource `json:"libraries"`
}

type catalog struct {
	SchemaVersion  int                      `json:"schemaVersion"`
	FfmpegReleases map[string]ffmpegRelease `json:"ffmpegReleases"`
}

var loadedCatalog = mustLoadCatalog()

func mustLoadCatalog() catalog {
	var parsed catalog
	if err := json.Unmarshal(librarySourcesFile, &parsed); err != nil {
		panic(fmt.Sprintf("library-sources.json is invalid: %v", err))
	}
	if len(parsed.FfmpegReleases) == 0 {
		panic("library-sources.json declares no FFmpeg releases")
	}
	return parsed
}

// ResolveLibrarySource returns the download pin for a library given the FFmpeg version
// being built. It prefers an exact FFmpeg-version match, then falls back to the highest
// recorded release. Returns false when no recorded release pins that library.
func ResolveLibrarySource(ffmpegVersion string, libraryId string) (LibrarySource, bool) {
	if release, exists := loadedCatalog.FfmpegReleases[ffmpegVersion]; exists {
		if source, ok := release.Libraries[libraryId]; ok {
			return source, true
		}
	}
	fallbackKey := highestReleaseKey()
	if fallbackKey == "" || fallbackKey == ffmpegVersion {
		return LibrarySource{}, false
	}
	if source, ok := loadedCatalog.FfmpegReleases[fallbackKey].Libraries[libraryId]; ok {
		return source, true
	}
	return LibrarySource{}, false
}

// ResolvedFfmpegReleaseKey reports which recorded release key a given FFmpeg version
// resolves to (itself when recorded, otherwise the highest recorded release). Useful for
// surfacing the fallback in diagnostics.
func ResolvedFfmpegReleaseKey(ffmpegVersion string) string {
	if _, exists := loadedCatalog.FfmpegReleases[ffmpegVersion]; exists {
		return ffmpegVersion
	}
	return highestReleaseKey()
}

// highestReleaseKey returns the recorded FFmpeg release with the greatest semantic
// version. Keys that are not dotted-numeric sort below any numeric key.
func highestReleaseKey() string {
	highest := ""
	for key := range loadedCatalog.FfmpegReleases {
		if highest == "" || compareSemver(key, highest) > 0 {
			highest = key
		}
	}
	return highest
}

// compareSemver compares two dotted-numeric version strings, returning -1, 0, or 1.
// Non-numeric components compare as 0 (lowest), so a malformed key never outranks a
// well-formed one.
func compareSemver(left string, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for index := 0; index < len(leftParts) || index < len(rightParts); index++ {
		leftValue := numericComponent(leftParts, index)
		rightValue := numericComponent(rightParts, index)
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func numericComponent(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(parts[index]))
	if err != nil {
		return 0
	}
	return value
}

// CompareVersions exposes semantic version comparison for callers that validate a pinned
// version against an FFmpeg-required minimum. Returns -1, 0, or 1, plus false when either
// version is not dotted-numeric (so the caller can skip an undecidable comparison, e.g. a
// moving "master" pin).
func CompareVersions(left string, right string) (int, bool) {
	if !isDottedNumeric(left) || !isDottedNumeric(right) {
		return 0, false
	}
	return compareSemver(left, right), true
}

func isDottedNumeric(version string) bool {
	if version == "" {
		return false
	}
	for _, part := range strings.Split(version, ".") {
		if _, err := strconv.Atoi(strings.TrimSpace(part)); err != nil {
			return false
		}
	}
	return true
}
