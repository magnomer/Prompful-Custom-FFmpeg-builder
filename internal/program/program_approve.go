package program

import (
	"errors"
	"time"

	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/internal/reviewsession"
)

func (program *LProgram) LPlanToolchainRequest(buildConfigSettings planning.LSettingsToolchain) (planning.LReviewToolchain, error) {
	plan, err := planning.LPlanToolchainCreate(buildConfigSettings)
	if err != nil {
		return planning.LReviewToolchain{}, err
	}
	reviewSession, err := reviewsession.LSessionReviewCreate(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.LReviewToolchain{}, err
	}
	program.LMutexReviewSession.Lock()
	program.LToolchainReviewStorage[reviewSession.ReviewSessionId] = LReviewToolchainStored{ReviewSession: reviewSession, Plan: plan}
	program.LMutexReviewSession.Unlock()
	return planning.LReviewToolchain{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedLConsentText: reviewSession.ExpectedLConsentText, ExpectedLConsentTextHash: reviewSession.ExpectedLConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

func (program *LProgram) LPlanFfmpegRequest(ffmpegBuildSettings planning.LSettingsFfmpeg) (planning.LReviewFfmpeg, error) {
	plan, err := planning.LPlanFfmpegCreate(ffmpegBuildSettings)
	if err != nil {
		return planning.LReviewFfmpeg{}, err
	}
	reviewSession, err := reviewsession.LSessionReviewCreate(plan.ActionName, plan.PlanHash, 30*time.Minute)
	if err != nil {
		return planning.LReviewFfmpeg{}, err
	}
	program.LMutexReviewSession.Lock()
	program.LFfmpegReviewStorage[reviewSession.ReviewSessionId] = LReviewFfmpegStored{ReviewSession: reviewSession, Plan: plan}
	program.LMutexReviewSession.Unlock()
	return planning.LReviewFfmpeg{ReviewSessionId: reviewSession.ReviewSessionId, ExpectedLConsentText: reviewSession.ExpectedLConsentText, ExpectedLConsentTextHash: reviewSession.ExpectedLConsentTextHash, ExpiresAtUnixTime: reviewSession.ExpiresAtUnixTime, Plan: plan}, nil
}

// LPlanToolchainApprove validates the review, confirms, and prepares the
// toolchain asynchronously (GUI behavior).
func (program *LProgram) LPlanToolchainApprove(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.lToolchainApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lToolchainPrepareLaunch(reviewSessionId, plan, approval, false)
}

// LToolchainApproveSync is the CLI counterpart: it prepares the toolchain
// inline and returns only after preparation finishes, so the caller can map the
// outcome to an exit code. The final status arrives through the reporter.
func (program *LProgram) LToolchainApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.lToolchainApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lToolchainPrepareLaunch(reviewSessionId, plan, approval, true)
}

// lToolchainApproveValidate performs the shared validation for both toolchain
// approve paths. It atomically claims the single-use session up front so a
// concurrent approval of the same review is rejected, then rolls the claim back
// if a pre-confirmation check or the native dialog fails, so the session stays
// retryable. Actual consumption is deferred to lToolchainPrepareLaunch, which
// consumes only once the run has started.
func (program *LProgram) lToolchainApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanToolchain, error) {
	storedReviewSession, err := program.lToolchainReviewClaim(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanToolchain{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return planning.LPlanToolchain{}, err
	}
	if err := LHashToolchainVerify(plan); err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return planning.LPlanToolchain{}, err
	}
	confirmed, err := program.lNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return planning.LPlanToolchain{}, err
	}
	if !confirmed {
		program.lToolchainReviewRestore(reviewSessionId)
		return planning.LPlanToolchain{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	// The confirmation dialog/prompt can block for an unbounded time, during
	// which the review lifetime may lapse. Re-enforce expiry against the
	// monotonic deadline before consuming the session, so approval is refused
	// when the user confirms after the review has expired.
	if err := reviewsession.LReviewExpiryCheck(storedReviewSession.ReviewSession); err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return planning.LPlanToolchain{}, err
	}
	return plan, nil
}

// lToolchainPrepareLaunch builds the per-action consents and starts the
// toolchain worker, inline when runInline is true (CLI) or on a goroutine
// otherwise (GUI). It commits the claimed review only after the launch actually
// starts (consents built and action slot acquired); if any of those steps
// fails, it restores the claim so the confirmed review is not lost.
func (program *LProgram) lToolchainPrepareLaunch(reviewSessionId string, plan planning.LPlanToolchain, approval consent.LRequestApproval, runInline bool) (LResultAction, error) {
	userLConsentMsys, err := consent.LConsentMsysCreate(approval)
	if err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		program.lToolchainReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	program.lToolchainReviewConsume(reviewSessionId)
	if runInline {
		program.LToolchainPrepare(LContextAction, LRunId, reviewSessionId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	} else {
		go program.LToolchainPrepare(LContextAction, LRunId, reviewSessionId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	}
	return LResultAction{RunId: LRunId, ReviewSessionId: reviewSessionId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// LPlanFfmpegApprove validates the review, confirms, and launches the build
// asynchronously (GUI behavior: returns a RunId immediately, progress arrives
// through the reporter).
func (program *LProgram) LPlanFfmpegApprove(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, expiresAtUnixTime, err := program.lFfmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lFfmpegCompilationLaunch(reviewSessionId, plan, approval, expiresAtUnixTime, false)
}

// LFfmpegApproveSync is the CLI counterpart of LPlanFfmpegApprove: it runs
// the build inline and returns only after the build finishes, so the caller can
// map the outcome to an exit code. The final status is delivered through the
// reporter, exactly as in the async path.
func (program *LProgram) LFfmpegApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, expiresAtUnixTime, err := program.lFfmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lFfmpegCompilationLaunch(reviewSessionId, plan, approval, expiresAtUnixTime, true)
}

// lFfmpegApproveValidate performs the shared, side-effect-ordered validation for
// both approve paths: session claim, executability, hash, toolchain readiness,
// and the backend-owned confirmation. It atomically claims the single-use
// session up front so a concurrent approval of the same review is rejected, and
// rolls the claim back if any pre-confirmation check or the native dialog fails,
// so the session stays retryable. It returns the reviewed plan and the original
// review expiry; consumption is deferred to lFfmpegCompilationLaunch, which
// consumes only after the run has started.
func (program *LProgram) lFfmpegApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanFfmpeg, int64, error) {
	storedReviewSession, err := program.lFfmpegReviewClaim(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanFfmpeg{}, 0, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	if err := LHashFfmpegVerify(plan); err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	if err := LToolchainPreparedCheck(plan.WorkspaceDirectory, plan.WindowsShellProfileName); err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	if err := LToolchainPackageCheck(plan.WorkspaceDirectory, plan.WindowsShellProfileName, plan.RequiredMsys2PackageNames); err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	confirmed, err := program.lNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	if !confirmed {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, errors.New("user rejected approval in backend-owned confirmation")
	}
	// The confirmation dialog/prompt can block for an unbounded time, during
	// which the review lifetime may lapse. Re-enforce expiry against the
	// monotonic deadline before consuming the session, so approval is refused
	// when the user confirms after the review has expired.
	if err := reviewsession.LReviewExpiryCheck(storedReviewSession.ReviewSession); err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return planning.LPlanFfmpeg{}, 0, err
	}
	return plan, storedReviewSession.ReviewSession.ExpiresAtUnixTime, nil
}

// LFfmpegRetryRun re-launches the last confirmed FFmpeg run after a stall. It
// is the only approved way for the frontend to resume, because the launch
// helpers are unexported and never bound. It re-enforces the full approval
// boundary against server-owned state: the run must still be within its
// original review lifetime, its retained plan must still match its hash, and
// the user must confirm again in the backend-owned native dialog.
func (program *LProgram) LFfmpegRetryRun() (LResultAction, error) {
	program.LMutexReviewSession.Lock()
	retained := program.lApprovalFfmpegRetained
	program.LMutexReviewSession.Unlock()
	if retained == nil {
		return LResultAction{}, errors.New("no approved FFmpeg run is available to retry")
	}
	if time.Now().UTC().Unix() >= retained.ExpiresAtUnixTime {
		return LResultAction{}, errors.New("the approval for this FFmpeg run has expired; review and approve the build again")
	}
	if err := LHashFfmpegVerify(retained.Plan); err != nil {
		return LResultAction{}, err
	}
	confirmed, err := program.lNativeConsentAsk(retained.Plan.ActionName, retained.Plan.PlanHash)
	if err != nil {
		return LResultAction{}, err
	}
	if !confirmed {
		return LResultAction{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	// The confirmation can block past the retained approval window; re-enforce
	// expiry before re-launching so a stale run cannot start after it lapsed.
	if time.Now().UTC().Unix() >= retained.ExpiresAtUnixTime {
		return LResultAction{}, errors.New("the approval for this FFmpeg run has expired; review and approve the build again")
	}
	return program.lFfmpegCompilationLaunch(retained.ReviewSessionId, retained.Plan, retained.Approval, retained.ExpiresAtUnixTime, false)
}

// lFfmpegCompilationLaunch builds the per-action consents and starts the build
// worker, inline when runInline is true (CLI) or on a goroutine otherwise (GUI).
// It commits the claimed review only after the launch actually starts (consents
// built and action slot acquired); if any of those steps fails, it restores the
// claim so the confirmed review is not lost. On success it retains the
// backend-verified plan, its approval, and the original review expiry so a
// post-stall Retry re-launches from this server-owned state instead of a
// frontend-supplied plan, and cannot outlive the review window.
func (program *LProgram) lFfmpegCompilationLaunch(reviewSessionId string, plan planning.LPlanFfmpeg, approval consent.LRequestApproval, expiresAtUnixTime int64, runInline bool) (LResultAction, error) {
	userLConsentFfmpeg, err := consent.LConsentFfmpegCreate(approval)
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	userExternalLConsentCommand, err := consent.LConsentCommandCreate(approval)
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		program.lFfmpegReviewRestore(reviewSessionId)
		return LResultAction{}, err
	}
	program.lFfmpegReviewConsume(reviewSessionId)
	program.LMutexReviewSession.Lock()
	program.lApprovalFfmpegRetained = &lApprovalFfmpegStored{ReviewSessionId: reviewSessionId, Plan: plan, Approval: approval, ExpiresAtUnixTime: expiresAtUnixTime}
	program.LMutexReviewSession.Unlock()
	if runInline {
		program.LFfmpegCompile(LContextAction, LRunId, reviewSessionId, plan, userLConsentFfmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	} else {
		go program.LFfmpegCompile(LContextAction, LRunId, reviewSessionId, plan, userLConsentFfmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	}
	return LResultAction{RunId: LRunId, ReviewSessionId: reviewSessionId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}
