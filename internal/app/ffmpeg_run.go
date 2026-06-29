package app

import (
	"context"
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

const ffmpegReleaseSigningKeyUrl = "https://ffmpeg.org/ffmpeg-devel.asc"
const ffmpegReleaseSigningKeyFingerprint = "FCF986EA15E6E293A5644F10B4322F04D67658D8"

func verifyFfmpegDetachedSignature(signaturePath string, archivePath string, publicKeyPath string, emitProgress func(string, string)) error {
	return verifyDetachedSignatureWithPublicKey(signaturePath, archivePath, publicKeyPath, ffmpegReleaseSigningKeyFingerprint, "FFmpeg .asc", emitProgress)
}

func (app *App) buildFfmpeg(ctx context.Context, runId string, plan planning.FfmpegBuildPlan, userFfmpegSourceDownloadConsent consent.FfmpegSourceDownloadConsent, userArchiveExtractionConsent consent.ArchiveExtractionConsent, userLibraryPackageInstallConsent consent.PacmanInstallConsent, userExternalCommandExecutionConsent consent.CommandExecutionConsent) {
	actionSucceeded := false
	copyFailed := false
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	sourceRootDirectory := filepath.Join(workspaceLayout.SourcesDirectory, "ffmpeg-"+runId)
	ffmpegSourceDirectory := ""
	defer func() {
		if actionSucceeded {
			app.finishApprovedAction("completed")
			return
		}
		if copyFailed && ffmpegSourceDirectory != "" {
			app.emitLog("warn", localize("run.log.copyFailedFilesKept", nil))
			app.emitLog("warn", localize("run.log.copyFailedFilesLocation", map[string]string{"path": ffmpegSourceDirectory}))
			app.finishApprovedAction("failed")
			return
		}
		app.saveConfigLog(ffmpegSourceDirectory, workspaceLayout.WorkspaceDirectory)
		app.cleanupFailedFfmpegRun(plan, workspaceLayout, sourceRootDirectory, runId)
		app.finishApprovedAction("failed")
	}()
	app.emitStatus("building-ffmpeg")
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
	_ = auditWriter.WriteEvent("action-started", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build started.")
	emitProgress("info", localize("run.log.ffmpegStarted", map[string]string{"runId": runId}))

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+archiveExtensionFromUrl(plan.FfmpegSourceArchiveUrl))
	downloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg", DownloadUrl: plan.FfmpegSourceArchiveUrl, ExpectedSha256Hash: plan.FfmpegSourceSha256Hash, DestinationFilePath: archivePath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 1_000_000, ExpectedFileSizeMaximum: 200_000_000, FileConflictPolicy: downloadPolicyForHash(plan.FfmpegSourceSha256Hash)}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, downloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSourceDownload", "FFmpeg source download failed", err)
		return
	}
	signaturePath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-approved-source"+archiveExtensionFromUrl(plan.FfmpegSourceArchiveUrl)+".asc")
	signatureDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg signature", DownloadUrl: plan.FfmpegSourceSignatureUrl, DestinationFilePath: signaturePath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 100, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, signatureDownloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSignatureDownload", "FFmpeg source signature download failed", err)
		return
	}
	publicKeyPath := filepath.Join(workspaceLayout.DownloadsDirectory, "ffmpeg-devel.asc")
	publicKeyDownloadPlan := download.DownloadPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, WorkspaceDirectory: plan.WorkspaceDirectory, DownloadSourceName: "FFmpeg release signing key", DownloadUrl: ffmpegReleaseSigningKeyUrl, DestinationFilePath: publicKeyPath, AllowedHosts: []string{"ffmpeg.org", "www.ffmpeg.org"}, ExpectedFileSizeMinimum: 1000, ExpectedFileSizeMaximum: 100_000, FileConflictPolicy: download.OverwriteFile}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userFfmpegSourceDownloadConsent, publicKeyDownloadPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSigningKeyDownload", "FFmpeg signing key download failed", err)
		return
	}
	if err := verifyFfmpegDetachedSignature(signaturePath, archivePath, publicKeyPath, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSignatureVerification", "FFmpeg source signature verification failed", err)
		return
	}
	extractPlan := extraction.ExtractPlan{ActionName: plan.ActionName, PlanHash: plan.PlanHash, ArchiveFilePath: archivePath, DestinationDirectory: sourceRootDirectory, WorkspaceDirectory: plan.WorkspaceDirectory, ArchiveFormat: archiveFormatFromUrl(plan.FfmpegSourceArchiveUrl), ExtractDestinationPolicy: extraction.RequireNewDirectory, ExtractedFileModePolicy: extraction.PreserveSafeExecutableBits, MaximumFileCount: 50000, MaximumExtractedByteCount: 2_000_000_000, MaximumSingleFileByteCount: 500_000_000}
	if err := extraction.ExtractArchiveWithConsent(ctx, userArchiveExtractionConsent, extractPlan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSourceExtraction", "FFmpeg source extraction failed", err)
		return
	}
	ffmpegSourceDirectory, err = findSingleChildDirectory(sourceRootDirectory)
	if err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegSourceDirectoryMissing", "Could not locate extracted FFmpeg source directory", err)
		return
	}
	if err := app.validateLibraryVersionsAgainstFfmpeg(plan, ffmpegSourceDirectory, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.libraryVersionIncompatible", "A prepared library version is incompatible with the selected FFmpeg release", err)
		return
	}
	if err := app.installFfmpegLibraryPackages(ctx, plan, userLibraryPackageInstallConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegLibraryPackageInstall", "FFmpeg library package installation failed", err)
		return
	}
	if err := app.prepareNonNativeLibraries(ctx, plan, userFfmpegSourceDownloadConsent, userArchiveExtractionConsent, userLibraryPackageInstallConsent, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.libraryPreparation", "Non-Native library preparation failed", err)
		return
	}
	if err := app.executeFfmpegConfigure(ctx, plan, ffmpegSourceDirectory, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegConfigure", "FFmpeg configure failed", err)
		return
	}
	if err := app.executeFfmpegMake(ctx, plan, ffmpegSourceDirectory, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.ffmpegBuild", "FFmpeg build failed", err)
		return
	}
	if err := copyFfmpegBuildOutputs(ffmpegSourceDirectory, workspaceLayout, plan, emitProgress); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		copyFailed = true
		app.emitLocalizedFailure("run.failure.copyArtifacts", "Could not copy FFmpeg artifacts", err)
		return
	}
	if err := writeArtifactReport(workspaceLayout, runId, plan); err != nil {
		_ = auditWriter.WriteEvent("action-failed", plan.ActionName, plan.PlanHash, "error", err.Error())
		app.emitLocalizedFailure("run.failure.writeArtifactReport", "Could not write artifact report", err)
		return
	}
	_ = auditWriter.WriteEvent("action-completed", plan.ActionName, plan.PlanHash, "info", "Approved FFmpeg build completed.")
	emitProgress("info", localize("run.log.ffmpegCompleted", nil))
	actionSucceeded = true
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
		app.emitLog("warn", localize("run.log.configReadFailed", map[string]string{"message": err.Error()}))
		return
	}
	if err := os.WriteFile(destPath, data, 0o600); err != nil {
		app.emitLog("warn", localize("run.log.configSaveFailed", map[string]string{"message": err.Error()}))
		return
	}
	app.emitLog("info", localize("run.log.configSaved", map[string]string{"path": destPath}))
}

func (app *App) cleanupFailedFfmpegRun(plan planning.FfmpegBuildPlan, workspaceLayout workspace.WorkspaceLayout, sourceRootDirectory string, runId string) {
	app.emitLog("warn", localize("run.log.cleaningFfmpegPartial", nil))
	cleanupTargets := []string{
		sourceRootDirectory,
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-library-packages-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-configure-"+plan.PlanHash+".sh"),
		filepath.Join(workspaceLayout.BuildDirectory, "scripts", "ffmpeg-make-"+plan.PlanHash+".sh"),
	}
	for _, preparation := range plan.LibraryPreparations {
		cleanupTargets = append(cleanupTargets,
			filepath.Join(workspaceLayout.BuildDirectory, "prep", preparation.LibraryId+"-"+plan.PlanHash),
			filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
			filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-builddeps-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		)
	}
	app.cleanupWorkspaceTargets(plan.WorkspaceDirectory, cleanupTargets)
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
	// Privately-installed libraries (e.g. libtls) live in their own per-library prefix to
	// avoid archive-name collisions in the shared prefix; their pkgconfig dirs go first so
	// pkg-config resolves them (and their isolated libssl/libcrypto) ahead of the shared
	// prefix's same-named modules.
	paths := append([]string{}, privatePkgConfigDirsFor(plan)...)
	paths = append(paths,
		msys2Prefix+"/lib/pkgconfig",
		msys2Prefix+"/share/pkgconfig",
		"/usr/lib/pkgconfig",
		"/usr/share/pkgconfig",
	)
	return strings.Join(paths, ":")
}

// privatePkgConfigDirsFor returns the unix pkgconfig directories of every privately-installed
// library in the plan (see planning.LibraryPreparation.PrivatePrefixInstall), in plan order.
func privatePkgConfigDirsFor(plan planning.FfmpegBuildPlan) []string {
	profileDirectoryName := strings.ToLower(plan.WindowsShellProfileName)
	if profileDirectoryName == "" {
		profileDirectoryName = "ucrt64"
	}
	msys2Prefix := "/" + profileDirectoryName
	dirs := []string{}
	for _, preparation := range plan.LibraryPreparations {
		if preparation.PrivatePrefixInstall && preparation.PkgConfigName != "" {
			dirs = append(dirs, scripting.PrivateLibraryPkgConfigDir(msys2Prefix, preparation.PkgConfigName))
		}
	}
	return dirs
}

func (app *App) executeFfmpegConfigure(ctx context.Context, plan planning.FfmpegBuildPlan, ffmpegSourceDirectory string, userExternalCommandExecutionConsent consent.CommandExecutionConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.ConfigureScriptLines(plan.ConfigureFlags, privatePkgConfigDirsFor(plan), planning.FfmpegVersionFromArchiveUrl(plan.FfmpegSourceArchiveUrl))
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
