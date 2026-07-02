package planning

// LCatalogSourceBuildResolved returns the Libraries-tab catalog through the
// embedded version resolver. A known release line resolves to that line's catalog;
// an unknown newer release line uses the highest known catalog line. An empty,
// custom, or unparseable FFmpeg URL still returns no catalog.
func LCatalogSourceBuildResolved(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []LLibraryChoice {
	ffmpegVersion := LVersionArchiveParse(ffmpegSourceArchiveUrl)
	if ffmpegVersion == "" {
		return []LLibraryChoice{}
	}
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return []LLibraryChoice{}
	}
	resolvedPlan, err := resolver.LVersionResolve(LCatalogResolutionSettings{
		FfmpegVersion:           ffmpegVersion,
		WindowsShellProfileName: windowsShellProfileName,
	})
	if err != nil {
		return []LLibraryChoice{}
	}
	return LResolvedLibraryChoicesCreate(resolvedPlan.VisibleLibraries)
}

// LResolvedLibraryChoicesCreate converts resolved library rows into the legacy
// LLibraryChoice shape consumed by the current frontend. The compatibility field
// carries the resolver state forward without forcing the frontend to understand
// the new catalog model yet.
func LResolvedLibraryChoicesCreate(resolvedLibraries []LResolvedLibrary) []LLibraryChoice {
	choices := make([]LLibraryChoice, 0, len(resolvedLibraries))
	for _, resolvedLibrary := range resolvedLibraries {
		choices = append(choices, LResolvedLibraryChoiceCreate(resolvedLibrary))
	}
	return choices
}

func LResolvedLibraryChoiceCreate(resolvedLibrary LResolvedLibrary) LLibraryChoice {
	return LLibraryChoice{
		LibraryId:            resolvedLibrary.LibraryId,
		TrackName:            resolvedLibrary.TrackName,
		DisplayName:          resolvedLibrary.DisplayName,
		CategoryName:         resolvedLibrary.CategoryName,
		ConfigureFlags:       append([]string{}, resolvedLibrary.ConfigureFlags...),
		PackageNames:         append([]string{}, resolvedLibrary.PackageNames...),
		OfficialWebpageUrl:   resolvedLibrary.OfficialWebpageUrl,
		LicenseEffectName:    resolvedLibrary.LicenseEffectName,
		PlainExplanation:     resolvedLibrary.PlainExplanation,
		TechnicalExplanation: resolvedLibrary.TechnicalExplanation,
		DefaultChecked:       resolvedLibrary.DefaultChecked,
		Locked:               resolvedLibrary.Locked,
		SupportState:         resolvedLibrary.SupportState,
		PreparationStatus:    resolvedLibrary.PreparationStatus,
		UnavailableReasons:   append([]string{}, resolvedLibrary.UnavailableReasons...),
		UnavailableProfiles:  append([]string{}, resolvedLibrary.UnavailableProfiles...),
		VersionCompatibility: resolvedLibrary.VersionCompatibility,
	}
}

// LLibraryConfigureFlagMatchFromChoices returns library entries from an already
// resolved catalog whose configure flags overlap with the supplied extra flags.
// This is the version-aware equivalent of LLibraryConfigureFlagMatch; it keeps
// extra raw flags on the same catalog authority as checkbox selections.
func LLibraryConfigureFlagMatchFromChoices(catalog []LLibraryChoice, flags []string, skip []LLibraryChoice) []LLibraryChoice {
	flagSet := map[string]bool{}
	for _, flag := range flags {
		if flag != "" {
			flagSet[flag] = true
		}
	}
	skipIds := map[string]bool{}
	for _, library := range skip {
		skipIds[library.LibraryId] = true
	}
	result := []LLibraryChoice{}
	seen := map[string]bool{}
	for _, library := range catalog {
		if skipIds[library.LibraryId] || seen[library.LibraryId] {
			continue
		}
		for _, flag := range library.ConfigureFlags {
			if flagSet[flag] {
				seen[library.LibraryId] = true
				result = append(result, library)
				break
			}
		}
	}
	return result
}

// LCatalogResolvedPlanFromFFmpegSettings creates the resolver plan associated
// with a current FFmpeg settings object. Known release lines resolve to their
// catalog line; unknown newer release lines resolve through the highest known
// catalog line. Unparseable or older unsupported versions are reported as missing.
func LCatalogResolvedPlanFromFFmpegSettings(ffmpegBuildSettings LSettingsFFmpeg) (LResolvedVersionPlan, bool, error) {
	ffmpegVersion := LVersionArchiveParse(ffmpegBuildSettings.FfmpegSourceArchiveUrl)
	if ffmpegVersion == "" {
		return LResolvedVersionPlan{}, false, nil
	}
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return LResolvedVersionPlan{}, false, err
	}
	resolvedPlan, err := resolver.LVersionResolve(LCatalogResolutionSettings{
		FfmpegVersion:           ffmpegVersion,
		WindowsShellProfileName: ffmpegBuildSettings.WindowsShellProfileName,
		SelectedLibraryIds:      ffmpegBuildSettings.SelectedLibraryIds,
	})
	if err != nil {
		return LResolvedVersionPlan{}, false, err
	}
	return resolvedPlan, true, nil
}

func LCatalogResolvedBuildPlanFromFFmpegSettings(ffmpegBuildSettings LSettingsFFmpeg, resolvedVersionPlan LResolvedVersionPlan) LResolvedBuildPlan {
	selectedById := map[string]LResolvedLibrary{}
	for _, library := range resolvedVersionPlan.VisibleLibraries {
		selectedById[library.LibraryId] = library
	}
	for _, library := range resolvedVersionPlan.UnsupportedLibraries {
		selectedById[library.LibraryId] = library
	}
	selectedLibraries := []LResolvedLibrary{}
	workIds := []string{}
	for _, libraryId := range resolvedVersionPlan.NormalizedLibraryIds {
		library, exists := selectedById[libraryId]
		if !exists {
			continue
		}
		selectedLibraries = append(selectedLibraries, library)
		workIds = append(workIds, library.WorkIds...)
	}
	versionWorks := LVersionLibraryWorksFromIdsResolve(resolvedVersionPlan.FfmpegVersion, workIds)
	return LResolvedBuildPlan{
		FfmpegVersion:              resolvedVersionPlan.FfmpegVersion,
		RequestedFfmpegVersion:     LVersionArchiveParse(ffmpegBuildSettings.FfmpegSourceArchiveUrl),
		CompatibilityFfmpegVersion: resolvedVersionPlan.FfmpegVersion,
		FfmpegSourceArchiveUrl:     ffmpegBuildSettings.FfmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegBuildSettings.FfmpegSourceSignatureUrl,
		SelectedLibraries:          selectedLibraries,
		VersionLibraryWorks:        versionWorks,
		RequiredMsys2PackageNames:  resolvedVersionPlan.RequiredPackageNames,
		ConfigureFlags:             resolvedVersionPlan.ConfigureFlags,
		Warnings:                   resolvedVersionPlan.Warnings,
	}
}

func LVersionLibraryWorksFromIdsResolve(compatibilityFfmpegVersion string, workIds []string) []LVersionLibraryWork {
	registry, err := LVersionLibraryWorkRegistryLoad()
	if err != nil {
		return nil
	}
	works, _ := registry.LWorksResolve(workIds)
	if compatibilityFfmpegVersion == "" {
		return works
	}
	filteredWorks := []LVersionLibraryWork{}
	for _, work := range works {
		if work.FfmpegVersion == compatibilityFfmpegVersion {
			filteredWorks = append(filteredWorks, work)
		}
	}
	return filteredWorks
}

func LLibraryIdFromVersionLibraryWorkId(ffmpegVersion string, workId string) string {
	prefix := "ffmpeg-" + ffmpegVersion + "-"
	suffix := "-work"
	if len(workId) <= len(prefix)+len(suffix) {
		return ""
	}
	if workId[:len(prefix)] != prefix || workId[len(workId)-len(suffix):] != suffix {
		return ""
	}
	return workId[len(prefix) : len(workId)-len(suffix)]
}

// LPresetLibraryChoice is the frontend-facing library preset record resolved from
// the embedded /presets catalog for one FFmpeg version and one shell profile.
type LPresetLibraryChoice struct {
	PresetId           string   `json:"presetId"`
	LibraryIds         []string `json:"libraryIds"`
	ExtendedLibraryIds []string `json:"extendedLibraryIds,omitempty"`
	Hidden             bool     `json:"hidden,omitempty"`
	Dev                bool     `json:"dev,omitempty"`
}

// LCatalogPresetSourceBuildResolved returns library presets through the embedded
// V5 preset catalog. Known release lines resolve to their catalog line; unknown
// newer release lines resolve through the highest known catalog line.
func LCatalogPresetSourceBuildResolved(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []LPresetLibraryChoice {
	ffmpegVersion := LVersionArchiveParse(ffmpegSourceArchiveUrl)
	if ffmpegVersion == "" {
		return []LPresetLibraryChoice{}
	}
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return []LPresetLibraryChoice{}
	}
	presets, err := resolver.LPresetLibraryChoicesResolve(LCatalogResolutionSettings{
		FfmpegVersion:           ffmpegVersion,
		WindowsShellProfileName: windowsShellProfileName,
	})
	if err != nil {
		return []LPresetLibraryChoice{}
	}
	if presets == nil {
		return []LPresetLibraryChoice{}
	}
	return presets
}

// LPresetLibraryChoicesResolve resolves all presets exposed by the version record.
func (resolver LCatalogResolver) LPresetLibraryChoicesResolve(settings LCatalogResolutionSettings) ([]LPresetLibraryChoice, error) {
	settings = LCatalogResolutionSettingsNormalize(settings)
	catalogFfmpegVersion, exists := resolver.LCatalogVersionForRequestedResolve(settings.FfmpegVersion)
	if !exists {
		return nil, nil
	}
	versionRecord := resolver.VersionRecords[catalogFfmpegVersion]
	presetIds := LCatalogVersionPresetOrderRead(versionRecord)
	choices := make([]LPresetLibraryChoice, 0, len(presetIds))
	for _, presetId := range presetIds {
		presetRecord, exists := resolver.PresetRecords[presetId]
		if !exists {
			continue
		}
		presetVersionRecord, exists := LCatalogPresetVersionRecordRead(presetRecord, catalogFfmpegVersion)
		if !exists {
			continue
		}
		normalLibraryIds, normalOk := resolver.LPresetModeLibraryIdsResolve(settings, presetId, LCatalogDefaultPresetModeName)
		if !normalOk {
			continue
		}
		extendedLibraryIds, extendedOk := resolver.LPresetModeLibraryIdsResolve(settings, presetId, "extended")
		choice := LPresetLibraryChoice{
			PresetId:   presetId,
			LibraryIds: normalLibraryIds,
			Hidden:     LCatalogBoolField(presetVersionRecord, "hiddenInCurrentV4Ui"),
			Dev:        LCatalogBoolField(presetVersionRecord, "devInCurrentV4Ui"),
		}
		if extendedOk {
			choice.ExtendedLibraryIds = extendedLibraryIds
		}
		choices = append(choices, choice)
	}
	return choices, nil
}

func (resolver LCatalogResolver) LPresetModeLibraryIdsResolve(settings LCatalogResolutionSettings, presetId string, modeName string) ([]string, bool) {
	resolvedPlan, err := resolver.LVersionResolve(LCatalogResolutionSettings{
		FfmpegVersion:           settings.FfmpegVersion,
		PresetId:                presetId,
		PresetModeName:          modeName,
		WindowsShellProfileName: settings.WindowsShellProfileName,
	})
	if err != nil || len(resolvedPlan.SelectedLibraryIds) == 0 {
		return nil, false
	}
	return append([]string{}, resolvedPlan.NormalizedLibraryIds...), true
}

func LCatalogVersionPresetOrderRead(versionRecord map[string]any) []string {
	presetsObject, ok := versionRecord["presets"].(map[string]any)
	if !ok {
		return nil
	}
	presetIds := []string{}
	for _, sectionName := range []string{"availableInCurrentV4Ui", "hiddenOrDevInCurrentV4Ui"} {
		items, ok := presetsObject[sectionName].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemObject, ok := item.(map[string]any)
			if !ok {
				continue
			}
			presetId := LCatalogStringField(itemObject, "presetId")
			if presetId != "" {
				presetIds = append(presetIds, presetId)
			}
		}
	}
	return LCatalogStringsUniqueStable(presetIds)
}
