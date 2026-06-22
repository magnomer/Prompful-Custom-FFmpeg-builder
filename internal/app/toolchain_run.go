package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"customffmpegbuilder/internal/audit"
	"customffmpegbuilder/internal/consent"
	"customffmpegbuilder/internal/download"
	"customffmpegbuilder/internal/execution"
	"customffmpegbuilder/internal/extraction"
	"customffmpegbuilder/internal/planning"
	"customffmpegbuilder/internal/scripting"
	"customffmpegbuilder/internal/workspace"
)

const msys2InstallerSigningKeyUrl = "https://keyserver.ubuntu.com/pks/lookup?op=get&options=mr&search=0x0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"
const msys2InstallerPrimaryFingerprint = "0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"

func verifyMsys2DetachedSignature(signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return verifyDetachedSignatureWithPublicKey(signaturePath, archivePath, publicKeyPath, msys2InstallerPrimaryFingerprint, "MSYS2 .sig", emitProgress)
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
		app.emitLocalizedFailure("run.failure.createWorkspaceDirectories", "Could not create workspace directories", err)
		return
	}
	auditWriter, err := audit.NewWriter(workspaceLayout.LogsDirectory, runId)
	if err != nil {
		app.emitLocalizedFailure("run.failure.createAuditLog", "Could not create audit log", err)
		return
	}
	emitProgress := app.createAuditedProgressFunc(auditWriter, plan.ActionName, plan.PlanHash)
	_ = auditWriter.WriteEvent("action-started", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 preparation started.")
	emitProgress("info", "Approved private MSYS2 preparation started. Run: "+runId)

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-approved"+archiveExtensionFromUrl(plan.Msys2ArchiveUrl))
	downloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 archive", DownloadUrl: plan.Msys2ArchiveUrl, ExpectedSha256Hash: plan.Msys2ArchiveSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 500_000_000, FileConflictPolicy: downloadPolicyForHash(plan.Msys2ArchiveSha256Hash)}
	if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, downloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.msys2ArchiveDownload", "MSYS2 archive download failed", err)
		return
	}
	if plan.Msys2ArchiveSignatureUrl != "" {
		signaturePath := archivePath + ".sig"
		signatureDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 signature", DownloadUrl: plan.Msys2ArchiveSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
		if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, signatureDownloadPlan, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitLocalizedFailure("run.failure.msys2SignatureDownload", "MSYS2 signature download failed", err)
			return
		}
		publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-installer-signing-key.asc")
		publicKeyDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 installer signing key", DownloadUrl: msys2InstallerSigningKeyUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"keyserver.ubuntu.com"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 1_000_000, FileConflictPolicy: download.OverwriteFile}
		if err := download.DownloadMsys2WithConsent(ctx, userMsys2DownloadConsent, publicKeyDownloadPlan, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitLocalizedFailure("run.failure.msys2SigningKeyDownload", "MSYS2 signing key download failed", err)
			return
		}
		if err := verifyMsys2DetachedSignature(signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
			_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
			app.emitLocalizedFailure("run.failure.msys2SignatureVerification", "MSYS2 signature verification failed", err)
			return
		}
	}
	if err := app.prepareFreshToolchainDirectory(plan.WorkspaceDirectory, plan.Msys2RootDirectory, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.msys2Cleanup", "MSYS2 private toolchain folder cleanup failed", err)
		return
	}
	extractPlan := extraction.ExtractPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, ArchiveFormat: archiveFormatFromUrl(plan.Msys2ArchiveUrl), ExtractDestinationPolicy: extraction.RequireNewDirectory, ExtractedFileModePolicy: extraction.PreserveSafeExecutableBits, MaximumFileCount: 250000, MaximumExtractedByteCount: 10_000_000_000, MaximumSingleFileByteCount: 2_000_000_000}
	if err := extraction.ExtractArchiveWithConsent(ctx, userArchiveExtractionConsent, extractPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.msys2Extraction", "MSYS2 archive extraction failed", err)
		return
	}
	if err := normalizeMsys2RootDirectory(plan.Msys2RootDirectory, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.msys2Layout", "MSYS2 archive layout check failed", err)
		return
	}
	if err := app.installPacmanPackages(ctx, plan, userPacmanPackageInstallConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.msys2PackageInstall", "MSYS2 package installation failed", err)
		return
	}
	_ = auditWriter.WriteEvent("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 environment is ready.")
	emitProgress("info", localize("run.log.toolchainReady", nil))
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
	emitProgress("warn", localize("run.log.previousMsys2Exists", nil))
	stopPrivateMsys2BackgroundAgents(msys2RootDirectory, emitProgress)
	if err := removeAllWithRetry(msys2RootDirectory); err != nil {
		return fmt.Errorf("could not remove previous private MSYS2 folder: %w", err)
	}
	emitProgress("info", localize("run.log.previousMsys2Removed", nil))
	return nil
}

func normalizeMsys2RootDirectory(msys2RootDirectory string, emitProgress func(string, string)) error {
	if fileExists(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		emitProgress("info", localize("run.log.msys2LayoutConfirmed", nil))
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
	emitProgress("info", localize("run.log.msys2TopFolderMoving", nil))
	if err := moveDirectoryContents(nestedRootDirectory, msys2RootDirectory); err != nil {
		return err
	}
	if err := os.Remove(nestedRootDirectory); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not remove empty extracted MSYS2 wrapper folder: %w", err)
	}
	if !fileExists(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		return errors.New("MSYS2 archive layout normalization did not produce usr/bin/bash.exe")
	}
	emitProgress("info", localize("run.log.msys2LayoutNormalized", nil))
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

func (app *App) cleanupFailedToolchainRun(plan planning.ToolchainPreparationPlan, workspaceLayout workspace.WorkspaceLayout, runId string) {
	app.emitLog("warn", localize("run.log.cleaningToolchainPartial", nil))
	cleanupTargets := []string{
		plan.Msys2RootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "pacman-install-"+plan.PlanHash+".sh"),
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
			app.emitLog("warn", localize("run.log.cleanupInspectFailed", map[string]string{"message": err.Error()}))
			continue
		}
		if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, targetPath); err != nil {
			app.emitLog("warn", localize("run.log.cleanupOutsideWorkspace", map[string]string{"message": err.Error()}))
			continue
		}
		if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, targetPath); err != nil {
			app.emitLog("warn", localize("run.log.cleanupUnsafe", map[string]string{"message": err.Error()}))
			continue
		}
		stopPrivateMsys2BackgroundAgents(targetPath, app.emitLog)
		if err := removeAllWithRetry(targetPath); err != nil {
			app.emitLog("warn", localize("run.log.cleanupRemoveFailed", map[string]string{"message": err.Error()}))
			continue
		}
		app.emitLog("info", localize("run.log.cleanupRemoved", map[string]string{"path": targetPath}))
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
		emitProgress("info", localize("run.log.stoppedMsys2Agents", nil))
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
