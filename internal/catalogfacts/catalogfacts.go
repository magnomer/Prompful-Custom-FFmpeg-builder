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

// CatalogDataFile is one embedded catalog JSON file exposed to planner packages.
type CatalogDataFile struct {
	Path       string
	Base       string
	RawContent []byte
}

// CatalogDataDomainFilesLoad returns the embedded JSON files for one catalog domain.
func CatalogDataDomainFilesLoad(domainPath string) ([]CatalogDataFile, error) {
	entries, err := embeddedCatalog.ReadDir(domainPath)
	if err != nil {
		return nil, err
	}
	files := []CatalogDataFile{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		pathName := path.Join(domainPath, entry.Name())
		rawContent, err := embeddedCatalog.ReadFile(pathName)
		if err != nil {
			return nil, err
		}
		files = append(files, CatalogDataFile{
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

// CatalogDataFileRead returns one embedded catalog JSON file.
func CatalogDataFileRead(pathName string) ([]byte, error) {
	return embeddedCatalog.ReadFile(pathName)
}

type LibrarySupport struct {
	MinVersion  string
	Unavailable bool
	SourceBuild bool
}

type ReleaseSupport struct {
	Libraries map[string]LibrarySupport
	Options   []string
}

type catalogFile struct {
	Path string
	Base string
	Data map[string]any
}

type catalog struct {
	Libraries map[string]map[string]any
	Versions  map[string]map[string]any
}

func loadCatalog() (catalog, error) {
	loaded := catalog{Libraries: map[string]map[string]any{}, Versions: map[string]map[string]any{}}
	libraryFiles, err := loadDomain("catalogdata/libraries")
	if err != nil {
		return catalog{}, err
	}
	versionFiles, err := loadDomain("catalogdata/versions")
	if err != nil {
		return catalog{}, err
	}
	for _, file := range libraryFiles {
		libraryId := stringField(file.Data, "libraryId")
		if libraryId != "" {
			loaded.Libraries[libraryId] = file.Data
		}
	}
	for _, file := range versionFiles {
		ffmpeg, _ := file.Data["ffmpeg"].(map[string]any)
		versionId := stringField(ffmpeg, "version")
		if versionId != "" {
			loaded.Versions[versionId] = file.Data
		}
	}
	return loaded, nil
}

func loadDomain(dir string) ([]catalogFile, error) {
	entries, err := embeddedCatalog.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	files := []catalogFile{}
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
		files = append(files, catalogFile{Path: pathName, Base: strings.TrimSuffix(entry.Name(), ".json"), Data: record})
	}
	return files, nil
}

func ReleaseSupportResolve(version string) (ReleaseSupport, bool) {
	lineKey := ReleaseLineKeyGet(version)
	if lineKey == "" {
		return ReleaseSupport{}, false
	}
	if recommended, found := ReleaseRecommendedForLineGet(lineKey); found {
		return ReleaseSupportForVersionGet(recommended)
	}
	chosenMajor, chosenMinor, ok := ReleaseLineSplit(lineKey)
	if !ok {
		return ReleaseSupport{}, false
	}
	highestKey := ReleaseLineHighestGet()
	highestMajor, highestMinor, ok := ReleaseLineSplit(highestKey)
	if !ok {
		return ReleaseSupport{}, false
	}
	if chosenMajor > highestMajor || (chosenMajor == highestMajor && chosenMinor > highestMinor) {
		return ReleaseLineGet(highestKey)
	}
	return ReleaseSupport{}, false
}

func ReleaseLineGet(releaseLineKey string) (ReleaseSupport, bool) {
	version, found := ReleaseRecommendedForLineGet(releaseLineKey)
	if !found {
		return ReleaseSupport{}, false
	}
	return ReleaseSupportForVersionGet(version)
}

func ReleaseSupportForVersionGet(ffmpegVersion string) (ReleaseSupport, bool) {
	catalog, err := loadCatalog()
	if err != nil {
		return ReleaseSupport{}, false
	}
	versionRecord, found := catalog.Versions[ffmpegVersion]
	if !found {
		return ReleaseSupport{}, false
	}
	release := ReleaseSupport{Libraries: map[string]LibrarySupport{}, Options: supportedOptionIdsRead(versionRecord)}
	for libraryId, libraryRecord := range catalog.Libraries {
		versionObject, exists := libraryVersionRecordRead(libraryRecord, ffmpegVersion)
		if !exists || !boolField(versionObject, "supportedByFfmpeg") {
			continue
		}
		release.Libraries[libraryId] = LibrarySupport{
			MinVersion:  stringField(versionObject, "ffmpegPkgConfigMinimumVersion"),
			Unavailable: !boolField(versionObject, "availableInCurrentV4"),
			SourceBuild: boolField(versionObject, "sourceBuildRequiredByReleaseManifest"),
		}
	}
	return release, true
}

func ReleaseRecommendedForLineGet(releaseLineKey string) (string, bool) {
	catalog, err := loadCatalog()
	if err != nil {
		return "", false
	}
	for versionId, versionRecord := range catalog.Versions {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		if stringField(ffmpeg, "releaseLine") == releaseLineKey {
			return versionId, true
		}
	}
	return "", false
}

func ReleaseLineHighestGet() string {
	best := ""
	bestMajor, bestMinor := -1, -1
	for _, lineKey := range ReleaseLineListGet() {
		major, minor, ok := ReleaseLineSplit(lineKey)
		if !ok {
			continue
		}
		if major > bestMajor || (major == bestMajor && minor > bestMinor) {
			bestMajor, bestMinor, best = major, minor, lineKey
		}
	}
	return best
}

func ReleaseLineListGet() []string {
	catalog, err := loadCatalog()
	if err != nil {
		return nil
	}
	lineKeys := []string{}
	for _, versionRecord := range catalog.Versions {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		lineKey := stringField(ffmpeg, "releaseLine")
		if lineKey != "" {
			lineKeys = append(lineKeys, lineKey)
		}
	}
	return stringsUniqueSortedStable(lineKeys)
}

func (release ReleaseSupport) LibrarySupportGet(libraryId string) (LibrarySupport, bool) {
	support, supported := release.Libraries[libraryId]
	return support, supported
}

func (release ReleaseSupport) OptionSupportCheck(optionId string) bool {
	for _, supportedOptionId := range release.Options {
		if supportedOptionId == optionId {
			return true
		}
	}
	return false
}

func supportedOptionIdsRead(versionRecord map[string]any) []string {
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
		optionId := stringField(itemObject, "optionId")
		if optionId != "" {
			optionIds = append(optionIds, optionId)
		}
	}
	return stringsUniqueSortedStable(optionIds)
}

func libraryVersionRecordRead(libraryRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := libraryRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func stringField(record map[string]any, fieldName string) string {
	value, ok := record[fieldName].(string)
	if !ok {
		return ""
	}
	return value
}

func boolField(record map[string]any, fieldName string) bool {
	value, ok := record[fieldName].(bool)
	return ok && value
}

func ReleaseLineKeyGet(version string) string {
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

func ReleaseLineSplit(lineKey string) (int, int, bool) {
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

func stringsUniqueSortedStable(values []string) []string {
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
