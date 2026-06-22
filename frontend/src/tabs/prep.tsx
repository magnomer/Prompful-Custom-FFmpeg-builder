import React from "react";
import { PageHeader, ApprovalPanel, EmptyReview, LiveBuildProgress } from "./shared";
import { getToolchainPipeline } from "./logutils";
import { t } from "../i18n";
import type { LiveProgress } from "./logs";

export type PrepTabProps = {
  toolchainPreparationPlanReview: ToolchainPreparationPlanReview | null;
  toolchainLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  toolchainProgress: LiveProgress;
  canCancelToolchain: boolean;
  approveToolchainPreparationPlan: () => Promise<void>;
  cancelApprovedAction: () => Promise<void>;
};

export function PrepTab({ toolchainPreparationPlanReview, toolchainLogEntries, approvedActionPhase, approvedActionStatus, toolchainProgress, canCancelToolchain, approveToolchainPreparationPlan, cancelApprovedAction }: PrepTabProps) {
  return (
    <section className="tab-page prep-page">
      <PageHeader title={t("prep.title")} text={t("prep.intro")} />
      {toolchainPreparationPlanReview && (
        <ApprovalPanel
          title={t("prep.plan.title")}
          actionName={toolchainPreparationPlanReview.plan.actionName}
          planHash={toolchainPreparationPlanReview.plan.planHash}
          expectedConsentText={toolchainPreparationPlanReview.expectedConsentText}
          operations={toolchainPreparationPlanReview.plan.operations}
          warnings={toolchainPreparationPlanReview.plan.warnings}
          isExecutable={toolchainPreparationPlanReview.plan.isExecutable}
          onRequestBackendConfirmation={approveToolchainPreparationPlan}
        />
      )}
      {!toolchainPreparationPlanReview && toolchainLogEntries.length === 0 && (
        <EmptyReview text={t("prep.empty")} />
      )}
      {!toolchainPreparationPlanReview && toolchainLogEntries.length > 0 && (
        <LiveBuildProgress
          isActive={approvedActionPhase === "toolchain"}
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
