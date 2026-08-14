package program

import "promptfulcustomffmpegbuilder/internal/planning"

type LStateInitial struct {
	HostOS                        string                          `json:"hostOs"`
	KindExplanation               string                          `json:"kindExplanation"`
	SecurityRuleSummary           string                          `json:"securityRuleSummary"`
	NamingRuleSummary             string                          `json:"namingRuleSummary"`
	LBuildSettingsDefault         planning.LSettingsToolchain     `json:"defaultBuildConfigSettings"`
	LFfmpegSettingsDefault        planning.LSettingsFfmpeg        `json:"defaultFfmpegBuildSettings"`
	DefaultLibraryCatalog         []planning.LLibraryChoice       `json:"defaultLibraryCatalog"`
	DefaultLibraryPresetCatalog   []planning.LPresetLibraryChoice `json:"defaultLibraryPresetCatalog"`
	DefaultConfigureOptionCatalog []planning.LOptionChoice        `json:"defaultConfigureOptionCatalog"`
	LReleaseSupportedCatalog      []planning.LReleaseChoice       `json:"supportedFfmpegReleases"`
}

type LResultAction struct {
	RunId     string `json:"runId"`
	StartedAt string `json:"startedAt"`
}

type LFileResult struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	SizeBytes  int64  `json:"sizeBytes"`
	Sha256Hash string `json:"sha256Hash"`
}

type LResultState struct {
	ArtifactsDirectory        string        `json:"artifactsDirectory"`
	ReportPath                string        `json:"reportPath"`
	FfmpegVersion             string        `json:"ffmpegVersion"`
	Files                     []LFileResult `json:"files"`
	SelectedLibraries         []string      `json:"selectedLibraries"`
	SelectedConfigureOptions  []string      `json:"selectedConfigureOptions"`
	RequiredMsys2PackageNames []string      `json:"requiredMsys2PackageNames"`
	ConfigureFlags            []string      `json:"configureFlags"`
	LicenseProfileName        string        `json:"licenseProfileName"`
	CreatedAt                 string        `json:"createdAt"`
}

type LReportArtifact struct {
	CreatedAt                 string                    `json:"createdAt"`
	FfmpegVersion             string                    `json:"ffmpegVersion"`
	FfmpegSourceArchiveUrl    string                    `json:"ffmpegSourceArchiveUrl"`
	SelectedLibraries         []planning.LLibraryChoice `json:"selectedLibraries"`
	SelectedConfigureOptions  []planning.LOptionChoice  `json:"selectedConfigureOptions"`
	RequiredMsys2PackageNames []string                  `json:"requiredMsys2PackageNames"`
	ConfigureFlags            []string                  `json:"configureFlags"`
	LicenseProfileName        string                    `json:"licenseProfileName"`
}
