import React from "react";
import { PageHeader, ApprovalPanel, EmptyReview, LiveBuildProgress } from "./shared";
import { getFfmpegPipeline } from "./logutils";
import { t } from "../i18n";
import type { LiveProgress } from "./logs";

export type FFmpegBuildTabProps = {
  ffmpegBuildPlanReview: FfmpegBuildPlanReview | null;
  ffmpegLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  ffmpegProgress: LiveProgress;
  canCancelFfmpeg: boolean;
  approveFfmpegBuildPlan: () => Promise<void>;
  cancelApprovedAction: () => Promise<void>;
};

export function FFmpegBuildTab({ ffmpegBuildPlanReview, ffmpegLogEntries, approvedActionPhase, approvedActionStatus, ffmpegProgress, canCancelFfmpeg, approveFfmpegBuildPlan, cancelApprovedAction }: FFmpegBuildTabProps) {
  return (
    <section className="tab-page ffmpeg-build-page">
      <PageHeader title={t("ffmpegBuild.title")} text={t("ffmpegBuild.intro")} />
      {ffmpegBuildPlanReview && (
        <ApprovalPanel
          title={t("ffmpegBuild.plan.title")}
          actionName={ffmpegBuildPlanReview.plan.actionName}
          planHash={ffmpegBuildPlanReview.plan.planHash}
          expectedConsentText={ffmpegBuildPlanReview.expectedConsentText}
          operations={ffmpegBuildPlanReview.plan.operations}
          warnings={ffmpegBuildPlanReview.plan.warnings}
          isExecutable={ffmpegBuildPlanReview.plan.isExecutable}
          selectedLibraries={ffmpegBuildPlanReview.plan.selectedLibraries}
          requiredMsys2PackageNames={ffmpegBuildPlanReview.plan.requiredMsys2PackageNames}
          generatedConfigureFlags={ffmpegBuildPlanReview.plan.generatedConfigureFlags}
          selectedConfigureOptions={ffmpegBuildPlanReview.plan.selectedConfigureOptions}
          generatedOptionFlags={ffmpegBuildPlanReview.plan.generatedOptionFlags}
          extraConfigureFlags={ffmpegBuildPlanReview.plan.extraConfigureFlags}
          finalConfigureFlags={ffmpegBuildPlanReview.plan.configureFlags}
          onRequestBackendConfirmation={approveFfmpegBuildPlan}
        />
      )}
      {!ffmpegBuildPlanReview && ffmpegLogEntries.length === 0 && (
        <EmptyReview text={t("ffmpegBuild.empty")} />
      )}
      {!ffmpegBuildPlanReview && ffmpegLogEntries.length > 0 && (
        <LiveBuildProgress
          isActive={approvedActionPhase === "ffmpeg"}
          approvedActionStatus={approvedActionStatus}
          progress={ffmpegProgress}
          pipeline={getFfmpegPipeline()}
          completionLabel={t("ffmpegBuild.progress.completionLabel")}
          onCancel={cancelApprovedAction}
          canCancel={canCancelFfmpeg}
        />
      )}
    </section>
  );
}
