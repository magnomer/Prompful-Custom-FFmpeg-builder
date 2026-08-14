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
	LFFmpegReviewStorage    map[string]LReviewFFmpegStored
	LStateWindowStartup     LStateWindow
	LLocaleUi               string
	LMutexLocaleUi          sync.RWMutex
	LReporter               reporting.LReporter
	LConfirmer              reporting.LConfirmer
}

type LReviewToolchainStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanToolchain
}

type LReviewFFmpegStored struct {
	ReviewSession reviewsession.LSessionReview
	Plan          planning.LPlanFFmpeg
}

func LProgramCreate() *LProgram {
	return &LProgram{LStateWindowStartup: LStateWindowLoad(), LToolchainReviewStorage: map[string]LReviewToolchainStored{}, LFFmpegReviewStorage: map[string]LReviewFFmpegStored{}}
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
	program.LActionApprovedCancel()
}

// LLocaleSet records the UI language so the backend-rendered native confirmation
// dialog can follow it. Only the dialog is localized; log and error messages stay
// English. Unknown locales fall back to English.
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
