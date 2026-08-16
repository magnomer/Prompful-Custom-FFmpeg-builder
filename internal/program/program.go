package program

import (
	"context"
	"sync"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/reporting"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
)

type LProgram struct {
	LContext                context.Context
	LContextAction          context.Context
	LActionCancelFunction   context.CancelFunc
	LMutexAction            sync.Mutex
	LMutexReviewSession     sync.Mutex
	LToolchainReviewStorage map[string]LReviewToolchainStored
	LFfmpegReviewStorage    map[string]LReviewFfmpegStored
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
	program.LWindowGeometryRestore()
}

// LWindowCloseCheck runs while the window still exists, so it is the safe place to
// read window geometry. Returning false allows the close to proceed.
func (program *LProgram) LWindowCloseCheck(LContext context.Context) bool {
	program.LWindowGeometrySave(LContext)
	return false
}

func (program *LProgram) LProgramStop(LContext context.Context) {
	program.lApprovalConfirmationCancel()
	program.LActionApprovedCancel()
}

// LLocaleSet records the UI language for backend-owned native surfaces such as
// the workspace picker. Logs and technical error details stay English. Unknown
// locales fall back to English.
func (program *LProgram) LLocaleSet(locale string) {
	normalizedLocale := "en"
	if locale == "ko" {
		normalizedLocale = "ko"
	}
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
