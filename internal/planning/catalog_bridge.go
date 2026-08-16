package planning

import (
	"fmt"
	"strings"
)

// LCatalogLibraryGet returns the Libraries-tab catalog through the
// embedded version resolver. A known release line resolves to that line's catalog;
// an unknown newer release line uses the highest known catalog line. An empty,
// custom, or unparseable FFmpeg URL still returns no catalog.
func LCatalogLibraryGet(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []LLibraryChoice {
	choices, _ := LCatalogLibraryResolve(ffmpegSourceArchiveUrl, windowsShellProfileName)
	return choices
}

// LCatalogLibraryResolve preserves resolver failures for interactive callers.
// An empty source is a legitimate pre-selection state; a non-empty source that
// cannot resolve is not a successful empty catalog.
func LCatalogLibraryResolve(ffmpegSourceArchiveUrl string, windowsShellProfileName string) ([]LLibraryChoice, error) {
	if strings.TrimSpace(ffmpegSourceArchiveUrl) == "" {
		return []LLibraryChoice{}, nil
	}
	ffmpegVersion := LVersionArchiveParse(ffmpegSourceArchiveUrl)
	if ffmpegVersion == "" {
		return nil, fmt.Errorf("cannot determine FFmpeg version from source URL")
	}
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return nil, fmt.Errorf("load library catalog: %w", err)
	}
	resolvedPlan, err := resolver.LVersionResolve(LCatalogResolutionSettings{
		FfmpegVersion:           ffmpegVersion,
		WindowsShellProfileName: windowsShellProfileName,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve library catalog for FFmpeg %s: %w", ffmpegVersion, err)
	}
	return LLibraryChoicesCreate(resolvedPlan.VisibleLibraries), nil
}

// LLibraryChoicesCreate converts resolved library rows into the legacy
// LLibraryChoice shape consumed by the current frontend. The compatibility field
// carries the resolver state forward without forcing the frontend to understand
// the new catalog model yet.
func LLibraryChoicesCreate(resolvedLibraries []LResolvedLibrary) []LLibraryChoice {
	choices := make([]LLibraryChoice, 0, len(resolvedLibraries))
	for _, resolvedLibrary := range resolvedLibraries {
		choices = append(choices, LLibraryChoiceCreate(resolvedLibrary))
	}
	return choices
}

func LLibraryChoiceCreate(resolvedLibrary LResolvedLibrary) LLibraryChoice {
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

// LLibraryFlagMatch returns library entries from an already
// resolved catalog whose configure flags overlap with the supplied extra flags.
// This is the version-aware equivalent of LConfigureFlagMatch; it keeps
// extra raw flags on the same catalog authority as checkbox selections.
func LLibraryFlagMatch(catalog []LLibraryChoice, flags []string, skip []LLibraryChoice) []LLibraryChoice {
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

// LCatalogPlanResolve creates the resolver plan associated
// with a current FFmpeg settings object. Known release lines resolve to their
// catalog line; unknown newer release lines resolve through the highest known
// catalog line. Unparseable or older unsupported versions are reported as missing.
func LCatalogPlanResolve(ffmpegBuildSettings LSettingsFfmpeg) (LResolvedVersionPlan, bool, error) {
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

func LCatalogBuildResolve(ffmpegBuildSettings LSettingsFfmpeg, resolvedVersionPlan LResolvedVersionPlan) LResolvedBuildPlan {
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
	versionWorks := LLibraryWorkResolve(resolvedVersionPlan.FfmpegVersion, workIds)
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

func LLibraryWorkResolve(compatibilityFfmpegVersion string, workIds []string) []LVersionLibraryWork {
	registry, err := LWorkRegistryLoad()
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

func LLibraryIdGet(ffmpegVersion string, workId string) string {
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

// LCatalogPresetGet returns library presets through the embedded
// V5 preset catalog. Known release lines resolve to their catalog line; unknown
// newer release lines resolve through the highest known catalog line.
func LCatalogPresetGet(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []LPresetLibraryChoice {
	presets, _ := LCatalogPresetResolve(ffmpegSourceArchiveUrl, windowsShellProfileName)
	return presets
}

func LCatalogPresetResolve(ffmpegSourceArchiveUrl string, windowsShellProfileName string) ([]LPresetLibraryChoice, error) {
	if strings.TrimSpace(ffmpegSourceArchiveUrl) == "" {
		return []LPresetLibraryChoice{}, nil
	}
	ffmpegVersion := LVersionArchiveParse(ffmpegSourceArchiveUrl)
	if ffmpegVersion == "" {
		return nil, fmt.Errorf("cannot determine FFmpeg version from source URL")
	}
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return nil, fmt.Errorf("load preset catalog: %w", err)
	}
	presets, err := resolver.LPresetChoicesResolve(LCatalogResolutionSettings{
		FfmpegVersion:           ffmpegVersion,
		WindowsShellProfileName: windowsShellProfileName,
	})
	if err != nil {
		return nil, fmt.Errorf("resolve preset catalog for FFmpeg %s: %w", ffmpegVersion, err)
	}
	if presets == nil {
		return nil, fmt.Errorf("no preset catalog exists for FFmpeg %s", ffmpegVersion)
	}
	return presets, nil
}

// LPresetChoicesResolve resolves all presets exposed by the version record.
func (resolver LCatalogResolver) LPresetChoicesResolve(settings LCatalogResolutionSettings) ([]LPresetLibraryChoice, error) {
	settings = LCatalogSettingsNormalize(settings)
	catalogFfmpegVersion, exists := resolver.LCatalogVersionResolve(settings.FfmpegVersion)
	if !exists {
		return nil, nil
	}
	versionRecord := resolver.VersionRecords[catalogFfmpegVersion]
	presetIds := LPresetOrderRead(versionRecord)
	choices := make([]LPresetLibraryChoice, 0, len(presetIds))
	for _, presetId := range presetIds {
		presetRecord, exists := resolver.PresetRecords[presetId]
		if !exists {
			continue
		}
		presetVersionRecord, exists := LPresetRecordRead(presetRecord, catalogFfmpegVersion)
		if !exists {
			continue
		}
		normalLibraryIds, normalOk := resolver.LPresetModeResolve(settings, presetId, LPresetModeName)
		if !normalOk {
			continue
		}
		extendedLibraryIds, extendedOk := resolver.LPresetModeResolve(settings, presetId, "extended")
		choice := LPresetLibraryChoice{
			PresetId:   presetId,
			LibraryIds: normalLibraryIds,
			Hidden:     LCatalogBooleanGet(presetVersionRecord, "hiddenInCurrentCatalogUi"),
			Dev:        LCatalogBooleanGet(presetVersionRecord, "devInCurrentCatalogUi"),
		}
		if extendedOk {
			choice.ExtendedLibraryIds = extendedLibraryIds
		}
		choices = append(choices, choice)
	}
	return choices, nil
}

func (resolver LCatalogResolver) LPresetModeResolve(settings LCatalogResolutionSettings, presetId string, modeName string) ([]string, bool) {
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

func LPresetOrderRead(versionRecord map[string]any) []string {
	presetsObject, ok := versionRecord["presets"].(map[string]any)
	if !ok {
		return nil
	}
	presetIds := []string{}
	for _, sectionName := range []string{"availableInCurrentCatalogUi", "hiddenOrDevInCurrentCatalogUi"} {
		items, ok := presetsObject[sectionName].([]any)
		if !ok {
			continue
		}
		for _, item := range items {
			itemObject, ok := item.(map[string]any)
			if !ok {
				continue
			}
			presetId := LCatalogFieldGet(itemObject, "presetId")
			if presetId != "" {
				presetIds = append(presetIds, presetId)
			}
		}
	}
	return LStringsUniqueGet(presetIds)
}
