package planning

type RiskLevel string

type LibraryTrackName string

const (
	LibraryTrackNative   LibraryTrackName = "native"
	LibraryTrackInternal LibraryTrackName = "internal"
	LibraryTrackExternal LibraryTrackName = "external"
)

const (
	RiskLevelInfo    RiskLevel = "info"
	RiskLevelWarning RiskLevel = "warning"
	RiskLevelBlocked RiskLevel = "blocked"
)

type PlanWarning struct {
	RiskLevel     RiskLevel         `json:"riskLevelName"`
	Message       string            `json:"message"`
	MessageKey    string            `json:"messageKey,omitempty"`
	MessageValues map[string]string `json:"messageValues,omitempty"`
}

type PlanOperation struct {
	OperationName string            `json:"operationName"`
	Summary       string            `json:"summary"`
	SummaryKey    string            `json:"summaryKey,omitempty"`
	SummaryValues map[string]string `json:"summaryValues,omitempty"`
}

type LibraryChoice struct {
	LibraryId            string           `json:"libraryId"`
	TrackName            LibraryTrackName `json:"trackName"`
	DisplayName          string           `json:"displayName"`
	CategoryName         string           `json:"categoryName"`
	ConfigureFlags       []string         `json:"configureFlags"`
	PackageNames         []string         `json:"packageNames"`
	OfficialWebpageUrl   string           `json:"officialWebpageUrl"`
	LicenseEffectName    string           `json:"licenseEffectName"`
	PlainExplanation     string           `json:"plainExplanation"`
	TechnicalExplanation string           `json:"technicalExplanation"`
	DefaultChecked       bool             `json:"defaultChecked"`
	Locked               bool             `json:"locked"`
	// VersionCompatibility, when set, reports whether this library is supported by the
	// FFmpeg release being built, per that release's support manifest. It is populated only
	// on a plan's resolved SelectedLibraries (where the FFmpeg version is known) and only
	// when the chosen release line is manifested; the version-free catalog-fetch path and
	// un-manifested release lines leave it nil. omitempty keeps it additive: the existing UI
	// ignores it.
	VersionCompatibility *LibraryVersionCompatibility `json:"versionCompatibility,omitempty"`
}

// LibraryVersionCompatibility describes how a selected library relates to the chosen FFmpeg
// release's support manifest. Supported is false when the release is manifested but does not
// support the library (its --enable switch does not exist in that release). Available is false
// when the switch exists but the package this builder can supply cannot satisfy it for this
// release (manifest Unavailable, e.g. lensfun); Available is therefore Supported AND not
// package-unavailable. MinVersion echoes FFmpeg's pkg-config minimum for that library in that
// release ("" when none). The UI uses Available to show/hide rows; the backend blocks the plan
// when a selected library is Supported-but-unavailable or unsupported.
type LibraryVersionCompatibility struct {
	Supported  bool   `json:"supported"`
	Available  bool   `json:"available"`
	MinVersion string `json:"minVersion,omitempty"`
}

type TrackedLibrarySelection struct {
	TrackName LibraryTrackName `json:"trackName"`
	Libraries []LibraryChoice  `json:"libraries"`
}

type ConfigureOptionChoice struct {
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

type BuildConfigSettings struct {
	WorkspaceDirectory       string   `json:"workspaceDirectory"`
	Msys2ArchiveUrl          string   `json:"msys2ArchiveUrl"`
	Msys2ArchiveSha256Hash   string   `json:"msys2ArchiveSha256Hash"`
	Msys2ArchiveSignatureUrl string   `json:"msys2ArchiveSignatureUrl"`
	Msys2PackageNames        []string `json:"msys2PackageNames"`
	WindowsShellProfileName  string   `json:"windowsShellProfileName"`
}

type FfmpegBuildSettings struct {
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

type ToolchainPreparationPlan struct {
	ActionName                 string          `json:"actionName"`
	PlanHash                   string          `json:"planHash"`
	WorkspaceDirectory         string          `json:"workspaceDirectory"`
	Msys2RootDirectory         string          `json:"msys2RootDirectory"`
	Msys2ArchiveUrl            string          `json:"msys2ArchiveUrl"`
	Msys2ArchiveSha256Hash     string          `json:"msys2ArchiveSha256Hash"`
	Msys2ArchiveSignatureUrl   string          `json:"msys2ArchiveSignatureUrl"`
	Msys2PackageNames          []string        `json:"msys2PackageNames"`
	WindowsShellProfileName    string          `json:"windowsShellProfileName"`
	WillModifySystemPath       bool            `json:"willModifySystemPath"`
	WillRequireAdminRights     bool            `json:"willRequireAdminRights"`
	WillUseExistingMsys2       bool            `json:"willUseExistingMsys2"`
	WillDeleteFiles            bool            `json:"willDeleteFiles"`
	DownloadConflictPolicyName string          `json:"downloadConflictPolicyName"`
	ExtractDestinationPolicy   string          `json:"extractionDestinationPolicyName"`
	Operations                 []PlanOperation `json:"operations"`
	Warnings                   []PlanWarning   `json:"warnings"`
	IsExecutable               bool            `json:"isExecutable"`
}

type FfmpegBuildPlan struct {
	ActionName                 string                    `json:"actionName"`
	PlanHash                   string                    `json:"planHash"`
	WorkspaceDirectory         string                    `json:"workspaceDirectory"`
	Msys2RootDirectory         string                    `json:"msys2RootDirectory"`
	FfmpegSourceArchiveUrl     string                    `json:"ffmpegSourceArchiveUrl"`
	FfmpegSourceSignatureUrl   string                    `json:"ffmpegSourceSignatureUrl"`
	FfmpegSourceSha256Hash     string                    `json:"ffmpegSourceSha256Hash"`
	SelectedLibraryIds         []string                  `json:"selectedLibraryIds"`
	SelectedLibraries          []LibraryChoice           `json:"selectedLibraries"`
	SelectedNativeLibraries    []LibraryChoice           `json:"selectedNativeLibraries"`
	SelectedInternalLibraries  []LibraryChoice           `json:"selectedInternalLibraries"`
	SelectedExternalLibraries  []LibraryChoice           `json:"selectedExternalLibraries"`
	SelectedLibrariesByTrack   []TrackedLibrarySelection `json:"selectedLibrariesByTrack"`
	LibraryPreparations        []LibraryPreparation      `json:"libraryPreparations"`
	RequiredMsys2PackageNames  []string                  `json:"requiredMsys2PackageNames"`
	GeneratedConfigureFlags    []string                  `json:"generatedConfigureFlags"`
	SelectedConfigureOptions   []ConfigureOptionChoice   `json:"selectedConfigureOptions"`
	GeneratedOptionFlags       []string                  `json:"generatedOptionFlags"`
	ExtraConfigureFlags        []string                  `json:"extraConfigureFlags"`
	ConfigureFlags             []string                  `json:"configureFlags"`
	ParallelJobCount           int                       `json:"parallelJobCount"`
	WindowsShellProfileName    string                    `json:"windowsShellProfileName"`
	LicenseProfileName         string                    `json:"licenseProfileName"`
	WillModifySystemPath       bool                      `json:"willModifySystemPath"`
	WillRequireAdminRights     bool                      `json:"willRequireAdminRights"`
	WillUseExistingMsys2       bool                      `json:"willUseExistingMsys2"`
	WillDeleteFiles            bool                      `json:"willDeleteFiles"`
	DownloadConflictPolicyName string                    `json:"downloadConflictPolicyName"`
	ExtractDestinationPolicy   string                    `json:"extractionDestinationPolicyName"`
	Operations                 []PlanOperation           `json:"operations"`
	Warnings                   []PlanWarning             `json:"warnings"`
	IsExecutable               bool                      `json:"isExecutable"`
}

type ToolchainPreparationPlanReview struct {
	ReviewSessionId         string                   `json:"reviewSessionId"`
	ExpectedConsentText     string                   `json:"expectedConsentText"`
	ExpectedConsentTextHash string                   `json:"expectedConsentTextHash"`
	ExpiresAtUnixTime       int64                    `json:"expiresAtUnixTime"`
	Plan                    ToolchainPreparationPlan `json:"plan"`
}

type FfmpegBuildPlanReview struct {
	ReviewSessionId         string          `json:"reviewSessionId"`
	ExpectedConsentText     string          `json:"expectedConsentText"`
	ExpectedConsentTextHash string          `json:"expectedConsentTextHash"`
	ExpiresAtUnixTime       int64           `json:"expiresAtUnixTime"`
	Plan                    FfmpegBuildPlan `json:"plan"`
}
