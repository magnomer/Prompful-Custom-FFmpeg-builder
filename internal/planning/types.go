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
