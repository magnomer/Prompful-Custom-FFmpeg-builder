import React from "react";
import { PageHeader, ApprovalPanel, LiveBuildProgress, ReviewList, DescriptionLines } from "./shared";
import { getToolchainPipeline } from "./logutils";
import { t } from "../i18n";
import notPreparedIcon from "../assets/prep-card-icons/NotPrepared.svg";
import planToImplementIcon from "../assets/prep-card-icons/PlanToImplement.svg";
import planReadyIcon from "../assets/prep-card-icons/PlanReady.svg";
import type { LiveProgress } from "./logs";

export type PrepTabProps = {
  toolchainPreparationPlanReview: ToolchainPreparationPlanReview | null;
  toolchainLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  toolchainProgress: LiveProgress;
  canCancelToolchain: boolean;
  toolchainStatus: ToolchainStatus | null;
  installedToolchainProfiles: ToolchainStatus[];
  toolchainVerification: ToolchainVerification | null;
  isVerifyingToolchain: boolean;
  isApprovedActionRunning: boolean;
  currentShellProfileName: string;
  configuredMsys2PackageNames: string[];
  approveToolchainPreparationPlan: () => Promise<void>;
  cancelToolchainPreparationPlan: () => void;
  cancelApprovedAction: () => Promise<void>;
  onVerifyToolchain: () => void;
  onReuseToolchain: () => void;
  onReinstallToolchain: () => void;
  onGoToBuildConfig: () => void;
  onClearBuildEnvironments: () => void;
};

function ToolchainEmptyCard(props: { onGoToBuildConfig: () => void }) {
  return (
    <section className="card card--blue">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={notPreparedIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{t("prep.empty.title")}</h2>
        <DescriptionLines text={t("prep.empty.text")} />
      </div>
      <div className="card__control">
        <button className="button button--primary" type="button" onClick={props.onGoToBuildConfig}>
          {t("prep.empty.action")}<span className="button__chevron" aria-hidden="true">›</span>
        </button>
      </div>
    </section>
  );
}

function InstalledProfilesPanel(props: { profiles: ToolchainStatus[]; currentShellProfileName: string; onClearBuildEnvironments: () => void }) {
  const { profiles, currentShellProfileName, onClearBuildEnvironments } = props;
  return (
    <section className="prep-profiles">
      <div className="prep-profiles__header">
        <h2 className="prep-profiles__title">{t("prep.profiles.title")}</h2>
        <button className="button prep-profiles__clear" type="button" onClick={onClearBuildEnvironments}>
          {t("prep.profiles.clear")}
        </button>
      </div>
      <ul className="prep-profiles__list">
        {profiles.map((profile) => {
          const isCurrent = profile.windowsShellProfileName === currentShellProfileName;
          const detail = profile.packageCount > 0
            ? t("prep.profiles.detail", { count: profile.packageCount })
            : t("prep.profiles.detailNoManifest");
          return (
            <li className={`prep-profiles__item ${isCurrent ? "prep-profiles__item--current" : ""}`} key={profile.windowsShellProfileName}>
              <span className="prep-profiles__name">{profile.windowsShellProfileName}</span>
              <span className="prep-profiles__detail">{detail}</span>
              {isCurrent && <span className="prep-profiles__tag">{t("prep.profiles.current")}</span>}
            </li>
          );
        })}
      </ul>
    </section>
  );
}

// Set difference of package lists: what the current config adds vs what the
// prepared toolchain has but the config no longer lists.
function computePackageDrift(configured: string[], prepared: string[]): { added: string[]; removed: string[] } {
  const configuredSet = new Set(configured);
  const preparedSet = new Set(prepared);
  return {
    added: configured.filter((name) => !preparedSet.has(name)).sort(),
    removed: prepared.filter((name) => !configuredSet.has(name)).sort(),
  };
}

function ToolchainRecoveryCard(props: {
  status: ToolchainStatus;
  verification: ToolchainVerification | null;
  isVerifying: boolean;
  configuredPackageNames: string[];
  onVerify: () => void;
  onReuse: () => void;
  onReinstall: () => void;
}) {
  const { status, verification, isVerifying, configuredPackageNames, onVerify, onReuse, onReinstall } = props;
  const installedDate = status.createdAt ? new Date(status.createdAt).toLocaleString() : "";
  const hasManifest = status.packageCount > 0;
  // Drift is only meaningful when the prepared profile recorded its package list.
  const drift = hasManifest ? computePackageDrift(configuredPackageNames, status.packageNames) : { added: [], removed: [] };
  const hasDrift = drift.added.length > 0 || drift.removed.length > 0;

  return (
    <section className="card card--green">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={planReadyIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{t("prep.recovery.title")}</h2>
        {hasManifest ? (
          <p className="card__desc">{t("prep.recovery.summary", { date: installedDate, profile: status.windowsShellProfileName, count: status.packageCount })}</p>
        ) : (
          <p className="card__desc">{t("prep.recovery.summaryNoManifest")}</p>
        )}
      </div>

      {hasDrift && (
        <div className="prep-recovery__drift">
          <p className="empty-text empty-text--warn">{t("prep.recovery.driftSummary", { added: drift.added.length, removed: drift.removed.length })}</p>
          {drift.added.length > 0 && <ReviewList title={t("prep.recovery.driftAdded")} items={drift.added} dense />}
          {drift.removed.length > 0 && <ReviewList title={t("prep.recovery.driftRemoved")} items={drift.removed} dense />}
        </div>
      )}

      {verification && (
        verification.verified ? (
          <p className="empty-text empty-text--ok">{verification.message}</p>
        ) : (
          <div className="prep-recovery__verify-result">
            <p className="empty-text empty-text--warn">{verification.message}</p>
            {verification.missingPackageNames.length > 0 && (
              <ReviewList title={t("prep.recovery.missingPackages")} items={verification.missingPackageNames} />
            )}
          </div>
        )
      )}

      <div className="card__control">
        <button className="button button--primary" type="button" onClick={onReuse}>{t("prep.recovery.reuse")}</button>
        <button className="button" type="button" onClick={onReinstall}>{t("prep.recovery.reinstall")}</button>
        <button className="button" type="button" disabled={isVerifying} onClick={onVerify}>
          {isVerifying ? t("prep.recovery.verifying") : t("prep.recovery.verify")}
        </button>
      </div>
    </section>
  );
}

export function PrepTab({ toolchainPreparationPlanReview, toolchainLogEntries, approvedActionPhase, approvedActionStatus, toolchainProgress, canCancelToolchain, toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain, isApprovedActionRunning, currentShellProfileName, configuredMsys2PackageNames, approveToolchainPreparationPlan, cancelToolchainPreparationPlan, cancelApprovedAction, onVerifyToolchain, onReuseToolchain, onReinstallToolchain, onGoToBuildConfig, onClearBuildEnvironments }: PrepTabProps) {
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
      <PageHeader title={t("prep.title")} text={t("prep.intro")} />
      {isIdle && installedToolchainProfiles.length > 0 && (
        <InstalledProfilesPanel profiles={installedToolchainProfiles} currentShellProfileName={currentShellProfileName} onClearBuildEnvironments={onClearBuildEnvironments} />
      )}
      {showApproval && (
        <ApprovalPanel
          variant="blue"
          icon={planToImplementIcon}
          title={t("prep.plan.title")}
          actionName={toolchainPreparationPlanReview.plan.actionName}
          planHash={toolchainPreparationPlanReview.plan.planHash}
          expectedConsentText={toolchainPreparationPlanReview.expectedConsentText}
          operations={toolchainPreparationPlanReview.plan.operations}
          warnings={toolchainPreparationPlanReview.plan.warnings}
          isExecutable={toolchainPreparationPlanReview.plan.isExecutable}
          onCancelPlan={cancelToolchainPreparationPlan}
          onRequestBackendConfirmation={approveToolchainPreparationPlan}
          isConfirmationBusy={isApprovedActionRunning}
        />
      )}
      {showRecovery && (
        <ToolchainRecoveryCard
          status={toolchainStatus}
          verification={toolchainVerification}
          isVerifying={isVerifyingToolchain}
          configuredPackageNames={configuredMsys2PackageNames}
          onVerify={onVerifyToolchain}
          onReuse={onReuseToolchain}
          onReinstall={onReinstallToolchain}
        />
      )}
      {isIdle && !showRecovery && (
        <ToolchainEmptyCard onGoToBuildConfig={onGoToBuildConfig} />
      )}
      {showProgress && (
        <LiveBuildProgress
          isActive={isToolchainRunning}
          approvedActionStatus={approvedActionStatus}
          progress={toolchainProgress}
          pipeline={getToolchainPipeline()}
          completionLabel={t("prep.progress.completionLabel")}
          onCancel={cancelApprovedAction}
          canCancel={canCancelToolchain}
        />
      )}
    </section>
  );
}
