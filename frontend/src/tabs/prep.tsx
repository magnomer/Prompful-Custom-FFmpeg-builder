import React from "react";
import { PHeaderPageRender, PApprovalPanelRender, PProgressLiveRender, PListReviewRender, PTextDescriptionRender } from "./shared";
import { LPipelineToolchainGet } from "./logutils";
import { LLocaleGet, LLocaleMessageGet, LLocaleTextGet } from "../i18n";
import notPreparedIcon from "../assets/prep-card-icons/NotPrepared.svg";
import planToImplementIcon from "../assets/prep-card-icons/PlanToImplement.svg";
import planReadyIcon from "../assets/prep-card-icons/PlanReady.svg";
import type { LProgressLive } from "./logs";

export type LPreparationProperties = {
  toolchainPreparationPlanReview: LReviewToolchain | null;
  toolchainLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  toolchainProgress: LProgressLive;
  canCancelToolchain: boolean;
  toolchainStatus: LStatusToolchain | null;
  installedToolchainProfiles: LStatusToolchain[];
  toolchainVerification: LVerificationToolchain | null;
  isVerifyingToolchain: boolean;
  isApprovedActionRunning: boolean;
  isPlanningToolchain: boolean;
  currentShellProfileName: string;
  configuredMsys2PackageNames: string[];
  approveToolchainPreparationPlan: () => Promise<void>;
  LPlanToolchainCancel: () => void;
  cancelApprovedAction: () => Promise<void>;
  LActionApprovedClear: () => void;
  onVerifyToolchain: () => void;
  onReuseToolchain: () => void;
  onReinstallToolchain: () => void;
  onGoToBuildConfig: () => void;
  onClearBuildEnvironments: () => void;
};

function PCardToolchainRender(props: { onGoToBuildConfig: () => void }) {
  return (
    <section className="card card--blue">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={notPreparedIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("prep.empty.title")}</h2>
        <PTextDescriptionRender text={LLocaleTextGet("prep.empty.text")} />
      </div>
      <div className="card__control">
        <button className="button button--primary" type="button" onClick={props.onGoToBuildConfig}>
          {LLocaleTextGet("prep.empty.action")}<span className="button__chevron" aria-hidden="true">›</span>
        </button>
      </div>
    </section>
  );
}

function PPanelProfileRender(props: { profiles: LStatusToolchain[]; currentShellProfileName: string; onClearBuildEnvironments: () => void }) {
  const { profiles, currentShellProfileName, onClearBuildEnvironments } = props;
  return (
    <section className="prep-profiles">
      <div className="prep-profiles__header">
        <h2 className="prep-profiles__title">{LLocaleTextGet("prep.profiles.title")}</h2>
        <button className="button prep-profiles__clear" type="button" onClick={onClearBuildEnvironments}>
          {LLocaleTextGet("prep.profiles.clear")}
        </button>
      </div>
      <ul className="prep-profiles__list">
        {profiles.map((profile) => {
          const isCurrent = profile.windowsShellProfileName === currentShellProfileName;
          const detail = profile.packageCount > 0
            ? LLocaleTextGet("prep.profiles.detail", { count: profile.packageCount })
            : LLocaleTextGet("prep.profiles.detailNoManifest");
          return (
            <li className={`prep-profiles__item ${isCurrent ? "prep-profiles__item--current" : ""}`} key={profile.windowsShellProfileName}>
              <span className="prep-profiles__name">{profile.windowsShellProfileName}</span>
              <span className="prep-profiles__detail">{detail}</span>
              {isCurrent && <span className="prep-profiles__tag">{LLocaleTextGet("prep.profiles.current")}</span>}
            </li>
          );
        })}
      </ul>
    </section>
  );
}

function PNoticeToolchainRender(props: { onShowExisting: () => void }) {
  return (
    <section className="prep-existing" aria-label={LLocaleTextGet("prep.existingNotice.title")}>
      <span className="prep-existing__mark" aria-hidden="true"><img className="prep-existing__icon" src={planReadyIcon} alt="" /></span>
      <div className="prep-existing__body">
        <h2 className="prep-existing__title">{LLocaleTextGet("prep.existingNotice.title")}</h2>
        <p className="prep-existing__text">{LLocaleTextGet("prep.existingNotice.detail")}</p>
      </div>
      <button className="button prep-existing__action" type="button" onClick={props.onShowExisting}>
        {LLocaleTextGet("prep.existingNotice.action")}
      </button>
    </section>
  );
}

// Set difference of package lists: what the current config adds vs what the
// prepared toolchain has but the config no longer lists.
function LPackageDriftGet(configured: string[], prepared: string[]): { added: string[]; removed: string[] } {
  const configuredSet = new Set(configured);
  const preparedSet = new Set(prepared);
  return {
    added: configured.filter((name) => !preparedSet.has(name)).sort(),
    removed: prepared.filter((name) => !configuredSet.has(name)).sort(),
  };
}

function PCardRecoveryRender(props: {
  status: LStatusToolchain;
  verification: LVerificationToolchain | null;
  isVerifying: boolean;
  configuredPackageNames: string[];
  onVerify: () => void;
  onReuse: () => void;
  onReinstall: () => void;
  isPlanning: boolean;
}) {
  const { status, verification, isVerifying, configuredPackageNames, onVerify, onReuse, onReinstall, isPlanning } = props;
  const installedDate = status.createdAt ? new Date(status.createdAt).toLocaleString(LLocaleGet() === "ko" ? "ko-KR" : "en-US") : "";
  const hasManifest = status.packageCount > 0;
  // Drift is only meaningful when the prepared profile recorded its package list.
  const drift = hasManifest ? LPackageDriftGet(configuredPackageNames, status.packageNames) : { added: [], removed: [] };
  const hasDrift = drift.added.length > 0 || drift.removed.length > 0;

  return (
    <section className="card card--green">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={planReadyIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("prep.recovery.title")}</h2>
        {hasManifest ? (
          <p className="card__desc">{LLocaleTextGet("prep.recovery.summary", { date: installedDate, profile: status.windowsShellProfileName, count: status.packageCount })}</p>
        ) : (
          <p className="card__desc">{LLocaleTextGet("prep.recovery.summaryNoManifest")}</p>
        )}
      </div>

      {hasDrift && (
        <div className="prep-recovery__drift">
          <p className="empty-text empty-text--warn">{LLocaleTextGet("prep.recovery.driftSummary", { added: drift.added.length, removed: drift.removed.length })}</p>
          {drift.added.length > 0 && <PListReviewRender title={LLocaleTextGet("prep.recovery.driftAdded")} items={drift.added} dense />}
          {drift.removed.length > 0 && <PListReviewRender title={LLocaleTextGet("prep.recovery.driftRemoved")} items={drift.removed} dense />}
        </div>
      )}

      {verification && (
        verification.verified ? (
          <p className="empty-text empty-text--ok">{LLocaleMessageGet(verification)}</p>
        ) : (
          <div className="prep-recovery__verify-result">
            <p className="empty-text empty-text--warn">{LLocaleMessageGet(verification)}</p>
            {verification.missingPackageNames.length > 0 && (
              <PListReviewRender title={LLocaleTextGet("prep.recovery.missingPackages")} items={verification.missingPackageNames} />
            )}
          </div>
        )
      )}

      <div className="card__control">
        <button className="button button--primary" type="button" onClick={onReuse}>{LLocaleTextGet("prep.recovery.reuse")}</button>
        <button className="button" type="button" disabled={isPlanning} onClick={onReinstall}>{LLocaleTextGet("prep.recovery.reinstall")}</button>
        <button className="button" type="button" disabled={isVerifying} onClick={onVerify}>
          {isVerifying ? LLocaleTextGet("prep.recovery.verifying") : LLocaleTextGet("prep.recovery.verify")}
        </button>
      </div>
    </section>
  );
}

export function PPrepRender({ toolchainPreparationPlanReview, toolchainLogEntries, approvedActionPhase, approvedActionStatus, toolchainProgress, canCancelToolchain, toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain, isApprovedActionRunning, isPlanningToolchain, currentShellProfileName, configuredMsys2PackageNames, approveToolchainPreparationPlan, LPlanToolchainCancel, cancelApprovedAction, LActionApprovedClear, onVerifyToolchain, onReuseToolchain, onReinstallToolchain, onGoToBuildConfig, onClearBuildEnvironments }: LPreparationProperties) {
  // A running toolchain action takes priority: show its live progress and hide
  // the approval panel, so a refused/duplicate confirm can never strand the UI on
  // the plan while the install is actually progressing.
  const isToolchainRunning = approvedActionPhase === "toolchain";
  const showProgress = isToolchainRunning || toolchainLogEntries.length > 0;
  const showApproval = !!toolchainPreparationPlanReview && !isToolchainRunning;
  const isIdle = !showApproval && !showProgress;
  const showRecovery = isIdle && !!toolchainStatus?.installed;

  return (
    <section className="tab-page prep-page">
      <PHeaderPageRender title={LLocaleTextGet("prep.title")} text={LLocaleTextGet("prep.intro")} />
      {isIdle && installedToolchainProfiles.length > 0 && (
        <PPanelProfileRender profiles={installedToolchainProfiles} currentShellProfileName={currentShellProfileName} onClearBuildEnvironments={onClearBuildEnvironments} />
      )}
      {showApproval && toolchainStatus?.installed && (
        <PNoticeToolchainRender onShowExisting={LPlanToolchainCancel} />
      )}
      {showApproval && (
        <PApprovalPanelRender
          variant="blue"
          icon={planToImplementIcon}
          title={LLocaleTextGet("prep.plan.title")}
          actionName={toolchainPreparationPlanReview.plan.actionName}
          planHash={toolchainPreparationPlanReview.plan.planHash}
          expectedLConsentText={toolchainPreparationPlanReview.expectedLConsentText}
          operations={toolchainPreparationPlanReview.plan.operations}
          warnings={toolchainPreparationPlanReview.plan.warnings}
          isExecutable={toolchainPreparationPlanReview.plan.isExecutable}
          onCancelPlan={LPlanToolchainCancel}
          onRequestBackendConfirmation={approveToolchainPreparationPlan}
          isConfirmationBusy={isApprovedActionRunning}
        />
      )}
      {showRecovery && (
        <PCardRecoveryRender
          status={toolchainStatus}
          verification={toolchainVerification}
          isVerifying={isVerifyingToolchain}
          configuredPackageNames={configuredMsys2PackageNames}
          onVerify={onVerifyToolchain}
          onReuse={onReuseToolchain}
          onReinstall={onReinstallToolchain}
          isPlanning={isPlanningToolchain}
        />
      )}
      {isIdle && !showRecovery && (
        <PCardToolchainRender onGoToBuildConfig={onGoToBuildConfig} />
      )}
      {showProgress && (
        <PProgressLiveRender
          isActive={isToolchainRunning}
          approvedActionStatus={approvedActionStatus}
          progress={toolchainProgress}
          pipeline={LPipelineToolchainGet()}
          completionLabel={LLocaleTextGet("prep.progress.completionLabel")}
          onCancel={cancelApprovedAction}
          canCancel={canCancelToolchain}
          onClear={LActionApprovedClear}
        />
      )}
    </section>
  );
}
