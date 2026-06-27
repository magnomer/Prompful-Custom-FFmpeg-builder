import React from "react";
import { PageHeader, LiveBuildProgress, DescriptionLines } from "./shared";
import { getFfmpegPipeline } from "./logutils";
import { t } from "../i18n";
import { configureOptionText, libraryLicenseLabel, libraryText } from "../catalogText";
import type { LiveProgress } from "./logs";
import emptyStateBlueIcon from "../assets/empty-card-icons/EmptyStateBlue.svg";

export type FFmpegBuildTabProps = {
  ffmpegBuildPlanReview: FfmpegBuildPlanReview | null;
  ffmpegLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  ffmpegProgress: LiveProgress;
  canCancelFfmpeg: boolean;
  approveFfmpegBuildPlan: () => Promise<void>;
  cancelApprovedAction: () => Promise<void>;
  onGoToOptions: () => void;
};

type BuildPlanTabId = "libraries" | "packages" | "options" | "flags" | "operations";

function listOrEmpty<T>(items: T[] | null | undefined): T[] {
  return Array.isArray(items) ? items : [];
}

function FfmpegPlanEmptyCard(props: { onGoToOptions: () => void }) {
  return (
    <section className="card card--blue ffmpeg-build-empty-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={emptyStateBlueIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{t("ffmpegBuild.empty.title")}</h2>
        <DescriptionLines text={t("ffmpegBuild.empty.text")} />
      </div>
      <div className="card__control">
        <button className="button button--primary" type="button" onClick={props.onGoToOptions}>
          {t("ffmpegBuild.empty.action")}<span className="button__chevron" aria-hidden="true">›</span>
        </button>
      </div>
    </section>
  );
}

function planWarningText(warning: PlanWarning): string {
  if (warning.messageKey) return t(warning.messageKey, warning.messageValues ?? {});
  return warning.message ?? "";
}

function planOperationText(operation: PlanOperation): string {
  if (operation.summaryKey) return t(operation.summaryKey, operation.summaryValues ?? {});
  return operation.summary ?? operation.operationName ?? "";
}

function libraryPlanLabel(library: LibraryChoice): string {
  const flags = listOrEmpty(library.configureFlags).join(" ");
  const flagText = flags.length > 0 ? ` — ${flags}` : "";
  return `${libraryText(library, "displayName")} · ${libraryLicenseLabel(library.licenseEffectName)}${flagText}`;
}

function preparationPlanLabel(preparation: LibraryPreparation): string {
  const name = preparation.version ? `${preparation.displayName} ${preparation.version}` : preparation.displayName;
  const buildDependencyPackages = listOrEmpty(preparation.buildDependencyPackages);
  const dependencyText = buildDependencyPackages.length > 0
    ? ` · ${t("approval.preparation.buildDependencies")}: ${buildDependencyPackages.join(" ")}`
    : "";
  return `${name} · ${t(`approval.preparation.method.${preparation.method}`)} · ${preparation.allowedDownloadHost}${dependencyText}`;
}

function optionPlanLabel(option: ConfigureOptionChoice): string {
  const flags = listOrEmpty(option.configureFlags).join(" ");
  const flagText = flags.length > 0 ? ` — ${flags}` : "";
  return `${configureOptionText(option, "displayName")}${flagText}`;
}

function BuildPlanList(props: { title: string; items: string[]; emptyText?: string; code?: boolean }) {
  return (
    <section className="build-plan-list-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{props.title}</h2>
        <span className="result-files__count">{t("result.details.count", { count: props.items.length })}</span>
      </div>
      {props.items.length === 0 ? <p className="empty-text">{props.emptyText ?? t("result.files.empty")}</p> : (
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

function BuildPlanWarnings(props: { warnings: PlanWarning[] }) {
  if (props.warnings.length === 0) return null;
  return (
    <section className="build-plan-warning-card">
      <h3 className="build-plan-section-title">{t("review.warnings.title")}</h3>
      <div className="build-plan-warning-list">
        {props.warnings.map((warning, index) => {
          // A nonfree license boundary makes the build non-redistributable — flag
          // it red even though it is not build-blocking.
          const nonRedistributable = warning.messageKey === "plan.warnings.licenseNonfree";
          return (
            <article className={`build-plan-warning build-plan-warning--${warning.riskLevelName} ${nonRedistributable ? "build-plan-warning--danger" : ""}`} key={`${warning.riskLevelName}-${index}`}>
              <span className="build-plan-warning__icon" aria-hidden="true" />
              <p className="build-plan-warning__text">{planWarningText(warning)}</p>
            </article>
          );
        })}
      </div>
    </section>
  );
}

function BuildPlanConfirmation(props: { planHash: string; expectedConsentText: string }) {
  return (
    <section className="result-location-panel build-plan-confirmation-panel">
      <div className="result-location-row">
        <strong>{t("approval.metadata.planHash")}</strong>
        <span className="build-plan-confirmation-value">{props.planHash}</span>
      </div>
      <div className="result-location-row">
        <strong>{t("approval.metadata.backendPhrase")}</strong>
        <span className="build-plan-confirmation-value">{props.expectedConsentText}</span>
      </div>
    </section>
  );
}

function FfmpegBuildPlanCard(props: { review: FfmpegBuildPlanReview }) {
  const [activeTabId, setActiveTabId] = React.useState<BuildPlanTabId>("libraries");
  const plan = props.review.plan;
  const selectedLibraries = listOrEmpty(plan.selectedLibraries);
  const requiredMsys2PackageNames = listOrEmpty(plan.requiredMsys2PackageNames);
  const libraryPreparations = listOrEmpty(plan.libraryPreparations);
  const selectedConfigureOptions = listOrEmpty(plan.selectedConfigureOptions);
  const generatedOptionFlags = listOrEmpty(plan.generatedOptionFlags);
  const extraConfigureFlags = listOrEmpty(plan.extraConfigureFlags);
  const configureFlags = listOrEmpty(plan.configureFlags);
  const operationsRaw = listOrEmpty(plan.operations);
  const warnings = listOrEmpty(plan.warnings);
  const libraries = React.useMemo(() => selectedLibraries.map(libraryPlanLabel), [selectedLibraries]);
  const packages = React.useMemo(() => [
    ...requiredMsys2PackageNames,
    ...libraryPreparations.map(preparationPlanLabel),
  ], [requiredMsys2PackageNames, libraryPreparations]);
  const options = React.useMemo(() => [
    ...selectedConfigureOptions.map(optionPlanLabel),
    ...generatedOptionFlags,
    ...extraConfigureFlags,
  ], [selectedConfigureOptions, generatedOptionFlags, extraConfigureFlags]);
  const operations = React.useMemo(() => operationsRaw.map(planOperationText), [operationsRaw]);
  const tabs: { id: BuildPlanTabId; label: string; count: number; code?: boolean }[] = [
    { id: "libraries", label: t("ffmpegBuild.plan.tabs.libraries"), count: libraries.length },
    { id: "packages", label: t("ffmpegBuild.plan.tabs.packages"), count: packages.length },
    { id: "options", label: t("ffmpegBuild.plan.tabs.options"), count: options.length },
    { id: "flags", label: t("ffmpegBuild.plan.tabs.flags"), count: configureFlags.length, code: true },
    { id: "operations", label: t("approval.review.operations"), count: operations.length },
  ];

  return (
    <section className="build-plan-shell">
      <BuildPlanConfirmation planHash={plan.planHash} expectedConsentText={props.review.expectedConsentText} />
      <BuildPlanWarnings warnings={warnings} />

      <section className="result-details-card build-plan-details-card">
        <div className="result-details-tabs" role="tablist" aria-label={t("ffmpegBuild.plan.tabs.ariaLabel")}>
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
          {activeTabId === "libraries" && <BuildPlanList title={t("approval.review.selectedLibraries")} items={libraries} emptyText={t("result.details.empty.libraries")} />}
          {activeTabId === "packages" && <BuildPlanList title={t("approval.review.requiredLibraryPackages")} items={packages} emptyText={t("result.details.empty.packages")} code />}
          {activeTabId === "options" && <BuildPlanList title={t("approval.review.selectedBuiltInOptions")} items={options} emptyText={t("result.details.empty.options")} />}
          {activeTabId === "flags" && <BuildPlanList title={t("approval.review.finalConfigureFlags")} items={configureFlags} emptyText={t("result.details.empty.flags")} code />}
          {activeTabId === "operations" && <BuildPlanList title={t("approval.review.operations")} items={operations} emptyText={t("ffmpegBuild.plan.empty.operations")} />}
        </div>
      </section>
    </section>
  );
}

export function FFmpegBuildTab({ ffmpegBuildPlanReview, ffmpegLogEntries, approvedActionPhase, approvedActionStatus, ffmpegProgress, canCancelFfmpeg, cancelApprovedAction, onGoToOptions }: FFmpegBuildTabProps) {
  // A running FFmpeg build takes priority: show live progress, hide the plan.
  const isFfmpegRunning = approvedActionPhase === "ffmpeg";
  const showProgress = isFfmpegRunning || ffmpegLogEntries.length > 0;
  const showApproval = !!ffmpegBuildPlanReview && !isFfmpegRunning;
  return (
    <section className="tab-page ffmpeg-build-page">
      <PageHeader title={t("ffmpegBuild.title")} text={t("ffmpegBuild.intro")} />
      {showApproval && <FfmpegBuildPlanCard review={ffmpegBuildPlanReview} />}
      {!showApproval && !showProgress && (
        <FfmpegPlanEmptyCard onGoToOptions={onGoToOptions} />
      )}
      {showProgress && (
        <LiveBuildProgress
          isActive={isFfmpegRunning}
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
