package planning

import "fmt"

func (resolver LCatalogResolver) LSelectionResolve(settings LCatalogResolutionSettings) ([]string, []string) {
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
	versionRecord, exists := LPresetRecordRead(presetRecord, settings.FfmpegVersion)
	if !exists {
		return nil, nil
	}
	selectedLibraryIds := LPresetIdentifiersRead(versionRecord, settings.PresetModeName)
	return selectedLibraryIds, selectedLibraryIds
}

func LPresetIdentifiersRead(versionRecord map[string]any, modeName string) []string {
	// Flat preset shape: every preset/version directly owns its complete library
	// list. The extended list is also direct data, not an inherited subgroup,
	// mode, result cache, or composition of another preset.
	if modeName == "extended" {
		return LArrayFieldGet(versionRecord, "extendedLibraryIds")
	}
	return LArrayFieldGet(versionRecord, "libraryIds")
}

func LPresetRecordRead(presetRecord map[string]any, ffmpegVersion string) (map[string]any, bool) {
	versionRecords, ok := presetRecord["ffmpegVersions"].(map[string]any)
	if !ok {
		return nil, false
	}
	versionRecord, ok := versionRecords[ffmpegVersion].(map[string]any)
	return versionRecord, ok
}

func LLibrarySelectNormalize(ffmpegVersion string, selectedLibraryIds []string, resolvedById map[string]LResolvedLibrary, preserveInvalidSelectedLibraries bool) ([]string, []LWarningPlan) {
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
			if preserveInvalidSelectedLibraries && !LVersionReplacementCheck(libraryId, selectedLibraryIds) {
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
	normalizedLibraryIds = LConflictGroupNormalize(normalizedLibraryIds)
	return LStringsSortedGet(normalizedLibraryIds), warnings
}

func LConflictGroupNormalize(selectedLibraryIds []string) []string {
	selectedSet := map[string]bool{}
	for _, libraryId := range selectedLibraryIds {
		selectedSet[libraryId] = true
	}

	// Match the current catalog's backend selection normalization: TLS is priority-ordered, while
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

func LVersionReplacementCheck(libraryId string, selectedLibraryIds []string) bool {
	if libraryId == "libvpl" && LStringContainsCheck(selectedLibraryIds, "libmfx") {
		return true
	}
	return false
}
