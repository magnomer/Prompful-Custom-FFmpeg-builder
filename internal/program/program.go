package program

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"promptfulcustomffmpegbuilder/internal/audit"
	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/download"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/reporting"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
	"promptfulcustomffmpegbuilder/internal/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type LProgram struct {
	LContext                context.Context
	LContextAction          context.Context
	LActionCancelFunction   context.CancelFunc
	LMutexAction            sync.Mutex
	LMutexReviewSession     sync.Mutex
	LToolchainReviewStorage map[string]LReviewToolchainStored
	LFFmpegReviewStorage    map[string]LReviewFFmpegStored
	LStateWindowStartup     LStateWindow
	LLocaleUi               string
	LMutexLocaleUi          sync.RWMutex
	LReporter               reporting.LReporter
	LConfirmer              reporting.LConfirmer
}

type LReviewToolchainStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanToolchain
}

type LReviewFFmpegStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanFFmpeg
}

type LStateInitial struct {
	HostOS                        string                          `json:"hostOs"`
	KindExplanation               string                          `json:"kindExplanation"`
	SecurityRuleSummary           string                          `json:"securityRuleSummary"`
	NamingRuleSummary             string                          `json:"namingRuleSummary"`
	LBuildSettingsDefault         planning.LSettingsToolchain     `json:"defaultBuildConfigSettings"`
	LFFmpegSettingsDefault        planning.LSettingsFFmpeg        `json:"defaultFfmpegBuildSettings"`
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

func LProgramCreate() *LProgram {
	return &LProgram{LStateWindowStartup: LStateWindowLoad(), LToolchainReviewStorage: map[string]LReviewToolchainStored{}, LFFmpegReviewStorage: map[string]LReviewFFmpegStored{}}
}

func LWindowInitialRead(program *LProgram) (int, int) {
	return program.LStateWindowStartup.Width, program.LStateWindowStartup.Height
}

func LLocaleTextGet(key string, values map[string]string) string {
	return LLocaleTextGetInternal(key, values)
}

func (program *LProgram) LProgramStart(LContext context.Context) {
	program.LContext = LContext
	program.LReporter = LReporterWails{program: program}
	program.LConfirmer = LConfirmerWails{program: program}
	program.LWindowGeometryRestore()
}

// LWindowCloseCheck runs while the window still exists, so it is the safe place to
// read window geometry. Returning false allows the close to proceed.
func (program *LProgram) LWindowCloseCheck(LContext context.Context) bool {
	program.LWindowGeometrySave(LContext)
	return false
}

func (program *LProgram) LProgramStop(LContext context.Context) {
	program.LActionApprovedCancel()
}

// LLocaleSet records the UI language so the backend-rendered native confirmation
// dialog can follow it. Only the dialog is localized; log and error messages stay
// English. Unknown locales fall back to English.
func (program *LProgram) LLocaleSet(locale string) {
	normalizedLocale := "en"
	if locale == "ko" {
		normalizedLocale = "ko"
	}
	program.LMutexLocaleUi.Lock()
	program.LLocaleUi = normalizedLocale
	program.LMutexLocaleUi.Unlock()
}

func (program *LProgram) lLocaleCurrentGet() string {
	program.LMutexLocaleUi.RLock()
	defer program.LMutexLocaleUi.RUnlock()
	if program.LLocaleUi == "" {
		return "en"
	}
	return program.LLocaleUi
}

func (program *LProgram) LStateInitialGet() LStateInitial {
	return LStateInitial{
		HostOS:                        runtime.GOOS,
		KindExplanation:               LLocaleTextGetInternal("initial.kindExplanation", nil),
		SecurityRuleSummary:           LLocaleTextGetInternal("initial.securityRuleSummary", nil),
		NamingRuleSummary:             LLocaleTextGetInternal("initial.namingRuleSummary", nil),
		LBuildSettingsDefault:         planning.LSettingsBuildCreate(),
		LFFmpegSettingsDefault:        planning.LSettingsFFmpegCreate(),
		DefaultLibraryCatalog:         planning.LCatalogLibraryGet(planning.LSettingsFFmpegCreate().FfmpegSourceArchiveUrl, planning.LSettingsFFmpegCreate().WindowsShellProfileName),
		DefaultLibraryPresetCatalog:   planning.LCatalogPresetGet(planning.LSettingsFFmpegCreate().FfmpegSourceArchiveUrl, planning.LSettingsFFmpegCreate().WindowsShellProfileName),
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

func (program *LProgram) LDirectoryResultOpen(workspaceDirectory string) error {
	workspaceLayout := LArtifactLayoutFind(workspaceDirectory)
	if err := os.MkdirAll(workspaceLayout.ArtifactsDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		return exec.Command("explorer.exe", workspaceLayout.ArtifactsDirectory).Start()
	case "darwin":
		return exec.Command("open", workspaceLayout.ArtifactsDirectory).Start()
	default:
		return exec.Command("xdg-open", workspaceLayout.ArtifactsDirectory).Start()
	}
}

func (program *LProgram) LReportResultOpen(workspaceDirectory string) error {
	workspaceLayout := LArtifactLayoutFind(workspaceDirectory)
	reportPath, _, err := LReportLatestRead(workspaceLayout)
	if err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, reportPath); err != nil {
		return err
	}
	switch runtime.GOOS {
	case "windows":
		return exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", reportPath).Start()
	case "darwin":
		return exec.Command("open", reportPath).Start()
	default:
		return exec.Command("xdg-open", reportPath).Start()
	}
}

func (program *LProgram) LLinkExternalOpen(urlToOpen string) error {
	wailsRuntime.BrowserOpenURL(program.LContext, urlToOpen)
	return nil
}

func (program *LProgram) LWorkspaceSelect() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(program.LContext, wailsRuntime.OpenDialogOptions{Title: LLocaleTextGetInternal("native.selectWorkspace.title", nil)})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (program *LProgram) LPlanToolchainRequest(buildConfigSettings planning.LSettingsToolchain) (planning.LReviewToolchain, error) {
	plan, err := planning.LPlanToolchainCreate(buildConfigSettings)
	if err != nil {
		return planning.LReviewToolchain{}, err
	}
	reviewSession, err := reviewsession.LSessionReviewCreate(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.LReviewToolchain{}, err
	}
	program.LMutexReviewSession.Lock()
	program.LToolchainReviewStorage[reviewSession.ReviewSessionId] = LReviewToolchainStored{ReviewSession: reviewSession, Plan: plan}
	program.LMutexReviewSession.Unlock()
	return planning.LReviewToolchain{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedLConsentText: reviewSession.ExpectedLConsentText, ExpectedLConsentTextHash: reviewSession.ExpectedLConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

func (program *LProgram) LPlanFFmpegRequest(ffmpegBuildSettings planning.LSettingsFFmpeg) (planning.LReviewFFmpeg, error) {
	plan, err := planning.LPlanFFmpegCreate(ffmpegBuildSettings)
	if err != nil {
		return planning.LReviewFFmpeg{}, err
	}
	reviewSession, err := reviewsession.LSessionReviewCreate(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.LReviewFFmpeg{}, err
	}
	program.LMutexReviewSession.Lock()
	program.LFFmpegReviewStorage[reviewSession.ReviewSessionId] = LReviewFFmpegStored{ReviewSession: reviewSession, Plan: plan}
	program.LMutexReviewSession.Unlock()
	return planning.LReviewFFmpeg{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedLConsentText: reviewSession.ExpectedLConsentText, ExpectedLConsentTextHash: reviewSession.ExpectedLConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

// LPlanToolchainApprove validates the review, confirms, and prepares the
// toolchain asynchronously (GUI behavior).
func (program *LProgram) LPlanToolchainApprove(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.LToolchainApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.LToolchainPrepareLaunch(plan, approval, false)
}

// LToolchainApproveSync is the CLI counterpart: it prepares the toolchain
// inline and returns only after preparation finishes, so the caller can map the
// outcome to an exit code. The final status arrives through the reporter.
func (program *LProgram) LToolchainApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.LToolchainApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.LToolchainPrepareLaunch(plan, approval, true)
}

// LToolchainApproveValidate performs the shared validation for both toolchain
// approve paths and consumes the single-use session only after confirmation.
func (program *LProgram) LToolchainApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanToolchain, error) {
	storedReviewSession, err := program.LToolchainReviewValidate(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanToolchain{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		return planning.LPlanToolchain{}, err
	}
	if err := LHashToolchainVerify(plan); err != nil {
		return planning.LPlanToolchain{}, err
	}
	confirmed, err := program.LNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		return planning.LPlanToolchain{}, err
	}
	if !confirmed {
		return planning.LPlanToolchain{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	program.LToolchainReviewConsume(reviewSessionId)
	return plan, nil
}

// LToolchainPrepareLaunch builds the per-action consents and starts the
// toolchain worker, inline when runInline is true (CLI) or on a goroutine
// otherwise (GUI).
func (program *LProgram) LToolchainPrepareLaunch(plan planning.LPlanToolchain, approval consent.LRequestApproval, runInline bool) (LResultAction, error) {
	userLConsentMsys, err := consent.LConsentMsysCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		return LResultAction{}, err
	}
	if runInline {
		program.LToolchainPrepare(LContextAction, LRunId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	} else {
		go program.LToolchainPrepare(LContextAction, LRunId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	}
	return LResultAction{RunId: LRunId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// LPlanFFmpegApprove validates the review, confirms, and launches the build
// asynchronously (GUI behavior: returns a RunId immediately, progress arrives
// through the reporter).
func (program *LProgram) LPlanFFmpegApprove(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.LFFmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.LFFmpegCompilationLaunch(plan, approval, false)
}

// LFFmpegApproveSync is the CLI counterpart of LPlanFFmpegApprove: it runs
// the build inline and returns only after the build finishes, so the caller can
// map the outcome to an exit code. The final status is delivered through the
// reporter, exactly as in the async path.
func (program *LProgram) LFFmpegApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.LFFmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.LFFmpegCompilationLaunch(plan, approval, true)
}

// LFFmpegApproveValidate performs the shared, side-effect-ordered validation for
// both approve paths: session check, executability, hash, toolchain readiness,
// and the backend-owned confirmation. It consumes the single-use session only
// after confirmation succeeds, so a rejected confirmation leaves it retryable.
func (program *LProgram) LFFmpegApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanFFmpeg, error) {
	storedReviewSession, err := program.LFFmpegReviewValidate(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanFFmpeg{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		return planning.LPlanFFmpeg{}, err
	}
	if err := LHashFFmpegVerify(plan); err != nil {
		return planning.LPlanFFmpeg{}, err
	}
	if err := LToolchainPreparedCheck(plan.WorkspaceDirectory, plan.WindowsShellProfileName); err != nil {
		return planning.LPlanFFmpeg{}, err
	}
	confirmed, err := program.LNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		return planning.LPlanFFmpeg{}, err
	}
	if !confirmed {
		return planning.LPlanFFmpeg{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	program.LFFmpegReviewConsume(reviewSessionId)
	return plan, nil
}

// LFFmpegCompilationLaunch builds the per-action consents and starts the build worker,
// inline when runInline is true (CLI) or on a goroutine otherwise (GUI).
func (program *LProgram) LFFmpegCompilationLaunch(plan planning.LPlanFFmpeg, approval consent.LRequestApproval, runInline bool) (LResultAction, error) {
	userLConsentFFmpeg, err := consent.LConsentFFmpegCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userExternalLConsentCommand, err := consent.LConsentCommandCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		return LResultAction{}, err
	}
	if runInline {
		program.LFFmpegCompile(LContextAction, LRunId, plan, userLConsentFFmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	} else {
		go program.LFFmpegCompile(LContextAction, LRunId, plan, userLConsentFFmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	}
	return LResultAction{RunId: LRunId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// LToolchainReviewValidate checks the session without consuming it, so a
// later step (the native confirmation dialog) can still be cancelled or retried.
// The session is only removed by LToolchainReviewConsume once the user has
// confirmed, which keeps it single-use without losing it on a rejected dialog.
func (program *LProgram) LToolchainReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewToolchainStored, error) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LToolchainReviewStorage[reviewSessionId]
	if !exists {
		return LReviewToolchainStored{}, errors.New("toolchain review session was not found")
	}
	if err := reviewsession.LReviewApprovalCheck(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.LConsentText); err != nil {
		return LReviewToolchainStored{}, err
	}
	return storedReviewSession, nil
}

func (program *LProgram) LToolchainReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LToolchainReviewStorage, reviewSessionId)
}

func (program *LProgram) LFFmpegReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewFFmpegStored, error) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LFFmpegReviewStorage[reviewSessionId]
	if !exists {
		return LReviewFFmpegStored{}, errors.New("FFmpeg review session was not found")
	}
	if err := reviewsession.LReviewApprovalCheck(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.LConsentText); err != nil {
		return LReviewFFmpegStored{}, err
	}
	return storedReviewSession, nil
}

func (program *LProgram) LFFmpegReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LFFmpegReviewStorage, reviewSessionId)
}

func (program *LProgram) LNativeConsentAsk(actionName string, planHash string) (bool, error) {
	if program.LConfirmer == nil {
		return false, errors.New("no approval confirmer is configured")
	}
	return program.LConfirmer.LConfirmerApprovalGet(actionName, planHash)
}

func (program *LProgram) LActionApprovedCancel() bool {
	program.LMutexAction.Lock()
	defer program.LMutexAction.Unlock()
	if program.LActionCancelFunction == nil {
		return false
	}
	program.LActionCancelFunction()
	program.LLogEmit("warn", LLocaleTextGetInternal("logs.system.cancellationRequested", nil))
	return true
}

func (program *LProgram) LActionApprovedStart() (string, context.Context, error) {
	program.LMutexAction.Lock()
	defer program.LMutexAction.Unlock()
	if program.LActionCancelFunction != nil {
		return "", nil, errors.New("an approved action is already running")
	}
	LContextAction, LActionCancelFunction := context.WithCancel(context.Background())
	program.LContextAction = LContextAction
	program.LActionCancelFunction = LActionCancelFunction
	LRunId := time.Now().UTC().Format("20060102T150405Z")
	return LRunId, LContextAction, nil
}

func (program *LProgram) LActionApprovedFinish(status string) {
	program.LMutexAction.Lock()
	program.LActionCancelFunction = nil
	program.LContextAction = nil
	program.LMutexAction.Unlock()
	program.LStatusEmit(status)
}

func (program *LProgram) LAuditProgressCreate(auditWriter *audit.LAuditWriter, actionName string, planHash string) func(string, string) {
	return func(level string, message string) {
		_ = auditWriter.LAuditEventWrite("log", actionName, planHash, level, message)
		program.LLogEmit(level, message)
	}
}

func (program *LProgram) LStatusEmit(status string) {
	if program.LReporter != nil {
		program.LReporter.LReporterStatusEmit(status)
	}
}

func (program *LProgram) LLogEmit(level string, message string) {
	if program.LReporter != nil {
		program.LReporter.LReporterLogEmit(level, message)
	}
}

func (program *LProgram) lStalledEmit(addresses []string) {
	if program.LReporter != nil {
		program.LReporter.LReporterStalledEmit(addresses)
	}
}

func (program *LProgram) LStatusFailureEmit(message string, err error) {
	program.LLogEmit("error", message+": "+err.Error())
	program.LStatusEmit("failed")
}

func (program *LProgram) LErrorLocalizedEmit(messageKey string, fallback string, err error) {
	message := LLocaleTextGetInternal(messageKey, nil)
	if message == messageKey {
		message = fallback
	}
	program.LStatusFailureEmit(message, err)
}

func LPolicyHashResolve(expectedSha256Hash string) download.LPolicyFile {
	if expectedSha256Hash == "" {
		return download.LPolicyFileOverwrite
	}
	return download.LHashReusePolicy
}
