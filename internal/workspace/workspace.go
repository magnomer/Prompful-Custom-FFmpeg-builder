package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// LPatternArtifactVersion matches the dotted-numeric FFmpeg release used as the
// per-version artifact subdirectory name (e.g. "8.1.2"). Anything else is
// rejected so a malformed version can never introduce path separators or ".."
// into the artifact path.
var LPatternArtifactVersion = regexp.MustCompile(`^\d+(?:\.\d+){0,2}$`)

type LWorkspaceLayout struct {
	WorkspaceDirectory  string `json:"workspaceDirectory"`
	CacheDirectory      string `json:"cacheDirectory"`
	DownloadsDirectory  string `json:"downloadsDirectory"`
	SourcesDirectory    string `json:"sourcesDirectory"`
	BuildDirectory      string `json:"buildDirectory"`
	PrefixDirectory     string `json:"prefixDirectory"`
	// ArtifactsBaseDirectory is the parent that holds one per-version subdirectory
	// per built FFmpeg release (e.g. <workspace>/FFmpeg). ArtifactsDirectory is the
	// directory a specific build's files live in: the version subdirectory for a
	// versioned resolve, or the base itself when no version is known.
	ArtifactsBaseDirectory string `json:"artifactsBaseDirectory"`
	ArtifactsDirectory     string `json:"artifactsDirectory"`
	LogsDirectory          string `json:"logsDirectory"`
	ToolchainsDirectory    string `json:"toolchainsDirectory"`
}

func LWorkspaceLayoutResolve(workspaceDirectory string) LWorkspaceLayout {
	artifactsBaseDirectory := filepath.Join(workspaceDirectory, "FFmpeg")
	return LWorkspaceLayout{
		WorkspaceDirectory:     workspaceDirectory,
		CacheDirectory:         filepath.Join(workspaceDirectory, "cache"),
		DownloadsDirectory:     filepath.Join(workspaceDirectory, "cache", "downloads"),
		SourcesDirectory:       filepath.Join(workspaceDirectory, "sources"),
		BuildDirectory:         filepath.Join(workspaceDirectory, "build"),
		PrefixDirectory:        filepath.Join(workspaceDirectory, "prefix"),
		ArtifactsBaseDirectory: artifactsBaseDirectory,
		ArtifactsDirectory:     artifactsBaseDirectory,
		LogsDirectory:          filepath.Join(workspaceDirectory, "logs"),
		ToolchainsDirectory:    filepath.Join(workspaceDirectory, "toolchains"),
	}
}

// LWorkspaceLayoutResolveVersioned resolves a layout whose ArtifactsDirectory is
// the per-version subdirectory <workspace>/FFmpeg/<version>. A build stores its
// executables, shared libraries, and build report there so releases no longer
// overwrite one another. If the version is not a valid dotted-numeric release the
// artifact path falls back to the base FFmpeg directory rather than risk an
// unsafe subdirectory name.
func LWorkspaceLayoutResolveVersioned(workspaceDirectory string, version string) LWorkspaceLayout {
	layout := LWorkspaceLayoutResolve(workspaceDirectory)
	if LPatternArtifactVersion.MatchString(version) {
		layout.ArtifactsDirectory = filepath.Join(layout.ArtifactsBaseDirectory, version)
	}
	return layout
}

func LWorkspaceFolderCreate(workspaceLayout LWorkspaceLayout) error {
	directoryPaths := []string{
		workspaceLayout.WorkspaceDirectory,
		workspaceLayout.CacheDirectory,
		workspaceLayout.DownloadsDirectory,
		workspaceLayout.SourcesDirectory,
		workspaceLayout.BuildDirectory,
		workspaceLayout.PrefixDirectory,
		workspaceLayout.ArtifactsDirectory,
		workspaceLayout.LogsDirectory,
		workspaceLayout.ToolchainsDirectory,
	}
	for _, directoryPath := range directoryPaths {
		if err := os.MkdirAll(directoryPath, 0o755); err != nil {
			return err
		}
		if err := LPathRealCheck(workspaceLayout.WorkspaceDirectory, directoryPath); err != nil {
			return err
		}
	}
	return nil
}

func LPathWorkspaceCheck(workspaceDirectory string, candidatePath string) error {
	absoluteWorkspaceDirectory, err := filepath.Abs(workspaceDirectory)
	if err != nil {
		return err
	}
	absoluteCandidatePath, err := filepath.Abs(candidatePath)
	if err != nil {
		return err
	}
	relativePath, err := filepath.Rel(absoluteWorkspaceDirectory, absoluteCandidatePath)
	if err != nil {
		return err
	}
	if relativePath == "." {
		return nil
	}
	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return errors.New("path escapes selected workspace")
	}
	return nil
}

// LPathRealCheck verifies both the string path and the existing
// filesystem object or nearest existing parent after symlink/reparse resolution.
// Use this before reads, writes, renames, directory creation, and command execution.
func LPathRealCheck(workspaceDirectory string, candidatePath string) error {
	if workspaceDirectory == "" || candidatePath == "" {
		return errors.New("workspace and candidate paths must not be empty")
	}
	if err := LPathWorkspaceCheck(workspaceDirectory, candidatePath); err != nil {
		return err
	}

	absoluteWorkspaceDirectory, err := filepath.Abs(workspaceDirectory)
	if err != nil {
		return err
	}
	workspaceRealPath, err := filepath.EvalSymlinks(absoluteWorkspaceDirectory)
	if err != nil {
		return err
	}

	LPathParentFind, err := LPathParentFind(candidatePath)
	if err != nil {
		return err
	}
	if err := LPathSymlinkReject(workspaceRealPath, LPathParentFind); err != nil {
		return err
	}
	candidateRealPath, err := filepath.EvalSymlinks(LPathParentFind)
	if err != nil {
		return err
	}
	return LPathBaseInsideCheck(workspaceRealPath, candidateRealPath, "resolved path escapes selected workspace")
}

func LDirectorySymlinkCheck(directoryPath string) error {
	fileInfo, err := os.Lstat(directoryPath)
	if err != nil {
		return err
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("directory must not be a symlink: %s", directoryPath)
	}
	if !fileInfo.IsDir() {
		return fmt.Errorf("path is not a directory: %s", directoryPath)
	}
	return nil
}

func LPathSymlinkReject(basePath string, candidatePath string) error {
	absoluteBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	absoluteCandidatePath, err := filepath.Abs(candidatePath)
	if err != nil {
		return err
	}
	if err := LPathBaseInsideCheck(absoluteBasePath, absoluteCandidatePath, "path escapes selected workspace"); err != nil {
		return err
	}
	relativePath, err := filepath.Rel(absoluteBasePath, absoluteCandidatePath)
	if err != nil {
		return err
	}
	if relativePath == "." {
		return LDirectorySymlinkCheck(absoluteBasePath)
	}
	currentPath := absoluteBasePath
	components := strings.Split(relativePath, string(os.PathSeparator))
	for _, component := range components {
		if component == "" || component == "." {
			continue
		}
		currentPath = filepath.Join(currentPath, component)
		fileInfo, err := os.Lstat(currentPath)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		if err != nil {
			return err
		}
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("path component must not be a symlink: %s", currentPath)
		}
	}
	return nil
}

func LPathParentFind(candidatePath string) (string, error) {
	absoluteCandidatePath, err := filepath.Abs(candidatePath)
	if err != nil {
		return "", err
	}
	currentPath := absoluteCandidatePath
	for {
		_, statError := os.Lstat(currentPath)
		if statError == nil {
			return currentPath, nil
		}
		if !errors.Is(statError, os.ErrNotExist) {
			return "", statError
		}
		parentPath := filepath.Dir(currentPath)
		if parentPath == currentPath {
			return "", os.ErrNotExist
		}
		currentPath = parentPath
	}
}

func LPathBaseInsideCheck(basePath string, candidatePath string, errorMessage string) error {
	absoluteBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	absoluteCandidatePath, err := filepath.Abs(candidatePath)
	if err != nil {
		return err
	}
	relativePath, err := filepath.Rel(absoluteBasePath, absoluteCandidatePath)
	if err != nil {
		return err
	}
	if relativePath == "." {
		return nil
	}
	if strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return errors.New(errorMessage)
	}
	return nil
}
