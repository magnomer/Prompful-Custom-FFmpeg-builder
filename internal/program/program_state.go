package program

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

func (program *LProgram) LStateInitialGet() LStateInitial {
	return LStateInitial{
		HostOS:                        runtime.GOOS,
		KindExplanation:               LLocaleTextGetInternal("initial.kindExplanation", nil),
		SecurityRuleSummary:           LLocaleTextGetInternal("initial.securityRuleSummary", nil),
		NamingRuleSummary:             LLocaleTextGetInternal("initial.namingRuleSummary", nil),
		LBuildSettingsDefault:         planning.LSettingsBuildCreate(),
		LFfmpegSettingsDefault:        planning.LSettingsFfmpegCreate(),
		DefaultLibraryCatalog:         planning.LCatalogLibraryGet(planning.LSettingsFfmpegCreate().FfmpegSourceArchiveUrl, planning.LSettingsFfmpegCreate().WindowsShellProfileName),
		DefaultLibraryPresetCatalog:   planning.LCatalogPresetGet(planning.LSettingsFfmpegCreate().FfmpegSourceArchiveUrl, planning.LSettingsFfmpegCreate().WindowsShellProfileName),
		DefaultConfigureOptionCatalog: planning.LCatalogOptionBuild(),
		LReleaseSupportedCatalog:      planning.LReleaseSupportedGet(),
	}
}

func (program *LProgram) LCatalogSourceGet(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []planning.LLibraryChoice {
	return planning.LCatalogLibraryGet(ffmpegSourceArchiveUrl, windowsShellProfileName)
}

func (program *LProgram) LPresetSourceGet(ffmpegSourceArchiveUrl string, windowsShellProfileName string) []planning.LPresetLibraryChoice {
	return planning.LCatalogPresetGet(ffmpegSourceArchiveUrl, windowsShellProfileName)
}

func (program *LProgram) LResultBuildGet(workspaceDirectory string) (LResultState, error) {
	workspaceLayout := LArtifactLayoutFind(workspaceDirectory)
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return LResultState{}, err
	}
	if err := os.MkdirAll(workspaceLayout.ArtifactsDirectory, 0o755); err != nil {
		return LResultState{}, err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return LResultState{}, err
	}
	result := LResultState{ArtifactsDirectory: workspaceLayout.ArtifactsDirectory, Files: []LFileResult{}, SelectedLibraries: []string{}, SelectedConfigureOptions: []string{}, RequiredMsys2PackageNames: []string{}, ConfigureFlags: []string{}}
	artifactEntries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return LResultState{}, err
	}
	for _, artifactEntry := range artifactEntries {
		if artifactEntry.IsDir() {
			continue
		}
		artifactName := artifactEntry.Name()
		artifactNameLower := strings.ToLower(artifactName)
		if artifactNameLower != "ffmpeg.exe" && artifactNameLower != "ffprobe.exe" && !strings.HasSuffix(artifactNameLower, ".dll") {
			continue
		}
		artifactPath := filepath.Join(workspaceLayout.ArtifactsDirectory, artifactName)
		if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, artifactPath); err != nil {
			return LResultState{}, err
		}
		fileInfo, err := os.Stat(artifactPath)
		if err != nil {
			return LResultState{}, err
		}
		result.Files = append(result.Files, LFileResult{Name: artifactName, Path: artifactPath, SizeBytes: fileInfo.Size(), Sha256Hash: LHashFileCreate(artifactPath)})
	}
	sort.Slice(result.Files, func(leftIndex, rightIndex int) bool {
		return strings.ToLower(result.Files[leftIndex].Name) < strings.ToLower(result.Files[rightIndex].Name)
	})
	reportPath, report, err := LReportLatestRead(workspaceLayout)
	if err != nil {
		result.FfmpegVersion = LArtifactVersionRead(workspaceLayout)
		return result, nil
	}
	result.ReportPath = reportPath
	result.CreatedAt = report.CreatedAt
	result.FfmpegVersion = report.FfmpegVersion
	if result.FfmpegVersion == "" {
		result.FfmpegVersion = planning.LVersionArchiveParse(report.FfmpegSourceArchiveUrl)
	}
	if result.FfmpegVersion == "" {
		result.FfmpegVersion = LArtifactVersionRead(workspaceLayout)
	}
	result.RequiredMsys2PackageNames = report.RequiredMsys2PackageNames
	result.ConfigureFlags = report.ConfigureFlags
	result.LicenseProfileName = report.LicenseProfileName
	for _, library := range report.SelectedLibraries {
		if library.DisplayName == "" {
			continue
		}
		if library.LicenseEffectName != "" && library.LicenseEffectName != "none" {
			result.SelectedLibraries = append(result.SelectedLibraries, "library:"+library.LibraryId+":"+library.LicenseEffectName)
		} else {
			result.SelectedLibraries = append(result.SelectedLibraries, "library:"+library.LibraryId+":"+library.LicenseEffectName)
		}
	}
	for _, option := range report.SelectedConfigureOptions {
		if option.DisplayName != "" {
			result.SelectedConfigureOptions = append(result.SelectedConfigureOptions, "option:"+option.OptionId)
		}
	}
	return result, nil
}

func (program *LProgram) LDirectoryResultOpen(workspaceDirectory string, artifactsDirectory string) error {
	if err := LResultDirectoryValidate(workspaceDirectory, artifactsDirectory); err != nil {
		return err
	}
	return LDirectoryOpen(artifactsDirectory)
}

func (program *LProgram) LReportResultOpen(workspaceDirectory string, reportPath string) error {
	if err := LResultReportValidate(workspaceDirectory, reportPath); err != nil {
		return err
	}
	return LPathOpen(reportPath)
}

func (program *LProgram) LLinkExternalOpen(urlToOpen string) error {
	if err := planning.LExternalWebURLValidate(urlToOpen); err != nil {
		return err
	}
	wailsRuntime.BrowserOpenURL(program.LContext, urlToOpen)
	return nil
}

func LResultDirectoryValidate(workspaceDirectory string, artifactsDirectory string) error {
	if workspaceDirectory == "" || artifactsDirectory == "" {
		return errors.New("workspace and artifact directories are required")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.ArtifactsBaseDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(layout.ArtifactsBaseDirectory, artifactsDirectory); err != nil {
		return errors.New("artifact directory is outside the workspace artifact folder")
	}
	info, err := os.Stat(artifactsDirectory)
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return errors.New("artifact path is not a directory")
	}
	return nil
}

func LResultReportValidate(workspaceDirectory string, reportPath string) error {
	if workspaceDirectory == "" || reportPath == "" {
		return errors.New("workspace directory and report path are required")
	}
	layout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	if err := workspace.LPathRealCheck(layout.WorkspaceDirectory, layout.ArtifactsBaseDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(layout.ArtifactsBaseDirectory, reportPath); err != nil {
		return errors.New("report is outside the workspace artifact folder")
	}
	reportName := filepath.Base(reportPath)
	if !strings.HasPrefix(reportName, "build-report-") || !strings.HasSuffix(reportName, ".json") {
		return errors.New("invalid build report path")
	}
	info, err := os.Stat(reportPath)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return errors.New("build report path is a directory")
	}
	return nil
}

func (program *LProgram) LWorkspaceSelect() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(program.LContext, wailsRuntime.OpenDialogOptions{Title: LLocaleTextForGet(program.lLocaleCurrentGet(), "native.selectWorkspace.title", nil)})
	if err != nil {
		return "", err
	}
	return selection, nil
}
