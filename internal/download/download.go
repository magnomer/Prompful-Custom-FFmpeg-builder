package download

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"customffmpegbuilder/internal/consent"
	"customffmpegbuilder/internal/workspace"
)

type FileConflictPolicy string

const (
	FailIfExists       FileConflictPolicy = "fail-if-exists"
	ReuseIfHashMatches FileConflictPolicy = "reuse-if-hash-matches"
	OverwriteFile      FileConflictPolicy = "overwrite-approved"
)

type DownloadPlan struct {
	ActionName              string             `json:"actionName"`
	PlanHash                string             `json:"planHash"`
	WorkspaceDirectory      string             `json:"workspaceDirectory"`
	DownloadSourceName      string             `json:"downloadSourceName"`
	DownloadUrl             string             `json:"downloadUrl"`
	ExpectedSha256Hash      string             `json:"expectedSha256Hash"`
	DestinationFilePath     string             `json:"destinationFilePath"`
	AllowedHosts            []string           `json:"allowedHostNames"`
	ExpectedFileSizeMinimum int64              `json:"expectedFileSizeMinimum"`
	ExpectedFileSizeMaximum int64              `json:"expectedFileSizeMaximum"`
	FileConflictPolicy      FileConflictPolicy `json:"destinationConflictPolicyName"`
}

type ProgressFunc func(level string, message string)

func DownloadMsys2WithConsent(ctx context.Context, userDownloadConsent consent.Msys2DownloadConsent, downloadPlan DownloadPlan, emitProgress ProgressFunc) error {
	if err := consent.CheckConsent(userDownloadConsent.Consent, consent.ConsentKindMsys2Download, downloadPlan.ActionName, downloadPlan.PlanHash); err != nil {
		return err
	}
	return downloadFile(ctx, downloadPlan, emitProgress)
}

func DownloadFfmpegSourceWithConsent(ctx context.Context, userDownloadConsent consent.FfmpegSourceDownloadConsent, downloadPlan DownloadPlan, emitProgress ProgressFunc) error {
	if err := consent.CheckConsent(userDownloadConsent.Consent, consent.ConsentKindFfmpegSourceDownload, downloadPlan.ActionName, downloadPlan.PlanHash); err != nil {
		return err
	}
	return downloadFile(ctx, downloadPlan, emitProgress)
}

func FileSha256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func CheckFileHash(filePath string, expectedSha256Hash string) error {
	actualSha256Hash, err := FileSha256(filePath)
	if err != nil {
		return err
	}
	if !strings.EqualFold(actualSha256Hash, expectedSha256Hash) {
		return fmt.Errorf("sha256 mismatch: expected %s but got %s", expectedSha256Hash, actualSha256Hash)
	}
	return nil
}

func downloadFile(ctx context.Context, downloadPlan DownloadPlan, emitProgress ProgressFunc) error {
	if err := checkDownloadPlan(downloadPlan); err != nil {
		return err
	}
	if reuseMatchingFile(downloadPlan, emitProgress) {
		return nil
	}
	if emitProgress != nil {
		emitProgress("info", "Downloading approved file from "+downloadPlan.DownloadSourceName)
	}
	downloadDirectory := filepath.Dir(downloadPlan.DestinationFilePath)
	if err := workspace.CheckRealPathInsideWorkspace(downloadPlan.WorkspaceDirectory, downloadDirectory); err != nil {
		return err
	}
	if err := os.MkdirAll(downloadDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(downloadPlan.WorkspaceDirectory, downloadPlan.DestinationFilePath); err != nil {
		return err
	}
	temporaryPath := downloadPlan.DestinationFilePath + ".part"
	if err := workspace.CheckRealPathInsideWorkspace(downloadPlan.WorkspaceDirectory, temporaryPath); err != nil {
		return err
	}
	downloadSucceeded := false
	defer func() {
		if !downloadSucceeded {
			_ = os.Remove(temporaryPath)
		}
	}()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, downloadPlan.DownloadUrl, nil)
	if err != nil {
		return err
	}
	client := &http.Client{
		Timeout: 60 * time.Minute,
		CheckRedirect: func(redirectRequest *http.Request, previousRequests []*http.Request) error {
			if len(previousRequests) >= 5 {
				return errors.New("too many download redirects")
			}
			if redirectRequest.URL.Scheme != "https" {
				return errors.New("download redirect target must use https")
			}
			if !isAllowedHost(redirectRequest.URL.Hostname(), downloadPlan.AllowedHosts) {
				return fmt.Errorf("download redirect target host is not allowlisted: %s", redirectRequest.URL.Hostname())
			}
			return nil
		},
	}
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return fmt.Errorf("unexpected download status: %s", response.Status)
	}
	if response.ContentLength > downloadPlan.ExpectedFileSizeMaximum && downloadPlan.ExpectedFileSizeMaximum > 0 {
		return errors.New("download is larger than allowed maximum")
	}
	outputFile, err := os.OpenFile(temporaryPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	hash := sha256.New()
	writtenByteCount, copyErr := io.Copy(io.MultiWriter(outputFile, hash), response.Body)
	closeErr := outputFile.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if downloadPlan.ExpectedFileSizeMinimum > 0 && writtenByteCount < downloadPlan.ExpectedFileSizeMinimum {
		return errors.New("download is smaller than allowed minimum")
	}
	if downloadPlan.ExpectedFileSizeMaximum > 0 && writtenByteCount > downloadPlan.ExpectedFileSizeMaximum {
		return errors.New("download is larger than allowed maximum")
	}
	actualSha256Hash := hex.EncodeToString(hash.Sum(nil))
	if downloadPlan.ExpectedSha256Hash != "" {
		if !strings.EqualFold(actualSha256Hash, downloadPlan.ExpectedSha256Hash) {
			return fmt.Errorf("sha256 mismatch: expected %s but got %s", downloadPlan.ExpectedSha256Hash, actualSha256Hash)
		}
	} else if emitProgress != nil {
		emitProgress("info", "Calculated SHA-256 for "+downloadPlan.DownloadSourceName+": "+actualSha256Hash)
	}
	if downloadPlan.FileConflictPolicy == OverwriteFile {
		_ = os.Remove(downloadPlan.DestinationFilePath)
	}
	if err := os.Rename(temporaryPath, downloadPlan.DestinationFilePath); err != nil {
		return err
	}
	downloadSucceeded = true
	return nil
}

func reuseMatchingFile(downloadPlan DownloadPlan, emitProgress ProgressFunc) bool {
	if downloadPlan.FileConflictPolicy != ReuseIfHashMatches {
		return false
	}
	if _, err := os.Stat(downloadPlan.DestinationFilePath); err != nil {
		return false
	}
	if downloadPlan.ExpectedSha256Hash == "" {
		return false
	}
	if err := CheckFileHash(downloadPlan.DestinationFilePath, downloadPlan.ExpectedSha256Hash); err != nil {
		return false
	}
	if emitProgress != nil {
		emitProgress("info", "Reusing existing verified download: "+filepath.Base(downloadPlan.DestinationFilePath))
	}
	return true
}

func checkDownloadPlan(downloadPlan DownloadPlan) error {
	if downloadPlan.DownloadUrl == "" {
		return errors.New("download url is empty")
	}
	if downloadPlan.ExpectedSha256Hash != "" && !isSha256Hex(downloadPlan.ExpectedSha256Hash) {
		return errors.New("expected sha256 hash must be exactly 64 hexadecimal characters")
	}
	if downloadPlan.WorkspaceDirectory == "" {
		return errors.New("download workspace directory is empty")
	}
	parsedUrl, err := url.Parse(downloadPlan.DownloadUrl)
	if err != nil {
		return err
	}
	if parsedUrl.Scheme != "https" {
		return errors.New("download url must use https")
	}
	if !isAllowedHost(parsedUrl.Hostname(), downloadPlan.AllowedHosts) {
		return fmt.Errorf("download host is not allowlisted: %s", parsedUrl.Hostname())
	}
	if downloadPlan.DestinationFilePath == "" {
		return errors.New("destination file path is empty")
	}
	if err := workspace.CheckPathInsideWorkspace(downloadPlan.WorkspaceDirectory, downloadPlan.DestinationFilePath); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(downloadPlan.WorkspaceDirectory, downloadPlan.DestinationFilePath); err != nil {
		return err
	}
	if err := checkFileConflictPolicy(downloadPlan); err != nil {
		return err
	}
	return nil
}

func checkFileConflictPolicy(downloadPlan DownloadPlan) error {
	policyName := downloadPlan.FileConflictPolicy
	if policyName == "" {
		policyName = ReuseIfHashMatches
	}
	_, statError := os.Stat(downloadPlan.DestinationFilePath)
	if errors.Is(statError, os.ErrNotExist) {
		return nil
	}
	if statError != nil {
		return statError
	}
	switch policyName {
	case FailIfExists:
		return errors.New("download destination already exists")
	case ReuseIfHashMatches:
		if downloadPlan.ExpectedSha256Hash == "" {
			return errors.New("download destination already exists and cannot be reused without an expected SHA-256")
		}
		return CheckFileHash(downloadPlan.DestinationFilePath, downloadPlan.ExpectedSha256Hash)
	case OverwriteFile:
		return nil
	default:
		return fmt.Errorf("unknown download destination conflict policy: %s", policyName)
	}
}

func isAllowedHost(hostName string, allowedHostNames []string) bool {
	for _, allowedHostName := range allowedHostNames {
		if strings.EqualFold(hostName, allowedHostName) {
			return true
		}
	}
	return false
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
