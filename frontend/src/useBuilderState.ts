import { useEffect, useMemo, useRef, useState } from "react";
import {
  LPlanFFmpegApprove,
  LFFmpegCompilationLaunch,
  LPlanToolchainApprove,
  LActionApprovedCancel,
  LToolchainEnvironmentClear,
  LResultBuildGet,
  LVerificationBuildRun,
  LStateInitialGet,
  LStatusToolchainGet,
  LToolchainProfileList,
  LRecordLogGet,
  LLogRecordList,
  LFolderRecordOpen,
  LFileRecordOpen,
  LFolderLogOpen,
  LToolchainInstallVerify,
  LCatalogSourceGet,
  LPresetSourceGet,
  LStateUiLoad,
  LStateUiSave,
  LPlanFFmpegRequest,
  LPlanToolchainRequest,
  LDirectoryResultOpen,
  LReportResultOpen,
  LWorkspaceSelect,
  LLocaleSet,
} from "../wailsjs/go/program/LProgram";
import { BrowserOpenURL, EventsOn, WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../wailsjs/runtime/runtime";
import { planning } from "../wailsjs/go/models";

import { LLogSecurityEntry, LLogSecurityPayload, LStatusActionPayload, LStalledActionPayload, LProgressLive, LProgressGet } from "./tabs/logutils";
import {
  LPresetLibraryId, LPresetLibrary, LPresetLibraryClean, LPresetLibraryResolve,
  LLibrarySelectionNormalize, LPresetLibraryMatch, LPresetLibraryValidate,
  LLicenseBoundaryGet, LLibraryExclusiveRemove,
  LLibraryTestGet,
} from "./tabs/libraries";
import { LPresetOptionId, LPresetOptionCatalog } from "./tabs/options";
import {
  LTabIdentifier, LStateUiSaved, LStateWindowSaved, LStateWindowKey,
  LSettingsBuildEmpty, LSettingsFFmpegEmpty, LStateInitialDefault,
  LTextLineSplit, LLogLevelNormalize, LTabIdValidate,
  LRequestApprovalCreate, LStateUiParse, LStateWindowRead,
  LPackagePrefixUpdate, LSettingsBuildNormalize, LSettingsFFmpegNormalize, LStateInitialNormalize,
} from "./programstate";
import { LLocaleTextGet, LLocaleGet } from "./i18n";

// ─── Window state helpers ─────────────────────────────────────────────────────

function LRuntimeNumberRead(value: unknown, tupleIndex: number, primary: string, fallback: string): number {
  if (Array.isArray(value)) return Number(value[tupleIndex]);
  if (value && typeof value === "object") {
    const r = value as Record<string, unknown>;
    return Number(r[primary] ?? r[fallback] ?? 0);
  }
  return 0;
}

function LStateWindowRestore() {
  const s = LStateWindowRead();
  if (Number.isFinite(s.width) && Number.isFinite(s.height)) WindowSetSize(Number(s.width), Number(s.height));
  if (Number.isFinite(s.x) && Number.isFinite(s.y)) WindowSetPosition(Number(s.x), Number(s.y));
}

async function LStateWindowSave() {
  try {
    const sz = await WindowGetSize();
    const pos = await WindowGetPosition();
    const width  = LRuntimeNumberRead(sz,  0, "w", "width");
    const height = LRuntimeNumberRead(sz,  1, "h", "height");
    const x = LRuntimeNumberRead(pos, 0, "x", "left");
    const y = LRuntimeNumberRead(pos, 1, "y", "top");
    window.localStorage.setItem(LStateWindowKey, JSON.stringify({ width, height, x, y } satisfies LStateWindowSaved));
  } catch {
    // Window persistence is best-effort. Never block the program if the runtime is unavailable.
  }
}

// ─── Hook ────────────────────────────────────────────────────────────────────

export function LStateBuilderUse() {
  const [activeTabId, setActiveTabId] = useState<LTabIdentifier>("source");
  const hasLoadedSavedState = useRef(false);
  const tabPanelRef = useRef<HTMLElement>(null);
  // Tabs whose scroll position is remembered when leaving and restored when
  // returning (within the session). Every other tab still scrolls back to top.
  const scrollRememberedTabIds = useRef(new Set<LTabIdentifier>(["options", "buildFfmpeg"]));
  const tabScrollPositions = useRef<Partial<Record<LTabIdentifier, number>>>({});
  const activeTabIdRef = useRef<LTabIdentifier>("source");
  const [initialProgramState, setInitialProgramState] = useState<LStateInitial>(LStateInitialDefault);
  const [libraryCatalog, setLibraryCatalog] = useState<LLibraryChoice[]>([]);
  const [libraryPresetCatalog, setLibraryPresetCatalog] = useState<LPresetLibrary[]>([]);
  const [buildConfigSettings, setBuildConfigSettings] = useState<LSettingsToolchain>(LSettingsBuildEmpty);
  const [ffmpegBuildSettings, setFfmpegBuildSettings] = useState<LSettingsFFmpeg>(LSettingsFFmpegEmpty);
  const [libraryPresetId, setLibraryPresetId] = useState<LPresetLibraryId>("default");
  const [extendedLibraries, setExtendedLibrariesState] = useState(false);
  const [libraryDetailedView, setLibraryDetailedView] = useState(false);
  const [optionsDetailedView, setOptionsDetailedView] = useState(false);
  const [libraryTechnicalDetails, setLibraryTechnicalDetails] = useState(false);
  const [optionsTechnicalDetails, setOptionsTechnicalDetails] = useState(false);
  const [librarySectionFilters, setLLibrarySectionFilters] = useState<string[]>([]);
  const [msys2PackageText, setMsys2PackageText] = useState("");
  const [extraConfigureFlagText, setExtraConfigureFlagText] = useState("");
  const [toolchainPreparationPlanReview, setToolchainPreparationPlanReview] = useState<LReviewToolchain | null>(null);
  const [ffmpegBuildPlanReview, setFfmpegBuildPlanReview] = useState<LReviewFFmpeg | null>(null);
  const [approvedActionStatus, setApprovedActionStatus] = useState("idle");
  const [approvedActionPhase, setApprovedActionPhase] = useState<"toolchain" | "ffmpeg" | null>(null);
  // Mirror addresses tried before a transient-network stall halted the run,
  // delivered by the "approved-action-stalled" event for the stalled banner.
  const [ffmpegStalledAddresses, setFfmpegStalledAddresses] = useState<string[]>([]);
  // The plan and approval of the last-launched FFmpeg run, kept so Retry can
  // re-invoke the same approved action after a stall (the backend resumes from
  // cache). The review session is single-use and gone by then, so Retry uses
  // the direct run-start binding rather than the review-approval path.
  const pFfmpegRunLast = useRef<{ plan: planning.LPlanFFmpeg; approval: LRequestApproval } | null>(null);
  const [toolchainLogEntries, setToolchainLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [ffmpegLogEntries, setFfmpegLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [localLogRecords, setLocalLogRecords] = useState<LRecordLog[]>([]);
  const [localLogRecordsError, setLocalLogRecordsError] = useState("");
  const [buildResult, setBuildResult] = useState<LResultState | null>(null);
  const [buildResultError, setBuildResultError] = useState("");
  const [isLoadingBuildResult, setIsLoadingBuildResult] = useState(false);
  const [buildVerification, setBuildVerification] = useState<LVerificationState | null>(null);
  const [buildVerificationError, setBuildVerificationError] = useState("");
  const [isVerifyingBuild, setIsVerifyingBuild] = useState(false);
  const [toolchainStatus, setToolchainStatus] = useState<LStatusToolchain | null>(null);
  const [installedToolchainProfiles, setInstalledToolchainProfiles] = useState<LStatusToolchain[]>([]);
  const [toolchainVerification, setToolchainVerification] = useState<LVerificationToolchain | null>(null);
  const [isVerifyingToolchain, setIsVerifyingToolchain] = useState(false);

  const approvedActionPhaseRef = useRef<"toolchain" | "ffmpeg" | null>(null);
  const libraryPresetIdRef = useRef<LPresetLibraryId>("default");
  approvedActionPhaseRef.current = approvedActionPhase;
  libraryPresetIdRef.current = libraryPresetId;
  activeTabIdRef.current = activeTabId;

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed" && approvedActionStatus !== "stalled", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";
  const toolchainProgress = useMemo<LProgressLive>(() => LProgressGet(toolchainLogEntries, approvedActionStatus, "toolchain"), [toolchainLogEntries, approvedActionStatus]);
  const ffmpegProgress = useMemo<LProgressLive>(() => LProgressGet(ffmpegLogEntries, approvedActionStatus, "ffmpeg"), [ffmpegLogEntries, approvedActionStatus]);
  const securityLogEntries = useMemo(() => [...toolchainLogEntries, ...ffmpegLogEntries], [toolchainLogEntries, ffmpegLogEntries]);
  // Current Build configuration package list (live textarea), used by Prep to flag
  // drift between what is configured now and what the prepared toolchain installed.
  const configuredMsys2PackageNames = useMemo(() => LTextLineSplit(msys2PackageText), [msys2PackageText]);

  useEffect(() => {
    // Keep the backend in sync with the UI language so the native confirmation
    // dialog is shown in the same language. Logs stay English by design.
    void LLocaleSet(LLocaleGet());
    const onLocaleChange = () => { void LLocaleSet(LLocaleGet()); };
    window.addEventListener("customffmpeg-locale-change", onLocaleChange);

    LStateInitialGet().then(async (rawNextState: LStateInitial) => {
      const nextState = LStateInitialNormalize(rawNextState);
      const saved = LStateUiParse(await LStateUiLoad());
      const savedBts = LSettingsBuildNormalize(saved.buildConfigSettings, nextState.defaultBuildConfigSettings);
      let resolvedFbs = LSettingsFFmpegNormalize(saved.ffmpegBuildSettings, nextState.defaultFfmpegBuildSettings);
      const initialPresetCatalog = LPresetLibraryClean(nextState.defaultLibraryPresetCatalog);
      const hasSavedPreset = LPresetLibraryValidate(saved.libraryPresetId);
      const resolvedPresetId: LPresetLibraryId = hasSavedPreset ? saved.libraryPresetId! : "default";
      if (!hasSavedPreset && !saved.ffmpegBuildSettings) {
        const defaultPreset = initialPresetCatalog.find((p) => p.presetId === "default");
        if (defaultPreset) {
          const nextIds = LLibrarySelectionNormalize(defaultPreset.libraryIds, resolvedFbs.windowsShellProfileName, nextState.defaultLibraryCatalog);
          resolvedFbs = { ...resolvedFbs, selectedLibraryIds: nextIds, licenseProfileName: LLicenseBoundaryGet(nextIds, nextState.defaultLibraryCatalog, resolvedFbs.windowsShellProfileName) };
        }
      }
      setInitialProgramState(nextState);
      setLibraryCatalog(nextState.defaultLibraryCatalog);
      setLibraryPresetCatalog(initialPresetCatalog);
      setBuildConfigSettings(savedBts);
      setFfmpegBuildSettings(resolvedFbs);
      setMsys2PackageText(saved.msys2PackageText ?? savedBts.msys2PackageNames.join("\n"));
      setExtraConfigureFlagText(saved.extraConfigureFlagText ?? resolvedFbs.extraConfigureFlags.join("\n"));
      setLibraryPresetId(resolvedPresetId);
      if (typeof saved.extendedLibraries === "boolean") setExtendedLibrariesState(saved.extendedLibraries);
      if (typeof saved.libraryDetailedView === "boolean") setLibraryDetailedView(saved.libraryDetailedView);
      if (typeof saved.optionsDetailedView === "boolean") setOptionsDetailedView(saved.optionsDetailedView);
      if (typeof saved.libraryTechnicalDetails === "boolean") setLibraryTechnicalDetails(saved.libraryTechnicalDetails);
      if (typeof saved.optionsTechnicalDetails === "boolean") setOptionsTechnicalDetails(saved.optionsTechnicalDetails);
      if (Array.isArray(saved.librarySectionFilters)) setLLibrarySectionFilters(saved.librarySectionFilters.filter((value): value is string => typeof value === "string"));
      if (LTabIdValidate(saved.activeTabId)) setActiveTabId(saved.activeTabId);
      hasLoadedSavedState.current = true;
      LStateWindowRestore();
    }).catch((err) => {
      setLocalLogRecordsError(err instanceof Error ? err.message : String(err));
      hasLoadedSavedState.current = true;
    });

    const makeEntry = (payload: LLogSecurityPayload): LLogSecurityEntry => ({
      level: LLogLevelNormalize(payload.level),
      message: payload.message,
      timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }),
    });

    const removeLogListener = EventsOn("security-log", (payload: LLogSecurityPayload) => {
      const entry = makeEntry(payload);
      if (approvedActionPhaseRef.current === "ffmpeg") setFfmpegLogEntries((prev) => [...prev, entry]);
      else setToolchainLogEntries((prev) => [...prev, entry]);
    });
    const removeStatusListener = EventsOn("approved-action-status", (payload: LStatusActionPayload) => {
      setApprovedActionStatus(payload.status);
      if (payload.status === "failed") { setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null); setBuildResult(null); }
      if (payload.status === "completed") { setBuildResult(null); setApprovedActionPhase(null); }
      // A stall halts the run in a non-active retryable state: drop the "ffmpeg"
      // phase so the live progress stops reading as in-flight and the orange
      // stalled banner (not a spinner) becomes the authoritative signal.
      if (payload.status === "stalled") { setApprovedActionPhase(null); }
    });
    const removeStalledListener = EventsOn("approved-action-stalled", (payload: LStalledActionPayload) => {
      setFfmpegStalledAddresses(Array.isArray(payload.addresses) ? payload.addresses : []);
    });
    return () => { removeLogListener(); removeStatusListener(); removeStalledListener(); window.removeEventListener("customffmpeg-locale-change", onLocaleChange); };
  }, []);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    let isCurrent = true;
    Promise.all([
      LCatalogSourceGet(ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName),
      LPresetSourceGet(ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName),
    ])
      .then(([LCatalogLibrarySource, LCatalogPresetSource]) => {
        if (!isCurrent) return;
        const nextLibraryCatalog = Array.isArray(LCatalogLibrarySource) ? LCatalogLibrarySource : [];
        const nextPresetCatalog = LPresetLibraryClean(LCatalogPresetSource);
        setLibraryCatalog(nextLibraryCatalog);
        setLibraryPresetCatalog(nextPresetCatalog);
        setFfmpegBuildSettings((settings) => {
          const activePreset = nextPresetCatalog.find((preset) => preset.presetId === libraryPresetIdRef.current && preset.presetId !== "custom");
          const nextLibraryIds = activePreset
            ? (activePreset.dev
              ? LLibraryTestGet(nextLibraryCatalog, settings.windowsShellProfileName)
              : LLibrarySelectionNormalize(LPresetLibraryResolve(activePreset, extendedLibraries), settings.windowsShellProfileName, nextLibraryCatalog))
            : LLibrarySelectionNormalize(settings.selectedLibraryIds, settings.windowsShellProfileName, nextLibraryCatalog);
          setLibraryPresetId(LPresetLibraryMatch(nextLibraryIds, nextPresetCatalog, settings.windowsShellProfileName, nextLibraryCatalog, extendedLibraries, libraryPresetIdRef.current));
          return {
            ...settings,
            selectedLibraryIds: nextLibraryIds,
            licenseProfileName: LLicenseBoundaryGet(nextLibraryIds, nextLibraryCatalog, settings.windowsShellProfileName),
          };
        });
        setFfmpegBuildPlanReview(null);
      })
      .catch(() => {
        if (isCurrent) { setLibraryCatalog(initialProgramState.defaultLibraryCatalog); setLibraryPresetCatalog(LPresetLibraryClean(initialProgramState.defaultLibraryPresetCatalog)); }
      });
    return () => { isCurrent = false; };
  }, [ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName, initialProgramState.defaultLibraryCatalog, initialProgramState.defaultLibraryPresetCatalog, extendedLibraries]);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    void LStateUiSave(JSON.stringify({
      activeTabId,
      buildConfigSettings: { ...buildConfigSettings, msys2PackageNames: LTextLineSplit(msys2PackageText) },
      ffmpegBuildSettings: { ...ffmpegBuildSettings, extraConfigureFlags: LTextLineSplit(extraConfigureFlagText), configureFlags: LTextLineSplit(extraConfigureFlagText) },
      msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters,
    } satisfies LStateUiSaved));
  }, [activeTabId, buildConfigSettings, ffmpegBuildSettings, msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters]);

  useEffect(() => {
    const id = window.setInterval(() => { LStateWindowSave(); }, 2000);
    window.addEventListener("beforeunload", LStateWindowSave);
    return () => { window.clearInterval(id); window.removeEventListener("beforeunload", LStateWindowSave); LStateWindowSave(); };
  }, []);

  useEffect(() => {
    if (activeTabId === "result") refreshBuildResult();
    if (activeTabId === "logs") refreshLocalLogRecords();
  }, [activeTabId, buildConfigSettings.workspaceDirectory]);

  // Recover the "already prepared" state from disk so Prep remembers a prior
  // successful install across relaunches. Re-check when entering Prep or when the
  // workspace changes; clear any stale deep-verify result on workspace change.
  useEffect(() => {
    if (activeTabId === "prep") refreshToolchainStatus();
  }, [activeTabId, buildConfigSettings.workspaceDirectory, buildConfigSettings.windowsShellProfileName]);

  useEffect(() => {
    setToolchainVerification(null);
  }, [buildConfigSettings.workspaceDirectory, buildConfigSettings.windowsShellProfileName]);

  // After a toolchain run finishes, re-read disk so the recovery card reflects
  // the fresh install. This effect re-runs with current settings, avoiding the
  // stale closure of the once-registered event listener.
  useEffect(() => {
    if (approvedActionStatus === "completed") refreshToolchainStatus();
    if (activeTabId === "logs" && (approvedActionStatus === "completed" || approvedActionStatus === "failed")) refreshLocalLogRecords();
  }, [approvedActionStatus]);

  // Continuously record the scroll position of the active tab so it can be
  // restored on return. Registered once; reads the current tab via a ref.
  useEffect(() => {
    const el = tabPanelRef.current;
    if (!el) return;
    const PScrollSave = () => { tabScrollPositions.current[activeTabIdRef.current] = el.scrollTop; };
    el.addEventListener("scroll", PScrollSave, { passive: true });
    return () => el.removeEventListener("scroll", PScrollSave);
  }, []);

  // On tab change, restore the remembered scroll position for the remembered
  // tabs; every other tab scrolls back to the top.
  useEffect(() => {
    const el = tabPanelRef.current;
    if (!el) return;
    el.scrollTop = scrollRememberedTabIds.current.has(activeTabId)
      ? tabScrollPositions.current[activeTabId] ?? 0
      : 0;
  }, [activeTabId]);

  async function refreshLocalLogRecords() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setLocalLogRecords([]); setLocalLogRecordsError(""); return; }
    try {
      const records = await LLogRecordList(dir);
      setLocalLogRecords((previousRecords) => {
        const previousByRunId = new Map(previousRecords.map((record) => [record.runId, record]));
        return records.map((record) => {
          const previous = previousByRunId.get(record.runId);
          const normalizedRecord = {
            ...record,
            entries: (record.entries ?? []).map((entry) => ({ ...entry, level: LLogLevelNormalize(entry.level) })),
          };
          if (previous && ((previous.entries?.length ?? 0) > 0 || (previous.rawText ?? "").trim())) {
            return { ...normalizedRecord, entries: previous.entries, rawText: previous.rawText };
          }
          return normalizedRecord;
        });
      });
      setLocalLogRecordsError("");
    }
    catch (err) { setLocalLogRecords([]); setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function loadLocalLogRecord(runId: string) {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir || !runId || runId.startsWith("live-")) return;
    const existing = localLogRecords.find((record) => record.runId === runId);
    if (existing && ((existing.entries?.length ?? 0) > 0 || (existing.rawText ?? "").trim())) return;
    try {
      const record = await LRecordLogGet(dir, runId);
      const normalizedRecord = {
        ...record,
        entries: (record.entries ?? []).map((entry) => ({ ...entry, level: LLogLevelNormalize(entry.level) })),
      };
      setLocalLogRecords((records) => records.map((item) => item.runId === runId ? normalizedRecord : item));
      setLocalLogRecordsError("");
    }
    catch (err) { setLocalLogRecordsError(err instanceof Error ? err.message : String(err)); }
  }

  async function refreshBuildResult() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResult(null); setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    setBuildVerification(null); setBuildVerificationError("");
    setIsLoadingBuildResult(true);
    try { const r = await LResultBuildGet(dir); setBuildResult(r); setBuildResultError(""); }
    catch (err) { setBuildResult(null); setBuildResultError(err instanceof Error ? err.message : String(err)); }
    finally { setIsLoadingBuildResult(false); }
  }

  async function verifyBuildResult() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildVerificationError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    setIsVerifyingBuild(true);
    setBuildVerificationError("");
    try { setBuildVerification(await LVerificationBuildRun(dir)); }
    catch (err) { setBuildVerification(null); setBuildVerificationError(err instanceof Error ? err.message : String(err)); }
    finally { setIsVerifyingBuild(false); }
  }

  async function refreshToolchainStatus() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setToolchainStatus(null); setInstalledToolchainProfiles([]); return; }
    try { setToolchainStatus(await LStatusToolchainGet(dir, buildConfigSettings.windowsShellProfileName)); }
    catch { setToolchainStatus(null); }
    try { setInstalledToolchainProfiles(await LToolchainProfileList(dir)); }
    catch { setInstalledToolchainProfiles([]); }
  }


  async function clearBuildEnvironments() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) return;
    const confirmed = window.confirm(LLocaleTextGet("prep.profiles.clearConfirm"));
    if (!confirmed) return;
    setToolchainVerification(null);
    await LToolchainEnvironmentClear(dir);
    await refreshToolchainStatus();
  }

  async function verifyToolchain() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) return;
    setIsVerifyingToolchain(true);
    setToolchainVerification(null);
    try { setToolchainVerification(await LToolchainInstallVerify(dir, buildConfigSettings.windowsShellProfileName)); }
    catch (err) { setToolchainVerification({ verified: false, checkedPackageCount: 0, missingPackageNames: [], message: err instanceof Error ? err.message : String(err) }); }
    finally { setIsVerifyingToolchain(false); }
  }

  async function openResultFolder() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    await LDirectoryResultOpen(dir);
  }

  async function openResultReport() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResultError(LLocaleTextGet("result.error.chooseWorkspaceFirst")); return; }
    await LReportResultOpen(dir);
  }

  async function openLocalLogsFolder() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) return;
    await LFolderLogOpen(dir);
  }

  async function openLocalLogRecordFolder(runId: string) {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir || !runId || runId.startsWith("live-")) return;
    await LFolderRecordOpen(dir, runId);
  }

  async function openLocalLogRecordFile(runId: string, fileName: string) {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir || !runId || runId.startsWith("live-")) return;
    await LFileRecordOpen(dir, runId, fileName);
  }

  function LSettingsToolchainUpdate(next: Partial<LSettingsToolchain>) {
    setBuildConfigSettings((s) => ({ ...s, ...next }));
    setToolchainPreparationPlanReview(null);
  }

  function LSettingsFFmpegUpdate(next: Partial<LSettingsFFmpeg>) {
    setFfmpegBuildSettings((s) => ({ ...s, ...next }));
    setFfmpegBuildPlanReview(null);
  }

  // Changing the shell profile must remap the MSYS2 package prefixes in the
  // toolchain textarea too, otherwise a non-ucrt64 profile would still install
  // ucrt64 toolchain packages (wrong compiler/runtime for the chosen prefix).
  function LProfileShellUpdate(profileName: string) {
    // Re-normalize the library selection for the new profile so libraries with no
    // package there (e.g. onnxruntime on mingw64) are dropped, and recompute the
    // preset/license to match.
    const nextLibraryIds = LLibrarySelectionNormalize(ffmpegBuildSettings.selectedLibraryIds, profileName, libraryCatalog);
    LSettingsToolchainUpdate({ windowsShellProfileName: profileName });
    LSettingsFFmpegUpdate({
      windowsShellProfileName: profileName,
      selectedLibraryIds: nextLibraryIds,
      licenseProfileName: LLicenseBoundaryGet(nextLibraryIds, libraryCatalog, profileName),
    });
    setLibraryPresetId(LPresetLibraryMatch(nextLibraryIds, libraryPresetCatalog, profileName, libraryCatalog, extendedLibraries, libraryPresetId));
    setMsys2PackageText((text) => LPackagePrefixUpdate(text, profileName));
  }

  function LMSYSArchiveUpdate(nextUrl: string) {
    setBuildConfigSettings((s) => {
      const oldAuto = s.msys2ArchiveUrl ? `${s.msys2ArchiveUrl}.sig` : "";
      const shouldUpdate = s.msys2ArchiveSignatureUrl === "" || s.msys2ArchiveSignatureUrl === oldAuto;
      return { ...s, msys2ArchiveUrl: nextUrl, msys2ArchiveSignatureUrl: shouldUpdate && nextUrl ? `${nextUrl}.sig` : s.msys2ArchiveSignatureUrl };
    });
    setToolchainPreparationPlanReview(null);
  }

  async function chooseWorkspaceDirectory() {
    const dir = await LWorkspaceSelect();
    if (!dir) return;
    LSettingsToolchainUpdate({ workspaceDirectory: dir });
    LSettingsFFmpegUpdate({ workspaceDirectory: dir });
  }

  async function addBuildConfigPlanAndContinueToPrep() {
    const review = await LPlanToolchainRequest({ ...buildConfigSettings, msys2PackageNames: LTextLineSplit(msys2PackageText) });
    setToolchainPreparationPlanReview(review);
    setActiveTabId("prep");
  }

  function LPlanToolchainCancel() {
    setToolchainPreparationPlanReview(null);
  }

  async function reviewFfmpegPlans() {
    const flags = LTextLineSplit(extraConfigureFlagText);
    const review = await LPlanFFmpegRequest({ ...ffmpegBuildSettings, extraConfigureFlags: flags, configureFlags: flags });
    setFfmpegBuildPlanReview(review);
    setActiveTabId("buildFfmpeg");
  }

  async function approveToolchainPreparationPlan() {
    if (!toolchainPreparationPlanReview) return;
    const r = toolchainPreparationPlanReview;
    setToolchainLogEntries([]); setApprovedActionPhase("toolchain"); setApprovedActionStatus("starting");
    try {
      await LPlanToolchainApprove(r.reviewSessionId, LRequestApprovalCreate(r.plan.actionName, r.plan.planHash, r.expectedLConsentText));
      setToolchainPreparationPlanReview(null); setActiveTabId("prep");
    } catch (err) {
      setApprovedActionStatus("failed");
      setToolchainLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function approveFfmpegBuildPlan() {
    if (!ffmpegBuildPlanReview) return;
    const r = ffmpegBuildPlanReview;
    const approval = LRequestApprovalCreate(r.plan.actionName, r.plan.planHash, r.expectedLConsentText);
    setFfmpegLogEntries([]); setFfmpegStalledAddresses([]); setApprovedActionStatus("starting");
    try {
      await LPlanFFmpegApprove(r.reviewSessionId, approval);
      // Keep the approved plan and approval so Retry can resume the same action.
      // r.plan is the full backend plan; the ambient review type omits a few
      // generated fields, so narrow it to the binding's plan type for reuse.
      pFfmpegRunLast.current = { plan: r.plan as unknown as planning.LPlanFFmpeg, approval };
      setApprovedActionPhase("ffmpeg");
      setFfmpegBuildPlanReview(null); setActiveTabId("buildFfmpeg");
    } catch (err) {
      setApprovedActionStatus("failed");
      setFfmpegLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  // Re-run the same approved FFmpeg action after a stall. The cache-resumable
  // backend picks up where it halted, so this reuses the stored plan and approval
  // through the direct run-start binding rather than forking a second run path.
  async function retryFfmpegBuildPlan() {
    const last = pFfmpegRunLast.current;
    if (!last) return;
    setFfmpegLogEntries([]); setFfmpegStalledAddresses([]); setApprovedActionStatus("starting");
    try {
      await LFFmpegCompilationLaunch(last.plan, last.approval, false);
      setApprovedActionPhase("ffmpeg");
      setActiveTabId("buildFfmpeg");
    } catch (err) {
      setApprovedActionStatus("failed");
      setFfmpegLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function cancelApprovedAction() { await LActionApprovedCancel(); }
  // Clear a finished-but-unsuccessful run (failed or cancelled) so the user can
  // discard the dead plan and its logs, returning the page to its idle state.
  function LActionApprovedClear() {
    setToolchainLogEntries([]); setFfmpegLogEntries([]); setFfmpegStalledAddresses([]);
    setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null);
    setBuildResult(null); setBuildResultError("");
    setApprovedActionPhase(null); setApprovedActionStatus("idle");
  }
  async function openInUserBrowser(url: string) { BrowserOpenURL(url); }

  function LLibraryToggle(libraryId: string) {
    const library = libraryCatalog.find((l) => l.libraryId === libraryId);
    if (library?.locked) return;
    const current = ffmpegBuildSettings.selectedLibraryIds;
    const removing = current.includes(libraryId);
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    let next = removing ? current.filter((id) => id !== libraryId) : [...current, libraryId];
    if (!removing) next = LLibraryExclusiveRemove(next, libraryId);
    next = LLibrarySelectionNormalize(next, profile, libraryCatalog);
    setLibraryPresetId(LPresetLibraryMatch(next, libraryPresetCatalog, profile, libraryCatalog, extendedLibraries));
    LSettingsFFmpegUpdate({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryGet(next, libraryCatalog, profile) });
  }

  function LPresetLibraryApply(presetId: LPresetLibraryId) {
    const preset = libraryPresetCatalog.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const next = preset.dev
      ? LLibraryTestGet(libraryCatalog, profile)
      : LLibrarySelectionNormalize(LPresetLibraryResolve(preset, extendedLibraries), profile, libraryCatalog);
    setLibraryPresetId(presetId);
    LSettingsFFmpegUpdate({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryGet(next, libraryCatalog, profile) });
  }

  // Toggling the Extended mode re-applies the active named preset under the new mode so
  // its source-build extras are added/removed and the highlighted button stays truthful.
  // Custom/dev selections keep their libraries and are just re-matched against the new mode.
  function LLibraryExtendedUpdate(next: boolean) {
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const preset = libraryPresetCatalog.find((p) => p.presetId === libraryPresetId);
    if (preset && preset.presetId !== "custom" && !preset.dev) {
      const nextIds = LLibrarySelectionNormalize(LPresetLibraryResolve(preset, next), profile, libraryCatalog);
      LSettingsFFmpegUpdate({ selectedLibraryIds: nextIds, licenseProfileName: LLicenseBoundaryGet(nextIds, libraryCatalog, profile) });
    } else {
      setLibraryPresetId(LPresetLibraryMatch(ffmpegBuildSettings.selectedLibraryIds, libraryPresetCatalog, profile, libraryCatalog, next));
    }
    setExtendedLibrariesState(next);
  }

  function LOptionToggle(optionId: string) {
    const option = initialProgramState.defaultConfigureOptionCatalog.find((o) => o.optionId === optionId);
    if (option?.locked) return;
    const current = ffmpegBuildSettings.selectedConfigureOptionIds;
    LSettingsFFmpegUpdate({ selectedConfigureOptionIds: current.includes(optionId) ? current.filter((id) => id !== optionId) : [...current, optionId] });
  }

  function LPresetOptionApply(presetId: LPresetOptionId) {
    const preset = LPresetOptionCatalog.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    LSettingsFFmpegUpdate({ selectedConfigureOptionIds: preset.optionIds });
  }

  function LPackageToolchainRestore() {
    const recommended = initialProgramState.defaultBuildConfigSettings.msys2PackageNames.join("\n");
    setMsys2PackageText(LPackagePrefixUpdate(recommended, buildConfigSettings.windowsShellProfileName));
    setToolchainPreparationPlanReview(null);
  }

  function LFlagExtraRestore() {
    const flags = initialProgramState.defaultFfmpegBuildSettings.extraConfigureFlags || [];
    const optIds = initialProgramState.defaultFfmpegBuildSettings.selectedConfigureOptionIds || ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"];
    setExtraConfigureFlagText(flags.join("\n"));
    LSettingsFFmpegUpdate({ selectedConfigureOptionIds: optIds, extraConfigureFlags: flags, configureFlags: flags });
  }

  function LMSYSPackageUpdate(text: string) { setMsys2PackageText(text); setToolchainPreparationPlanReview(null); }
  function LFlagExtraUpdate(text: string) { setExtraConfigureFlagText(text); setFfmpegBuildPlanReview(null); }

  return {
    tabPanelRef,
    activeTabId, setActiveTabId,
    initialProgramState,
    libraryCatalog,
    libraryPresetCatalog,
    buildConfigSettings,
    ffmpegBuildSettings,
    libraryPresetId,
    extendedLibraries,
    libraryDetailedView, setLibraryDetailedView,
    optionsDetailedView, setOptionsDetailedView,
    libraryTechnicalDetails, setLibraryTechnicalDetails,
    optionsTechnicalDetails, setOptionsTechnicalDetails,
    librarySectionFilters, setLLibrarySectionFilters,
    msys2PackageText,
    extraConfigureFlagText,
    toolchainPreparationPlanReview,
    ffmpegBuildPlanReview,
    approvedActionStatus,
    approvedActionPhase,
    ffmpegStalledAddresses,
    toolchainLogEntries,
    ffmpegLogEntries,
    localLogRecords, localLogRecordsError,
    buildResult, buildResultError, isLoadingBuildResult,
    buildVerification, buildVerificationError, isVerifyingBuild, verifyBuildResult,
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    configuredMsys2PackageNames,
    canCancelToolchain, canCancelFfmpeg,
    isApprovedActionRunning: canCancelApprovedAction,
    toolchainProgress, ffmpegProgress,
    securityLogEntries,
    LSettingsToolchainUpdate, LSettingsFFmpegUpdate, LMSYSArchiveUpdate, LProfileShellUpdate,
    chooseWorkspaceDirectory,
    addBuildConfigPlanAndContinueToPrep, reviewFfmpegPlans,
    approveToolchainPreparationPlan, approveFfmpegBuildPlan, retryFfmpegBuildPlan, LPlanToolchainCancel, cancelApprovedAction, LActionApprovedClear,
    openInUserBrowser,
    LLibraryToggle, LPresetLibraryApply, LLibraryExtendedUpdate, LOptionToggle, LPresetOptionApply,
    LPackageToolchainRestore, LFlagExtraRestore,
    LMSYSPackageUpdate, LFlagExtraUpdate,
    refreshBuildResult, openResultFolder, openResultReport,
    refreshLocalLogRecords, loadLocalLogRecord, openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  };
}
