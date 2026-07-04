package catalogfacts

import (
	"embed"
	"encoding/json"
	"fmt"
	"path"
	"sort"
	"strconv"
	"strings"
)

//go:embed catalogdata/libraries/*.json catalogdata/versions/*.json catalogdata/presets/*.json catalogdata/librarysources/*.json
var embeddedCatalog embed.FS

// LCatalogDataFile is one embedded catalog JSON file exposed to planner packages.
type LCatalogDataFile struct {
	Path       string
	Base       string
	RawContent []byte
}

// LCatalogDomainLoad returns the embedded JSON files for one catalog domain.
func LCatalogDomainLoad(domainPath string) ([]LCatalogDataFile, error) {
	entries, err := embeddedCatalog.ReadDir(domainPath)
	if err != nil {
		return nil, err
	}
	files := []LCatalogDataFile{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		pathName := path.Join(domainPath, entry.Name())
		rawContent, err := embeddedCatalog.ReadFile(pathName)
		if err != nil {
			return nil, err
		}
		files = append(files, LCatalogDataFile{
			Path:       pathName,
			Base:       strings.TrimSuffix(entry.Name(), ".json"),
			RawContent: rawContent,
		})
	}
	sort.Slice(files, func(leftIndex, rightIndex int) bool {
		return files[leftIndex].Path < files[rightIndex].Path
	})
	return files, nil
}

// LCatalogFileRead returns one embedded catalog JSON file.
func LCatalogFileRead(pathName string) ([]byte, error) {
	return embeddedCatalog.ReadFile(pathName)
}

type LCatalogLibrarySupport struct {
	MinVersion  string
	Unavailable bool
	SourceBuild bool
}

type LCatalogReleaseSupport struct {
	Libraries map[string]LCatalogLibrarySupport
	Options   []string
}

type LCatalogFile struct {
	Path string
	Base string
	Data map[string]any
}

type LCatalogFacts struct {
	Libraries map[string]map[string]any
	Versions  map[string]map[string]any
}

func LCatalogLoad() (LCatalogFacts, error) {
	loaded := LCatalogFacts{Libraries: map[string]map[string]any{}, Versions: map[string]map[string]any{}}
	libraryFiles, err := LCatalogRecordLoad("catalogdata/libraries")
	if err != nil {
		return LCatalogFacts{}, err
	}
	versionFiles, err := LCatalogRecordLoad("catalogdata/versions")
	if err != nil {
		return LCatalogFacts{}, err
	}
	for _, file := range libraryFiles {
		libraryId := LStringFieldGet(file.Data, "libraryId")
		if libraryId != "" {
			loaded.Libraries[libraryId] = file.Data
		}
	}
	for _, file := range versionFiles {
		ffmpeg, _ := file.Data["ffmpeg"].(map[string]any)
		versionId := LStringFieldGet(ffmpeg, "version")
		if versionId != "" {
			loaded.Versions[versionId] = file.Data
		}
	}
	return loaded, nil
}

func LCatalogRecordLoad(dir string) ([]LCatalogFile, error) {
	entries, err := embeddedCatalog.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := []LCatalogFile{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		pathName := path.Join(dir, entry.Name())
		rawContent, err := embeddedCatalog.ReadFile(pathName)
		if err != nil {
			return nil, err
		}
		record := map[string]any{}
		if err := json.Unmarshal(rawContent, &record); err != nil {
			return nil, fmt.Errorf("decode embedded catalog file %q: %w", pathName, err)
		}
		files = append(files, LCatalogFile{Path: pathName, Base: strings.TrimSuffix(entry.Name(), ".json"), Data: record})
	}
	return files, nil
}

func LReleaseSupportResolve(version string) (LCatalogReleaseSupport, bool) {
	lineKey := LReleaseKeyGet(version)
	if lineKey == "" {
		return LCatalogReleaseSupport{}, false
	}
	if recommended, found := LReleaseRecommendGet(lineKey); found {
		return LReleaseVersionGet(recommended)
	}
	chosenMajor, chosenMinor, ok := LReleaseLineSplit(lineKey)
	if !ok {
		return LCatalogReleaseSupport{}, false
	}
	highestKey := LReleaseHighestGet()
	highestMajor, highestMinor, ok := LReleaseLineSplit(highestKey)
	if !ok {
		return LCatalogReleaseSupport{}, false
	}
	if chosenMajor > highestMajor || (chosenMajor == highestMajor && chosenMinor > highestMinor) {
		return LReleaseLineGet(highestKey)
	}
	return LCatalogReleaseSupport{}, false
}

func LReleaseLineGet(releaseLineKey string) (LCatalogReleaseSupport, bool) {
	version, found := LReleaseRecommendGet(releaseLineKey)
	if !found {
		return LCatalogReleaseSupport{}, false
	}
	return LReleaseVersionGet(version)
}

func LReleaseVersionGet(ffmpegVersion string) (LCatalogReleaseSupport, bool) {
	LCatalogFacts, err := LCatalogLoad()
	if err != nil {
		return LCatalogReleaseSupport{}, false
	}
	versionRecord, found := LCatalogFacts.Versions[ffmpegVersion]
	if !found {
		return LCatalogReleaseSupport{}, false
	}
	release := LCatalogReleaseSupport{Libraries: map[string]LCatalogLibrarySupport{}, Options: LOptionIdentifiersRead(versionRecord)}
	for libraryId, libraryRecord := range LCatalogFacts.Libraries {
		versionObject, exists := LLibraryRecordRead(libraryRecord, ffmpegVersion)
		if !exists || !LBooleanFieldGet(versionObject, "supportedByFfmpeg") {
			continue
		}
		release.Libraries[libraryId] = LCatalogLibrarySupport{
			MinVersion:  LStringFieldGet(versionObject, "ffmpegPkgConfigMinimumVersion"),
			Unavailable: !LBooleanFieldGet(versionObject, "availableInCurrentV4"),
			SourceBuild: LBooleanFieldGet(versionObject, "sourceBuildRequiredByReleaseManifest"),
		}
	}
	return release, true
}

func LReleaseRecommendGet(releaseLineKey string) (string, bool) {
	LCatalogFacts, err := LCatalogLoad()
	if err != nil {
		return "", false
	}
	for versionId, versionRecord := range LCatalogFacts.Versions {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		if LStringFieldGet(ffmpeg, "releaseLine") == releaseLineKey {
			return versionId, true
		}
	}
	return "", false
}

func LReleaseHighestGet() string {
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
	LCatalogFacts, err := LCatalogLoad()
	if err != nil {
		return nil
	}
	lineKeys := []string{}
	for _, versionRecord := range LCatalogFacts.Versions {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		lineKey := LStringFieldGet(ffmpeg, "releaseLine")
		if lineKey != "" {
			lineKeys = append(lineKeys, lineKey)
		}
	}
	return LStringsSortedGet(lineKeys)
}

func (release LCatalogReleaseSupport) LLibrarySupportGet(libraryId string) (LCatalogLibrarySupport, bool) {
	support, supported := release.Libraries[libraryId]
	return support, supported
}

func (release LCatalogReleaseSupport) LOptionSupportCheck(optionId string) bool {
	for _, supportedOptionId := range release.Options {
		if supportedOptionId == optionId {
			return true
		}
	}
	return false
}

func LOptionIdentifiersRead(versionRecord map[string]any) []string {
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
		optionId := LStringFieldGet(itemObject, "optionId")
		if optionId != "" {
			optionIds = append(optionIds, optionId)
		}
	}
	return LStringsSortedGet(optionIds)
}

func LLibraryRecordRead(libraryRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := libraryRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func LStringFieldGet(record map[string]any, fieldName string) string {
	value, ok := record[fieldName].(string)
	if !ok {
		return ""
	}
	return value
}

func LBooleanFieldGet(record map[string]any, fieldName string) bool {
	value, ok := record[fieldName].(bool)
	return ok && value
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

func LStringsSortedGet(values []string) []string {
	seen := map[string]bool{}
	unique := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		unique = append(unique, value)
	}
	sort.Strings(unique)
	return unique
}
