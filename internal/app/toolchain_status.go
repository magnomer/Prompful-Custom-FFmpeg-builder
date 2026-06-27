package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

// toolchainManifestFileName is written inside the private MSYS2 root on a
// successful toolchain preparation. It lives inside the root so it shares the
// root's lifecycle: wiping the root (re-prepare / failed cleanup) removes it too,
// which keeps the recorded metadata from outliving the install it describes.
const toolchainManifestFileName = ".ffmpeg-builder-toolchain.json"

type toolchainManifest struct {
	CreatedAt               string   `json:"createdAt"`
	WindowsShellProfileName string   `json:"windowsShellProfileName"`
	Msys2ArchiveUrl         string   `json:"msys2ArchiveUrl"`
	Msys2PackageNames       []string `json:"msys2PackageNames"`
	PlanHash                string   `json:"planHash"`
}

// ToolchainStatus is the fast, disk-based answer to "is this workspace already
// prepared?". The presence of usr/bin/bash.exe is authoritative; the manifest
// only enriches it with metadata and may be absent for installs made before the
// manifest existed.
type ToolchainStatus struct {
	Installed               bool     `json:"installed"`
	Healthy                 bool     `json:"healthy"`
	Msys2RootDirectory      string   `json:"msys2RootDirectory"`
	CreatedAt               string   `json:"createdAt"`
	WindowsShellProfileName string   `json:"windowsShellProfileName"`
	Msys2ArchiveUrl         string   `json:"msys2ArchiveUrl"`
	PackageCount            int      `json:"packageCount"`
	PackageNames            []string `json:"packageNames"`
	PlanHash                string   `json:"planHash"`
}

// ToolchainVerification is the deep, on-demand check: it queries the private
// MSYS2 package database and reports which recorded packages are actually
// installed versus missing.
type ToolchainVerification struct {
	Verified            bool     `json:"verified"`
	CheckedPackageCount int      `json:"checkedPackageCount"`
	MissingPackageNames []string `json:"missingPackageNames"`
	Message             string   `json:"message"`
}

func msys2RootDirectoryFor(workspaceDirectory string, windowsShellProfileName string) string {
	return planning.Msys2RootDirectoryForProfile(workspaceDirectory, windowsShellProfileName)
}

func toolchainBashPath(msys2RootDirectory string) string {
	return filepath.Join(msys2RootDirectory, "usr", "bin", "bash.exe")
}

// writeToolchainManifest records the successful install metadata. It is
// best-effort: a write failure does not invalidate the (already complete)
// toolchain, only the recovery metadata.
func writeToolchainManifest(plan planning.ToolchainPreparationPlan) error {
	manifestPath := filepath.Join(plan.Msys2RootDirectory, toolchainManifestFileName)
	if err := workspace.CheckRealPathInsideWorkspace(plan.WorkspaceDirectory, manifestPath); err != nil {
		return err
	}
	manifest := toolchainManifest{
		CreatedAt:               time.Now().UTC().Format(time.RFC3339),
		WindowsShellProfileName: plan.WindowsShellProfileName,
		Msys2ArchiveUrl:         plan.Msys2ArchiveUrl,
		Msys2PackageNames:       plan.Msys2PackageNames,
		PlanHash:                plan.PlanHash,
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath, data, 0o644)
}

func readToolchainManifest(msys2RootDirectory string) (toolchainManifest, error) {
	manifest := toolchainManifest{}
	data, err := os.ReadFile(filepath.Join(msys2RootDirectory, toolchainManifestFileName))
	if err != nil {
		return manifest, err
	}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return manifest, err
	}
	return manifest, nil
}

// checkToolchainPreparedForBuild refuses an FFmpeg build when the build's shell
// profile has no prepared toolchain of its own. With per-profile roots each
// profile is independent, so the only failure mode is "this profile was never
// prepared" ??caught here with a clear message instead of a cryptic failure deep
// in configure/make.
func checkToolchainPreparedForBuild(workspaceDirectory string, buildShellProfileName string) error {
	if workspaceDirectory == "" {
		return errors.New("workspace directory is empty")
	}
	msys2RootDirectory := msys2RootDirectoryFor(workspaceDirectory, buildShellProfileName)
	if !fileExists(toolchainBashPath(msys2RootDirectory)) {
		return fmt.Errorf("%s", localize("run.failure.toolchainNotPreparedForProfile", map[string]string{
			"profile": buildShellProfileName,
		}))
	}
	return nil
}

// GetToolchainStatus reports whether a private MSYS2 toolchain already exists for
// the given shell profile, so Prep can recover the "already prepared" state of
// each profile independently across relaunches. Fast: a stat plus an optional
// manifest read.
func (app *App) GetToolchainStatus(workspaceDirectory string, windowsShellProfileName string) (ToolchainStatus, error) {
	status := ToolchainStatus{PackageNames: []string{}}
	if workspaceDirectory == "" {
		return status, nil
	}
	msys2RootDirectory := msys2RootDirectoryFor(workspaceDirectory, windowsShellProfileName)
	status.Msys2RootDirectory = msys2RootDirectory
	if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, msys2RootDirectory); err != nil {
		return status, err
	}
	status.Healthy = fileExists(toolchainBashPath(msys2RootDirectory))
	status.Installed = status.Healthy
	if !status.Installed {
		return status, nil
	}
	manifest, err := readToolchainManifest(msys2RootDirectory)
	if err != nil {
		// Installed without a manifest (older install). Still report it as present.
		return status, nil
	}
	status.CreatedAt = manifest.CreatedAt
	status.WindowsShellProfileName = manifest.WindowsShellProfileName
	status.Msys2ArchiveUrl = manifest.Msys2ArchiveUrl
	status.PackageNames = manifest.Msys2PackageNames
	status.PackageCount = len(manifest.Msys2PackageNames)
	status.PlanHash = manifest.PlanHash
	return status, nil
}

// supportedShellProfileNames lists the profiles that can have their own private
// toolchain root. Kept in sync with isSupportedWindowsShellProfileName in planning.
var supportedShellProfileNames = []string{"ucrt64", "mingw64", "clang64"}

// GetInstalledToolchainProfiles returns the status of every shell profile that
// already has a prepared toolchain in this workspace. Profiles are independent
// (per-profile roots), so several can be installed at once; this lets Prep show
// all of them, not just the currently selected one.
func (app *App) GetInstalledToolchainProfiles(workspaceDirectory string) ([]ToolchainStatus, error) {
	installedProfiles := []ToolchainStatus{}
	if workspaceDirectory == "" {
		return installedProfiles, nil
	}
	for _, profileName := range supportedShellProfileNames {
		status, err := app.GetToolchainStatus(workspaceDirectory, profileName)
		if err != nil {
			continue
		}
		if !status.Installed {
			continue
		}
		// The per-profile root is authoritative for which profile this is, even
		// if an older install has no manifest to read the name from.
		status.WindowsShellProfileName = profileName
		installedProfiles = append(installedProfiles, status)
	}
	return installedProfiles, nil
}

// ClearBuildEnvironments removes every app-managed private build environment in
// the selected workspace. These environments live under workspace/toolchains and
// contain the per-profile MSYS2 roots prepared by Prep. The workspace boundary is
// checked before deletion so a malformed workspace path cannot delete outside
// the selected workspace.
func (app *App) ClearBuildEnvironments(workspaceDirectory string) error {
	if workspaceDirectory == "" {
		return errors.New("workspace directory is empty")
	}
	workspaceLayout := workspace.WorkspaceLayoutFor(workspaceDirectory)
	if err := workspace.CheckPathInsideWorkspace(workspaceDirectory, workspaceLayout.ToolchainsDirectory); err != nil {
		return err
	}
	toolchainsInfo, err := os.Stat(workspaceLayout.ToolchainsDirectory)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	if !toolchainsInfo.IsDir() {
		return fmt.Errorf("toolchains path is not a directory: %s", workspaceLayout.ToolchainsDirectory)
	}
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, workspaceLayout.ToolchainsDirectory); err != nil {
		return err
	}
	return os.RemoveAll(workspaceLayout.ToolchainsDirectory)
}

// VerifyToolchainInstallation deeply checks the private MSYS2 install by querying
// its package database (pacman -Qq) and comparing against the recorded package
// list. Read-only: it runs a fixed query command, never installs or downloads.
func (app *App) VerifyToolchainInstallation(workspaceDirectory string, windowsShellProfileName string) (ToolchainVerification, error) {
	verification := ToolchainVerification{MissingPackageNames: []string{}}
	if workspaceDirectory == "" {
		return verification, errors.New("workspace directory is empty")
	}
	msys2RootDirectory := msys2RootDirectoryFor(workspaceDirectory, windowsShellProfileName)
	if err := workspace.CheckRealPathInsideWorkspace(workspaceDirectory, msys2RootDirectory); err != nil {
		return verification, err
	}
	locale := app.currentLocale()
	if !fileExists(toolchainBashPath(msys2RootDirectory)) {
		verification.Message = localizeForLocale(locale, "verify.toolchain.notInstalled", nil)
		return verification, nil
	}
	manifest, err := readToolchainManifest(msys2RootDirectory)
	if err != nil || len(manifest.Msys2PackageNames) == 0 {
		verification.Message = localizeForLocale(locale, "verify.toolchain.noManifest", nil)
		return verification, nil
	}
	installedNames, err := queryInstalledPackageNames(msys2RootDirectory, manifest.WindowsShellProfileName)
	if err != nil {
		return verification, err
	}
	installedSet := make(map[string]bool, len(installedNames))
	for _, name := range installedNames {
		installedSet[name] = true
	}
	missingPackageNames := []string{}
	for _, expectedName := range manifest.Msys2PackageNames {
		if !installedSet[expectedName] {
			missingPackageNames = append(missingPackageNames, expectedName)
		}
	}
	sort.Strings(missingPackageNames)
	verification.CheckedPackageCount = len(manifest.Msys2PackageNames)
	verification.MissingPackageNames = missingPackageNames
	verification.Verified = len(missingPackageNames) == 0
	if verification.Verified {
		verification.Message = localizeForLocale(locale, "verify.toolchain.allPresent", map[string]string{"count": strconv.Itoa(verification.CheckedPackageCount)})
	} else {
		verification.Message = localizeForLocale(locale, "verify.toolchain.missing", map[string]string{"count": strconv.Itoa(len(missingPackageNames))})
	}
	return verification, nil
}

// queryInstalledPackageNames lists every package installed in the private MSYS2
// database. The package database is shared across shell profiles, so MSYSTEM is
// set only for correctness; the command and arguments are fixed (no user input).
func queryInstalledPackageNames(msys2RootDirectory string, windowsShellProfileName string) ([]string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	shellSystemName := strings.ToUpper(strings.TrimSpace(windowsShellProfileName))
	if shellSystemName == "" {
		shellSystemName = "UCRT64"
	}
	command := exec.CommandContext(ctx, toolchainBashPath(msys2RootDirectory), "-lc", "pacman -Qq")
	command.Dir = msys2RootDirectory
	command.Env = append(os.Environ(), "MSYSTEM="+shellSystemName, "MSYS2_PATH_TYPE=minimal", "CHERE_INVOKING=1")
	output, err := command.Output()
	if err != nil {
		return nil, fmt.Errorf("could not query installed packages with pacman -Qq: %w", err)
	}
	installedNames := []string{}
	for _, line := range strings.Split(string(output), "\n") {
		trimmedLine := strings.TrimSpace(line)
		if trimmedLine != "" {
			installedNames = append(installedNames, trimmedLine)
		}
	}
	return installedNames, nil
}
