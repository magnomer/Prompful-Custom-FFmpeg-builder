package planning

// LCatalogDomainName identifies one top-level catalog domain that will become
// embedded data in the next structural phases. Phase 1 only defines the model;
// it does not change the existing hardcoded planner behavior.
type LCatalogDomainName string

const (
	LCatalogDomainLibraries LCatalogDomainName = "libraries"
	LCatalogDomainVersions  LCatalogDomainName = "versions"
	LCatalogDomainPresets   LCatalogDomainName = "presets"
)

// LLibrarySupportState describes the version-resolved state of a library.
// It is intentionally broader than the current Supported/Available pair so the
// later resolver can represent hidden, replaced, source-build, and blocked rows
// without inventing new UI/build meanings.
type LLibrarySupportState string

const (
	LLibrarySupportUnknown     LLibrarySupportState = "unknown"
	LLibrarySupportSupported   LLibrarySupportState = "supported"
	LLibrarySupportUnsupported LLibrarySupportState = "unsupported"
	LLibrarySupportUnavailable LLibrarySupportState = "unavailable"
	LLibrarySupportHidden      LLibrarySupportState = "hidden"
	LLibrarySupportReplaced    LLibrarySupportState = "replaced"
	LSourceBuildRequired       LLibrarySupportState = "source-build-required"
	LPreparationMissing        LLibrarySupportState = "preparation-unimplemented"
	LUIDisabledSupport         LLibrarySupportState = "ui-disabled"
)

// LWorkPhaseName identifies when version-specific library work is allowed
// to run. The term is deliberately "work", not only "preparation": some library
// fixes occur during configure/build/install, pkg-config repair, or before FFmpeg
// configure/link.
type LWorkPhaseName string

const (
	LSourceFetchBefore      LWorkPhaseName = "before-library-source-fetch"
	LSourceExtractAfter     LWorkPhaseName = "after-library-source-extract"
	LLibraryConfigureBefore LWorkPhaseName = "before-library-configure"
	LLibraryConfigurePhase  LWorkPhaseName = "library-configure"
	LLibraryBuildBefore     LWorkPhaseName = "before-library-build"
	LLibraryBuildPhase      LWorkPhaseName = "library-build"
	LLibraryInstallPhase    LWorkPhaseName = "library-install"
	LLibraryInstallAfter    LWorkPhaseName = "after-library-install"
	LFfmpegConfigureBefore  LWorkPhaseName = "before-ffmpeg-configure"
	LFfmpegConfigureAfter   LWorkPhaseName = "after-ffmpeg-configure"
	LFfmpegBuildBefore      LWorkPhaseName = "before-ffmpeg-build"
	LFfmpegBuildAfter       LWorkPhaseName = "after-ffmpeg-build"
)

// LLibraryManifest is the planned data shape for /libraries/{libraryId}.json.
// It stores stable library facts and may refer to version-specific work ids, but
// it does not execute work itself.
type LLibraryManifest struct {
	LibraryId            string            `json:"libraryId"`
	DisplayName          string            `json:"displayName"`
	CategoryName         string            `json:"categoryName"`
	TrackName            LLibraryTrack     `json:"trackName"`
	ConfigureFlags       []string          `json:"configureFlags"`
	PackageNames         []string          `json:"packageNames"`
	LicenseEffectName    string            `json:"licenseEffectName"`
	PlainExplanation     string            `json:"plainExplanation,omitempty"`
	TechnicalExplanation string            `json:"technicalExplanation,omitempty"`
	ConflictLibraryIds   []string          `json:"conflictLibraryIds,omitempty"`
	DefaultWorkId        string            `json:"defaultWorkId,omitempty"`
	VersionWorkIds       map[string]string `json:"versionWorkIds,omitempty"`
}

// LVersionManifest is the planned data shape for /versions/{ffmpegVersion}.json.
// It records FFmpeg-release facts and version-level library/option rules.
type LVersionManifest struct {
	FfmpegVersion                 string                         `json:"ffmpegVersion"`
	ArchiveUrl                    string                         `json:"archiveUrl"`
	SignatureUrl                  string                         `json:"signatureUrl"`
	SupportedConfigureOptionIds   []string                       `json:"supportedConfigureOptionIds,omitempty"`
	UnsupportedConfigureOptionIds []string                       `json:"unsupportedConfigureOptionIds,omitempty"`
	LibraryRules                  map[string]LVersionLibraryRule `json:"libraryRules,omitempty"`
}

// LVersionLibraryRule describes how one FFmpeg release treats one library.
type LVersionLibraryRule struct {
	SupportState LLibrarySupportState `json:"supportState"`
	TrackName    LLibraryTrack        `json:"trackName,omitempty"`
	MinVersion   string               `json:"minVersion,omitempty"`
	WorkId       string               `json:"workId,omitempty"`
	MessageKey   string               `json:"messageKey,omitempty"`
}

// LPresetManifest is the planned data shape for /presets/{presetId}.json.
// Presets describe intent; the resolver will normalize them against the chosen
// FFmpeg version instead of treating them as final build decisions.
type LPresetManifest struct {
	PresetId           string   `json:"presetId"`
	DisplayName        string   `json:"displayName"`
	Description        string   `json:"description,omitempty"`
	WantedLibraryIds   []string `json:"wantedLibraryIds"`
	OptionalLibraryIds []string `json:"optionalLibraryIds,omitempty"`
	ExcludedLibraryIds []string `json:"excludedLibraryIds,omitempty"`
}

// LVersionLibraryWork identifies executable version-specific library work.
// The implementation may live under /versions/x.x.x/*.go in later phases.
type LVersionLibraryWork struct {
	WorkId        string           `json:"workId"`
	FfmpegVersion string           `json:"ffmpegVersion"`
	LibraryId     string           `json:"libraryId"`
	GoFilePath    string           `json:"goFilePath"`
	PhaseNames    []LWorkPhaseName `json:"phaseNames"`
	Summary       string           `json:"summary,omitempty"`
}

// LResolvedLibrary is the future resolver-facing shape for one library after the
// selected FFmpeg version has been applied.
type LResolvedLibrary struct {
	LibraryId            string                     `json:"libraryId"`
	DisplayName          string                     `json:"displayName"`
	CategoryName         string                     `json:"categoryName"`
	TrackName            LLibraryTrack              `json:"trackName"`
	SupportState         LLibrarySupportState       `json:"supportState"`
	ConfigureFlags       []string                   `json:"configureFlags"`
	PackageNames         []string                   `json:"packageNames"`
	OfficialWebpageUrl   string                     `json:"officialWebpageUrl,omitempty"`
	LicenseEffectName    string                     `json:"licenseEffectName,omitempty"`
	PlainExplanation     string                     `json:"plainExplanation,omitempty"`
	TechnicalExplanation string                     `json:"technicalExplanation,omitempty"`
	DefaultChecked       bool                       `json:"defaultChecked"`
	Locked               bool                       `json:"locked"`
	WorkIds              []string                   `json:"workIds,omitempty"`
	PreparationStatus    *LLibraryPreparationStatus `json:"preparationStatus,omitempty"`
	UnavailableReasons   []string                   `json:"unavailableReasons,omitempty"`
	UnavailableProfiles  []string                   `json:"unavailableProfiles,omitempty"`
	Warnings             []LWarningPlan             `json:"warnings,omitempty"`
	VersionCompatibility *LLibraryCompatibility     `json:"versionCompatibility,omitempty"`
}

// LLibraryPreparationStatus reports the catalog's build-preparation state for a
// version-resolved library row. It distinguishes implemented version-specific
// preparation from known catalog rows that are intentionally blocked because no
// preparation implementation exists yet.
type LLibraryPreparationStatus struct {
	Required               bool   `json:"required"`
	Kind                   string `json:"kind,omitempty"`
	Implemented            bool   `json:"implemented"`
	Implementation         string `json:"implementation,omitempty"`
	ImplementationLanguage string `json:"implementationLanguage,omitempty"`
	Reason                 string `json:"reason,omitempty"`
}

// LResolvedVersionPlan is the planned single source of truth for the Libraries
// UI, preset normalization, and build planning after a version is selected.
// FfmpegVersion is kept as the compatibility/catalog version for older callers.
type LResolvedVersionPlan struct {
	FfmpegVersion              string             `json:"ffmpegVersion"`
	RequestedFfmpegVersion     string             `json:"requestedFfmpegVersion,omitempty"`
	CompatibilityFfmpegVersion string             `json:"compatibilityFfmpegVersion,omitempty"`
	VisibleLibraries           []LResolvedLibrary `json:"visibleLibraries"`
	HiddenLibraries            []LResolvedLibrary `json:"hiddenLibraries,omitempty"`
	UnsupportedLibraries       []LResolvedLibrary `json:"unsupportedLibraries,omitempty"`
	SelectedLibraryIds         []string           `json:"selectedLibraryIds"`
	NormalizedLibraryIds       []string           `json:"normalizedLibraryIds"`
	RequiredWorkIds            []string           `json:"requiredWorkIds,omitempty"`
	ConfigureFlags             []string           `json:"configureFlags"`
	RequiredPackageNames       []string           `json:"requiredPackageNames,omitempty"`
	Warnings                   []LWarningPlan     `json:"warnings,omitempty"`
}

// LResolvedBuildPlan is the planned build-runner input. The build runner should
// eventually execute this plan instead of recalculating compatibility.
// FfmpegVersion is kept as the compatibility/catalog version for older callers.
type LResolvedBuildPlan struct {
	FfmpegVersion              string                `json:"ffmpegVersion"`
	RequestedFfmpegVersion     string                `json:"requestedFfmpegVersion,omitempty"`
	CompatibilityFfmpegVersion string                `json:"compatibilityFfmpegVersion,omitempty"`
	FfmpegSourceArchiveUrl     string                `json:"ffmpegSourceArchiveUrl"`
	FfmpegSourceSignatureUrl   string                `json:"ffmpegSourceSignatureUrl"`
	SelectedLibraries          []LResolvedLibrary    `json:"selectedLibraries"`
	VersionLibraryWorks        []LVersionLibraryWork `json:"versionLibraryWorks,omitempty"`
	RequiredMsys2PackageNames  []string              `json:"requiredMsys2PackageNames,omitempty"`
	ConfigureFlags             []string              `json:"configureFlags"`
	Warnings                   []LWarningPlan        `json:"warnings,omitempty"`
}

// LStructuralSummaryCreate returns the Phase 1 design rule as data so tests can
// pin the intended split while later phases fill in the implementation.
func LStructuralSummaryCreate() map[LCatalogDomainName]string {
	return map[LCatalogDomainName]string{
		LCatalogDomainLibraries: "library facts and work references",
		LCatalogDomainVersions:  "FFmpeg release facts and compatibility rules",
		LCatalogDomainPresets:   "preset intent before version normalization",
	}
}
