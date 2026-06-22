import { useEffect, useMemo, useRef, useState } from "react";
import {
  ApproveFfmpegBuildPlan,
  ApproveToolchainPreparationPlan,
  CancelApprovedAction,
  GetBuildResult,
  GetInitialApplicationState,
  LoadUiState,
  SaveUiState,
  RequestFfmpegBuildPlan,
  RequestToolchainPreparationPlan,
  OpenResultFolder,
  SelectWorkspace,
} from "../wailsjs/go/app/App";
import { BrowserOpenURL, EventsOn, WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../wailsjs/runtime/runtime";

import { SecurityLogEntry, SecurityLogPayload, ApprovedActionStatusPayload, LiveProgress, computeProgress } from "./tabs/logutils";
import {
  LibraryPresetId, libraryPresets,
  normalizeLibrarySelection, matchLibraryPresetId, isValidLibraryPresetId,
  deriveLicenseBoundaryFromSelectedLibraries, removeMutuallyExclusiveLibraries,
} from "./tabs/libraries";
import { OptionPresetId, optionPresets } from "./tabs/options";
import {
  TabId, SavedUiState, SavedWindowState, savedWindowStateKey,
  emptyBuildToolSettings, emptyFfmpegBuildSettings, defaultInitialApplicationState,
  splitLines, normalizeLogLevel, isValidTabId,
  createApprovalRequest, parseSavedUiState, readSavedWindowState,
} from "./appstate";
import { t } from "./i18n";

// ─── Window state helpers ─────────────────────────────────────────────────────

function readRuntimeNumber(value: unknown, tupleIndex: number, primary: string, fallback: string): number {
  if (Array.isArray(value)) return Number(value[tupleIndex]);
  if (value && typeof value === "object") {
    const r = value as Record<string, unknown>;
    return Number(r[primary] ?? r[fallback] ?? 0);
  }
  return 0;
}

function restoreWindowState() {
  const s = readSavedWindowState();
  if (Number.isFinite(s.width) && Number.isFinite(s.height)) WindowSetSize(Number(s.width), Number(s.height));
  if (Number.isFinite(s.x) && Number.isFinite(s.y)) WindowSetPosition(Number(s.x), Number(s.y));
}

async function saveWindowState() {
  try {
    const sz = await WindowGetSize();
    const pos = await WindowGetPosition();
    const width  = readRuntimeNumber(sz,  0, "w", "width");
    const height = readRuntimeNumber(sz,  1, "h", "height");
    const x = readRuntimeNumber(pos, 0, "x", "left");
    const y = readRuntimeNumber(pos, 1, "y", "top");
    window.localStorage.setItem(savedWindowStateKey, JSON.stringify({ width, height, x, y } satisfies SavedWindowState));
  } catch {
    // Window persistence is best-effort. Never block the app if the runtime is unavailable.
  }
}

// ─── Hook ────────────────────────────────────────────────────────────────────

export function useBuilderState() {
  const [activeTabId, setActiveTabId] = useState<TabId>("source");
  const hasLoadedSavedState = useRef(false);
  const tabPanelRef = useRef<HTMLElement>(null);
  const [initialApplicationState, setInitialApplicationState] = useState<InitialApplicationState>(defaultInitialApplicationState);
  const [buildToolSettings, setBuildToolSettings] = useState<BuildToolSettings>(emptyBuildToolSettings);
  const [ffmpegBuildSettings, setFfmpegBuildSettings] = useState<FfmpegBuildSettings>(emptyFfmpegBuildSettings);
  const [libraryPresetId, setLibraryPresetId] = useState<LibraryPresetId>("default");
  const [msys2PackageText, setMsys2PackageText] = useState("");
  const [extraConfigureFlagText, setExtraConfigureFlagText] = useState("");
  const [toolchainPreparationPlanReview, setToolchainPreparationPlanReview] = useState<ToolchainPreparationPlanReview | null>(null);
  const [ffmpegBuildPlanReview, setFfmpegBuildPlanReview] = useState<FfmpegBuildPlanReview | null>(null);
  const [approvedActionStatus, setApprovedActionStatus] = useState("idle");
  const [approvedActionPhase, setApprovedActionPhase] = useState<"toolchain" | "ffmpeg" | null>(null);
  const [toolchainLogEntries, setToolchainLogEntries] = useState<SecurityLogEntry[]>([]);
  const [ffmpegLogEntries, setFfmpegLogEntries] = useState<SecurityLogEntry[]>([]);
  const [buildResult, setBuildResult] = useState<BuildResult | null>(null);
  const [buildResultError, setBuildResultError] = useState("");

  const approvedActionPhaseRef = useRef<"toolchain" | "ffmpeg" | null>(null);
  approvedActionPhaseRef.current = approvedActionPhase;

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";
  const toolchainProgress = useMemo<LiveProgress>(() => computeProgress(toolchainLogEntries, approvedActionStatus, "toolchain"), [toolchainLogEntries, approvedActionStatus]);
  const ffmpegProgress = useMemo<LiveProgress>(() => computeProgress(ffmpegLogEntries, approvedActionStatus, "ffmpeg"), [ffmpegLogEntries, approvedActionStatus]);
  const securityLogEntries = useMemo(() => [...toolchainLogEntries, ...ffmpegLogEntries], [toolchainLogEntries, ffmpegLogEntries]);

  useEffect(() => {
    GetInitialApplicationState().then(async (nextState: InitialApplicationState) => {
      const saved = parseSavedUiState(await LoadUiState());
      const savedBts = saved.buildToolSettings ? { ...nextState.defaultBuildToolSettings, ...saved.buildToolSettings } : nextState.defaultBuildToolSettings;
      let resolvedFbs = saved.ffmpegBuildSettings ? { ...nextState.defaultFfmpegBuildSettings, ...saved.ffmpegBuildSettings } : nextState.defaultFfmpegBuildSettings;
      const hasSavedPreset = isValidLibraryPresetId(saved.libraryPresetId);
      const resolvedPresetId: LibraryPresetId = hasSavedPreset ? saved.libraryPresetId! : "default";
      if (!hasSavedPreset && !saved.ffmpegBuildSettings) {
        const defaultPreset = libraryPresets.find((p) => p.presetId === "default");
        if (defaultPreset) {
          const nextIds = normalizeLibrarySelection(defaultPreset.libraryIds);
          resolvedFbs = { ...resolvedFbs, selectedLibraryIds: nextIds, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextIds, nextState.defaultLibraryCatalog) };
        }
      }
      setInitialApplicationState(nextState);
      setBuildToolSettings(savedBts);
      setFfmpegBuildSettings(resolvedFbs);
      setMsys2PackageText(saved.msys2PackageText ?? savedBts.msys2PackageNames.join("\n"));
      setExtraConfigureFlagText(saved.extraConfigureFlagText ?? resolvedFbs.extraConfigureFlags.join("\n"));
      setLibraryPresetId(resolvedPresetId);
      if (isValidTabId(saved.activeTabId)) setActiveTabId(saved.activeTabId);
      hasLoadedSavedState.current = true;
      restoreWindowState();
    });

    const makeEntry = (payload: SecurityLogPayload): SecurityLogEntry => ({
      level: normalizeLogLevel(payload.level),
      message: payload.message,
      timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }),
    });

    const removeLogListener = EventsOn("security-log", (payload: SecurityLogPayload) => {
      const entry = makeEntry(payload);
      if (approvedActionPhaseRef.current === "ffmpeg") setFfmpegLogEntries((prev) => [...prev, entry]);
      else setToolchainLogEntries((prev) => [...prev, entry]);
    });
    const removeStatusListener = EventsOn("approved-action-status", (payload: ApprovedActionStatusPayload) => {
      setApprovedActionStatus(payload.status);
      if (payload.status === "failed") { setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null); setBuildResult(null); }
      if (payload.status === "completed") { setBuildResult(null); setApprovedActionPhase(null); }
    });
    return () => { removeLogListener(); removeStatusListener(); };
  }, []);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    void SaveUiState(JSON.stringify({
      activeTabId,
      buildToolSettings: { ...buildToolSettings, msys2PackageNames: splitLines(msys2PackageText) },
      ffmpegBuildSettings: { ...ffmpegBuildSettings, extraConfigureFlags: splitLines(extraConfigureFlagText), configureFlags: splitLines(extraConfigureFlagText) },
      msys2PackageText, extraConfigureFlagText, libraryPresetId,
    } satisfies SavedUiState));
  }, [activeTabId, buildToolSettings, ffmpegBuildSettings, msys2PackageText, extraConfigureFlagText, libraryPresetId]);

  useEffect(() => {
    const id = window.setInterval(() => { saveWindowState(); }, 2000);
    window.addEventListener("beforeunload", saveWindowState);
    return () => { window.clearInterval(id); window.removeEventListener("beforeunload", saveWindowState); saveWindowState(); };
  }, []);

  useEffect(() => {
    if (activeTabId === "result") refreshBuildResult();
  }, [activeTabId, buildToolSettings.workspaceDirectory]);

  useEffect(() => {
    if (tabPanelRef.current) tabPanelRef.current.scrollTop = 0;
  }, [activeTabId]);

  async function refreshBuildResult() {
    const dir = buildToolSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResult(null); setBuildResultError(t("result.error.chooseWorkspaceFirst")); return; }
    try { const r = await GetBuildResult(dir); setBuildResult(r); setBuildResultError(""); }
    catch (err) { setBuildResult(null); setBuildResultError(err instanceof Error ? err.message : String(err)); }
  }

  async function openResultFolder() {
    const dir = buildToolSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResultError(t("result.error.chooseWorkspaceFirst")); return; }
    await OpenResultFolder(dir);
  }

  function updateBuildToolSettings(next: Partial<BuildToolSettings>) {
    setBuildToolSettings((s) => ({ ...s, ...next }));
    setToolchainPreparationPlanReview(null);
  }

  function updateFfmpegBuildSettings(next: Partial<FfmpegBuildSettings>) {
    setFfmpegBuildSettings((s) => ({ ...s, ...next }));
    setFfmpegBuildPlanReview(null);
  }

  function updateMsys2ArchiveUrl(nextUrl: string) {
    setBuildToolSettings((s) => {
      const oldAuto = s.msys2ArchiveUrl ? `${s.msys2ArchiveUrl}.sig` : "";
      const shouldUpdate = s.msys2ArchiveSignatureUrl === "" || s.msys2ArchiveSignatureUrl === oldAuto;
      return { ...s, msys2ArchiveUrl: nextUrl, msys2ArchiveSignatureUrl: shouldUpdate && nextUrl ? `${nextUrl}.sig` : s.msys2ArchiveSignatureUrl };
    });
    setToolchainPreparationPlanReview(null);
  }

  async function chooseWorkspaceDirectory() {
    const dir = await SelectWorkspace();
    if (!dir) return;
    updateBuildToolSettings({ workspaceDirectory: dir });
    updateFfmpegBuildSettings({ workspaceDirectory: dir });
  }

  async function addBuildToolsPlanAndContinueToPrep() {
    const review = await RequestToolchainPreparationPlan({ ...buildToolSettings, msys2PackageNames: splitLines(msys2PackageText) });
    setToolchainPreparationPlanReview(review);
    setActiveTabId("prep");
  }

  async function reviewFfmpegPlans() {
    const flags = splitLines(extraConfigureFlagText);
    const review = await RequestFfmpegBuildPlan({ ...ffmpegBuildSettings, extraConfigureFlags: flags, configureFlags: flags });
    setFfmpegBuildPlanReview(review);
    setActiveTabId("buildFfmpeg");
  }

  async function approveToolchainPreparationPlan() {
    if (!toolchainPreparationPlanReview) return;
    const r = toolchainPreparationPlanReview;
    setToolchainLogEntries([]); setApprovedActionPhase("toolchain"); setApprovedActionStatus("starting");
    try {
      await ApproveToolchainPreparationPlan(r.reviewSessionId, createApprovalRequest(r.plan.actionName, r.plan.planHash, r.expectedConsentText));
      setToolchainPreparationPlanReview(null); setActiveTabId("prep");
    } catch (err) {
      setApprovedActionStatus("failed");
      setToolchainLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function approveFfmpegBuildPlan() {
    if (!ffmpegBuildPlanReview) return;
    const r = ffmpegBuildPlanReview;
    setFfmpegLogEntries([]); setApprovedActionPhase("ffmpeg"); setApprovedActionStatus("starting");
    try {
      await ApproveFfmpegBuildPlan(r.reviewSessionId, createApprovalRequest(r.plan.actionName, r.plan.planHash, r.expectedConsentText));
      setFfmpegBuildPlanReview(null); setActiveTabId("buildFfmpeg");
    } catch (err) {
      setApprovedActionStatus("failed");
      setFfmpegLogEntries((prev) => [...prev, { level: "error", message: err instanceof Error ? err.message : String(err), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function cancelApprovedAction() { await CancelApprovedAction(); }
  async function openInUserBrowser(url: string) { BrowserOpenURL(url); }

  function toggleLibrary(libraryId: string) {
    const library = initialApplicationState.defaultLibraryCatalog.find((l) => l.libraryId === libraryId);
    if (library?.locked) return;
    const current = ffmpegBuildSettings.selectedLibraryIds;
    const removing = current.includes(libraryId);
    let next = removing ? current.filter((id) => id !== libraryId) : [...current, libraryId];
    if (!removing) next = removeMutuallyExclusiveLibraries(next, libraryId);
    next = normalizeLibrarySelection(next);
    setLibraryPresetId(matchLibraryPresetId(next));
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(next, initialApplicationState.defaultLibraryCatalog) });
  }

  function applyLibraryPreset(presetId: LibraryPresetId) {
    const preset = libraryPresets.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    const next = normalizeLibrarySelection(preset.libraryIds);
    setLibraryPresetId(presetId);
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(next, initialApplicationState.defaultLibraryCatalog) });
  }

  function toggleConfigureOption(optionId: string) {
    const option = initialApplicationState.defaultConfigureOptionCatalog.find((o) => o.optionId === optionId);
    if (option?.locked) return;
    const current = ffmpegBuildSettings.selectedConfigureOptionIds;
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: current.includes(optionId) ? current.filter((id) => id !== optionId) : [...current, optionId] });
  }

  function applyOptionPreset(presetId: OptionPresetId) {
    const preset = optionPresets.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: preset.optionIds });
  }

  function restoreRecommendedToolchainPackages() {
    setMsys2PackageText(initialApplicationState.defaultBuildToolSettings.msys2PackageNames.join("\n"));
    setToolchainPreparationPlanReview(null);
  }

  function restoreRecommendedExtraFlags() {
    const flags = initialApplicationState.defaultFfmpegBuildSettings.extraConfigureFlags || [];
    const optIds = initialApplicationState.defaultFfmpegBuildSettings.selectedConfigureOptionIds || ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"];
    setExtraConfigureFlagText(flags.join("\n"));
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: optIds, extraConfigureFlags: flags, configureFlags: flags });
  }

  function handleMsys2PackageTextChange(text: string) { setMsys2PackageText(text); setToolchainPreparationPlanReview(null); }
  function handleExtraFlagTextChange(text: string) { setExtraConfigureFlagText(text); setFfmpegBuildPlanReview(null); }

  return {
    tabPanelRef,
    activeTabId, setActiveTabId,
    initialApplicationState,
    buildToolSettings,
    ffmpegBuildSettings,
    libraryPresetId,
    msys2PackageText,
    extraConfigureFlagText,
    toolchainPreparationPlanReview,
    ffmpegBuildPlanReview,
    approvedActionStatus,
    approvedActionPhase,
    toolchainLogEntries,
    ffmpegLogEntries,
    buildResult, buildResultError,
    canCancelToolchain, canCancelFfmpeg,
    toolchainProgress, ffmpegProgress,
    securityLogEntries,
    updateBuildToolSettings, updateFfmpegBuildSettings, updateMsys2ArchiveUrl,
    chooseWorkspaceDirectory,
    addBuildToolsPlanAndContinueToPrep, reviewFfmpegPlans,
    approveToolchainPreparationPlan, approveFfmpegBuildPlan, cancelApprovedAction,
    openInUserBrowser,
    toggleLibrary, applyLibraryPreset, toggleConfigureOption, applyOptionPreset,
    restoreRecommendedToolchainPackages, restoreRecommendedExtraFlags,
    handleMsys2PackageTextChange, handleExtraFlagTextChange,
    refreshBuildResult, openResultFolder,
  };
}
