package program

import (
	"errors"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
)

// LToolchainReviewValidate checks the session without consuming it, so a
// later step (the native confirmation dialog) can still be cancelled or retried.
// The session is only removed by LToolchainReviewConsume once the user has
// confirmed, which keeps it single-use without losing it on a rejected dialog.
func (program *LProgram) LToolchainReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewToolchainStored, error) {
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

func (program *LProgram) LToolchainReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LToolchainReviewStorage, reviewSessionId)
}

func (program *LProgram) LFFmpegReviewValidate(reviewSessionId string, approval consent.LRequestApproval) (LReviewFFmpegStored, error) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LFFmpegReviewStorage[reviewSessionId]
	if !exists {
		return LReviewFFmpegStored{}, errors.New("FFmpeg review session was not found")
	}
	if err := reviewsession.LReviewApprovalCheck(storedReviewSession.ReviewSession, approval.ApprovedActionName, approval.ApprovedPlanHash, approval.LConsentText); err != nil {
		return LReviewFFmpegStored{}, err
	}
	return storedReviewSession, nil
}

func (program *LProgram) LFFmpegReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LFFmpegReviewStorage, reviewSessionId)
}

func (program *LProgram) LNativeConsentAsk(actionName string, planHash string) (bool, error) {
	if program.LConfirmer == nil {
		return false, errors.New("no approval confirmer is configured")
	}
	return program.LConfirmer.LConfirmerApprovalGet(actionName, planHash)
}
