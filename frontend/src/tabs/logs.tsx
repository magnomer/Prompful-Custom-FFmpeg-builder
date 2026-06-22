import React, { useMemo, useState } from "react";
import { t } from "../i18n";
import { PageHeader } from "./shared";
import {
  SecurityLogEntry, LogPhaseId, LogPhaseGroup,
  TOOLCHAIN_PHASE_ORDER, FFMPEG_PHASE_ORDER,
  parseLogEntry, buildPhaseGroups, runtimeLogText,
} from "./logutils";

export type { SecurityLogEntry, LogPhaseGroup };
export type { SecurityLogPayload, ApprovedActionStatusPayload, LiveProgress, LogPhaseId, ParsedLogEntry } from "./logutils";
export { computeProgress, getToolchainPipeline, getFfmpegPipeline } from "./logutils";

function SmartLogViewer(props: { entries: SecurityLogEntry[]; context?: "toolchain" | "ffmpeg" }) {
  const [expandedPhases, setExpandedPhases] = useState<Set<LogPhaseId>>(new Set());
  const [showSystemDlls, setShowSystemDlls] = useState(false);
  const [viewMode, setViewMode] = useState<"smart" | "raw">("smart");

  const parsed = useMemo(() => props.entries.map((e) => parseLogEntry(e, props.context ?? "ffmpeg")), [props.entries, props.context]);
  const phaseOrder = props.context === "toolchain" ? TOOLCHAIN_PHASE_ORDER : props.context === "ffmpeg" ? FFMPEG_PHASE_ORDER : [...TOOLCHAIN_PHASE_ORDER, ...FFMPEG_PHASE_ORDER];
  const phaseGroups = useMemo(() => buildPhaseGroups(parsed, phaseOrder), [parsed, phaseOrder]);
  const finalEntry = useMemo(() => [...parsed].reverse().find((e) => e.isFinalStatus), [parsed]);
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
      <div className="smart-log__toolbar">
        <button className={`smart-log__mode-btn ${viewMode === "smart" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setViewMode("smart")}>{t("logs.view.summary")}</button>
        <button className={`smart-log__mode-btn ${viewMode === "raw" ? "smart-log__mode-btn--active" : ""}`} type="button" onClick={() => setViewMode("raw")}>{t("logs.view.raw", { count: props.entries.length })}</button>
      </div>

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
          {finalEntry && (
            <div className={`smart-log__banner smart-log__banner--${finalEntry.level}`}>
              <span className="smart-log__banner-icon">{finalEntry.level === "error" ? t("logs.banner.errorMark") : t("logs.banner.okMark")}</span>
              <span className="smart-log__banner-text">{runtimeLogText(finalEntry.message)}</span>
              <time className="smart-log__banner-time">{finalEntry.timestamp}</time>
            </div>
          )}
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
          <details className="smart-log__details"><summary className="smart-log__details-summary">{t("logs.packages.showAll", { count: group.entries.length })}</summary>{group.entries.map((e, i) => <RawLogEntry entry={e} id={`pkg-raw-${i}`} key={`pkg-raw-${i}`} />)}</details>
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

export type LogsTabProps = {
  toolchainLogEntries: SecurityLogEntry[];
  ffmpegLogEntries: SecurityLogEntry[];
};

export function LogsTab({ toolchainLogEntries, ffmpegLogEntries }: LogsTabProps) {
  return (
    <section className="tab-page">
      <PageHeader title={t("logs.title")} text={t("logs.intro")} />
      {toolchainLogEntries.length > 0 && <section className="smart-log__section"><h2 className="smart-log__section-title">{t("logs.sections.toolchain")}</h2><SmartLogViewer entries={toolchainLogEntries} context="toolchain" /></section>}
      {ffmpegLogEntries.length > 0 && <section className="smart-log__section"><h2 className="smart-log__section-title">{t("logs.sections.ffmpeg")}</h2><SmartLogViewer entries={ffmpegLogEntries} context="ffmpeg" /></section>}
      {toolchainLogEntries.length === 0 && ffmpegLogEntries.length === 0 && <p className="empty-text">{t("logs.empty")}</p>}
    </section>
  );
}
