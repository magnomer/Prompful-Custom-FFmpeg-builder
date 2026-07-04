package planning

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"promptfulcustomffmpegbuilder/internal/scripting"
)

func LSettingsBuildCreate() LSettingsToolchain {
	return LSettingsToolchain{
		WorkspaceDirectory:       filepath.Join(LUserDirectoryResolve(), "CustomFFmpegBuilder", "workspace"),
		Msys2ArchiveUrl:          "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst",
		Msys2ArchiveSha256Hash:   "",
		Msys2ArchiveSignatureUrl: "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig",
		Msys2PackageNames:        LPackageDefaultGet("ucrt64"),
		WindowsShellProfileName:  "ucrt64",
	}
}

func LSettingsFFmpegCreate() LSettingsFFmpeg {
	return LSettingsFFmpeg{
		WorkspaceDirectory:         filepath.Join(LUserDirectoryResolve(), "CustomFFmpegBuilder", "workspace"),
		FfmpegSourceArchiveUrl:     "",
		FfmpegSourceSignatureUrl:   "",
		FfmpegSourceSha256Hash:     "",
		SelectedLibraryIds:         []string{},
		SelectedConfigureOptionIds: LOptionDefaultGet(),
		ExtraConfigureFlags:        []string{},
		ConfigureFlags:             []string{},
		ParallelJobCount:           LNumberMaxGet(1, runtime.NumCPU()),
		WindowsShellProfileName:    "ucrt64",
		LicenseProfileName:         "lgpl-local",
	}
}

func LWarningLocalizedCreate(riskLevel LRiskLevel, messageKey string, fallback string, values map[string]string) LWarningPlan {
	return LWarningPlan{LRiskLevel: riskLevel, Message: fallback, MessageKey: messageKey, MessageValues: values}
}

func LOperationLocalizedCreate(operationName string, fallback string) LOperationPlan {
	return LOperationPlan{OperationName: operationName, Summary: fallback, SummaryKey: "plan.operations." + operationName}
}

// LWarningTrackAppend blocks the build only for selected non-Native
// libraries that do not have an implemented preparation recipe yet. Libraries with a
// recipe are prepared before configure and do not block.
func LWarningTrackAppend(warnings []LWarningPlan, blockedLibraries []LLibraryChoice) ([]LWarningPlan, bool) {
	blockedInternal := LTrackFilterGet(blockedLibraries, LLibraryTrackInternal)
	blockedExternal := LTrackFilterGet(blockedLibraries, LLibraryTrackExternal)
	hasBlockedWarning := false
	if len(blockedInternal) > 0 {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.internalTrackNotPrepared", "Internal-track libraries are selected, but MSYS2-internal source-build preparation is not implemented yet for them. The build is blocked so configure flags are not approved before those libraries are prepared: "+LLibraryNameJoin(blockedInternal)+".", map[string]string{"libraries": LLibraryNameJoin(blockedInternal)}))
		hasBlockedWarning = true
	}
	if len(blockedExternal) > 0 {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.externalTrackNotPrepared", "External-track libraries are selected, but outside-build/import preparation is not implemented yet for them. The build is blocked so configure flags are not approved before those libraries are imported and verified: "+LLibraryNameJoin(blockedExternal)+".", map[string]string{"libraries": LLibraryNameJoin(blockedExternal)}))
		hasBlockedWarning = true
	}
	return warnings, hasBlockedWarning
}

func LLibraryNameJoin(libraries []LLibraryChoice) string {
	names := make([]string, 0, len(libraries))
	for _, library := range libraries {
		names = append(names, library.DisplayName)
	}
	return strings.Join(names, ", ")
}

func LOperationFFmpegBuild(hasInternalLibraries bool, hasExternalLibraries bool) []LOperationPlan {
	operations := []LOperationPlan{
		LOperationLocalizedCreate("download-ffmpeg-source", "Download the approved FFmpeg source archive."),
		LOperationLocalizedCreate("verify-ffmpeg-source-signature", "Verify the FFmpeg source archive with the matching .asc PGP signature before extraction."),
		LOperationLocalizedCreate("extract-ffmpeg-source", "Extract source into the private workspace."),
		LOperationLocalizedCreate("review-selected-libraries", "Show selected FFmpeg libraries, generated package names, generated configure flags, and license effects."),
		LOperationLocalizedCreate("install-selected-library-packages", "Install only the MSYS2 packages required by the selected FFmpeg libraries before configure runs."),
	}
	if hasInternalLibraries {
		operations = append(operations, LOperationLocalizedCreate("prepare-internal-libraries", "Build selected Internal-track libraries inside the selected MSYS2 environment before configure runs."))
	}
	if hasExternalLibraries {
		operations = append(operations, LOperationLocalizedCreate("prepare-external-libraries", "Import selected External-track libraries into the selected MSYS2 environment before configure runs."))
	}
	if hasInternalLibraries || hasExternalLibraries {
		operations = append(operations, LOperationLocalizedCreate("verify-prepared-libraries", "Verify prepared non-Native libraries before their FFmpeg configure flags are approved."))
	}
	operations = append(operations,
		LOperationLocalizedCreate("run-approved-configure-script", "Run FFmpeg configure with exactly the approved final flags."),
		LOperationLocalizedCreate("run-approved-make-command", "Run make with the approved parallel job count."),
		LOperationLocalizedCreate("create-artifact-report", "Write a build report with source hashes, libraries, flags, and artifact paths."),
	)
	return operations
}

func LPlanToolchainCreate(buildConfigSettings LSettingsToolchain) (LPlanToolchain, error) {
	buildConfigSettings = LSettingsBuildClean(buildConfigSettings)
	warnings := LWorkspaceWindowsValidate(buildConfigSettings.WorkspaceDirectory)
	isExecutable := !LWarningBlockedCheck(warnings)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.windowsOnly", "This project profile is Windows-only.", nil))
		isExecutable = false
	}
	if buildConfigSettings.Msys2ArchiveUrl == "" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.msys2ArchiveUrlEmpty", "MSYS2 archive URL is empty. Use an official MSYS2 tar archive URL before approval. .tar.zst is recommended, and .tar.xz is accepted as a fallback.", nil))
		isExecutable = false
	} else if strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".sig") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".exe") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".sfx.exe") {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.msys2ArchiveUrlNotTar", "Use an MSYS2 tar archive URL here. The official .exe installer is valid MSYS2, but this program does not run installers; it verifies and extracts tar archives inside the selected workspace. Use .tar.zst when possible, or .tar.xz as a fallback. Put the matching .sig URL in the signature field.", nil))
		isExecutable = false
	} else if !(strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".tar.zst") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".tar.xz")) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.msys2ArchiveUrlBadExtension", "MSYS2 archive URL must end with .tar.zst or .tar.xz. This program uses tar archives so it can verify and extract files itself without running an installer.", nil))
		isExecutable = false
	}
	if buildConfigSettings.Msys2ArchiveSignatureUrl == "" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.msys2SignatureMissing", "No MSYS2 .sig URL was supplied. The program can calculate SHA-256, but signature verification is the better official authenticity check.", nil))
	} else if !strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveSignatureUrl), ".sig") {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.msys2SignatureBadExtension", "MSYS2 signature URL must end with .sig.", nil))
		isExecutable = false
	} else if buildConfigSettings.Msys2ArchiveUrl != "" && buildConfigSettings.Msys2ArchiveSignatureUrl != buildConfigSettings.Msys2ArchiveUrl+".sig" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.msys2SignatureMismatch", "MSYS2 signature URL does not exactly match the archive URL plus .sig. This may be intentional, but usually the signature URL should be the archive URL followed by .sig.", nil))
	}
	if buildConfigSettings.Msys2ArchiveSha256Hash != "" && !LHashSHA256Check(buildConfigSettings.Msys2ArchiveSha256Hash) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.msys2ShaBad", "MSYS2 SHA-256 must be exactly 64 hexadecimal characters. If you pasted a .sig file, remove it; .sig is a signature file, not a hash.", nil))
		isExecutable = false
	}
	for _, packageName := range buildConfigSettings.Msys2PackageNames {
		if err := scripting.LPackageMsysValidate(packageName); err != nil {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	if !LProfileShellCheck(buildConfigSettings.WindowsShellProfileName) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.shellProfileBad", "Windows shell profile must be ucrt64, mingw64, or clang64.", nil))
		isExecutable = false
	}

	plan := LPlanToolchain{
		ActionName:                 "prepare-private-msys2-toolchain",
		WorkspaceDirectory:         buildConfigSettings.WorkspaceDirectory,
		Msys2RootDirectory:         LDirectoryProfileResolve(buildConfigSettings.WorkspaceDirectory, buildConfigSettings.WindowsShellProfileName),
		Msys2ArchiveUrl:            buildConfigSettings.Msys2ArchiveUrl,
		Msys2ArchiveSha256Hash:     buildConfigSettings.Msys2ArchiveSha256Hash,
		Msys2ArchiveSignatureUrl:   buildConfigSettings.Msys2ArchiveSignatureUrl,
		Msys2PackageNames:          buildConfigSettings.Msys2PackageNames,
		WindowsShellProfileName:    buildConfigSettings.WindowsShellProfileName,
		WillModifySystemPath:       false,
		WillRequireAdminRights:     false,
		WillUseExistingMsys2:       false,
		WillDeleteFiles:            false,
		DownloadConflictPolicyName: "reuse-if-hash-matches",
		LPolicyExtraction:          "must-not-exist",
		Operations: []LOperationPlan{
			LOperationLocalizedCreate("create-workspace-directories", "Create directories inside the selected workspace only."),
			LOperationLocalizedCreate("download-msys2-archive", "Download the approved MSYS2 archive from the approved URL."),
			LOperationLocalizedCreate("verify-msys2-signature", "Verify the downloaded MSYS2 archive with its official .sig file using the built-in verifier."),
			LOperationLocalizedCreate("record-msys2-sha256", "Calculate and log the archive SHA-256 for the audit record."),
			LOperationLocalizedCreate("extract-private-msys2", "Extract MSYS2 into the private workspace toolchain directory."),
			LOperationLocalizedCreate("install-approved-pacman-packages", "Install only the package names listed in this plan."),
		},
		Warnings:     warnings,
		IsExecutable: isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := LPlanHashCreate(planWithoutHash)
	if err != nil {
		return LPlanToolchain{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func LPlanFFmpegCreate(ffmpegBuildSettings LSettingsFFmpeg) (LPlanFFmpeg, error) {
	ffmpegBuildSettings = LSettingsFFmpegClean(ffmpegBuildSettings)
	resolvedVersionPlan, hasResolvedVersionPlan, err := LCatalogPlanResolve(ffmpegBuildSettings)
	if err != nil {
		return LPlanFFmpeg{}, err
	}
	var resolvedVersionPlanPointer *LResolvedVersionPlan
	var resolvedBuildPlanPointer *LResolvedBuildPlan
	if hasResolvedVersionPlan {
		resolvedVersionPlanCopy := resolvedVersionPlan
		resolvedVersionPlanPointer = &resolvedVersionPlanCopy
		resolvedBuildPlan := LCatalogBuildResolve(ffmpegBuildSettings, resolvedVersionPlan)
		resolvedBuildPlanPointer = &resolvedBuildPlan
	}
	warnings := LWorkspaceWindowsValidate(ffmpegBuildSettings.WorkspaceDirectory)
	isExecutable := !LWarningBlockedCheck(warnings)
	ffmpegVersion := LArchiveURLParse(ffmpegBuildSettings.FfmpegSourceArchiveUrl)
	compatibilityFfmpegVersion := ffmpegVersion
	if hasResolvedVersionPlan {
		compatibilityFfmpegVersion = resolvedVersionPlan.FfmpegVersion
	}
	selectedLibraryIdsForPlan := LStringsSortedGet(ffmpegBuildSettings.SelectedLibraryIds)
	selectedLibraries := []LLibraryChoice{}
	unknownLibraryIds := []string{}
	extraLibraries := []LLibraryChoice{}
	if hasResolvedVersionPlan && resolvedBuildPlanPointer != nil {
		// For catalog-supported FFmpeg releases, the V5 resolved catalog is the
		// authoritative build input. The legacy catalog is no longer allowed to
		// recalculate selected libraries, tracks, package names, or library flags.
		selectedLibraryIdsForPlan = append([]string{}, resolvedVersionPlan.NormalizedLibraryIds...)
		selectedLibraries = LLibraryChoicesCreate(resolvedBuildPlanPointer.SelectedLibraries)
		resolvedVisibleLibraryChoices := LLibraryChoicesCreate(resolvedVersionPlan.VisibleLibraries)
		extraLibraries = LLibraryFlagMatch(resolvedVisibleLibraryChoices, ffmpegBuildSettings.ExtraConfigureFlags, selectedLibraries)
		warnings = append(warnings, resolvedBuildPlanPointer.Warnings...)
		if LWarningBlockedCheck(resolvedBuildPlanPointer.Warnings) {
			isExecutable = false
		}
	} else if len(ffmpegBuildSettings.SelectedLibraryIds) > 0 || len(ffmpegBuildSettings.ExtraConfigureFlags) > 0 {
		// No hard-coded library/version fallback is allowed. If the selected FFmpeg
		// source does not resolve to an exact embedded catalog version, library ids
		// and extra library flags remain unresolved and the plan is blocked below.
		unknownLibraryIds = append(unknownLibraryIds, ffmpegBuildSettings.SelectedLibraryIds...)
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.ffmpegVersionUnresolved", "FFmpeg source archive URL does not resolve to an embedded supported version; library selection was not resolved through a fallback catalog.", nil))
		isExecutable = false
	}
	allEffectiveLibraries := append(append([]LLibraryChoice{}, selectedLibraries...), extraLibraries...)
	selectedNativeLibraries := LTrackFilterGet(selectedLibraries, LLibraryTrackNative)
	selectedInternalLibraries := LTrackFilterGet(selectedLibraries, LLibraryTrackInternal)
	selectedExternalLibraries := LTrackFilterGet(selectedLibraries, LLibraryTrackExternal)
	selectedLibrariesByTrack := LTrackGroupCreate(selectedLibraries)
	// Track gate and prep operations key off every internal/external library that ends up
	// in the build, including extra-flag ones, so a raw flag cannot bypass the gate.
	// Libraries with an implemented prep recipe are prepared before configure; libraries
	// without one yet still block the build.
	gatedInternalLibraries := LTrackFilterGet(allEffectiveLibraries, LLibraryTrackInternal)
	gatedExternalLibraries := LTrackFilterGet(allEffectiveLibraries, LLibraryTrackExternal)
	gatedNonNativeLibraries := append(append([]LLibraryChoice{}, gatedInternalLibraries...), gatedExternalLibraries...)
	libraryPreparations, blockedNonNativeLibraries := LLibraryPartitionCreate(gatedNonNativeLibraries, compatibilityFfmpegVersion)
	LDependencyPrefixGet(libraryPreparations, ffmpegBuildSettings.WindowsShellProfileName)
	warnings, hasNonNativeBlockedWarning := LWarningTrackAppend(warnings, blockedNonNativeLibraries)
	if hasNonNativeBlockedWarning {
		isExecutable = false
	}
	warnings, hasFfmpegVersionBlock := LWarningFFmpegAppend(warnings, ffmpegVersion, allEffectiveLibraries, ffmpegBuildSettings.SelectedConfigureOptionIds)
	if hasFfmpegVersionBlock {
		isExecutable = false
	}
	for _, unknownLibraryId := range unknownLibraryIds {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.unknownLibraryId", "Unknown library id: "+unknownLibraryId, map[string]string{"id": unknownLibraryId}))
		isExecutable = false
	}
	libraryPackages := LLibraryPackageGet(selectedLibraries)
	libraryFlags := LLibraryFlagGet(selectedLibraries)
	selectedConfigureOptions, unknownConfigureOptionIds := LOptionConfigureSelect(ffmpegBuildSettings.SelectedConfigureOptionIds)
	for _, unknownConfigureOptionId := range unknownConfigureOptionIds {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.unknownOptionId", "Unknown FFmpeg option id: "+unknownConfigureOptionId, map[string]string{"id": unknownConfigureOptionId}))
		isExecutable = false
	}
	optionFlags := LOptionFlagGet(selectedConfigureOptions)
	finalConfigureFlags := LTextUniqueMerge(libraryFlags, optionFlags)
	finalConfigureFlags = LTextUniqueMerge(finalConfigureFlags, ffmpegBuildSettings.ExtraConfigureFlags)
	libraryPackages = LTextUniqueMerge(libraryPackages, LLibraryPackageGet(extraLibraries))
	derivedLicenseProfileName := LLicenseProfileGet(allEffectiveLibraries, finalConfigureFlags)
	finalConfigureFlags = LLicenseFlagAdd(finalConfigureFlags, derivedLicenseProfileName, allEffectiveLibraries)
	// Force the Windows-only base flags on (e.g. --disable-vaapi) regardless of selection, but
	// judge "did the user select anything" before adding them so the no-flags hint still fires.
	userSelectedAnyConfigureFlags := len(finalConfigureFlags) > 0
	finalConfigureFlags = LTextUniqueMerge(LHardwareFlagList(), finalConfigureFlags)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.windowsOnly", "This project profile is Windows-only.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceArchiveUrl == "" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.ffmpegArchiveUrlEmpty", "FFmpeg source archive URL is empty. Paste an official fixed release archive URL before approval.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSignatureUrl == "" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.ffmpegSignatureMissing", "FFmpeg source signature URL is empty. FFmpeg releases are verified through the matching .asc PGP signature.", nil))
		isExecutable = false
	} else if !strings.HasSuffix(strings.ToLower(ffmpegBuildSettings.FfmpegSourceSignatureUrl), ".asc") {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.ffmpegSignatureBadExtension", "FFmpeg signature URL must end in .asc. Do not paste the PGP signature text; use the URL of the matching .asc file.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSha256Hash == "" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskInfo, "plan.warnings.ffmpegShaMissing", "No FFmpeg SHA-256 was supplied. This is normal for the official release page: the program will verify the .asc PGP signature and calculate SHA-256 for the log.", nil))
	} else if !LHashSHA256Check(ffmpegBuildSettings.FfmpegSourceSha256Hash) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.ffmpegShaBad", "FFmpeg SHA-256 must be exactly 64 hexadecimal characters. If you have a .asc or .sig file, do not paste it into this field; it is a signature file, not a hash.", nil))
		isExecutable = false
	}
	if !userSelectedAnyConfigureFlags {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.noConfigureFlags", "No configure flags were selected.", nil))
	}
	for _, configureFlag := range finalConfigureFlags {
		if err := scripting.LFlagConfigureValidate(configureFlag); err != nil {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	for _, packageName := range libraryPackages {
		if err := scripting.LPackageMsysValidate(packageName); err != nil {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	configureConflictWarnings, hasConfigureConflicts := LOptionConflictValidate(finalConfigureFlags)
	warnings = append(warnings, configureConflictWarnings...)
	if hasConfigureConflicts {
		isExecutable = false
	}
	if !LProfileShellCheck(ffmpegBuildSettings.WindowsShellProfileName) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.shellProfileBad", "Windows shell profile must be ucrt64, mingw64, or clang64.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.ParallelJobCount > 256 {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.parallelJobTooHigh", "Parallel job count must not exceed 256.", nil))
		isExecutable = false
	}
	licenseWarnings, licenseBlocked := LLicenseProfileValidate(derivedLicenseProfileName, allEffectiveLibraries, finalConfigureFlags)
	warnings = append(warnings, licenseWarnings...)
	if LVersionMajorCheck(allEffectiveLibraries) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskInfo, "plan.warnings.version3Added", "FFmpeg version-3 license switch was added because a selected library requires --enable-version3.", nil))
	}
	if licenseBlocked {
		isExecutable = false
	}

	plan := LPlanFFmpeg{
		ActionName:                 "build-ffmpeg-from-approved-source",
		WorkspaceDirectory:         ffmpegBuildSettings.WorkspaceDirectory,
		Msys2RootDirectory:         LDirectoryProfileResolve(ffmpegBuildSettings.WorkspaceDirectory, ffmpegBuildSettings.WindowsShellProfileName),
		FfmpegSourceArchiveUrl:     ffmpegBuildSettings.FfmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegBuildSettings.FfmpegSourceSignatureUrl,
		FfmpegSourceSha256Hash:     ffmpegBuildSettings.FfmpegSourceSha256Hash,
		RequestedFfmpegVersion:     ffmpegVersion,
		CompatibilityFfmpegVersion: compatibilityFfmpegVersion,
		SelectedLibraryIds:         selectedLibraryIdsForPlan,
		SelectedLibraries:          selectedLibraries,
		SelectedNativeLibraries:    selectedNativeLibraries,
		SelectedInternalLibraries:  selectedInternalLibraries,
		SelectedExternalLibraries:  selectedExternalLibraries,
		SelectedLibrariesByTrack:   selectedLibrariesByTrack,
		LPreparationCatalog:        libraryPreparations,
		RequiredMsys2PackageNames:  libraryPackages,
		GeneratedConfigureFlags:    libraryFlags,
		SelectedConfigureOptions:   selectedConfigureOptions,
		GeneratedOptionFlags:       optionFlags,
		ExtraConfigureFlags:        ffmpegBuildSettings.ExtraConfigureFlags,
		ConfigureFlags:             finalConfigureFlags,
		ParallelJobCount:           ffmpegBuildSettings.ParallelJobCount,
		WindowsShellProfileName:    ffmpegBuildSettings.WindowsShellProfileName,
		LicenseProfileName:         derivedLicenseProfileName,
		WillModifySystemPath:       false,
		WillRequireAdminRights:     false,
		WillUseExistingMsys2:       false,
		WillDeleteFiles:            false,
		DownloadConflictPolicyName: "reuse-if-hash-matches",
		LPolicyExtraction:          "must-not-exist",
		Operations:                 LOperationFFmpegBuild(len(LLibraryTrackFilter(libraryPreparations, LLibraryTrackInternal)) > 0, len(LLibraryTrackFilter(libraryPreparations, LLibraryTrackExternal)) > 0),
		Warnings:                   warnings,
		ResolvedVersionPlan:        resolvedVersionPlanPointer,
		ResolvedBuildPlan:          resolvedBuildPlanPointer,
		IsExecutable:               isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := LPlanHashCreate(planWithoutHash)
	if err != nil {
		return LPlanFFmpeg{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func LPlanRunCheck(isExecutable bool) error {
	if !isExecutable {
		return errors.New("plan is blocked and cannot be executed")
	}
	return nil
}

func LSettingsBuildClean(buildConfigSettings LSettingsToolchain) LSettingsToolchain {
	defaults := LSettingsBuildCreate()
	if buildConfigSettings.WorkspaceDirectory == "" {
		buildConfigSettings.WorkspaceDirectory = defaults.WorkspaceDirectory
	}
	if buildConfigSettings.Msys2ArchiveSignatureUrl == "" && buildConfigSettings.Msys2ArchiveUrl != "" {
		buildConfigSettings.Msys2ArchiveSignatureUrl = buildConfigSettings.Msys2ArchiveUrl + ".sig"
	}
	if buildConfigSettings.WindowsShellProfileName == "" {
		buildConfigSettings.WindowsShellProfileName = defaults.WindowsShellProfileName
	}
	if len(buildConfigSettings.Msys2PackageNames) == 0 {
		buildConfigSettings.Msys2PackageNames = defaults.Msys2PackageNames
	}
	return buildConfigSettings
}

func LSettingsFFmpegClean(ffmpegBuildSettings LSettingsFFmpeg) LSettingsFFmpeg {
	defaults := LSettingsFFmpegCreate()
	if ffmpegBuildSettings.WorkspaceDirectory == "" {
		ffmpegBuildSettings.WorkspaceDirectory = defaults.WorkspaceDirectory
	}
	if ffmpegBuildSettings.FfmpegSourceSignatureUrl == "" && ffmpegBuildSettings.FfmpegSourceArchiveUrl != "" {
		ffmpegBuildSettings.FfmpegSourceSignatureUrl = ffmpegBuildSettings.FfmpegSourceArchiveUrl + ".asc"
	}
	if ffmpegBuildSettings.WindowsShellProfileName == "" {
		ffmpegBuildSettings.WindowsShellProfileName = defaults.WindowsShellProfileName
	}
	if ffmpegBuildSettings.ParallelJobCount < 1 {
		ffmpegBuildSettings.ParallelJobCount = defaults.ParallelJobCount
	}
	if len(ffmpegBuildSettings.SelectedConfigureOptionIds) == 0 {
		ffmpegBuildSettings.SelectedConfigureOptionIds = defaults.SelectedConfigureOptionIds
	}
	if len(ffmpegBuildSettings.ExtraConfigureFlags) == 0 && len(ffmpegBuildSettings.ConfigureFlags) > 0 {
		ffmpegBuildSettings.ExtraConfigureFlags = ffmpegBuildSettings.ConfigureFlags
	}
	if ffmpegBuildSettings.LicenseProfileName == "" {
		ffmpegBuildSettings.LicenseProfileName = defaults.LicenseProfileName
	}
	return ffmpegBuildSettings
}

func LWorkspaceWindowsValidate(workspaceDirectory string) []LWarningPlan {
	warnings := []LWarningPlan{}
	if workspaceDirectory == "" {
		return append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.workspaceEmpty", "Workspace directory is empty.", nil))
	}
	if filepath.IsAbs(workspaceDirectory) == false {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.workspaceNotAbsolute", "Workspace directory must be an absolute path.", nil))
	}
	if LTextSpaceCheck(workspaceDirectory) {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.workspaceContainsSpace", "Workspace path contains a space. Some FFmpeg dependency builds may fail with spaces in paths.", nil))
	}
	return warnings
}

func LWarningBlockedCheck(planWarnings []LWarningPlan) bool {
	for _, planWarning := range planWarnings {
		if planWarning.LRiskLevel == LRiskBlocked {
			return true
		}
	}
	return false
}

// LDirectoryProfileResolve returns the per-profile private MSYS2 root. Each
// shell profile gets its own isolated install (toolchains/msys2-<profile>) so a
// ucrt64 environment survives when the user switches to mingw64/clang64 and
// prepares those separately, instead of one shared root being wiped on each prep.
func LDirectoryProfileResolve(workspaceDirectory string, windowsShellProfileName string) string {
	profileName := windowsShellProfileName
	if profileName == "" {
		profileName = "ucrt64"
	}
	return filepath.Join(workspaceDirectory, "toolchains", "msys2-"+profileName)
}

func LProfileShellCheck(windowsShellProfileName string) bool {
	switch windowsShellProfileName {
	case "ucrt64", "mingw64", "clang64":
		return true
	default:
		return false
	}
}

func LOptionConflictValidate(finalConfigureFlags []string) ([]LWarningPlan, bool) {
	flagSet := map[string]bool{}
	for _, configureFlag := range finalConfigureFlags {
		flagSet[configureFlag] = true
	}
	warnings := []LWarningPlan{}
	blocked := false
	tlsBackendCount := 0
	for _, tlsFlag := range []string{"--enable-openssl", "--enable-gnutls", "--enable-mbedtls", "--enable-libtls"} {
		if flagSet[tlsFlag] {
			tlsBackendCount++
		}
	}
	if tlsBackendCount > 1 {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.tlsBackendConflict", "Choose one TLS backend: OpenSSL, GnuTLS, mbedTLS, or libtls. FFmpeg cannot configure more than one TLS backend at the same time.", nil))
		blocked = true
	}
	if flagSet["--enable-libshaderc"] && flagSet["--enable-libglslang"] {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.shaderCompilerConflict", "Choose one runtime shader compiler: libshaderc or libglslang. FFmpeg configure rejects --enable-libshaderc and --enable-libglslang together; if in doubt, keep libshaderc and disable libglslang.", nil))
		blocked = true
	}
	// FFmpeg forbids enabling the full-profile and baseline-profile EVC bindings of the same
	// codec together (they bind the same XEVD/XEVE library): "libxevd and libxevdb must not be
	// enabled at the same time" (and likewise for the encoder). Keep the full-profile binding,
	// which is the superset.
	if flagSet["--enable-libxevd"] && flagSet["--enable-libxevdb"] {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.evcDecoderConflict", "Choose one EVC decoder binding: libxevd (full profile) or libxevdb (baseline profile). FFmpeg configure rejects enabling both; if in doubt, keep libxevd and disable libxevdb.", nil))
		blocked = true
	}
	if flagSet["--enable-libxeve"] && flagSet["--enable-libxeveb"] {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.evcEncoderConflict", "Choose one EVC encoder binding: libxeve (full profile) or libxeveb (baseline profile). FFmpeg configure rejects enabling both; if in doubt, keep libxeve and disable libxeveb.", nil))
		blocked = true
	}
	// Intel Hardware Acceleration has two mutually exclusive backends: oneVPL (--enable-libvpl, the
	// libvpl row) and the legacy Media SDK (--enable-libmfx). FFmpeg configure dies with "can not use
	// libmfx and libvpl together" when both are enabled, so block the combination early. Prefer oneVPL,
	// the maintained path (FFmpeg itself deprecates libmfx in favor of libvpl).
	if flagSet["--enable-libvpl"] && flagSet["--enable-libmfx"] {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.intelHwaccelBackendConflict", "Choose one Intel Hardware Acceleration backend: oneVPL (--enable-libvpl) or the legacy libmfx (--enable-libmfx). FFmpeg configure rejects enabling both; if in doubt, keep oneVPL and disable libmfx.", nil))
		blocked = true
	}
	if flagSet["--disable-network"] {
		networkLibraries := []struct{ flag, name string }{
			{"--enable-libsrt", "SRT"},
			{"--enable-libssh", "libssh"},
			{"--enable-librtmp", "librtmp"},
			{"--enable-librist", "librist"},
			{"--enable-libzmq", "ZeroMQ"},
			{"--enable-librabbitmq", "RabbitMQ-C"},
		}
		enabledNames := []string{}
		for _, networkLibrary := range networkLibraries {
			if flagSet[networkLibrary.flag] {
				enabledNames = append(enabledNames, networkLibrary.name)
			}
		}
		if len(enabledNames) > 0 {
			joinedNames := strings.Join(enabledNames, ", ")
			warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.networkDisabledWithProtocol", fmt.Sprintf("Remove all networking support (--disable-network) is selected, so these selected network libraries will not work: %s. Deselect --disable-network or remove those libraries.", joinedNames), map[string]string{"libraries": joinedNames}))
		}
	}
	return warnings, blocked
}

func LLicenseProfileValidate(licenseProfileName string, selectedLibraries []LLibraryChoice, finalConfigureFlags []string) ([]LWarningPlan, bool) {
	warnings := []LWarningPlan{}
	blocked := false
	switch licenseProfileName {
	case "lgpl-local", "gpl-local", "nonfree-local":
	default:
		return []LWarningPlan{LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.licenseBoundaryBad", "Derived license boundary must be lgpl-local, gpl-local, or nonfree-local.", nil)}, true
	}
	if licenseProfileName == "gpl-local" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskInfo, "plan.warnings.licenseGpl", "License boundary was set to GPL local because selected libraries or flags require --enable-gpl.", nil))
	}
	if licenseProfileName == "nonfree-local" {
		warnings = append(warnings, LWarningLocalizedCreate(LRiskWarning, "plan.warnings.licenseNonfree", "License boundary was set to nonfree local because selected libraries or flags require --enable-nonfree. Do not redistribute this build unless you have reviewed the license obligations.", nil))
	}
	for _, library := range selectedLibraries {
		if library.LicenseEffectName == "nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.libraryNeedsNonfree", fmt.Sprintf("Library %s requires nonfree-local license boundary.", library.DisplayName), map[string]string{"library": library.DisplayName}))
			blocked = true
		}
		if library.LicenseEffectName == "gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.libraryNeedsGpl", fmt.Sprintf("Library %s requires GPL-compatible license boundary.", library.DisplayName), map[string]string{"library": library.DisplayName}))
			blocked = true
		}
	}
	for _, configureFlag := range finalConfigureFlags {
		if configureFlag == "--enable-nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.enableNonfreeNeedsBoundary", "--enable-nonfree requires nonfree-local license boundary.", nil))
			blocked = true
		}
		if configureFlag == "--enable-gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, LWarningLocalizedCreate(LRiskBlocked, "plan.warnings.enableGplNeedsBoundary", "--enable-gpl requires GPL-compatible license boundary.", nil))
			blocked = true
		}
	}
	return warnings, blocked
}

func LUserDirectoryResolve() string {
	if localAppData := LEnvironmentVariable("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	return "."
}

var LEnvironmentVariable = os.Getenv

func LTextSpaceCheck(value string) bool {
	for _, valueRune := range value {
		if valueRune == ' ' || valueRune == '\t' || valueRune == '\n' || valueRune == '\r' {
			return true
		}
	}
	return false
}

func LNumberMaxGet(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func LHashSHA256Check(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
