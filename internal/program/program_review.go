package program

import (
	"errors"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
)

// lConsentTextLimit bounds the caller-supplied approval consent text before
// it is hashed under the shared review-session mutex. The legitimate consent
// text is a short fixed sentence over a backend-generated action name and plan
// hash, so any submission beyond this is malformed. Rejecting it before locking
// keeps an oversized string from blocking unrelated review work while it hashes.
const lConsentTextLimit = 4096

// lToolchainReviewValidate checks the session without consuming it, so a
// later step (the native confirmation dialog) can still be cancelled or retried.
// The session is only removed by lToolchainReviewConsume once the user has
// confirmed, which keeps it single-use without losing it on a rejected dialog.
// It is unexported so Wails does not bind it: only the outer approve entrypoints
// may drive review validation.
func (program *LProgram) lToolchainReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewToolchainStored, error) {
	if len(approval.LConsentText) > lConsentTextLimit {
		return LReviewToolchainStored{}, errors.New("approval consent text exceeds maximum length")
	}
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LToolchainReviewStorage[reviewSessionId]
	if !exists {
		return LReviewToolchainStored{}, errors.New("toolchain review session was not found")
	}
	if err := reviewsession.LReviewApprovalCheck(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.LConsentText); err != nil {
		return LReviewToolchainStored{}, err
	}
	return storedReviewSession, nil
}

func (program *LProgram) lToolchainReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LToolchainReviewStorage, reviewSessionId)
}

func (program *LProgram) lFfmpegReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewFfmpegStored, error) {
	if len(approval.LConsentText) > lConsentTextLimit {
		return LReviewFfmpegStored{}, errors.New("approval consent text exceeds maximum length")
	}
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LFfmpegReviewStorage[reviewSessionId]
	if !exists {
		return LReviewFfmpegStored{}, errors.New("FFmpeg review session was not found")
	}
	if err := reviewsession.LReviewApprovalCheck(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.LConsentText); err != nil {
		return LReviewFfmpegStored{}, err
	}
	return storedReviewSession, nil
}

func (program *LProgram) lFfmpegReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LFfmpegReviewStorage, reviewSessionId)
}

// lNativeConsentAsk opens the backend-owned native confirmation dialog. It is
// unexported so Wails does not bind it: the trusted native dialog can only be
// raised from inside a validated approve path, never as a standalone frontend
// call.
func (program *LProgram) lNativeConsentAsk(actionName string, planHash string) (bool, error) {
	if program.LConfirmer == nil {
		return false, errors.New("no approval confirmer is configured")
	}
	return program.LConfirmer.LConfirmerApprovalGet(actionName, planHash)
}
