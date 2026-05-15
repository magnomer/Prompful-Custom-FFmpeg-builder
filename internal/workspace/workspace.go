package workspace

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type WorkspaceLayout struct {
	WorkspaceDirectory  string `json:"workspaceDirectory"`
	CacheDirectory      string `json:"cacheDirectory"`
	DownloadsDirectory  string `json:"downloadsDirectory"`
	SourcesDirectory    string `json:"sourcesDirectory"`
	BuildDirectory      string `json:"buildDirectory"`
	PrefixDirectory     string `json:"prefixDirectory"`
	ArtifactsDirectory  string `json:"artifactsDirectory"`
	LogsDirectory       string `json:"logsDirectory"`
	ToolchainsDirectory string `json:"toolchainsDirectory"`
}

func WorkspaceLayoutFor(workspaceDirectory string) WorkspaceLayout {
	return WorkspaceLayout{
		WorkspaceDirectory:  workspaceDirectory,
		CacheDirectory:      filepath.Join(workspaceDirectory, "cache"),
		DownloadsDirectory:  filepath.Join(workspaceDirectory, "cache", "downloads"),
		SourcesDirectory:    filepath.Join(workspaceDirectory, "sources"),
		BuildDirectory:      filepath.Join(workspaceDirectory, "build"),
		PrefixDirectory:     filepath.Join(workspaceDirectory, "prefix"),
		ArtifactsDirectory:  filepath.Join(workspaceDirectory, "artifacts"),
		LogsDirectory:       filepath.Join(workspaceDirectory, "logs"),
		ToolchainsDirectory: filepath.Join(workspaceDirectory, "toolchains"),
	}
}

func CreateWorkspaceFolders(workspaceLayout WorkspaceLayout) error {
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
		if err := CheckRealPathInsideWorkspace(workspaceLayout.WorkspaceDirectory, directoryPath); err != nil {
			return err
		}
	}
	return nil
}

func CheckPathInsideWorkspace(workspaceDirectory string, candidatePath string) error {
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

// CheckRealPathInsideWorkspace verifies both the string path and the existing
// filesystem object or nearest existing parent after symlink/reparse resolution.
// Use this before reads, writes, renames, directory creation, and command execution.
func CheckRealPathInsideWorkspace(workspaceDirectory string, candidatePath string) error {
	if workspaceDirectory == "" || candidatePath == "" {
		return errors.New("workspace and candidate paths must not be empty")
	}
	if err := CheckPathInsideWorkspace(workspaceDirectory, candidatePath); err != nil {
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

	nearestExistingParent, err := nearestExistingParent(candidatePath)
	if err != nil {
		return err
	}
	if err := RejectSymlinkComponents(workspaceRealPath, nearestExistingParent); err != nil {
		return err
	}
	candidateRealPath, err := filepath.EvalSymlinks(nearestExistingParent)
	if err != nil {
		return err
	}
	return checkPathInsideBase(workspaceRealPath, candidateRealPath, "resolved path escapes selected workspace")
}

func CheckDirectoryNotSymlink(directoryPath string) error {
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

func RejectSymlinkComponents(basePath string, candidatePath string) error {
	absoluteBasePath, err := filepath.Abs(basePath)
	if err != nil {
		return err
	}
	absoluteCandidatePath, err := filepath.Abs(candidatePath)
	if err != nil {
		return err
	}
	if err := checkPathInsideBase(absoluteBasePath, absoluteCandidatePath, "path escapes selected workspace"); err != nil {
		return err
	}
	relativePath, err := filepath.Rel(absoluteBasePath, absoluteCandidatePath)
	if err != nil {
		return err
	}
	if relativePath == "." {
		return CheckDirectoryNotSymlink(absoluteBasePath)
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

func nearestExistingParent(candidatePath string) (string, error) {
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

func checkPathInsideBase(basePath string, candidatePath string, errorMessage string) error {
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
