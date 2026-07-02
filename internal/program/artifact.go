package program

import (
	"crypto/sha256"
	"debug/pe"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

func LHashToolchainVerify(plan planning.LPlanToolchain) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.LPlanHashCreate(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("toolchain plan hash does not match plan content")
	}
	return nil
}

func LHashFFmpegVerify(plan planning.LPlanFFmpeg) error {
	planWithoutHash := plan
	originalPlanHash := planWithoutHash.PlanHash
	planWithoutHash.PlanHash = ""
	computedPlanHash, err := planning.LPlanHashCreate(planWithoutHash)
	if err != nil {
		return err
	}
	if computedPlanHash != originalPlanHash {
		return errors.New("FFmpeg build plan hash does not match plan content")
	}
	return nil
}

func LDirectoryChildFind(parentDirectory string) (string, error) {
	entries, err := os.ReadDir(parentDirectory)
	if err != nil {
		return "", err
	}
	childDirectories := make([]string, 0)
	for _, entry := range entries {
		if entry.IsDir() {
			childDirectories = append(childDirectories, filepath.Join(parentDirectory, entry.Name()))
		}
	}
	if len(childDirectories) != 1 {
		return "", fmt.Errorf("expected exactly one extracted source directory, found %d", len(childDirectories))
	}
	return childDirectories[0], nil
}

func LArtifactFFmpegCopy(ffmpegSourceDirectory string, workspaceLayout workspace.LWorkspaceLayout, plan planning.LPlanFFmpeg, emitProgress func(string, string)) error {
	if err := LArtifactFFmpegEmpty(workspaceLayout, emitProgress); err != nil {
		return err
	}

	// Step 1: copy the built executables from the FFmpeg source directory.
	exeEntries, err := os.ReadDir(ffmpegSourceDirectory)
	if err != nil {
		return fmt.Errorf("could not read FFmpeg build directory: %w", err)
	}
	var copiedExePaths []string
	for _, entry := range exeEntries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".exe") {
			continue
		}
		sourcePath := filepath.Join(ffmpegSourceDirectory, entry.Name())
		if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, sourcePath); err != nil {
			return err
		}
		destinationPath := filepath.Join(workspaceLayout.ArtifactsDirectory, entry.Name())
		if err := LFileCopy(workspaceLayout.WorkspaceDirectory, sourcePath, destinationPath); err != nil {
			return err
		}
		copiedExePaths = append(copiedExePaths, destinationPath)
	}
	if len(copiedExePaths) == 0 {
		return fmt.Errorf("no .exe files found in FFmpeg build directory: %s", ffmpegSourceDirectory)
	}

	// Step 1b: copy this build's own FFmpeg shared libraries (libav*/libsw*/
	// libpostproc*). A shared, in-tree build emits these into per-library
	// subdirectories rather than next to the executables, so the .exe-only copy
	// above misses them. Without this step the PE traversal below resolves
	// libavcodec-*.dll and friends from the MSYS2 bin index, bundling the stock
	// MSYS2 FFmpeg libraries whose configure differs from this build. That mixes
	// in libraries that omit selected encoders (for example libfdk_aac,
	// libopenh264, libilbc, libtwolame, libvo_amrwbenc) and triggers FFmpeg's
	// "library configuration mismatch" warning at runtime.
	builtSharedLibraryPaths, err := LDllFFmpegCopy(ffmpegSourceDirectory, workspaceLayout, LPlanSharedCheck(plan), emitProgress)
	if err != nil {
		return err
	}

	// Step 2: build a lookup index for MSYS2 DLLs. This does not copy them. Only DLLs reached from the PE dependency traversal below are copied.
	profileDirectoryName := strings.ToLower(plan.WindowsShellProfileName)
	if profileDirectoryName == "" {
		profileDirectoryName = "ucrt64"
	}
	msys2BinDirectory := filepath.Join(plan.Msys2RootDirectory, profileDirectoryName, "bin")
	// String-only workspace check ??LPathRealCheck uses
	// filepath.EvalSymlinks which fails on MSYS2's internal reparse points.
	// The path is trusted by construction: Msys2RootDirectory is always
	// planning.LDirectoryProfileResolve(workspaceDirectory, profile), i.e.
	// workspaceDirectory/toolchains/msys2-<profile>.
	if err := workspace.LPathWorkspaceCheck(workspaceLayout.WorkspaceDirectory, msys2BinDirectory); err != nil {
		return fmt.Errorf("MSYS2 bin directory is outside workspace: %w", err)
	}
	dllIndex, err := LDllIndexBuild(msys2BinDirectory)
	if err != nil {
		return err
	}
	emitProgress("info", fmt.Sprintf("DLL lookup index built: %d DLLs indexed in %s", len(dllIndex), msys2BinDirectory))

	// Step 3: BFS from the copied exes and this build's own shared libraries
	// through their PE import tables, resolving only DLLs present in the MSYS2
	// bin index. The built shared libraries are pre-marked as resolved so their
	// names are never re-copied from the MSYS2 bin index (which would replace
	// them with the stock MSYS2 build), while still being traversed so their own
	// dependencies (the selected encoder backends) are pulled in.
	rootPePaths := append(append([]string{}, copiedExePaths...), builtSharedLibraryPaths...)
	preResolvedNames := make([]string, 0, len(builtSharedLibraryPaths))
	for _, builtPath := range builtSharedLibraryPaths {
		preResolvedNames = append(preResolvedNames, filepath.Base(builtPath))
	}
	return LDllDependencyCopy(rootPePaths, preResolvedNames, dllIndex, workspaceLayout, emitProgress)
}

// LDllFFmpegCopy copies the FFmpeg shared libraries produced by
// this build (libav*/libsw*/libpostproc*) from the build directory into the
// artifact directory. A shared, in-tree build places these DLLs in per-library
// subdirectories, so they are located with a recursive walk. The returned paths
// are the artifact-directory destinations.
func LDllFFmpegCopy(ffmpegSourceDirectory string, workspaceLayout workspace.LWorkspaceLayout, isSharedBuild bool, emitProgress func(string, string)) ([]string, error) {
	copiedByBaseName := map[string]bool{}
	var copiedPaths []string
	walkErr := filepath.WalkDir(ffmpegSourceDirectory, func(path string, dirEntry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if dirEntry.IsDir() {
			return nil
		}
		if !LDllFFmpegCheck(dirEntry.Name()) {
			return nil
		}
		baseNameLower := strings.ToLower(dirEntry.Name())
		if copiedByBaseName[baseNameLower] {
			return nil
		}
		if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, path); err != nil {
			return err
		}
		destinationPath := filepath.Join(workspaceLayout.ArtifactsDirectory, dirEntry.Name())
		if err := LFileCopy(workspaceLayout.WorkspaceDirectory, path, destinationPath); err != nil {
			return err
		}
		copiedByBaseName[baseNameLower] = true
		copiedPaths = append(copiedPaths, destinationPath)
		emitProgress("info", fmt.Sprintf("built FFmpeg shared library %s: copied OK -> %s", dirEntry.Name(), destinationPath))
		return nil
	})
	if walkErr != nil {
		return nil, fmt.Errorf("could not collect built FFmpeg shared libraries from %s: %w", ffmpegSourceDirectory, walkErr)
	}
	if len(copiedPaths) == 0 {
		if isSharedBuild {
			emitProgress("warn", "No built FFmpeg shared libraries (libav*/libsw*) were found in the build directory. The artifact may load stock MSYS2 FFmpeg libraries instead, which can omit selected encoders.")
		} else {
			emitProgress("info", "Static build: the FFmpeg libraries are linked into ffmpeg.exe/ffprobe.exe, so there are no separate shared libraries to bundle.")
		}
	} else {
		emitProgress("info", fmt.Sprintf("Copied %d built FFmpeg shared libraries from the build directory.", len(copiedPaths)))
	}
	return copiedPaths, nil
}

// LPlanSharedCheck reports whether the plan configures a shared (DLL) FFmpeg
// build, i.e. its configure flags request --enable-shared. A static build (the default)
// links the FFmpeg libraries into the executables, so the absence of libav*/libsw* DLLs
// in the build directory is expected rather than a problem.
func LPlanSharedCheck(plan planning.LPlanFFmpeg) bool {
	for _, configureFlag := range plan.ConfigureFlags {
		if configureFlag == "--enable-shared" {
			return true
		}
	}
	return false
}

// LDllFFmpegCheck reports whether fileName is an FFmpeg shared library
// such as libavcodec-62.dll or avutil-60.dll. It accepts the name with or
// without the "lib" prefix and requires a numeric ABI suffix.
func LDllFFmpegCheck(fileName string) bool {
	lowerName := strings.ToLower(fileName)
	if !strings.HasSuffix(lowerName, ".dll") {
		return false
	}
	base := strings.TrimSuffix(lowerName, ".dll")
	base = strings.TrimPrefix(base, "lib")
	stems := []string{"avutil", "avcodec", "avformat", "avdevice", "avfilter", "swscale", "swresample", "postproc"}
	for _, stem := range stems {
		if !strings.HasPrefix(base, stem+"-") {
			continue
		}
		suffix := base[len(stem)+1:]
		if suffix == "" {
			return false
		}
		for _, character := range suffix {
			if character < '0' || character > '9' {
				return false
			}
		}
		return true
	}
	return false
}

func LArtifactFFmpegEmpty(workspaceLayout workspace.LWorkspaceLayout, emitProgress func(string, string)) error {
	if workspaceLayout.ArtifactsDirectory == "" {
		return errors.New("FFmpeg artifact directory is empty")
	}
	if err := workspace.LPathWorkspaceCheck(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return fmt.Errorf("FFmpeg artifact directory is outside workspace: %w", err)
	}
	workspaceAbsolutePath, err := filepath.Abs(workspaceLayout.WorkspaceDirectory)
	if err != nil {
		return err
	}
	artifactAbsolutePath, err := filepath.Abs(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return err
	}
	if artifactAbsolutePath == workspaceAbsolutePath {
		return errors.New("refusing to empty the workspace root as the FFmpeg artifact directory")
	}
	if err := os.MkdirAll(workspaceLayout.ArtifactsDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return err
	}
	if err := workspace.LDirectorySymlinkCheck(workspaceLayout.ArtifactsDirectory); err != nil {
		return err
	}
	entries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return err
	}
	emitProgress("info", fmt.Sprintf("Emptying FFmpeg artifact directory before copying the new build: %s", workspaceLayout.ArtifactsDirectory))
	removedEntryCount := 0
	for _, entry := range entries {
		entryPath := filepath.Join(workspaceLayout.ArtifactsDirectory, entry.Name())
		if err := workspace.LPathWorkspaceCheck(workspaceLayout.WorkspaceDirectory, entryPath); err != nil {
			return err
		}
		if err := os.RemoveAll(entryPath); err != nil {
			return fmt.Errorf("could not remove stale FFmpeg artifact %s: %w", entryPath, err)
		}
		removedEntryCount++
	}
	emitProgress("info", fmt.Sprintf("Removed %d stale FFmpeg artifact entries before copying the new build.", removedEntryCount))
	return nil
}

func LDllIndexBuild(msys2BinDirectory string) (map[string]string, error) {
	entries, err := os.ReadDir(msys2BinDirectory)
	if err != nil {
		return nil, fmt.Errorf("could not read MSYS2 bin directory: %w", err)
	}
	index := make(map[string]string, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if !strings.HasSuffix(strings.ToLower(entry.Name()), ".dll") {
			continue
		}
		index[strings.ToLower(entry.Name())] = filepath.Join(msys2BinDirectory, entry.Name())
	}
	return index, nil
}

func LDllDependencyCopy(rootPePaths []string, preResolvedNames []string, dllIndex map[string]string, workspaceLayout workspace.LWorkspaceLayout, emitProgress func(string, string)) error {
	resolved := map[string]bool{}
	// Names already present in the artifact directory (this build's own shared
	// libraries). Marking them resolved prevents the traversal from copying a
	// same-named DLL out of the MSYS2 bin index over the build's own library.
	for _, preResolvedName := range preResolvedNames {
		resolved[strings.ToLower(preResolvedName)] = true
	}
	queue := append([]string{}, rootPePaths...)
	visited := map[string]bool{}

	for len(queue) > 0 {
		currentPath := queue[0]
		queue = queue[1:]
		currentKey := strings.ToLower(filepath.Base(currentPath))
		if visited[currentKey] {
			continue
		}
		visited[currentKey] = true

		peFile, err := pe.Open(currentPath)
		if err != nil {
			// Log and skip files that cannot be parsed as PE.
			emitProgress("warn", fmt.Sprintf("Skipping PE inspection for %s: %s", filepath.Base(currentPath), err.Error()))
			continue
		}
		// ImportedLibraries is unimplemented in Go's debug/pe.
		// ImportedSymbols returns strings in the form "FunctionName:DLLNAME.dll".
		// Extract the library side and use it only as a dependency name.
		symbols, err := peFile.ImportedSymbols()
		peFile.Close()
		if err != nil {
			emitProgress("warn", fmt.Sprintf("Could not read PE symbols for %s: %s", filepath.Base(currentPath), err.Error()))
			continue
		}
		imports := LDllSymbolRead(symbols)
		emitProgress("info", fmt.Sprintf("PE DLL dependencies for %s: %d DLLs found", filepath.Base(currentPath), len(imports)))

		for _, importedName := range imports {
			nameLower := strings.ToLower(importedName)
			if resolved[nameLower] {
				emitProgress("info", fmt.Sprintf("  dependency %s: already copied, skipping", importedName))
				continue
			}
			sourcePath, found := dllIndex[nameLower]
			if !found {
				emitProgress("info", fmt.Sprintf("  dependency %s: not in MSYS2 bin (system DLL or statically linked)", importedName))
				continue
			}
			emitProgress("info", fmt.Sprintf("  dependency %s: found at %s", importedName, sourcePath))
			destinationPath := filepath.Join(workspaceLayout.ArtifactsDirectory, importedName)
			if err := LDllMsysCopy(workspaceLayout.WorkspaceDirectory, sourcePath, destinationPath); err != nil {
				emitProgress("warn", fmt.Sprintf("  dependency %s: copy FAILED: %s", importedName, err.Error()))
				return err
			}
			emitProgress("info", fmt.Sprintf("  dependency %s: copied OK -> %s", importedName, destinationPath))
			resolved[nameLower] = true
			// A copied DLL can itself import other MSYS2 DLLs. Queue it so the
			// artifact directory receives the complete required dependency closure,
			// not every DLL from MSYS2.
			queue = append(queue, destinationPath)
		}
	}
	return nil
}

func LDllSymbolRead(symbols []string) []string {
	seenDlls := map[string]bool{}
	dllNames := []string{}
	for _, symbol := range symbols {
		parts := strings.SplitN(symbol, ":", 2)
		if len(parts) != 2 {
			continue
		}
		dllName := strings.TrimSpace(parts[1])
		if dllName == "" || !strings.HasSuffix(strings.ToLower(dllName), ".dll") {
			continue
		}
		dllNameLower := strings.ToLower(dllName)
		if seenDlls[dllNameLower] {
			continue
		}
		seenDlls[dllNameLower] = true
		dllNames = append(dllNames, dllName)
	}
	sort.Slice(dllNames, func(leftIndex, rightIndex int) bool {
		return strings.ToLower(dllNames[leftIndex]) < strings.ToLower(dllNames[rightIndex])
	})
	return dllNames
}

// LDllMsysCopy copies a DLL from the MSYS2 bin directory to the destination.
// It uses LPathWorkspaceCheck (string-only) for the source instead of
// LPathRealCheck because filepath.EvalSymlinks fails on MSYS2's
// internal reparse points on Windows (see golang/go#63703).
// The destination still uses the full real-path check.
func LDllMsysCopy(workspaceDirectory string, sourcePath string, destinationPath string) error {
	if err := workspace.LPathWorkspaceCheck(workspaceDirectory, sourcePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, filepath.Dir(destinationPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, destinationPath); err != nil {
		return err
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destinationFile, sourceFile)
	closeErr := destinationFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func LFileCopy(workspaceDirectory string, sourcePath string, destinationPath string) error {
	if err := workspace.LPathRealCheck(workspaceDirectory, sourcePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, filepath.Dir(destinationPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(destinationPath), 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, destinationPath); err != nil {
		return err
	}
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return err
	}
	defer sourceFile.Close()
	destinationFile, err := os.OpenFile(destinationPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o700)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(destinationFile, sourceFile)
	closeErr := destinationFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func LReportArtifactWrite(workspaceLayout workspace.LWorkspaceLayout, LRunId string, plan planning.LPlanFFmpeg) error {
	reportPath := filepath.Join(workspaceLayout.ArtifactsDirectory, "build-report-"+LRunId+".json")
	ffmpegExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffmpeg.exe")
	ffprobeExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffprobe.exe")
	report := map[string]interface{}{"runId": LRunId, "createdAt": time.Now().UTC().Format(time.RFC3339), "approvedPlanHash": plan.PlanHash, "ffmpegVersion": planning.LVersionArchiveParse(plan.FfmpegSourceArchiveUrl), "ffmpegSourceArchiveUrl": plan.FfmpegSourceArchiveUrl, "ffmpegSourceSignatureUrl": plan.FfmpegSourceSignatureUrl, "ffmpegSourceSha256Hash": plan.FfmpegSourceSha256Hash, "selectedLibraries": plan.SelectedLibraries, "selectedConfigureOptions": plan.SelectedConfigureOptions, "requiredMsys2PackageNames": plan.RequiredMsys2PackageNames, "generatedConfigureFlags": plan.GeneratedConfigureFlags, "generatedOptionFlags": plan.GeneratedOptionFlags, "extraConfigureFlags": plan.ExtraConfigureFlags, "configureFlags": plan.ConfigureFlags, "licenseProfileName": plan.LicenseProfileName, "ffmpegExecutablePath": ffmpegExecutablePath, "ffmpegExecutableSha256Hash": LHashFileCreate(ffmpegExecutablePath), "ffmpegExecutableSizeBytes": LFileSizeRead(ffmpegExecutablePath), "ffprobeExecutablePath": ffprobeExecutablePath, "ffprobeExecutableSha256Hash": LHashFileCreate(ffprobeExecutablePath), "ffprobeExecutableSizeBytes": LFileSizeRead(ffprobeExecutablePath), "artifactFiles": LFileArtifactList(workspaceLayout)}
	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, filepath.Dir(reportPath)); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, reportPath); err != nil {
		return err
	}
	return os.WriteFile(reportPath, reportBytes, 0o600)
}

func LArtifactVersionRead(workspaceLayout workspace.LWorkspaceLayout) string {
	ffmpegExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffmpeg.exe")
	if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, ffmpegExecutablePath); err != nil {
		return ""
	}
	versionOutput, err := LFFmpegVersionRun(ffmpegExecutablePath)
	if err != nil {
		return ""
	}
	return LVersionFFmpegParse(versionOutput)
}

// LArtifactLatestLayoutFind returns the workspace layout whose ArtifactsDirectory
// holds the most recently written build report. Builds are stored per FFmpeg
// version under <workspace>/FFmpeg/<version>/, so this scans each version
// subdirectory (and, for backward compatibility with pre-versioning builds, the
// FFmpeg base directory itself) and selects the directory with the newest
// build-report-*.json. When no report exists anywhere it returns the base layout,
// so callers still get a valid (empty) artifacts directory to report on.
func LArtifactLatestLayoutFind(workspaceDirectory string) workspace.LWorkspaceLayout {
	baseLayout := workspace.LWorkspaceLayoutResolve(workspaceDirectory)
	candidateDirectories := []string{baseLayout.ArtifactsBaseDirectory}
	if baseEntries, err := os.ReadDir(baseLayout.ArtifactsBaseDirectory); err == nil {
		for _, baseEntry := range baseEntries {
			if baseEntry.IsDir() {
				candidateDirectories = append(candidateDirectories, filepath.Join(baseLayout.ArtifactsBaseDirectory, baseEntry.Name()))
			}
		}
	}
	bestDirectory := ""
	var bestModTime time.Time
	for _, candidateDirectory := range candidateDirectories {
		entries, err := os.ReadDir(candidateDirectory)
		if err != nil {
			continue
		}
		for _, entry := range entries {
			if entry.IsDir() || !strings.HasPrefix(entry.Name(), "build-report-") || !strings.HasSuffix(entry.Name(), ".json") {
				continue
			}
			info, err := entry.Info()
			if err != nil {
				continue
			}
			if bestDirectory == "" || info.ModTime().After(bestModTime) {
				bestDirectory = candidateDirectory
				bestModTime = info.ModTime()
			}
		}
	}
	layout := baseLayout
	if bestDirectory != "" {
		layout.ArtifactsDirectory = bestDirectory
	}
	return layout
}

func LReportLatestRead(workspaceLayout workspace.LWorkspaceLayout) (string, LReportArtifact, error) {
	entries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return "", LReportArtifact{}, err
	}
	latestPath := ""
	var latestModTime time.Time
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasPrefix(entry.Name(), "build-report-") || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}
		candidatePath := filepath.Join(workspaceLayout.ArtifactsDirectory, entry.Name())
		if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, candidatePath); err != nil {
			return "", LReportArtifact{}, err
		}
		info, err := entry.Info()
		if err != nil {
			return "", LReportArtifact{}, err
		}
		if latestPath == "" || info.ModTime().After(latestModTime) {
			latestPath = candidatePath
			latestModTime = info.ModTime()
		}
	}
	if latestPath == "" {
		return "", LReportArtifact{}, errors.New("no build report found")
	}
	reportBytes, err := os.ReadFile(latestPath)
	if err != nil {
		return "", LReportArtifact{}, err
	}
	var report LReportArtifact
	if err := json.Unmarshal(reportBytes, &report); err != nil {
		return "", LReportArtifact{}, err
	}
	return latestPath, report, nil
}

func LFileSizeRead(filePath string) int64 {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return 0
	}
	return fileInfo.Size()
}

func LFileArtifactList(workspaceLayout workspace.LWorkspaceLayout) []LFileResult {
	artifactFiles := []LFileResult{}
	entries, err := os.ReadDir(workspaceLayout.ArtifactsDirectory)
	if err != nil {
		return artifactFiles
	}
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		artifactName := entry.Name()
		artifactNameLower := strings.ToLower(artifactName)
		if artifactNameLower != "ffmpeg.exe" && artifactNameLower != "ffprobe.exe" && !strings.HasSuffix(artifactNameLower, ".dll") {
			continue
		}
		artifactPath := filepath.Join(workspaceLayout.ArtifactsDirectory, artifactName)
		if err := workspace.LPathRealCheck(workspaceLayout.WorkspaceDirectory, artifactPath); err != nil {
			continue
		}
		fileInfo, err := os.Stat(artifactPath)
		if err != nil {
			continue
		}
		artifactFiles = append(artifactFiles, LFileResult{Name: artifactName, Path: artifactPath, SizeBytes: fileInfo.Size(), Sha256Hash: LHashFileCreate(artifactPath)})
	}
	sort.Slice(artifactFiles, func(leftIndex, rightIndex int) bool {
		return strings.ToLower(artifactFiles[leftIndex].Name) < strings.ToLower(artifactFiles[rightIndex].Name)
	})
	return artifactFiles
}

func LHashFileCreate(filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		return ""
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return ""
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func LFileExistCheck(filePath string) bool {
	info, err := os.Stat(filePath)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

func LTextErrorTrim(value string) string {
	value = strings.TrimSpace(value)
	if len(value) > 600 {
		return value[:600] + "..."
	}
	return value
}
