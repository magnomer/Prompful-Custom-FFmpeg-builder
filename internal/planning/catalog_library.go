package planning

import (
	"fmt"
	"strings"
)

func (resolver LCatalogResolver) LLibraryResolve(ffmpegVersion string, libraryId string, shellProfileName string) (LResolvedLibrary, error) {
	libraryRecord, exists := resolver.LibraryRecords[libraryId]
	if !exists {
		return LResolvedLibrary{}, fmt.Errorf("library %q has no embedded library record", libraryId)
	}
	versionObject, exists := LVersionRecordRead(libraryRecord, ffmpegVersion)
	if !exists {
		return LResolvedLibrary{}, fmt.Errorf("library %q has no record for FFmpeg %q", libraryId, ffmpegVersion)
	}
	preparationObject, _ := versionObject["preparation"].(map[string]any)
	preparationStatus := LPreparationStatusRead(preparationObject)
	workIds := []string{}
	if preparationStatus.Required && preparationStatus.Implemented {
		workIds = append(workIds, LWorkIdentifierCreate(ffmpegVersion, libraryId))
	}
	profileUnavailable := LShellProfileCheck(versionObject, shellProfileName)
	packageNames := LShellPackageRead(versionObject, shellProfileName)
	if preparationStatus.Required || LCatalogBooleanGet(versionObject, "sourceBuildRequiredByReleaseManifest") {
		packageNames = nil
	}
	return LResolvedLibrary{
		LibraryId:            libraryId,
		DisplayName:          LStringDefaultGet(versionObject, "displayName", libraryId),
		CategoryName:         LCatalogFieldGet(versionObject, "categoryName"),
		TrackName:            LLibraryTrack(LCatalogFieldGet(versionObject, "trackName")),
		SupportState:         LCatalogSupportResolve(versionObject, preparationStatus, profileUnavailable),
		ConfigureFlags:       LArrayFieldGet(versionObject, "ffmpegConfigureFlags"),
		PackageNames:         packageNames,
		OfficialWebpageUrl:   LCatalogFieldGet(versionObject, "officialWebpageUrl"),
		LicenseEffectName:    LCatalogFieldGet(versionObject, "licenseEffectName"),
		PlainExplanation:     LCatalogFieldGet(versionObject, "plainExplanation"),
		TechnicalExplanation: LCatalogFieldGet(versionObject, "technicalExplanation"),
		DefaultChecked:       LCatalogBooleanGet(versionObject, "defaultChecked"),
		Locked:               LCatalogBooleanGet(versionObject, "locked"),
		WorkIds:              workIds,
		PreparationStatus:    preparationStatus,
		UnavailableReasons:   LCatalogReasonRead(versionObject),
		UnavailableProfiles:  LArrayFieldGet(versionObject, "unavailableShellProfiles"),
		VersionCompatibility: &LLibraryCompatibility{
			Supported:  LCatalogBooleanGet(versionObject, "supportedByFfmpeg"),
			Available:  LCatalogBooleanGet(versionObject, "provisionableInCurrentCatalog"),
			MinVersion: LCatalogFieldGet(versionObject, "ffmpegPkgConfigMinimumVersion"),
		},
	}, nil
}

func LCatalogSupportResolve(versionObject map[string]any, preparationStatus *LLibraryPreparationStatus, profileUnavailable bool) LLibrarySupportState {
	stateName := LCatalogFieldGet(versionObject, "state")
	if strings.Contains(stateName, "unsupported") || !LCatalogBooleanGet(versionObject, "supportedByFfmpeg") {
		return LLibrarySupportUnsupported
	}
	if profileUnavailable || strings.Contains(stateName, "unavailable") || !LCatalogBooleanGet(versionObject, "provisionableInCurrentCatalog") {
		return LLibrarySupportUnavailable
	}
	if LCatalogReasonCheck(versionObject, "disabled-in-current-catalog-ui") {
		return LUiDisabled
	}
	if preparationStatus != nil && preparationStatus.Required {
		if !preparationStatus.Implemented {
			return LPreparationMissing
		}
		return LSourceBuildRequired
	}
	if LCatalogBooleanGet(versionObject, "sourceBuildRequiredByReleaseManifest") {
		return LSourceBuildRequired
	}
	if stateName == "compatible" {
		return LLibrarySupportSupported
	}
	return LLibrarySupportUnknown
}

func LPreparationStatusRead(preparationObject map[string]any) *LLibraryPreparationStatus {
	if preparationObject == nil {
		return nil
	}
	implementation := LCatalogFieldGet(preparationObject, "implementation")
	required := LCatalogBooleanGet(preparationObject, "required")
	status := &LLibraryPreparationStatus{
		Required:               required,
		Kind:                   LCatalogFieldGet(preparationObject, "kind"),
		Implemented:            required && implementation != "",
		Implementation:         implementation,
		ImplementationLanguage: LCatalogFieldGet(preparationObject, "implementationLanguage"),
		Reason:                 LCatalogFieldGet(preparationObject, "reason"),
	}
	if !required && status.Kind == "" && status.Implementation == "" && status.Reason == "" {
		return nil
	}
	return status
}

func LCatalogReasonRead(versionObject map[string]any) []string {
	uiState, ok := versionObject["currentCatalogUiState"].(map[string]any)
	if !ok {
		return nil
	}
	return LArrayFieldGet(uiState, "reasonsWhenUnavailable")
}

func LCatalogReasonCheck(versionObject map[string]any, reason string) bool {
	for _, unavailableReason := range LCatalogReasonRead(versionObject) {
		if unavailableReason == reason {
			return true
		}
	}
	return false
}

func LVersionRecordRead(libraryRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := libraryRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func LLibraryOrderRead(versionRecord map[string]any) []string {
	librariesObject, ok := versionRecord["libraries"].(map[string]any)
	if !ok {
		return nil
	}
	libraryIds := []string{}
	for _, sectionName := range []string{"includedByOfficialFfmpegSource", "compatible", "unavailableInCurrentCatalog", "unsupportedByThisFfmpegVersion", "unsupported", "unavailable"} {
		items, ok := librariesObject[sectionName].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemObject, ok := item.(map[string]any)
			if !ok {
				continue
			}
			libraryId := LCatalogFieldGet(itemObject, "libraryId")
			if libraryId != "" {
				libraryIds = append(libraryIds, libraryId)
			}
		}
	}
	return LStringsUniqueGet(libraryIds)
}
