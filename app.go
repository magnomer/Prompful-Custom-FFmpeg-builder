package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"customffmpegbuilder/internal/audit"
	"customffmpegbuilder/internal/consent"
	"customffmpegbuilder/internal/download"
	"customffmpegbuilder/internal/execution"
	"customffmpegbuilder/internal/extraction"
	"customffmpegbuilder/internal/planning"
	"customffmpegbuilder/internal/reviewsession"
	"customffmpegbuilder/internal/scripting"
	"customffmpegbuilder/internal/workspace"

	"github.com/ProtonMail/go-crypto/openpgp"

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
	DefaultBuildToolSettings      planning.BuildToolSettings       `json:"defaultBuildToolSettings"`
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

func NewApp() *App {
	return &App{toolchainReviewSessionStore: map[string]storedToolchainPreparationReviewSession{}, ffmpegReviewSessionStore: map[string]storedFfmpegBuildReviewSession{}}
}

func (app *App) Startup(ctx context.Context)  { app.ctx = ctx }
func (app *App) Shutdown(ctx context.Context) { app.CancelApprovedAction() }

func (app *App) GetInitialApplicationState() InitialApplicationState {
	return InitialApplicationState{
		HostOS:                        runtime.GOOS,
		KindExplanation:               "This app contains only the GUI and build orchestrator. It does not bundle FFmpeg, codec libraries, multimedia libraries, or generated FFmpeg binaries.",
		SecurityRuleSummary:           "Every download, extraction, package installation, deletion, or command execution requires an action-specific user consent value that matches the approved plan hash.",
		NamingRuleSummary:             "Go uses PascalCase exported names and lowerCamel local names. TypeScript uses PascalCase types and lowerCamel functions. CSS uses strict BEM: block, block__element, block--modifier.",
		DefaultBuildToolSettings:      planning.DefaultBuildToolSettings(),
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
	for _, artifactName := range []string{"ffmpeg.exe", "ffprobe.exe"} {
		artifactPath := filepath.Join(workspaceLayout.ArtifactsDirectory, artifactName)
		if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, artifactPath); err != nil {
			return BuildResult{}, err
		}
		fileInfo, err := os.Stat(artifactPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return BuildResult{}, err
		}
		result.Files = append(result.Files, BuildResultFile{Name: artifactName, Path: artifactPath, SizeBytes: fileInfo.Size(), Sha256Hash: createFileHashOrEmpty(artifactPath)})
	}
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
			result.SelectedLibraries = append(result.SelectedLibraries, library.DisplayName+" ("+library.LicenseEffectName+")")
		} else {
			result.SelectedLibraries = append(result.SelectedLibraries, library.DisplayName)
		}
	}
	for _, option := range report.SelectedConfigureOptions {
		if option.DisplayName != "" {
			result.SelectedConfigureOptions = append(result.SelectedConfigureOptions, option.DisplayName)
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

func (app *App) OpenExternalUrl(urlToOpen string) error {
	wailsRuntime.BrowserOpenURL(app.ctx, urlToOpen)
	return nil
}

func (app *App) SelectWorkspace() (string, error) {
	selection, err := wailsRuntime.OpenDirectoryDialog(app.ctx, wailsRuntime.OpenDialogOptions{Title: "Choose CustomFFmpeg workspace"})
	if err != nil {
		return "", err
	}
	return selection, nil
}

func (app *App) RequestToolchainPreparationPlan(buildToolSettings planning.BuildToolSettings) (planning.ToolchainPreparationPlanReview, error) {
	plan, err := planning.PlanToolchainSetup(buildToolSettings)
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
	storedReviewSession, err := app.takeToolchainReviewSession(reviewSessionId, approval)
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
	storedReviewSession, err := app.takeFfmpegReviewSession(reviewSessionId, approval)
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
	confirmedByNativeDialog, err := app.askNativeUserApproval(plan.ActionName, plan.PlanHash)
	if err != nil {
		return ApprovedActionResult{}, err
	}
	if !confirmedByNativeDialog {
		return ApprovedActionResult{}, errors.New("user rejected approval in backend-owned native confirmation dialog")
	}
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

func (app *App) takeToolchainReviewSession(reviewSessionId string, approval consent.ApprovalRequest) (storedToolchainPreparationReviewSession, error) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	storedReviewSession, exists := app.toolchainReviewSessionStore[reviewSessionId]
	if !exists {
		return storedToolchainPreparationReviewSession{}, errors.New("toolchain review session was not found")
	}
	if err := reviewsession.CheckReviewApproval(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.ConsentText); err != nil {
		return storedToolchainPreparationReviewSession{}, err
	}
	delete(app.toolchainReviewSessionStore, reviewSessionId)
	return storedReviewSession, nil
}

func (app *App) takeFfmpegReviewSession(reviewSessionId string, approval consent.ApprovalRequest) (storedFfmpegBuildReviewSession, error) {
	app.reviewSessionMutex.Lock()
	defer app.reviewSessionMutex.Unlock()
	storedReviewSession, exists := app.ffmpegReviewSessionStore[reviewSessionId]
	if !exists {
		return storedFfmpegBuildReviewSession{}, errors.New("FFmpeg review session was not found")
	}
	if err := reviewsession.CheckReviewApproval(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.ConsentText); err != nil {
		return storedFfmpegBuildReviewSession{}, err
	}
	delete(app.ffmpegReviewSessionStore, reviewSessionId)
	return storedReviewSession, nil
}

func (app *App) askNativeUserApproval(actionName string, planHash string) (bool, error) {
	if app.ctx == nil {
		return false, errors.New("application context is not ready for native approval dialog")
	}
	message := "The frontend sent an approval request.\n\n" +
		"Approve running this exact backend plan?\n\n" +
		"Action: " + actionName + "\n" +
		"Plan hash: " + planHash + "\n\n" +
		"Choose No if you did not intentionally approve this action."
	choice, err := wailsRuntime.MessageDialog(app.ctx, wailsRuntime.MessageDialogOptions{
		Type:          wailsRuntime.QuestionDialog,
		Title:         "Confirm backend approval",
		Message:       message,
		Buttons:       []string{"No", "Yes"},
		DefaultButton: "No",
		CancelButton:  "No",
	})
	if err != nil {
		return false, err
	}
	return choice == "Yes", nil
}

func (app *App) CancelApprovedAction() bool {
	app.actionMutex.Lock()
	defer app.actionMutex.Unlock()
	if app.actionCancelFunction == nil {
		return false
	}
	app.actionCancelFunction()
	app.emitLog("warn", "Cancellation requested by user.")
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

func (app *App) prepareToolchain(ctx context.Context, runId string, plan planning.ToolchainPreparationPlan, userMsys2DownloadConsent consent.Msys2DownloadConsent, userArchiveExtractionConsent consent.ArchiveExtractionConsent, userPacmanPackageInstallConsent consent.PacmanInstallConsent) {
	actionSucceeded := false
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	defer func() {
		if actionSucceeded {
			app.finishApprovedAction("completed")
			return
		}
		app.cleanupFailedToolchainRun(plan, workspaceLayout, runId)
		app.finishApprovedAction("failed")
	}()
	app.emitStatus("preparing-toolchain")
	if err := workspace.CreateWorkspaceFolders(workspaceLayout); err != nil {
		app.emitFailure("Could not create workspace directories", err)
		return
	}
	auditWriter, err := audit.NewWriter(workspaceLayout.LogsDirectory, runId)
	if err != nil {
		app.emitFailure("Could not create audit log", err)
		return
	}
	emitProgress := app.createAuditedProgressFunc(auditWriter, plan.ActionName, plan.PlanHash)
	_ = auditWriter.WriteEvent("action-started", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 preparation started.")
	emitProgress("info", "Approved private MSYS2 preparation started. Run: "+runId)

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-approved"+archiveExtensionFromUrl(plan.Msys2ArchiveUrl))
	downloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 archive", DownloadUrl: plan.Msys2ArchiveUrl, ExpectedSha256Hash: plan.Msys2ArchiveSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 500_000_000, FileConflictPolicy: downloadPolicyForHash(plan.Msys2ArchiveSha256Hash)}
	if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, downloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("MSYS2 archive download failed", err)
		return
	}
	if plan.Msys2ArchiveSignatureUrl != "" {
		signaturePath := archivePath + ".sig"
		signatureDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 signature", DownloadUrl: plan.Msys2ArchiveSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
		if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, signatureDownloadPlan, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitFailure("MSYS2 signature download failed", err)
			return
		}
		publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-installer-signing-key.asc")
		publicKeyDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 installer signing key", DownloadUrl: msys2InstallerSigningKeyUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"keyserver.ubuntu.com"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 1_000_000, FileConflictPolicy: download.OverwriteFile}
		if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, publicKeyDownloadPlan, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitFailure("MSYS2 signing key download failed", err)
			return
		}
		if err := verifyMsys2DetachedSignature(signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitFailure("MSYS2 signature verification failed", err)
			return
		}
	}
	if err := app.prepareFreshToolchainDirectory(plan.WorkspaceDirectory, plan.Msys2RootDirectory, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("MSYS2 private toolchain folder cleanup failed", err)
		return
	}
	extractPlan := extraction.ExtractPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, ArchiveFormat: archiveFormatFromUrl(plan.Msys2ArchiveUrl), ExtractDestinationPolicy: extraction.RequireNewDirectory, ExtractedFileModePolicy: extraction.PreserveSafeExecutableBits, MaximumFileCount: 250000, MaximumExtractedByteCount: 10_000_000_000, MaximumSingleFileByteCount: 2_000_000_000}
	if err := extraction.ExtractArchiveWithConsent(ctx, userArchiveExtractionConsent, extractPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("MSYS2 archive extraction failed", err)
		return
	}
	if err := normalizeMsys2RootDirectory(plan.Msys2RootDirectory, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("MSYS2 archive layout check failed", err)
		return
	}
	if err := app.installPacmanPackages(ctx, plan, userPacmanPackageInstallConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("MSYS2 package installation failed", err)
		return
	}
	_ = auditWriter.WriteEvent("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 environment is ready.")
	emitProgress("info", "Approved private MSYS2 environment is ready.")
	actionSucceeded = true
}

func (app *App) prepareFreshToolchainDirectory(workspaceDirectory string, msys2RootDirectory string, emitProgress func(string, string)) error {
	if msys2RootDirectory == "" {
		return errors.New("private MSYS2 toolchain folder is empty")
	}
	if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, msys2RootDirectory); err != nil {
		return err
	}
	if _, err := os.Lstat(msys2RootDirectory); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("could not inspect existing private MSYS2 folder: %w", err)
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, msys2RootDirectory); err != nil {
		return fmt.Errorf("existing private MSYS2 folder is unsafe to clean: %w", err)
	}
	emitProgress("warn", "A previous private MSYS2 folder already exists. Removing it before extracting the newly approved archive.")
	stopPrivateMsys2BackgroundAgents(msys2RootDirectory, emitProgress)
	if err := removeAllWithRetry(msys2RootDirectory); err != nil {
		return fmt.Errorf("could not remove previous private MSYS2 folder: %w", err)
	}
	emitProgress("info", "Previous private MSYS2 folder removed. Fresh extraction can start.")
	return nil
}

func normalizeMsys2RootDirectory(msys2RootDirectory string, emitProgress func(string, string)) error {
	if fileExists(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		emitProgress("info", "MSYS2 archive layout confirmed.")
		return nil
	}

	entries, err := os.ReadDir(msys2RootDirectory)
	if err != nil {
		return fmt.Errorf("could not inspect extracted MSYS2 directory: %w", err)
	}

	candidateDirectories := []string{}
	for _, entry := range entries {
		if entry.IsDir() {
			candidateDirectories = append(candidateDirectories, filepath.Join(msys2RootDirectory, entry.Name()))
		}
	}

	matchingDirectories := []string{}
	for _, candidateDirectory := range candidateDirectories {
		if fileExists(filepath.Join(candidateDirectory, "usr", "bin", "bash.exe")) {
			matchingDirectories = append(matchingDirectories, candidateDirectory)
		}
	}

	if len(matchingDirectories) == 0 {
		return errors.New("could not find usr/bin/bash.exe after extraction; check that the MSYS2 URL points to a base tar archive, not an installer, signature, or HTML page")
	}
	if len(matchingDirectories) > 1 {
		return fmt.Errorf("found multiple possible MSYS2 roots after extraction: %s", strings.Join(matchingDirectories, ", "))
	}

	nestedRootDirectory := matchingDirectories[0]
	emitProgress("info", "MSYS2 archive contains a top-level folder. Moving its contents into the private toolchain folder.")
	if err := moveDirectoryContents(nestedRootDirectory, msys2RootDirectory); err != nil {
		return err
	}
	if err := os.Remove(nestedRootDirectory); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not remove empty extracted MSYS2 wrapper folder: %w", err)
	}
	if !fileExists(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		return errors.New("MSYS2 archive layout normalization did not produce usr/bin/bash.exe")
	}
	emitProgress("info", "MSYS2 archive layout normalized.")
	return nil
}

func moveDirectoryContents(sourceDirectory string, destinationDirectory string) error {
	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return fmt.Errorf("could not read extracted MSYS2 wrapper folder: %w", err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDirectory, entry.Name())
		destinationPath := filepath.Join(destinationDirectory, entry.Name())
		if fileExists(destinationPath) {
			return fmt.Errorf("cannot move extracted MSYS2 entry because destination already exists: %s", destinationPath)
		}
		if err := os.Rename(sourcePath, destinationPath); err != nil {
			return fmt.Errorf("could not move extracted MSYS2 entry %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (app *App) buildFfmpeg(ctx context.Context, runId string, plan planning.FfmpegBuildPlan, userFfmpegSourceDownloadConsent consent.FfmpegSourceDownloadConsent, userArchiveExtractionConsent consent.ArchiveExtractionConsent, userLibraryPackageInstallConsent consent.PacmanInstallConsent, userExternalCommandExecutionConsent consent.CommandExecutionConsent) {
	actionSucceeded := false
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	sourceRootDirectory := filepath.Join(workspaceLayout.SourcesDirectory, "ffmpeg-"+runId)
	ffmpegSourceDirectory := ""
	defer func() {
		if actionSucceeded {
			app.finishApprovedAction("completed")
			return
		}
		app.saveConfigLog(ffmpegSourceDirectory, workspaceLayout.WorkspaceDirectory)
		app.cleanupFailedFfmpegRun(plan, workspaceLayout, sourceRootDirectory, runId)
		app.finishApprovedAction("failed")
	}()
	app.emitStatus("building-ffmpeg")
	if err := workspace.CreateWorkspaceFolders(workspaceLayout); err != nil {
		app.emitFailure("Could not create workspace directories", err)
		return
	}
	auditWriter, err := audit.NewWriter(workspaceLayout.LogsDirectory, runId)
	if err != nil {
		app.emitFailure("Could not create audit log", err)
		return
	}
	emitProgress := app.createAuditedProgressFunc(auditWriter, plan.ActionName, plan.PlanHash)
	_ = auditWriter.WriteEvent("action-started", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build started.")
	emitProgress("info", "Approved FFmpeg build started. Run: "+runId)

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+archiveExtensionFromUrl(plan.FfmpegSourceArchiveUrl))
	downloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg", DownloadUrl: plan.FfmpegSourceArchiveUrl, ExpectedSha256Hash: plan.FfmpegSourceSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"ffmpeg.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 200_000_000, FileConflictPolicy: downloadPolicyForHash(plan.FfmpegSourceSha256Hash)}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, downloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg source download failed", err)
		return
	}
	signaturePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+archiveExtensionFromUrl(plan.FfmpegSourceArchiveUrl)+".asc")
	signatureDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg signature", DownloadUrl: plan.FfmpegSourceSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"ffmpeg.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, signatureDownloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg source signature download failed", err)
		return
	}
	publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-devel.asc")
	publicKeyDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg release signing key", DownloadUrl: ffmpegReleaseSigningKeyUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"ffmpeg.org"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, publicKeyDownloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg signing key download failed", err)
		return
	}
	if err := verifyFfmpegDetachedSignature(signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg source signature verification failed", err)
		return
	}
	extractPlan := extraction.ExtractPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: sourceRootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, ArchiveFormat: archiveFormatFromUrl(plan.FfmpegSourceArchiveUrl), ExtractDestinationPolicy: extraction.RequireNewDirectory, ExtractedFileModePolicy: extraction.PreserveSafeExecutableBits, MaximumFileCount: 50000, MaximumExtractedByteCount: 2_000_000_000, MaximumSingleFileByteCount: 500_000_000}
	if err := extraction.ExtractArchiveWithConsent(ctx, userArchiveExtractionConsent, extractPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg source extraction failed", err)
		return
	}
	ffmpegSourceDirectory, err = findSingleChildDirectory(sourceRootDirectory)
	if err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("Could not locate extracted FFmpeg source directory", err)
		return
	}
	if err := app.installFfmpegLibraryPackages(ctx, plan, userLibraryPackageInstallConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg library package installation failed", err)
		return
	}
	if err := app.executeFfmpegConfigure(ctx, plan, ffmpegSourceDirectory, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg configure failed", err)
		return
	}
	if err := app.executeFfmpegMake(ctx, plan, ffmpegSourceDirectory, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("FFmpeg build failed", err)
		return
	}
	if err := copyFfmpegArtifacts(ffmpegSourceDirectory, workspaceLayout); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("Could not copy FFmpeg artifacts", err)
		return
	}
	if err := writeArtifactReport(workspaceLayout, runId, plan); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitFailure("Could not write artifact report", err)
		return
	}
	_ = auditWriter.WriteEvent("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build completed.")
	emitProgress("info", "Approved FFmpeg build completed. Artifact report written.")
	actionSucceeded = true
}

func (app *App) cleanupFailedToolchainRun(plan planning.ToolchainPreparationPlan, workspaceLayout workspace.WorkspaceLayout, runId string) {
	app.emitLog("warn", "Cleaning partial build-tool files from the failed run.")
	cleanupTargets := []string{
		plan.Msys2RootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "pacman-install-"+plan.PlanHash+".sh"),
	}
	app.cleanupWorkspaceTargets(plan.WorkspaceDirectory, cleanupTargets)
}

func (app *App) saveConfigLog(ffmpegSourceDirectory string, workspaceDirectory string) {
	if ffmpegSourceDirectory == "" {
		return
	}
	configLogPath := filepath.Join(ffmpegSourceDirectory, "ffbuild", "config.log")
	if _, err := os.Stat(configLogPath); err != nil {
		return
	}
	destPath := filepath.Join(workspaceDirectory, "ffmpeg-config.log")
	data, err := os.ReadFile(configLogPath)
	if err != nil {
		app.emitLog("warn", "Could not read config.log: "+err.Error())
		return
	}
	if err := os.WriteFile(destPath, data, 0o600); err != nil {
		app.emitLog("warn", "Could not save config.log to workspace: "+err.Error())
		return
	}
	app.emitLog("info", "Saved FFmpeg config.log to: "+destPath)
}

func (app *App) cleanupFailedFfmpegRun(plan planning.FfmpegBuildPlan, workspaceLayout workspace.WorkspaceLayout, sourceRootDirectory string, runId string) {
	app.emitLog("warn", "Cleaning partial FFmpeg build files from the failed run.")
	cleanupTargets := []string{
		sourceRootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-library-packages-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-configure-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-make-"+plan.PlanHash+".sh"),
	}
	app.cleanupWorkspaceTargets(plan.WorkspaceDirectory, cleanupTargets)
}

func (app *App) cleanupWorkspaceTargets(workspaceDirectory string, targets []string) {
	for _, targetPath := range targets {
		if targetPath == "" {
			continue
		}
		if _, err := os.Lstat(targetPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			app.emitLog("warn", "Could not inspect partial path before cleanup: "+err.Error())
			continue
		}
		if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, targetPath); err != nil {
			app.emitLog("warn", "Skipped cleanup target outside workspace: "+err.Error())
			continue
		}
		if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, targetPath); err != nil {
			app.emitLog("warn", "Skipped unsafe cleanup target: "+err.Error())
			continue
		}
		stopPrivateMsys2BackgroundAgents(targetPath, app.emitLog)
		if err := removeAllWithRetry(targetPath); err != nil {
			app.emitLog("warn", "Could not remove partial path from failed run: "+err.Error())
			continue
		}
		app.emitLog("info", "Removed partial path from failed run: "+targetPath)
	}
}

func stopPrivateMsys2BackgroundAgents(msys2RootDirectory string, emitProgress func(string, string)) {
	bashPath := filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")
	if !fileExists(bashPath) {
		return
	}
	command := exec.Command(bashPath, "-lc", "GNUPGHOME=/etc/pacman.d/gnupg gpgconf --kill gpg-agent >/dev/null 2>&1 || true; GNUPGHOME=/etc/pacman.d/gnupg gpgconf --kill all >/dev/null 2>&1 || true")
	command.Dir = msys2RootDirectory
	command.Env = append(os.Environ(), "MSYSTEM=UCRT64", "MSYS2_PATH_TYPE=inherit", "CHERE_INVOKING=1")
	if err := command.Run(); err == nil && emitProgress != nil {
		emitProgress("info", "Stopped private MSYS2 signing agents before cleanup.")
	}
}

func removeAllWithRetry(targetPath string) error {
	var lastErr error
	for attempt := 0; attempt < 8; attempt++ {
		if err := os.RemoveAll(targetPath); err != nil {
			lastErr = err
			time.Sleep(time.Duration(250+attempt*150) * time.Millisecond)
			continue
		}
		if _, err := os.Lstat(targetPath); errors.Is(err, os.ErrNotExist) {
			return nil
		} else if err != nil {
			return err
		}
		lastErr = fmt.Errorf("path still exists after cleanup attempt: %s", targetPath)
		time.Sleep(time.Duration(250+attempt*150) * time.Millisecond)
	}
	if lastErr == nil {
		lastErr = fmt.Errorf("could not remove path: %s", targetPath)
	}
	return lastErr
}

func (app *App) installPacmanPackages(ctx context.Context, plan planning.ToolchainPreparationPlan, userPacmanPackageInstallConsent consent.PacmanInstallConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	if err := consent.CheckConsent(userPacmanPackageInstallConsent.Consent, consent.ConsentKindPacmanPackageInstallation, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.PacmanInstallScriptLines(plan.Msys2PackageNames)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "pacman-install-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, ScriptKind: execution.PacmanInstallScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLogDirectory: auditWriter.LogDirectory()}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved pacman package installation script.")
	return execution.RunPacmanWithConsent(ctx, userPacmanPackageInstallConsent, commandPlan, emitProgress)
}

func (app *App) installFfmpegLibraryPackages(ctx context.Context, plan planning.FfmpegBuildPlan, userLibraryPackageInstallConsent consent.PacmanInstallConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	if len(plan.RequiredMsys2PackageNames) == 0 {
		emitProgress("info", "No extra MSYS2 library packages are required by the selected FFmpeg libraries.")
		return nil
	}
	if err := consent.CheckConsent(userLibraryPackageInstallConsent.Consent, consent.ConsentKindPacmanPackageInstallation, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.FfmpegLibraryPackageInstallScriptLines(plan.RequiredMsys2PackageNames)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-library-packages-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, ScriptKind: execution.PacmanInstallScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLogDirectory: auditWriter.LogDirectory()}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg library package installation script.")
	return execution.RunPacmanWithConsent(ctx, userLibraryPackageInstallConsent, commandPlan, emitProgress)
}

func pkgConfigPathFor(plan planning.FfmpegBuildPlan) string {
	profileDirectoryName := strings.ToLower(plan.WindowsShellProfileName)
	if profileDirectoryName == "" {
		profileDirectoryName = "ucrt64"
	}
	msys2Prefix := "/" + profileDirectoryName
	paths := []string{
		msys2Prefix + "/lib/pkgconfig",
		msys2Prefix + "/share/pkgconfig",
		"/usr/lib/pkgconfig",
		"/usr/share/pkgconfig",
	}
	return strings.Join(paths, ":")
}

func (app *App) executeFfmpegConfigure(ctx context.Context, plan planning.FfmpegBuildPlan, ffmpegSourceDirectory string, userExternalCommandExecutionConsent consent.CommandExecutionConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.ConfigureScriptLines(plan.ConfigureFlags)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-configure-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: ffmpegSourceDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{"PKG_CONFIG_PATH": pkgConfigPathFor(plan)}, AllowedExecutableBasenames: []string{"bash.exe"}, ScriptKind: execution.FfmpegConfigureScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLogDirectory: auditWriter.LogDirectory()}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg configure script.")
	return execution.RunCommandWithConsent(ctx, userExternalCommandExecutionConsent, commandPlan, emitProgress)
}

func (app *App) executeFfmpegMake(ctx context.Context, plan planning.FfmpegBuildPlan, ffmpegSourceDirectory string, userExternalCommandExecutionConsent consent.CommandExecutionConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.MakeScriptLines(plan.ParallelJobCount)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-make-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: ffmpegSourceDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, ScriptKind: execution.FfmpegMakeScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLogDirectory: auditWriter.LogDirectory()}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg make script.")
	return execution.RunCommandWithConsent(ctx, userExternalCommandExecutionConsent, commandPlan, emitProgress)
}

func verifyToolchainPlanHash(plan planning.ToolchainPreparationPlan) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.HashPlan(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("toolchain plan hash does not match plan content")
	}
	return nil
}

func verifyFfmpegPlanHash(plan planning.FfmpegBuildPlan) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.HashPlan(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("FFmpeg build plan hash does not match plan content")
	}
	return nil
}

func findSingleChildDirectory(parentDirectory string) (string, error) {
	entries, err := os.ReadDir(parentDirectory)
	if err != nil {
		return "", err
	}
	childDirectories := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			childDirectories = append(childDirectories, filepath.Join(parentDirectory, entry.Name()))
		}
	}
	if len(childDirectories) != 1 {
		return "", fmt.Errorf("expected exactly one extracted source directory, found %d", len(childDirectories))
	}
	return childDirectories[0], nil
}

func copyFfmpegArtifacts(ffmpegSourceDirectory string, workspaceLayout workspace.WorkspaceLayout) error {
	artifactNames := []string{"ffmpeg.exe", "ffprobe.exe"}
	for _, artifactName := range artifactNames {
		sourcePath := filepath.Join(ffmpegSourceDirectory, artifactName)
		if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, sourcePath); err != nil {
			return err
		}
		if _, err := os.Stat(sourcePath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return err
		}
		destinationPath := filepath.Join(workspaceLayout.ArtifactsDirectory, artifactName)
		if err := copyFile(workspaceLayout.WorkspaceDirectory, sourcePath, destinationPath); err != nil {
			return err
		}
	}
	return nil
}

func copyFile(workspaceDirectory string, sourcePath string, destinationPath string) error {
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, sourcePath); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, filepath.Dir(destinationPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, destinationPath); err != nil {
		return err
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destinationFile, sourceFile)
	closeErr := destinationFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func writeArtifactReport(workspaceLayout workspace.WorkspaceLayout, runId string, plan planning.FfmpegBuildPlan) error {
	reportPath := filepath.Join(workspaceLayout.ArtifactsDirectory, "build-report-"+runId+".json")
	ffmpegExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffmpeg.exe")
	ffprobeExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffprobe.exe")
	report := map[string]interface{}{"runId": runId, "createdAt": time.Now().UTC().Format(time.RFC3339), "approvedPlanHash": plan.PlanHash, "ffmpegSourceArchiveUrl": plan.FfmpegSourceArchiveUrl, "ffmpegSourceSignatureUrl": plan.FfmpegSourceSignatureUrl, "ffmpegSourceSha256Hash": plan.FfmpegSourceSha256Hash, "selectedLibraries": plan.SelectedLibraries, "selectedConfigureOptions": plan.SelectedConfigureOptions, "requiredMsys2PackageNames": plan.RequiredMsys2PackageNames, "generatedConfigureFlags": plan.GeneratedConfigureFlags, "generatedOptionFlags": plan.GeneratedOptionFlags, "extraConfigureFlags": plan.ExtraConfigureFlags, "configureFlags": plan.ConfigureFlags, "licenseProfileName": plan.LicenseProfileName, "ffmpegExecutablePath": ffmpegExecutablePath, "ffmpegExecutableSha256Hash": createFileHashOrEmpty(ffmpegExecutablePath), "ffmpegExecutableSizeBytes": fileSizeOrZero(ffmpegExecutablePath), "ffprobeExecutablePath": ffprobeExecutablePath, "ffprobeExecutableSha256Hash": createFileHashOrEmpty(ffprobeExecutablePath), "ffprobeExecutableSizeBytes": fileSizeOrZero(ffprobeExecutablePath)}
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, filepath.Dir(reportPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, reportPath); err != nil {
		return err
	}
	return os.WriteFile(reportPath, reportBytes, 0o600)
}

func readLatestArtifactReport(workspaceLayout workspace.WorkspaceLayout) (string, artifactReport, error) {
	entries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return "", artifactReport{}, err
	}
	latestPath := ""
	var latestModTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "build-report-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		candidatePath := filepath.Join(workspaceLayout.ArtifactsDirectory, entry.Name())
		if err := workspace.CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, candidatePath); err != nil {
			return "", artifactReport{}, err
		}
		info, err := entry.Info()
		if err != nil {
			return "", artifactReport{}, err
		}
		if latestPath == "" || info.ModTime().After(latestModTime) {
			latestPath = candidatePath
			latestModTime = info.ModTime()
		}
	}
	if latestPath == "" {
		return "", artifactReport{}, errors.New("no build report found")
	}
	reportBytes, err := os.ReadFile(latestPath)
	if err != nil {
		return "", artifactReport{}, err
	}
	var report artifactReport
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		return "", artifactReport{}, err
	}
	return latestPath, report, nil
}

func fileSizeOrZero(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func createFileHashOrEmpty(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
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

func downloadPolicyForHash(expectedSha256Hash string) download.FileConflictPolicy {
	if expectedSha256Hash == "" {
		return download.OverwriteFile
	}
	return download.ReuseIfHashMatches
}

func archiveExtensionFromUrl(downloadUrl string) string {
	lowerUrl := strings.ToLower(downloadUrl)
	switch {
	case strings.HasSuffix(lowerUrl, ".tar.xz"):
		return ".tar.xz"
	case strings.HasSuffix(lowerUrl, ".tar.zst"):
		return ".tar.zst"
	case strings.HasSuffix(lowerUrl, ".tar.bz2"):
		return ".tar.bz2"
	case strings.HasSuffix(lowerUrl, ".tar.gz") || strings.HasSuffix(lowerUrl, ".tgz"):
		return ".tar.gz"
	default:
		return ".tar.xz"
	}
}

const ffmpegReleaseSigningKeyUrl = "https://ffmpeg.org/ffmpeg-devel.asc"
const ffmpegReleaseSigningKeyFingerprint = "FCF986EA15E6E293A5644F10B4322F04D67658D8"

func verifyFfmpegDetachedSignature(signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return verifyDetachedSignatureWithPublicKey(signaturePath, archivePath, publicKeyPath, ffmpegReleaseSigningKeyFingerprint, "FFmpeg .asc", emitProgress)
}

const msys2InstallerSigningKeyUrl = "https://keyserver.ubuntu.com/pks/lookup?op=get&options=mr&search=0x0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"
const msys2InstallerPrimaryFingerprint = "0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"

func verifyMsys2DetachedSignature(signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return verifyDetachedSignatureWithPublicKey(signaturePath, archivePath, publicKeyPath, msys2InstallerPrimaryFingerprint, "MSYS2 .sig", emitProgress)
}

func verifyDetachedSignatureWithPublicKey(signaturePath string, archivePath string, publicKeyPath string, expectedPrimaryFingerprint string, signatureLabel string, emitProgress func(string, string)) error {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("could not read public signing key: %w", err)
	}
	keyRing, err := readOpenPgpKeyRing(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("could not read public signing key: %w", err)
	}
	if !keyRingContainsPrimaryFingerprint(keyRing, expectedPrimaryFingerprint) {
		return fmt.Errorf("downloaded signing key did not match the expected fingerprint %s", expectedPrimaryFingerprint)
	}
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("could not read archive for signature verification: %w", err)
	}
	signatureBytes, err := os.ReadFile(signaturePath)
	if err != nil {
		return fmt.Errorf("could not read detached signature: %w", err)
	}
	signer, err := checkDetachedSignature(keyRing, archiveBytes, signatureBytes)
	if err != nil {
		return fmt.Errorf("detached signature check failed: %w", err)
	}
	if signer == nil || !strings.EqualFold(hex.EncodeToString(signer.PrimaryKey.Fingerprint[:]), expectedPrimaryFingerprint) {
		return fmt.Errorf("signature was valid, but it was not made by the expected key %s", expectedPrimaryFingerprint)
	}
	if emitProgress != nil {
		emitProgress("info", signatureLabel+" verification passed without requiring system GPG.")
	}
	return nil
}

func readOpenPgpKeyRing(publicKeyBytes []byte) (openpgp.EntityList, error) {
	trimmedKey := bytes.TrimSpace(publicKeyBytes)
	if bytes.HasPrefix(trimmedKey, []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----")) {
		return openpgp.ReadArmoredKeyRing(bytes.NewReader(publicKeyBytes))
	}
	return openpgp.ReadKeyRing(bytes.NewReader(publicKeyBytes))
}

func checkDetachedSignature(keyRing openpgp.EntityList, archiveBytes []byte, signatureBytes []byte) (*openpgp.Entity, error) {
	trimmedSignature := bytes.TrimSpace(signatureBytes)
	if bytes.HasPrefix(trimmedSignature, []byte("-----BEGIN PGP SIGNATURE-----")) {
		return openpgp.CheckArmoredDetachedSignature(keyRing, bytes.NewReader(archiveBytes), bytes.NewReader(signatureBytes), nil)
	}
	return openpgp.CheckDetachedSignature(keyRing, bytes.NewReader(archiveBytes), bytes.NewReader(signatureBytes), nil)
}

func keyRingContainsPrimaryFingerprint(keyRing openpgp.EntityList, expectedPrimaryFingerprint string) bool {
	for _, entity := range keyRing {
		if entity == nil || entity.PrimaryKey == nil {
			continue
		}
		if strings.EqualFold(hex.EncodeToString(entity.PrimaryKey.Fingerprint[:]), expectedPrimaryFingerprint) {
			return true
		}
	}
	return false
}

func fileExists(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func trimForError(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 600 {
		return value[:600] + "..."
	}
	return value
}

func archiveFormatFromUrl(downloadUrl string) extraction.ArchiveFormat {
	lowerUrl := strings.ToLower(downloadUrl)
	switch {
	case strings.HasSuffix(lowerUrl, ".tar.bz2"):
		return extraction.TarBz2
	case strings.HasSuffix(lowerUrl, ".tar.gz") || strings.HasSuffix(lowerUrl, ".tgz"):
		return extraction.ArchiveFormatTarGz
	case strings.HasSuffix(lowerUrl, ".tar.zst"):
		return extraction.TarZst
	case strings.HasSuffix(lowerUrl, ".tar.xz"):
		return extraction.TarXz
	case strings.HasSuffix(lowerUrl, ".tar"):
		return extraction.ArchiveFormatTar
	default:
		return extraction.TarXz
	}
}
