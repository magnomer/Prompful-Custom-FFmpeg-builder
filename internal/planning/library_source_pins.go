package planning

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"promptfulcustomffmpegbuilder/internal/catalogfacts"
)

// LLibrarySourcePin is the archive pin for one source-built or external library.
type LLibrarySourcePin struct {
	Version           string   `json:"version"`
	Url               string   `json:"url"`
	Sha256            string   `json:"sha256"`
	Format            string   `json:"format"`
	Host              string   `json:"host"`
	ExtraCMakeOptions []string `json:"extraCMakeOptions,omitempty"`
}

type LLibrarySourceCatalog struct {
	SchemaVersion   int                              `json:"schemaVersion"`
	LReleaseFFmpegs map[string]LSourceCatalogRelease `json:"LReleaseFFmpegs"`
}

type LSourceCatalogRelease struct {
	Libraries map[string]LLibrarySourcePin `json:"libraries"`
}

func LSourceSpecificResolve(ffmpegVersion string, libraryId string) (LLibrarySourcePin, bool) {
	catalog, err := LLibrarySourceLoad()
	if err != nil {
		return LLibrarySourcePin{}, false
	}
	return catalog.LSourceResolve(ffmpegVersion, libraryId)
}

func LLibrarySourceLoad() (LLibrarySourceCatalog, error) {
	rawContent, err := catalogfacts.LCatalogFileRead("catalogdata/librarysources/library-sources.json")
	if err != nil {
		return LLibrarySourceCatalog{}, fmt.Errorf("read embedded library source catalog: %w", err)
	}
	var catalog LLibrarySourceCatalog
	if err := json.Unmarshal(rawContent, &catalog); err != nil {
		return LLibrarySourceCatalog{}, fmt.Errorf("decode embedded library source catalog: %w", err)
	}
	return catalog, nil
}

func (catalog LLibrarySourceCatalog) LSourceResolve(ffmpegVersion string, libraryId string) (LLibrarySourcePin, bool) {
	ffmpegVersion = strings.TrimSpace(ffmpegVersion)
	libraryId = strings.TrimSpace(libraryId)
	if source, exists := catalog.LSourceExactResolve(ffmpegVersion, libraryId); exists {
		return source, true
	}
	fallbackVersion := catalog.LReleasePinnedGet()
	if fallbackVersion == "" || fallbackVersion == ffmpegVersion {
		return LLibrarySourcePin{}, false
	}
	return catalog.LSourceExactResolve(fallbackVersion, libraryId)
}

func (catalog LLibrarySourceCatalog) LSourceExactResolve(ffmpegVersion string, libraryId string) (LLibrarySourcePin, bool) {
	release, exists := catalog.LReleaseFFmpegs[ffmpegVersion]
	if !exists {
		return LLibrarySourcePin{}, false
	}
	source, exists := release.Libraries[libraryId]
	if !exists {
		return LLibrarySourcePin{}, false
	}
	return source, strings.TrimSpace(source.Url) != "" && strings.TrimSpace(source.Sha256) != ""
}

func (catalog LLibrarySourceCatalog) LReleasePinnedGet() string {
	highest := ""
	for version := range catalog.LReleaseFFmpegs {
		if highest == "" || LVersionKeyCompare(version, highest) > 0 {
			highest = version
		}
	}
	return highest
}

func LVersionKeyCompare(left string, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for index := 0; index < len(leftParts) || index < len(rightParts); index++ {
		leftValue := LVersionPartGet(leftParts, index)
		rightValue := LVersionPartGet(rightParts, index)
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func LVersionPartGet(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(parts[index]))
	if err != nil {
		return 0
	}
	return value
}
