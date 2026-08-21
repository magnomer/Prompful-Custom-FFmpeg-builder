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
	return program.lToolchainPrepareLaunch(plan, approval, false)
}

// LToolchainApproveSync is the CLI counterpart: it prepares the toolchain
// inline and returns only after preparation finishes, so the caller can map the
// outcome to an exit code. The final status arrives through the reporter.
func (program *LProgram) LToolchainApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.lToolchainApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lToolchainPrepareLaunch(plan, approval, true)
}

// lToolchainApproveValidate performs the shared validation for both toolchain
// approve paths and consumes the single-use session only after confirmation.
func (program *LProgram) lToolchainApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanToolchain, error) {
	storedReviewSession, err := program.lToolchainReviewValidate(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanToolchain{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		return planning.LPlanToolchain{}, err
	}
	if err := LHashToolchainVerify(plan); err != nil {
		return planning.LPlanToolchain{}, err
	}
	confirmed, err := program.lNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		return planning.LPlanToolchain{}, err
	}
	if !confirmed {
		return planning.LPlanToolchain{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	program.lToolchainReviewConsume(reviewSessionId)
	return plan, nil
}

// lToolchainPrepareLaunch builds the per-action consents and starts the
// toolchain worker, inline when runInline is true (CLI) or on a goroutine
// otherwise (GUI).
func (program *LProgram) lToolchainPrepareLaunch(plan planning.LPlanToolchain, approval consent.LRequestApproval, runInline bool) (LResultAction, error) {
	userLConsentMsys, err := consent.LConsentMsysCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		return LResultAction{}, err
	}
	if runInline {
		program.LToolchainPrepare(LContextAction, LRunId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	} else {
		go program.LToolchainPrepare(LContextAction, LRunId, plan, userLConsentMsys, userLConsentArchive, userPacmanPackageInstallLConsent)
	}
	return LResultAction{RunId: LRunId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}

// LPlanFfmpegApprove validates the review, confirms, and launches the build
// asynchronously (GUI behavior: returns a RunId immediately, progress arrives
// through the reporter).
func (program *LProgram) LPlanFfmpegApprove(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.lFfmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lFfmpegCompilationLaunch(plan, approval, false)
}

// LFfmpegApproveSync is the CLI counterpart of LPlanFfmpegApprove: it runs
// the build inline and returns only after the build finishes, so the caller can
// map the outcome to an exit code. The final status is delivered through the
// reporter, exactly as in the async path.
func (program *LProgram) LFfmpegApproveSync(reviewSessionId string, approval consent.LRequestApproval) (LResultAction, error) {
	plan, err := program.lFfmpegApproveValidate(reviewSessionId, approval)
	if err != nil {
		return LResultAction{}, err
	}
	return program.lFfmpegCompilationLaunch(plan, approval, true)
}

// lFfmpegApproveValidate performs the shared, side-effect-ordered validation for
// both approve paths: session check, executability, hash, toolchain readiness,
// and the backend-owned confirmation. It consumes the single-use session only
// after confirmation succeeds, so a rejected confirmation leaves it retryable.
func (program *LProgram) lFfmpegApproveValidate(reviewSessionId string, approval consent.LRequestApproval) (planning.LPlanFfmpeg, error) {
	storedReviewSession, err := program.lFfmpegReviewValidate(reviewSessionId, approval)
	if err != nil {
		return planning.LPlanFfmpeg{}, err
	}
	plan := storedReviewSession.Plan
	if err := planning.LPlanRunCheck(plan.IsExecutable); err != nil {
		return planning.LPlanFfmpeg{}, err
	}
	if err := LHashFfmpegVerify(plan); err != nil {
		return planning.LPlanFfmpeg{}, err
	}
	if err := LToolchainPreparedCheck(plan.WorkspaceDirectory, plan.WindowsShellProfileName); err != nil {
		return planning.LPlanFfmpeg{}, err
	}
	confirmed, err := program.lNativeConsentAsk(plan.ActionName, plan.PlanHash)
	if err != nil {
		return planning.LPlanFfmpeg{}, err
	}
	if !confirmed {
		return planning.LPlanFfmpeg{}, errors.New("user rejected approval in backend-owned confirmation")
	}
	program.lFfmpegReviewConsume(reviewSessionId)
	// Retain the backend-verified plan, its approval, and the original review
	// expiry so a post-stall Retry re-launches from this server-owned state
	// instead of a frontend-supplied plan, and cannot outlive the review window.
	program.LMutexReviewSession.Lock()
	program.lApprovalFfmpegRetained = &lApprovalFfmpegStored{Plan: plan, Approval: approval, ExpiresAtUnixTime: storedReviewSession.ReviewSession.ExpiresAtUnixTime}
	program.LMutexReviewSession.Unlock()
	return plan, nil
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
	if time.Now().UTC().Unix() > retained.ExpiresAtUnixTime {
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
	return program.lFfmpegCompilationLaunch(retained.Plan, retained.Approval, false)
}

// lFfmpegCompilationLaunch builds the per-action consents and starts the build worker,
// inline when runInline is true (CLI) or on a goroutine otherwise (GUI).
func (program *LProgram) lFfmpegCompilationLaunch(plan planning.LPlanFfmpeg, approval consent.LRequestApproval, runInline bool) (LResultAction, error) {
	userLConsentFfmpeg, err := consent.LConsentFfmpegCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userLConsentArchive, err := consent.LConsentArchiveCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userExternalLConsentCommand, err := consent.LConsentCommandCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	userPacmanPackageInstallLConsent, err := consent.LConsentPacmanCreate(approval)
	if err != nil {
		return LResultAction{}, err
	}
	LRunId, LContextAction, err := program.LActionApprovedStart()
	if err != nil {
		return LResultAction{}, err
	}
	if runInline {
		program.LFfmpegCompile(LContextAction, LRunId, plan, userLConsentFfmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	} else {
		go program.LFfmpegCompile(LContextAction, LRunId, plan, userLConsentFfmpeg, userLConsentArchive, userPacmanPackageInstallLConsent, userExternalLConsentCommand)
	}
	return LResultAction{RunId: LRunId, StartedAt: time.Now().UTC().Format(time.RFC3339)}, nil
}
