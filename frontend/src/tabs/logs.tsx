import React, { useMemo, useState } from "react";
import { LLocaleGet, LLocaleTextGet } from "../i18n";
import emptyStatePurpleIcon from "../assets/empty-card-icons/EmptyStatePurple.svg";
import {
  LLogSecurityEntry, LPhaseLogId, LPhaseLogGroup,
  LPhaseToolchainOrder, LPhaseFfmpegOrder,
  LLogEntryParse, LPhaseGroupBuild, LLogRuntimeBuild,
} from "./logutils";
import { LTabKeyDown } from "./tabkeyboard";

export type { LLogSecurityEntry, LPhaseLogGroup };
export type { LLogSecurityPayload, LStatusActionPayload, LProgressLive, LPhaseLogId, LLogParsedEntry } from "./logutils";
export { LProgressGet, LPipelineToolchainGet, LPipelineFfmpegGet } from "./logutils";

function PLogSmartRender(props: { entries: LLogSecurityEntry[]; context?: "toolchain" | "ffmpeg"; viewMode?: "smart" | "raw"; hideModeToolbar?: boolean }) {
  const locale = LLocaleGet();
  const [expandedPhases, setExpandedPhases] = useState<Set<LPhaseLogId>>(new Set());
  const [showSystemDlls, setShowSystemDlls] = useState(false);
  const [internalViewMode, setInternalViewMode] = useState<"smart" | "raw">("smart");
  const viewMode = props.viewMode ?? internalViewMode;

  const parsed = useMemo(() => props.entries.map((e) => LLogEntryParse(e, props.context ?? "ffmpeg")), [props.entries, props.context]);
  const phaseOrder = props.context === "toolchain" ? LPhaseToolchainOrder : props.context === "ffmpeg" ? LPhaseFfmpegOrder : [...LPhaseToolchainOrder, ...LPhaseFfmpegOrder];
  const phaseGroups = useMemo(() => LPhaseGroupBuild(parsed, phaseOrder), [parsed, phaseOrder, locale]);
  const errorEntries = useMemo(() => parsed.filter((e) => e.level === "error"), [parsed]);
  const warnEntries = useMemo(() => parsed.filter((e) => e.level === "warn"), [parsed]);

  const totalCompile = phaseGroups.reduce((sum, g) => sum + g.compileCount, 0);
  const totalAssemble = phaseGroups.reduce((sum, g) => sum + g.assembleCount, 0);
  const totalCopied = phaseGroups.reduce((sum, g) => sum + g.copiedDlls.length, 0);

  function LPhaseToggle(phase: LPhaseLogId) {
    setExpandedPhases((prev) => {
      const next = new Set(prev);
      if (next.has(phase)) next.delete(phase);
      else next.add(phase);
      return next;
    });
  }

  if (props.entries.length === 0) return <p className="empty-text">{LLocaleTextGet("logs.empty")}</p>;

  return (
    <div className="smart-log">
      {!props.hideModeToolbar && (
        <div className="smart-log__toolbar">
          <button className={`smart-log__mode-btn ${viewMode === "smart" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setInternalViewMode("smart")}>{LLocaleTextGet("logs.view.summary")}</button>
          <button className={`smart-log__mode-btn ${viewMode === "raw" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setInternalViewMode("raw")}>{LLocaleTextGet("logs.view.raw", { count: props.entries.length })}</button>
        </div>
      )}

      {viewMode === "raw" && (
        <div className="log-list" aria-live="polite">
          {[...props.entries].reverse().map((entry, index) => (
            <p className={`log-list__entry log-list__entry--${entry.level}`} key={`${entry.level}-${index}-${entry.message}`}>
              <strong>{entry.level}</strong><time className="log-list__time">{entry.timestamp}</time><span>{LLogRuntimeBuild(entry.message)}</span>
            </p>
          ))}
        </div>
      )}

      {viewMode === "smart" && (
        <>
          <div className="smart-log__stats">
            {totalCompile > 0 && <PStatRender value={totalCompile} label={LLocaleTextGet("logs.stats.cFilesCompiled")} />}
            {totalAssemble > 0 && <PStatRender value={totalAssemble} label={LLocaleTextGet("logs.stats.asmFilesAssembled")} />}
            {totalCopied > 0 && <PStatRender value={totalCopied} label={LLocaleTextGet("logs.stats.dllsCopied")} />}
            {errorEntries.length > 0 && <PStatRender value={errorEntries.length} label={LLocaleTextGet("logs.stats.errors")} className="smart-log__stat--error" />}
            {warnEntries.length > 0 && <PStatRender value={warnEntries.length} label={LLocaleTextGet("logs.stats.warnings")} className="smart-log__stat--warn" />}
          </div>
          {errorEntries.length > 0 && (
            <section className="smart-log__surface smart-log__surface--error">
              <h3 className="smart-log__surface-title">{LLocaleTextGet("logs.errors.title")}</h3>
              {errorEntries.map((e, i) => <PLogEntryRender entry={e} key={`err-${i}`} />)}
            </section>
          )}
          {warnEntries.length > 0 && (
            <section className="smart-log__surface smart-log__surface--warn">
              <h3 className="smart-log__surface-title">{LLocaleTextGet("logs.warnings.title", { count: warnEntries.length })}</h3>
              <details className="smart-log__details">
                <summary className="smart-log__details-summary">{LLocaleTextGet("logs.warnings.showLines", { count: warnEntries.length })}</summary>
                {warnEntries.map((e, i) => <PLogEntryRender entry={e} key={`warn-${i}`} />)}
              </details>
            </section>
          )}
          {phaseGroups.map((group) => (
            <section className="smart-log__phase" key={group.phase}>
              <button className="smart-log__phase-header" type="button" onClick={() => LPhaseToggle(group.phase)}>
                <span className="smart-log__phase-label">{group.label}</span>
                <span className="smart-log__phase-meta">
                  {group.phase === "ff-compile" && group.compileCount > 0 && <span className="smart-log__badge">{LLocaleTextGet("logs.badge.cFiles", { count: group.compileCount })}</span>}
                  {group.phase === "ff-compile" && group.assembleCount > 0 && <span className="smart-log__badge">{LLocaleTextGet("logs.badge.asmFiles", { count: group.assembleCount })}</span>}
                  {(group.phase === "tc-install" || group.phase === "ff-pkgconfig") && <span className="smart-log__badge">{LLocaleTextGet("logs.badge.packages", { count: LPackageInstalledCount(group) })}</span>}
                  {group.phase === "ff-extraction" && (
                    <>
                      {group.copiedDlls.length > 0 && <span className="smart-log__badge smart-log__badge--ok">{LLocaleTextGet("logs.badge.copied", { count: group.copiedDlls.length })}</span>}
                      {group.skippedDllCount > 0 && <span className="smart-log__badge">{LLocaleTextGet("logs.badge.alreadyPresent", { count: group.skippedDllCount })}</span>}
                      {group.systemDllCount > 0 && <span className="smart-log__badge smart-log__badge--dim">{LLocaleTextGet("logs.badge.systemDlls", { count: group.systemDllCount })}</span>}
                    </>
                  )}
                  {group.startTime && group.endTime && group.startTime !== group.endTime && <span className="smart-log__time-range">{LLocaleTextGet("logs.timeRange", { start: group.startTime, end: group.endTime })}</span>}
                  {group.startTime && group.startTime === group.endTime && <span className="smart-log__time-range">{group.startTime}</span>}
                </span>
                <span className="smart-log__phase-chevron">{expandedPhases.has(group.phase) ? LLocaleTextGet("logs.phase.collapseMark") : LLocaleTextGet("logs.phase.expandMark")}</span>
              </button>
              {expandedPhases.has(group.phase) && <PPhaseBodyRender group={group} showSystemDlls={showSystemDlls} onToggleSystemDlls={() => setShowSystemDlls((v) => !v)} />}
            </section>
          ))}
        </>
      )}
    </div>
  );
}

function PStatRender(props: { value: number; label: string; className?: string }) {
  return <div className={`smart-log__stat ${props.className ?? ""}`}><span className="smart-log__stat-value">{props.value}</span><span className="smart-log__stat-label">{props.label}</span></div>;
}

function PLogEntryRender(props: { entry: LLogSecurityEntry }) {
  return <p className="smart-log__surface-entry"><time className="log-list__time">{props.entry.timestamp}</time><span>{LLogRuntimeBuild(props.entry.message)}</span></p>;
}

function PLogRawRender(props: { entry: LLogSecurityEntry; id: string }) {
  return <p className={`log-list__entry log-list__entry--${props.entry.level}`} key={props.id}><strong>{props.entry.level}</strong><time className="log-list__time">{props.entry.timestamp}</time><span>{LLogRuntimeBuild(props.entry.message)}</span></p>;
}

function LPackageInstalledCount(group: LPhaseLogGroup) {
  return group.entries.filter((e) => e.message.startsWith("installing ") || e.message.startsWith("reinstalling ")).length;
}

function PPhaseBodyRender({ group, showSystemDlls, onToggleSystemDlls }: { group: LPhaseLogGroup; showSystemDlls: boolean; onToggleSystemDlls: () => void }) {
  return (
    <div className="smart-log__phase-body">
      {group.phase === "ff-compile" && (
        <>
          {group.compileCount + group.assembleCount > 0 && <div className="smart-log__compile-summary">{LLogSummaryCreate(group.compileCount, group.assembleCount)}</div>}
          <div className="smart-log__compile-list">
            {group.entries.filter((e) => e.compileOp === "CC" || e.compileOp === "CXX" || e.compileOp === "HOSTCC").map((e, i) => (
              <span className="smart-log__compile-file" key={`cc-${i}`} title={e.message}>{e.compileTarget?.split("/").pop()}</span>
            ))}
          </div>
          {group.assembleCount > 0 && (
            <div className="smart-log__compile-list">
              {group.entries.filter((e) => e.compileOp === "X86ASM").map((e, i) => (
                <span className="smart-log__compile-file smart-log__compile-file--asm" key={`asm-${i}`} title={e.message}>{e.compileTarget?.split("/").pop()}</span>
              ))}
            </div>
          )}
        </>
      )}
      {group.phase === "ff-extraction" && (
        <>
          {group.copiedDlls.length > 0 && <div className="smart-log__dll-section"><strong className="smart-log__dll-heading">{LLocaleTextGet("logs.dll.copiedHeading")}</strong><div className="smart-log__dll-list">{group.copiedDlls.map((dll, i) => <span className="smart-log__dll-tag smart-log__dll-tag--copied" key={`dll-${i}`}>{dll}</span>)}</div></div>}
          {group.skippedDllCount > 0 && <p className="smart-log__dll-note">{LLocaleTextGet(group.skippedDllCount === 1 ? "logs.dll.skipped.one" : "logs.dll.skipped.other", { count: group.skippedDllCount })}</p>}
          {group.systemDllCount > 0 && (
            <div className="smart-log__dll-section">
              <button className="smart-log__toggle-btn" type="button" onClick={onToggleSystemDlls}>{LLocaleTextGet(showSystemDlls ? "logs.dll.hideSystem" : "logs.dll.showSystem", { count: group.systemDllCount })}</button>
              {showSystemDlls && <div className="smart-log__dll-list">{group.entries.filter((e) => e.dllAction === "system").map((e, i) => <span className="smart-log__dll-tag smart-log__dll-tag--system" key={`sys-${i}`}>{e.dllDep}</span>)}</div>}
            </div>
          )}
        </>
      )}
      {group.phase === "tc-install" && (
        <>
          <div className="smart-log__pkg-list">{group.entries.filter((e) => e.message.startsWith("reinstalling ")).map((e, i) => <span className="smart-log__pkg-tag" key={`pkg-${i}`}>{e.message.replace("reinstalling ", "").trim()}</span>)}</div>
          {group.entries.map((e, i) => <PLogRawRender entry={e} id={`pkg-raw-${i}`} key={`pkg-raw-${i}`} />)}
        </>
      )}
      {group.phase === "ff-configure" && (
        <>
          {group.entries.filter(LLogConfigureCheck).map((e, i) => <PLogRawRender entry={e} id={`cfg-key-${i}`} key={`cfg-key-${i}`} />)}
          <details className="smart-log__details"><summary className="smart-log__details-summary">{LLocaleTextGet("logs.configure.showAll", { count: group.entries.length })}</summary>{group.entries.map((e, i) => <PLogRawRender entry={e} id={`cfg-raw-${i}`} key={`cfg-raw-${i}`} />)}</details>
        </>
      )}
      {group.phase !== "ff-compile" && group.phase !== "ff-extraction" && group.phase !== "tc-install" && group.phase !== "ff-configure" && (
        group.entries.map((e, i) => <PLogRawRender entry={e} id={`${group.phase}-${i}`} key={`${group.phase}-${i}`} />)
      )}
    </div>
  );
}

function LLogSummaryCreate(compileCount: number, assembleCount: number): string {
  if (assembleCount > 0) {
    const compileKey = compileCount === 1 ? "logs.compile.unit.cFile.one" : "logs.compile.unit.cFile.other";
    const assembleKey = assembleCount === 1 ? "logs.compile.unit.asmFile.one" : "logs.compile.unit.asmFile.other";
    return LLocaleTextGet("logs.compile.summaryWithAsm", {
      compileCount,
      compileUnit: LLocaleTextGet(compileKey),
      assembleCount,
      assembleUnit: LLocaleTextGet(assembleKey),
    });
  }
  const compileKey = compileCount === 1 ? "logs.compile.unit.cFile.one" : "logs.compile.unit.cFile.other";
  return LLocaleTextGet("logs.compile.summary", { compileCount, compileUnit: LLocaleTextGet(compileKey) });
}

function LLogConfigureCheck(e: LLogSecurityEntry): boolean {
  return e.message.startsWith("FFmpeg configure") || e.message.startsWith("Starting FFmpeg configure") || e.message.startsWith("License:") || e.message.startsWith("C compiler") || e.message.startsWith("C library") || e.message.startsWith("ARCH ") || e.message.startsWith("threading") || e.message.startsWith("static ") || e.message.startsWith("shared ") || e.message.startsWith("x86 assembler") || e.message.startsWith("Running approved");
}

function PLogEmptyRender(props: { onGoToPrep: () => void; onGoToBuild: () => void }) {
  return (
    <section className="card card--purple logs-empty-card">
      <span className="card__badge" aria-hidden="true">
        <img className="card__badge-icon" src={emptyStatePurpleIcon} alt="" />
      </span>
      <div className="card__head logs-empty-card__head">
        <h2 className="card__title">{LLocaleTextGet("logs.empty")}</h2>
        <p className="card__desc">{LLocaleTextGet("logs.empty.description")}</p>
      </div>
      <div className="logs-empty-card__actions">
        <button className="button" type="button" onClick={props.onGoToBuild}>{LLocaleTextGet("logs.actions.goToBuild")}</button>
        <button className="button button--primary" type="button" onClick={props.onGoToPrep}>{LLocaleTextGet("logs.actions.goToPrep")}</button>
      </div>
    </section>
  );
}

type LLogViewIdentifier = "summary" | "raw";
type LLogKindNormalized = "toolchain" | "ffmpeg" | "unknown";

function LRecordKindNormalize(kind: string): LLogKindNormalized {
  if (kind === "toolchain") return "toolchain";
  if (kind === "ffmpeg") return "ffmpeg";
  return "unknown";
}

function LLogKindGet(kind: string): string {
  if (kind === "toolchain") return LLocaleTextGet("logs.local.kind.toolchain");
  if (kind === "ffmpeg") return LLocaleTextGet("logs.local.kind.ffmpeg");
  return LLocaleTextGet("logs.local.kind.unknown");
}

function LLogStatusGet(status: string): string {
  return LLocaleTextGet(`logs.local.status.${status}`);
}

function LRecordLiveCreate(toolchainLogEntries: LLogSecurityEntry[], ffmpegLogEntries: LLogSecurityEntry[]): LRecordLog[] {
  const records: LRecordLog[] = [];
  if (ffmpegLogEntries.length > 0) {
    records.push({
      runId: "live-ffmpeg",
      createdAt: "",
      displayTime: LLocaleTextGet("logs.local.live"),
      kind: "ffmpeg",
      status: "running",
      directory: "",
      entries: ffmpegLogEntries,
      rawText: ffmpegLogEntries.map((entry) => `[${entry.timestamp}] ${entry.level}: ${entry.message}`).join("\n"),
      errorCount: ffmpegLogEntries.filter((entry) => entry.level === "error").length,
      warnCount: ffmpegLogEntries.filter((entry) => entry.level === "warn").length,
      hasStdoutLog: false,
      hasStderrLog: false,
      hasSecurityLAuditEvents: false,
    });
  }
  if (toolchainLogEntries.length > 0) {
    records.push({
      runId: "live-toolchain",
      createdAt: "",
      displayTime: LLocaleTextGet("logs.local.live"),
      kind: "toolchain",
      status: "running",
      directory: "",
      entries: toolchainLogEntries,
      rawText: toolchainLogEntries.map((entry) => `[${entry.timestamp}] ${entry.level}: ${entry.message}`).join("\n"),
      errorCount: toolchainLogEntries.filter((entry) => entry.level === "error").length,
      warnCount: toolchainLogEntries.filter((entry) => entry.level === "warn").length,
      hasStdoutLog: false,
      hasStderrLog: false,
      hasSecurityLAuditEvents: false,
    });
  }
  return records;
}

function PLogSelectorRender(props: { records: LRecordLog[]; selectedRunId: string; onSelect: (runId: string) => void; onOpenLogsFolder: () => Promise<void>; onOpenRecordFolder: (runId: string) => Promise<void> }) {
  const selectedRecord = props.records.find((record) => record.runId === props.selectedRunId);
  const canOpenRecordFolder = Boolean(selectedRecord && !selectedRecord.runId.startsWith("live-") && selectedRecord.directory);

  return (
    <section className="log-record-selector" aria-label={LLocaleTextGet("logs.local.selector.ariaLabel")}>
      <label className="log-record-selector__label" htmlFor="local-log-record-select">{LLocaleTextGet("logs.local.selector.label")}</label>
      <select
        id="local-log-record-select"
        className="log-record-selector__select"
        value={props.selectedRunId}
        onChange={(event) => props.onSelect(event.target.value)}
      >
        {props.records.map((record) => (
          <option value={record.runId} key={record.runId}>
            {record.displayTime} · {LLogKindGet(record.kind)} · {LLogStatusGet(record.status)}
          </option>
        ))}
      </select>
      <button
        className="button log-record-selector__open-button"
        type="button"
        disabled={!canOpenRecordFolder}
        onClick={() => selectedRecord && props.onOpenRecordFolder(selectedRecord.runId)}
      >
        {LLocaleTextGet("logs.local.actions.openRecordFolder")}
      </button>
      <button
        className="button log-record-selector__open-button"
        type="button"
        onClick={props.onOpenLogsFolder}
      >
        {LLocaleTextGet("logs.local.actions.openLogsFolder")}
      </button>
    </section>
  );
}

function PLogViewerRender(props: { text: string }) {
  if (!props.text.trim()) return <p className="empty-text">{LLocaleTextGet("logs.local.empty.raw")}</p>;
  return <pre className="log-raw-file-viewer">{props.text}</pre>;
}

function PLogFilesRender(props: { record: LRecordLog; onOpenFile: (runId: string, fileName: string) => Promise<void> }) {
  const files: { fileName: string; key: string; available: boolean }[] = [
    { fileName: "stdout.log", key: "stdout", available: props.record.hasStdoutLog },
    { fileName: "stderr.log", key: "stderr", available: props.record.hasStderrLog },
    { fileName: "security-events.jsonl", key: "events", available: props.record.hasSecurityLAuditEvents },
  ];
  if (!files.some((file) => file.available)) return <p className="empty-text">{LLocaleTextGet("logs.local.empty.raw")}</p>;
  return (
    <div className="log-raw-file-sections">
      <p className="log-raw-file-sections__intro">{LLocaleTextGet("logs.local.raw.intro")}</p>
      {files.map((file) => (
        <section className="log-raw-file-card" key={file.fileName}>
          <div className="log-raw-file-card__head">
            <h3 className="log-raw-file-card__title">{LLocaleTextGet(`logs.local.raw.${file.key}.title`)}</h3>
            <code className="log-raw-file-card__name">{file.fileName}</code>
          </div>
          <p className="log-raw-file-card__simple">{LLocaleTextGet(`logs.local.raw.${file.key}.simple`)}</p>
          <p className="log-raw-file-card__tech">{LLocaleTextGet(`logs.local.raw.${file.key}.tech`)}</p>
          <button
            className="button log-raw-file-card__open"
            type="button"
            disabled={!file.available}
            onClick={() => props.onOpenFile(props.record.runId, file.fileName)}
          >
            {file.available ? LLocaleTextGet("logs.local.raw.open") : LLocaleTextGet("logs.local.raw.missing")}
          </button>
        </section>
      ))}
    </div>
  );
}

function PLogDetailsRender(props: { record: LRecordLog; onOpenRecordFile: (runId: string, fileName: string) => Promise<void> }) {
  const [activeTabId, setActiveTabId] = useState<LLogViewIdentifier>("summary");
  const kind = LRecordKindNormalize(props.record.kind);
  const context = kind === "toolchain" ? "toolchain" : kind === "ffmpeg" ? "ffmpeg" : null;
  const isLive = props.record.runId.startsWith("live-");
  const hasDetails = isLive || props.record.entries.length > 0 || props.record.rawText.trim().length > 0;
  const rawCount = isLive
    ? (props.record.rawText.trim() ? 1 : 0)
    : [props.record.hasStdoutLog, props.record.hasStderrLog, props.record.hasSecurityLAuditEvents].filter(Boolean).length;
  const tabs: { id: LLogViewIdentifier; label: string; count: number }[] = kind === "unknown"
    ? [{ id: "raw", label: LLocaleTextGet("logs.local.tabs.unknownRaw"), count: rawCount }]
    : [
      { id: "summary", label: kind === "toolchain" ? LLocaleTextGet("logs.local.tabs.environmentSummary") : LLocaleTextGet("logs.local.tabs.ffmpegSummary"), count: props.record.entries.length },
      { id: "raw", label: kind === "toolchain" ? LLocaleTextGet("logs.local.tabs.environmentRaw") : LLocaleTextGet("logs.local.tabs.ffmpegRaw"), count: rawCount },
    ];

  const effectiveActiveTabId = tabs.some((tab) => tab.id === activeTabId) ? activeTabId : tabs[0]?.id ?? "raw";

  return (
    <section className="result-details-card log-details-card">
      <div className="result-details-tabs" role="tablist" aria-label={LLocaleTextGet("logs.local.tabs.ariaLabel")}>
        {tabs.map((tab, index) => (
          <button
            className={`result-details-tab ${effectiveActiveTabId === tab.id ? "result-details-tab--active" : ""}`}
            type="button"
            role="tab"
            aria-selected={effectiveActiveTabId === tab.id}
            aria-controls="log-detail-tabpanel"
            id={`log-detail-tab-${tab.id}`}
            tabIndex={effectiveActiveTabId === tab.id ? 0 : -1}
            onClick={() => setActiveTabId(tab.id)}
            onKeyDown={(event) => LTabKeyDown(event, index, tabs.length, (nextIndex) => setActiveTabId(tabs[nextIndex].id))}
            key={tab.id}
          >
            <span>{tab.label}</span>
            <span className="result-details-tab__count">{tab.count}</span>
          </button>
        ))}
      </div>
      <div className="result-details-body log-details-body" role="tabpanel" id="log-detail-tabpanel" aria-labelledby={`log-detail-tab-${effectiveActiveTabId}`}>
        <div className="log-record-meta">
          <span>{props.record.displayTime}</span>
          <span>{LLogKindGet(props.record.kind)}</span>
          <span>{LLogStatusGet(props.record.status)}</span>
          {props.record.errorCount > 0 && <span>{LLocaleTextGet("logs.local.meta.errors", { count: props.record.errorCount })}</span>}
          {props.record.warnCount > 0 && <span>{LLocaleTextGet("logs.local.meta.warnings", { count: props.record.warnCount })}</span>}
        </div>
        {(() => {
          const showRaw = effectiveActiveTabId === "raw" || !context;
          if (showRaw && !isLive) {
            return <PLogFilesRender record={props.record} onOpenFile={props.onOpenRecordFile} />;
          }
          if (!hasDetails) {
            return <p className="empty-text">{LLocaleTextGet("logs.local.loading")}</p>;
          }
          if (showRaw) {
            return <PLogViewerRender text={props.record.rawText} />;
          }
          if (context) {
            return <PLogSmartRender entries={props.record.entries} context={context} viewMode="smart" hideModeToolbar />;
          }
          return null;
        })()}
      </div>
    </section>
  );
}

export type LLogProperties = {
  toolchainLogEntries: LLogSecurityEntry[];
  ffmpegLogEntries: LLogSecurityEntry[];
  localLogRecords: LRecordLog[];
  localLogRecordsError: string;
  refreshLocalLogRecords: () => Promise<void>;
  loadLocalLogRecord: (runId: string) => Promise<void>;
  openLocalLogsFolder: () => Promise<void>;
  openLocalLogRecordFolder: (runId: string) => Promise<void>;
  openLocalLogRecordFile: (runId: string, fileName: string) => Promise<void>;
  onGoToPrep: () => void;
  onGoToBuild: () => void;
};

export function PLogRender({ toolchainLogEntries, ffmpegLogEntries, localLogRecords, localLogRecordsError, refreshLocalLogRecords, loadLocalLogRecord, openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile, onGoToPrep, onGoToBuild }: LLogProperties) {
  const liveRecords = useMemo(() => LRecordLiveCreate(toolchainLogEntries, ffmpegLogEntries), [toolchainLogEntries, ffmpegLogEntries]);
  const records = useMemo(() => [...liveRecords, ...localLogRecords], [liveRecords, localLogRecords]);
  const [selectedRunId, setSelectedRunId] = useState("");
  const requestedDetailRunIds = React.useRef(new Set<string>());

  React.useEffect(() => {
    if (records.length === 0) {
      setSelectedRunId("");
      return;
    }
    if (!records.some((record) => record.runId === selectedRunId)) {
      setSelectedRunId(records[0].runId);
    }
  }, [records, selectedRunId]);

  const selectedRecord = records.find((record) => record.runId === selectedRunId) ?? records[0];

  React.useEffect(() => {
    if (!selectedRecord || selectedRecord.runId.startsWith("live-")) return;
    if (selectedRecord.entries.length > 0 || selectedRecord.rawText.trim()) return;
    if (requestedDetailRunIds.current.has(selectedRecord.runId)) return;
    const runId = selectedRecord.runId;
    requestedDetailRunIds.current.add(runId);
    void loadLocalLogRecord(runId).finally(() => requestedDetailRunIds.current.delete(runId));
  }, [selectedRecord?.runId, selectedRecord?.entries.length, selectedRecord?.rawText]);

  return (
    <section className="tab-page log-page">
      <header className="result-page__header">
        <div>
          <h1 className="page-header__title">{LLocaleTextGet("logs.title")}</h1>
          <p className="page-header__text">{LLocaleTextGet("logs.intro")}</p>
        </div>
        <div className="result-page__actions">
          <button className="button" type="button" onClick={refreshLocalLogRecords}>{LLocaleTextGet("logs.actions.refresh")}</button>
        </div>
      </header>
      {localLogRecordsError && <p className="empty-text">{localLogRecordsError}</p>}
      {!localLogRecordsError && records.length > 0 && selectedRecord && (
        <>
          <PLogSelectorRender records={records} selectedRunId={selectedRecord.runId} onSelect={setSelectedRunId} onOpenLogsFolder={openLocalLogsFolder} onOpenRecordFolder={openLocalLogRecordFolder} />
          <PLogDetailsRender record={selectedRecord} onOpenRecordFile={openLocalLogRecordFile} />
        </>
      )}
      {!localLogRecordsError && records.length === 0 && <PLogEmptyRender onGoToPrep={onGoToPrep} onGoToBuild={onGoToBuild} />}
    </section>
  );
}
