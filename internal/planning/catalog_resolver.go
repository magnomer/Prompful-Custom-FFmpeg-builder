package planning

import (
	"fmt"
	"strings"
)

const (
	LPresetModeName      = "normal"
	LShellProfileDefault = "ucrt64"
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
	catalog, report, err := LCatalogLoadValidate()
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
		record, err := LJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		libraryId := LCatalogFieldGet(record, "libraryId")
		resolver.LibraryRecords[libraryId] = record
		resolver.LibraryFilesById[libraryId] = file
	}
	for _, file := range catalog.VersionFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		versionId := LVersionIdentifierRead(record)
		resolver.VersionRecords[versionId] = record
		resolver.VersionFilesById[versionId] = file
	}
	for _, file := range catalog.PresetFiles {
		record, err := LJsonObjectRead(file)
		if err != nil {
			return LCatalogResolver{}, err
		}
		presetId := LCatalogFieldGet(record, "presetId")
		resolver.PresetRecords[presetId] = record
		resolver.PresetFilesById[presetId] = file
	}
	return resolver, nil
}

func LCatalogEmbeddedResolve(settings LCatalogResolutionSettings) (LResolvedVersionPlan, error) {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return LResolvedVersionPlan{}, err
	}
	return resolver.LVersionResolve(settings)
}

func (resolver LCatalogResolver) LVersionResolve(settings LCatalogResolutionSettings) (LResolvedVersionPlan, error) {
	settings = LCatalogSettingsNormalize(settings)
	catalogFfmpegVersion, exists := resolver.LCatalogVersionResolve(settings.FfmpegVersion)
	if !exists {
		return LResolvedVersionPlan{}, fmt.Errorf("unknown FFmpeg version %q", settings.FfmpegVersion)
	}
	versionRecord := resolver.VersionRecords[catalogFfmpegVersion]
	libraryOrder := LLibraryOrderRead(versionRecord)
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
	selectedLibraryIds, _ := resolver.LSelectionResolve(selectedSettings)
	preserveInvalidSelectedLibraries := len(settings.SelectedLibraryIds) > 0 && settings.PresetId == ""
	normalizedLibraryIds, warnings := LLibrarySelectNormalize(catalogFfmpegVersion, selectedLibraryIds, resolvedById, preserveInvalidSelectedLibraries)
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
	configureFlags = LStringsSortedGet(configureFlags)
	requiredPackageNames = LStringsSortedGet(requiredPackageNames)
	requiredWorkIds = LStringsSortedGet(requiredWorkIds)
	return LResolvedVersionPlan{
		FfmpegVersion:              catalogFfmpegVersion,
		RequestedFfmpegVersion:     settings.FfmpegVersion,
		CompatibilityFfmpegVersion: catalogFfmpegVersion,
		VisibleLibraries:           visibleLibraries,
		UnsupportedLibraries:       unsupportedLibraries,
		SelectedLibraryIds:         LStringsSortedGet(selectedLibraryIds),
		NormalizedLibraryIds:       normalizedLibraryIds,
		RequiredWorkIds:            requiredWorkIds,
		ConfigureFlags:             configureFlags,
		RequiredPackageNames:       requiredPackageNames,
		Warnings:                   warnings,
	}, nil
}

func LCatalogSettingsNormalize(settings LCatalogResolutionSettings) LCatalogResolutionSettings {
	settings.FfmpegVersion = strings.TrimSpace(settings.FfmpegVersion)
	settings.PresetId = strings.TrimSpace(settings.PresetId)
	settings.PresetModeName = strings.TrimSpace(settings.PresetModeName)
	settings.WindowsShellProfileName = strings.TrimSpace(settings.WindowsShellProfileName)
	if settings.PresetModeName == "" {
		settings.PresetModeName = LPresetModeName
	}
	if settings.WindowsShellProfileName == "" {
		settings.WindowsShellProfileName = LShellProfileDefault
	}
	settings.SelectedLibraryIds = LStringsSortedGet(settings.SelectedLibraryIds)
	return settings
}

func (resolver LCatalogResolver) LBuildPlanCreate(settings LCatalogResolutionSettings) (LResolvedBuildPlan, error) {
	settings = LCatalogSettingsNormalize(settings)
	resolvedVersionPlan, err := resolver.LVersionResolve(settings)
	if err != nil {
		return LResolvedBuildPlan{}, err
	}
	versionRecord := resolver.VersionRecords[resolvedVersionPlan.FfmpegVersion]
	ffmpegObject, _ := versionRecord["ffmpeg"].(map[string]any)
	ffmpegSourceArchiveUrl := LArchiveUrlResolve(settings.FfmpegVersion, resolvedVersionPlan.FfmpegVersion, ffmpegObject)
	ffmpegSourceSignatureUrl := LSignatureUrlResolve(settings.FfmpegVersion, resolvedVersionPlan.FfmpegVersion, ffmpegObject, ffmpegSourceArchiveUrl)
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
