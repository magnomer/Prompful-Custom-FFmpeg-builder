package program

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

// lLibraryNonnativePrepare prepares every Internal- and External-track library in
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
func (program *LProgram) lLibraryNonnativePrepare(LContext context.Context, plan planning.LPlanFFmpeg, userDownloadLConsent consent.LConsentFFmpeg, userLConsentArchive consent.LConsentArchive, userLibraryPackageInstallLConsent consent.LConsentPacman, userExternalLConsentCommand consent.LConsentCommand, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	if len(plan.LLibraryPreparationList) == 0 {
		return nil
	}
	for _, preparation := range plan.LLibraryPreparationList {
		emitProgress("info", LLocaleTextGetInternal("run.log.libraryPreparationStarted", map[string]string{"library": preparation.DisplayName}))
		if err := program.lLibrarySinglePrepare(LContext, plan, preparation, userDownloadLConsent, userLConsentArchive, userLibraryPackageInstallLConsent, userExternalLConsentCommand, auditWriter, emitProgress); err != nil {
			return fmt.Errorf("%s: %w", preparation.DisplayName, err)
		}
		emitProgress("info", LLocaleTextGetInternal("run.log.libraryPreparationFinished", map[string]string{"library": preparation.DisplayName}))
	}
	return nil
}

func (program *LProgram) lLibrarySinglePrepare(LContext context.Context, plan planning.LPlanFFmpeg, preparation planning.LLibraryPreparation, userDownloadLConsent consent.LConsentFFmpeg, userLConsentArchive consent.LConsentArchive, userLibraryPackageInstallLConsent consent.LConsentPacman, userExternalLConsentCommand consent.LConsentCommand, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	if err := program.lPackageDependencyInstall(LContext, plan, preparation, userLibraryPackageInstallLConsent, auditWriter, emitProgress); err != nil {
		return err
	}
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	archiveFormat, archiveExtension, err := LArchivePreparationResolve(preparation.LArchiveFormat)
	if err != nil {
		return err
	}

	archivePath := filepath.Join(workspaceLayout.DownloadsDirectory, "prep-"+preparation.LibraryId+"-"+plan.PlanHash+archiveExtension)
	downloadPlan := download.LPlanDownload{
		ActionName:              plan.ActionName,
		PlanHash:                plan.PlanHash,
		WorkspaceDirectory:      plan.WorkspaceDirectory,
		DownloadSourceName:      preparation.DisplayName + " preparation archive",
		DownloadUrl:             preparation.ArchiveUrl,
		ExpectedSha256Hash:      preparation.ArchiveSha256Hash,
		DestinationFilePath:     archivePath,
		AllowedHosts:            LDownloadHostList(preparation.AllowedDownloadHost),
		ExpectedFileSizeMinimum: 1_000,
		ExpectedFileSizeMaximum: 5_000_000_000,
		LPolicyFile:             LPolicyHashResolve(preparation.ArchiveSha256Hash),
	}
	if err := download.LDownloadFFmpegRun(LContext, userDownloadLConsent, downloadPlan, emitProgress); err != nil {
		return err
	}

	extractRootDirectory := filepath.Join(workspaceLayout.BuildDirectory, "prep", preparation.LibraryId+"-"+plan.PlanHash)
	if err := program.lDirectoryPreparationRemove(plan.WorkspaceDirectory, extractRootDirectory); err != nil {
		return err
	}
	extractPlan := extraction.LPlanExtraction{
		ActionName:                 plan.ActionName,
		PlanHash:                   plan.PlanHash,
		ArchiveFilePath:            archivePath,
		DestinationDirectory:       extractRootDirectory,
		WorkspaceDirectory:         plan.WorkspaceDirectory,
		LArchiveFormat:             archiveFormat,
		LPolicyExtraction:          extraction.LPolicyExtractionRequireNewDirectory,
		LPolicyFilemode:            extraction.LPolicyFilemodeExecutablePreserve,
		MaximumFileCount:           250000,
		MaximumExtractedByteCount:  10_000_000_000,
		MaximumSingleFileByteCount: 2_000_000_000,
	}
	if err := extraction.LArchiveConsentExtract(LContext, userLConsentArchive, extractPlan, emitProgress); err != nil {
		return err
	}

	// The build/import script runs with its working directory at the extracted source
	// root, auto-detected so no version-specific directory name is hardcoded: archives
	// that wrap everything in a single top-level folder (typical source tarballs like
	// uavs3d-<rev>/) descend into it; archives that lay their contents out at the top
	// level (typical vendor zips with include/ + lib/) stay at the extract root. The
	// script never receives the archive URL, only already-verified content.
	workingDirectory, err := LPreparationSourceResolve(extractRootDirectory)
	if err != nil {
		return err
	}

	scriptLines, err := LScriptPreparationBuild(preparation)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{
		WorkspaceDirectory: plan.WorkspaceDirectory,
		ScriptFilePath:     filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		ScriptLines:        scriptLines,
	})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{
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
		LScriptKind:                execution.LScriptLibraryPreparation,
		ApprovedScriptFilePath:     scriptFile.ScriptFilePath,
		ApprovedScriptSha256Hash:   scriptFile.ScriptSha256Hash,
		RunLAuditDirectoryGet:      auditWriter.LAuditDirectoryGet(),
	}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved "+preparation.DisplayName+" preparation script.")
	return execution.LCommandConsentRun(LContext, userExternalLConsentCommand, commandPlan, emitProgress)
}

// lPackageDependencyInstall installs the MSYS2 packages a library's build
// system needs at configure time (e.g. a Python3 interpreter) into the build environment
// before its source build runs. It reuses the same hash-pinned, consent-gated pacman
// installation path as the FFmpeg library packages, so the same one pacman consent the
// run already collected covers it. Libraries with no build dependencies are skipped.
func (program *LProgram) lPackageDependencyInstall(LContext context.Context, plan planning.LPlanFFmpeg, preparation planning.LLibraryPreparation, userLibraryPackageInstallLConsent consent.LConsentPacman, auditWriter *audit.LAuditWriter, emitProgress func(string, string)) error {
	// Profile-prefixed mingw packages plus verbatim MSYS base packages (e.g. autotools)
	// are installed together in one pacman transaction.
	dependencyPackages := append(append([]string{}, preparation.BuildDependencyPackages...), preparation.MsysBuildDependencyPackages...)
	if len(dependencyPackages) == 0 {
		return nil
	}
	if err := consent.LConsentCheck(userLibraryPackageInstallLConsent.LConsent, consent.LConsentKindPacman, plan.ActionName, plan.PlanHash); err != nil {
		return err
	}
	emitProgress("info", LLocaleTextGetInternal("run.log.libraryBuildDependenciesStarted", map[string]string{"library": preparation.DisplayName}))
	workspaceLayout := workspace.LWorkspaceLayoutResolve(plan.WorkspaceDirectory)
	scriptLines, err := scripting.LScriptPackageLinesCreate(dependencyPackages)
	if err != nil {
		return err
	}
	scriptFile, err := scripting.LScriptFileWrite(scripting.LPlanScript{
		WorkspaceDirectory: plan.WorkspaceDirectory,
		ScriptFilePath:     filepath.Join(workspaceLayout.BuildDirectory, "scripts", "prep-builddeps-"+preparation.LibraryId+"-"+plan.PlanHash+".sh"),
		ScriptLines:        scriptLines,
	})
	if err != nil {
		return err
	}
	commandPlan := execution.LPlanCommand{
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
		LScriptKind:                execution.LScriptPacmanInstall,
		ApprovedScriptFilePath:     scriptFile.ScriptFilePath,
		ApprovedScriptSha256Hash:   scriptFile.ScriptSha256Hash,
		RunLAuditDirectoryGet:      auditWriter.LAuditDirectoryGet(),
	}
	_ = auditWriter.LAuditEventWrite("command-started", plan.ActionName, plan.PlanHash, "info", "Running approved "+preparation.DisplayName+" build dependency installation script.")
	return execution.LCommandPacmanRun(LContext, userLibraryPackageInstallLConsent, commandPlan, emitProgress)
}

// LDownloadHostList returns the hosts trusted for a preparation archive
// download. It expands github.com to the CDN hosts GitHub always redirects downloads to
// (codeload.github.com for /archive/refs tarballs, objects.githubusercontent.com for
// release assets). Those are GitHub's own infrastructure and the archive is still gated
// by SHA-256, so trusting the redirect target removes a spurious "untrusted host"
// warning without weakening verification.
func LDownloadHostList(allowedDownloadHost string) []string {
	hosts := []string{allowedDownloadHost}
	if allowedDownloadHost == "github.com" {
		hosts = append(hosts, "codeload.github.com", "objects.githubusercontent.com")
	}
	return hosts
}

func (program *LProgram) lDirectoryPreparationRemove(workspaceDirectory string, preparationDirectory string) error {
	if _, err := os.Lstat(preparationDirectory); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if err := workspace.LPathWorkspaceCheck(workspaceDirectory, preparationDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, preparationDirectory); err != nil {
		return err
	}
	return LPathRetryRemove(preparationDirectory)
}

// LScriptPreparationBuild maps a planning preparation recipe to the scripting-layer
// build spec and returns the generated, validated script lines for its LMethod.
func LScriptPreparationBuild(preparation planning.LLibraryPreparation) ([]string, error) {
	buildSpec := scripting.LLibraryBuildSpec{
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
		PkgConfigAppendCFlags:    preparation.PkgConfigAppendCFlags,
		PkgConfigLibsLine:        preparation.PkgConfigLibsLine,
		PkgConfigLibsLinePatches: LPkgConfigLibsLinePatchConvert(preparation.PkgConfigLibsLinePatches),
		PrivatePrefixInstall:     preparation.PrivatePrefixInstall,
		VerifyHeaderRelativePath: preparation.VerifyHeaderRelativePath,
		VerifyLibStem:            preparation.VerifyLibStem,
		SourcePatches:            LPatchSourceConvert(preparation.SourcePatches),
		GeneratedSourceFiles:     LFileGeneratedConvert(preparation.GeneratedSourceFiles),
	}
	switch preparation.Method {
	case planning.LLibraryInternalMethod:
		return scripting.LScriptLibraryInternalCreate(buildSpec)
	case planning.LLibraryExternalMethod:
		return scripting.LScriptLibraryExternalCreate(buildSpec)
	default:
		return nil, fmt.Errorf("unknown library preparation LMethod: %s", preparation.Method)
	}
}

func LPkgConfigLibsLinePatchConvert(patches []planning.LPkgConfigLibsLinePatch) []scripting.LPkgConfigLibsLinePatch {
	if len(patches) == 0 {
		return nil
	}
	mapped := make([]scripting.LPkgConfigLibsLinePatch, 0, len(patches))
	for _, patch := range patches {
		mapped = append(mapped, scripting.LPkgConfigLibsLinePatch{Module: patch.Module, LibsLine: patch.LibsLine})
	}
	return mapped
}

// LPatchSourceConvert maps planning-layer source patches to the scripting-layer type
// (scripting cannot import planning without an import cycle).
func LPatchSourceConvert(patches []planning.LSourcePatch) []scripting.LSourcePatch {
	if len(patches) == 0 {
		return nil
	}
	mapped := make([]scripting.LSourcePatch, 0, len(patches))
	for _, patch := range patches {
		mapped = append(mapped, scripting.LSourcePatch{File: patch.File, Find: patch.Find, Replace: patch.Replace})
	}
	return mapped
}

// LFileGeneratedConvert maps planning-layer generated source files to the
// scripting-layer type (scripting cannot import planning without an import cycle).
func LFileGeneratedConvert(files []planning.LFileGenerated) []scripting.LFileGenerated {
	if len(files) == 0 {
		return nil
	}
	mapped := make([]scripting.LFileGenerated, 0, len(files))
	for _, file := range files {
		mapped = append(mapped, scripting.LFileGenerated{Path: file.Path, Lines: file.Lines})
	}
	return mapped
}

// LPreparationSourceResolve returns the directory the build/import script should run
// in. If the extracted archive contains exactly one entry and it is a directory (a
// source tarball wrapped in a single top-level folder), it descends into that folder;
// otherwise it uses the extract root (a vendor archive whose contents sit at the top
// level). This keeps the generic layer free of any version-specific directory name.
func LPreparationSourceResolve(extractRootDirectory string) (string, error) {
	entries, err := os.ReadDir(extractRootDirectory)
	if err != nil {
		return "", err
	}
	if len(entries) == 1 && entries[0].IsDir() {
		return filepath.Join(extractRootDirectory, entries[0].Name()), nil
	}
	return extractRootDirectory, nil
}

func LArchivePreparationResolve(archiveFormatName string) (extraction.LArchiveFormat, string, error) {
	switch archiveFormatName {
	case "tar.gz", "tgz":
		return extraction.LArchiveTarGz, ".tar.gz", nil
	case "tar.xz":
		return extraction.LArchiveTarXz, ".tar.xz", nil
	case "tar.zst":
		return extraction.LArchiveTarZst, ".tar.zst", nil
	case "tar.bz2":
		return extraction.LArchiveTarBz2, ".tar.bz2", nil
	case "zip":
		return extraction.LArchiveZip, ".zip", nil
	default:
		return "", "", fmt.Errorf("unsupported preparation archive format: %s", archiveFormatName)
	}
}
