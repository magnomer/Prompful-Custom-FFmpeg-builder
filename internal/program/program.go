package program

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/reporting"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
	"promptfulcustomffmpegbuilder/internal/workspace"

	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

type LProgram struct {
	LContext                context.Context
	LContextAction          context.Context
	LActionCancelFunction   context.CancelFunc
	lActionDone             chan struct{}
	lProgramStopping        bool
	lWorkspaceOwner         *workspace.LWorkspaceOwner
	LMutexAction            sync.Mutex
	LMutexReviewSession     sync.Mutex
	LToolchainReviewStorage map[string]LReviewToolchainStored
	LFfmpegReviewStorage    map[string]LReviewFfmpegStored
	lApprovalFfmpegRetained *lApprovalFfmpegStored
	LStateWindowStartup     LStateWindow
	LLocaleUi               string
	LMutexLocaleUi          sync.RWMutex
	LMutexConfirmation      sync.Mutex
	LConfirmationRequestId  string
	LConfirmationResponse   chan bool
	LReporter               reporting.LReporter
	LConfirmer              reporting.LConfirmer
}

type LReviewToolchainStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanToolchain
}

type LReviewFfmpegStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanFfmpeg
}

// lApprovalFfmpegStored retains the backend-verified plan and approval of the
// last confirmed FFmpeg run so a post-stall Retry re-launches from server-owned
// state, bounded by the original review's expiry, rather than from a
// frontend-held plan. ExpiresAtUnixTime carries the original review lifetime so
// Retry cannot renew approval past it.
type lApprovalFfmpegStored struct {
	ReviewSessionId   string
	Plan              planning.LPlanFfmpeg
	Approval          consent.LRequestApproval
	ExpiresAtUnixTime int64
}

func LProgramCreate() *LProgram {
	return &LProgram{LStateWindowStartup: LStateWindowLoad(), LToolchainReviewStorage: map[string]LReviewToolchainStored{}, LFfmpegReviewStorage: map[string]LReviewFfmpegStored{}}
}

func LWindowInitialRead(program *LProgram) (int, int) {
	return program.LStateWindowStartup.Width, program.LStateWindowStartup.Height
}

func LLocaleTextGet(key string, values map[string]string) string {
	return LLocaleTextGetInternal(key, values)
}

func (program *LProgram) LProgramStart(LContext context.Context) {
	program.LContext = LContext
	program.LReporter = LReporterWails{program: program}
	program.LConfirmer = LConfirmerWails{program: program}
}

// LProgramReady runs after the native window and WebView runtime are ready.
// Wails does not guarantee that window runtime calls work during OnStartup.
func (program *LProgram) LProgramReady(LContext context.Context) {
	program.LContext = LContext
	program.LWindowGeometryRestore()
}

// LWindowCloseCheck runs while the window still exists, so it is the safe place to
// read window geometry. Returning false allows the close to proceed.
func (program *LProgram) LWindowCloseCheck(LContext context.Context) bool {
	if err := program.LWindowGeometrySave(LContext); err != nil {
		_, _ = wailsRuntime.MessageDialog(LContext, wailsRuntime.MessageDialogOptions{
			Type:    wailsRuntime.ErrorDialog,
			Title:   LLocaleTextForGet(program.lLocaleCurrentGet(), "persistence.windowSaveFailed.title", nil),
			Message: LLocaleTextForGet(program.lLocaleCurrentGet(), "persistence.windowSaveFailed.message", map[string]string{"message": err.Error()}),
		})
	}
	return false
}

func (program *LProgram) LProgramStop(LContext context.Context) {
	program.lApprovalConfirmationCancel()
	program.lActionApprovedStop()
}

// LLocaleSet records the UI language for backend-owned native surfaces such as
// the workspace picker. Logs and technical error details stay English. Unknown
// locales fall back to English.
func LLocaleNormalize(locale string) string {
	normalizedLocale := "en"
	if strings.TrimSpace(locale) == "ko" {
		normalizedLocale = "ko"
	}
	return normalizedLocale
}

func LPathLocaleResolve() (string, error) {
	configDirectory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDirectory, "PromptfulCustomFfmpegBuilder", "locale-state.txt"), nil
}

// LLocaleLoad restores the UI locale from backend-owned storage before React
// mounts, so startup does not depend on WebView localStorage retention.
func (program *LProgram) LLocaleLoad() (string, error) {
	filePath, err := LPathLocaleResolve()
	if err != nil {
		return "", err
	}
	fileData, err := os.ReadFile(filePath)
	if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	normalizedLocale := LLocaleNormalize(string(fileData))
	program.lLocaleCommit(normalizedLocale)
	return normalizedLocale, nil
}

func (program *LProgram) LLocaleSet(locale string) error {
	normalizedLocale := LLocaleNormalize(locale)
	filePath, err := LPathLocaleResolve()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(filePath), 0o755); err != nil {
		return err
	}
	if err := LStateFileAtomicWrite(filePath, []byte(normalizedLocale+"\n"), 0o644); err != nil {
		return err
	}
	program.lLocaleCommit(normalizedLocale)
	return nil
}

func (program *LProgram) lLocaleCommit(normalizedLocale string) {
	program.LMutexLocaleUi.Lock()
	program.LLocaleUi = normalizedLocale
	program.LMutexLocaleUi.Unlock()
}

func (program *LProgram) lLocaleCurrentGet() string {
	program.LMutexLocaleUi.RLock()
	defer program.LMutexLocaleUi.RUnlock()
	if program.LLocaleUi == "" {
		return "en"
	}
	return program.LLocaleUi
}
