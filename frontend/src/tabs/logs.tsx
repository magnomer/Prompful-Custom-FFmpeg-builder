import React, { useMemo, useState } from "react";
import { t } from "../i18n";
import emptyStatePurpleIcon from "../assets/empty-card-icons/EmptyStatePurple.svg";
import {
  SecurityLogEntry, LogPhaseId, LogPhaseGroup,
  TOOLCHAIN_PHASE_ORDER, FFMPEG_PHASE_ORDER,
  parseLogEntry, buildPhaseGroups, runtimeLogText,
} from "./logutils";

export type { SecurityLogEntry, LogPhaseGroup };
export type { SecurityLogPayload, ApprovedActionStatusPayload, LiveProgress, LogPhaseId, ParsedLogEntry } from "./logutils";
export { computeProgress, getToolchainPipeline, getFfmpegPipeline } from "./logutils";

function SmartLogViewer(props: { entries: SecurityLogEntry[]; context?: "toolchain" | "ffmpeg"; viewMode?: "smart" | "raw"; hideModeToolbar?: boolean }) {
  const [expandedPhases, setExpandedPhases] = useState<Set<LogPhaseId>>(new Set());
  const [showSystemDlls, setShowSystemDlls] = useState(false);
  const [internalViewMode, setInternalViewMode] = useState<"smart" | "raw">("smart");
  const viewMode = props.viewMode ?? internalViewMode;

  const parsed = useMemo(() => props.entries.map((e) => parseLogEntry(e, props.context ?? "ffmpeg")), [props.entries, props.context]);
  const phaseOrder = props.context === "toolchain" ? TOOLCHAIN_PHASE_ORDER : props.context === "ffmpeg" ? FFMPEG_PHASE_ORDER : [...TOOLCHAIN_PHASE_ORDER, ...FFMPEG_PHASE_ORDER];
  const phaseGroups = useMemo(() => buildPhaseGroups(parsed, phaseOrder), [parsed, phaseOrder]);
  const errorEntries = useMemo(() => parsed.filter((e) => e.level === "error"), [parsed]);
  const warnEntries = useMemo(() => parsed.filter((e) => e.level === "warn"), [parsed]);

  const totalCompile = phaseGroups.reduce((sum, g) => sum + g.compileCount, 0);
  const totalAssemble = phaseGroups.reduce((sum, g) => sum + g.assembleCount, 0);
  const totalCopied = phaseGroups.reduce((sum, g) => sum + g.copiedDlls.length, 0);

  function togglePhase(phase: LogPhaseId) {
    setExpandedPhases((prev) => {
      const next = new Set(prev);
      if (next.has(phase)) next.delete(phase);
      else next.add(phase);
      return next;
    });
  }

  if (props.entries.length === 0) return <p className="empty-text">{t("logs.empty")}</p>;

  return (
    <div className="smart-log">
      {!props.hideModeToolbar && (
        <div className="smart-log__toolbar">
          <button className={`smart-log__mode-btn ${viewMode === "smart" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setInternalViewMode("smart")}>{t("logs.view.summary")}</button>
          <button className={`smart-log__mode-btn ${viewMode === "raw" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setInternalViewMode("raw")}>{t("logs.view.raw", { count: props.entries.length })}</button>
        </div>
      )}

      {viewMode === "raw" && (
        <div className="log-list" aria-live="polite">
          {[...props.entries].reverse().map((entry, index) => (
            <p className={`log-list__entry log-list__entry--${entry.level}`} key={`${entry.level}-${index}-${entry.message}`}>
              <strong>{entry.level}</strong><time className="log-list__time">{entry.timestamp}</time><span>{runtimeLogText(entry.message)}</span>
            </p>
          ))}
        </div>
      )}

      {viewMode === "smart" && (
        <>
          <div className="smart-log__stats">
            {totalCompile > 0 && <Stat value={totalCompile} label={t("logs.stats.cFilesCompiled")} />}
            {totalAssemble > 0 && <Stat value={totalAssemble} label={t("logs.stats.asmFilesAssembled")} />}
            {totalCopied > 0 && <Stat value={totalCopied} label={t("logs.stats.dllsCopied")} />}
            {errorEntries.length > 0 && <Stat value={errorEntries.length} label={t("logs.stats.errors")} className="smart-log__stat--error" />}
            {warnEntries.length > 0 && <Stat value={warnEntries.length} label={t("logs.stats.warnings")} className="smart-log__stat--warn" />}
          </div>
          {errorEntries.length > 0 && (
            <section className="smart-log__surface smart-log__surface--error">
              <h3 className="smart-log__surface-title">{t("logs.errors.title")}</h3>
              {errorEntries.map((e, i) => <LogSurfaceEntry entry={e} key={`err-${i}`} />)}
            </section>
          )}
          {warnEntries.length > 0 && (
            <section className="smart-log__surface smart-log__surface--warn">
              <h3 className="smart-log__surface-title">{t("logs.warnings.title", { count: warnEntries.length })}</h3>
              <details className="smart-log__details">
                <summary className="smart-log__details-summary">{t("logs.warnings.showLines", { count: warnEntries.length })}</summary>
                {warnEntries.map((e, i) => <LogSurfaceEntry entry={e} key={`warn-${i}`} />)}
              </details>
            </section>
          )}
          {phaseGroups.map((group) => (
            <section className="smart-log__phase" key={group.phase}>
              <button className="smart-log__phase-header" type="button" onClick={() => togglePhase(group.phase)}>
                <span className="smart-log__phase-label">{group.label}</span>
                <span className="smart-log__phase-meta">
                  {group.phase === "ff-compile" && group.compileCount > 0 && <span className="smart-log__badge">{t("logs.badge.cFiles", { count: group.compileCount })}</span>}
                  {group.phase === "ff-compile" && group.assembleCount > 0 && <span className="smart-log__badge">{t("logs.badge.asmFiles", { count: group.assembleCount })}</span>}
                  {(group.phase === "tc-install" || group.phase === "ff-pkgconfig") && <span className="smart-log__badge">{t("logs.badge.packages", { count: countInstalledPackages(group) })}</span>}
                  {group.phase === "ff-extraction" && (
                    <>
                      {group.copiedDlls.length > 0 && <span className="smart-log__badge smart-log__badge--ok">{t("logs.badge.copied", { count: group.copiedDlls.length })}</span>}
                      {group.skippedDllCount > 0 && <span className="smart-log__badge">{t("logs.badge.alreadyPresent", { count: group.skippedDllCount })}</span>}
                      {group.systemDllCount > 0 && <span className="smart-log__badge smart-log__badge--dim">{t("logs.badge.systemDlls", { count: group.systemDllCount })}</span>}
                    </>
                  )}
                  {group.startTime && group.endTime && group.startTime !== group.endTime && <span className="smart-log__time-range">{t("logs.timeRange", { start: group.startTime, end: group.endTime })}</span>}
                  {group.startTime && group.startTime === group.endTime && <span className="smart-log__time-range">{group.startTime}</span>}
                </span>
                <span className="smart-log__phase-chevron">{expandedPhases.has(group.phase) ? t("logs.phase.collapseMark") : t("logs.phase.expandMark")}</span>
              </button>
              {expandedPhases.has(group.phase) && <PhaseBody group={group} showSystemDlls={showSystemDlls} onToggleSystemDlls={() => setShowSystemDlls((v) => !v)} />}
            </section>
          ))}
        </>
      )}
    </div>
  );
}

function Stat(props: { value: number; label: string; className?: string }) {
  return <div className={`smart-log__stat ${props.className ?? ""}`}><span className="smart-log__stat-value">{props.value}</span><span className="smart-log__stat-label">{props.label}</span></div>;
}

function LogSurfaceEntry(props: { entry: SecurityLogEntry }) {
  return <p className="smart-log__surface-entry"><time className="log-list__time">{props.entry.timestamp}</time><span>{runtimeLogText(props.entry.message)}</span></p>;
}

function RawLogEntry(props: { entry: SecurityLogEntry; id: string }) {
  return <p className={`log-list__entry log-list__entry--${props.entry.level}`} key={props.id}><strong>{props.entry.level}</strong><time className="log-list__time">{props.entry.timestamp}</time><span>{runtimeLogText(props.entry.message)}</span></p>;
}

function countInstalledPackages(group: LogPhaseGroup) {
  return group.entries.filter((e) => e.message.startsWith("installing ") || e.message.startsWith("reinstalling ")).length;
}

function PhaseBody({ group, showSystemDlls, onToggleSystemDlls }: { group: LogPhaseGroup; showSystemDlls: boolean; onToggleSystemDlls: () => void }) {
  return (
    <div className="smart-log__phase-body">
      {group.phase === "ff-compile" && (
        <>
          {group.compileCount + group.assembleCount > 0 && <div className="smart-log__compile-summary">{compileSummary(group.compileCount, group.assembleCount)}</div>}
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
          {group.copiedDlls.length > 0 && <div className="smart-log__dll-section"><strong className="smart-log__dll-heading">{t("logs.dll.copiedHeading")}</strong><div className="smart-log__dll-list">{group.copiedDlls.map((dll, i) => <span className="smart-log__dll-tag smart-log__dll-tag--copied" key={`dll-${i}`}>{dll}</span>)}</div></div>}
          {group.skippedDllCount > 0 && <p className="smart-log__dll-note">{t(group.skippedDllCount === 1 ? "logs.dll.skipped.one" : "logs.dll.skipped.other", { count: group.skippedDllCount })}</p>}
          {group.systemDllCount > 0 && (
            <div className="smart-log__dll-section">
              <button className="smart-log__toggle-btn" type="button" onClick={onToggleSystemDlls}>{t(showSystemDlls ? "logs.dll.hideSystem" : "logs.dll.showSystem", { count: group.systemDllCount })}</button>
              {showSystemDlls && <div className="smart-log__dll-list">{group.entries.filter((e) => e.dllAction === "system").map((e, i) => <span className="smart-log__dll-tag smart-log__dll-tag--system" key={`sys-${i}`}>{e.dllDep}</span>)}</div>}
            </div>
          )}
        </>
      )}
      {group.phase === "tc-install" && (
        <>
          <div className="smart-log__pkg-list">{group.entries.filter((e) => e.message.startsWith("reinstalling ")).map((e, i) => <span className="smart-log__pkg-tag" key={`pkg-${i}`}>{e.message.replace("reinstalling ", "").trim()}</span>)}</div>
          {group.entries.map((e, i) => <RawLogEntry entry={e} id={`pkg-raw-${i}`} key={`pkg-raw-${i}`} />)}
        </>
      )}
      {group.phase === "ff-configure" && (
        <>
          {group.entries.filter(isKeyConfigureEntry).map((e, i) => <RawLogEntry entry={e} id={`cfg-key-${i}`} key={`cfg-key-${i}`} />)}
          <details className="smart-log__details"><summary className="smart-log__details-summary">{t("logs.configure.showAll", { count: group.entries.length })}</summary>{group.entries.map((e, i) => <RawLogEntry entry={e} id={`cfg-raw-${i}`} key={`cfg-raw-${i}`} />)}</details>
        </>
      )}
      {group.phase !== "ff-compile" && group.phase !== "ff-extraction" && group.phase !== "tc-install" && group.phase !== "ff-configure" && (
        group.entries.map((e, i) => <RawLogEntry entry={e} id={`${group.phase}-${i}`} key={`${group.phase}-${i}`} />)
      )}
    </div>
  );
}

function compileSummary(compileCount: number, assembleCount: number): string {
  if (assembleCount > 0) {
    const compileKey = compileCount === 1 ? "logs.compile.unit.cFile.one" : "logs.compile.unit.cFile.other";
    const assembleKey = assembleCount === 1 ? "logs.compile.unit.asmFile.one" : "logs.compile.unit.asmFile.other";
    return t("logs.compile.summaryWithAsm", {
      compileCount,
      compileUnit: t(compileKey),
      assembleCount,
      assembleUnit: t(assembleKey),
    });
  }
  const compileKey = compileCount === 1 ? "logs.compile.unit.cFile.one" : "logs.compile.unit.cFile.other";
  return t("logs.compile.summary", { compileCount, compileUnit: t(compileKey) });
}

function isKeyConfigureEntry(e: SecurityLogEntry): boolean {
  return e.message.startsWith("FFmpeg configure") || e.message.startsWith("Starting FFmpeg configure") || e.message.startsWith("License:") || e.message.startsWith("C compiler") || e.message.startsWith("C library") || e.message.startsWith("ARCH ") || e.message.startsWith("threading") || e.message.startsWith("static ") || e.message.startsWith("shared ") || e.message.startsWith("x86 assembler") || e.message.startsWith("Running approved");
}

function LogsEmptyCard(props: { onGoToPrep: () => void; onGoToBuild: () => void }) {
  return (
    <section className="card card--purple logs-empty-card">
      <span className="card__badge" aria-hidden="true">
        <img className="card__badge-icon" src={emptyStatePurpleIcon} alt="" />
      </span>
      <div className="card__head logs-empty-card__head">
        <h2 className="card__title">{t("logs.empty")}</h2>
        <p className="card__desc">{t("logs.empty.description")}</p>
      </div>
      <div className="logs-empty-card__actions">
        <button className="button" type="button" onClick={props.onGoToBuild}>{t("logs.actions.goToBuild")}</button>
        <button className="button button--primary" type="button" onClick={props.onGoToPrep}>{t("logs.actions.goToPrep")}</button>
      </div>
    </section>
  );
}

type LocalLogViewTabId = "summary" | "raw";
type NormalizedLocalLogKind = "toolchain" | "ffmpeg" | "unknown";

function normalizeLocalRecordKind(kind: string): NormalizedLocalLogKind {
  if (kind === "toolchain") return "toolchain";
  if (kind === "ffmpeg") return "ffmpeg";
  return "unknown";
}

function localLogKindLabel(kind: string): string {
  if (kind === "toolchain") return t("logs.local.kind.toolchain");
  if (kind === "ffmpeg") return t("logs.local.kind.ffmpeg");
  return t("logs.local.kind.unknown");
}

function localLogStatusLabel(status: string): string {
  return t(`logs.local.status.${status}`);
}

function makeLiveLocalRecords(toolchainLogEntries: SecurityLogEntry[], ffmpegLogEntries: SecurityLogEntry[]): LocalLogRecord[] {
  const records: LocalLogRecord[] = [];
  if (ffmpegLogEntries.length > 0) {
    records.push({
      runId: "live-ffmpeg",
      createdAt: "",
      displayTime: t("logs.local.live"),
      kind: "ffmpeg",
      status: "running",
      directory: "",
      entries: ffmpegLogEntries,
      rawText: ffmpegLogEntries.map((entry) => `[${entry.timestamp}] ${entry.level}: ${entry.message}`).join("\n"),
      errorCount: ffmpegLogEntries.filter((entry) => entry.level === "error").length,
      warnCount: ffmpegLogEntries.filter((entry) => entry.level === "warn").length,
      hasStdoutLog: false,
      hasStderrLog: false,
      hasSecurityEvents: false,
    });
  }
  if (toolchainLogEntries.length > 0) {
    records.push({
      runId: "live-toolchain",
      createdAt: "",
      displayTime: t("logs.local.live"),
      kind: "toolchain",
      status: "running",
      directory: "",
      entries: toolchainLogEntries,
      rawText: toolchainLogEntries.map((entry) => `[${entry.timestamp}] ${entry.level}: ${entry.message}`).join("\n"),
      errorCount: toolchainLogEntries.filter((entry) => entry.level === "error").length,
      warnCount: toolchainLogEntries.filter((entry) => entry.level === "warn").length,
      hasStdoutLog: false,
      hasStderrLog: false,
      hasSecurityEvents: false,
    });
  }
  return records;
}

function LocalLogSelector(props: { records: LocalLogRecord[]; selectedRunId: string; onSelect: (runId: string) => void; onOpenLogsFolder: () => Promise<void>; onOpenRecordFolder: (runId: string) => Promise<void> }) {
  const selectedRecord = props.records.find((record) => record.runId === props.selectedRunId);
  const canOpenRecordFolder = Boolean(selectedRecord && !selectedRecord.runId.startsWith("live-") && selectedRecord.directory);

  return (
    <section className="log-record-selector" aria-label={t("logs.local.selector.ariaLabel")}>
      <label className="log-record-selector__label" htmlFor="local-log-record-select">{t("logs.local.selector.label")}</label>
      <select
        id="local-log-record-select"
        className="log-record-selector__select"
        value={props.selectedRunId}
        onChange={(event) => props.onSelect(event.target.value)}
      >
        {props.records.map((record) => (
          <option value={record.runId} key={record.runId}>
            {record.displayTime} · {localLogKindLabel(record.kind)} · {localLogStatusLabel(record.status)}
          </option>
        ))}
      </select>
      <button
        className="button log-record-selector__open-button"
        type="button"
        disabled={!canOpenRecordFolder}
        onClick={() => selectedRecord && props.onOpenRecordFolder(selectedRecord.runId)}
      >
        {t("logs.local.actions.openRecordFolder")}
      </button>
      <button
        className="button log-record-selector__open-button"
        type="button"
        onClick={props.onOpenLogsFolder}
      >
        {t("logs.local.actions.openLogsFolder")}
      </button>
    </section>
  );
}

function RawLocalLogViewer(props: { text: string }) {
  if (!props.text.trim()) return <p className="empty-text">{t("logs.local.empty.raw")}</p>;
  return <pre className="log-raw-file-viewer">{props.text}</pre>;
}

function RawLocalLogFiles(props: { record: LocalLogRecord; onOpenFile: (runId: string, fileName: string) => Promise<void> }) {
  const files: { fileName: string; key: string; available: boolean }[] = [
    { fileName: "stdout.log", key: "stdout", available: props.record.hasStdoutLog },
    { fileName: "stderr.log", key: "stderr", available: props.record.hasStderrLog },
    { fileName: "security-events.jsonl", key: "events", available: props.record.hasSecurityEvents },
  ];
  if (!files.some((file) => file.available)) return <p className="empty-text">{t("logs.local.empty.raw")}</p>;
  return (
    <div className="log-raw-file-sections">
      <p className="log-raw-file-sections__intro">{t("logs.local.raw.intro")}</p>
      {files.map((file) => (
        <section className="log-raw-file-card" key={file.fileName}>
          <div className="log-raw-file-card__head">
            <h3 className="log-raw-file-card__title">{t(`logs.local.raw.${file.key}.title`)}</h3>
            <code className="log-raw-file-card__name">{file.fileName}</code>
          </div>
          <p className="log-raw-file-card__simple">{t(`logs.local.raw.${file.key}.simple`)}</p>
          <p className="log-raw-file-card__tech">{t(`logs.local.raw.${file.key}.tech`)}</p>
          <button
            className="button log-raw-file-card__open"
            type="button"
            disabled={!file.available}
            onClick={() => props.onOpenFile(props.record.runId, file.fileName)}
          >
            {file.available ? t("logs.local.raw.open") : t("logs.local.raw.missing")}
          </button>
        </section>
      ))}
    </div>
  );
}

function LocalLogDetails(props: { record: LocalLogRecord; onOpenRecordFile: (runId: string, fileName: string) => Promise<void> }) {
  const [activeTabId, setActiveTabId] = useState<LocalLogViewTabId>("summary");
  const kind = normalizeLocalRecordKind(props.record.kind);
  const context = kind === "toolchain" ? "toolchain" : kind === "ffmpeg" ? "ffmpeg" : null;
  const isLive = props.record.runId.startsWith("live-");
  const hasDetails = isLive || props.record.entries.length > 0 || props.record.rawText.trim().length > 0;
  const rawCount = isLive
    ? (props.record.rawText.trim() ? 1 : 0)
    : [props.record.hasStdoutLog, props.record.hasStderrLog, props.record.hasSecurityEvents].filter(Boolean).length;
  const tabs: { id: LocalLogViewTabId; label: string; count: number }[] = kind === "unknown"
    ? [{ id: "raw", label: t("logs.local.tabs.unknownRaw"), count: rawCount }]
    : [
      { id: "summary", label: kind === "toolchain" ? t("logs.local.tabs.environmentSummary") : t("logs.local.tabs.ffmpegSummary"), count: props.record.entries.length },
      { id: "raw", label: kind === "toolchain" ? t("logs.local.tabs.environmentRaw") : t("logs.local.tabs.ffmpegRaw"), count: rawCount },
    ];

  const effectiveActiveTabId = tabs.some((tab) => tab.id === activeTabId) ? activeTabId : tabs[0]?.id ?? "raw";

  return (
    <section className="result-details-card log-details-card">
      <div className="result-details-tabs" role="tablist" aria-label={t("logs.local.tabs.ariaLabel")}>
        {tabs.map((tab) => (
          <button
            className={`result-details-tab ${effectiveActiveTabId === tab.id ? "result-details-tab--active" : ""}`}
            type="button"
            role="tab"
            aria-selected={effectiveActiveTabId === tab.id}
            onClick={() => setActiveTabId(tab.id)}
            key={tab.id}
          >
            <span>{tab.label}</span>
            <span className="result-details-tab__count">{tab.count}</span>
          </button>
        ))}
      </div>
      <div className="result-details-body log-details-body">
        <div className="log-record-meta">
          <span>{props.record.displayTime}</span>
          <span>{localLogKindLabel(props.record.kind)}</span>
          <span>{localLogStatusLabel(props.record.status)}</span>
          {props.record.errorCount > 0 && <span>{t("logs.local.meta.errors", { count: props.record.errorCount })}</span>}
          {props.record.warnCount > 0 && <span>{t("logs.local.meta.warnings", { count: props.record.warnCount })}</span>}
        </div>
        {(() => {
          const showRaw = effectiveActiveTabId === "raw" || !context;
          if (showRaw && !isLive) {
            return <RawLocalLogFiles record={props.record} onOpenFile={props.onOpenRecordFile} />;
          }
          if (!hasDetails) {
            return <p className="empty-text">{t("logs.local.loading")}</p>;
          }
          if (showRaw) {
            return <RawLocalLogViewer text={props.record.rawText} />;
          }
          if (context) {
            return <SmartLogViewer entries={props.record.entries} context={context} viewMode="smart" hideModeToolbar />;
          }
          return null;
        })()}
      </div>
    </section>
  );
}

export type LogsTabProps = {
  toolchainLogEntries: SecurityLogEntry[];
  ffmpegLogEntries: SecurityLogEntry[];
  localLogRecords: LocalLogRecord[];
  localLogRecordsError: string;
  refreshLocalLogRecords: () => Promise<void>;
  loadLocalLogRecord: (runId: string) => Promise<void>;
  openLocalLogsFolder: () => Promise<void>;
  openLocalLogRecordFolder: (runId: string) => Promise<void>;
  openLocalLogRecordFile: (runId: string, fileName: string) => Promise<void>;
  onGoToPrep: () => void;
  onGoToBuild: () => void;
};

export function LogsTab({ toolchainLogEntries, ffmpegLogEntries, localLogRecords, localLogRecordsError, refreshLocalLogRecords, loadLocalLogRecord, openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile, onGoToPrep, onGoToBuild }: LogsTabProps) {
  const liveRecords = useMemo(() => makeLiveLocalRecords(toolchainLogEntries, ffmpegLogEntries), [toolchainLogEntries, ffmpegLogEntries]);
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
    requestedDetailRunIds.current.add(selectedRecord.runId);
    void loadLocalLogRecord(selectedRecord.runId);
  }, [selectedRecord?.runId, selectedRecord?.entries.length, selectedRecord?.rawText]);

  return (
    <section className="tab-page log-page">
      <header className="result-page__header">
        <div>
          <h1 className="page-header__title">{t("logs.title")}</h1>
          <p className="page-header__text">{t("logs.intro")}</p>
        </div>
        <div className="result-page__actions">
          <button className="button" type="button" onClick={refreshLocalLogRecords}>{t("logs.actions.refresh")}</button>
        </div>
      </header>
      {localLogRecordsError && <p className="empty-text">{localLogRecordsError}</p>}
      {!localLogRecordsError && records.length > 0 && selectedRecord && (
        <>
          <LocalLogSelector records={records} selectedRunId={selectedRecord.runId} onSelect={setSelectedRunId} onOpenLogsFolder={openLocalLogsFolder} onOpenRecordFolder={openLocalLogRecordFolder} />
          <LocalLogDetails record={selectedRecord} onOpenRecordFile={openLocalLogRecordFile} />
        </>
      )}
      {!localLogRecordsError && records.length === 0 && <LogsEmptyCard onGoToPrep={onGoToPrep} onGoToBuild={onGoToBuild} />}
    </section>
  );
}
