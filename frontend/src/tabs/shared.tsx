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

export function DescriptionLines(props: { text: string; className?: string; groupSentences?: boolean }) {
  const sentences = props.text.split(/(?<=[.!?。])\s+/).filter((sentence) => sentence.trim().length > 0);
  const isSubordinate = (sentence: string) => sentence.trim().startsWith("(");

  let lines: { text: string; sub: boolean }[];
  if (props.groupSentences) {
    // Keep regular sentences flowing together; only break parentheticals onto their own line.
    lines = [];
    let buffer: string[] = [];
    for (const sentence of sentences) {
      if (isSubordinate(sentence)) {
        if (buffer.length > 0) { lines.push({ text: buffer.join(" "), sub: false }); buffer = []; }
        lines.push({ text: sentence, sub: true });
      } else {
        buffer.push(sentence);
      }
    }
    if (buffer.length > 0) lines.push({ text: buffer.join(" "), sub: false });
  } else {
    // One sentence per line.
    lines = sentences.map((sentence) => ({ text: sentence, sub: isSubordinate(sentence) }));
  }

  return (
    <p className={props.className ?? "card__desc"}>
      {lines.map((line, index) => (
        <span className={`card__desc-line ${line.sub ? "card__desc-line--sub" : ""}`} key={index}>{line.text}</span>
      ))}
    </p>
  );
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

function CopyableField(props: { value: string; ariaLabel: string }) {
  const [copied, setCopied] = React.useState(false);
  async function copyToClipboard() {
    try {
      await navigator.clipboard.writeText(props.value);
      setCopied(true);
      window.setTimeout(() => setCopied(false), 1500);
    } catch {
      // Clipboard may be unavailable; the field is still selectable for manual copy.
    }
  }
  return (
    <div className="copy-field">
      <textarea
        className="card__input metadata__field"
        readOnly
        rows={2}
        value={props.value}
        onFocus={(event) => event.currentTarget.select()}
        aria-label={props.ariaLabel}
      />
      <button className="button copy-field__btn" type="button" onClick={copyToClipboard}>
        {copied ? t("actions.copied") : t("actions.copy")}
      </button>
    </div>
  );
}

export function ReviewList(props: { title: string; items: string[]; dense?: boolean }) {
  return (
    <section className={`review-list ${props.dense ? "review-list--dense" : ""}`}>
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
  variant?: string;
  icon?: string;
  selectedLibraries?: LibraryChoice[];
  libraryPreparations?: LibraryPreparation[];
  requiredMsys2PackageNames?: string[];
  generatedConfigureFlags?: string[];
  selectedConfigureOptions?: ConfigureOptionChoice[];
  generatedOptionFlags?: string[];
  extraConfigureFlags?: string[];
  finalConfigureFlags?: string[];
  onCancelPlan?: () => void;
  onRequestBackendConfirmation: () => void;
  isConfirmationBusy?: boolean;
};

function ApprovalPanelBody(props: ApprovalPanelProps) {
  return (
    <>
      <dl className="metadata">
        <dt>{t("approval.metadata.action")}</dt><dd>{actionNameLabel(props.actionName)}</dd>
        <dt>{t("approval.metadata.planHash")}</dt>
        <dd><CopyableField value={props.planHash} ariaLabel={t("approval.metadata.planHash")} /></dd>
        <dt>{t("approval.metadata.backendPhrase")}</dt>
        <dd><CopyableField value={props.expectedConsentText} ariaLabel={t("approval.metadata.backendPhrase")} /></dd>
      </dl>
      {props.selectedLibraries && props.selectedLibraries.length > 0 && (
        <ReviewList title={t("approval.review.selectedLibraries")} items={props.selectedLibraries.map((library) => `${libraryText(library, "displayName")} | ${libraryLicenseLabel(library.licenseEffectName)} | ${library.configureFlags.join(" ")}`)} />
      )}
      {props.libraryPreparations && props.libraryPreparations.length > 0 && (
        <ReviewList title={t("approval.review.libraryPreparations")} items={props.libraryPreparations.map((preparation) => {
          const name = preparation.version ? `${preparation.displayName} ${preparation.version}` : preparation.displayName;
          const base = `${name} | ${t(`approval.preparation.method.${preparation.method}`)} | ${preparation.allowedDownloadHost}`;
          return preparation.buildDependencyPackages && preparation.buildDependencyPackages.length > 0
            ? `${base} | ${t("approval.preparation.buildDependencies")}: ${preparation.buildDependencyPackages.join(" ")}`
            : base;
        })} />
      )}
      {props.requiredMsys2PackageNames && props.requiredMsys2PackageNames.length > 0 && <ReviewList title={t("approval.review.requiredLibraryPackages")} items={props.requiredMsys2PackageNames} />}
      {props.generatedConfigureFlags && props.generatedConfigureFlags.length > 0 && <ReviewList title={t("approval.review.generatedLibraryFlags")} items={props.generatedConfigureFlags} />}
      {props.selectedConfigureOptions && props.selectedConfigureOptions.length > 0 && <ReviewList title={t("approval.review.selectedBuiltInOptions")} items={props.selectedConfigureOptions.map((option) => `${configureOptionText(option, "displayName")} | ${option.configureFlags.join(" ")}`)} />}
      {props.generatedOptionFlags && props.generatedOptionFlags.length > 0 && <ReviewList title={t("approval.review.generatedOptionFlags")} items={props.generatedOptionFlags} />}
      {props.extraConfigureFlags && props.extraConfigureFlags.length > 0 && <ReviewList title={t("approval.review.advancedManualFlags")} items={props.extraConfigureFlags} />}
      {props.finalConfigureFlags && props.finalConfigureFlags.length > 0 && <ReviewList title={t("approval.review.finalConfigureFlags")} items={props.finalConfigureFlags} />}
      <ReviewList title={t("approval.review.operations")} items={props.operations.map((operation) => planOperationText(operation))} dense />
      {props.warnings.length > 0 && <WarningList warnings={props.warnings} />}
    </>
  );
}

function ApprovalPanelActions(props: ApprovalPanelProps) {
  return (
    <div className="approval-card__actions">
      {props.onCancelPlan && <button className="button" type="button" onClick={props.onCancelPlan}>{t("actions.cancel")}</button>}
      <button className="button button--primary" type="button" disabled={!props.isExecutable || props.isConfirmationBusy} onClick={props.onRequestBackendConfirmation}>{t("actions.requestBackendConfirmation")}</button>
    </div>
  );
}

export function ApprovalPanel(props: ApprovalPanelProps) {
  // Card variant borrows the Source tab card design (colored accent + badge +
  // card head). The plain variant keeps the original review-panel look.
  if (props.variant) {
    return (
      <section className={`card card--${props.variant} approval-card`}>
        <span className="card__badge" aria-hidden="true">{props.icon && <img className="card__badge-icon" src={props.icon} alt="" />}</span>
        <div className="card__head">
          <h2 className="card__title">{props.title}</h2>
          <DescriptionLines text={t("approval.summary")} />
        </div>
        <div className="approval-card__body">
          <ApprovalPanelBody {...props} />
        </div>
        <ApprovalPanelActions {...props} />
      </section>
    );
  }
  return (
    <section className="approval-panel">
      <h2 className="approval-panel__title">{props.title}</h2>
      <p className="approval-panel__summary">{t("approval.summary")}</p>
      <ApprovalPanelBody {...props} />
      <ApprovalPanelActions {...props} />
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

      {hasFailed && progress.failureMessages.length > 0 && (
        <div className="live-progress__failure">
          <span className="live-progress__failure-label">{t("progress.failureReason")}</span>
          {progress.failureMessages.map((failureMessage, index) => (
            <span className="live-progress__failure-msg" key={index}>{runtimeLogText(failureMessage)}</span>
          ))}
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
