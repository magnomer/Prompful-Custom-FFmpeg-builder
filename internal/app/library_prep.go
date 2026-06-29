package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"promptfulcustomffmpegbuilder/internal/audit"
	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/download"
	"promptfulcustomffmpegbuilder/internal/execution"
	"promptfulcustomffmpegbuilder/internal/extraction"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/scripting"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

// prepareNonNativeLibraries prepares every Internal- and External-track library in
// the plan before FFmpeg configure runs. Internal-track libraries are built from a
// verified upstream source archive inside the private MSYS2 environment; External-
// track libraries are imported from a verified vendor binary archive. Each library:
//
//  1. downloads a SHA-256-pinned archive (host-allowlisted, size-bounded) through the
//     same download layer the FFmpeg source uses,
//  2. extracts it into a private per-library prep directory inside the workspace,
//  3. runs a hash-pinned, consent-gated prep script that builds/imports into the
//     selected MSYS2 profile prefix and verifies the installed header and library.
//
// The plan only ever contains libraries that have an implemented recipe; libraries
// without one block the plan in the planner and never reach here.
func (app *App) prepareNonNativeLibraries(ctx context.Context, plan planning.FfmpegBuildPlan, userDownloadConsent consent.FfmpegSourceDownloadConsent, userArchiveExtractionConsent consent.ArchiveExtractionConsent, userLibraryPackageInstallConsent consent.PacmanInstallConsent, userExternalCommandExecutionConsent consent.CommandExecutionConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	if len(plan.LibraryPreparations) == 0 {
		return nil
	}
	for _, preparation := range plan.LibraryPreparations {
		emitProgress("info", localize("run.log.libraryPreparationStarted", map[string]string{"library": preparation.DisplayName}))
		if err := app.prepareSingleNonNativeLibrary(ctx, plan, preparation, userDownloadConsent, userArchiveExtractionConsent, userLibraryPackageInstallConsent, userExternalCommandExecutionConsent, auditWriter, emitProgress); err != nil {
			return fmt.Errorf("%s: %w", preparation.DisplayName, err)
		}
		emitProgress("info", localize("run.log.libraryPreparationFinished", map[string]string{"library": preparation.DisplayName}))
	}
	return nil
}

func (app *App) prepareSingleNonNativeLibrary(ctx context.Context, plan planning.FfmpegBuildPlan, preparation planning.LibraryPreparation, userDownloadConsent consent.FfmpegSourceDownloadConsent, userArchiveExtractionConsent consent.ArchiveExtractionConsent, userLibraryPackageInstallConsent consent.PacmanInstallConsent, userExternalCommandExecutionConsent consent.CommandExecutionConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	if err := app.installLibraryBuildDependencyPackages(ctx, plan, preparation, userLibraryPackageInstallConsent, auditWriter, emitProgress); err != nil {
		return err
	}
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	archiveFormat, archiveExtension, err := preparationArchiveFormat(preparation.ArchiveFormat)
	if err != nil {
		return err
	}

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "prep-"+preparation.LibraryId+"-"+plan.PlanHash+archiveExtension)
	downloadPlan := download.DownloadPlan{
		ActionName:              plan.ActionName,
		PlanHash:                plan.PlanHash,
		WorkspaceDirectory:      plan.WorkspaceDirectory,
		DownloadSourceName:      preparation.DisplayName + " preparation archive",
		DownloadUrl:             preparation.ArchiveUrl,
		ExpectedSha256Hash:      preparation.ArchiveSha256Hash,
		DestinationFilePath:     archivePath,
		AllowedHosts:            preparationDownloadAllowedHosts(preparation.AllowedDownloadHost),
		ExpectedFileSizeMinimum: 1_000,
		ExpectedFileSizeMaximum: 5_000_000_000,
		FileConflictPolicy:      downloadPolicyForHash(preparation.ArchiveSha256Hash),
	}
	if err := download.DownloadFfmpegSourceWithConsent(ctx, userDownloadConsent, downloadPlan, emitProgress); err != nil {
		return err
	}

	extractRootDirectory := filepath.Join(workspaceLayout.BuildDirectory, "prep", preparation.LibraryId+"-"+plan.PlanHash)
	if err := app.removeExistingPreparationDirectory(plan.WorkspaceDirectory, extractRootDirectory); err != nil {
		return err
	}
	extractPlan := extraction.ExtractPlan{
		ActionName:                 plan.ActionName,
		PlanHash:                   plan.PlanHash,
		ArchiveFilePath:            archivePath,
		DestinationDirectory:       extractRootDirectory,
		WorkspaceDirectory:         plan.WorkspaceDirectory,
		ArchiveFormat:              archiveFormat,
		ExtractDestinationPolicy:   extraction.RequireNewDirectory,
		ExtractedFileModePolicy:    extraction.PreserveSafeExecutableBits,
		MaximumFileCount:           250000,
		MaximumExtractedByteCount:  10_000_000_000,
		MaximumSingleFileByteCount: 2_000_000_000,
	}
	if err := extraction.ExtractArchiveWithConsent(ctx, userArchiveExtractionConsent, extractPlan, emitProgress); err != nil {
		return err
	}

	// The build/import script runs with its working directory at the extracted source
	// root, auto-detected so no version-specific directory name is hardcoded: archives
	// that wrap everything in a single top-level folder (typical source tarballs like
	// uavs3d-<rev>/) descend into it; archives that lay their contents out at the top
	// level (typical vendor zips with include/ + lib/) stay at the extract root. The
	// script never receives the archive URL, only already-verified content.
	workingDirectory, err := resolvePreparationSourceRoot(extractRootDirectory)
	if err != nil {
		return err
	}

	scriptLines, err := preparationScriptLines(preparation)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{
		WorkspaceDirectory: plan.WorkspaceDirectory,
		ScriptFilePath:     filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		ScriptLines:        scriptLines,
	})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{
		ActionName:                 plan.ActionName,
		PlanHash:                   plan.PlanHash,
		ExecutablePath:             filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"),
		ArgumentValues:             []string{filepath.ToSlash(scriptFile.ScriptFilePath)},
		WorkingDirectory:           workingDirectory,
		WorkspaceDirectory:         plan.WorkspaceDirectory,
		Msys2RootDirectory:         plan.Msys2RootDirectory,
		WindowsShellProfileName:    plan.WindowsShellProfileName,
		EnvironmentVariables:       map[string]string{},
		AllowedExecutableBasenames: []string{"bash.exe"},
		ScriptKind:                 execution.LibraryPreparationScript,
		ApprovedScriptFilePath:     scriptFile.ScriptFilePath,
		ApprovedScriptSha256Hash:   scriptFile.ScriptSha256Hash,
		RunLogDirectory:            auditWriter.LogDirectory(),
	}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved "+preparation.DisplayName+" preparation script.")
	return execution.RunCommandWithConsent(ctx, userExternalCommandExecutionConsent, commandPlan, emitProgress)
}

// installLibraryBuildDependencyPackages installs the MSYS2 packages a library's build
// system needs at configure time (e.g. a Python3 interpreter) into the build environment
// before its source build runs. It reuses the same hash-pinned, consent-gated pacman
// installation path as the FFmpeg library packages, so the same one pacman consent the
// run already collected covers it. Libraries with no build dependencies are skipped.
func (app *App) installLibraryBuildDependencyPackages(ctx context.Context, plan planning.FfmpegBuildPlan, preparation planning.LibraryPreparation, userLibraryPackageInstallConsent consent.PacmanInstallConsent, auditWriter *audit.Writer, emitProgress func(string, string)) error {
	// Profile-prefixed mingw packages plus verbatim MSYS base packages (e.g. autotools)
	// are installed together in one pacman transaction.
	dependencyPackages := append(append([]string{}, preparation.BuildDependencyPackages...), preparation.MsysBuildDependencyPackages...)
	if len(dependencyPackages) == 0 {
		return nil
	}
	if err := consent.CheckConsent(userLibraryPackageInstallConsent.Consent, consent.ConsentKindPacmanPackageInstallation, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	emitProgress("info", localize("run.log.libraryBuildDependenciesStarted", map[string]string{"library": preparation.DisplayName}))
	workspaceLayout := workspace.WorkspaceLayoutFor(plan.WorkspaceDirectory)
	scriptLines, err := scripting.FfmpegLibraryPackageInstallScriptLines(dependencyPackages)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.WriteScriptFile(scripting.ScriptFilePlan{
		WorkspaceDirectory: plan.WorkspaceDirectory,
		ScriptFilePath:     filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-builddeps-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		ScriptLines:        scriptLines,
	})
	if err != nil {
		return err
	}
	commandPlan := execution.CommandPlan{
		ActionName:                 plan.ActionName,
		PlanHash:                   plan.PlanHash,
		ExecutablePath:             filepath.Join(plan.Msys2RootDirectory, "usr", "bin", "bash.exe"),
		ArgumentValues:             []string{filepath.ToSlash(scriptFile.ScriptFilePath)},
		WorkingDirectory:           plan.Msys2RootDirectory,
		WorkspaceDirectory:         plan.WorkspaceDirectory,
		Msys2RootDirectory:         plan.Msys2RootDirectory,
		WindowsShellProfileName:    plan.WindowsShellProfileName,
		EnvironmentVariables:       map[string]string{},
		AllowedExecutableBasenames: []string{"bash.exe"},
		ScriptKind:                 execution.PacmanInstallScript,
		ApprovedScriptFilePath:     scriptFile.ScriptFilePath,
		ApprovedScriptSha256Hash:   scriptFile.ScriptSha256Hash,
		RunLogDirectory:            auditWriter.LogDirectory(),
	}
	_ = auditWriter.WriteEvent("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved "+preparation.DisplayName+" build dependency installation script.")
	return execution.RunPacmanWithConsent(ctx, userLibraryPackageInstallConsent, commandPlan, emitProgress)
}

// preparationDownloadAllowedHosts returns the hosts trusted for a preparation archive
// download. It expands github.com to the CDN hosts GitHub always redirects downloads to
// (codeload.github.com for /archive/refs tarballs, objects.githubusercontent.com for
// release assets). Those are GitHub's own infrastructure and the archive is still gated
// by SHA-256, so trusting the redirect target removes a spurious "untrusted host"
// warning without weakening verification.
func preparationDownloadAllowedHosts(allowedDownloadHost string) []string {
	hosts := []string{allowedDownloadHost}
	if allowedDownloadHost == "github.com" {
		hosts = append(hosts, "codeload.github.com", "objects.githubusercontent.com")
	}
	return hosts
}

func (app *App) removeExistingPreparationDirectory(workspaceDirectory string, preparationDirectory string) error {
	if _, err := os.Lstat(preparationDirectory); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, preparationDirectory); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, preparationDirectory); err != nil {
		return err
	}
	return removeAllWithRetry(preparationDirectory)
}

// preparationScriptLines maps a planning preparation recipe to the scripting-layer
// build spec and returns the generated, validated script lines for its method.
func preparationScriptLines(preparation planning.LibraryPreparation) ([]string, error) {
	buildSpec := scripting.LibraryBuildSpec{
		LibraryId:                preparation.LibraryId,
		DisplayName:              preparation.DisplayName,
		BuildSystem:              string(preparation.BuildSystem),
		CFlags:                   preparation.CFlags,
		CMakeOptions:             preparation.CMakeOptions,
		CMakeBuildTargets:        preparation.CMakeBuildTargets,
		ConfigureSubdir:          preparation.ConfigureSubdir,
		ConfigureOptions:         preparation.ConfigureOptions,
		RunAutogen:               preparation.RunAutogen,
		MakeBuildTargets:         preparation.MakeBuildTargets,
		MakeInstallTargets:       preparation.MakeInstallTargets,
		MakeVariables:            preparation.MakeVariables,
		MakeInstallHeaderFiles:   preparation.MakeInstallHeaderFiles,
		MakeStaticLibFile:        preparation.MakeStaticLibFile,
		ImportIncludeSubdir:      preparation.ImportIncludeSubdir,
		ImportLibSubdir:          preparation.ImportLibSubdir,
		PkgConfigName:            preparation.PkgConfigName,
		PkgConfigAppendLibs:      preparation.PkgConfigAppendLibs,
		PkgConfigLibsLine:        preparation.PkgConfigLibsLine,
		PrivatePrefixInstall:     preparation.PrivatePrefixInstall,
		VerifyHeaderRelativePath: preparation.VerifyHeaderRelativePath,
		VerifyLibStem:            preparation.VerifyLibStem,
		SourcePatches:            preparationSourcePatches(preparation.SourcePatches),
		GeneratedSourceFiles:     preparationGeneratedSourceFiles(preparation.GeneratedSourceFiles),
	}
	switch preparation.Method {
	case planning.PreparationMethodInternalSource:
		return scripting.InternalLibrarySourceBuildScriptLines(buildSpec)
	case planning.PreparationMethodExternalImport:
		return scripting.ExternalLibraryImportScriptLines(buildSpec)
	default:
		return nil, fmt.Errorf("unknown library preparation method: %s", preparation.Method)
	}
}

// preparationSourcePatches maps planning-layer source patches to the scripting-layer type
// (scripting cannot import planning without an import cycle).
func preparationSourcePatches(patches []planning.LibrarySourcePatch) []scripting.LibrarySourcePatch {
	if len(patches) == 0 {
		return nil
	}
	mapped := make([]scripting.LibrarySourcePatch, 0, len(patches))
	for _, patch := range patches {
		mapped = append(mapped, scripting.LibrarySourcePatch{File: patch.File, Find: patch.Find, Replace: patch.Replace})
	}
	return mapped
}

// preparationGeneratedSourceFiles maps planning-layer generated source files to the
// scripting-layer type (scripting cannot import planning without an import cycle).
func preparationGeneratedSourceFiles(files []planning.GeneratedSourceFile) []scripting.GeneratedSourceFile {
	if len(files) == 0 {
		return nil
	}
	mapped := make([]scripting.GeneratedSourceFile, 0, len(files))
	for _, file := range files {
		mapped = append(mapped, scripting.GeneratedSourceFile{Path: file.Path, Lines: file.Lines})
	}
	return mapped
}

// resolvePreparationSourceRoot returns the directory the build/import script should run
// in. If the extracted archive contains exactly one entry and it is a directory (a
// source tarball wrapped in a single top-level folder), it descends into that folder;
// otherwise it uses the extract root (a vendor archive whose contents sit at the top
// level). This keeps the generic layer free of any version-specific directory name.
func resolvePreparationSourceRoot(extractRootDirectory string) (string, error) {
	entries, err := os.ReadDir(extractRootDirectory)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(extractRootDirectory, entries[0].Name()), nil
	}
	return extractRootDirectory, nil
}

func preparationArchiveFormat(archiveFormatName string) (extraction.ArchiveFormat, string, error) {
	switch archiveFormatName {
	case "tar.gz", "tgz":
		return extraction.ArchiveFormatTarGz, ".tar.gz", nil
	case "tar.xz":
		return extraction.TarXz, ".tar.xz", nil
	case "tar.zst":
		return extraction.TarZst, ".tar.zst", nil
	case "tar.bz2":
		return extraction.TarBz2, ".tar.bz2", nil
	case "zip":
		return extraction.ArchiveFormatZip, ".zip", nil
	default:
		return "", "", fmt.Errorf("unsupported preparation archive format: %s", archiveFormatName)
	}
}
