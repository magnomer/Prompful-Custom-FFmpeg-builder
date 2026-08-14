package planning

type LRiskLevel string

type LLibraryTrack string

const (
	LLibraryTrackNative   LLibraryTrack = "native"
	LLibraryTrackInternal LLibraryTrack = "internal"
	LLibraryTrackExternal LLibraryTrack = "external"
)

const (
	LRiskInfo    LRiskLevel = "info"
	LRiskWarning LRiskLevel = "warning"
	LRiskBlocked LRiskLevel = "blocked"
)

type LWarningPlan struct {
	LRiskLevel    LRiskLevel        `json:"riskLevelName"`
	Message       string            `json:"message"`
	MessageKey    string            `json:"messageKey,omitempty"`
	MessageValues map[string]string `json:"messageValues,omitempty"`
}

type LOperationPlan struct {
	OperationName string            `json:"operationName"`
	Summary       string            `json:"summary"`
	SummaryKey    string            `json:"summaryKey,omitempty"`
	SummaryValues map[string]string `json:"summaryValues,omitempty"`
}

type LLibraryChoice struct {
	LibraryId            string                     `json:"libraryId"`
	TrackName            LLibraryTrack              `json:"trackName"`
	DisplayName          string                     `json:"displayName"`
	CategoryName         string                     `json:"categoryName"`
	ConfigureFlags       []string                   `json:"configureFlags"`
	PackageNames         []string                   `json:"packageNames"`
	OfficialWebpageUrl   string                     `json:"officialWebpageUrl"`
	LicenseEffectName    string                     `json:"licenseEffectName"`
	PlainExplanation     string                     `json:"plainExplanation"`
	TechnicalExplanation string                     `json:"technicalExplanation"`
	DefaultChecked       bool                       `json:"defaultChecked"`
	Locked               bool                       `json:"locked"`
	SupportState         LLibrarySupportState       `json:"supportState,omitempty"`
	PreparationStatus    *LLibraryPreparationStatus `json:"preparationStatus,omitempty"`
	UnavailableReasons   []string                   `json:"unavailableReasons,omitempty"`
	UnavailableProfiles  []string                   `json:"unavailableProfiles,omitempty"`
	// VersionCompatibility, when set, reports whether this library is supported by the
	// FFmpeg release being built, per that release's support release-support manifest. It is populated only
	// on a plan's resolved SelectedLibraries (where the FFmpeg version is known) and only
	// when the chosen release line is release-supported; the version-free library catalog-fetch path and
	// unsupported release lines leave it nil. omitempty keeps it additive: the existing UI
	// ignores it.
	VersionCompatibility *LLibraryCompatibility `json:"versionCompatibility,omitempty"`
}

// LLibraryCompatibility describes how a selected library relates to the chosen FFmpeg
// release's support release-support manifest. Supported is false when the release is release-supported but does not
// support the library (its --enable switch does not exist in that release). Available is false
// when the switch exists but the package this builder can supply cannot satisfy it for this
// release (release-support manifest Unavailable, e.g. lensfun); Available is therefore Supported AND not
// package-unavailable. MinVersion echoes FFmpeg's pkg-config minimum for that library in that
// release ("" when none). The UI uses Available to show/hide rows; the backend blocks the plan
// when a selected library is Supported-but-unavailable or unsupported.
type LLibraryCompatibility struct {
	Supported  bool   `json:"supported"`
	Available  bool   `json:"available"`
	MinVersion string `json:"minVersion,omitempty"`
}

type LLibraryTrackSelection struct {
	TrackName LLibraryTrack    `json:"trackName"`
	Libraries []LLibraryChoice `json:"libraries"`
}

type LOptionChoice struct {
	OptionId         string   `json:"optionId"`
	DisplayName      string   `json:"displayName"`
	CategoryName     string   `json:"categoryName"`
	ConfigureFlags   []string `json:"configureFlags"`
	PlainExplanation string   `json:"plainExplanation"`
	TechnicalNote    string   `json:"technicalNote"`
	DefaultEnabled   bool     `json:"defaultEnabled"`
	Locked           bool     `json:"locked"`
	RiskLevelName    string   `json:"riskLevelName"`
}

type LSettingsToolchain struct {
	WorkspaceDirectory       string   `json:"workspaceDirectory"`
	Msys2ArchiveUrl          string   `json:"msys2ArchiveUrl"`
	Msys2ArchiveSha256Hash   string   `json:"msys2ArchiveSha256Hash"`
	Msys2ArchiveSignatureUrl string   `json:"msys2ArchiveSignatureUrl"`
	Msys2PackageNames        []string `json:"msys2PackageNames"`
	WindowsShellProfileName  string   `json:"windowsShellProfileName"`
}

type LSettingsFfmpeg struct {
	WorkspaceDirectory         string   `json:"workspaceDirectory"`
	FfmpegSourceArchiveUrl     string   `json:"ffmpegSourceArchiveUrl"`
	FfmpegSourceSignatureUrl   string   `json:"ffmpegSourceSignatureUrl"`
	FfmpegSourceSha256Hash     string   `json:"ffmpegSourceSha256Hash"`
	SelectedLibraryIds         []string `json:"selectedLibraryIds"`
	SelectedConfigureOptionIds []string `json:"selectedConfigureOptionIds"`
	ExtraConfigureFlags        []string `json:"extraConfigureFlags"`
	ConfigureFlags             []string `json:"configureFlags"`
	ParallelJobCount           int      `json:"parallelJobCount"`
	WindowsShellProfileName    string   `json:"windowsShellProfileName"`
	LicenseProfileName         string   `json:"licenseProfileName"`
}

type LPlanToolchain struct {
	ActionName                 string           `json:"actionName"`
	PlanHash                   string           `json:"planHash"`
	WorkspaceDirectory         string           `json:"workspaceDirectory"`
	Msys2RootDirectory         string           `json:"msys2RootDirectory"`
	Msys2ArchiveUrl            string           `json:"msys2ArchiveUrl"`
	Msys2ArchiveSha256Hash     string           `json:"msys2ArchiveSha256Hash"`
	Msys2ArchiveSignatureUrl   string           `json:"msys2ArchiveSignatureUrl"`
	Msys2PackageNames          []string         `json:"msys2PackageNames"`
	WindowsShellProfileName    string           `json:"windowsShellProfileName"`
	WillModifySystemPath       bool             `json:"willModifySystemPath"`
	WillRequireAdminRights     bool             `json:"willRequireAdminRights"`
	WillUseExistingMsys2       bool             `json:"willUseExistingMsys2"`
	WillDeleteFiles            bool             `json:"willDeleteFiles"`
	DownloadConflictPolicyName string           `json:"downloadConflictPolicyName"`
	LPolicyExtraction          string           `json:"extractionDestinationPolicyName"`
	Operations                 []LOperationPlan `json:"operations"`
	Warnings                   []LWarningPlan   `json:"warnings"`
	IsExecutable               bool             `json:"isExecutable"`
}

type LPlanFfmpeg struct {
	ActionName                 string                   `json:"actionName"`
	PlanHash                   string                   `json:"planHash"`
	WorkspaceDirectory         string                   `json:"workspaceDirectory"`
	Msys2RootDirectory         string                   `json:"msys2RootDirectory"`
	FfmpegSourceArchiveUrl     string                   `json:"ffmpegSourceArchiveUrl"`
	FfmpegSourceSignatureUrl   string                   `json:"ffmpegSourceSignatureUrl"`
	FfmpegSourceSha256Hash     string                   `json:"ffmpegSourceSha256Hash"`
	RequestedFfmpegVersion     string                   `json:"requestedFfmpegVersion,omitempty"`
	CompatibilityFfmpegVersion string                   `json:"compatibilityFfmpegVersion,omitempty"`
	SelectedLibraryIds         []string                 `json:"selectedLibraryIds"`
	SelectedLibraries          []LLibraryChoice         `json:"selectedLibraries"`
	SelectedNativeLibraries    []LLibraryChoice         `json:"selectedNativeLibraries"`
	SelectedInternalLibraries  []LLibraryChoice         `json:"selectedInternalLibraries"`
	SelectedExternalLibraries  []LLibraryChoice         `json:"selectedExternalLibraries"`
	SelectedLibrariesByTrack   []LLibraryTrackSelection `json:"selectedLibrariesByTrack"`
	LPreparationCatalog        []LLibraryPreparation    `json:"libraryPreparations"`
	RequiredMsys2PackageNames  []string                 `json:"requiredMsys2PackageNames"`
	GeneratedConfigureFlags    []string                 `json:"generatedConfigureFlags"`
	SelectedConfigureOptions   []LOptionChoice          `json:"selectedConfigureOptions"`
	GeneratedOptionFlags       []string                 `json:"generatedOptionFlags"`
	ExtraConfigureFlags        []string                 `json:"extraConfigureFlags"`
	ConfigureFlags             []string                 `json:"configureFlags"`
	ParallelJobCount           int                      `json:"parallelJobCount"`
	WindowsShellProfileName    string                   `json:"windowsShellProfileName"`
	LicenseProfileName         string                   `json:"licenseProfileName"`
	WillModifySystemPath       bool                     `json:"willModifySystemPath"`
	WillRequireAdminRights     bool                     `json:"willRequireAdminRights"`
	WillUseExistingMsys2       bool                     `json:"willUseExistingMsys2"`
	WillDeleteFiles            bool                     `json:"willDeleteFiles"`
	DownloadConflictPolicyName string                   `json:"downloadConflictPolicyName"`
	LPolicyExtraction          string                   `json:"extractionDestinationPolicyName"`
	Operations                 []LOperationPlan         `json:"operations"`
	Warnings                   []LWarningPlan           `json:"warnings"`
	ResolvedVersionPlan        *LResolvedVersionPlan    `json:"resolvedVersionPlan,omitempty"`
	ResolvedBuildPlan          *LResolvedBuildPlan      `json:"resolvedBuildPlan,omitempty"`
	IsExecutable               bool                     `json:"isExecutable"`
}

type LReviewToolchain struct {
	ReviewSessionId          string         `json:"reviewSessionId"`
	ExpectedLConsentText     string         `json:"expectedLConsentText"`
	ExpectedLConsentTextHash string         `json:"expectedLConsentTextHash"`
	ExpiresAtUnixTime        int64          `json:"expiresAtUnixTime"`
	Plan                     LPlanToolchain `json:"plan"`
}

type LReviewFfmpeg struct {
	ReviewSessionId          string      `json:"reviewSessionId"`
	ExpectedLConsentText     string      `json:"expectedLConsentText"`
	ExpectedLConsentTextHash string      `json:"expectedLConsentTextHash"`
	ExpiresAtUnixTime        int64       `json:"expiresAtUnixTime"`
	Plan                     LPlanFfmpeg `json:"plan"`
}
