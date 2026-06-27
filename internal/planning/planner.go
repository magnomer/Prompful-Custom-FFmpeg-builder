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

func DefaultBuildConfigSettings() BuildConfigSettings {
	return BuildConfigSettings{
		WorkspaceDirectory:       filepath.Join(defaultUserDataDirectory(), "CustomFFmpegBuilder", "workspace"),
		Msys2ArchiveUrl:          "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst",
		Msys2ArchiveSha256Hash:   "",
		Msys2ArchiveSignatureUrl: "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig",
		Msys2PackageNames:        defaultMsys2PackageNames("ucrt64"),
		WindowsShellProfileName:  "ucrt64",
	}
}

func DefaultFfmpegBuildSettings() FfmpegBuildSettings {
	return FfmpegBuildSettings{
		WorkspaceDirectory:         filepath.Join(defaultUserDataDirectory(), "CustomFFmpegBuilder", "workspace"),
		FfmpegSourceArchiveUrl:     "",
		FfmpegSourceSignatureUrl:   "",
		FfmpegSourceSha256Hash:     "",
		SelectedLibraryIds:         defaultLibraryIds(),
		SelectedConfigureOptionIds: defaultConfigureOptionIds(),
		ExtraConfigureFlags:        []string{},
		ConfigureFlags:             []string{},
		ParallelJobCount:           maxInt(1, runtime.NumCPU()-1),
		WindowsShellProfileName:    "ucrt64",
		LicenseProfileName:         "lgpl-local",
	}
}

func localizedWarning(riskLevel RiskLevel, messageKey string, fallback string, values map[string]string) PlanWarning {
	return PlanWarning{RiskLevel: riskLevel, Message: fallback, MessageKey: messageKey, MessageValues: values}
}

func localizedOperation(operationName string, fallback string) PlanOperation {
	return PlanOperation{OperationName: operationName, Summary: fallback, SummaryKey: "plan.operations." + operationName}
}

// appendUnpreparedTrackWarnings blocks the build only for selected non-Native
// libraries that do not have an implemented preparation recipe yet. Libraries with a
// recipe are prepared before configure and do not block.
func appendUnpreparedTrackWarnings(warnings []PlanWarning, blockedLibraries []LibraryChoice) ([]PlanWarning, bool) {
	blockedInternal := librariesForTrack(blockedLibraries, LibraryTrackInternal)
	blockedExternal := librariesForTrack(blockedLibraries, LibraryTrackExternal)
	hasBlockedWarning := false
	if len(blockedInternal) > 0 {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.internalTrackNotPrepared", "Internal-track libraries are selected, but MSYS2-internal source-build preparation is not implemented yet for them. The build is blocked so configure flags are not approved before those libraries are prepared: "+joinLibraryDisplayNames(blockedInternal)+".", map[string]string{"libraries": joinLibraryDisplayNames(blockedInternal)}))
		hasBlockedWarning = true
	}
	if len(blockedExternal) > 0 {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.externalTrackNotPrepared", "External-track libraries are selected, but outside-build/import preparation is not implemented yet for them. The build is blocked so configure flags are not approved before those libraries are imported and verified: "+joinLibraryDisplayNames(blockedExternal)+".", map[string]string{"libraries": joinLibraryDisplayNames(blockedExternal)}))
		hasBlockedWarning = true
	}
	return warnings, hasBlockedWarning
}

func joinLibraryDisplayNames(libraries []LibraryChoice) string {
	names := make([]string, 0, len(libraries))
	for _, library := range libraries {
		names = append(names, library.DisplayName)
	}
	return strings.Join(names, ", ")
}

func ffmpegBuildOperations(hasInternalLibraries bool, hasExternalLibraries bool) []PlanOperation {
	operations := []PlanOperation{
		localizedOperation("download-ffmpeg-source", "Download the approved FFmpeg source archive."),
		localizedOperation("verify-ffmpeg-source-signature", "Verify the FFmpeg source archive with the matching .asc PGP signature before extraction."),
		localizedOperation("extract-ffmpeg-source", "Extract source into the private workspace."),
		localizedOperation("review-selected-libraries", "Show selected FFmpeg libraries, generated package names, generated configure flags, and license effects."),
		localizedOperation("install-selected-library-packages", "Install only the MSYS2 packages required by the selected FFmpeg libraries before configure runs."),
	}
	if hasInternalLibraries {
		operations = append(operations, localizedOperation("prepare-internal-libraries", "Build selected Internal-track libraries inside the selected MSYS2 environment before configure runs."))
	}
	if hasExternalLibraries {
		operations = append(operations, localizedOperation("prepare-external-libraries", "Import selected External-track libraries into the selected MSYS2 environment before configure runs."))
	}
	if hasInternalLibraries || hasExternalLibraries {
		operations = append(operations, localizedOperation("verify-prepared-libraries", "Verify prepared non-Native libraries before their FFmpeg configure flags are approved."))
	}
	operations = append(operations,
		localizedOperation("run-approved-configure-script", "Run FFmpeg configure with exactly the approved final flags."),
		localizedOperation("run-approved-make-command", "Run make with the approved parallel job count."),
		localizedOperation("create-artifact-report", "Write a build report with source hashes, libraries, flags, and artifact paths."),
	)
	return operations
}

func PlanToolchainSetup(buildConfigSettings BuildConfigSettings) (ToolchainPreparationPlan, error) {
	buildConfigSettings = cleanBuildConfigSettings(buildConfigSettings)
	warnings := validateCommonWindowsWorkspace(buildConfigSettings.WorkspaceDirectory)
	isExecutable := !hasBlockedWarnings(warnings)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.windowsOnly", "This project profile is Windows-only.", nil))
		isExecutable = false
	}
	if buildConfigSettings.Msys2ArchiveUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlEmpty", "MSYS2 archive URL is empty. Use an official MSYS2 tar archive URL before approval. .tar.zst is recommended, and .tar.xz is accepted as a fallback.", nil))
		isExecutable = false
	} else if strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".sig") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".exe") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".sfx.exe") {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlNotTar", "Use an MSYS2 tar archive URL here. The official .exe installer is valid MSYS2, but this app does not run installers; it verifies and extracts tar archives inside the selected workspace. Use .tar.zst when possible, or .tar.xz as a fallback. Put the matching .sig URL in the signature field.", nil))
		isExecutable = false
	} else if !(strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".tar.zst") || strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveUrl), ".tar.xz")) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlBadExtension", "MSYS2 archive URL must end with .tar.zst or .tar.xz. This app uses tar archives so it can verify and extract files itself without running an installer.", nil))
		isExecutable = false
	}
	if buildConfigSettings.Msys2ArchiveSignatureUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.msys2SignatureMissing", "No MSYS2 .sig URL was supplied. The app can calculate SHA-256, but signature verification is the better official authenticity check.", nil))
	} else if !strings.HasSuffix(strings.ToLower(buildConfigSettings.Msys2ArchiveSignatureUrl), ".sig") {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2SignatureBadExtension", "MSYS2 signature URL must end with .sig.", nil))
		isExecutable = false
	} else if buildConfigSettings.Msys2ArchiveUrl != "" && buildConfigSettings.Msys2ArchiveSignatureUrl != buildConfigSettings.Msys2ArchiveUrl+".sig" {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.msys2SignatureMismatch", "MSYS2 signature URL does not exactly match the archive URL plus .sig. This may be intentional, but usually the signature URL should be the archive URL followed by .sig.", nil))
	}
	if buildConfigSettings.Msys2ArchiveSha256Hash != "" && !isSha256Hex(buildConfigSettings.Msys2ArchiveSha256Hash) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ShaBad", "MSYS2 SHA-256 must be exactly 64 hexadecimal characters. If you pasted a .sig file, remove it; .sig is a signature file, not a hash.", nil))
		isExecutable = false
	}
	for _, packageName := range buildConfigSettings.Msys2PackageNames {
		if err := scripting.ValidateMsys2PackageName(packageName); err != nil {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	if !isSupportedWindowsShellProfileName(buildConfigSettings.WindowsShellProfileName) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.shellProfileBad", "Windows shell profile must be ucrt64, mingw64, or clang64.", nil))
		isExecutable = false
	}

	plan := ToolchainPreparationPlan{
		ActionName:                 "prepare-private-msys2-toolchain",
		WorkspaceDirectory:         buildConfigSettings.WorkspaceDirectory,
		Msys2RootDirectory:         Msys2RootDirectoryForProfile(buildConfigSettings.WorkspaceDirectory, buildConfigSettings.WindowsShellProfileName),
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
		ExtractDestinationPolicy:   "must-not-exist",
		Operations: []PlanOperation{
			localizedOperation("create-workspace-directories", "Create directories inside the selected workspace only."),
			localizedOperation("download-msys2-archive", "Download the approved MSYS2 archive from the approved URL."),
			localizedOperation("verify-msys2-signature", "Verify the downloaded MSYS2 archive with its official .sig file using the built-in verifier."),
			localizedOperation("record-msys2-sha256", "Calculate and log the archive SHA-256 for the audit record."),
			localizedOperation("extract-private-msys2", "Extract MSYS2 into the private workspace toolchain directory."),
			localizedOperation("install-approved-pacman-packages", "Install only the package names listed in this plan."),
		},
		Warnings:     warnings,
		IsExecutable: isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := HashPlan(planWithoutHash)
	if err != nil {
		return ToolchainPreparationPlan{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func PlanFfmpegBuild(ffmpegBuildSettings FfmpegBuildSettings) (FfmpegBuildPlan, error) {
	ffmpegBuildSettings = cleanFfmpegBuildSettings(ffmpegBuildSettings)
	warnings := validateCommonWindowsWorkspace(ffmpegBuildSettings.WorkspaceDirectory)
	isExecutable := !hasBlockedWarnings(warnings)
	selectedLibraries, unknownLibraryIds := selectLibraries(ffmpegBuildSettings.WindowsShellProfileName, ffmpegBuildSettings.SelectedLibraryIds)
	// Resolve raw ExtraConfigureFlags back to catalog libraries (excluding ones already
	// explicitly selected) so a flag like --enable-libdavs2 typed in the extra-flags box
	// is subject to the same track gating, operations, and license effects as a checkbox.
	extraLibraries := librariesForConfigureFlags(ffmpegBuildSettings.WindowsShellProfileName, ffmpegBuildSettings.ExtraConfigureFlags, selectedLibraries)
	allEffectiveLibraries := append(append([]LibraryChoice{}, selectedLibraries...), extraLibraries...)
	selectedNativeLibraries := librariesForTrack(selectedLibraries, LibraryTrackNative)
	selectedInternalLibraries := librariesForTrack(selectedLibraries, LibraryTrackInternal)
	selectedExternalLibraries := librariesForTrack(selectedLibraries, LibraryTrackExternal)
	selectedLibrariesByTrack := groupLibrariesByTrack(selectedLibraries)
	// Track gate and prep operations key off every internal/external library that ends up
	// in the build, including extra-flag ones, so a raw flag cannot bypass the gate.
	// Libraries with an implemented prep recipe are prepared before configure; libraries
	// without one yet still block the build.
	gatedInternalLibraries := librariesForTrack(allEffectiveLibraries, LibraryTrackInternal)
	gatedExternalLibraries := librariesForTrack(allEffectiveLibraries, LibraryTrackExternal)
	gatedNonNativeLibraries := append(append([]LibraryChoice{}, gatedInternalLibraries...), gatedExternalLibraries...)
	ffmpegVersion := ffmpegVersionFromArchiveUrl(ffmpegBuildSettings.FfmpegSourceArchiveUrl)
	libraryPreparations, blockedNonNativeLibraries := partitionNonNativeLibraries(gatedNonNativeLibraries, ffmpegVersion)
	prefixPreparationBuildDependencyPackages(libraryPreparations, ffmpegBuildSettings.WindowsShellProfileName)
	warnings, hasNonNativeBlockedWarning := appendUnpreparedTrackWarnings(warnings, blockedNonNativeLibraries)
	if hasNonNativeBlockedWarning {
		isExecutable = false
	}
	for _, unknownLibraryId := range unknownLibraryIds {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.unknownLibraryId", "Unknown library id: "+unknownLibraryId, map[string]string{"id": unknownLibraryId}))
		isExecutable = false
	}
	libraryPackages := uniquePackagesFromLibraries(selectedLibraries)
	libraryFlags := uniqueFlagsFromLibraries(selectedLibraries)
	selectedConfigureOptions, unknownConfigureOptionIds := selectConfigureOptions(ffmpegBuildSettings.SelectedConfigureOptionIds)
	for _, unknownConfigureOptionId := range unknownConfigureOptionIds {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.unknownOptionId", "Unknown FFmpeg option id: "+unknownConfigureOptionId, map[string]string{"id": unknownConfigureOptionId}))
		isExecutable = false
	}
	optionFlags := uniqueFlagsFromConfigureOptions(selectedConfigureOptions)
	finalConfigureFlags := mergeUniqueStrings(libraryFlags, optionFlags)
	finalConfigureFlags = mergeUniqueStrings(finalConfigureFlags, ffmpegBuildSettings.ExtraConfigureFlags)
	libraryPackages = mergeUniqueStrings(libraryPackages, uniquePackagesFromLibraries(extraLibraries))
	derivedLicenseProfileName := deriveLicenseProfileName(allEffectiveLibraries, finalConfigureFlags)
	finalConfigureFlags = addLicenseFlags(finalConfigureFlags, derivedLicenseProfileName, allEffectiveLibraries)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.windowsOnly", "This project profile is Windows-only.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceArchiveUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.ffmpegArchiveUrlEmpty", "FFmpeg source archive URL is empty. Paste an official fixed release archive URL before approval.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSignatureUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.ffmpegSignatureMissing", "FFmpeg source signature URL is empty. FFmpeg releases are verified through the matching .asc PGP signature.", nil))
		isExecutable = false
	} else if !strings.HasSuffix(strings.ToLower(ffmpegBuildSettings.FfmpegSourceSignatureUrl), ".asc") {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.ffmpegSignatureBadExtension", "FFmpeg signature URL must end in .asc. Do not paste the PGP signature text; use the URL of the matching .asc file.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSha256Hash == "" {
		warnings = append(warnings, localizedWarning(RiskLevelInfo, "plan.warnings.ffmpegShaMissing", "No FFmpeg SHA-256 was supplied. This is normal for the official release page: the app will verify the .asc PGP signature and calculate SHA-256 for the log.", nil))
	} else if !isSha256Hex(ffmpegBuildSettings.FfmpegSourceSha256Hash) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.ffmpegShaBad", "FFmpeg SHA-256 must be exactly 64 hexadecimal characters. If you have a .asc or .sig file, do not paste it into this field; it is a signature file, not a hash.", nil))
		isExecutable = false
	}
	if len(finalConfigureFlags) == 0 {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.noConfigureFlags", "No configure flags were selected.", nil))
	}
	for _, configureFlag := range finalConfigureFlags {
		if err := scripting.ValidateConfigureFlag(configureFlag); err != nil {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	for _, packageName := range libraryPackages {
		if err := scripting.ValidateMsys2PackageName(packageName); err != nil {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	configureConflictWarnings, hasConfigureConflicts := validateConfigureFlagConflicts(finalConfigureFlags)
	warnings = append(warnings, configureConflictWarnings...)
	if hasConfigureConflicts {
		isExecutable = false
	}
	if !isSupportedWindowsShellProfileName(ffmpegBuildSettings.WindowsShellProfileName) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.shellProfileBad", "Windows shell profile must be ucrt64, mingw64, or clang64.", nil))
		isExecutable = false
	}
	if ffmpegBuildSettings.ParallelJobCount > 256 {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.parallelJobTooHigh", "Parallel job count must not exceed 256.", nil))
		isExecutable = false
	}
	licenseWarnings, licenseBlocked := validateLicenseProfile(derivedLicenseProfileName, allEffectiveLibraries, finalConfigureFlags)
	warnings = append(warnings, licenseWarnings...)
	if selectedLibrariesRequireVersion3(allEffectiveLibraries) {
		warnings = append(warnings, localizedWarning(RiskLevelInfo, "plan.warnings.version3Added", "FFmpeg version-3 license switch was added because a selected library requires --enable-version3.", nil))
	}
	if licenseBlocked {
		isExecutable = false
	}

	plan := FfmpegBuildPlan{
		ActionName:                 "build-ffmpeg-from-approved-source",
		WorkspaceDirectory:         ffmpegBuildSettings.WorkspaceDirectory,
		Msys2RootDirectory:         Msys2RootDirectoryForProfile(ffmpegBuildSettings.WorkspaceDirectory, ffmpegBuildSettings.WindowsShellProfileName),
		FfmpegSourceArchiveUrl:     ffmpegBuildSettings.FfmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegBuildSettings.FfmpegSourceSignatureUrl,
		FfmpegSourceSha256Hash:     ffmpegBuildSettings.FfmpegSourceSha256Hash,
		SelectedLibraryIds:         ffmpegBuildSettings.SelectedLibraryIds,
		SelectedLibraries:          selectedLibraries,
		SelectedNativeLibraries:    selectedNativeLibraries,
		SelectedInternalLibraries:  selectedInternalLibraries,
		SelectedExternalLibraries:  selectedExternalLibraries,
		SelectedLibrariesByTrack:   selectedLibrariesByTrack,
		LibraryPreparations:        libraryPreparations,
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
		ExtractDestinationPolicy:   "must-not-exist",
		Operations:                 ffmpegBuildOperations(len(preparationsForTrack(libraryPreparations, LibraryTrackInternal)) > 0, len(preparationsForTrack(libraryPreparations, LibraryTrackExternal)) > 0),
		Warnings:                   warnings,
		IsExecutable:               isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := HashPlan(planWithoutHash)
	if err != nil {
		return FfmpegBuildPlan{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func CheckPlanCanRun(isExecutable bool) error {
	if !isExecutable {
		return errors.New("plan is blocked and cannot be executed")
	}
	return nil
}

func cleanBuildConfigSettings(buildConfigSettings BuildConfigSettings) BuildConfigSettings {
	defaults := DefaultBuildConfigSettings()
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

func cleanFfmpegBuildSettings(ffmpegBuildSettings FfmpegBuildSettings) FfmpegBuildSettings {
	defaults := DefaultFfmpegBuildSettings()
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
	ffmpegBuildSettings.SelectedLibraryIds = mergeDefaultLibraryIds(ffmpegBuildSettings.SelectedLibraryIds)
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

func validateCommonWindowsWorkspace(workspaceDirectory string) []PlanWarning {
	warnings := []PlanWarning{}
	if workspaceDirectory == "" {
		return append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.workspaceEmpty", "Workspace directory is empty.", nil))
	}
	if filepath.IsAbs(workspaceDirectory) == false {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.workspaceNotAbsolute", "Workspace directory must be an absolute path.", nil))
	}
	if containsSpace(workspaceDirectory) {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.workspaceContainsSpace", "Workspace path contains a space. Some FFmpeg dependency builds may fail with spaces in paths.", nil))
	}
	return warnings
}

func hasBlockedWarnings(planWarnings []PlanWarning) bool {
	for _, planWarning := range planWarnings {
		if planWarning.RiskLevel == RiskLevelBlocked {
			return true
		}
	}
	return false
}

// Msys2RootDirectoryForProfile returns the per-profile private MSYS2 root. Each
// shell profile gets its own isolated install (toolchains/msys2-<profile>) so a
// ucrt64 environment survives when the user switches to mingw64/clang64 and
// prepares those separately, instead of one shared root being wiped on each prep.
func Msys2RootDirectoryForProfile(workspaceDirectory string, windowsShellProfileName string) string {
	profileName := windowsShellProfileName
	if profileName == "" {
		profileName = "ucrt64"
	}
	return filepath.Join(workspaceDirectory, "toolchains", "msys2-"+profileName)
}

func isSupportedWindowsShellProfileName(windowsShellProfileName string) bool {
	switch windowsShellProfileName {
	case "ucrt64", "mingw64", "clang64":
		return true
	default:
		return false
	}
}

func validateConfigureFlagConflicts(finalConfigureFlags []string) ([]PlanWarning, bool) {
	flagSet := map[string]bool{}
	for _, configureFlag := range finalConfigureFlags {
		flagSet[configureFlag] = true
	}
	warnings := []PlanWarning{}
	blocked := false
	tlsBackendCount := 0
	for _, tlsFlag := range []string{"--enable-openssl", "--enable-gnutls", "--enable-mbedtls", "--enable-libtls"} {
		if flagSet[tlsFlag] {
			tlsBackendCount++
		}
	}
	if tlsBackendCount > 1 {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.tlsBackendConflict", "Choose one TLS backend: OpenSSL, GnuTLS, mbedTLS, or libtls. FFmpeg cannot configure more than one TLS backend at the same time.", nil))
		blocked = true
	}
	if flagSet["--enable-libshaderc"] && flagSet["--enable-libglslang"] {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.shaderCompilerConflict", "Choose one runtime shader compiler: libshaderc or libglslang. FFmpeg configure rejects --enable-libshaderc and --enable-libglslang together; if in doubt, keep libshaderc and disable libglslang.", nil))
		blocked = true
	}
	// FFmpeg forbids enabling the full-profile and baseline-profile EVC bindings of the same
	// codec together (they bind the same XEVD/XEVE library): "libxevd and libxevdb must not be
	// enabled at the same time" (and likewise for the encoder). Keep the full-profile binding,
	// which is the superset.
	if flagSet["--enable-libxevd"] && flagSet["--enable-libxevdb"] {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.evcDecoderConflict", "Choose one EVC decoder binding: libxevd (full profile) or libxevdb (baseline profile). FFmpeg configure rejects enabling both; if in doubt, keep libxevd and disable libxevdb.", nil))
		blocked = true
	}
	if flagSet["--enable-libxeve"] && flagSet["--enable-libxeveb"] {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.evcEncoderConflict", "Choose one EVC encoder binding: libxeve (full profile) or libxeveb (baseline profile). FFmpeg configure rejects enabling both; if in doubt, keep libxeve and disable libxeveb.", nil))
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
			warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.networkDisabledWithProtocol", fmt.Sprintf("Remove all networking support (--disable-network) is selected, so these selected network libraries will not work: %s. Deselect --disable-network or remove those libraries.", joinedNames), map[string]string{"libraries": joinedNames}))
		}
	}
	return warnings, blocked
}

func validateLicenseProfile(licenseProfileName string, selectedLibraries []LibraryChoice, finalConfigureFlags []string) ([]PlanWarning, bool) {
	warnings := []PlanWarning{}
	blocked := false
	switch licenseProfileName {
	case "lgpl-local", "gpl-local", "nonfree-local":
	default:
		return []PlanWarning{localizedWarning(RiskLevelBlocked, "plan.warnings.licenseBoundaryBad", "Derived license boundary must be lgpl-local, gpl-local, or nonfree-local.", nil)}, true
	}
	if licenseProfileName == "gpl-local" {
		warnings = append(warnings, localizedWarning(RiskLevelInfo, "plan.warnings.licenseGpl", "License boundary was set to GPL local because selected libraries or flags require --enable-gpl.", nil))
	}
	if licenseProfileName == "nonfree-local" {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.licenseNonfree", "License boundary was set to nonfree local because selected libraries or flags require --enable-nonfree. Do not redistribute this build unless you have reviewed the license obligations.", nil))
	}
	for _, library := range selectedLibraries {
		if library.LicenseEffectName == "nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.libraryNeedsNonfree", fmt.Sprintf("Library %s requires nonfree-local license boundary.", library.DisplayName), map[string]string{"library": library.DisplayName}))
			blocked = true
		}
		if library.LicenseEffectName == "gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.libraryNeedsGpl", fmt.Sprintf("Library %s requires GPL-compatible license boundary.", library.DisplayName), map[string]string{"library": library.DisplayName}))
			blocked = true
		}
	}
	for _, configureFlag := range finalConfigureFlags {
		if configureFlag == "--enable-nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.enableNonfreeNeedsBoundary", "--enable-nonfree requires nonfree-local license boundary.", nil))
			blocked = true
		}
		if configureFlag == "--enable-gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.enableGplNeedsBoundary", "--enable-gpl requires GPL-compatible license boundary.", nil))
			blocked = true
		}
	}
	return warnings, blocked
}

func defaultUserDataDirectory() string {
	if localAppData := getenv("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	return "."
}

var getenv = os.Getenv

func containsSpace(value string) bool {
	for _, valueRune := range value {
		if valueRune == ' ' || valueRune == '\t' || valueRune == '\n' || valueRune == '\r' {
			return true
		}
	}
	return false
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func isSha256Hex(value string) bool {
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
