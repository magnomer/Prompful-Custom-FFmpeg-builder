import { useEffect, useMemo, useRef, useState } from "react";
import {
  LPlanFFmpegApprove,
  LPlanToolchainApprove,
  LActionApprovedCancel,
  LToolchainEnvironmentClear,
  LResultBuildGet,
  LVerificationBuildRun,
  LStateInitialGet,
  LStatusToolchainGet,
  LProfileToolchainList,
  LRecordLogGet,
  LRecordLogList,
  LFolderRecordOpen,
  LFileRecordOpen,
  LFolderLogOpen,
  LToolchainInstallVerify,
  LCatalogSourceGet,
  LCatalogPresetSourceGet,
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

import { LLogSecurityEntry, LLogSecurityPayload, LStatusActionPayload, LProgressLive, LProgressCompute } from "./tabs/logutils";
import {
  LPresetLibraryId, LPresetLibrary, LPresetLibraryListSanitize, LPresetLibraryResolve,
  LLibrarySelectionNormalize, LPresetLibraryMatch, LPresetLibraryValidate,
  LLicenseBoundaryDerive, LLibraryExclusiveRemove,
  LLibraryMaximumTestIdsGet,
} from "./tabs/libraries";
import { LPresetOptionId, LPresetOptionList } from "./tabs/options";
import {
  PTabId, LStateUiSaved, LStateWindowSaved, LStateWindowKey,
  LSettingsBuildEmpty, LSettingsFFmpegEmpty, LStateInitialDefault,
  LTextLineSplit, LLogLevelNormalize, LTabIdValidate,
  LRequestApprovalCreate, LStateUiParse, LStateWindowRead,
  LPackagePrefixRemap, LSettingsBuildNormalize, LSettingsFFmpegNormalize, LStateInitialNormalize,
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
  const [activeTabId, setActiveTabId] = useState<PTabId>("source");
  const hasLoadedSavedState = useRef(false);
  const tabPanelRef = useRef<HTMLElement>(null);
  // Tabs whose scroll position is remembered when leaving and restored when
  // returning (within the session). Every other tab still scrolls back to top.
  const scrollRememberedTabIds = useRef(new Set<PTabId>(["options", "buildFfmpeg"]));
  const tabScrollPositions = useRef<Partial<Record<PTabId, number>>>({});
  const activeTabIdRef = useRef<PTabId>("source");
  const [initialProgramState, setInitialProgramState] = useState<LStateInitial>(LStateInitialDefault);
  const [libraryCatalog, setLibraryCatalog] = useState<LLibraryChoice[]>([]);
  const [libraryPresetCatalog, setLibraryPresetCatalog] = useState<LPresetLibrary[]>([]);
  const [buildConfigSettings, setBuildConfigSettings] = useState<LSettingsBuild>(LSettingsBuildEmpty);
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
  const [toolchainLogEntries, setToolchainLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [ffmpegLogEntries, setFfmpegLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [localLogRecords, setLocalLogRecords] = useState<LRecordLog[]>([]);
  const [localLogRecordsError, setLocalLogRecordsError] = useState("");
  const [buildResult, setBuildResult] = useState<LResultBuild | null>(null);
  const [buildResultError, setBuildResultError] = useState("");
  const [isLoadingBuildResult, setIsLoadingBuildResult] = useState(false);
  const [buildVerification, setBuildVerification] = useState<LVerificationBuild | null>(null);
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

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";
  const toolchainProgress = useMemo<LProgressLive>(() => LProgressCompute(toolchainLogEntries, approvedActionStatus, "toolchain"), [toolchainLogEntries, approvedActionStatus]);
  const ffmpegProgress = useMemo<LProgressLive>(() => LProgressCompute(ffmpegLogEntries, approvedActionStatus, "ffmpeg"), [ffmpegLogEntries, approvedActionStatus]);
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
      const initialPresetCatalog = LPresetLibraryListSanitize(nextState.defaultLibraryPresetCatalog);
      const hasSavedPreset = LPresetLibraryValidate(saved.libraryPresetId);
      const resolvedPresetId: LPresetLibraryId = hasSavedPreset ? saved.libraryPresetId! : "default";
      if (!hasSavedPreset && !saved.ffmpegBuildSettings) {
        const defaultPreset = initialPresetCatalog.find((p) => p.presetId === "default");
        if (defaultPreset) {
          const nextIds = LLibrarySelectionNormalize(defaultPreset.libraryIds, resolvedFbs.windowsShellProfileName, nextState.defaultLibraryCatalog);
          resolvedFbs = { ...resolvedFbs, selectedLibraryIds: nextIds, licenseProfileName: LLicenseBoundaryDerive(nextIds, nextState.defaultLibraryCatalog, resolvedFbs.windowsShellProfileName) };
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
    });
    return () => { removeLogListener(); removeStatusListener(); window.removeEventListener("customffmpeg-locale-change", onLocaleChange); };
  }, []);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    let isCurrent = true;
    Promise.all([
      LCatalogSourceGet(ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName),
      LCatalogPresetSourceGet(ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName),
    ])
      .then(([LCatalogLibrarySource, LCatalogPresetSource]) => {
        if (!isCurrent) return;
        const nextLibraryCatalog = Array.isArray(LCatalogLibrarySource) ? LCatalogLibrarySource : [];
        const nextPresetCatalog = LPresetLibraryListSanitize(LCatalogPresetSource);
        setLibraryCatalog(nextLibraryCatalog);
        setLibraryPresetCatalog(nextPresetCatalog);
        setFfmpegBuildSettings((settings) => {
          const activePreset = nextPresetCatalog.find((preset) => preset.presetId === libraryPresetIdRef.current && preset.presetId !== "custom");
          const nextLibraryIds = activePreset
            ? (activePreset.dev
              ? LLibraryMaximumTestIdsGet(nextLibraryCatalog, settings.windowsShellProfileName)
              : LLibrarySelectionNormalize(LPresetLibraryResolve(activePreset, extendedLibraries), settings.windowsShellProfileName, nextLibraryCatalog))
            : LLibrarySelectionNormalize(settings.selectedLibraryIds, settings.windowsShellProfileName, nextLibraryCatalog);
          setLibraryPresetId(LPresetLibraryMatch(nextLibraryIds, nextPresetCatalog, settings.windowsShellProfileName, nextLibraryCatalog, extendedLibraries, libraryPresetIdRef.current));
          return {
            ...settings,
            selectedLibraryIds: nextLibraryIds,
            licenseProfileName: LLicenseBoundaryDerive(nextLibraryIds, nextLibraryCatalog, settings.windowsShellProfileName),
          };
        });
        setFfmpegBuildPlanReview(null);
      })
      .catch(() => {
        if (isCurrent) { setLibraryCatalog(initialProgramState.defaultLibraryCatalog); setLibraryPresetCatalog(LPresetLibraryListSanitize(initialProgramState.defaultLibraryPresetCatalog)); }
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
    const onScroll = () => { tabScrollPositions.current[activeTabIdRef.current] = el.scrollTop; };
    el.addEventListener("scroll", onScroll, { passive: true });
    return () => el.removeEventListener("scroll", onScroll);
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
      const records = await LRecordLogList(dir);
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
    try { setInstalledToolchainProfiles(await LProfileToolchainList(dir)); }
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

  function updateBuildConfigSettings(next: Partial<LSettingsBuild>) {
    setBuildConfigSettings((s) => ({ ...s, ...next }));
    setToolchainPreparationPlanReview(null);
  }

  function updateFfmpegBuildSettings(next: Partial<LSettingsFFmpeg>) {
    setFfmpegBuildSettings((s) => ({ ...s, ...next }));
    setFfmpegBuildPlanReview(null);
  }

  // Changing the shell profile must remap the MSYS2 package prefixes in the
  // toolchain textarea too, otherwise a non-ucrt64 profile would still install
  // ucrt64 toolchain packages (wrong compiler/runtime for the chosen prefix).
  function changeShellProfile(profileName: string) {
    // Re-normalize the library selection for the new profile so libraries with no
    // package there (e.g. onnxruntime on mingw64) are dropped, and recompute the
    // preset/license to match.
    const nextLibraryIds = LLibrarySelectionNormalize(ffmpegBuildSettings.selectedLibraryIds, profileName, libraryCatalog);
    updateBuildConfigSettings({ windowsShellProfileName: profileName });
    updateFfmpegBuildSettings({
      windowsShellProfileName: profileName,
      selectedLibraryIds: nextLibraryIds,
      licenseProfileName: LLicenseBoundaryDerive(nextLibraryIds, libraryCatalog, profileName),
    });
    setLibraryPresetId(LPresetLibraryMatch(nextLibraryIds, libraryPresetCatalog, profileName, libraryCatalog, extendedLibraries, libraryPresetId));
    setMsys2PackageText((text) => LPackagePrefixRemap(text, profileName));
  }

  function updateMsys2ArchiveUrl(nextUrl: string) {
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
    updateBuildConfigSettings({ workspaceDirectory: dir });
    updateFfmpegBuildSettings({ workspaceDirectory: dir });
  }

  async function addBuildConfigPlanAndContinueToPrep() {
    const review = await LPlanToolchainRequest({ ...buildConfigSettings, msys2PackageNames: LTextLineSplit(msys2PackageText) });
    setToolchainPreparationPlanReview(review);
    setActiveTabId("prep");
  }

  function cancelToolchainPreparationPlan() {
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
    setFfmpegLogEntries([]); setApprovedActionStatus("starting");
    try {
      await LPlanFFmpegApprove(r.reviewSessionId, LRequestApprovalCreate(r.plan.actionName, r.plan.planHash, r.expectedLConsentText));
      setApprovedActionPhase("ffmpeg");
      setFfmpegBuildPlanReview(null); setActiveTabId("buildFfmpeg");
    } catch (err) {
      setApprovedActionStatus("failed");
      setFfmpegLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function cancelApprovedAction() { await LActionApprovedCancel(); }
  // Clear a finished-but-unsuccessful run (failed or cancelled) so the user can
  // discard the dead plan and its logs, returning the page to its idle state.
  function clearApprovedAction() {
    setToolchainLogEntries([]); setFfmpegLogEntries([]);
    setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null);
    setBuildResult(null); setBuildResultError("");
    setApprovedActionPhase(null); setApprovedActionStatus("idle");
  }
  async function openInUserBrowser(url: string) { BrowserOpenURL(url); }

  function toggleLibrary(libraryId: string) {
    const library = libraryCatalog.find((l) => l.libraryId === libraryId);
    if (library?.locked) return;
    const current = ffmpegBuildSettings.selectedLibraryIds;
    const removing = current.includes(libraryId);
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    let next = removing ? current.filter((id) => id !== libraryId) : [...current, libraryId];
    if (!removing) next = LLibraryExclusiveRemove(next, libraryId);
    next = LLibrarySelectionNormalize(next, profile, libraryCatalog);
    setLibraryPresetId(LPresetLibraryMatch(next, libraryPresetCatalog, profile, libraryCatalog, extendedLibraries));
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryDerive(next, libraryCatalog, profile) });
  }

  function applyLibraryPreset(presetId: LPresetLibraryId) {
    const preset = libraryPresetCatalog.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const next = preset.dev
      ? LLibraryMaximumTestIdsGet(libraryCatalog, profile)
      : LLibrarySelectionNormalize(LPresetLibraryResolve(preset, extendedLibraries), profile, libraryCatalog);
    setLibraryPresetId(presetId);
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryDerive(next, libraryCatalog, profile) });
  }

  // Toggling the Extended mode re-applies the active named preset under the new mode so
  // its source-build extras are added/removed and the highlighted button stays truthful.
  // Custom/dev selections keep their libraries and are just re-matched against the new mode.
  function setExtendedLibraries(next: boolean) {
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const preset = libraryPresetCatalog.find((p) => p.presetId === libraryPresetId);
    if (preset && preset.presetId !== "custom" && !preset.dev) {
      const nextIds = LLibrarySelectionNormalize(LPresetLibraryResolve(preset, next), profile, libraryCatalog);
      updateFfmpegBuildSettings({ selectedLibraryIds: nextIds, licenseProfileName: LLicenseBoundaryDerive(nextIds, libraryCatalog, profile) });
    } else {
      setLibraryPresetId(LPresetLibraryMatch(ffmpegBuildSettings.selectedLibraryIds, libraryPresetCatalog, profile, libraryCatalog, next));
    }
    setExtendedLibrariesState(next);
  }

  function toggleConfigureOption(optionId: string) {
    const option = initialProgramState.defaultConfigureOptionCatalog.find((o) => o.optionId === optionId);
    if (option?.locked) return;
    const current = ffmpegBuildSettings.selectedConfigureOptionIds;
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: current.includes(optionId) ? current.filter((id) => id !== optionId) : [...current, optionId] });
  }

  function applyOptionPreset(presetId: LPresetOptionId) {
    const preset = LPresetOptionList.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: preset.optionIds });
  }

  function restoreRecommendedToolchainPackages() {
    const recommended = initialProgramState.defaultBuildConfigSettings.msys2PackageNames.join("\n");
    setMsys2PackageText(LPackagePrefixRemap(recommended, buildConfigSettings.windowsShellProfileName));
    setToolchainPreparationPlanReview(null);
  }

  function restoreRecommendedExtraFlags() {
    const flags = initialProgramState.defaultFfmpegBuildSettings.extraConfigureFlags || [];
    const optIds = initialProgramState.defaultFfmpegBuildSettings.selectedConfigureOptionIds || ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"];
    setExtraConfigureFlagText(flags.join("\n"));
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: optIds, extraConfigureFlags: flags, configureFlags: flags });
  }

  function handleMsys2PackageTextChange(text: string) { setMsys2PackageText(text); setToolchainPreparationPlanReview(null); }
  function handleExtraFlagTextChange(text: string) { setExtraConfigureFlagText(text); setFfmpegBuildPlanReview(null); }

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
    updateBuildConfigSettings, updateFfmpegBuildSettings, updateMsys2ArchiveUrl, changeShellProfile,
    chooseWorkspaceDirectory,
    addBuildConfigPlanAndContinueToPrep, reviewFfmpegPlans,
    approveToolchainPreparationPlan, approveFfmpegBuildPlan, cancelToolchainPreparationPlan, cancelApprovedAction, clearApprovedAction,
    openInUserBrowser,
    toggleLibrary, applyLibraryPreset, setExtendedLibraries, toggleConfigureOption, applyOptionPreset,
    restoreRecommendedToolchainPackages, restoreRecommendedExtraFlags,
    handleMsys2PackageTextChange, handleExtraFlagTextChange,
    refreshBuildResult, openResultFolder, openResultReport,
    refreshLocalLogRecords, loadLocalLogRecord, openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  };
}
