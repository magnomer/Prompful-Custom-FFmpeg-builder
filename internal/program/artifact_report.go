package program

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/workspace"
)

func LReportArtifactWrite(workspaceLayout workspace.LWorkspaceLayout, LRunId string, reviewSessionId string, plan planning.LPlanFfmpeg) error {
	reportPath := filepath.Join(workspaceLayout.ArtifactsDirectory, "build-report-"+LRunId+".json")
	ffmpegExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffmpeg.exe")
	ffprobeExecutablePath := filepath.Join(workspaceLayout.ArtifactsDirectory, "ffprobe.exe")
	report := map[string]interface{}{"runId": LRunId, "reviewSessionId": reviewSessionId, "createdAt": time.Now().UTC().Format(time.RFC3339), "approvedPlanHash": plan.PlanHash, "ffmpegVersion": planning.LVersionArchiveParse(plan.FfmpegSourceArchiveUrl), "ffmpegSourceArchiveUrl": plan.FfmpegSourceArchiveUrl, "ffmpegSourceSignatureUrl": plan.FfmpegSourceSignatureUrl, "ffmpegSourceSha256Hash": plan.FfmpegSourceSha256Hash, "selectedLibraries": plan.SelectedLibraries, "selectedConfigureOptions": plan.SelectedConfigureOptions, "requiredMsys2PackageNames": plan.RequiredMsys2PackageNames, "generatedConfigureFlags": plan.GeneratedConfigureFlags, "generatedOptionFlags": plan.GeneratedOptionFlags, "extraConfigureFlags": plan.ExtraConfigureFlags, "configureFlags": plan.ConfigureFlags, "licenseProfileName": plan.LicenseProfileName, "ffmpegExecutablePath": ffmpegExecutablePath, "ffmpegExecutableSha256Hash": LHashFileCreate(ffmpegExecutablePath), "ffmpegExecutableSizeBytes": LFileSizeRead(ffmpegExecutablePath), "ffprobeExecutablePath": ffprobeExecutablePath, "ffprobeExecutableSha256Hash": LHashFileCreate(ffprobeExecutablePath), "ffprobeExecutableSizeBytes": LFileSizeRead(ffprobeExecutablePath), "artifactFiles": LArtifactFileList(workspaceLayout)}
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
	versionOutput, err := LFfmpegVersionRun(ffmpegExecutablePath)
	if err != nil {
		return ""
	}
	return LVersionFfmpegParse(versionOutput)
}

// LArtifactLayoutFind returns the workspace layout whose ArtifactsDirectory
// holds the most recently written build report. Builds are stored per FFmpeg
// version under <workspace>/FFmpeg/<version>/, so this scans each version
// subdirectory (and, for backward compatibility with pre-versioning builds, the
// FFmpeg base directory itself) and selects the directory with the newest
// build-report-*.json. When no report exists anywhere it returns the base layout,
// so callers still get a valid (empty) artifacts directory to report on.
func LArtifactLayoutFind(workspaceDirectory string) workspace.LWorkspaceLayout {
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

func LArtifactFileList(workspaceLayout workspace.LWorkspaceLayout) []LFileResult {
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
