import React from "react";
import { PHeaderPageRender, PProgressLiveRender, PTextDescriptionRender, PNoticeExpiredRender } from "./shared";
import { LPipelineFfmpegGet } from "./logutils";
import { LLocaleGet, LLocaleTextGet } from "../i18n";
import { LOptionTextGet, LLicenseLabelGet, LLibraryNameGet, LLibraryTextGet } from "../catalogText";
import type { LProgressLive } from "./logs";
import emptyStateBlueIcon from "../assets/empty-card-icons/EmptyStateBlue.svg";
import { LTabKeyDown } from "./tabkeyboard";

export type LBuildProperties = {
  ffmpegBuildPlanReview: LReviewFfmpeg | null;
  ffmpegLogEntries: { timestamp: string; level: "info" | "warn" | "error"; message: string }[];
  approvedActionPhase: "toolchain" | "ffmpeg" | null;
  approvedActionStatus: string;
  ffmpegStalledAddresses: string[];
  ffmpegProgress: LProgressLive;
  canCancelFfmpeg: boolean;
  isReviewExpired: boolean;
  approveFfmpegBuildPlan: () => Promise<void>;
  retryFfmpegBuildPlan: () => Promise<void>;
  cancelApprovedAction: () => Promise<void>;
  LActionApprovedClear: () => void;
  onGoToOptions: () => void;
};

type LBuildPlanIdentifier = "libraries" | "packages" | "options" | "flags" | "operations";

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

function LWarningTextGet(warning: LWarningPlan): string {
  if (warning.messageKey) return LLocaleTextGet(warning.messageKey, warning.messageValues ?? {});
  return warning.message ?? "";
}

function LOperationTextGet(operation: LOperationPlan): string {
  if (operation.summaryKey) return LLocaleTextGet(operation.summaryKey, operation.summaryValues ?? {});
  return operation.summary ?? operation.operationName ?? "";
}

function LLibraryLabelGet(library: LLibraryChoice): string {
  const flags = LListEmptyNormalize(library.configureFlags).join(" ");
  const flagText = flags.length > 0 ? ` — ${flags}` : "";
  return `${LLibraryTextGet(library, "displayName")} · ${LLicenseLabelGet(library.licenseEffectName)}${flagText}`;
}

function LPreparationLabelGet(preparation: LLibraryPreparation): string {
  const localizedName = LLibraryNameGet(preparation.libraryId, preparation.displayName);
  const name = preparation.version ? `${localizedName} ${preparation.version}` : localizedName;
  const buildDependencyPackages = LListEmptyNormalize(preparation.buildDependencyPackages);
  const dependencyText = buildDependencyPackages.length > 0
    ? ` · ${LLocaleTextGet("approval.preparation.buildDependencies")}: ${buildDependencyPackages.join(" ")}`
    : "";
  return `${name} · ${LLocaleTextGet(`approval.preparation.method.${preparation.method}`)} · ${preparation.allowedDownloadHost}${dependencyText}`;
}

function LOptionLabelGet(option: LOptionChoice): string {
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
              <p className="build-plan-warning__text">{LWarningTextGet(warning)}</p>
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

function PCardBuildRender(props: { review: LReviewFfmpeg; isReviewExpired: boolean }) {
  const locale = LLocaleGet();
  const [activeTabId, setActiveTabId] = React.useState<LBuildPlanIdentifier>("libraries");
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
  const libraries = React.useMemo(() => selectedLibraries.map(LLibraryLabelGet), [selectedLibraries, locale]);
  const packages = React.useMemo(() => [
    ...requiredMsys2PackageNames,
    ...libraryPreparations.map(LPreparationLabelGet),
  ], [requiredMsys2PackageNames, libraryPreparations, locale]);
  const options = React.useMemo(() => [
    ...selectedConfigureOptions.map(LOptionLabelGet),
    ...generatedOptionFlags,
    ...extraConfigureFlags,
  ], [selectedConfigureOptions, generatedOptionFlags, extraConfigureFlags, locale]);
  const operations = React.useMemo(() => operationsRaw.map(LOperationTextGet), [operationsRaw, locale]);
  const tabs: { id: LBuildPlanIdentifier; label: string; count: number; code?: boolean }[] = [
    { id: "libraries", label: LLocaleTextGet("ffmpegBuild.plan.tabs.libraries"), count: libraries.length },
    { id: "packages", label: LLocaleTextGet("ffmpegBuild.plan.tabs.packages"), count: packages.length },
    { id: "options", label: LLocaleTextGet("ffmpegBuild.plan.tabs.options"), count: options.length },
    { id: "flags", label: LLocaleTextGet("ffmpegBuild.plan.tabs.flags"), count: configureFlags.length, code: true },
    { id: "operations", label: LLocaleTextGet("approval.review.operations"), count: operations.length },
  ];

  return (
    <section className="build-plan-shell">
      <PPanelConfirmationRender planHash={plan.planHash} expectedLConsentText={props.review.expectedLConsentText} />
      {props.isReviewExpired && <PNoticeExpiredRender />}
      <PListWarningRender warnings={warnings} />

      <section className="result-details-card build-plan-details-card">
        <div className="result-details-tabs" role="tablist" aria-label={LLocaleTextGet("ffmpegBuild.plan.tabs.ariaLabel")}>
          {tabs.map((tab, index) => (
            <button
              className={`result-details-tab ${activeTabId === tab.id ? "result-details-tab--active" : ""}`}
              type="button"
              role="tab"
              aria-selected={activeTabId === tab.id}
              aria-controls="ffmpeg-plan-tabpanel"
              id={`ffmpeg-plan-tab-${tab.id}`}
              tabIndex={activeTabId === tab.id ? 0 : -1}
              key={tab.id}
              onClick={() => setActiveTabId(tab.id)}
              onKeyDown={(event) => LTabKeyDown(event, index, tabs.length, (nextIndex) => setActiveTabId(tabs[nextIndex].id))}
            >
              <span>{tab.label}</span>
              <span className="result-details-tab__count">{tab.count}</span>
            </button>
          ))}
        </div>
        <div className="result-details-body" role="tabpanel" id="ffmpeg-plan-tabpanel" aria-labelledby={`ffmpeg-plan-tab-${activeTabId}`}>
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

// A transient-network stall is halted but retryable, not a hard failure: an
// orange (never red, never green) banner lists the mirror addresses that were
// tried and offers Retry, which resumes the same approved action from cache.
function PBuildStalledRender(props: { addresses: string[]; onRetry: () => void }) {
  return (
    <section className="build-plan-stalled-card">
      <div className="build-plan-stalled-head">
        <span className="build-plan-stalled__icon" aria-hidden="true" />
        <h3 className="build-plan-section-title">{LLocaleTextGet("ffmpegBuild.stalled.title")}</h3>
      </div>
      {props.addresses.length > 0 && (
        <div className="build-plan-stalled-addresses">
          <span className="build-plan-stalled__label">{LLocaleTextGet("ffmpegBuild.stalled.triedAddresses")}</span>
          <ul className="build-plan-stalled-list">
            {props.addresses.map((address, index) => (
              <li className="build-plan-stalled-list__item" key={`${address}-${index}`}><code>{address}</code></li>
            ))}
          </ul>
        </div>
      )}
      <div className="build-plan-stalled-control">
        <button className="button button--primary" type="button" onClick={props.onRetry}>{LLocaleTextGet("ffmpegBuild.stalled.retry")}</button>
      </div>
    </section>
  );
}

export function PBuildRender({ ffmpegBuildPlanReview, ffmpegLogEntries, approvedActionPhase, approvedActionStatus, ffmpegStalledAddresses, ffmpegProgress, canCancelFfmpeg, isReviewExpired, retryFfmpegBuildPlan, cancelApprovedAction, LActionApprovedClear, onGoToOptions }: LBuildProperties) {
  // A running FFmpeg build takes priority: show live progress, hide the plan.
  const isFfmpegRunning = approvedActionPhase === "ffmpeg";
  const showProgress = isFfmpegRunning || ffmpegLogEntries.length > 0;
  const showApproval = !!ffmpegBuildPlanReview && !isFfmpegRunning;
  return (
    <section className="tab-page ffmpeg-build-page">
      <PHeaderPageRender title={LLocaleTextGet("ffmpegBuild.title")} text={LLocaleTextGet("ffmpegBuild.intro")} />
      {showApproval && <PCardBuildRender review={ffmpegBuildPlanReview} isReviewExpired={isReviewExpired} />}
      {!showApproval && !showProgress && (
        <PCardPlanRender onGoToOptions={onGoToOptions} />
      )}
      {approvedActionStatus === "stalled" && (
        <PBuildStalledRender addresses={ffmpegStalledAddresses} onRetry={retryFfmpegBuildPlan} />
      )}
      {showProgress && (
        <PProgressLiveRender
          isActive={isFfmpegRunning}
          approvedActionStatus={approvedActionStatus}
          progress={ffmpegProgress}
          pipeline={LPipelineFfmpegGet()}
          completionLabel={LLocaleTextGet("ffmpegBuild.progress.completionLabel")}
          onCancel={cancelApprovedAction}
          canCancel={canCancelFfmpeg}
          onClear={LActionApprovedClear}
        />
      )}
    </section>
  );
}
