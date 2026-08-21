package program

import (
	"debug/pe"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

func LArtifactFfmpegCopy(ffmpegSourceDirectory string, workspaceLayout workspace.LWorkspaceLayout, plan planning.LPlanFfmpeg, emitProgress func(string, string)) (retErr error) {
	// Publish into a staging directory and swap it into the artifact directory
	// only once the whole set is complete, so a failed copy cannot replace a
	// known-good result with a partial one.
	if err := LArtifactDirectoryValidate(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
		return err
	}
	stagingLayout := workspaceLayout
	stagingLayout.ArtifactsDirectory = workspaceLayout.ArtifactsDirectory + "-staging"
	if err := LFfmpegArtifactCheck(stagingLayout, emitProgress); err != nil {
		return err
	}
	defer func() {
		if retErr != nil {
			_ = os.RemoveAll(stagingLayout.ArtifactsDirectory)
		}
	}()

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
		if err := workspace.LPathRealCheck(stagingLayout.WorkspaceDirectory, sourcePath); err != nil {
			return err
		}
		destinationPath := filepath.Join(stagingLayout.ArtifactsDirectory, entry.Name())
		if err := LFileCopy(stagingLayout.WorkspaceDirectory, sourcePath, destinationPath); err != nil {
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
	builtSharedLibraryPaths, err := LDllFfmpegCopy(ffmpegSourceDirectory, stagingLayout, LPlanSharedCheck(plan), emitProgress)
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
	if err := workspace.LPathWorkspaceCheck(stagingLayout.WorkspaceDirectory, msys2BinDirectory); err != nil {
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
	if err := LDllDependencyCopy(rootPePaths, preResolvedNames, dllIndex, stagingLayout, emitProgress); err != nil {
		return err
	}
	return LArtifactStagingReplace(workspaceLayout, stagingLayout.ArtifactsDirectory, emitProgress)
}

// LArtifactDirectoryValidate confirms targetDirectory is a safe, in-workspace
// artifact directory that is not the workspace root, creating it if absent.
func LArtifactDirectoryValidate(workspaceDirectory string, targetDirectory string) error {
	if targetDirectory == "" {
		return errors.New("FFmpeg artifact directory is empty")
	}
	if err := workspace.LPathWorkspaceCheck(workspaceDirectory, targetDirectory); err != nil {
		return fmt.Errorf("FFmpeg artifact directory is outside workspace: %w", err)
	}
	workspaceAbsolutePath, err := filepath.Abs(workspaceDirectory)
	if err != nil {
		return err
	}
	artifactAbsolutePath, err := filepath.Abs(targetDirectory)
	if err != nil {
		return err
	}
	if artifactAbsolutePath == workspaceAbsolutePath {
		return errors.New("refusing to empty the workspace root as the FFmpeg artifact directory")
	}
	if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(workspaceDirectory, targetDirectory); err != nil {
		return err
	}
	return workspace.LDirectorySymlinkCheck(targetDirectory)
}

// LArtifactStagingReplace empties the real artifact directory and moves the
// completed staging entries into it. The previous result is destroyed only
// here, once the new set is known to be complete.
func LArtifactStagingReplace(workspaceLayout workspace.LWorkspaceLayout, stagingDirectory string, emitProgress func(string, string)) error {
	if err := LFfmpegArtifactCheck(workspaceLayout, emitProgress); err != nil {
		return err
	}
	entries, err := os.ReadDir(stagingDirectory)
	if err != nil {
		return err
	}
	for _, entry := range entries {
		destinationPath := filepath.Join(workspaceLayout.ArtifactsDirectory, entry.Name())
		if err := workspace.LPathWorkspaceCheck(workspaceLayout.WorkspaceDirectory, destinationPath); err != nil {
			return err
		}
		if err := os.Rename(filepath.Join(stagingDirectory, entry.Name()), destinationPath); err != nil {
			return fmt.Errorf("could not move staged artifact %s into place: %w", entry.Name(), err)
		}
	}
	if err := os.RemoveAll(stagingDirectory); err != nil {
		emitProgress("warn", fmt.Sprintf("Could not remove staging directory %s: %s", stagingDirectory, err.Error()))
	}
	emitProgress("info", fmt.Sprintf("Published %d artifact entries to %s.", len(entries), workspaceLayout.ArtifactsDirectory))
	return nil
}

// LDllFfmpegCopy copies the FFmpeg shared libraries produced by
// this build (libav*/libsw*/libpostproc*) from the build directory into the
// artifact directory. A shared, in-tree build places these DLLs in per-library
// subdirectories, so they are located with a recursive walk. The returned paths
// are the artifact-directory destinations.
func LDllFfmpegCopy(ffmpegSourceDirectory string, workspaceLayout workspace.LWorkspaceLayout, isSharedBuild bool, emitProgress func(string, string)) ([]string, error) {
	copiedByBaseName := map[string]bool{}
	var copiedPaths []string
	walkErr := filepath.WalkDir(ffmpegSourceDirectory, func(path string, dirEntry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if dirEntry.IsDir() {
			return nil
		}
		if !LDllFfmpegCheck(dirEntry.Name()) {
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
			// A shared build must bundle its own libav*/libsw* libraries; otherwise the
			// PE traversal resolves them from the MSYS2 bin index and ships stock FFmpeg.
			return nil, fmt.Errorf("shared FFmpeg build produced no libav*/libsw* shared libraries in %s: refusing to publish an artifact that would load stock MSYS2 FFmpeg libraries", ffmpegSourceDirectory)
		}
		emitProgress("info", "Static build: the FFmpeg libraries are linked into ffmpeg.exe/ffprobe.exe, so there are no separate shared libraries to bundle.")
	} else {
		emitProgress("info", fmt.Sprintf("Copied %d built FFmpeg shared libraries from the build directory.", len(copiedPaths)))
	}
	return copiedPaths, nil
}

// LPlanSharedCheck reports whether the plan configures a shared (DLL) FFmpeg
// build, i.e. its configure flags request --enable-shared. A static build (the default)
// links the FFmpeg libraries into the executables, so the absence of libav*/libsw* DLLs
// in the build directory is expected rather than a problem.
func LPlanSharedCheck(plan planning.LPlanFfmpeg) bool {
	return slices.Contains(plan.ConfigureFlags, "--enable-shared")
}

// LDllFfmpegCheck reports whether fileName is an FFmpeg shared library
// such as libavcodec-62.dll or avutil-60.dll. It accepts the name with or
// without the "lib" prefix and requires a numeric ABI suffix.
func LDllFfmpegCheck(fileName string) bool {
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

func LFfmpegArtifactCheck(workspaceLayout workspace.LWorkspaceLayout, emitProgress func(string, string)) error {
	if err := LArtifactDirectoryValidate(workspaceLayout.WorkspaceDirectory, workspaceLayout.ArtifactsDirectory); err != nil {
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
				if LDllSystemCheck(importedName) {
					emitProgress("info", fmt.Sprintf("  dependency %s: Windows system DLL, resolved by the OS", importedName))
					continue
				}
				// An import that is neither in the MSYS2 index nor a known system DLL is
				// an unresolved runtime dependency; failing keeps a broken artifact from exiting 0.
				return fmt.Errorf("unresolved runtime dependency %s imported by %s: not found in the MSYS2 bin index and not a known Windows system DLL", importedName, filepath.Base(currentPath))
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

// LDllSystemCheck reports whether name is a known Windows system DLL that the
// OS resolves at load time and must not be bundled. Imports that are neither in
// the MSYS2 index nor recognized here are treated as unresolved dependencies.
func LDllSystemCheck(name string) bool {
	lower := strings.ToLower(name)
	if strings.HasPrefix(lower, "api-ms-win-") || strings.HasPrefix(lower, "ext-ms-win-") {
		return true
	}
	switch lower {
	case "ntdll.dll", "kernel32.dll", "kernelbase.dll", "user32.dll", "gdi32.dll",
		"gdi32full.dll", "advapi32.dll", "shell32.dll", "shlwapi.dll", "ole32.dll",
		"oleaut32.dll", "combase.dll", "rpcrt4.dll", "sechost.dll", "ws2_32.dll",
		"wsock32.dll", "crypt32.dll", "secur32.dll", "bcrypt.dll", "bcryptprimitives.dll",
		"ncrypt.dll", "msvcrt.dll", "ucrtbase.dll", "vcruntime140.dll", "psapi.dll",
		"version.dll", "winmm.dll", "imm32.dll", "comdlg32.dll", "comctl32.dll",
		"setupapi.dll", "cfgmgr32.dll", "iphlpapi.dll", "dnsapi.dll", "userenv.dll",
		"netapi32.dll", "mpr.dll", "powrprof.dll", "dwmapi.dll", "uxtheme.dll",
		"wintrust.dll", "wldap32.dll", "hid.dll", "dwrite.dll", "windowscodecs.dll",
		"d3d9.dll", "d3d11.dll", "d3d12.dll", "dxgi.dll", "d2d1.dll", "opengl32.dll",
		"gdiplus.dll", "mf.dll", "mfplat.dll", "mfreadwrite.dll", "mfuuid.dll",
		"strmiids.dll", "avicap32.dll", "msvfw32.dll", "winspool.drv", "ksuser.dll",
		"avrt.dll", "dxva2.dll", "evr.dll", "d3dcompiler_47.dll", "normaliz.dll",
		"winhttp.dll", "wininet.dll", "urlmon.dll", "propsys.dll", "shcore.dll":
		return true
	}
	return false
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
