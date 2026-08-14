package program

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"

	"promptfulcustomffmpegbuilder/internal/audit"
	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/download"
	"promptfulcustomffmpegbuilder/internal/execution"
	"promptfulcustomffmpegbuilder/internal/extraction"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/scripting"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

const LSignatureFFmpegUrl = "https://ffmpeg.org/ffmpeg-devel.asc"
const LSignatureFFmpegFingerprint = "FCF986EA15E6E293A5644F10B4322F04D67658D8"

func LSignatureFFmpegVerify(signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return LSignatureDetachedVerify(signaturePath, archivePath, publicKeyPath, LSignatureFFmpegFingerprint, "FFmpeg .asc", emitProgress)
}

func (program *LProgram) LFFmpegCompile(LContext context.Context, LRunId string, plan planning.LPlanFFmpeg, userLConsentFFmpeg consent.LConsentFFmpeg, userLConsentArchive consent.LArchiveConsentState, userLibraryPackageInstallLConsent consent.LConsentPacman, userExternalLConsentCommand consent.LConsentCommand) {
	actionSucceeded := false
	copyFailed := false
	stalledHalt := false
	workspaceLayout := workspace.LLayoutVersionedGet(plan.WorkspaceDirectory, planning.LVersionArchiveParse(plan.FfmpegSourceArchiveUrl))
	sourceRootDirectory := filepath.Join(workspaceLayout.SourcesDirectory, "ffmpeg-"+LRunId)
	ffmpegSourceDirectory := ""
	defer func() {
		if actionSucceeded {
			program.LActionApprovedFinish("completed")
			return
		}
		if copyFailed && ffmpegSourceDirectory != "" {
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.copyFailedFilesKept", nil))
			program.LLogEmit("warn", LLocaleTextGetInternal("run.log.copyFailedFilesLocation", map[string]string{"path": ffmpegSourceDirectory}))
			program.LActionApprovedFinish("failed")
			return
		}
		program.LLogConfigurationSave(ffmpegSourceDirectory, workspaceLayout.WorkspaceDirectory)
		if stalledHalt {
			// A transient-network stall is retryable and the whole pipeline resumes
			// from cache, so preserve the partial build instead of cleaning it. The
			// "stalled" status was already emitted when the stall was classified.
			program.LActionApprovedFinish("stalled")
			return
		}
		program.LFFmpegFailureClean(plan, workspaceLayout, sourceRootDirectory)
		program.LActionApprovedFinish("failed")
	}()
	program.LStatusEmit("building-ffmpeg")
	if err := workspace.LWorkspaceFolderCreate(workspaceLayout); err != nil {
		program.LErrorLocalizedEmit("run.failure.createWorkspaceDirectories", "Could not create workspace directories", err)
		return
	}
	auditWriter, err := audit.LAuditWriterCreate(workspaceLayout.LogsDirectory, LRunId)
	if err != nil {
		program.LErrorLocalizedEmit("run.failure.createAuditLog", "Could not create audit log", err)
		return
	}
	emitProgress := program.LAuditProgressCreate(auditWriter, plan.ActionName, plan.PlanHash)
	_ = auditWriter.LAuditEventWrite("action-started", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build started.")
	emitProgress("info", LLocaleTextGetInternal("run.log.ffmpegStarted", map[string]string{"runId": LRunId}))

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+LArchiveExtensionResolve(plan.FfmpegSourceArchiveUrl))
	downloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg", DownloadUrl: plan.FfmpegSourceArchiveUrl, ExpectedSha256Hash: plan.FfmpegSourceSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 200_000_000, LPolicyFile: LPolicyHashResolve(plan.FfmpegSourceSha256Hash)}
	if err := download.LDownloadFFmpegRun(LContext, userLConsentFFmpeg, downloadPlan, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSourceDownload", "FFmpeg source download failed", err)
		return
	}
	signaturePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+LArchiveExtensionResolve(plan.FfmpegSourceArchiveUrl)+".asc")
	signatureDownloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg signature", DownloadUrl: plan.FfmpegSourceSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, LPolicyFile: download.LPolicyFileOverwrite}
	if err := download.LDownloadFFmpegRun(LContext, userLConsentFFmpeg, signatureDownloadPlan, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSignatureDownload", "FFmpeg source signature download failed", err)
		return
	}
	publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-devel.asc")
	publicKeyDownloadPlan := download.LDownloadPlanState{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg release signing key", DownloadUrl: LSignatureFFmpegUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 100_000, LPolicyFile: download.LPolicyFileOverwrite}
	if err := download.LDownloadFFmpegRun(LContext, userLConsentFFmpeg, publicKeyDownloadPlan, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSigningKeyDownload", "FFmpeg signing key download failed", err)
		return
	}
	if err := LSignatureFFmpegVerify(signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSignatureVerification", "FFmpeg source signature verification failed", err)
		return
	}
	extractPlan := extraction.LPlanExtraction{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: sourceRootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, LArchiveKind: LArchiveFormatResolve(plan.FfmpegSourceArchiveUrl), LPolicyExtraction: extraction.LNewDirectoryPolicy, LPolicyFilemode: extraction.LExecutablePreservePolicy, MaximumFileCount: 50000, MaximumExtractedByteCount: 2_000_000_000, MaximumSingleFileByteCount: 500_000_000}
	if err := extraction.LArchiveConsentExtract(LContext, userLConsentArchive, extractPlan, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSourceExtraction", "FFmpeg source extraction failed", err)
		return
	}
	ffmpegSourceDirectory, err = LDirectoryChildFind(sourceRootDirectory)
	if err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegSourceDirectoryMissing", "Could not locate extracted FFmpeg source directory", err)
		return
	}
	if err := program.LLibraryVersionValidate(plan, ffmpegSourceDirectory, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.libraryVersionIncompatible", "A prepared library version is incompatible with the selected FFmpeg release", err)
		return
	}
	if err := program.LLibraryPackageInstall(LContext, plan, userLibraryPackageInstallLConsent, auditWriter, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegLibraryPackageInstall", "FFmpeg library package installation failed", err)
		return
	}
	if err := program.LLibraryNonnativePrepare(LContext, plan, userLConsentFFmpeg, userLConsentArchive, userLibraryPackageInstallLConsent, userExternalLConsentCommand, auditWriter, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.libraryPreparation", "Non-Native library preparation failed", err)
		return
	}
	if err := program.LFFmpegConfigureRun(LContext, plan, ffmpegSourceDirectory, userExternalLConsentCommand, auditWriter, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegConfigure", "FFmpeg configure failed", err)
		return
	}
	if err := program.LFFmpegMakeRun(LContext, plan, ffmpegSourceDirectory, userExternalLConsentCommand, auditWriter, emitProgress); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.ffmpegBuild", "FFmpeg build failed", err)
		return
	}
	if err := LArtifactFFmpegCopy(ffmpegSourceDirectory, workspaceLayout, plan, emitProgress); err != nil {
		copyFailed = true
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.copyArtifacts", "Could not copy FFmpeg artifacts", err)
		return
	}
	if err := LReportArtifactWrite(workspaceLayout, LRunId, plan); err != nil {
		stalledHalt = program.LActionFailureEmit(auditWriter, plan, "run.failure.writeArtifactReport", "Could not write artifact report", err)
		return
	}
	_ = auditWriter.LAuditEventWrite("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build completed.")
	emitProgress("info", LLocaleTextGetInternal("run.log.ffmpegCompleted", nil))
	actionSucceeded = true
}

// LActionFailureEmit records the terminal outcome of a failed FFmpeg build stage.
// A transient-network stall (execution.LErrorNetworkStalled) is halted in the
// retryable "stalled" state: a final warn-level audit event carries the tried
// addresses (localized run.log.downloadStalled) and the live status becomes
// "stalled". Every other error is a genuine failure: an error-level "action-failed"
// event and the localized failure status. Returns true when the stage stalled so
// the caller can preserve the partial build for a later resume.
func (program *LProgram) LActionFailureEmit(auditWriter *audit.LAuditWriter, plan planning.LPlanFFmpeg, messageKey string, fallback string, err error) bool {
	var stalled *execution.LErrorNetworkStalled
	if errors.As(err, &stalled) {
		message := LLocaleTextGetInternal("run.log.downloadStalled", map[string]string{"addresses": strings.Join(stalled.LNetworkAddresses, ", ")})
		_ = auditWriter.LAuditEventWrite("action-stalled", plan.ActionName, plan.PlanHash, "warn", message)
		program.LLogEmit("warn", message)
		program.LStatusEmit("stalled")
		return true
	}
	_ = auditWriter.LAuditEventWrite("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
	program.LErrorLocalizedEmit(messageKey, fallback, err)
	return false
}

func (program *LProgram) LLogConfigurationSave(ffmpegSourceDirectory string, workspaceDirectory string) {
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
		program.LLogEmit("warn", LLocaleTextGetInternal("run.log.configReadFailed", map[string]string{"message": err.Error()}))
		return
	}
	if err := os.WriteFile(destPath, data, 0o600); err != nil {
		program.LLogEmit("warn", LLocaleTextGetInternal("run.log.configSaveFailed", map[string]string{"message": err.Error()}))
		return
	}
	program.LLogEmit("info", LLocaleTextGetInternal("run.log.configSaved", map[string]string{"path": destPath}))
}

func (program *LProgram) LFFmpegFailureClean(plan planning.LPlanFFmpeg, workspaceLayout workspace.LWorkspaceLayout, sourceRootDirectory string) {
	program.LLogEmit("warn", LLocaleTextGetInternal("run.log.cleaningFfmpegPartial", nil))
	cleanupTargets := []string{
		sourceRootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-library-packages-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-configure-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-make-"+plan.PlanHash+".sh"),
	}
	for _, preparation := range plan.LPreparationCatalog {
		cleanupTargets = append(cleanupTargets,
			filepath.Join(workspaceLayout.BuildDirectory, "prep", preparation.LibraryId+"-"+plan.PlanHash),
			filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
			filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-builddeps-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		)
	}
	program.LWorkspaceTargetsClean(plan.WorkspaceDirectory, cleanupTargets)
}

func (program *LProgram) LLibraryPackageInstall(LContext context.Context, plan planning.LPlanFFmpeg, userLibraryPackageInstallLConsent consent.LConsentPacman, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	if len(plan.RequiredMsys2PackageNames) == 0 {
		emitProgress("info", "No extra MSYS2 library packages are required by the selected FFmpeg libraries.")
		return nil
	}
	if err := consent.LConsentCheck(userLibraryPackageInstallLConsent.LConsent, consent.LConsentKindPacman, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	scriptLines, err := scripting.LPackageScriptCreate(plan.RequiredMsys2PackageNames)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-library-packages-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: plan.Msys2RootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, LScriptKind: execution.LPacmanInstallScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLAuditDirectoryGet: auditWriter.LAuditDirectoryGet()}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg library package installation script.")
	return execution.LCommandPacmanRun(LContext, userLibraryPackageInstallLConsent, commandPlan, emitProgress)
}

func LPathPkgconfigResolve(plan planning.LPlanFFmpeg) string {
	profileDirectoryName := strings.ToLower(plan.WindowsShellProfileName)
	if profileDirectoryName == "" {
		profileDirectoryName = "ucrt64"
	}
	msys2Prefix := "/" + profileDirectoryName
	// Privately-installed libraries (e.g. libtls) live in their own per-library prefix to
	// avoid archive-name collisions in the shared prefix; their pkgconfig dirs go first so
	// pkg-config resolves them (and their isolated libssl/libcrypto) ahead of the shared
	// prefix's same-named modules.
	paths := append([]string{}, LPackagePathList(plan)...)
	paths = append(paths,
		msys2Prefix+"/lib/pkgconfig",
		msys2Prefix+"/share/pkgconfig",
		"/usr/lib/pkgconfig",
		"/usr/share/pkgconfig",
	)
	return strings.Join(paths, ":")
}

// LPackagePathList returns the unix pkgconfig directories of every privately-installed
// library in the plan (see planning.LLibraryPreparation.PrivatePrefixInstall), in plan order.
func LPackagePathList(plan planning.LPlanFFmpeg) []string {
	profileDirectoryName := strings.ToLower(plan.WindowsShellProfileName)
	if profileDirectoryName == "" {
		profileDirectoryName = "ucrt64"
	}
	msys2Prefix := "/" + profileDirectoryName
	dirs := []string{}
	for _, preparation := range plan.LPreparationCatalog {
		if preparation.PrivatePrefixInstall && preparation.PkgConfigName != "" {
			dirs = append(dirs, scripting.LPrivateDirectoryGet(msys2Prefix, preparation.PkgConfigName))
		}
	}
	return dirs
}

func (program *LProgram) LFFmpegConfigureRun(LContext context.Context, plan planning.LPlanFFmpeg, ffmpegSourceDirectory string, userExternalLConsentCommand consent.LConsentCommand, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	scriptLines, err := scripting.LConfigureScriptCreate(plan.ConfigureFlags, LPackagePathList(plan), planning.LVersionArchiveParse(plan.FfmpegSourceArchiveUrl))
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-configure-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: ffmpegSourceDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{"PKG_CONFIG_PATH": LPathPkgconfigResolve(plan)}, AllowedExecutableBasenames: []string{"bash.exe"}, LScriptKind: execution.LScriptFFmpegConfigure, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLAuditDirectoryGet: auditWriter.LAuditDirectoryGet()}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg configure script.")
	return execution.LCommandConsentRun(LContext, userExternalLConsentCommand, commandPlan, emitProgress)
}

func (program *LProgram) LFFmpegMakeRun(LContext context.Context, plan planning.LPlanFFmpeg, ffmpegSourceDirectory string, userExternalLConsentCommand consent.LConsentCommand, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	scriptLines, err := scripting.LMakeLinesCreate(plan.ParallelJobCount)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{WorkspaceDirectory: plan.WorkspaceDirectory, ScriptFilePath: filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-make-"+plan.PlanHash+".sh"), ScriptLines: scriptLines})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ExecutablePath: filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"), ArgumentValues: []string{filepath.ToSlash(scriptFile.ScriptFilePath)}, WorkingDirectory: ffmpegSourceDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, Msys2RootDirectory: plan.Msys2RootDirectory, WindowsShellProfileName: plan.WindowsShellProfileName, EnvironmentVariables: map[string]string{}, AllowedExecutableBasenames: []string{"bash.exe"}, LScriptKind: execution.LFFmpegMakeScript, ApprovedScriptFilePath: scriptFile.ScriptFilePath, ApprovedScriptSha256Hash: scriptFile.ScriptSha256Hash, RunLAuditDirectoryGet: auditWriter.LAuditDirectoryGet()}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved FFmpeg make script.")
	return execution.LCommandConsentRun(LContext, userExternalLConsentCommand, commandPlan, emitProgress)
}
