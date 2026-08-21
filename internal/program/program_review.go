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

// lToolchainReviewClaim atomically reserves the single-use session: under one
// lock it checks the approval against the stored review and, on success, marks
// the review used so a concurrent approval of the same session fails the used
// check. The reservation is a claim, not a consumption: it is committed by
// lToolchainReviewConsume only after the launch actually starts, and rolled
// back by lToolchainReviewRestore if any step before launch (native dialog,
// consent creation, action-slot acquisition) fails, which keeps the review
// single-use without losing it on a recoverable pre-launch failure. It is
// unexported so Wails does not bind it: only the outer approve entrypoints may
// drive review validation.
func (program *LProgram) lToolchainReviewClaim(reviewSessionId string, approval consent.LRequestApproval) (LReviewToolchainStored, error) {
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
	storedReviewSession.ReviewSession.WasUsed = true
	program.LToolchainReviewStorage[reviewSessionId] = storedReviewSession
	return storedReviewSession, nil
}

// lToolchainReviewRestore rolls back a claim made by lToolchainReviewClaim when
// a pre-launch step fails, so the review returns to its unused state and the
// user can retry from the same session instead of being stranded.
func (program *LProgram) lToolchainReviewRestore(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LToolchainReviewStorage[reviewSessionId]
	if !exists {
		return
	}
	storedReviewSession.ReviewSession.WasUsed = false
	program.LToolchainReviewStorage[reviewSessionId] = storedReviewSession
}

func (program *LProgram) lToolchainReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LToolchainReviewStorage, reviewSessionId)
}

// LToolchainReviewCancel revokes a stored toolchain review the frontend has
// abandoned (plan cancelled or its settings changed) so the backend session
// cannot still validate and launch its snapshot until the 30-minute expiry. It
// only removes authority and is idempotent: cancelling an already-consumed or
// unknown session is a no-op.
func (program *LProgram) LToolchainReviewCancel(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LToolchainReviewStorage, reviewSessionId)
}

func (program *LProgram) lFfmpegReviewClaim(reviewSessionId string, approval consent.LRequestApproval) (LReviewFfmpegStored, error) {
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
	storedReviewSession.ReviewSession.WasUsed = true
	program.LFfmpegReviewStorage[reviewSessionId] = storedReviewSession
	return storedReviewSession, nil
}

func (program *LProgram) lFfmpegReviewRestore(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	storedReviewSession, exists := program.LFfmpegReviewStorage[reviewSessionId]
	if !exists {
		return
	}
	storedReviewSession.ReviewSession.WasUsed = false
	program.LFfmpegReviewStorage[reviewSessionId] = storedReviewSession
}

func (program *LProgram) lFfmpegReviewConsume(reviewSessionId string) {
	program.LMutexReviewSession.Lock()
	defer program.LMutexReviewSession.Unlock()
	delete(program.LFfmpegReviewStorage, reviewSessionId)
}

// LFfmpegReviewCancel revokes a stored FFmpeg review the frontend has abandoned
// (plan cancelled or its settings changed) so the backend session cannot still
// validate and launch its snapshot until the 30-minute expiry. It only removes
// authority and is idempotent: cancelling an already-consumed or unknown session
// is a no-op.
func (program *LProgram) LFfmpegReviewCancel(reviewSessionId string) {
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
