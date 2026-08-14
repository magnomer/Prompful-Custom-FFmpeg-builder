package planning

import (
	"fmt"
	"strconv"
	"strings"
)

// LLibrarySupport is one library's FFmpeg-release support fact resolved from the
// embedded /libraries and /versions catalog. It replaces the retired
// shared/releasesupport data package; the old files are archived under /backup
// and must not be imported by live code.
type LLibrarySupport struct {
	MinVersion  string `json:"minVersion,omitempty"`
	Unavailable bool   `json:"unavailable,omitempty"`
	SourceBuild bool   `json:"sourceBuild,omitempty"`
}

// LReleaseSupport is one FFmpeg release line's support surface resolved from the
// embedded catalog. Libraries are keyed by builder library id; Options are the
// configure-option ids supported by that FFmpeg release.
type LReleaseSupport struct {
	Libraries map[string]LLibrarySupport `json:"libraries"`
	Options   []string                   `json:"options"`
}

func LReleaseLineGet(releaseLineKey string) (LReleaseSupport, bool) {
	version, found := LReleaseRecommendGet(releaseLineKey)
	if !found {
		return LReleaseSupport{}, false
	}
	return LReleaseVersionGet(version)
}

func LReleaseSupportResolve(version string) (LReleaseSupport, bool) {
	lineKey := LReleaseKeyGet(version)
	if lineKey == "" {
		return LReleaseSupport{}, false
	}
	if recommended, found := LReleaseRecommendGet(lineKey); found {
		return LReleaseVersionGet(recommended)
	}
	chosenMajor, chosenMinor, ok := LReleaseLineSplit(lineKey)
	if !ok {
		return LReleaseSupport{}, false
	}
	highestKey := LReleaseNewestGet()
	highestMajor, highestMinor, ok := LReleaseLineSplit(highestKey)
	if !ok {
		return LReleaseSupport{}, false
	}
	if chosenMajor > highestMajor || (chosenMajor == highestMajor && chosenMinor > highestMinor) {
		return LReleaseLineGet(highestKey)
	}
	return LReleaseSupport{}, false
}

func LReleaseVersionGet(ffmpegVersion string) (LReleaseSupport, bool) {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return LReleaseSupport{}, false
	}
	versionRecord, found := resolver.VersionRecords[ffmpegVersion]
	if !found {
		return LReleaseSupport{}, false
	}
	release := LReleaseSupport{Libraries: map[string]LLibrarySupport{}, Options: LCatalogOptionRead(versionRecord)}
	for libraryId, libraryRecord := range resolver.LibraryRecords {
		versionObject, exists := LVersionRecordRead(libraryRecord, ffmpegVersion)
		if !exists {
			continue
		}
		if !LCatalogBooleanGet(versionObject, "supportedByFfmpeg") {
			continue
		}
		release.Libraries[libraryId] = LLibrarySupport{
			MinVersion:  LCatalogFieldGet(versionObject, "ffmpegPkgConfigMinimumVersion"),
			Unavailable: !LCatalogBooleanGet(versionObject, "provisionableInCurrentCatalog"),
			SourceBuild: LCatalogBooleanGet(versionObject, "sourceBuildRequiredByReleaseManifest"),
		}
	}
	return release, true
}

func LCatalogOptionRead(versionRecord map[string]any) []string {
	configureOptions, ok := versionRecord["configureOptions"].(map[string]any)
	if !ok {
		return nil
	}
	items, ok := configureOptions["supported"].([]any)
	if !ok {
		return nil
	}
	optionIds := []string{}
	for _, item := range items {
		itemObject, ok := item.(map[string]any)
		if !ok {
			continue
		}
		optionId := LCatalogFieldGet(itemObject, "optionId")
		if optionId != "" {
			optionIds = append(optionIds, optionId)
		}
	}
	return LStringsSortedGet(optionIds)
}

func (release LReleaseSupport) LLibrarySupportGet(libraryId string) (LLibrarySupport, bool) {
	support, supported := release.Libraries[libraryId]
	return support, supported
}

func (release LReleaseSupport) LLibraryAvailableCheck(libraryId string) bool {
	support, supported := release.Libraries[libraryId]
	return supported && !support.Unavailable
}

func (release LReleaseSupport) LOptionSupportCheck(optionId string) bool {
	for _, supportedOptionId := range release.Options {
		if supportedOptionId == optionId {
			return true
		}
	}
	return false
}

func LReleaseRecommendGet(releaseLineKey string) (string, bool) {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return "", false
	}
	for versionId, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		if LCatalogFieldGet(ffmpeg, "releaseLine") == releaseLineKey {
			return versionId, true
		}
	}
	return "", false
}

func LReleaseNewestGet() string {
	best := ""
	bestMajor, bestMinor := -1, -1
	for _, lineKey := range LReleaseListGet() {
		major, minor, ok := LReleaseLineSplit(lineKey)
		if !ok {
			continue
		}
		if major > bestMajor || (major == bestMajor && minor > bestMinor) {
			bestMajor, bestMinor, best = major, minor, lineKey
		}
	}
	return best
}

func LReleaseListGet() []string {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return nil
	}
	lineKeys := []string{}
	for _, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		lineKey := LCatalogFieldGet(ffmpeg, "releaseLine")
		if lineKey != "" {
			lineKeys = append(lineKeys, lineKey)
		}
	}
	return LStringsSortedGet(lineKeys)
}

func LReleaseKeyGet(version string) string {
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

func LReleaseLineSplit(lineKey string) (int, int, bool) {
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

func LVersionCompare(left string, right string) (int, bool) {
	if !LVersionDottedCheck(left) || !LVersionDottedCheck(right) {
		return 0, false
	}
	return LVersionSemverCompare(left, right), true
}

func LVersionSemverCompare(left string, right string) int {
	leftParts := strings.Split(left, ".")
	rightParts := strings.Split(right, ".")
	for index := 0; index < len(leftParts) || index < len(rightParts); index++ {
		leftValue := LVersionNumberGet(leftParts, index)
		rightValue := LVersionNumberGet(rightParts, index)
		if leftValue < rightValue {
			return -1
		}
		if leftValue > rightValue {
			return 1
		}
	}
	return 0
}

func LVersionNumberGet(parts []string, index int) int {
	if index >= len(parts) {
		return 0
	}
	value, err := strconv.Atoi(strings.TrimSpace(parts[index]))
	if err != nil {
		return 0
	}
	return value
}

func LVersionDottedCheck(version string) bool {
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

func LReleaseRequiredGet(lineKey string) (string, error) {
	version, found := LReleaseRecommendGet(lineKey)
	if !found {
		return "", fmt.Errorf("unknown FFmpeg release line %q", lineKey)
	}
	return version, nil
}
