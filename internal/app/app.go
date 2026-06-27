package app

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
	"promptfulcustomffmpegbuilder/internal/reviewsession"
	"promptfulcustomffmpegbuilder/internal/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                         context.Context
	actionContext               context.Context
	actionCancelFunction        context.CancelFunc
	actionMutex                 sync.Mutex
	reviewSessionMutex          sync.Mutex
	toolchainReviewSessionStore map[string]storedToolchainPreparationReviewSession
	ffmpegReviewSessionStore    map[string]storedFfmpegBuildReviewSession
	startupWindowState          windowState
	uiLocale                    string
	uiLocaleMutex               sync.RWMutex
}

type storedToolchainPreparationReviewSession struct {
	ReviewSession reviewsession.PlanReviewSession
	Plan          planning.ToolchainPreparationPlan
}

type storedFfmpegBuildReviewSession struct {
	ReviewSession reviewsession.PlanReviewSession
	Plan          planning.FfmpegBuildPlan
}

type InitialApplicationState struct {
	HostOS                        string                           `json:"hostOs"`
	KindExplanation               string                           `json:"kindExplanation"`
	SecurityRuleSummary           string                           `json:"securityRuleSummary"`
	NamingRuleSummary             string                           `json:"namingRuleSummary"`
	DefaultBuildConfigSettings    planning.BuildConfigSettings     `json:"defaultBuildConfigSettings"`
	DefaultFfmpegBuildSettings    planning.FfmpegBuildSettings     `json:"defaultFfmpegBuildSettings"`
	DefaultLibraryCatalog         []planning.LibraryChoice         `json:"defaultLibraryCatalog"`
	DefaultConfigureOptionCatalog []planning.ConfigureOptionChoice `json:"defaultConfigureOptionCatalog"`
}

type ApprovedActionResult struct {
	RunId     string `json:"runId"`
	StartedAt string `json:"startedAt"`
}

type BuildResultFile struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	SizeBytes  int64  `json:"sizeBytes"`
	Sha256Hash string `json:"sha256Hash"`
}

type BuildResult struct {
	ArtifactsDirectory        string            `json:"artifactsDirectory"`
	ReportPath                string            `json:"reportPath"`
	Files                     []BuildResultFile `json:"files"`
	SelectedLibraries         []string          `json:"selectedLibraries"`
	SelectedConfigureOptions  []string          `json:"selectedConfigureOptions"`
	RequiredMsys2PackageNames []string          `json:"requiredMsys2PackageNames"`
	ConfigureFlags            []string          `json:"configureFlags"`
	LicenseProfileName        string            `json:"licenseProfileName"`
	CreatedAt                 string            `json:"createdAt"`
}

type artifactReport struct {
	CreatedAt                 string                           `json:"createdAt"`
	SelectedLibraries         []planning.LibraryChoice         `json:"selectedLibraries"`
	SelectedConfigureOptions  []planning.ConfigureOptionChoice `json:"selectedConfigureOptions"`
	RequiredMsys2PackageNames []string                         `json:"requiredMsys2PackageNames"`
	ConfigureFlags            []string                         `json:"configureFlags"`
	LicenseProfileName        string                           `json:"licenseProfileName"`
}

func New() *App {
	return &App{startupWindowState: loadWindowState(), toolchainReviewSessionStore: map[string]storedToolchainPreparationReviewSession{}, ffmpegReviewSessionStore: map[string]storedFfmpegBuildReviewSession{}}
}

func InitialWindowSize(app *App) (int, int) {
	return app.startupWindowState.Width, app.startupWindowState.Height
}

func Localize(key string, values map[string]string) string {
	return localize(key, values)
}

func (app *App) Startup(ctx context.Context) {
	app.ctx = ctx
	app.restoreWindowGeometry()
}

// BeforeClose runs while the window still exists, so it is the safe place to
// read window geometry. Returning false allows the close to proceed.
func (app *App) BeforeClose(ctx context.Context) bool {
	app.persistWindowGeometry(ctx)
	return false
}

func (app *App) Shutdown(ctx context.Context) {
	app.CancelApprovedAction()
}

// SetLocale records the UI language so the backend-rendered native confirmation
// dialog can follow it. Only the dialog is localized; log and error messages stay
// English. Unknown locales fall back to English.
func (app *App) SetLocale(locale string) {
	normalizedLocale := "en"
	if locale == "ko" {
		normalizedLocale = "ko"
	}
	app.uiLocaleMutex.Lock()
	app.uiLocale = normalizedLocale
	app.uiLocaleMutex.Unlock()
}

func (app *App) currentLocale() string {
	app.uiLocaleMutex.RLock()
	defer app.uiLocaleMutex.RUnlock()
	if app.uiLocale == "" {
		return "en"
	}
	return app.uiLocale
}

func (app *App) GetInitialApplicationState() InitialApplicationState {
	return InitialApplicationState{
		HostOS:                        runtime.GOOS,
		KindExplanation:               localize("initial.kindExplanation", nil),
		SecurityRuleSummary:           localize("initial.securityRuleSummary", nil),
		NamingRuleSummary:             localize("initial.namingRuleSummary", nil),
		DefaultBuildConfigSettings:    planning.DefaultBuildConfigSettings(),
		DefaultFfmpegBuildSettings:    planning.DefaultFfmpegBuildSettings(),
		DefaultLibraryCatalog:         planning.LibraryCatalogForShellProfile(planning.DefaultFfmpegBuildSettings().WindowsShellProfileName),
		DefaultConfigureOptionCatalog: planning.ConfigureOptionCatalog(),
	}
}

func (app *App) GetBuildResult(workspaceDirectory string) (BuildResult, error) {
	workspaceLayout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return BuildResult{}, err
	}
	if err := os.MkdirAll(workspaceLayout.ArtifactsDirectory, 0o755); err != nil {
		return BuildResult{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return BuildResult{}, err
	}
	result := BuildResult{ArtifactsDirectory: workspaceLayout.ArtifactsDirectory, Files: []BuildResultFile{}, SelectedLibraries: []string{}, SelectedConfigureOptions: []string{}, RequiredMsys2PackageNames: []string{}, ConfigureFlags: []string{}}
	artifactEntries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return BuildResult{}, err
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
		if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, artifactPath); err != nil {
			return BuildResult{}, err
		}
		fileInfo, err := os.Stat(artifactPath)
		if err != nil {
			return BuildResult{}, err
		}
		result.Files = append(result.Files, BuildResultFile{Name: artifactName, Path: artifactPath, SizeBytes: fileInfo.Size(), Sha256Hash: createFileHashOrEmpty(artifactPath)})
	}
	sort.Slice(result.Files, func(leftIndex, rightIndex int) bool {
		return strings.ToLower(result.Files[leftIndex].Name) < strings.ToLower(result.Files[rightIndex].Name)
	})
	reportPath, report, err := readLatestArtifactReport(workspaceLayout)
	if err != nil {
		return result, nil
	}
	result.ReportPath = reportPath
	result.CreatedAt = report.CreatedAt
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

func (app *App) OpenResultFolder(workspaceDirectory string) error {
	workspaceLayout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	if err := os.MkdirAll(workspaceLayout.ArtifactsDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
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

func (app *App) OpenResultReport(workspaceDirectory string) error {
	workspaceLayout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	reportPath, _, err := readLatestArtifactReport(workspaceLayout)
	if err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, reportPath); err != nil {
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

func (app *App) OpenExternalUrl(urlToOpen string) error {
	wailsRuntime.BrowserOpenURL(app.ctx, urlToOpen)
	return nil
}

func (app *App) SelectWorkspace() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(app.ctx, wailsRuntime.OpenDialogOptions{Title: localize("native.selectWorkspace.title", nil)})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (app *App) RequestToolchainPreparationPlan(buildConfigSettings planning.BuildConfigSettings) (planning.ToolchainPreparationPlanReview, error) {
	plan, err := planning.PlanToolchainSetup(buildConfigSettings)
	if err != nil {
		return planning.ToolchainPreparationPlanReview{}, err
	}
	reviewSession, err := reviewsession.NewPlanReviewSession(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.ToolchainPreparationPlanReview{}, err
	}
	app.reviewSessionMutex.Lock()
	app.toolchainReviewSessionStore[reviewSession.ReviewSessionId] = storedToolchainPreparationReviewSession{ReviewSession: reviewSession, Plan: plan}
	app.reviewSessionMutex.Unlock()
	return planning.ToolchainPreparationPlanReview{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedConsentText: reviewSession.ExpectedConsentText, ExpectedConsentTextHash: reviewSession.ExpectedConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

func (app *App) RequestFfmpegBuildPlan(ffmpegBuildSettings planning.FfmpegBuildSettings) (planning.FfmpegBuildPlanReview, error) {
	plan, err := planning.PlanFfmpegBuild(ffmpegBuildSettings)
	if err != nil {
		return planning.FfmpegBuildPlanReview{}, err
	}
	reviewSession, err := reviewsession.NewPlanReviewSession(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.FfmpegBuildPlanReview{}, err
	}
	app.reviewSessionMutex.Lock()
	app.ffmpegReviewSessionStore[reviewSession.ReviewSessionId] = storedFfmpegBuildReviewSession{ReviewSession: reviewSession, Plan: plan}
	app.reviewSessionMutex.Unlock()
	return planning.FfmpegBuildPlanReview{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedConsentText: reviewSession.ExpectedConsentText, ExpectedConsentTextHash: reviewSession.ExpectedConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

func (app *App) ApproveToolchainPreparationPlan(reviewSessionId string, approval consent.ApprovalRequest) (ApprovedActionResult, error) {
	storedReviewSession, err := app.validateToolchainReviewSession(reviewSessionId, approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.CheckPlanCanRun(plan.IsExecutable); err != nil {
		return ApprovedActionResult{}, err
	}
	if err := verifyToolchainPlanHash(plan); err != nil {
		return ApprovedActionResult{}, err
	}
	confirmedByNativeDialog, err := app.askNativeUserApproval(plan.ActionName, plan.PlanHash)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	if !confirmedByNativeDialog {
		return ApprovedActionResult{}, errors.New("user rejected approval in backend-owned native confirmation dialog")
	}
	// Confirmed: consume the session now so it is single-use, but only after the
	// dialog succeeded, so a cancelled/failed dialog leaves it retryable.
	app.consumeToolchainReviewSession(reviewSessionId)
	userMsys2DownloadConsent, err := consent.Msys2DownloadApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	userArchiveExtractionConsent, err := consent.ArchiveExtractionApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	userPacmanPackageInstallConsent, err := consent.PacmanInstallApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	runId, actionContext, err := app.startApprovedAction()
	if err != nil {
		return ApprovedActionResult{}, err
	}
	go app.prepareToolchain(actionContext, runId, plan, userMsys2DownloadConsent, userArchiveExtractionConsent, userPacmanPackageInstallConsent)
	return ApprovedActionResult{RunId: runId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

func (app *App) ApproveFfmpegBuildPlan(reviewSessionId string, approval consent.ApprovalRequest) (ApprovedActionResult, error) {
	storedReviewSession, err := app.validateFfmpegReviewSession(reviewSessionId, approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.CheckPlanCanRun(plan.IsExecutable); err != nil {
		return ApprovedActionResult{}, err
	}
	if err := verifyFfmpegPlanHash(plan); err != nil {
		return ApprovedActionResult{}, err
	}
	if err := checkToolchainPreparedForBuild(plan.WorkspaceDirectory, plan.WindowsShellProfileName); err != nil {
		return ApprovedActionResult{}, err
	}
	confirmedByNativeDialog, err := app.askNativeUserApproval(plan.ActionName, plan.PlanHash)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	if !confirmedByNativeDialog {
		return ApprovedActionResult{}, errors.New("user rejected approval in backend-owned native confirmation dialog")
	}
	// Confirmed: consume the session now so it is single-use, but only after the
	// dialog succeeded, so a cancelled/failed dialog leaves it retryable.
	app.consumeFfmpegReviewSession(reviewSessionId)
	userFfmpegSourceDownloadConsent, err := consent.FfmpegSourceDownloadApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	userArchiveExtractionConsent, err := consent.ArchiveExtractionApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	userExternalCommandExecutionConsent, err := consent.CommandExecutionApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	userPacmanPackageInstallConsent, err := consent.PacmanInstallApproval(approval)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	runId, actionContext, err := app.startApprovedAction()
	if err != nil {
		return ApprovedActionResult{}, err
	}
	go app.buildFfmpeg(actionContext, runId, plan, userFfmpegSourceDownloadConsent, userArchiveExtractionConsent, userPacmanPackageInstallConsent, userExternalCommandExecutionConsent)
	return ApprovedActionResult{RunId: runId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// validateToolchainReviewSession checks the session without consuming it, so a
// later step (the native confirmation dialog) can still be cancelled or retried.
// The session is only removed by consumeToolchainReviewSession once the user has
// confirmed, which keeps it single-use without losing it on a rejected dialog.
func (app *App) validateToolchainReviewSession(reviewSessionId string, approval consent.ApprovalRequest) (storedToolchainPreparationReviewSession, error) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	storedReviewSession, exists := app.toolchainReviewSessionStore[reviewSessionId]
	if !exists {
		return storedToolchainPreparationReviewSession{}, errors.New("toolchain review session was not found")
	}
	if err := reviewsession.CheckReviewApproval(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.ConsentText); err != nil {
		return storedToolchainPreparationReviewSession{}, err
	}
	return storedReviewSession, nil
}

func (app *App) consumeToolchainReviewSession(reviewSessionId string) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	delete(app.toolchainReviewSessionStore, reviewSessionId)
}

func (app *App) validateFfmpegReviewSession(reviewSessionId string, approval consent.ApprovalRequest) (storedFfmpegBuildReviewSession, error) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	storedReviewSession, exists := app.ffmpegReviewSessionStore[reviewSessionId]
	if !exists {
		return storedFfmpegBuildReviewSession{}, errors.New("FFmpeg review session was not found")
	}
	if err := reviewsession.CheckReviewApproval(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.ConsentText); err != nil {
		return storedFfmpegBuildReviewSession{}, err
	}
	return storedReviewSession, nil
}

func (app *App) consumeFfmpegReviewSession(reviewSessionId string) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	delete(app.ffmpegReviewSessionStore, reviewSessionId)
}

func (app *App) askNativeUserApproval(actionName string, planHash string) (bool, error) {
	if app.ctx == nil {
		return false, errors.New("application context is not ready for native approval dialog")
	}
	locale := app.currentLocale()
	message := localizeForLocale(locale, "native.approval.message", map[string]string{"action": localizeForLocale(locale, "approval.action."+actionName, nil), "planHash": planHash})
	noButtonLabel := localizeForLocale(locale, "native.approval.no", nil)
	yesButtonLabel := localizeForLocale(locale, "native.approval.yes", nil)
	choice, err := wailsRuntime.MessageDialog(app.ctx, wailsRuntime.MessageDialogOptions{
		Type:          wailsRuntime.QuestionDialog,
		Title:         localizeForLocale(locale, "native.approval.title", nil),
		Message:       message,
		Buttons:       []string{noButtonLabel, yesButtonLabel},
		DefaultButton: noButtonLabel,
		CancelButton:  noButtonLabel,
	})
	if err != nil {
		return false, err
	}
	// On Windows, Wails' QuestionDialog ignores custom button labels and returns
	// the native "Yes"/"No" strings, so a localized yes label would never match.
	// Accept the localized label or the native English "Yes".
	return choice == yesButtonLabel || choice == "Yes", nil
}

func (app *App) CancelApprovedAction() bool {
	app.actionMutex.Lock()
	defer app.actionMutex.Unlock()
	if app.actionCancelFunction == nil {
		return false
	}
	app.actionCancelFunction()
	app.emitLog("warn", localize("logs.system.cancellationRequested", nil))
	return true
}

func (app *App) startApprovedAction() (string, context.Context, error) {
	app.actionMutex.Lock()
	defer app.actionMutex.Unlock()
	if app.actionCancelFunction != nil {
		return "", nil, errors.New("an approved action is already running")
	}
	actionContext, actionCancelFunction := context.WithCancel(context.Background())
	app.actionContext = actionContext
	app.actionCancelFunction = actionCancelFunction
	runId := time.Now().UTC().Format("20060102T150405Z")
	return runId, actionContext, nil
}

func (app *App) finishApprovedAction(status string) {
	app.actionMutex.Lock()
	app.actionCancelFunction = nil
	app.actionContext = nil
	app.actionMutex.Unlock()
	app.emitStatus(status)
}

func (app *App) createAuditedProgressFunc(auditWriter *audit.Writer, actionName string, planHash string) func(string, string) {
	return func(level string, message string) {
		_ = auditWriter.WriteEvent("log", actionName, planHash, level, message)
		app.emitLog(level, message)
	}
}

func (app *App) emitStatus(status string) {
	if app.ctx != nil {
		wailsRuntime.EventsEmit(app.ctx, "approved-action-status", map[string]string{"status": status})
	}
}

func (app *App) emitLog(level string, message string) {
	if app.ctx != nil {
		wailsRuntime.EventsEmit(app.ctx, "security-log", map[string]string{"level": level, "message": message})
	}
}

func (app *App) emitFailure(message string, err error) {
	app.emitLog("error", message+": "+err.Error())
	app.emitStatus("failed")
}

func (app *App) emitLocalizedFailure(messageKey string, fallback string, err error) {
	message := localize(messageKey, nil)
	if message == messageKey {
		message = fallback
	}
	app.emitFailure(message, err)
}

func downloadPolicyForHash(expectedSha256Hash string) download.FileConflictPolicy {
	if expectedSha256Hash == "" {
		return download.OverwriteFile
	}
	return download.ReuseIfHashMatches
}
