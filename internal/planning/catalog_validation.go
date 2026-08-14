package planning

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type LValidationIssueLevel string

const (
	LValidationIssueError   LValidationIssueLevel = "error"
	LValidationIssueWarning LValidationIssueLevel = "warning"
)

type LCatalogValidationIssue struct {
	LevelName  LValidationIssueLevel `json:"levelName"`
	DomainName LCatalogDomainName    `json:"domainName"`
	PathName   string                `json:"pathName,omitempty"`
	Message    string                `json:"message"`
}

type LCatalogValidationReport struct {
	LibraryFileCount int                       `json:"libraryFileCount"`
	VersionFileCount int                       `json:"versionFileCount"`
	PresetFileCount  int                       `json:"presetFileCount"`
	LibraryIds       []string                  `json:"libraryIds"`
	VersionIds       []string                  `json:"versionIds"`
	PresetIds        []string                  `json:"presetIds"`
	Issues           []LCatalogValidationIssue `json:"issues,omitempty"`
}

func (report LCatalogValidationReport) LErrorCount() int {
	count := 0
	for _, issue := range report.Issues {
		if issue.LevelName == LValidationIssueError {
			count++
		}
	}
	return count
}

func (report LCatalogValidationReport) LWarningCount() int {
	count := 0
	for _, issue := range report.Issues {
		if issue.LevelName == LValidationIssueWarning {
			count++
		}
	}
	return count
}

func LCatalogEmbeddedValidate(catalog LCatalogEmbedded) LCatalogValidationReport {
	report := LCatalogValidationReport{
		LibraryFileCount: len(catalog.LibraryFiles),
		VersionFileCount: len(catalog.VersionFiles),
		PresetFileCount:  len(catalog.PresetFiles),
	}
	libraryIds := map[string]string{}
	versionIds := map[string]string{}
	presetIds := map[string]string{}
	for _, file := range catalog.LibraryFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, err.Error()))
			continue
		}
		LRecordKindCheck(&report, file, record, "ffmpeg-version-aware-library")
		libraryId := LCatalogFieldGet(record, "libraryId")
		LCatalogIdRegister(&report, file, libraryIds, libraryId, file.BaseName, "libraryId")
		LVersionRecordCheck(&report, file, record)
	}
	for _, file := range catalog.VersionFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, err.Error()))
			continue
		}
		LRecordKindCheck(&report, file, record, "ffmpeg-version")
		versionId := LVersionIdentifierRead(record)
		LCatalogIdRegister(&report, file, versionIds, versionId, file.BaseName, "ffmpeg.version")
	}
	for _, file := range catalog.PresetFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, err.Error()))
			continue
		}
		LRecordKindCheck(&report, file, record, "ffmpeg-preset")
		presetId := LCatalogFieldGet(record, "presetId")
		LCatalogIdRegister(&report, file, presetIds, presetId, file.BaseName, "presetId")
		LVersionRecordCheck(&report, file, record)
	}
	report.LibraryIds = LMapKeysSort(libraryIds)
	report.VersionIds = LMapKeysSort(versionIds)
	report.PresetIds = LMapKeysSort(presetIds)
	LCrossReferenceCheck(&report, catalog, libraryIds, versionIds, presetIds)
	return report
}

func LCatalogLoadValidate() (LCatalogEmbedded, LCatalogValidationReport, error) {
	catalog, err := LCatalogEmbeddedLoad()
	if err != nil {
		return LCatalogEmbedded{}, LCatalogValidationReport{}, err
	}
	report := LCatalogEmbeddedValidate(catalog)
	if report.LErrorCount() > 0 {
		return catalog, report, fmt.Errorf("embedded catalog validation failed with %d error(s)", report.LErrorCount())
	}
	return catalog, report, nil
}

func LJsonObjectRead(file LCatalogEmbeddedFile) (map[string]any, error) {
	record := map[string]any{}
	if err := json.Unmarshal(file.RawContent, &record); err != nil {
		return nil, fmt.Errorf("decode embedded catalog file %q: %w", file.PathName, err)
	}
	return record, nil
}

func LRecordKindCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, expectedKind string) {
	actualKind := LCatalogFieldGet(record, "recordKind")
	if actualKind == "" {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "missing recordKind"))
		return
	}
	if actualKind != expectedKind {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("recordKind %q does not match expected %q", actualKind, expectedKind)))
	}
}

func LCatalogIdRegister(report *LCatalogValidationReport, file LCatalogEmbeddedFile, ids map[string]string, actualId string, expectedId string, fieldName string) {
	if actualId == "" {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "missing "+fieldName))
		return
	}
	if actualId != expectedId {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("%s %q does not match file basename %q", fieldName, actualId, expectedId)))
	}
	if previousPathName, exists := ids[actualId]; exists {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("duplicate id %q already defined in %s", actualId, previousPathName)))
		return
	}
	ids[actualId] = file.PathName
}

func LVersionRecordCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any) {
	ffmpegVersions, ok := record["ffmpegVersions"].(map[string]any)
	if !ok {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "missing ffmpegVersions object"))
		return
	}
	if len(ffmpegVersions) == 0 {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "ffmpegVersions object is empty"))
	}
	for versionId, versionRecord := range ffmpegVersions {
		versionObject, ok := versionRecord.(map[string]any)
		if !ok {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("ffmpegVersions.%s is not an object", versionId)))
			continue
		}
		actualVersion := LCatalogFieldGet(versionObject, "ffmpegVersion")
		if actualVersion != "" && actualVersion != versionId {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("ffmpegVersions key %q does not match embedded ffmpegVersion %q", versionId, actualVersion)))
		}
	}
}

func LCrossReferenceCheck(report *LCatalogValidationReport, catalog LCatalogEmbedded, libraryIds map[string]string, versionIds map[string]string, presetIds map[string]string) {
	for _, file := range catalog.LibraryFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			continue
		}
		libraryId := LCatalogFieldGet(record, "libraryId")
		ffmpegVersions, _ := record["ffmpegVersions"].(map[string]any)
		for versionId, versionRecord := range ffmpegVersions {
			if _, exists := versionIds[versionId]; !exists {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("library %q references unknown FFmpeg version %q", libraryId, versionId)))
			}
			versionObject, ok := versionRecord.(map[string]any)
			if !ok {
				continue
			}
			versionLibraryId := LCatalogFieldGet(versionObject, "libraryId")
			if versionLibraryId != "" && versionLibraryId != libraryId {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("ffmpegVersions.%s libraryId %q does not match top-level libraryId %q", versionId, versionLibraryId, libraryId)))
			}
			preparationObject, _ := versionObject["preparation"].(map[string]any)
			implementationPath := LCatalogFieldGet(preparationObject, "implementation")
			if implementationPath != "" {
				LSourcePathCheck(report, file, implementationPath, versionId)
			}
		}
	}
	for _, file := range catalog.VersionFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			continue
		}
		LVersionReferenceCheck(report, file, record, libraryIds)
	}
	for _, file := range catalog.PresetFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			continue
		}
		LPresetReferenceCheck(report, file, record, libraryIds, versionIds)
	}
	_ = presetIds
}

func LVersionReferenceCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, libraryIds map[string]string) {
	libraries, ok := record["libraries"].(map[string]any)
	if !ok {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "missing libraries object"))
		return
	}
	for _, sectionName := range []string{"includedByOfficialFfmpegSource", "compatible", "unsupported", "unavailable"} {
		items, ok := libraries[sectionName].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemObject, ok := item.(map[string]any)
			if !ok {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "library reference in "+sectionName+" is not an object"))
				continue
			}
			libraryId := LCatalogFieldGet(itemObject, "libraryId")
			if libraryId == "" {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, "library reference in "+sectionName+" is missing libraryId"))
				continue
			}
			if _, exists := libraryIds[libraryId]; !exists {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("library reference %q in %s has no library metadata file", libraryId, sectionName)))
			}
			metadataPath := LCatalogFieldGet(itemObject, "metadata")
			if metadataPath != "" && metadataPath != "libraries/"+libraryId+".json" {
				report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("library %q metadata path %q is not libraries/%s.json", libraryId, metadataPath, libraryId)))
			}
			preparationPath := LCatalogFieldGet(itemObject, "preparation")
			if preparationPath != "" {
				LSourcePathCheck(report, file, preparationPath, LVersionIdentifierRead(record))
			}
		}
	}
}

func LPresetReferenceCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, libraryIds map[string]string, versionIds map[string]string) {
	ffmpegVersions, ok := record["ffmpegVersions"].(map[string]any)
	if !ok {
		return
	}
	for versionId, versionRecord := range ffmpegVersions {
		if _, exists := versionIds[versionId]; !exists {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("preset references unknown FFmpeg version %q", versionId)))
		}
		versionObject, ok := versionRecord.(map[string]any)
		if !ok {
			continue
		}
		if _, hasModes := versionObject["modes"]; hasModes {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("preset %s must be flat and must not contain modes", versionId)))
		}
		LPresetFlatCheck(report, file, versionId, versionObject, libraryIds)
	}
}

func LPresetFlatCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, versionId string, versionObject map[string]any, libraryIds map[string]string) {
	for _, fieldName := range []string{"libraryIds", "extendedLibraryIds"} {
		LIdentifierArrayCheck(report, file, versionObject, fieldName, libraryIds)
	}
	for _, fieldName := range []string{"declaredLibraryIdsFromCurrentV4Preset", "selectedLibraryIdsForUcrt64", "selectedLibraryIdsForMingw64", "selectedLibraryIdsForClang64", "removedLibraryIdsForUcrt64", "removedLibraryIdsForMingw64", "removedLibraryIdsForClang64"} {
		if _, exists := versionObject[fieldName]; exists {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("preset %s must not contain legacy/result field %s", versionId, fieldName)))
		}
	}
}

func LIdentifierArrayCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, fieldName string, libraryIds map[string]string) {
	values, ok := record[fieldName].([]any)
	if !ok {
		return
	}
	for _, value := range values {
		libraryId, ok := value.(string)
		if !ok {
			report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fieldName+" contains a non-string library id"))
			continue
		}
		LIdentifierReferenceCheck(report, file, libraryId, libraryIds, fieldName)
	}
}

func LIdentifierReferenceCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, libraryId string, libraryIds map[string]string, fieldName string) {
	if _, exists := libraryIds[libraryId]; !exists {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("%s references unknown library %q", fieldName, libraryId)))
	}
}

func LVersionIdentifierRead(record map[string]any) string {
	ffmpeg, ok := record["ffmpeg"].(map[string]any)
	if !ok {
		return ""
	}
	return LCatalogFieldGet(ffmpeg, "version")
}

func LCatalogFieldGet(record map[string]any, fieldName string) string {
	value, ok := record[fieldName].(string)
	if !ok {
		return ""
	}
	return value
}

func LSourcePathCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, sourcePath string, expectedVersion string) {
	cleanPath := filepath.ToSlash(filepath.Clean(sourcePath))
	if strings.HasPrefix(cleanPath, "../") || strings.HasPrefix(cleanPath, "/") {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("unsafe source path %q", sourcePath)))
		return
	}
	if !strings.HasPrefix(cleanPath, "versions/") || !strings.HasSuffix(cleanPath, ".go") {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueWarning, file, fmt.Sprintf("source path %q is not a recognized version work path", sourcePath)))
		return
	}
	pathVersion := LSourcePathVersionGet(cleanPath)
	if expectedVersion != "" && pathVersion != expectedVersion {
		report.Issues = append(report.Issues, LValidationIssueCreate(LValidationIssueError, file, fmt.Sprintf("source path %q targets FFmpeg version %q but the owning record is FFmpeg %q", sourcePath, pathVersion, expectedVersion)))
	}
}

// LSourcePathVersionGet extracts the version segment of a "versions/<version>/<file>.go" path.
func LSourcePathVersionGet(cleanPath string) string {
	segments := strings.Split(cleanPath, "/")
	if len(segments) < 3 {
		return ""
	}
	return segments[1]
}

func LValidationIssueCreate(levelName LValidationIssueLevel, file LCatalogEmbeddedFile, message string) LCatalogValidationIssue {
	return LCatalogValidationIssue{
		LevelName:  levelName,
		DomainName: file.DomainName,
		PathName:   file.PathName,
		Message:    message,
	}
}

func LMapKeysSort(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
