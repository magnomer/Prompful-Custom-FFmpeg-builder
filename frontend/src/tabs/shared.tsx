import React from "react";
import { t, tFallback, tStatus } from "../i18n";
import type { LiveProgress, LogPhaseId } from "./logs";
import { getPhaseLabel, runtimeLogText } from "./logutils";
import { configureOptionText, libraryLicenseLabel, libraryText } from "../catalogText";


function planWarningText(warning: PlanWarning): string {
  if (warning.messageKey) return t(warning.messageKey, warning.messageValues ?? {});
  return warning.message;
}

function planOperationText(operation: PlanOperation): string {
  if (operation.summaryKey) return t(operation.summaryKey, operation.summaryValues ?? {});
  return operation.summary;
}

function actionNameLabel(actionName: string): string {
  return tFallback(`approval.action.${actionName}`, actionName);
}

export function PageHeader(props: { title: string; text: string }) {
  return (
    <header className="page-header">
      <h1 className="page-header__title">{props.title}</h1>
      <p className="page-header__text">{props.text}</p>
    </header>
  );
}

export function InfoBox(props: { title?: string; children: React.ReactNode }) {
  return (
    <section className="description">
      {props.title && <h2 className="description__title">{props.title}</h2>}
      <div className="description__body">{props.children}</div>
    </section>
  );
}

export function ExternalLinkButton(props: { label: string; url: string; onOpen: (urlToOpen: string) => Promise<void> }) {
  return (
    <button className="button button--link" type="button" onClick={() => props.onOpen(props.url)}>
      {props.label}
    </button>
  );
}

export function EmptyReview(props: { text: string }) {
  return <p className="empty-text">{props.text}</p>;
}

export function ReviewList(props: { title: string; items: string[] }) {
  return (
    <section className="review-list">
      <h3 className="review-list__title">{props.title}</h3>
      <ul className="review-list__items">
        {props.items.map((item) => <li className="review-list__item" key={item}>{item}</li>)}
      </ul>
    </section>
  );
}

export function WarningList(props: { warnings: PlanWarning[] }) {
  return (
    <section className="review-list">
      <h3 className="review-list__title">{t("review.warnings.title")}</h3>
      <ul className="review-list__items">
        {props.warnings.map((warning, index) => <li className={`review-list__item review-list__item--${warning.riskLevelName}`} key={`${warning.riskLevelName}-${index}`}>{planWarningText(warning)}</li>)}
      </ul>
    </section>
  );
}

export type ApprovalPanelProps = {
  title: string;
  actionName: string;
  planHash: string;
  expectedConsentText: string;
  operations: PlanOperation[];
  warnings: PlanWarning[];
  isExecutable: boolean;
  selectedLibraries?: LibraryChoice[];
  requiredMsys2PackageNames?: string[];
  generatedConfigureFlags?: string[];
  selectedConfigureOptions?: ConfigureOptionChoice[];
  generatedOptionFlags?: string[];
  extraConfigureFlags?: string[];
  finalConfigureFlags?: string[];
  onRequestBackendConfirmation: () => void;
};

export function ApprovalPanel(props: ApprovalPanelProps) {
  return (
    <section className="approval-panel">
      <h2 className="approval-panel__title">{props.title}</h2>
      <p className="approval-panel__summary">{t("approval.summary")}</p>
      <dl className="metadata">
        <dt>{t("approval.metadata.action")}</dt><dd>{actionNameLabel(props.actionName)}</dd>
        <dt>{t("approval.metadata.planHash")}</dt><dd className="metadata__hash">{props.planHash}</dd>
        <dt>{t("approval.metadata.backendPhrase")}</dt><dd>{props.expectedConsentText}</dd>
      </dl>
      {props.selectedLibraries && props.selectedLibraries.length > 0 && (
        <ReviewList title={t("approval.review.selectedLibraries")} items={props.selectedLibraries.map((library) => `${libraryText(library, "displayName")} | ${libraryLicenseLabel(library.licenseEffectName)} | ${library.configureFlags.join(" ")}`)} />
      )}
      {props.requiredMsys2PackageNames && props.requiredMsys2PackageNames.length > 0 && <ReviewList title={t("approval.review.requiredLibraryPackages")} items={props.requiredMsys2PackageNames} />}
      {props.generatedConfigureFlags && props.generatedConfigureFlags.length > 0 && <ReviewList title={t("approval.review.generatedLibraryFlags")} items={props.generatedConfigureFlags} />}
      {props.selectedConfigureOptions && props.selectedConfigureOptions.length > 0 && <ReviewList title={t("approval.review.selectedBuiltInOptions")} items={props.selectedConfigureOptions.map((option) => `${configureOptionText(option, "displayName")} | ${option.configureFlags.join(" ")}`)} />}
      {props.generatedOptionFlags && props.generatedOptionFlags.length > 0 && <ReviewList title={t("approval.review.generatedOptionFlags")} items={props.generatedOptionFlags} />}
      {props.extraConfigureFlags && props.extraConfigureFlags.length > 0 && <ReviewList title={t("approval.review.advancedManualFlags")} items={props.extraConfigureFlags} />}
      {props.finalConfigureFlags && props.finalConfigureFlags.length > 0 && <ReviewList title={t("approval.review.finalConfigureFlags")} items={props.finalConfigureFlags} />}
      <ReviewList title={t("approval.review.operations")} items={props.operations.map((operation) => planOperationText(operation))} />
      {props.warnings.length > 0 && <WarningList warnings={props.warnings} />}
    </section>
  );
}

export function LiveBuildProgress(props: {
  isActive: boolean;
  approvedActionStatus: string;
  progress: LiveProgress;
  pipeline: { id: LogPhaseId; label: string; short: string }[];
  completionLabel: string;
  onCancel: () => void;
  canCancel: boolean;
}) {
  const { isActive, approvedActionStatus, progress, pipeline, completionLabel, onCancel, canCancel } = props;

  const currentPhaseId = progress.currentPhaseId;
  const pipelineIds = pipeline.map((s) => s.id);
  const currentPipelineIndex = currentPhaseId ? pipelineIds.indexOf(currentPhaseId) : -1;
  const completedPhaseIds = new Set<LogPhaseId>(
    progress.isComplete ? pipelineIds : pipelineIds.filter((_, idx) => idx < currentPipelineIndex),
  );
  const currentGroup = progress.phaseGroups?.find((g) => g.phase === currentPhaseId);
  const isIdle = !isActive && approvedActionStatus === "idle";
  const isComplete = progress.isComplete;
  const hasFailed = progress.hasFailed && !isComplete;
  const lastMeaningful = progress.lastMessage ?? null;

  return (
    <div className={`live-progress ${isComplete ? "live-progress--done" : ""} ${hasFailed ? "live-progress--failed" : ""} ${isActive && !isComplete && !hasFailed ? "live-progress--running" : ""}`}>
      <div className="live-progress__header">
        <span className="live-progress__title">
          {isIdle && t("progress.waitingToStart")}
          {isActive && !isComplete && !hasFailed && (
            <><span className="live-progress__spinner" aria-hidden="true" /> {progress.currentPhaseId ? getPhaseLabel(progress.currentPhaseId) : tStatus(approvedActionStatus)}</>
          )}
          {isComplete && t("progress.complete", { label: completionLabel })}
          {hasFailed && t("progress.failed", { label: completionLabel })}
        </span>
        {isActive && !isComplete && !isIdle && (
          <button className="button button--danger live-progress__cancel" type="button" disabled={!canCancel} onClick={onCancel}>{t("actions.cancel")}</button>
        )}
      </div>

      {!isIdle && (
        <div className="live-progress__pipeline" aria-label={t("progress.phases.ariaLabel")}>
          {pipeline.map((step) => {
            const isDone = completedPhaseIds.has(step.id);
            const isFailed = step.id === currentPhaseId && hasFailed;
            const isCurrent = step.id === currentPhaseId && !isComplete && !hasFailed;
            const isPending = !isDone && !isCurrent && !isFailed;
            return (
              <div
                key={step.id}
                className={`live-progress__step ${isDone ? "live-progress__step--done" : ""} ${isCurrent ? "live-progress__step--current" : ""} ${isFailed ? "live-progress__step--failed" : ""} ${isPending ? "live-progress__step--pending" : ""}`}
                title={step.label}
              >
                <span className="live-progress__step-mark">{isDone ? t("progress.step.doneMark") : isFailed ? t("progress.step.failedMark") : isCurrent ? t("progress.step.currentMark") : ""}</span>
                <span className="live-progress__step-dot" aria-hidden="true" />
                <span className="live-progress__step-label">{step.short}</span>
              </div>
            );
          })}
        </div>
      )}

      {isActive && !isComplete && !hasFailed && (
        <div className="live-progress__counters">
          {progress.compileCount > 0 && <ProgressCounter value={progress.compileCount} label={t("progress.counters.cFilesCompiled")} />}
          {progress.assembleCount > 0 && <ProgressCounter value={progress.assembleCount} label={t("progress.counters.asmFiles")} />}
          {progress.copiedDllCount > 0 && <ProgressCounter value={progress.copiedDllCount} label={t("progress.counters.dllsBundled")} />}
          {currentGroup && currentGroup.phase === "ff-compile" && currentGroup.entries.length > 0 && (
            <WideProgressCounter label={t("progress.counters.lastCompiled")} value={currentGroup.entries[currentGroup.entries.length - 1].compileTarget?.split("/").pop()} />
          )}
          {currentGroup && currentGroup.phase === "tc-install" && (
            <WideProgressCounter label={t("progress.counters.packagesInstalled")} value={currentGroup.entries.filter((e) => e.message.startsWith("installing ") || e.message.startsWith("reinstalling ")).length} />
          )}
          {currentGroup && currentGroup.phase === "ff-pkgconfig" && currentGroup.entries.filter((e) => e.message.startsWith("reinstalling ") || e.message.startsWith("downgrading ")).length > 0 && (
            <WideProgressCounter label={t("progress.counters.packagesRefreshed")} value={currentGroup.entries.filter((e) => e.message.startsWith("reinstalling ") || e.message.startsWith("downgrading ")).length} />
          )}
        </div>
      )}

      {isActive && !isComplete && lastMeaningful && (
        <div className="live-progress__last-line">
          <span className="live-progress__last-label">{t("progress.lastMessage")}</span>
          <span className="live-progress__last-msg">{runtimeLogText(lastMeaningful)}</span>
        </div>
      )}

      {isComplete && (
        <div className="live-progress__counters">
          {progress.compileCount > 0 && <ProgressCounter value={progress.compileCount} label={t("progress.counters.cFiles")} />}
          {progress.assembleCount > 0 && <ProgressCounter value={progress.assembleCount} label={t("progress.counters.asmFiles")} />}
          {progress.copiedDllCount > 0 && <ProgressCounter value={progress.copiedDllCount} label={t("progress.counters.dllsBundled")} />}
        </div>
      )}
    </div>
  );
}

function ProgressCounter(props: { value: number; label: string }) {
  return (
    <div className="live-progress__counter">
      <span className="live-progress__counter-value">{props.value}</span>
      <span className="live-progress__counter-label">{props.label}</span>
    </div>
  );
}

function WideProgressCounter(props: { label: string; value: React.ReactNode }) {
  return (
    <div className="live-progress__counter live-progress__counter--wide">
      <span className="live-progress__counter-label">{props.label}</span>
      <code className="live-progress__counter-file">{props.value}</code>
    </div>
  );
}
