package planning

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type LCatalogValidationIssueLevel string

const (
	LCatalogValidationIssueError   LCatalogValidationIssueLevel = "error"
	LCatalogValidationIssueWarning LCatalogValidationIssueLevel = "warning"
)

type LCatalogValidationIssue struct {
	LevelName  LCatalogValidationIssueLevel `json:"levelName"`
	DomainName LCatalogDomainName           `json:"domainName"`
	PathName   string                       `json:"pathName,omitempty"`
	Message    string                       `json:"message"`
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
		if issue.LevelName == LCatalogValidationIssueError {
			count++
		}
	}
	return count
}

func (report LCatalogValidationReport) LWarningCount() int {
	count := 0
	for _, issue := range report.Issues {
		if issue.LevelName == LCatalogValidationIssueWarning {
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
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, err.Error()))
			continue
		}
		LCatalogRecordKindCheck(&report, file, record, "ffmpeg-version-aware-library")
		libraryId := LCatalogStringField(record, "libraryId")
		LCatalogIdRegister(&report, file, libraryIds, libraryId, file.BaseName, "libraryId")
		LCatalogFfmpegVersionRecordMapCheck(&report, file, record)
	}
	for _, file := range catalog.VersionFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, err.Error()))
			continue
		}
		LCatalogRecordKindCheck(&report, file, record, "ffmpeg-version")
		versionId := LCatalogVersionIdRead(record)
		LCatalogIdRegister(&report, file, versionIds, versionId, file.BaseName, "ffmpeg.version")
	}
	for _, file := range catalog.PresetFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, err.Error()))
			continue
		}
		LCatalogRecordKindCheck(&report, file, record, "ffmpeg-preset")
		presetId := LCatalogStringField(record, "presetId")
		LCatalogIdRegister(&report, file, presetIds, presetId, file.BaseName, "presetId")
		LCatalogFfmpegVersionRecordMapCheck(&report, file, record)
	}
	report.LibraryIds = LCatalogMapKeysSorted(libraryIds)
	report.VersionIds = LCatalogMapKeysSorted(versionIds)
	report.PresetIds = LCatalogMapKeysSorted(presetIds)
	LCatalogCrossReferenceCheck(&report, catalog, libraryIds, versionIds, presetIds)
	return report
}

func LCatalogEmbeddedLoadAndValidate() (LCatalogEmbedded, LCatalogValidationReport, error) {
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

func LCatalogJsonObjectRead(file LCatalogEmbeddedFile) (map[string]any, error) {
	record := map[string]any{}
	if err := json.Unmarshal(file.RawContent, &record); err != nil {
		return nil, fmt.Errorf("decode embedded catalog file %q: %w", file.PathName, err)
	}
	return record, nil
}

func LCatalogRecordKindCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, expectedKind string) {
	actualKind := LCatalogStringField(record, "recordKind")
	if actualKind == "" {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "missing recordKind"))
		return
	}
	if actualKind != expectedKind {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("recordKind %q does not match expected %q", actualKind, expectedKind)))
	}
}

func LCatalogIdRegister(report *LCatalogValidationReport, file LCatalogEmbeddedFile, ids map[string]string, actualId string, expectedId string, fieldName string) {
	if actualId == "" {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "missing "+fieldName))
		return
	}
	if actualId != expectedId {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("%s %q does not match file basename %q", fieldName, actualId, expectedId)))
	}
	if previousPathName, exists := ids[actualId]; exists {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("duplicate id %q already defined in %s", actualId, previousPathName)))
		return
	}
	ids[actualId] = file.PathName
}

func LCatalogFfmpegVersionRecordMapCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any) {
	ffmpegVersions, ok := record["ffmpegVersions"].(map[string]any)
	if !ok {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "missing ffmpegVersions object"))
		return
	}
	if len(ffmpegVersions) == 0 {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "ffmpegVersions object is empty"))
	}
	for versionId, versionRecord := range ffmpegVersions {
		versionObject, ok := versionRecord.(map[string]any)
		if !ok {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("ffmpegVersions.%s is not an object", versionId)))
			continue
		}
		actualVersion := LCatalogStringField(versionObject, "ffmpegVersion")
		if actualVersion != "" && actualVersion != versionId {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("ffmpegVersions key %q does not match embedded ffmpegVersion %q", versionId, actualVersion)))
		}
	}
}

func LCatalogCrossReferenceCheck(report *LCatalogValidationReport, catalog LCatalogEmbedded, libraryIds map[string]string, versionIds map[string]string, presetIds map[string]string) {
	for _, file := range catalog.LibraryFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			continue
		}
		libraryId := LCatalogStringField(record, "libraryId")
		ffmpegVersions, _ := record["ffmpegVersions"].(map[string]any)
		for versionId, versionRecord := range ffmpegVersions {
			if _, exists := versionIds[versionId]; !exists {
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("library %q references unknown FFmpeg version %q", libraryId, versionId)))
			}
			versionObject, ok := versionRecord.(map[string]any)
			if !ok {
				continue
			}
			versionLibraryId := LCatalogStringField(versionObject, "libraryId")
			if versionLibraryId != "" && versionLibraryId != libraryId {
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("ffmpegVersions.%s libraryId %q does not match top-level libraryId %q", versionId, versionLibraryId, libraryId)))
			}
		}
	}
	for _, file := range catalog.VersionFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			continue
		}
		LCatalogVersionLibraryReferencesCheck(report, file, record, libraryIds)
	}
	for _, file := range catalog.PresetFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			continue
		}
		LCatalogPresetReferencesCheck(report, file, record, libraryIds, versionIds)
	}
	_ = presetIds
}

func LCatalogVersionLibraryReferencesCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, libraryIds map[string]string) {
	libraries, ok := record["libraries"].(map[string]any)
	if !ok {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "missing libraries object"))
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
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "library reference in "+sectionName+" is not an object"))
				continue
			}
			libraryId := LCatalogStringField(itemObject, "libraryId")
			if libraryId == "" {
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, "library reference in "+sectionName+" is missing libraryId"))
				continue
			}
			if _, exists := libraryIds[libraryId]; !exists {
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("library reference %q in %s has no library metadata file", libraryId, sectionName)))
			}
			metadataPath := LCatalogStringField(itemObject, "metadata")
			if metadataPath != "" && metadataPath != "libraries/"+libraryId+".json" {
				report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("library %q metadata path %q is not libraries/%s.json", libraryId, metadataPath, libraryId)))
			}
			preparationPath := LCatalogStringField(itemObject, "preparation")
			if preparationPath != "" {
				LCatalogSourcePathCheck(report, file, preparationPath)
			}
		}
	}
}

func LCatalogPresetReferencesCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, libraryIds map[string]string, versionIds map[string]string) {
	ffmpegVersions, ok := record["ffmpegVersions"].(map[string]any)
	if !ok {
		return
	}
	for versionId, versionRecord := range ffmpegVersions {
		if _, exists := versionIds[versionId]; !exists {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("preset references unknown FFmpeg version %q", versionId)))
		}
		versionObject, ok := versionRecord.(map[string]any)
		if !ok {
			continue
		}
		if _, hasModes := versionObject["modes"]; hasModes {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("preset %s must be flat and must not contain modes", versionId)))
		}
		LCatalogPresetReferencesForFlatVersionCheck(report, file, versionId, versionObject, libraryIds)
	}
}

func LCatalogPresetReferencesForFlatVersionCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, versionId string, versionObject map[string]any, libraryIds map[string]string) {
	for _, fieldName := range []string{"libraryIds", "extendedLibraryIds"} {
		LCatalogLibraryIdArrayReferencesCheck(report, file, versionObject, fieldName, libraryIds)
	}
	for _, fieldName := range []string{"declaredLibraryIdsFromCurrentV4Preset", "selectedLibraryIdsForUcrt64", "selectedLibraryIdsForMingw64", "selectedLibraryIdsForClang64", "removedLibraryIdsForUcrt64", "removedLibraryIdsForMingw64", "removedLibraryIdsForClang64"} {
		if _, exists := versionObject[fieldName]; exists {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("preset %s must not contain legacy/result field %s", versionId, fieldName)))
		}
	}
}

func LCatalogLibraryIdArrayReferencesCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, record map[string]any, fieldName string, libraryIds map[string]string) {
	values, ok := record[fieldName].([]any)
	if !ok {
		return
	}
	for _, value := range values {
		libraryId, ok := value.(string)
		if !ok {
			report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fieldName+" contains a non-string library id"))
			continue
		}
		LCatalogLibraryIdReferenceCheck(report, file, libraryId, libraryIds, fieldName)
	}
}

func LCatalogLibraryIdReferenceCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, libraryId string, libraryIds map[string]string, fieldName string) {
	if _, exists := libraryIds[libraryId]; !exists {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("%s references unknown library %q", fieldName, libraryId)))
	}
}

func LCatalogVersionIdRead(record map[string]any) string {
	ffmpeg, ok := record["ffmpeg"].(map[string]any)
	if !ok {
		return ""
	}
	return LCatalogStringField(ffmpeg, "version")
}

func LCatalogStringField(record map[string]any, fieldName string) string {
	value, ok := record[fieldName].(string)
	if !ok {
		return ""
	}
	return value
}

func LCatalogSourcePathCheck(report *LCatalogValidationReport, file LCatalogEmbeddedFile, sourcePath string) {
	cleanPath := filepath.ToSlash(filepath.Clean(sourcePath))
	if strings.HasPrefix(cleanPath, "../") || strings.HasPrefix(cleanPath, "/") {
		report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueError, file, fmt.Sprintf("unsafe source path %q", sourcePath)))
		return
	}
	if strings.HasPrefix(cleanPath, "versions/") && strings.HasSuffix(cleanPath, ".go") {
		return
	}
	report.Issues = append(report.Issues, LCatalogValidationIssueCreate(LCatalogValidationIssueWarning, file, fmt.Sprintf("source path %q is not a recognized version work path", sourcePath)))
}

func LCatalogValidationIssueCreate(levelName LCatalogValidationIssueLevel, file LCatalogEmbeddedFile, message string) LCatalogValidationIssue {
	return LCatalogValidationIssue{
		LevelName:  levelName,
		DomainName: file.DomainName,
		PathName:   file.PathName,
		Message:    message,
	}
}

func LCatalogMapKeysSorted(values map[string]string) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
