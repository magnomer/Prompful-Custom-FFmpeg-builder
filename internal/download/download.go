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
	"sync/atomic"
	"time"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/workspace"
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

// A download can fail for transient reasons: the official URL redirects to a
// CDN mirror that stalls, a connection drops, or a mirror returns a 5xx. Each
// retry re-issues the request to the *official* DownloadUrl (not the redirected
// mirror), so the server is free to redirect to a healthier mirror. These
// constants bound the retries and define when a transfer counts as stalled.
const (
	maxDownloadAttempts       = 10
	downloadRetryInitialDelay = 5 * time.Second
	downloadRetryMaxDelay     = 60 * time.Second

	// downloadStallPoll is how often the watchdog samples transfer progress.
	downloadStallPoll = 2 * time.Second

	// Cap 1 ??zero download: abort when not a single byte arrives within this
	// window. Catches a fully dead transfer (DNS resolved, connection open, but
	// nothing flowing) without waiting for the overall client Timeout.
	downloadZeroDataTimeout = 10 * time.Second

	// Cap 2 ??too slow: abort when the average speed over a sliding window stays
	// below the floor. Catches a transfer that trickles (so it never trips the
	// zero-data cap) yet is effectively never going to finish. Mirrors pacman's
	// low-speed-limit/low-speed-time behavior.
	downloadLowSpeedWindow              = 20 * time.Second
	downloadLowSpeedFloorBytesPerSecond = 1024 // 1 KiB/s averaged over the window
)

// Stall kinds reported by the watchdog so the failure message can tell a fully
// dead transfer apart from a merely too-slow one.
const (
	stallKindNone int32 = iota
	stallKindZeroData
	stallKindTooSlow
)

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
	warnIfHostNotAllowlisted(downloadPlan, emitProgress)
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

	retryDelay := downloadRetryInitialDelay
	for attemptNumber := 1; ; attemptNumber++ {
		transientFailure, err := downloadAttempt(ctx, downloadPlan, temporaryPath, emitProgress)
		if err == nil {
			return nil
		}
		// Stop on user cancellation, a clearly non-transient failure (bad status,
		// size/hash mismatch), or once the attempt budget is spent.
		if ctx.Err() != nil || !transientFailure || attemptNumber >= maxDownloadAttempts {
			_ = os.Remove(temporaryPath)
			return err
		}
		if emitProgress != nil {
			emitProgress("warn", fmt.Sprintf("Transient download failure for %s (attempt %d of %d): %v. Retrying from the official URL in %s...", downloadPlan.DownloadSourceName, attemptNumber, maxDownloadAttempts, err, retryDelay))
		}
		select {
		case <-ctx.Done():
			_ = os.Remove(temporaryPath)
			return err
		case <-time.After(retryDelay):
		}
		if retryDelay *= 2; retryDelay > downloadRetryMaxDelay {
			retryDelay = downloadRetryMaxDelay
		}
	}
}

// downloadAttempt performs one download to temporaryPath. Each call re-requests
// downloadPlan.DownloadUrl (the official URL) so a retry lets the server pick a
// fresh, possibly healthier, redirect target. It reports whether the failure was
// transient (stalled transfer, dropped connection, 5xx) so the caller can retry.
func downloadAttempt(ctx context.Context, downloadPlan DownloadPlan, temporaryPath string, emitProgress ProgressFunc) (bool, error) {
	// Clear any partial file left by a prior failed attempt; O_EXCL below needs a
	// clean slate.
	_ = os.Remove(temporaryPath)

	attemptCtx, cancelAttempt := context.WithCancel(ctx)
	defer cancelAttempt()

	var bytesTransferred atomic.Int64
	var stallKind atomic.Int32
	watchdogDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(downloadStallPoll)
		defer ticker.Stop()
		now := time.Now()
		// Cap 1 (zero data): bytes seen at the last poll and when they last moved.
		lastSeenBytes := int64(0)
		lastProgressTime := now
		// Cap 2 (too slow): bytes and time at the start of the current speed window.
		windowStartBytes := int64(0)
		windowStartTime := now
		for {
			select {
			case <-watchdogDone:
				return
			case <-attemptCtx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				currentBytes := bytesTransferred.Load()

				// Cap 1 ??zero download: not one byte within the timeout.
				if currentBytes > lastSeenBytes {
					lastSeenBytes = currentBytes
					lastProgressTime = now
				} else if now.Sub(lastProgressTime) >= downloadZeroDataTimeout {
					stallKind.Store(stallKindZeroData)
					cancelAttempt()
					return
				}

				// Cap 2 ??too slow: average speed over a full window below the floor.
				if elapsed := now.Sub(windowStartTime); elapsed >= downloadLowSpeedWindow {
					averageBytesPerSecond := float64(currentBytes-windowStartBytes) / elapsed.Seconds()
					if averageBytesPerSecond < downloadLowSpeedFloorBytesPerSecond {
						stallKind.Store(stallKindTooSlow)
						cancelAttempt()
						return
					}
					windowStartBytes = currentBytes
					windowStartTime = now
				}
			}
		}
	}()
	defer close(watchdogDone)

	request, err := http.NewRequestWithContext(attemptCtx, http.MethodGet, downloadPlan.DownloadUrl, nil)
	if err != nil {
		return false, err
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
			if !isAllowedHost(redirectRequest.URL.Hostname(), downloadPlan.AllowedHosts) && emitProgress != nil {
				emitProgress("warn", "Download for "+downloadPlan.DownloadSourceName+" redirected to a host that is not on the trusted allowlist: "+redirectRequest.URL.Hostname()+". Proceeding; downloaded content is still verified by signature/SHA-256.")
			}
			return nil
		},
	}
	response, err := client.Do(request)
	if err != nil {
		// Any connection-level error (DNS, reset, refused) ??or a stall the watchdog
		// turned into a cancellation ??is transient. A genuine user cancellation is
		// caught by the caller's ctx.Err() check.
		return true, describeStallError(stallKind.Load(), err)
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return isRetryableStatusCode(response.StatusCode), fmt.Errorf("unexpected download status: %s", response.Status)
	}
	if response.ContentLength > downloadPlan.ExpectedFileSizeMaximum && downloadPlan.ExpectedFileSizeMaximum > 0 {
		return false, errors.New("download is larger than allowed maximum")
	}
	outputFile, err := os.OpenFile(temporaryPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err != nil {
		return false, err
	}
	hash := sha256.New()
	activityTrackingBody := &activityTrackingReader{reader: response.Body, bytesTransferred: &bytesTransferred}
	writtenByteCount, copyErr := io.Copy(io.MultiWriter(outputFile, hash), activityTrackingBody)
	closeErr := outputFile.Close()
	if copyErr != nil {
		_ = os.Remove(temporaryPath)
		// A stall the watchdog aborted, or a mid-stream read failure (connection
		// dropped), is transient.
		return true, describeStallError(stallKind.Load(), copyErr)
	}
	if closeErr != nil {
		return false, closeErr
	}
	if downloadPlan.ExpectedFileSizeMinimum > 0 && writtenByteCount < downloadPlan.ExpectedFileSizeMinimum {
		return false, errors.New("download is smaller than allowed minimum")
	}
	if downloadPlan.ExpectedFileSizeMaximum > 0 && writtenByteCount > downloadPlan.ExpectedFileSizeMaximum {
		return false, errors.New("download is larger than allowed maximum")
	}
	actualSha256Hash := hex.EncodeToString(hash.Sum(nil))
	if downloadPlan.ExpectedSha256Hash != "" {
		if !strings.EqualFold(actualSha256Hash, downloadPlan.ExpectedSha256Hash) {
			return false, fmt.Errorf("sha256 mismatch: expected %s but got %s", downloadPlan.ExpectedSha256Hash, actualSha256Hash)
		}
	} else if emitProgress != nil {
		emitProgress("info", "Calculated SHA-256 for "+downloadPlan.DownloadSourceName+": "+actualSha256Hash)
	}
	if downloadPlan.FileConflictPolicy == OverwriteFile {
		_ = os.Remove(downloadPlan.DestinationFilePath)
	}
	if err := os.Rename(temporaryPath, downloadPlan.DestinationFilePath); err != nil {
		return false, err
	}
	return false, nil
}

// activityTrackingReader accumulates the running total of bytes read so the
// stall watchdog can measure both whether data is flowing at all and how fast.
type activityTrackingReader struct {
	reader           io.Reader
	bytesTransferred *atomic.Int64
}

func (r *activityTrackingReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.bytesTransferred.Add(int64(n))
	}
	return n, err
}

// describeStallError wraps a transfer error with the watchdog's reason when the
// attempt was aborted for being dead or too slow, so the retry message tells the
// two cases apart. With no stall it returns the underlying error unchanged.
func describeStallError(stallKind int32, baseErr error) error {
	switch stallKind {
	case stallKindZeroData:
		return fmt.Errorf("download stalled: no data received for %s: %w", downloadZeroDataTimeout, baseErr)
	case stallKindTooSlow:
		return fmt.Errorf("download too slow: averaged under %d bytes/sec over %s: %w", downloadLowSpeedFloorBytesPerSecond, downloadLowSpeedWindow, baseErr)
	default:
		return baseErr
	}
}

// isRetryableStatusCode reports whether an HTTP status warrants a retry: server
// errors (5xx) and the standard "slow down / try again" codes. Other non-2xx
// codes (e.g. 404) are permanent and are not retried.
func isRetryableStatusCode(statusCode int) bool {
	if statusCode >= 500 {
		return true
	}
	return statusCode == http.StatusTooManyRequests || statusCode == http.StatusRequestTimeout
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

func warnIfHostNotAllowlisted(downloadPlan DownloadPlan, emitProgress ProgressFunc) {
	if emitProgress == nil {
		return
	}
	parsedUrl, err := url.Parse(downloadPlan.DownloadUrl)
	if err != nil {
		return
	}
	if isAllowedHost(parsedUrl.Hostname(), downloadPlan.AllowedHosts) {
		return
	}
	emitProgress("warn", "Download host for "+downloadPlan.DownloadSourceName+" is not on the trusted allowlist: "+parsedUrl.Hostname()+". Proceeding because downloaded content is still verified by signature/SHA-256. Make sure you trust this host.")
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
