package program

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/audit"
	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/download"
	"promptfulcustomffmpegbuilder/internal/hostexec"
	"promptfulcustomffmpegbuilder/internal/execution"
	"promptfulcustomffmpegbuilder/internal/extraction"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/scripting"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

const LMsysKeyUrl = "https://keyserver.ubuntu.com/pks/lookup?op=get&options=mr&search=0x0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"
const LSignatureMsysFingerprint = "0EBF782C5D53F7E5FB02A66746BD761F7A49B0EC"

func LSignatureMsysVerify(LContext context.Context, signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return LSignatureDetachedVerify(LContext, signaturePath, archivePath, publicKeyPath, LSignatureMsysFingerprint, "MSYS2 .sig", emitProgress)
}

func (program *LProgram) LToolchainPrepare(LContext context.Context, LRunId string, reviewSessionId string, plan planning.LPlanToolchain, userLConsentMsys consent.LConsentMsys, userLConsentArchive consent.LArchiveConsentState, userPacmanPackageInstallLConsent consent.LConsentPacman) {
	actionSucceeded := false
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	defer func() {
		if actionSucceeded {
			program.LActionApprovedFinish("completed")
			return
		}
		program.LToolchainFailureClean(plan, workspaceLayout)
		// A user-requested cancellation finishes "cancelled", not "failed", so the
		// stop is distinguishable from a genuine preparation failure.
		if program.LActionCancelledCheck() {
			program.LActionApprovedFinish("cancelled")
			return
		}
		program.LActionApprovedFinish("failed")
	}()
	program.LStatusEmit("preparing-toolchain")
	if err := workspace.LWorkspaceFolderCreate(workspaceLayout); err != nil {
		program.LErrorLocalizedEmit("run.failure.createWorkspaceDirectories", "Could not create workspace directories", err)
		return
	}
	auditWriter, err := audit.LAuditWriterCreate(workspaceLayout.LogsDirectory, LRunId, reviewSessionId)
	if err != nil {
		program.LErrorLocalizedEmit("run.failure.createAuditLog", "Could not create audit log", err)
		return
	}
	emitProgress := program.LAuditProgressCreate(auditWriter, plan.ActionName, plan.PlanHash)
	_ = auditWriter.LAuditEventWrite("action-started", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 preparation started.")
	emitProgress("info", "Approved private MSYS2 preparation started. Run: "+LRunId)

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-approved"+LArchiveExtensionResolve(plan.Msys2ArchiveUrl))
	downloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 archive", DownloadUrl: plan.Msys2ArchiveUrl, ExpectedSha256Hash: plan.Msys2ArchiveSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 500_000_000, LPolicyFile: LPolicyHashResolve(plan.Msys2ArchiveSha256Hash)}
	if err := download.LDownloadMsysRun(LContext, userLConsentMsys, downloadPlan, emitProgress); err != nil {
		program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2ArchiveDownload", "MSYS2 archive download failed", err)
		return
	}
	if plan.Msys2ArchiveSignatureUrl != "" {
		signaturePath := archivePath + ".sig"
		signatureDownloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 signature", DownloadUrl: plan.Msys2ArchiveSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"github.com", "repo.msys2.org", "mirror.msys2.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, LPolicyFile: download.LPolicyFileOverwrite}
		if err := download.LDownloadMsysRun(LContext, userLConsentMsys, signatureDownloadPlan, emitProgress); err != nil {
			program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2SignatureDownload", "MSYS2 signature download failed", err)
			return
		}
		publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "msys2-installer-signing-key.asc")
		publicKeyDownloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "MSYS2 installer signing key", DownloadUrl: LMsysKeyUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"keyserver.ubuntu.com"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 1_000_000, LPolicyFile: download.LPolicyFileOverwrite}
		if err := download.LDownloadMsysRun(LContext, userLConsentMsys, publicKeyDownloadPlan, emitProgress); err != nil {
			program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2SigningKeyDownload", "MSYS2 signing key download failed", err)
			return
		}
		if err := LSignatureMsysVerify(LContext, signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
			program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2SignatureVerification", "MSYS2 signature verification failed", err)
			return
		}
	}
	if err := program.LToolchainFreshPrepare(LContext, plan.WorkspaceDirectory, plan.Msys2RootDirectory, emitProgress); err != nil {
		program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2Cleanup", "MSYS2 private toolchain folder cleanup failed", err)
		return
	}
	extractPlan := extraction.LPlanExtraction{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, LArchiveKind: LArchiveFormatResolve(plan.Msys2ArchiveUrl), LPolicyExtraction: extraction.LPolicyExtractionNew, LPolicyFilemode: extraction.LPolicyFilemodeExecutable, MaximumFileCount: 250000, MaximumExtractedByteCount: 10_000_000_000, MaximumSingleFileByteCount: 2_000_000_000}
	if err := extraction.LArchiveConsentExtract(LContext, userLConsentArchive, extractPlan, emitProgress); err != nil {
		program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2Extraction", "MSYS2 archive extraction failed", err)
		return
	}
	if err := LMsysRootNormalize(LContext, plan.Msys2RootDirectory, emitProgress); err != nil {
		program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2Layout", "MSYS2 archive layout check failed", err)
		return
	}
	if err := program.LPacmanPackageInstall(LContext, plan, userPacmanPackageInstallLConsent, auditWriter, emitProgress); err != nil {
		program.LToolchainFailureEmit(auditWriter, plan, "run.failure.msys2PackageInstall", "MSYS2 package installation failed", err)
		return
	}
	if err := LManifestToolchainWrite(plan); err != nil {
		emitProgress("warn", LLocaleTextGetInternal("run.log.toolchainManifestFailed", nil))
	}
	_ = auditWriter.LAuditEventWrite("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved private MSYS2 environment is ready.")
	emitProgress("info", LLocaleTextGetInternal("run.log.toolchainReady", nil))
	actionSucceeded = true
}

// LToolchainFailureEmit records a failed toolchain stage. A user-requested
// cancellation is written as an "action-cancelled" event; every other error is
// written as "action-failed" with its localized failure log.
func (program *LProgram) LToolchainFailureEmit(auditWriter *audit.LAuditWriter, plan planning.LPlanToolchain, messageKey string, fallback string, err error) {
	if program.LActionCancelledCheck() {
		lActionCancelledEmit(program, auditWriter, plan.ActionName, plan.PlanHash)
		return
	}
	_ = auditWriter.LAuditEventWrite("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
	program.LErrorLocalizedEmit(messageKey, fallback, err)
}

func (program *LProgram) LToolchainFreshPrepare(LContext context.Context, workspaceDirectory string, msys2RootDirectory string, emitProgress func(string, string)) error {
	if msys2RootDirectory == "" {
		return errors.New("private MSYS2 toolchain folder is empty")
	}
	if err := LContext.Err(); err != nil {
		return err
	}
	if err := workspace.LPathWorkspaceCheck(workspaceDirectory, msys2RootDirectory); err != nil {
		return err
	}
	if _, err := os.Lstat(msys2RootDirectory); errors.Is(err, os.ErrNotExist) {
		return nil
	} else if err != nil {
		return fmt.Errorf("could not inspect existing private MSYS2 folder: %w", err)
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, msys2RootDirectory); err != nil {
		return fmt.Errorf("existing private MSYS2 folder is unsafe to clean: %w", err)
	}
	emitProgress("warn", LLocaleTextGetInternal("run.log.previousMsys2Exists", nil))
	LMsysProcessStop(msys2RootDirectory, emitProgress)
	if err := LPathRetryRemove(msys2RootDirectory); err != nil {
		return fmt.Errorf("could not remove previous private MSYS2 folder: %w", err)
	}
	emitProgress("info", LLocaleTextGetInternal("run.log.previousMsys2Removed", nil))
	return nil
}

func LMsysRootNormalize(LContext context.Context, msys2RootDirectory string, emitProgress func(string, string)) error {
	if LFileExistCheck(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		emitProgress("info", LLocaleTextGetInternal("run.log.msys2LayoutConfirmed", nil))
		return nil
	}
	if err := LContext.Err(); err != nil {
		return err
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
		if LFileExistCheck(filepath.Join(candidateDirectory, "usr", "bin", "bash.exe")) {
			matchingDirectories = append(matchingDirectories, candidateDirectory)
		}
	}

	if len(matchingDirectories) == 0 {
		return errors.New("could not find usr/bin/bash.exe after extraction; check that the MSYS2 URL points to a base tar archive, not an installer, signature, or HTML page")
	}
	if len(matchingDirectories) > 1 {
		return fmt.Errorf("found multiple possible MSYS2 roots after extraction: %s", strings.Join(matchingDirectories, ", "))
	}

	if err := LContext.Err(); err != nil {
		return err
	}
	nestedRootDirectory := matchingDirectories[0]
	emitProgress("info", LLocaleTextGetInternal("run.log.msys2TopFolderMoving", nil))
	if err := LDirectoryContentMove(nestedRootDirectory, msys2RootDirectory); err != nil {
		return err
	}
	if err := os.Remove(nestedRootDirectory); err != nil && !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("could not remove empty extracted MSYS2 wrapper folder: %w", err)
	}
	if !LFileExistCheck(filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")) {
		return errors.New("MSYS2 archive layout normalization did not produce usr/bin/bash.exe")
	}
	emitProgress("info", LLocaleTextGetInternal("run.log.msys2LayoutNormalized", nil))
	return nil
}

func LDirectoryContentMove(sourceDirectory string, destinationDirectory string) error {
	entries, err := os.ReadDir(sourceDirectory)
	if err != nil {
		return fmt.Errorf("could not read extracted MSYS2 wrapper folder: %w", err)
	}
	for _, entry := range entries {
		sourcePath := filepath.Join(sourceDirectory, entry.Name())
		destinationPath := filepath.Join(destinationDirectory, entry.Name())
		if LFileExistCheck(destinationPath) {
			return fmt.Errorf("cannot move extracted MSYS2 entry because destination already exists: %s", destinationPath)
		}
		if err := os.Rename(sourcePath, destinationPath); err != nil {
			return fmt.Errorf("could not move extracted MSYS2 entry %s: %w", entry.Name(), err)
		}
	}
	return nil
}

func (program *LProgram) LPacmanPackageInstall(LContext context.Context, plan planning.LPlanToolchain, userPacmanPackageInstallLConsent consent.LConsentPacman, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	if err := consent.LConsentCheck(userPacmanPackageInstallLConsent.LConsent, consent.LConsentKindPacman, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	scriptLines, err := scripting.LScriptPacmanBuild(plan.Msys2PackageNames, plan.WindowsShellProfileName)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "pacman-install-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, LScriptKind: execution.LPacmanInstallScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLAuditDirectoryGet: auditWriter.LAuditDirectoryGet()}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved pacman package installation script.")
	return execution.LCommandPacmanRun(LContext, userPacmanPackageInstallLConsent, commandPlan, emitProgress)
}

func (program *LProgram) LToolchainFailureClean(plan planning.LPlanToolchain, workspaceLayout workspace.LWorkspaceLayout) {
	program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleaningToolchainPartial", nil))
	cleanupTargets := []string{
		plan.Msys2RootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "pacman-install-"+plan.PlanHash+".sh"),
	}
	program.LWorkspaceTargetsClean(plan.WorkspaceDirectory, cleanupTargets)
}

func (program *LProgram) LWorkspaceTargetsClean(workspaceDirectory string, targets []string) {
	for _, targetPath := range targets {
		if targetPath == "" {
			continue
		}
		if _, err := os.Lstat(targetPath); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleanupInspectFailed", map[string]string{"message": err.Error()}))
			continue
		}
		if err := workspace.LPathWorkspaceCheck(workspaceDirectory, targetPath); err != nil {
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleanupOutsideWorkspace", map[string]string{"message": err.Error()}))
			continue
		}
		if err := workspace.LPathRealCheck(workspaceDirectory, targetPath); err != nil {
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleanupUnsafe", map[string]string{"message": err.Error()}))
			continue
		}
		LMsysProcessStop(targetPath, program.LLogEmit)
		if err := LPathRetryRemove(targetPath); err != nil {
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleanupRemoveFailed", map[string]string{"message": err.Error()}))
			continue
		}
		program.LLogEmit("info", LLocaleTextGetInternal("run.log.cleanupRemoved", map[string]string{"path": targetPath}))
	}
}

func LMsysProcessStop(msys2RootDirectory string, emitProgress func(string, string)) {
	bashPath := filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")
	if !LFileExistCheck(bashPath) {
		return
	}
	command := exec.Command(bashPath, "-lc", "GNUPGHOME=/etc/pacman.d/gnupg gpgconf --kill gpg-agent >/dev/null 2>&1 || true; GNUPGHOME=/etc/pacman.d/gnupg gpgconf --kill all >/dev/null 2>&1 || true")
	command.Dir = msys2RootDirectory
	command.Env = append(os.Environ(), "MSYSTEM=UCRT64", "MSYS2_PATH_TYPE=inherit", "CHERE_INVOKING=1")
	hostexec.LCommandWindowHide(command)
	if err := command.Run(); err == nil && emitProgress != nil {
		emitProgress("info", LLocaleTextGetInternal("run.log.stoppedMsys2Agents", nil))
	}
}

func LPathRetryRemove(targetPath string) error {
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

func LArchiveExtensionResolve(downloadUrl string) string {
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

func LArchiveFormatResolve(downloadUrl string) extraction.LArchiveKind {
	lowerUrl := strings.ToLower(downloadUrl)
	switch {
	case strings.HasSuffix(lowerUrl, ".tar.bz2"):
		return extraction.LArchiveTarBzip
	case strings.HasSuffix(lowerUrl, ".tar.gz") || strings.HasSuffix(lowerUrl, ".tgz"):
		return extraction.LArchiveTarGz
	case strings.HasSuffix(lowerUrl, ".tar.zst"):
		return extraction.LArchiveTarZst
	case strings.HasSuffix(lowerUrl, ".tar.xz"):
		return extraction.LArchiveTarXz
	case strings.HasSuffix(lowerUrl, ".tar"):
		return extraction.LArchiveTar
	default:
		return extraction.LArchiveTarXz
	}
}
