import React from "react";
import { PHeaderPageRender, PProgressBuildLiveRender, PTextDescriptionRender } from "./shared";
import { LPipelineFFmpegGet } from "./logutils";
import { LLocaleTextGet } from "../i18n";
import { LOptionTextGet, LLibraryLicenseLabelGet, LLibraryTextGet } from "../catalogText";
import type { LProgressLive } from "./logs";
import emptyStateBlueIcon from "../assets/empty-card-icons/EmptyStateBlue.svg";

export type PBuildProps = {
  ffmpegBuildPlanReview: LReviewFFmpeg | null;
  ffmpegLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  ffmpegProgress: LProgressLive;
  canCancelFfmpeg: boolean;
  approveFfmpegBuildPlan: () => Promise<void>;
  cancelApprovedAction: () => Promise<void>;
  clearApprovedAction: () => void;
  onGoToOptions: () => void;
};

type PBuildPlanTabId = "libraries" | "packages" | "options" | "flags" | "operations";

function LListEmptyNormalize<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : [];
}

function PCardPlanRender(props: { onGoToOptions: () => void }) {
  return (
    <section className="card card--blue ffmpeg-build-empty-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={emptyStateBlueIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("ffmpegBuild.empty.title")}</h2>
        <PTextDescriptionRender text={LLocaleTextGet("ffmpegBuild.empty.text")} />
      </div>
      <div className="card__control">
        <button className="button button--primary" type="button" onClick={props.onGoToOptions}>
          {LLocaleTextGet("ffmpegBuild.empty.action")}<span className="button__chevron" aria-hidden="true">›</span>
        </button>
      </div>
    </section>
  );
}

function LWarningPlanTextGet(warning: LWarningPlan): string {
  if (warning.messageKey) return LLocaleTextGet(warning.messageKey, warning.messageValues ?? {});
  return warning.message ?? "";
}

function LOperationPlanTextGet(operation: LOperationPlan): string {
  if (operation.summaryKey) return LLocaleTextGet(operation.summaryKey, operation.summaryValues ?? {});
  return operation.summary ?? operation.operationName ?? "";
}

function LLibraryPlanLabelGet(library: LLibraryChoice): string {
  const flags = LListEmptyNormalize(library.configureFlags).join(" ");
  const flagText = flags.length > 0 ? ` — ${flags}` : "";
  return `${LLibraryTextGet(library, "displayName")} · ${LLibraryLicenseLabelGet(library.licenseEffectName)}${flagText}`;
}

function LPreparationPlanLabelGet(preparation: LLibraryPreparation): string {
  const name = preparation.version ? `${preparation.displayName} ${preparation.version}` : preparation.displayName;
  const buildDependencyPackages = LListEmptyNormalize(preparation.buildDependencyPackages);
  const dependencyText = buildDependencyPackages.length > 0
    ? ` · ${LLocaleTextGet("approval.preparation.buildDependencies")}: ${buildDependencyPackages.join(" ")}`
    : "";
  return `${name} · ${LLocaleTextGet(`approval.preparation.method.${preparation.method}`)} · ${preparation.allowedDownloadHost}${dependencyText}`;
}

function LOptionPlanLabelGet(option: LOptionChoice): string {
  const flags = LListEmptyNormalize(option.configureFlags).join(" ");
  const flagText = flags.length > 0 ? ` — ${flags}` : "";
  return `${LOptionTextGet(option, "displayName")}${flagText}`;
}

function PListPlanRender(props: { title: string; items: string[]; emptyText?: string; code?: boolean }) {
  return (
    <section className="build-plan-list-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{props.title}</h2>
        <span className="result-files__count">{LLocaleTextGet("result.details.count", { count: props.items.length })}</span>
      </div>
      {props.items.length === 0 ? <p className="empty-text">{props.emptyText ?? LLocaleTextGet("result.files.empty")}</p> : (
        <div className="result-file-list result-plan-list build-plan-list">
          {props.items.map((item, index) => (
            <article className="result-plan-item build-plan-list__item" key={`${item}-${index}`}>
              <span className="result-plan-item__index">{String(index + 1).padStart(2, "0")}</span>
              {props.code ? <code className="result-plan-item__text result-plan-item__text--code">{item}</code> : <span className="result-plan-item__text">{item}</span>}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function PListWarningRender(props: { warnings: LWarningPlan[] }) {
  if (props.warnings.length === 0) return null;
  return (
    <section className="build-plan-warning-card">
      <h3 className="build-plan-section-title">{LLocaleTextGet("review.warnings.title")}</h3>
      <div className="build-plan-warning-list">
        {props.warnings.map((warning, index) => {
          // A nonfree license boundary makes the build non-redistributable — flag
          // it red even though it is not build-blocking.
          const nonRedistributable = warning.messageKey === "plan.warnings.licenseNonfree";
          return (
            <article className={`build-plan-warning build-plan-warning--${warning.riskLevelName} ${nonRedistributable ? "build-plan-warning--danger" : ""}`} key={`${warning.riskLevelName}-${index}`}>
              <span className="build-plan-warning__icon" aria-hidden="true" />
              <p className="build-plan-warning__text">{LWarningPlanTextGet(warning)}</p>
            </article>
          );
        })}
      </div>
    </section>
  );
}

function PPanelConfirmationRender(props: { planHash: string; expectedLConsentText: string }) {
  return (
    <section className="result-location-panel build-plan-confirmation-panel">
      <div className="result-location-row">
        <strong>{LLocaleTextGet("approval.metadata.planHash")}</strong>
        <span className="build-plan-confirmation-value">{props.planHash}</span>
      </div>
      <div className="result-location-row">
        <strong>{LLocaleTextGet("approval.metadata.backendPhrase")}</strong>
        <span className="build-plan-confirmation-value">{props.expectedLConsentText}</span>
      </div>
    </section>
  );
}

function PCardBuildRender(props: { review: LReviewFFmpeg }) {
  const [activeTabId, setActiveTabId] = React.useState<PBuildPlanTabId>("libraries");
  const plan = props.review.plan;
  const selectedLibraries = LListEmptyNormalize(plan.selectedLibraries);
  const requiredMsys2PackageNames = LListEmptyNormalize(plan.requiredMsys2PackageNames);
  const libraryPreparations = LListEmptyNormalize(plan.libraryPreparations);
  const selectedConfigureOptions = LListEmptyNormalize(plan.selectedConfigureOptions);
  const generatedOptionFlags = LListEmptyNormalize(plan.generatedOptionFlags);
  const extraConfigureFlags = LListEmptyNormalize(plan.extraConfigureFlags);
  const configureFlags = LListEmptyNormalize(plan.configureFlags);
  const operationsRaw = LListEmptyNormalize(plan.operations);
  const warnings = LListEmptyNormalize(plan.warnings);
  const libraries = React.useMemo(() => selectedLibraries.map(LLibraryPlanLabelGet), [selectedLibraries]);
  const packages = React.useMemo(() => [
    ...requiredMsys2PackageNames,
    ...libraryPreparations.map(LPreparationPlanLabelGet),
  ], [requiredMsys2PackageNames, libraryPreparations]);
  const options = React.useMemo(() => [
    ...selectedConfigureOptions.map(LOptionPlanLabelGet),
    ...generatedOptionFlags,
    ...extraConfigureFlags,
  ], [selectedConfigureOptions, generatedOptionFlags, extraConfigureFlags]);
  const operations = React.useMemo(() => operationsRaw.map(LOperationPlanTextGet), [operationsRaw]);
  const tabs: { id: PBuildPlanTabId; label: string; count: number; code?: boolean }[] = [
    { id: "libraries", label: LLocaleTextGet("ffmpegBuild.plan.tabs.libraries"), count: libraries.length },
    { id: "packages", label: LLocaleTextGet("ffmpegBuild.plan.tabs.packages"), count: packages.length },
    { id: "options", label: LLocaleTextGet("ffmpegBuild.plan.tabs.options"), count: options.length },
    { id: "flags", label: LLocaleTextGet("ffmpegBuild.plan.tabs.flags"), count: configureFlags.length, code: true },
    { id: "operations", label: LLocaleTextGet("approval.review.operations"), count: operations.length },
  ];

  return (
    <section className="build-plan-shell">
      <PPanelConfirmationRender planHash={plan.planHash} expectedLConsentText={props.review.expectedLConsentText} />
      <PListWarningRender warnings={warnings} />

      <section className="result-details-card build-plan-details-card">
        <div className="result-details-tabs" role="tablist" aria-label={LLocaleTextGet("ffmpegBuild.plan.tabs.ariaLabel")}>
          {tabs.map((tab) => (
            <button
              className={`result-details-tab ${activeTabId === tab.id ? "result-details-tab--active" : ""}`}
              type="button"
              role="tab"
              aria-selected={activeTabId === tab.id}
              key={tab.id}
              onClick={() => setActiveTabId(tab.id)}
            >
              <span>{tab.label}</span>
              <span className="result-details-tab__count">{tab.count}</span>
            </button>
          ))}
        </div>
        <div className="result-details-body">
          {activeTabId === "libraries" && <PListPlanRender title={LLocaleTextGet("approval.review.selectedLibraries")} items={libraries} emptyText={LLocaleTextGet("result.details.empty.libraries")} />}
          {activeTabId === "packages" && <PListPlanRender title={LLocaleTextGet("approval.review.requiredLibraryPackages")} items={packages} emptyText={LLocaleTextGet("result.details.empty.packages")} code />}
          {activeTabId === "options" && <PListPlanRender title={LLocaleTextGet("approval.review.selectedBuiltInOptions")} items={options} emptyText={LLocaleTextGet("result.details.empty.options")} />}
          {activeTabId === "flags" && <PListPlanRender title={LLocaleTextGet("approval.review.finalConfigureFlags")} items={configureFlags} emptyText={LLocaleTextGet("result.details.empty.flags")} code />}
          {activeTabId === "operations" && <PListPlanRender title={LLocaleTextGet("approval.review.operations")} items={operations} emptyText={LLocaleTextGet("ffmpegBuild.plan.empty.operations")} />}
        </div>
      </section>
    </section>
  );
}

export function PBuildRender({ ffmpegBuildPlanReview, ffmpegLogEntries, approvedActionPhase, approvedActionStatus, ffmpegProgress, canCancelFfmpeg, cancelApprovedAction, clearApprovedAction, onGoToOptions }: PBuildProps) {
  // A running FFmpeg build takes priority: show live progress, hide the plan.
  const isFfmpegRunning = approvedActionPhase === "ffmpeg";
  const showProgress = isFfmpegRunning || ffmpegLogEntries.length > 0;
  const showApproval = !!ffmpegBuildPlanReview && !isFfmpegRunning;
  return (
    <section className="tab-page ffmpeg-build-page">
      <PHeaderPageRender title={LLocaleTextGet("ffmpegBuild.title")} text={LLocaleTextGet("ffmpegBuild.intro")} />
      {showApproval && <PCardBuildRender review={ffmpegBuildPlanReview} />}
      {!showApproval && !showProgress && (
        <PCardPlanRender onGoToOptions={onGoToOptions} />
      )}
      {showProgress && (
        <PProgressBuildLiveRender
          isActive={isFfmpegRunning}
          approvedActionStatus={approvedActionStatus}
          progress={ffmpegProgress}
          pipeline={LPipelineFFmpegGet()}
          completionLabel={LLocaleTextGet("ffmpegBuild.progress.completionLabel")}
          onCancel={cancelApprovedAction}
          canCancel={canCancelFfmpeg}
          onClear={clearApprovedAction}
        />
      )}
    </section>
  );
}
