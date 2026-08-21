import { useEffect, useMemo, useRef, useState } from "react";
import {
  LPlanFfmpegApprove,
  LFfmpegRetryRun,
  LPlanToolchainApprove,
  LActionApprovedCancel,
  LStateInitialGet,
  LCatalogSourceGet,
  LPresetSourceGet,
  LStateUiLoad,
  LStateUiSave,
  LPlanFfmpegRequest,
  LPlanToolchainRequest,
  LWorkspaceSelect,
  LLinkExternalOpen,
  LApprovalConfirmationResolve,
} from "../wailsjs/go/program/LProgram";
import { EventsOn } from "../wailsjs/runtime/runtime";

import { LLogSecurityEntry, LLogSecurityPayload, LStatusActionPayload, LStalledActionPayload, LProgressLive, LProgressGet } from "./tabs/logutils";
import {
  LPresetLibraryId, LPresetLibrary, LPresetLibraryClean, LPresetLibraryResolve,
  LLibrarySelectionNormalize, LPresetLibraryMatch, LPresetLibraryValidate,
  LLicenseBoundaryGet, LLibraryExclusiveRemove,
  LLibraryTestGet,
} from "./tabs/libraries";
import { LSectionStateNormalize } from "./tabs/librarycatalog";
import { LPresetOptionId, LPresetOptionCatalog } from "./tabs/options";
import {
  LTabIdentifier, LStateUiSaved,
  LSettingsBuildEmpty, LSettingsFfmpegEmpty, LStateInitialDefault,
  LTextLineSplit, LLogLevelNormalize, LTabIdValidate,
  LRequestApprovalCreate, LStateUiParse,
  LPackagePrefixUpdate, LSettingsBuildNormalize, LSettingsFfmpegNormalize, LConfigureOptionSelectionNormalize, LStateInitialNormalize,
} from "./programstate";
import { LStateResultUse } from "./builderresultstate";
import { LStateLogUse } from "./builderlogstate";
import { LStateToolchainStatusUse } from "./buildertoolchainstate";
import { LLocaleGet } from "./i18n";

export function LStateBuilderUse() {
  const locale = LLocaleGet();
  const [activeTabId, setActiveTabId] = useState<LTabIdentifier>("source");
  const hasLoadedSavedState = useRef(false);
  const [startupState, setStartupState] = useState<{ status: "loading" | "ready" | "error"; error: string }>({ status: "loading", error: "" });
  const [uiStatePersistenceError, setUiStatePersistenceError] = useState<{ operation: "load" | "save"; detail: string } | null>(null);
  const tabPanelRef = useRef<HTMLElement>(null);
  // Tabs whose scroll position is remembered when leaving and restored when
  // returning (within the session). Every other tab still scrolls back to top.
  const scrollRememberedTabIds = useRef(new Set<LTabIdentifier>(["options", "buildFfmpeg"]));
  const tabScrollPositions = useRef<Partial<Record<LTabIdentifier, number>>>({});
  const activeTabIdRef = useRef<LTabIdentifier>("source");
  const buildConfigSettingsRef = useRef<LSettingsToolchain>(LSettingsBuildEmpty);
  const ffmpegBuildSettingsRef = useRef<LSettingsFfmpeg>(LSettingsFfmpegEmpty);
  const msys2PackageTextRef = useRef("");
  const extraConfigureFlagTextRef = useRef("");
  const uiStateSaveQueueRef = useRef<Promise<void>>(Promise.resolve());
  const toolchainPlanPendingRef = useRef(false);
  const ffmpegPlanPendingRef = useRef(false);
  const [initialProgramState, setInitialProgramState] = useState<LStateInitial>(LStateInitialDefault);
  const [libraryCatalog, setLibraryCatalog] = useState<LLibraryChoice[]>([]);
  const [libraryPresetCatalog, setLibraryPresetCatalog] = useState<LPresetLibrary[]>([]);
  const [buildConfigSettings, setBuildConfigSettings] = useState<LSettingsToolchain>(LSettingsBuildEmpty);
  const [ffmpegBuildSettings, setFfmpegBuildSettings] = useState<LSettingsFfmpeg>(LSettingsFfmpegEmpty);
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
  const [ffmpegBuildPlanReview, setFfmpegBuildPlanReview] = useState<LReviewFfmpeg | null>(null);
  const [approvedActionStatus, setApprovedActionStatus] = useState("idle");
  const [approvedActionPhase, setApprovedActionPhase] = useState<"toolchain" | "ffmpeg" | null>(null);
  // Set when an approval attempt is refused or fails before any worker starts.
  // The backend keeps the review retryable, so the frontend keeps the approval
  // panel visible and shows the reason there instead of a false failed-run card.
  const [toolchainApprovalError, setToolchainApprovalError] = useState<string | null>(null);
  const [ffmpegApprovalError, setFfmpegApprovalError] = useState<string | null>(null);
  const [approvalConfirmationRequest, setApprovalConfirmationRequest] = useState<{ requestId: string; actionName: string; planHash: string } | null>(null);
  // Mirror addresses tried before a transient-network stall halted the run,
  // delivered by the "approved-action-stalled" event for the stalled banner.
  const [ffmpegStalledAddresses, setFfmpegStalledAddresses] = useState<string[]>([]);
  // Whether an approved FFmpeg run exists for Retry to resume after a stall.
  // The plan and approval themselves live on the backend, which re-enforces the
  // full approval boundary (lifetime + native confirmation) on retry; the
  // frontend only needs to know a run is available and calls LFfmpegRetryRun.
  const pFfmpegRunLast = useRef<boolean>(false);
  const [toolchainLogEntries, setToolchainLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [ffmpegLogEntries, setFfmpegLogEntries] = useState<LLogSecurityEntry[]>([]);
  const [isPlanningToolchain, setIsPlanningToolchain] = useState(false);
  const [isPlanningFfmpeg, setIsPlanningFfmpeg] = useState(false);

  // Workspace directory shared by result, log, and toolchain state.
  const dir = buildConfigSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
  const {
    buildResult, buildResultError, isLoadingBuildResult,
    buildVerification, buildVerificationError, isVerifyingBuild,
    LResultClear, refreshBuildResult, verifyBuildResult, openResultFolder, openResultReport,
  } = LStateResultUse(dir);
  const {
    localLogRecords, localLogRecordsError, setLocalLogRecordsError,
    refreshLocalLogRecords, loadLocalLogRecord,
    openLocalLogsFolder, openLocalLogRecordFolder, openLocalLogRecordFile,
  } = LStateLogUse(dir);
  const {
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    refreshToolchainStatus, verifyToolchain, clearBuildEnvironments,
  } = LStateToolchainStatusUse(dir, buildConfigSettings.windowsShellProfileName);
  const approvedActionPhaseRef = useRef<"toolchain" | "ffmpeg" | null>(null);
  const libraryPresetIdRef = useRef<LPresetLibraryId>("default");
  approvedActionPhaseRef.current = approvedActionPhase;
  libraryPresetIdRef.current = libraryPresetId;
  activeTabIdRef.current = activeTabId;
  buildConfigSettingsRef.current = buildConfigSettings;
  ffmpegBuildSettingsRef.current = ffmpegBuildSettings;
  msys2PackageTextRef.current = msys2PackageText;
  extraConfigureFlagTextRef.current = extraConfigureFlagText;

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed" && approvedActionStatus !== "stalled", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";
  const toolchainProgress = useMemo<LProgressLive>(() => LProgressGet(toolchainLogEntries, approvedActionStatus, "toolchain"), [toolchainLogEntries, approvedActionStatus, locale]);
  const ffmpegProgress = useMemo<LProgressLive>(() => LProgressGet(ffmpegLogEntries, approvedActionStatus, "ffmpeg"), [ffmpegLogEntries, approvedActionStatus, locale]);
  const securityLogEntries = useMemo(() => [...toolchainLogEntries, ...ffmpegLogEntries], [toolchainLogEntries, ffmpegLogEntries]);
  // Current Build configuration package list (live textarea), used by Prep to flag
  // drift between what is configured now and what the prepared toolchain installed.
  const configuredMsys2PackageNames = useMemo(() => LTextLineSplit(msys2PackageText), [msys2PackageText]);

  useEffect(() => {
    let isHydrationCurrent = true;
    void (async () => {
      try {
        const nextState = LStateInitialNormalize(await LStateInitialGet());
        let saved: LStateUiSaved = {};
        try {
          saved = LStateUiParse(await LStateUiLoad());
        } catch (error) {
          if (isHydrationCurrent) setUiStatePersistenceError({ operation: "load", detail: error instanceof Error ? error.message : String(error) });
        }
        if (!isHydrationCurrent) return;
        const savedBts = LSettingsBuildNormalize(saved.buildConfigSettings, nextState.defaultBuildConfigSettings);
        let resolvedFbs = LSettingsFfmpegNormalize(saved.ffmpegBuildSettings, nextState.defaultFfmpegBuildSettings);
        resolvedFbs = {
          ...resolvedFbs,
          selectedConfigureOptionIds: LConfigureOptionSelectionNormalize(resolvedFbs.selectedConfigureOptionIds, nextState.defaultConfigureOptionCatalog),
        };
        // The profile selector is the single authority. Normal UI changes update
        // both settings objects, and hydration must preserve that invariant too.
        resolvedFbs = { ...resolvedFbs, windowsShellProfileName: savedBts.windowsShellProfileName };
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
        setMsys2PackageText(typeof saved.msys2PackageText === "string" ? saved.msys2PackageText : savedBts.msys2PackageNames.join("\n"));
        setExtraConfigureFlagText(typeof saved.extraConfigureFlagText === "string" ? saved.extraConfigureFlagText : resolvedFbs.extraConfigureFlags.join("\n"));
        setLibraryPresetId(resolvedPresetId);
        if (typeof saved.extendedLibraries === "boolean") setExtendedLibrariesState(saved.extendedLibraries);
        if (typeof saved.libraryDetailedView === "boolean") setLibraryDetailedView(saved.libraryDetailedView);
        if (typeof saved.optionsDetailedView === "boolean") setOptionsDetailedView(saved.optionsDetailedView);
        if (typeof saved.libraryTechnicalDetails === "boolean") setLibraryTechnicalDetails(saved.libraryTechnicalDetails);
        if (typeof saved.optionsTechnicalDetails === "boolean") setOptionsTechnicalDetails(saved.optionsTechnicalDetails);
        if (Array.isArray(saved.librarySectionFilters)) {
          const savedFilters = saved.librarySectionFilters.filter((value): value is string => typeof value === "string");
          setLLibrarySectionFilters(LSectionStateNormalize(savedFilters, nextState.defaultLibraryCatalog));
        }
        if (LTabIdValidate(saved.activeTabId)) setActiveTabId(saved.activeTabId);
        hasLoadedSavedState.current = true;
      } catch (error) {
        if (!isHydrationCurrent) return;
        setStartupState({ status: "error", error: error instanceof Error ? error.message : String(error) });
      }
    })();

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
      if (payload.status === "failed") { setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null); setToolchainApprovalError(null); setFfmpegApprovalError(null); LResultClear(); }
      if (payload.status === "completed") { LResultClear(); setApprovedActionPhase(null); }
      // A stall halts the run in a non-active retryable state: drop the "ffmpeg"
      // phase so the live progress stops reading as in-flight and the orange
      // stalled banner (not a spinner) becomes the authoritative signal.
      if (payload.status === "stalled") { setApprovedActionPhase(null); }
    });
    const removeStalledListener = EventsOn("approved-action-stalled", (payload: LStalledActionPayload) => {
      setFfmpegStalledAddresses(Array.isArray(payload.addresses) ? payload.addresses : []);
    });
    const removeConfirmationListener = EventsOn("approval-confirmation-request", (payload: { requestId?: string; actionName?: string; planHash?: string }) => {
      if (payload?.requestId && payload.actionName && payload.planHash) {
        setApprovalConfirmationRequest({ requestId: payload.requestId, actionName: payload.actionName, planHash: payload.planHash });
      }
    });
    const removeConfirmationClosedListener = EventsOn("approval-confirmation-closed", (payload: { requestId?: string }) => {
      setApprovalConfirmationRequest((request) => request?.requestId === payload?.requestId ? null : request);
    });
    return () => { isHydrationCurrent = false; removeLogListener(); removeStatusListener(); removeStalledListener(); removeConfirmationListener(); removeConfirmationClosedListener(); };
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
        setStartupState({ status: "ready", error: "" });
      })
      .catch((error) => {
        if (isCurrent) {
          setLibraryCatalog(initialProgramState.defaultLibraryCatalog);
          setLibraryPresetCatalog(LPresetLibraryClean(initialProgramState.defaultLibraryPresetCatalog));
          setLocalLogRecordsError(`Unable to resolve the FFmpeg catalog. ${error instanceof Error ? error.message : String(error)}`);
          // Resolver failure is non-destructive: retain the restored selection
          // and allow the UI to open with a visible diagnostic.
          setStartupState({ status: "ready", error: "" });
        }
      });
    return () => { isCurrent = false; };
  }, [ffmpegBuildSettings.ffmpegSourceArchiveUrl, ffmpegBuildSettings.windowsShellProfileName, initialProgramState.defaultLibraryCatalog, initialProgramState.defaultLibraryPresetCatalog, extendedLibraries]);

  useEffect(() => {
    if (!hasLoadedSavedState.current) return;
    const serializedState = JSON.stringify({
      activeTabId,
      buildConfigSettings: { ...buildConfigSettings, msys2PackageNames: LTextLineSplit(msys2PackageText) },
      ffmpegBuildSettings: { ...ffmpegBuildSettings, extraConfigureFlags: LTextLineSplit(extraConfigureFlagText), configureFlags: LTextLineSplit(extraConfigureFlagText) },
      msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters,
    } satisfies LStateUiSaved);
    uiStateSaveQueueRef.current = uiStateSaveQueueRef.current
      .then(async () => {
        await LStateUiSave(serializedState);
        setUiStatePersistenceError(null);
      })
      .catch((error) => {
        console.error("Failed to save UI state", error);
        setUiStatePersistenceError({ operation: "save", detail: error instanceof Error ? error.message : String(error) });
      });
  }, [activeTabId, buildConfigSettings, ffmpegBuildSettings, msys2PackageText, extraConfigureFlagText, libraryPresetId, extendedLibraries, libraryDetailedView, optionsDetailedView, libraryTechnicalDetails, optionsTechnicalDetails, librarySectionFilters]);

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

  function LSettingsToolchainUpdate(next: Partial<LSettingsToolchain>) {
    setBuildConfigSettings((s) => ({ ...s, ...next }));
    setToolchainPreparationPlanReview(null);
  }

  function LSettingsFfmpegUpdate(next: Partial<LSettingsFfmpeg>) {
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
    LSettingsFfmpegUpdate({
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
    const nextDir = await LWorkspaceSelect();
    if (!nextDir) return;
    LSettingsToolchainUpdate({ workspaceDirectory: nextDir });
    LSettingsFfmpegUpdate({ workspaceDirectory: nextDir });
  }

  async function addBuildConfigPlanAndContinueToPrep() {
    if (toolchainPlanPendingRef.current) return;
    toolchainPlanPendingRef.current = true;
    setIsPlanningToolchain(true);
    const originTabId = activeTabIdRef.current;
    const requestedSettings = { ...buildConfigSettings, msys2PackageNames: LTextLineSplit(msys2PackageText) };
    const requestSignature = JSON.stringify(requestedSettings);
    try {
      const review = await LPlanToolchainRequest(requestedSettings);
      const currentSignature = JSON.stringify({ ...buildConfigSettingsRef.current, msys2PackageNames: LTextLineSplit(msys2PackageTextRef.current) });
      if (requestSignature !== currentSignature) return;
      setToolchainApprovalError(null);
      setToolchainPreparationPlanReview(review);
      if (activeTabIdRef.current === originTabId) setActiveTabId("prep");
    } finally {
      toolchainPlanPendingRef.current = false;
      setIsPlanningToolchain(false);
    }
  }

  function LPlanToolchainCancel() {
    setToolchainPreparationPlanReview(null);
    setToolchainApprovalError(null);
  }

  async function reviewFfmpegPlans() {
    if (ffmpegPlanPendingRef.current) return;
    ffmpegPlanPendingRef.current = true;
    setIsPlanningFfmpeg(true);
    const originTabId = activeTabIdRef.current;
    const flags = LTextLineSplit(extraConfigureFlagText);
    const requestedSettings = { ...ffmpegBuildSettings, extraConfigureFlags: flags, configureFlags: flags };
    const requestSignature = JSON.stringify(requestedSettings);
    try {
      const review = await LPlanFfmpegRequest(requestedSettings);
      const currentFlags = LTextLineSplit(extraConfigureFlagTextRef.current);
      const currentSignature = JSON.stringify({ ...ffmpegBuildSettingsRef.current, extraConfigureFlags: currentFlags, configureFlags: currentFlags });
      if (requestSignature !== currentSignature) return;
      setFfmpegApprovalError(null);
      setFfmpegBuildPlanReview(review);
      if (activeTabIdRef.current === originTabId) setActiveTabId("buildFfmpeg");
    } finally {
      ffmpegPlanPendingRef.current = false;
      setIsPlanningFfmpeg(false);
    }
  }

  async function approveToolchainPreparationPlan() {
    if (!toolchainPreparationPlanReview) return;
    const originTabId = activeTabIdRef.current;
    const r = toolchainPreparationPlanReview;
    setToolchainLogEntries([]); setToolchainApprovalError(null); setApprovedActionPhase("toolchain"); setApprovedActionStatus("starting");
    try {
      await LPlanToolchainApprove(r.reviewSessionId, LRequestApprovalCreate(r.plan.actionName, r.plan.planHash, r.expectedLConsentText));
      setToolchainPreparationPlanReview(null);
      if (activeTabIdRef.current === originTabId) setActiveTabId("prep");
    } catch (err) {
      // Rejected or failed before a worker started; the backend left the review
      // retryable. Return to the approval panel (clear the pre-await phase, go
      // back to idle) and surface the reason on the panel rather than emitting a
      // failed progress card whose Clear would delete the still-valid review.
      setApprovedActionPhase(null); setApprovedActionStatus("idle");
      setToolchainApprovalError(err instanceof Error ? err.message : String(err));
    }
  }

  async function approveFfmpegBuildPlan() {
    if (!ffmpegBuildPlanReview) return;
    const originTabId = activeTabIdRef.current;
    const r = ffmpegBuildPlanReview;
    const approval = LRequestApprovalCreate(r.plan.actionName, r.plan.planHash, r.expectedLConsentText);
    setFfmpegLogEntries([]); setFfmpegStalledAddresses([]); setFfmpegApprovalError(null); setApprovedActionStatus("starting");
    try {
      await LPlanFfmpegApprove(r.reviewSessionId, approval);
      // The backend retained this run's plan and approval; mark that a retry
      // target exists so the stalled banner can offer Retry.
      pFfmpegRunLast.current = true;
      setApprovedActionPhase("ffmpeg");
      setFfmpegBuildPlanReview(null);
      if (activeTabIdRef.current === originTabId) setActiveTabId("buildFfmpeg");
    } catch (err) {
      // Rejected or failed before a worker started; the backend left the review
      // retryable. Keep the approval card (phase stays null, go back to idle) and
      // show the reason on it rather than a false failed-run card whose Clear
      // would delete the still-valid review.
      setApprovedActionStatus("idle");
      setFfmpegApprovalError(err instanceof Error ? err.message : String(err));
    }
  }

  // Re-run the same approved FFmpeg action after a stall. LFfmpegRetryRun
  // resumes from backend-retained state and re-enforces the approval boundary
  // (original lifetime + native confirmation) there, so the frontend passes no
  // plan and cannot renew or alter an expired approval.
  async function retryFfmpegBuildPlan() {
    if (!pFfmpegRunLast.current) return;
    const originTabId = activeTabIdRef.current;
    setFfmpegLogEntries([]); setFfmpegStalledAddresses([]); setApprovedActionStatus("starting");
    try {
      await LFfmpegRetryRun();
      setApprovedActionPhase("ffmpeg");
      if (activeTabIdRef.current === originTabId) setActiveTabId("buildFfmpeg");
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
    setToolchainApprovalError(null); setFfmpegApprovalError(null);
    setToolchainPreparationPlanReview(null); setFfmpegBuildPlanReview(null);
    LResultClear();
    setApprovedActionPhase(null); setApprovedActionStatus("idle");
  }
  async function openInUserBrowser(url: string) {
    try { await LLinkExternalOpen(url); }
    catch (err) { console.error("Unable to open external link", err); }
  }

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
    LSettingsFfmpegUpdate({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryGet(next, libraryCatalog, profile) });
  }

  function LPresetLibraryApply(presetId: LPresetLibraryId) {
    const preset = libraryPresetCatalog.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const next = preset.dev
      ? LLibraryTestGet(libraryCatalog, profile)
      : LLibrarySelectionNormalize(LPresetLibraryResolve(preset, extendedLibraries), profile, libraryCatalog);
    setLibraryPresetId(presetId);
    LSettingsFfmpegUpdate({ selectedLibraryIds: next, licenseProfileName: LLicenseBoundaryGet(next, libraryCatalog, profile) });
  }

  // Toggling the Extended mode re-applies the active named preset under the new mode so
  // its source-build extras are added/removed and the highlighted button stays truthful.
  // Custom/dev selections keep their libraries and are just re-matched against the new mode.
  function LLibraryExtendedUpdate(next: boolean) {
    const profile = ffmpegBuildSettings.windowsShellProfileName;
    const preset = libraryPresetCatalog.find((p) => p.presetId === libraryPresetId);
    if (preset && preset.presetId !== "custom" && !preset.dev) {
      const nextIds = LLibrarySelectionNormalize(LPresetLibraryResolve(preset, next), profile, libraryCatalog);
      LSettingsFfmpegUpdate({ selectedLibraryIds: nextIds, licenseProfileName: LLicenseBoundaryGet(nextIds, libraryCatalog, profile) });
    } else {
      setLibraryPresetId(LPresetLibraryMatch(ffmpegBuildSettings.selectedLibraryIds, libraryPresetCatalog, profile, libraryCatalog, next));
    }
    setExtendedLibrariesState(next);
  }

  function LOptionToggle(optionId: string) {
    const option = initialProgramState.defaultConfigureOptionCatalog.find((o) => o.optionId === optionId);
    if (option?.locked) return;
    const current = ffmpegBuildSettings.selectedConfigureOptionIds;
    LSettingsFfmpegUpdate({ selectedConfigureOptionIds: current.includes(optionId) ? current.filter((id) => id !== optionId) : [...current, optionId] });
  }

  function LPresetOptionApply(presetId: LPresetOptionId) {
    const preset = LPresetOptionCatalog.find((p) => p.presetId === presetId);
    if (!preset || preset.presetId === "custom") return;
    LSettingsFfmpegUpdate({ selectedConfigureOptionIds: preset.optionIds });
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
    LSettingsFfmpegUpdate({ selectedConfigureOptionIds: optIds, extraConfigureFlags: flags, configureFlags: flags });
  }

  function LMSYSPackageUpdate(text: string) { setMsys2PackageText(text); setToolchainPreparationPlanReview(null); }
  function LFlagExtraUpdate(text: string) { setExtraConfigureFlagText(text); setFfmpegBuildPlanReview(null); }
  async function resolveApprovalConfirmation(approved: boolean) {
    const request = approvalConfirmationRequest;
    if (!request) return;
    setApprovalConfirmationRequest(null);
    await LApprovalConfirmationResolve(request.requestId, approved);
  }
  return {
    startupState,
    uiStatePersistenceError, clearUiStatePersistenceError: () => setUiStatePersistenceError(null),
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
    toolchainApprovalError,
    ffmpegApprovalError,
    approvalConfirmationRequest, resolveApprovalConfirmation,
    ffmpegStalledAddresses,
    toolchainLogEntries,
    ffmpegLogEntries,
    localLogRecords, localLogRecordsError,
    buildResult, buildResultError, isLoadingBuildResult,
    buildVerification, buildVerificationError, isVerifyingBuild, verifyBuildResult,
    toolchainStatus, installedToolchainProfiles, toolchainVerification, isVerifyingToolchain,
    isPlanningToolchain, isPlanningFfmpeg,
    configuredMsys2PackageNames,
    canCancelToolchain, canCancelFfmpeg,
    isApprovedActionRunning: canCancelApprovedAction,
    toolchainProgress, ffmpegProgress,
    securityLogEntries,
    LSettingsToolchainUpdate, LSettingsFfmpegUpdate, LMSYSArchiveUpdate, LProfileShellUpdate,
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
