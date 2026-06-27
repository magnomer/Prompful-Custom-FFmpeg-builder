import { useEffect, useMemo, useRef, useState } from "react";
import {
  ApproveFfmpegBuildPlan,
  ApproveToolchainPreparationPlan,
  CancelApprovedAction,
  ClearBuildEnvironments,
  GetBuildResult,
  GetInitialApplicationState,
  GetToolchainStatus,
  GetInstalledToolchainProfiles,
  VerifyToolchainInstallation,
  LoadUiState,
  SaveUiState,
  RequestFfmpegBuildPlan,
  RequestToolchainPreparationPlan,
  OpenResultFolder,
  OpenResultReport,
  SelectWorkspace,
  SetLocale,
} from "../wailsjs/go/app/App";
import { BrowserOpenURL, EventsOn, WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../wailsjs/runtime/runtime";

import { SecurityLogEntry, SecurityLogPayload, ApprovedActionStatusPayload, LiveProgress, computeProgress } from "./tabs/logutils";
import {
  LibraryPresetId, libraryPresets, presetLibraryIds,
  normalizeLibrarySelection, matchLibraryPresetId, isValidLibraryPresetId,
  deriveLicenseBoundaryFromSelectedLibraries, removeMutuallyExclusiveLibraries,
  maximumTestLibraryIds,
} from "./tabs/libraries";
import { OptionPresetId, optionPresets } from "./tabs/options";
import {
  TabId, SavedUiState, SavedWindowState, savedWindowStateKey,
  emptyBuildConfigSettings, emptyFfmpegBuildSettings, defaultInitialApplicationState,
  splitLines, normalizeLogLevel, isValidTabId,
  createApprovalRequest, parseSavedUiState, readSavedWindowState,
  remapMsys2PackagePrefixes,
} from "./appstate";
import { t, getLocale } from "./i18n";

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
  // Tabs whose scroll position is remembered when leaving and restored when
  // returning (within the session). Every other tab still scrolls back to top.
  const scrollRememberedTabIds = useRef(new Set<TabId>(["options", "buildFfmpeg"]));
  const tabScrollPositions = useRef<Partial<Record<TabId, number>>>({});
  const activeTabIdRef = useRef<TabId>("source");
  const [initialApplicationState, setInitialApplicationState] = useState<InitialApplicationState>(defaultInitialApplicationState);
  const [buildConfigSettings, setBuildConfigSettings] = useState<BuildConfigSettings>(emptyBuildConfigSettings);
  const [ffmpegBuildSettings, setFfmpegBuildSettings] = useState<FfmpegBuildSettings>(emptyFfmpegBuildSettings);
  const [libraryPresetId, setLibraryPresetId] = useState<LibraryPresetId>("default");
  const [extendedLibraries, setExtendedLibrariesState] = useState(false);
  const [libraryDetailedView, setLibraryDetailedView] = useState(false);
  const [optionsDetailedView, setOptionsDetailedView] = useState(false);
  const [libraryTechnicalDetails, setLibraryTechnicalDetails] = useState(false);
  const [optionsTechnicalDetails, setOptionsTechnicalDetails] = useState(false);
  const [librarySectionFilters, setLibrarySectionFilters] = useState<string[]>([]);
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
  const [toolchainStatus, setToolchainStatus] = useState<ToolchainStatus | null>(null);
  const [installedToolchainProfiles, setInstalledToolchainProfiles] = useState<ToolchainStatus[]>([]);
  const [toolchainVerification, setToolchainVerification] = useState<ToolchainVerification | null>(null);
  const [isVerifyingToolchain, setIsVerifyingToolchain] = useState(false);

  const approvedActionPhaseRef = useRef<"toolchain" | "ffmpeg" | null>(null);
  approvedActionPhaseRef.current = approvedActionPhase;
  activeTabIdRef.current = activeTabId;

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";
  const toolchainProgress = useMemo<LiveProgress>(() => computeProgress(toolchainLogEntries, approvedActionStatus, "toolchain"), [toolchainLogEntries, approvedActionStatus]);
  const ffmpegProgress = useMemo<LiveProgress>(() => computeProgress(ffmpegLogEntries, approvedActionStatus, "ffmpeg"), [ffmpegLogEntries, approvedActionStatus]);
  const securityLogEntries = useMemo(() => [...toolchainLogEntries, ...ffmpegLogEntries], [toolchainLogEntries, ffmpegLogEntries]);
  // Current Build configuration package list (live textarea), used by Prep to flag
  // drift between what is configured now and what the prepared toolchain installed.
  const configuredMsys2PackageNames = useMemo(() => splitLines(msys2PackageText), [msys2PackageText]);

  useEffect(() => {
    // Keep the backend in sync with the UI language so the native confirmation
    // dialog is shown in the same language. Logs stay English by design.
    void SetLocale(getLocale());
    const onLocaleChange = () => { void SetLocale(getLocale()); };
    window.addEventListener("customffmpeg-locale-change", onLocaleChange);

    GetInitialApplicationState().then(async (nextState: InitialApplicationState) => {
      const saved = parseSavedUiState(await LoadUiState());
      const savedBts = saved.buildConfigSettings ? { ...nextState.defaultBuildConfigSettings, ...saved.buildConfigSettings } : nextState.defaultBuildConfigSettings;
      let resolvedFbs = saved.ffmpegBuildSettings ? { ...nextState.defaultFfmpegBuildSettings, ...saved.ffmpegBuildSettings } : nextState.defaultFfmpegBuildSettings;
      const hasSavedPreset = isValidLibraryPresetId(saved.libraryPresetId);
      const resolvedPresetId: LibraryPresetId = hasSavedPreset ? saved.libraryPresetId! : "default";
      if (!hasSavedPreset && !saved.ffmpegBuildSettings) {
        const defaultPreset = libraryPresets.find((p) => p.presetId === "default");
        if (defaultPreset) {
          const nextIds = normalizeLibrarySelection(defaultPreset.libraryIds, resolvedFbs.windowsShellProfileName);
          resolvedFbs = { ...resolvedFbs, selectedLibraryIds: nextIds, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextIds, nextState.defaultLibraryCatalog, resolvedFbs.windowsShellProfileName) };
        }
      }
      setInitialApplicationState(nextState);
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
      if (Array.isArray(saved.librarySectionFilters)) setLibrarySectionFilters(saved.librarySectionFilters.filter((value): value is string => typeof value === "string"));
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
    return () => { removeLogListener(); removeStatusListener(); window.removeEventListener("customffmpeg-locale-change", onLocaleChange); };
  }, []);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    void SaveUiState(JSON.stringify({
      activeTabId,
      buildConfigSettings: { ...buildConfigSettings, msys2PackageNames: splitLines(msys2PackageText) },
      ffmpegBuildSettings: { ...ffmpegBuildSettings, extraConfigureFlags: splitLines(extraConfigureFlagText), configureFlags: splitLines(extraConfigureFlagText) },
      msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters,
    } satisfies SavedUiState));
  }, [activeTabId, buildConfigSettings, ffmpegBuildSettings, msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters]);

  useEffect(() => {
    const id = window.setInterval(() => { saveWindowState(); }, 2000);
    window.addEventListener("beforeunload", saveWindowState);
    return () => { window.clearInterval(id); window.removeEventListener("beforeunload", saveWindowState); saveWindowState(); };
  }, []);

  useEffect(() => {
    if (activeTabId === "result") refreshBuildResult();
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

  async function refreshBuildResult() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResult(null); setBuildResultError(t("result.error.chooseWorkspaceFirst")); return; }
    try { const r = await GetBuildResult(dir); setBuildResult(r); setBuildResultError(""); }
    catch (err) { setBuildResult(null); setBuildResultError(err instanceof Error ? err.message : String(err)); }
  }

  async function refreshToolchainStatus() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setToolchainStatus(null); setInstalledToolchainProfiles([]); return; }
    try { setToolchainStatus(await GetToolchainStatus(dir, buildConfigSettings.windowsShellProfileName)); }
    catch { setToolchainStatus(null); }
    try { setInstalledToolchainProfiles(await GetInstalledToolchainProfiles(dir)); }
    catch { setInstalledToolchainProfiles([]); }
  }


  async function clearBuildEnvironments() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) return;
    const confirmed = window.confirm(t("prep.profiles.clearConfirm"));
    if (!confirmed) return;
    setToolchainVerification(null);
    await ClearBuildEnvironments(dir);
    await refreshToolchainStatus();
  }

  async function verifyToolchain() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) return;
    setIsVerifyingToolchain(true);
    setToolchainVerification(null);
    try { setToolchainVerification(await VerifyToolchainInstallation(dir, buildConfigSettings.windowsShellProfileName)); }
    catch (err) { setToolchainVerification({ verified: false, checkedPackageCount: 0, missingPackageNames: [], message: err instanceof Error ? err.message : String(err) }); }
    finally { setIsVerifyingToolchain(false); }
  }

  async function openResultFolder() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResultError(t("result.error.chooseWorkspaceFirst")); return; }
    await OpenResultFolder(dir);
  }

  async function openResultReport() {
    const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!dir) { setBuildResultError(t("result.error.chooseWorkspaceFirst")); return; }
    await OpenResultReport(dir);
  }

  function updateBuildConfigSettings(next: Partial<BuildConfigSettings>) {
    setBuildConfigSettings((s) => ({ ...s, ...next }));
    setToolchainPreparationPlanReview(null);
  }

  function updateFfmpegBuildSettings(next: Partial<FfmpegBuildSettings>) {
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
    const nextLibraryIds = normalizeLibrarySelection(ffmpegBuildSettings.selectedLibraryIds, profileName);
    updateBuildConfigSettings({ windowsShellProfileName: profileName });
    updateFfmpegBuildSettings({
      windowsShellProfileName: profileName,
      selectedLibraryIds: nextLibraryIds,
      licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextLibraryIds, initialApplicationState.defaultLibraryCatalog, profileName),
    });
    setLibraryPresetId(matchLibraryPresetId(nextLibraryIds, profileName, initialApplicationState.defaultLibraryCatalog, extendedLibraries));
    setMsys2PackageText((text) => remapMsys2PackagePrefixes(text, profileName));
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
    const dir = await SelectWorkspace();
    if (!dir) return;
    updateBuildConfigSettings({ workspaceDirectory: dir });
    updateFfmpegBuildSettings({ workspaceDirectory: dir });
  }

  async function addBuildConfigPlanAndContinueToPrep() {
    const review = await RequestToolchainPreparationPlan({ ...buildConfigSettings, msys2PackageNames: splitLines(msys2PackageText) });
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
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    let next = removing ? current.filter((id) => id !== libraryId) : [...current, libraryId];
    if (!removing) next = removeMutuallyExclusiveLibraries(next, libraryId);
    next = normalizeLibrarySelection(next, profile);
    setLibraryPresetId(matchLibraryPresetId(next, profile, initialApplicationState.defaultLibraryCatalog, extendedLibraries));
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(next, initialApplicationState.defaultLibraryCatalog, profile) });
  }

  function applyLibraryPreset(presetId: LibraryPresetId) {
    const preset = libraryPresets.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const next = preset.dev
      ? maximumTestLibraryIds(initialApplicationState.defaultLibraryCatalog, profile)
      : normalizeLibrarySelection(presetLibraryIds(preset, extendedLibraries), profile);
    setLibraryPresetId(presetId);
    updateFfmpegBuildSettings({ selectedLibraryIds: next, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(next, initialApplicationState.defaultLibraryCatalog, profile) });
  }

  // Toggling the Extended mode re-applies the active named preset under the new mode so
  // its source-build extras are added/removed and the highlighted button stays truthful.
  // Custom/dev selections keep their libraries and are just re-matched against the new mode.
  function setExtendedLibraries(next: boolean) {
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const preset = libraryPresets.find((p) => p.presetId === libraryPresetId);
    if (preset && preset.presetId !== "custom" && !preset.dev) {
      const nextIds = normalizeLibrarySelection(presetLibraryIds(preset, next), profile);
      updateFfmpegBuildSettings({ selectedLibraryIds: nextIds, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextIds, initialApplicationState.defaultLibraryCatalog, profile) });
    } else {
      setLibraryPresetId(matchLibraryPresetId(ffmpegBuildSettings.selectedLibraryIds, profile, initialApplicationState.defaultLibraryCatalog, next));
    }
    setExtendedLibrariesState(next);
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
    const recommended = initialApplicationState.defaultBuildConfigSettings.msys2PackageNames.join("\n");
    setMsys2PackageText(remapMsys2PackagePrefixes(recommended, buildConfigSettings.windowsShellProfileName));
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
    buildConfigSettings,
    ffmpegBuildSettings,
    libraryPresetId,
    extendedLibraries,
    libraryDetailedView, setLibraryDetailedView,
    optionsDetailedView, setOptionsDetailedView,
    libraryTechnicalDetails, setLibraryTechnicalDetails,
    optionsTechnicalDetails, setOptionsTechnicalDetails,
    librarySectionFilters, setLibrarySectionFilters,
    msys2PackageText,
    extraConfigureFlagText,
    toolchainPreparationPlanReview,
    ffmpegBuildPlanReview,
    approvedActionStatus,
    approvedActionPhase,
    toolchainLogEntries,
    ffmpegLogEntries,
    buildResult, buildResultError,
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    configuredMsys2PackageNames,
    canCancelToolchain, canCancelFfmpeg,
    isApprovedActionRunning: canCancelApprovedAction,
    toolchainProgress, ffmpegProgress,
    securityLogEntries,
    updateBuildConfigSettings, updateFfmpegBuildSettings, updateMsys2ArchiveUrl, changeShellProfile,
    chooseWorkspaceDirectory,
    addBuildConfigPlanAndContinueToPrep, reviewFfmpegPlans,
    approveToolchainPreparationPlan, approveFfmpegBuildPlan, cancelApprovedAction,
    openInUserBrowser,
    toggleLibrary, applyLibraryPreset, setExtendedLibraries, toggleConfigureOption, applyOptionPreset,
    restoreRecommendedToolchainPackages, restoreRecommendedExtraFlags,
    handleMsys2PackageTextChange, handleExtraFlagTextChange,
    refreshBuildResult, openResultFolder, openResultReport,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  };
}
