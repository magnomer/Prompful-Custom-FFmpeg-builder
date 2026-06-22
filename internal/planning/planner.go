package planning

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"customffmpegbuilder/internal/scripting"
)

func DefaultBuildToolSettings() BuildToolSettings {
	return BuildToolSettings{
		WorkspaceDirectory:       filepath.Join(defaultUserDataDirectory(), "CustomFFmpegBuilder", "workspace"),
		Msys2ArchiveUrl:          "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst",
		Msys2ArchiveSha256Hash:   "",
		Msys2ArchiveSignatureUrl: "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig",
		Msys2PackageNames:        defaultMsys2PackageNames(),
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

func PlanToolchainSetup(buildToolSettings BuildToolSettings) (ToolchainPreparationPlan, error) {
	buildToolSettings = cleanBuildToolSettings(buildToolSettings)
	warnings := validateCommonWindowsWorkspace(buildToolSettings.WorkspaceDirectory)
	isExecutable := !hasBlockedWarnings(warnings)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.windowsOnly", "This project profile is Windows-only.", nil))
		isExecutable = false
	}
	if buildToolSettings.Msys2ArchiveUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlEmpty", "MSYS2 archive URL is empty. Use an official MSYS2 tar archive URL before approval. .tar.zst is recommended, and .tar.xz is accepted as a fallback.", nil))
		isExecutable = false
	} else if strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".sig") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".exe") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".sfx.exe") {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlNotTar", "Use an MSYS2 tar archive URL here. The official .exe installer is valid MSYS2, but this app does not run installers; it verifies and extracts tar archives inside the selected workspace. Use .tar.zst when possible, or .tar.xz as a fallback. Put the matching .sig URL in the signature field.", nil))
		isExecutable = false
	} else if !(strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".tar.zst") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".tar.xz")) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ArchiveUrlBadExtension", "MSYS2 archive URL must end with .tar.zst or .tar.xz. This app uses tar archives so it can verify and extract files itself without running an installer.", nil))
		isExecutable = false
	}
	if buildToolSettings.Msys2ArchiveSignatureUrl == "" {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.msys2SignatureMissing", "No MSYS2 .sig URL was supplied. The app can calculate SHA-256, but signature verification is the better official authenticity check.", nil))
	} else if !strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveSignatureUrl), ".sig") {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2SignatureBadExtension", "MSYS2 signature URL must end with .sig.", nil))
		isExecutable = false
	} else if buildToolSettings.Msys2ArchiveUrl != "" && buildToolSettings.Msys2ArchiveSignatureUrl != buildToolSettings.Msys2ArchiveUrl+".sig" {
		warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.msys2SignatureMismatch", "MSYS2 signature URL does not exactly match the archive URL plus .sig. This may be intentional, but usually the signature URL should be the archive URL followed by .sig.", nil))
	}
	if buildToolSettings.Msys2ArchiveSha256Hash != "" && !isSha256Hex(buildToolSettings.Msys2ArchiveSha256Hash) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.msys2ShaBad", "MSYS2 SHA-256 must be exactly 64 hexadecimal characters. If you pasted a .sig file, remove it; .sig is a signature file, not a hash.", nil))
		isExecutable = false
	}
	for _, packageName := range buildToolSettings.Msys2PackageNames {
		if err := scripting.ValidateMsys2PackageName(packageName); err != nil {
			warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.validationError", err.Error(), map[string]string{"message": err.Error()}))
			isExecutable = false
		}
	}
	if !isSupportedWindowsShellProfileName(buildToolSettings.WindowsShellProfileName) {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.shellProfileBad", "Windows shell profile must be ucrt64, mingw64, or clang64.", nil))
		isExecutable = false
	}

	plan := ToolchainPreparationPlan{
		ActionName:                 "prepare-private-msys2-toolchain",
		WorkspaceDirectory:         buildToolSettings.WorkspaceDirectory,
		Msys2RootDirectory:         filepath.Join(buildToolSettings.WorkspaceDirectory, "toolchains", "msys2"),
		Msys2ArchiveUrl:            buildToolSettings.Msys2ArchiveUrl,
		Msys2ArchiveSha256Hash:     buildToolSettings.Msys2ArchiveSha256Hash,
		Msys2ArchiveSignatureUrl:   buildToolSettings.Msys2ArchiveSignatureUrl,
		Msys2PackageNames:          buildToolSettings.Msys2PackageNames,
		WindowsShellProfileName:    buildToolSettings.WindowsShellProfileName,
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
	extraLibraries := librariesForConfigureFlags(ffmpegBuildSettings.WindowsShellProfileName, ffmpegBuildSettings.ExtraConfigureFlags, selectedLibraries)
	libraryPackages = mergeUniqueStrings(libraryPackages, uniquePackagesFromLibraries(extraLibraries))
	derivedLicenseProfileName := deriveLicenseProfileName(selectedLibraries, finalConfigureFlags)
	finalConfigureFlags = addLicenseFlags(finalConfigureFlags, derivedLicenseProfileName, selectedLibraries)

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
	librariesWithoutMsys2Package := map[string]string{
		"xavs2": "xavs2",
		"vvenc": "vvenc",
	}
	for _, lib := range selectedLibraries {
		if upstreamName, exists := librariesWithoutMsys2Package[lib.LibraryId]; exists {
			warnings = append(warnings, localizedWarning(RiskLevelWarning, "plan.warnings.libraryNoMsys2Package", "No prebuilt MSYS2 package exists for "+lib.DisplayName+" ("+upstreamName+"). FFmpeg configure will fail for this library unless you build and install "+upstreamName+" into the selected MSYS2 prefix yourself.", map[string]string{"library": lib.DisplayName, "upstream": upstreamName}))
		}
	}
	if ffmpegBuildSettings.ParallelJobCount > 256 {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.parallelJobTooHigh", "Parallel job count must not exceed 256.", nil))
		isExecutable = false
	}
	licenseWarnings, licenseBlocked := validateLicenseProfile(derivedLicenseProfileName, selectedLibraries, finalConfigureFlags)
	warnings = append(warnings, licenseWarnings...)
	if selectedLibrariesRequireVersion3(selectedLibraries) {
		warnings = append(warnings, localizedWarning(RiskLevelInfo, "plan.warnings.version3Added", "FFmpeg version-3 license switch was added because a selected library requires --enable-version3.", nil))
	}
	if licenseBlocked {
		isExecutable = false
	}

	plan := FfmpegBuildPlan{
		ActionName:                 "build-ffmpeg-from-approved-source",
		WorkspaceDirectory:         ffmpegBuildSettings.WorkspaceDirectory,
		Msys2RootDirectory:         filepath.Join(ffmpegBuildSettings.WorkspaceDirectory, "toolchains", "msys2"),
		FfmpegSourceArchiveUrl:     ffmpegBuildSettings.FfmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegBuildSettings.FfmpegSourceSignatureUrl,
		FfmpegSourceSha256Hash:     ffmpegBuildSettings.FfmpegSourceSha256Hash,
		SelectedLibraryIds:         ffmpegBuildSettings.SelectedLibraryIds,
		SelectedLibraries:          selectedLibraries,
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
		Operations: []PlanOperation{
			localizedOperation("download-ffmpeg-source", "Download the approved FFmpeg source archive."),
			localizedOperation("verify-ffmpeg-source-signature", "Verify the FFmpeg source archive with the matching .asc PGP signature before extraction."),
			localizedOperation("extract-ffmpeg-source", "Extract source into the private workspace."),
			localizedOperation("review-selected-libraries", "Show selected FFmpeg libraries, generated package names, generated configure flags, and license effects."),
			localizedOperation("install-selected-library-packages", "Install only the MSYS2 packages required by the selected FFmpeg libraries before configure runs."),
			localizedOperation("run-approved-configure-script", "Run FFmpeg configure with exactly the approved final flags."),
			localizedOperation("run-approved-make-command", "Run make with the approved parallel job count."),
			localizedOperation("create-artifact-report", "Write a build report with source hashes, libraries, flags, and artifact paths."),
		},
		Warnings:     warnings,
		IsExecutable: isExecutable,
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

func cleanBuildToolSettings(buildToolSettings BuildToolSettings) BuildToolSettings {
	defaults := DefaultBuildToolSettings()
	if buildToolSettings.WorkspaceDirectory == "" {
		buildToolSettings.WorkspaceDirectory = defaults.WorkspaceDirectory
	}
	if buildToolSettings.Msys2ArchiveSignatureUrl == "" && buildToolSettings.Msys2ArchiveUrl != "" {
		buildToolSettings.Msys2ArchiveSignatureUrl = buildToolSettings.Msys2ArchiveUrl + ".sig"
	}
	if buildToolSettings.WindowsShellProfileName == "" {
		buildToolSettings.WindowsShellProfileName = defaults.WindowsShellProfileName
	}
	if len(buildToolSettings.Msys2PackageNames) == 0 {
		buildToolSettings.Msys2PackageNames = defaults.Msys2PackageNames
	}
	return buildToolSettings
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
	if flagSet["--enable-gnutls"] && flagSet["--enable-openssl"] {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.tlsBackendConflict", "Choose one TLS backend: OpenSSL or GnuTLS. FFmpeg cannot configure both --enable-openssl and --enable-gnutls at the same time.", nil))
		blocked = true
	}
	if flagSet["--enable-libshaderc"] && flagSet["--enable-libglslang"] {
		warnings = append(warnings, localizedWarning(RiskLevelBlocked, "plan.warnings.shaderCompilerConflict", "Choose one runtime shader compiler: libshaderc or libglslang. FFmpeg configure rejects --enable-libshaderc and --enable-libglslang together; if in doubt, keep libshaderc and disable libglslang.", nil))
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
