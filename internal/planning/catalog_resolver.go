package planning

import (
	"fmt"
	"sort"
	"strings"
)

const (
	LCatalogDefaultPresetModeName          = "normal"
	LCatalogDefaultWindowsShellProfileName = "ucrt64"
)

// LCatalogResolutionSettings describes one version-aware catalog resolution request.
type LCatalogResolutionSettings struct {
	FfmpegVersion           string   `json:"ffmpegVersion"`
	PresetId                string   `json:"presetId,omitempty"`
	PresetModeName          string   `json:"presetModeName,omitempty"`
	WindowsShellProfileName string   `json:"windowsShellProfileName,omitempty"`
	SelectedLibraryIds      []string `json:"selectedLibraryIds,omitempty"`
}

type LCatalogResolver struct {
	Catalog          LCatalogEmbedded
	LibraryRecords   map[string]map[string]any
	VersionRecords   map[string]map[string]any
	PresetRecords    map[string]map[string]any
	LibraryFilesById map[string]LCatalogEmbeddedFile
	VersionFilesById map[string]LCatalogEmbeddedFile
	PresetFilesById  map[string]LCatalogEmbeddedFile
}

func LCatalogResolverLoad() (LCatalogResolver, LCatalogValidationReport, error) {
	catalog, report, err := LCatalogEmbeddedLoadAndValidate()
	if err != nil {
		return LCatalogResolver{}, report, err
	}
	resolver, err := LCatalogResolverCreate(catalog)
	return resolver, report, err
}

func LCatalogResolverCreate(catalog LCatalogEmbedded) (LCatalogResolver, error) {
	resolver := LCatalogResolver{
		Catalog:          catalog,
		LibraryRecords:   map[string]map[string]any{},
		VersionRecords:   map[string]map[string]any{},
		PresetRecords:    map[string]map[string]any{},
		LibraryFilesById: map[string]LCatalogEmbeddedFile{},
		VersionFilesById: map[string]LCatalogEmbeddedFile{},
		PresetFilesById:  map[string]LCatalogEmbeddedFile{},
	}
	for _, file := range catalog.LibraryFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		libraryId := LCatalogStringField(record, "libraryId")
		resolver.LibraryRecords[libraryId] = record
		resolver.LibraryFilesById[libraryId] = file
	}
	for _, file := range catalog.VersionFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		versionId := LCatalogVersionIdRead(record)
		resolver.VersionRecords[versionId] = record
		resolver.VersionFilesById[versionId] = file
	}
	for _, file := range catalog.PresetFiles {
		record, err := LCatalogJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		presetId := LCatalogStringField(record, "presetId")
		resolver.PresetRecords[presetId] = record
		resolver.PresetFilesById[presetId] = file
	}
	return resolver, nil
}

func LCatalogEmbeddedVersionResolve(settings LCatalogResolutionSettings) (LResolvedVersionPlan, error) {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return LResolvedVersionPlan{}, err
	}
	return resolver.LVersionResolve(settings)
}

func (resolver LCatalogResolver) LVersionResolve(settings LCatalogResolutionSettings) (LResolvedVersionPlan, error) {
	settings = LCatalogResolutionSettingsNormalize(settings)
	catalogFfmpegVersion, exists := resolver.LCatalogVersionForRequestedResolve(settings.FfmpegVersion)
	if !exists {
		return LResolvedVersionPlan{}, fmt.Errorf("unknown FFmpeg version %q", settings.FfmpegVersion)
	}
	versionRecord := resolver.VersionRecords[catalogFfmpegVersion]
	libraryOrder := LCatalogVersionLibraryOrderRead(versionRecord)
	resolvedById := map[string]LResolvedLibrary{}
	visibleLibraries := []LResolvedLibrary{}
	unsupportedLibraries := []LResolvedLibrary{}
	for _, libraryId := range libraryOrder {
		resolvedLibrary, err := resolver.LLibraryResolve(catalogFfmpegVersion, libraryId, settings.WindowsShellProfileName)
		if err != nil {
			return LResolvedVersionPlan{}, err
		}
		resolvedById[libraryId] = resolvedLibrary
		switch resolvedLibrary.SupportState {
		case LLibrarySupportHidden:
			unsupportedLibraries = append(unsupportedLibraries, resolvedLibrary)
		case LLibrarySupportUnsupported, LLibrarySupportUnavailable:
			visibleLibraries = append(visibleLibraries, resolvedLibrary)
			unsupportedLibraries = append(unsupportedLibraries, resolvedLibrary)
		default:
			visibleLibraries = append(visibleLibraries, resolvedLibrary)
		}
	}
	selectedSettings := settings
	selectedSettings.FfmpegVersion = catalogFfmpegVersion
	selectedLibraryIds, _ := resolver.LSelectedLibraryIdsResolve(selectedSettings)
	preserveInvalidSelectedLibraries := len(settings.SelectedLibraryIds) > 0 && settings.PresetId == ""
	normalizedLibraryIds, warnings := LCatalogSelectedLibrariesNormalize(catalogFfmpegVersion, selectedLibraryIds, resolvedById, preserveInvalidSelectedLibraries)
	configureFlags := []string{}
	requiredPackageNames := []string{}
	requiredWorkIds := []string{}
	for _, libraryId := range normalizedLibraryIds {
		resolvedLibrary, exists := resolvedById[libraryId]
		if !exists {
			continue
		}
		configureFlags = append(configureFlags, resolvedLibrary.ConfigureFlags...)
		requiredPackageNames = append(requiredPackageNames, resolvedLibrary.PackageNames...)
		requiredWorkIds = append(requiredWorkIds, resolvedLibrary.WorkIds...)
	}
	configureFlags = LCatalogStringsUniqueSortedStable(configureFlags)
	requiredPackageNames = LCatalogStringsUniqueSortedStable(requiredPackageNames)
	requiredWorkIds = LCatalogStringsUniqueSortedStable(requiredWorkIds)
	return LResolvedVersionPlan{
		FfmpegVersion:              catalogFfmpegVersion,
		RequestedFfmpegVersion:     settings.FfmpegVersion,
		CompatibilityFfmpegVersion: catalogFfmpegVersion,
		VisibleLibraries:           visibleLibraries,
		UnsupportedLibraries:       unsupportedLibraries,
		SelectedLibraryIds:         LCatalogStringsUniqueSortedStable(selectedLibraryIds),
		NormalizedLibraryIds:       normalizedLibraryIds,
		RequiredWorkIds:            requiredWorkIds,
		ConfigureFlags:             configureFlags,
		RequiredPackageNames:       requiredPackageNames,
		Warnings:                   warnings,
	}, nil
}

func (resolver LCatalogResolver) LCatalogVersionForRequestedResolve(requestedVersion string) (string, bool) {
	requestedVersion = strings.TrimSpace(requestedVersion)
	if requestedVersion == "" {
		return "", false
	}
	if _, exists := resolver.VersionRecords[requestedVersion]; exists {
		return requestedVersion, true
	}
	lineKey := LReleaseLineKeyGet(requestedVersion)
	if lineKey == "" {
		return "", false
	}
	if recommendedVersion, found := resolver.LReleaseRecommendedForLineResolve(lineKey); found {
		return recommendedVersion, true
	}
	requestedMajor, requestedMinor, ok := LReleaseLineSplit(lineKey)
	if !ok {
		return "", false
	}
	highestLineKey, highestVersion, found := resolver.LHighestKnownReleaseLineResolve()
	if !found {
		return "", false
	}
	highestMajor, highestMinor, ok := LReleaseLineSplit(highestLineKey)
	if !ok {
		return "", false
	}
	if requestedMajor > highestMajor || (requestedMajor == highestMajor && requestedMinor > highestMinor) {
		return highestVersion, true
	}
	return "", false
}

func (resolver LCatalogResolver) LReleaseRecommendedForLineResolve(releaseLineKey string) (string, bool) {
	bestVersion := ""
	for versionId, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		if LCatalogStringField(ffmpeg, "releaseLine") != releaseLineKey {
			continue
		}
		if bestVersion == "" || LVersionSemverCompare(versionId, bestVersion) > 0 {
			bestVersion = versionId
		}
	}
	return bestVersion, bestVersion != ""
}

func (resolver LCatalogResolver) LHighestKnownReleaseLineResolve() (string, string, bool) {
	bestLineKey := ""
	bestVersion := ""
	bestMajor, bestMinor := -1, -1
	for versionId, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		lineKey := LCatalogStringField(ffmpeg, "releaseLine")
		major, minor, ok := LReleaseLineSplit(lineKey)
		if !ok {
			continue
		}
		if major > bestMajor || (major == bestMajor && minor > bestMinor) || (major == bestMajor && minor == bestMinor && LVersionSemverCompare(versionId, bestVersion) > 0) {
			bestMajor, bestMinor = major, minor
			bestLineKey, bestVersion = lineKey, versionId
		}
	}
	return bestLineKey, bestVersion, bestVersion != ""
}

func LCatalogResolutionSettingsNormalize(settings LCatalogResolutionSettings) LCatalogResolutionSettings {
	settings.FfmpegVersion = strings.TrimSpace(settings.FfmpegVersion)
	settings.PresetId = strings.TrimSpace(settings.PresetId)
	settings.PresetModeName = strings.TrimSpace(settings.PresetModeName)
	settings.WindowsShellProfileName = strings.TrimSpace(settings.WindowsShellProfileName)
	if settings.PresetModeName == "" {
		settings.PresetModeName = LCatalogDefaultPresetModeName
	}
	if settings.WindowsShellProfileName == "" {
		settings.WindowsShellProfileName = LCatalogDefaultWindowsShellProfileName
	}
	settings.SelectedLibraryIds = LCatalogStringsUniqueSortedStable(settings.SelectedLibraryIds)
	return settings
}

func (resolver LCatalogResolver) LLibraryResolve(ffmpegVersion string, libraryId string, shellProfileName string) (LResolvedLibrary, error) {
	libraryRecord, exists := resolver.LibraryRecords[libraryId]
	if !exists {
		return LResolvedLibrary{}, fmt.Errorf("library %q has no embedded library record", libraryId)
	}
	versionObject, exists := LCatalogLibraryVersionRecordRead(libraryRecord, ffmpegVersion)
	if !exists {
		return LResolvedLibrary{}, fmt.Errorf("library %q has no record for FFmpeg %q", libraryId, ffmpegVersion)
	}
	preparationObject, _ := versionObject["preparation"].(map[string]any)
	preparationStatus := LCatalogPreparationStatusRead(preparationObject)
	workIds := []string{}
	if preparationStatus.Required && preparationStatus.Implemented {
		workIds = append(workIds, LCatalogVersionLibraryWorkIdCreate(ffmpegVersion, libraryId))
	}
	profileUnavailable := LCatalogLibraryShellProfileUnavailableCheck(versionObject, shellProfileName)
	packageNames := LCatalogPackageNamesForShellRead(versionObject, shellProfileName)
	if preparationStatus.Required || LCatalogBoolField(versionObject, "sourceBuildRequiredByReleaseManifest") {
		packageNames = nil
	}
	return LResolvedLibrary{
		LibraryId:            libraryId,
		DisplayName:          LCatalogStringFieldDefault(versionObject, "displayName", libraryId),
		CategoryName:         LCatalogStringField(versionObject, "categoryName"),
		TrackName:            LLibraryTrack(LCatalogStringField(versionObject, "trackName")),
		SupportState:         LCatalogSupportStateResolve(versionObject, preparationStatus, profileUnavailable),
		ConfigureFlags:       LCatalogStringArrayField(versionObject, "ffmpegConfigureFlags"),
		PackageNames:         packageNames,
		OfficialWebpageUrl:   LCatalogStringField(versionObject, "officialWebpageUrl"),
		LicenseEffectName:    LCatalogStringField(versionObject, "licenseEffectName"),
		PlainExplanation:     LCatalogStringField(versionObject, "plainExplanation"),
		TechnicalExplanation: LCatalogStringField(versionObject, "technicalExplanation"),
		DefaultChecked:       LCatalogBoolField(versionObject, "defaultChecked"),
		Locked:               LCatalogBoolField(versionObject, "locked"),
		WorkIds:              workIds,
		PreparationStatus:    preparationStatus,
		UnavailableReasons:   LCatalogCurrentV4UiUnavailableReasonsRead(versionObject),
		UnavailableProfiles:  LCatalogStringArrayField(versionObject, "unavailableShellProfiles"),
		VersionCompatibility: &LLibraryCompatibility{
			Supported:  LCatalogBoolField(versionObject, "supportedByFfmpeg"),
			Available:  LCatalogBoolField(versionObject, "availableInCurrentV4"),
			MinVersion: LCatalogStringField(versionObject, "ffmpegPkgConfigMinimumVersion"),
		},
	}, nil
}

func LCatalogSupportStateResolve(versionObject map[string]any, preparationStatus *LLibraryPreparationStatus, profileUnavailable bool) LLibrarySupportState {
	stateName := LCatalogStringField(versionObject, "state")
	if strings.Contains(stateName, "unsupported") || !LCatalogBoolField(versionObject, "supportedByFfmpeg") {
		return LLibrarySupportUnsupported
	}
	if profileUnavailable || strings.Contains(stateName, "unavailable") || !LCatalogBoolField(versionObject, "availableInCurrentV4") {
		return LLibrarySupportUnavailable
	}
	if LCatalogCurrentV4UiUnavailableReasonCheck(versionObject, "disabled-in-current-v4-ui") {
		return LLibrarySupportUiDisabled
	}
	if preparationStatus != nil && preparationStatus.Required {
		if !preparationStatus.Implemented {
			return LLibrarySupportPreparationUnimplemented
		}
		return LLibrarySupportSourceBuildRequired
	}
	if LCatalogBoolField(versionObject, "sourceBuildRequiredByReleaseManifest") {
		return LLibrarySupportSourceBuildRequired
	}
	if stateName == "compatible" {
		return LLibrarySupportSupported
	}
	return LLibrarySupportUnknown
}

func LCatalogPreparationStatusRead(preparationObject map[string]any) *LLibraryPreparationStatus {
	if preparationObject == nil {
		return nil
	}
	implementation := LCatalogStringField(preparationObject, "implementation")
	required := LCatalogBoolField(preparationObject, "required")
	status := &LLibraryPreparationStatus{
		Required:               required,
		Kind:                   LCatalogStringField(preparationObject, "kind"),
		Implemented:            required && implementation != "",
		Implementation:         implementation,
		ImplementationLanguage: LCatalogStringField(preparationObject, "implementationLanguage"),
		Reason:                 LCatalogStringField(preparationObject, "reason"),
	}
	if !required && status.Kind == "" && status.Implementation == "" && status.Reason == "" {
		return nil
	}
	return status
}

func LCatalogCurrentV4UiUnavailableReasonsRead(versionObject map[string]any) []string {
	uiState, ok := versionObject["currentV4UiState"].(map[string]any)
	if !ok {
		return nil
	}
	return LCatalogStringArrayField(uiState, "reasonsWhenUnavailable")
}

func LCatalogCurrentV4UiUnavailableReasonCheck(versionObject map[string]any, reason string) bool {
	for _, unavailableReason := range LCatalogCurrentV4UiUnavailableReasonsRead(versionObject) {
		if unavailableReason == reason {
			return true
		}
	}
	return false
}

func LCatalogLibraryVersionRecordRead(libraryRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := libraryRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func LCatalogVersionLibraryOrderRead(versionRecord map[string]any) []string {
	librariesObject, ok := versionRecord["libraries"].(map[string]any)
	if !ok {
		return nil
	}
	libraryIds := []string{}
	for _, sectionName := range []string{"includedByOfficialFfmpegSource", "compatible", "unavailableInCurrentV4", "unsupportedByThisFfmpegVersion", "unsupported", "unavailable"} {
		items, ok := librariesObject[sectionName].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemObject, ok := item.(map[string]any)
			if !ok {
				continue
			}
			libraryId := LCatalogStringField(itemObject, "libraryId")
			if libraryId != "" {
				libraryIds = append(libraryIds, libraryId)
			}
		}
	}
	return LCatalogStringsUniqueStable(libraryIds)
}

func (resolver LCatalogResolver) LSelectedLibraryIdsResolve(settings LCatalogResolutionSettings) ([]string, []string) {
	if len(settings.SelectedLibraryIds) > 0 {
		return settings.SelectedLibraryIds, nil
	}
	if settings.PresetId == "" {
		return nil, nil
	}
	presetRecord, exists := resolver.PresetRecords[settings.PresetId]
	if !exists {
		return nil, nil
	}
	versionRecord, exists := LCatalogPresetVersionRecordRead(presetRecord, settings.FfmpegVersion)
	if !exists {
		return nil, nil
	}
	selectedLibraryIds := LCatalogPresetLibraryIdsRead(versionRecord, settings.PresetModeName)
	return selectedLibraryIds, selectedLibraryIds
}

func LCatalogPresetLibraryIdsRead(versionRecord map[string]any, modeName string) []string {
	// Flat preset shape: every preset/version directly owns its complete library
	// list. The extended list is also direct data, not an inherited subgroup,
	// mode, result cache, or composition of another preset.
	if modeName == "extended" {
		return LCatalogStringArrayField(versionRecord, "extendedLibraryIds")
	}
	return LCatalogStringArrayField(versionRecord, "libraryIds")
}

func LCatalogPresetVersionRecordRead(presetRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := presetRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func LCatalogSelectedLibrariesNormalize(ffmpegVersion string, selectedLibraryIds []string, resolvedById map[string]LResolvedLibrary, preserveInvalidSelectedLibraries bool) ([]string, []LWarningPlan) {
	normalizedLibraryIds := []string{}
	warnings := []LWarningPlan{}
	for _, libraryId := range selectedLibraryIds {
		resolvedLibrary, exists := resolvedById[libraryId]
		if !exists {
			warnings = append(warnings, LWarningPlan{
				LRiskLevel: LRiskWarning,
				Message:    fmt.Sprintf("Library %s is not known for FFmpeg %s and was removed from the resolved selection.", libraryId, ffmpegVersion),
			})
			continue
		}
		switch resolvedLibrary.SupportState {
		case LLibrarySupportUnsupported, LLibrarySupportUnavailable:
			if preserveInvalidSelectedLibraries && !LCatalogSelectedLibraryVersionReplacementCheck(libraryId, selectedLibraryIds) {
				normalizedLibraryIds = append(normalizedLibraryIds, libraryId)
				continue
			}
			warnings = append(warnings, LWarningPlan{
				LRiskLevel: LRiskWarning,
				Message:    fmt.Sprintf("Library %s is not available for FFmpeg %s and was removed from the resolved preset selection.", libraryId, ffmpegVersion),
			})
		case LLibrarySupportHidden:
			warnings = append(warnings, LWarningPlan{
				LRiskLevel: LRiskWarning,
				Message:    fmt.Sprintf("Library %s is hidden for FFmpeg %s and was removed from the resolved selection.", libraryId, ffmpegVersion),
			})
		default:
			normalizedLibraryIds = append(normalizedLibraryIds, libraryId)
		}
	}
	normalizedLibraryIds = LCatalogSelectedLibraryConflictGroupsNormalize(normalizedLibraryIds)
	return LCatalogStringsUniqueSortedStable(normalizedLibraryIds), warnings
}

func LCatalogSelectedLibraryConflictGroupsNormalize(selectedLibraryIds []string) []string {
	selectedSet := map[string]bool{}
	for _, libraryId := range selectedLibraryIds {
		selectedSet[libraryId] = true
	}

	// Match V4's backend selection normalization: TLS is priority-ordered, while
	// shader compilers, EVC bindings, and Intel acceleration backends are pick-one pairs.
	if selectedSet["openssl"] {
		delete(selectedSet, "gnutls")
		delete(selectedSet, "mbedtls")
		delete(selectedSet, "libtls")
	} else if selectedSet["gnutls"] {
		delete(selectedSet, "mbedtls")
		delete(selectedSet, "libtls")
	} else if selectedSet["mbedtls"] {
		delete(selectedSet, "libtls")
	}
	if selectedSet["shaderc"] && selectedSet["glslang"] {
		delete(selectedSet, "glslang")
	}
	if selectedSet["xevd"] && selectedSet["xevdb"] {
		delete(selectedSet, "xevdb")
	}
	if selectedSet["xeve"] && selectedSet["xeveb"] {
		delete(selectedSet, "xeveb")
	}
	if selectedSet["libvpl"] && selectedSet["libmfx"] {
		delete(selectedSet, "libmfx")
	}

	normalizedLibraryIds := []string{}
	for _, libraryId := range selectedLibraryIds {
		if selectedSet[libraryId] {
			normalizedLibraryIds = append(normalizedLibraryIds, libraryId)
			delete(selectedSet, libraryId)
		}
	}
	return normalizedLibraryIds
}

func LCatalogSelectedLibraryVersionReplacementCheck(libraryId string, selectedLibraryIds []string) bool {
	if libraryId == "libvpl" && LCatalogStringSliceContains(selectedLibraryIds, "libmfx") {
		return true
	}
	return false
}

func LCatalogStringSliceContains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func (resolver LCatalogResolver) LResolvedBuildPlanCreate(settings LCatalogResolutionSettings) (LResolvedBuildPlan, error) {
	settings = LCatalogResolutionSettingsNormalize(settings)
	resolvedVersionPlan, err := resolver.LVersionResolve(settings)
	if err != nil {
		return LResolvedBuildPlan{}, err
	}
	versionRecord := resolver.VersionRecords[resolvedVersionPlan.FfmpegVersion]
	ffmpegObject, _ := versionRecord["ffmpeg"].(map[string]any)
	ffmpegSourceArchiveUrl := LResolvedBuildPlanArchiveUrlResolve(settings.FfmpegVersion, resolvedVersionPlan.FfmpegVersion, ffmpegObject)
	ffmpegSourceSignatureUrl := LResolvedBuildPlanSignatureUrlResolve(settings.FfmpegVersion, resolvedVersionPlan.FfmpegVersion, ffmpegObject, ffmpegSourceArchiveUrl)
	resolvedLibrariesById := map[string]LResolvedLibrary{}
	for _, resolvedLibrary := range resolvedVersionPlan.VisibleLibraries {
		resolvedLibrariesById[resolvedLibrary.LibraryId] = resolvedLibrary
	}
	for _, resolvedLibrary := range resolvedVersionPlan.UnsupportedLibraries {
		resolvedLibrariesById[resolvedLibrary.LibraryId] = resolvedLibrary
	}
	selectedLibraries := []LResolvedLibrary{}
	versionWorks := []LVersionLibraryWork{}
	for _, libraryId := range resolvedVersionPlan.NormalizedLibraryIds {
		resolvedLibrary, exists := resolvedLibrariesById[libraryId]
		if !exists {
			continue
		}
		selectedLibraries = append(selectedLibraries, resolvedLibrary)
		for _, workId := range resolvedLibrary.WorkIds {
			versionWorks = append(versionWorks, LVersionLibraryWork{
				WorkId:        workId,
				FfmpegVersion: resolvedVersionPlan.FfmpegVersion,
				LibraryId:     libraryId,
				GoFilePath:    fmt.Sprintf("versions/%s/%s.go", resolvedVersionPlan.FfmpegVersion, libraryId),
			})
		}
	}
	return LResolvedBuildPlan{
		FfmpegVersion:              resolvedVersionPlan.FfmpegVersion,
		RequestedFfmpegVersion:     settings.FfmpegVersion,
		CompatibilityFfmpegVersion: resolvedVersionPlan.FfmpegVersion,
		FfmpegSourceArchiveUrl:     ffmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegSourceSignatureUrl,
		SelectedLibraries:          selectedLibraries,
		VersionLibraryWorks:        versionWorks,
		RequiredMsys2PackageNames:  resolvedVersionPlan.RequiredPackageNames,
		ConfigureFlags:             resolvedVersionPlan.ConfigureFlags,
		Warnings:                   resolvedVersionPlan.Warnings,
	}, nil
}

func LResolvedBuildPlanArchiveUrlResolve(requestedVersion string, resolvedCatalogVersion string, ffmpegObject map[string]any) string {
	requestedVersion = strings.TrimSpace(requestedVersion)
	resolvedCatalogVersion = strings.TrimSpace(resolvedCatalogVersion)
	if requestedVersion == "" || requestedVersion == resolvedCatalogVersion {
		return LCatalogStringField(ffmpegObject, "archiveUrl")
	}
	if LReleaseLineKeyGet(requestedVersion) == "" {
		return ""
	}
	return LReleaseArchiveResolve(requestedVersion)
}

func LResolvedBuildPlanSignatureUrlResolve(requestedVersion string, resolvedCatalogVersion string, ffmpegObject map[string]any, archiveUrl string) string {
	requestedVersion = strings.TrimSpace(requestedVersion)
	resolvedCatalogVersion = strings.TrimSpace(resolvedCatalogVersion)
	if requestedVersion == "" || requestedVersion == resolvedCatalogVersion {
		return LCatalogStringField(ffmpegObject, "signatureUrl")
	}
	if archiveUrl == "" {
		return ""
	}
	return archiveUrl + ".asc"
}

func LCatalogStringFieldDefault(record map[string]any, fieldName string, defaultValue string) string {
	value := LCatalogStringField(record, fieldName)
	if value == "" {
		return defaultValue
	}
	return value
}

func LCatalogBoolField(record map[string]any, fieldName string) bool {
	if record == nil {
		return false
	}
	value, ok := record[fieldName].(bool)
	return ok && value
}

func LCatalogStringArrayField(record map[string]any, fieldName string) []string {
	values, ok := record[fieldName].([]any)
	if !ok {
		return nil
	}
	result := []string{}
	for _, value := range values {
		stringValue, ok := value.(string)
		if ok && stringValue != "" {
			result = append(result, stringValue)
		}
	}
	return result
}

func LCatalogPackageNamesForShellRead(versionObject map[string]any, shellProfileName string) []string {
	if LCatalogLibraryShellProfileUnavailableCheck(versionObject, shellProfileName) {
		return nil
	}
	packageNamesByShellProfile, ok := versionObject["packageNamesByShellProfile"].(map[string]any)
	if !ok {
		return nil
	}
	return LCatalogStringArrayField(packageNamesByShellProfile, shellProfileName)
}

func LCatalogLibraryShellProfileUnavailableCheck(versionObject map[string]any, shellProfileName string) bool {
	for _, unavailableShellProfileName := range LCatalogStringArrayField(versionObject, "unavailableShellProfiles") {
		if unavailableShellProfileName == shellProfileName {
			return true
		}
	}
	return false
}

func LCatalogVersionLibraryWorkIdCreate(ffmpegVersion string, libraryId string) string {
	return "ffmpeg-" + ffmpegVersion + "-" + libraryId + "-work"
}

func LCatalogStringsUniqueStable(values []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	return result
}

func LCatalogStringsUniqueSortedStable(values []string) []string {
	result := LCatalogStringsUniqueStable(values)
	sort.Strings(result)
	return result
}
