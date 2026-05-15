import React, { useEffect, useMemo, useRef, useState } from "react";
import { createRoot } from "react-dom/client";
import {
  ApproveFfmpegBuildPlan,
  ApproveToolchainPreparationPlan,
  CancelApprovedAction,
  GetBuildResult,
  GetInitialApplicationState,
  RequestFfmpegBuildPlan,
  RequestToolchainPreparationPlan,
  OpenResultFolder,
  SelectWorkspace,
} from "../wailsjs/go/main/App";
import { BrowserOpenURL, EventsOn, WindowGetPosition, WindowGetSize, WindowSetPosition, WindowSetSize } from "../wailsjs/runtime/runtime";
import "./style.css";

type TabId = "source" | "buildTools" | "prep" | "library" | "options" | "buildFfmpeg" | "result" | "logs";

type LibraryPresetId = "default" | "efficiency" | "compatibility" | "editor" | "full" | "custom";

type LibraryPreset = {
  presetId: LibraryPresetId;
  displayName: string;
  plainExplanation: string;
  technicalExplanation: string;
  libraryIds: string[];
};

type SecurityLogEntry = {
  timestamp: string;
  level: "info" | "warn" | "error";
  message: string;
};

type SecurityLogPayload = {
  level: string;
  message: string;
};

type ApprovedActionStatusPayload = {
  status: string;
};

type SavedUiState = {
  activeTabId?: TabId;
  buildToolSettings?: BuildToolSettings;
  ffmpegBuildSettings?: FfmpegBuildSettings;
  msys2PackageText?: string;
  extraConfigureFlagText?: string;
  libraryPresetId?: LibraryPresetId;
};

type SavedWindowState = {
  width?: number;
  height?: number;
  x?: number;
  y?: number;
};

const savedUiStateKey = "customffmpeg.builder.uiState.v1";
const savedWindowStateKey = "customffmpeg.builder.windowState.v1";

const emptyBuildToolSettings: BuildToolSettings = {
  workspaceDirectory: "",
  msys2ArchiveUrl: "",
  msys2ArchiveSha256Hash: "",
  msys2ArchiveSignatureUrl: "",
  msys2PackageNames: [],
  windowsShellProfileName: "ucrt64",
};

const emptyFfmpegBuildSettings: FfmpegBuildSettings = {
  workspaceDirectory: "",
  ffmpegSourceArchiveUrl: "",
  ffmpegSourceSignatureUrl: "",
  ffmpegSourceSha256Hash: "",
  selectedLibraryIds: [],
  selectedConfigureOptionIds: ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"],
  extraConfigureFlags: [],
  configureFlags: [],
  parallelJobCount: 1,
  windowsShellProfileName: "ucrt64",
  licenseProfileName: "lgpl-local",
};

const defaultInitialApplicationState: InitialApplicationState = {
  hostOs: "unknown",
  kindExplanation: "This app does not bundle FFmpeg, libraries, codecs, or generated binaries.",
  securityRuleSummary: "The frontend reviews and requests. The backend confirms and runs.",
  namingRuleSummary: "Names must expose the real operation.",
  defaultBuildToolSettings: emptyBuildToolSettings,
  defaultFfmpegBuildSettings: emptyFfmpegBuildSettings,
  defaultLibraryCatalog: [],
  defaultConfigureOptionCatalog: [],
};

const officialLinks = {
  msys2ArchiveList: "https://repo.msys2.org/distrib/x86_64/",
  msys2InstallerDocs: "https://www.msys2.org/docs/installer/",
  ffmpegDownload: "https://www.ffmpeg.org/download.html",
  ffmpegReleaseIndex: "https://ffmpeg.org/releases/",
  ffmpegSigningKey: "https://ffmpeg.org/ffmpeg-devel.asc",
};

function BuilderApp() {
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
  const [securityLogEntries, setSecurityLogEntries] = useState<SecurityLogEntry[]>([]);
  const [buildResult, setBuildResult] = useState<BuildResult | null>(null);
  const [buildResultError, setBuildResultError] = useState("");

  const canCancelApprovedAction = useMemo(() => approvedActionStatus !== "idle" && approvedActionStatus !== "completed" && approvedActionStatus !== "failed", [approvedActionStatus]);
  const canCancelToolchain = canCancelApprovedAction && approvedActionPhase === "toolchain";
  const canCancelFfmpeg = canCancelApprovedAction && approvedActionPhase === "ffmpeg";

  useEffect(() => {
    GetInitialApplicationState().then((nextInitialApplicationState: InitialApplicationState) => {
      const savedUiState = readSavedUiState();
      const savedBuildToolSettings = savedUiState.buildToolSettings ? { ...nextInitialApplicationState.defaultBuildToolSettings, ...savedUiState.buildToolSettings } : nextInitialApplicationState.defaultBuildToolSettings;
      const savedFfmpegBuildSettings = savedUiState.ffmpegBuildSettings ? { ...nextInitialApplicationState.defaultFfmpegBuildSettings, ...savedUiState.ffmpegBuildSettings } : nextInitialApplicationState.defaultFfmpegBuildSettings;
      setInitialApplicationState(nextInitialApplicationState);
      setBuildToolSettings(savedBuildToolSettings);
      setFfmpegBuildSettings(savedFfmpegBuildSettings);
      setMsys2PackageText(savedUiState.msys2PackageText ?? savedBuildToolSettings.msys2PackageNames.join("\n"));
      setExtraConfigureFlagText(savedUiState.extraConfigureFlagText ?? savedFfmpegBuildSettings.extraConfigureFlags.join("\n"));
      setLibraryPresetId(isValidLibraryPresetId(savedUiState.libraryPresetId) ? savedUiState.libraryPresetId : matchLibraryPresetId(savedFfmpegBuildSettings.selectedLibraryIds));
      if (isValidTabId(savedUiState.activeTabId)) {
        setActiveTabId(savedUiState.activeTabId);
      }
      hasLoadedSavedState.current = true;
      restoreWindowState();
    });

    const removeSecurityLogListener = EventsOn("security-log", (payload: SecurityLogPayload) => {
      setSecurityLogEntries((currentEntries) => [...currentEntries, { level: normalizeLogLevel(payload.level), message: payload.message, timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    });
    const removeStatusListener = EventsOn("approved-action-status", (payload: ApprovedActionStatusPayload) => {
      setApprovedActionStatus(payload.status);
      if (payload.status === "failed") {
        setToolchainPreparationPlanReview(null);
        setFfmpegBuildPlanReview(null);
        setBuildResult(null);
      }
      if (payload.status === "completed") {
        setBuildResult(null);
        setApprovedActionPhase(null);
      }
    });
    return () => {
      removeSecurityLogListener();
      removeStatusListener();
    };
  }, []);

  useEffect(() => {
    if (!hasLoadedSavedState.current) {
      return;
    }
    saveUiState({
      activeTabId,
      buildToolSettings: { ...buildToolSettings, msys2PackageNames: splitLines(msys2PackageText) },
      ffmpegBuildSettings: { ...ffmpegBuildSettings, extraConfigureFlags: splitLines(extraConfigureFlagText), configureFlags: splitLines(extraConfigureFlagText) },
      msys2PackageText,
      extraConfigureFlagText,
      libraryPresetId,
    });
  }, [activeTabId, buildToolSettings, ffmpegBuildSettings, msys2PackageText, extraConfigureFlagText, libraryPresetId]);

  useEffect(() => {
    const intervalId = window.setInterval(() => {
      saveWindowState();
    }, 2000);
    window.addEventListener("beforeunload", saveWindowState);
    return () => {
      window.clearInterval(intervalId);
      window.removeEventListener("beforeunload", saveWindowState);
      saveWindowState();
    };
  }, []);

  useEffect(() => {
    if (activeTabId !== "result") {
      return;
    }
    refreshBuildResult();
  }, [activeTabId, buildToolSettings.workspaceDirectory]);

  useEffect(() => {
    if (tabPanelRef.current) {
      tabPanelRef.current.scrollTop = 0;
    }
  }, [activeTabId]);

  async function refreshBuildResult() {
    const workspaceDirectory = buildToolSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!workspaceDirectory) {
      setBuildResult(null);
      setBuildResultError("Choose a workspace first.");
      return;
    }
    try {
      const nextResult = await GetBuildResult(workspaceDirectory);
      setBuildResult(nextResult);
      setBuildResultError("");
    } catch (error) {
      setBuildResult(null);
      setBuildResultError(error instanceof Error ? error.message : String(error));
    }
  }

  async function openResultFolder() {
    const workspaceDirectory = buildToolSettings.workspaceDirectory || ffmpegBuildSettings.workspaceDirectory;
    if (!workspaceDirectory) {
      setBuildResultError("Choose a workspace first.");
      return;
    }
    await OpenResultFolder(workspaceDirectory);
  }

  function updateBuildToolSettings(nextPartialSettings: Partial<BuildToolSettings>) {
    setBuildToolSettings((currentSettings) => ({ ...currentSettings, ...nextPartialSettings }));
    setToolchainPreparationPlanReview(null);
  }

  function updateFfmpegBuildSettings(nextPartialSettings: Partial<FfmpegBuildSettings>) {
    setFfmpegBuildSettings((currentSettings) => ({ ...currentSettings, ...nextPartialSettings }));
    setFfmpegBuildPlanReview(null);
  }

  function updateMsys2ArchiveUrl(nextArchiveUrl: string) {
    setBuildToolSettings((currentSettings) => {
      const oldAutoSignatureUrl = currentSettings.msys2ArchiveUrl ? `${currentSettings.msys2ArchiveUrl}.sig` : "";
      const shouldUpdateSignatureUrl = currentSettings.msys2ArchiveSignatureUrl === "" || currentSettings.msys2ArchiveSignatureUrl === oldAutoSignatureUrl;
      return {
        ...currentSettings,
        msys2ArchiveUrl: nextArchiveUrl,
        msys2ArchiveSignatureUrl: shouldUpdateSignatureUrl && nextArchiveUrl ? `${nextArchiveUrl}.sig` : currentSettings.msys2ArchiveSignatureUrl,
      };
    });
    setToolchainPreparationPlanReview(null);
  }

  async function chooseWorkspaceDirectory() {
    const selectedWorkspaceDirectory = await SelectWorkspace();
    if (!selectedWorkspaceDirectory) {
      return;
    }
    updateBuildToolSettings({ workspaceDirectory: selectedWorkspaceDirectory });
    updateFfmpegBuildSettings({ workspaceDirectory: selectedWorkspaceDirectory });
  }

  async function addBuildToolsPlanAndContinueToPrep() {
    const nextSettings: BuildToolSettings = {
      ...buildToolSettings,
      msys2PackageNames: splitLines(msys2PackageText),
    };
    const nextPlanReview = await RequestToolchainPreparationPlan(nextSettings);
    setToolchainPreparationPlanReview(nextPlanReview);
    setActiveTabId("prep");
  }

  async function reviewFfmpegPlans() {
    const nextSettings: FfmpegBuildSettings = {
      ...ffmpegBuildSettings,
      extraConfigureFlags: splitLines(extraConfigureFlagText),
      configureFlags: splitLines(extraConfigureFlagText),
    };
    const nextPlanReview = await RequestFfmpegBuildPlan(nextSettings);
    setFfmpegBuildPlanReview(nextPlanReview);
    setActiveTabId("buildFfmpeg");
  }

  async function approveToolchainPreparationPlan() {
    if (!toolchainPreparationPlanReview) {
      return;
    }
    const reviewToRun = toolchainPreparationPlanReview;
    setSecurityLogEntries([]);
    setApprovedActionPhase("toolchain");
    setApprovedActionStatus("starting");
    try {
      await ApproveToolchainPreparationPlan(reviewToRun.reviewSessionId, createApprovalRequest(reviewToRun.plan.actionName, reviewToRun.plan.planHash, reviewToRun.expectedConsentText));
      setToolchainPreparationPlanReview(null);
      setActiveTabId("prep");
    } catch (error) {
      setApprovedActionStatus("failed");
      setSecurityLogEntries((currentEntries) => [...currentEntries, { level: "error", message: error instanceof Error ? error.message : String(error), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function approveFfmpegBuildPlan() {
    if (!ffmpegBuildPlanReview) {
      return;
    }
    const reviewToRun = ffmpegBuildPlanReview;
    setSecurityLogEntries([]);
    setApprovedActionPhase("ffmpeg");
    setApprovedActionStatus("starting");
    try {
      await ApproveFfmpegBuildPlan(reviewToRun.reviewSessionId, createApprovalRequest(reviewToRun.plan.actionName, reviewToRun.plan.planHash, reviewToRun.expectedConsentText));
      setFfmpegBuildPlanReview(null);
      setActiveTabId("buildFfmpeg");
    } catch (error) {
      setApprovedActionStatus("failed");
      setSecurityLogEntries((currentEntries) => [...currentEntries, { level: "error", message: error instanceof Error ? error.message : String(error), timestamp: new Date().toLocaleTimeString('en-GB', { hour: '2-digit', minute: '2-digit', second: '2-digit', hour12: false }) }]);
    }
  }

  async function cancelApprovedAction() {
    await CancelApprovedAction();
  }

  async function openInUserBrowser(urlToOpen: string) {
    BrowserOpenURL(urlToOpen);
  }

  function toggleLibrary(libraryId: string) {
    const library = initialApplicationState.defaultLibraryCatalog.find((candidateLibrary) => candidateLibrary.libraryId === libraryId);
    if (library?.locked) {
      return;
    }
    const currentIds = ffmpegBuildSettings.selectedLibraryIds;
    const isRemoving = currentIds.includes(libraryId);
    let nextIds = isRemoving ? currentIds.filter((currentId) => currentId !== libraryId) : [...currentIds, libraryId];
    if (!isRemoving) {
      nextIds = removeMutuallyExclusiveLibraries(nextIds, libraryId);
    }
    nextIds = normalizeLibrarySelection(nextIds);
    setLibraryPresetId(matchLibraryPresetId(nextIds));
    updateFfmpegBuildSettings({ selectedLibraryIds: nextIds, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextIds, initialApplicationState.defaultLibraryCatalog) });
  }

  function applyLibraryPreset(nextPresetId: LibraryPresetId) {
    const preset = libraryPresets.find((candidatePreset) => candidatePreset.presetId === nextPresetId);
    if (!preset || preset.presetId === "custom") {
      return;
    }
    const nextIds = normalizeLibrarySelection(preset.libraryIds);
    setLibraryPresetId(nextPresetId);
    updateFfmpegBuildSettings({ selectedLibraryIds: nextIds, licenseProfileName: deriveLicenseBoundaryFromSelectedLibraries(nextIds, initialApplicationState.defaultLibraryCatalog) });
  }

  function toggleConfigureOption(optionId: string) {
    const option = initialApplicationState.defaultConfigureOptionCatalog.find((candidateOption) => candidateOption.optionId === optionId);
    if (option?.locked) {
      return;
    }
    const currentIds = ffmpegBuildSettings.selectedConfigureOptionIds;
    const nextIds = currentIds.includes(optionId) ? currentIds.filter((currentId) => currentId !== optionId) : [...currentIds, optionId];
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: nextIds });
  }


  function restoreRecommendedToolchainPackages() {
    setMsys2PackageText(initialApplicationState.defaultBuildToolSettings.msys2PackageNames.join("\n"));
    setToolchainPreparationPlanReview(null);
  }

  function restoreRecommendedExtraFlags() {
    const nextFlags = initialApplicationState.defaultFfmpegBuildSettings.extraConfigureFlags || [];
    const nextOptionIds = initialApplicationState.defaultFfmpegBuildSettings.selectedConfigureOptionIds || ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"];
    setExtraConfigureFlagText(nextFlags.join("\n"));
    updateFfmpegBuildSettings({ selectedConfigureOptionIds: nextOptionIds, extraConfigureFlags: nextFlags, configureFlags: nextFlags });
  }

  const selectedLibraryCount = ffmpegBuildSettings.selectedLibraryIds.length;
  const tabItems: { id: TabId; label: string; description: string }[] = [
    { id: "source", label: "Source", description: "Downloads and signatures" },
    { id: "buildTools", label: "Build Tools", description: "Private toolchain" },
    { id: "prep", label: "Prep", description: "Review build tools" },
    { id: "library", label: "FFmpeg Libraries", description: `${selectedLibraryCount} selected` },
    { id: "options", label: "FFmpeg Options", description: "Build settings" },
    { id: "buildFfmpeg", label: "Build FFmpeg", description: approvedActionStatus },
    { id: "result", label: "Result", description: "Output files" },
    { id: "logs", label: "Logs", description: `${securityLogEntries.length} entries` },
  ];
  return (
    <main className="app-shell">
      <aside className="left-nav" aria-label="Build stages">
        <div className="left-nav__brand">CustomFFmpeg</div>
        <p className="left-nav__rule">Frontend reviews. Backend confirms.</p>
        <nav className="left-nav__items">
          {tabItems.map((tabItem) => (
            <button className={`left-nav__item ${activeTabId === tabItem.id ? "left-nav__item--active" : ""}`} key={tabItem.id} type="button" onClick={() => setActiveTabId(tabItem.id)}>
              <span className="left-nav__label">{tabItem.label}</span>
              <span className="left-nav__description">{tabItem.description}</span>
            </button>
          ))}
        </nav>
      </aside>

      <div className="tab-right-column">
      <section className="tab-panel" ref={tabPanelRef}>
        {activeTabId === "source" && (
          <section className="tab-page">
            <PageHeader title="Source" text="Choose the two files the app will download. Nothing installs or builds from this tab." />

            <InfoBox title="Plain explanation">
              <p>You need one MSYS2 archive and one FFmpeg source archive. MSYS2 is the private build toolbox. FFmpeg source is the code that will be compiled.</p>
              <p>For MSYS2, use the official .sig signature file instead of hunting for a hash. The app verifies the signature internally and still calculates SHA-256 for the log.</p>
            </InfoBox>

            <InfoBox title="Important file rule">
              <p>Use the MSYS2 archive URL in the archive field and the matching .sig URL in the signature field. A .sig file is not readable text; do not paste its contents into any field.</p>
            </InfoBox>

            <label className="field">
              <span className="field__label">Workspace folder</span>
              <span className="field__hint">The private folder where downloads, tools, source files, logs, and build output will be stored.</span>
              <span className="field__row">
                <input className="field__input" value={buildToolSettings.workspaceDirectory} onChange={(event) => { updateBuildToolSettings({ workspaceDirectory: event.target.value }); updateFfmpegBuildSettings({ workspaceDirectory: event.target.value }); }} placeholder="Example: D:\ffmpeg-workspace" />
                <button className="button" type="button" onClick={chooseWorkspaceDirectory}>Browse</button>
              </span>
            </label>

            <section className="source-group">
              <h2 className="source-group__title">MSYS2 private tool environment</h2>
              <p className="source-group__text">MSYS2 provides pacman, make, compilers, and Unix-like build tools on Windows. The official .exe installer is valid MSYS2, but this app does not run installers. It uses a tar archive so it can verify the download and unpack it inside the selected workspace without changing your system install.</p>
              <div className="link-row">
                <ExternalLinkButton label="Open MSYS2 archive list in my browser" url={officialLinks.msys2ArchiveList} onOpen={openInUserBrowser} />
                <ExternalLinkButton label="Open MSYS2 notes in my browser" url={officialLinks.msys2InstallerDocs} onOpen={openInUserBrowser} />
              </div>
              <label className="field">
                <span className="field__label">MSYS2 base archive URL</span>
                <span className="field__hint">Use an MSYS2 tar archive. <code>.tar.zst</code> is recommended, and <code>.tar.xz</code> is accepted as a fallback. Do not use <code>.sig</code> here because it belongs in the signature field. Do not use <code>.exe</code> or <code>.sfx.exe</code> here: those are official MSYS2 files, but they are installer/self-extractor formats, and this app intentionally does not run installers.</span>
                <input className="field__input" value={buildToolSettings.msys2ArchiveUrl} onChange={(event) => updateMsys2ArchiveUrl(event.target.value)} placeholder="https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst or .tar.xz" />
              </label>

              <label className="field">
                <span className="field__label">MSYS2 signature URL</span>
                <span className="field__hint">Use the matching <code>.sig</code> file. The app downloads the MSYS2 public signing key, checks its fingerprint, and verifies the archive internally. You do not need to install GPG manually.</span>
                <input className="field__input" value={buildToolSettings.msys2ArchiveSignatureUrl} onChange={(event) => updateBuildToolSettings({ msys2ArchiveSignatureUrl: event.target.value })} placeholder="https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig" />
              </label>
              <label className="field">
                <span className="field__label">MSYS2 SHA-256, optional audit check</span>
                <span className="field__hint">This is optional. The app calculates and logs SHA-256 after download. Use .sig above for authenticity.</span>
                <input className="field__input field__input--mono" value={buildToolSettings.msys2ArchiveSha256Hash} onChange={(event) => updateBuildToolSettings({ msys2ArchiveSha256Hash: event.target.value })} placeholder="Optional: 64 hexadecimal characters; do not paste .sig text" />
              </label>
            </section>

            <section className="source-group">
              <h2 className="source-group__title">FFmpeg source code</h2>
              <p className="source-group__text">This is the FFmpeg source code. The app compiles it into your own FFmpeg executable. It is not a prebuilt Windows FFmpeg download.</p>
              <p className="source-group__text">FFmpeg publishes release tarballs with matching <code>.asc</code> PGP signatures. Use the <code>.asc</code> URL for authenticity. SHA-256 is kept only as an optional audit check.</p>
              <div className="link-row">
                <ExternalLinkButton label="Open FFmpeg download page in my browser" url={officialLinks.ffmpegDownload} onOpen={openInUserBrowser} />
                <ExternalLinkButton label="Open FFmpeg release archive in my browser" url={officialLinks.ffmpegReleaseIndex} onOpen={openInUserBrowser} />
                <ExternalLinkButton label="Open FFmpeg signing key in my browser" url={officialLinks.ffmpegSigningKey} onOpen={openInUserBrowser} />
              </div>
              <label className="field">
                <span className="field__label">FFmpeg source archive URL</span>
                <span className="field__hint">Use a source archive such as <code>ffmpeg-8.1.tar.xz</code>. Use the source-code release, not a Windows executable package.</span>
                <input className="field__input" value={ffmpegBuildSettings.ffmpegSourceArchiveUrl} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceArchiveUrl: event.target.value, ffmpegSourceSignatureUrl: event.target.value ? `${event.target.value}.asc` : "" })} placeholder="https://ffmpeg.org/releases/ffmpeg-8.1.tar.xz" />
              </label>
              <label className="field">
                <span className="field__label">FFmpeg source signature URL</span>
                <span className="field__hint">Use the matching <code>.asc</code> file, for example <code>ffmpeg-8.1.tar.xz.asc</code>. Do not paste the PGP signature text into this field.</span>
                <input className="field__input" value={ffmpegBuildSettings.ffmpegSourceSignatureUrl} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSignatureUrl: event.target.value })} placeholder="https://ffmpeg.org/releases/ffmpeg-8.1.tar.xz.asc" />
              </label>
              <label className="field">
                <span className="field__label">FFmpeg SHA-256, optional audit check</span>
                <span className="field__hint">Official FFmpeg releases are normally verified with the <code>.asc</code> signature above. Leave this empty unless you already have a trusted 64-character SHA-256 from another trusted source.</span>
                <input className="field__input field__input--mono" value={ffmpegBuildSettings.ffmpegSourceSha256Hash} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSha256Hash: event.target.value })} placeholder="Optional: 64 hexadecimal characters; do not paste .asc text" />
              </label>
            </section>
          </section>
        )}

        {activeTabId === "buildTools" && (
          <section className="tab-page">
            <PageHeader title="Build Tools" text="Choose the private Windows build tools installed inside the selected workspace." />

            <InfoBox title="Plain explanation">
              <p>This page chooses the tools used to compile FFmpeg. They are installed only into the workspace copy of MSYS2.</p>
              <p>In detail: these are MSYS2 pacman packages such as compilers, make, pkg-config, CMake, Ninja, NASM, and YASM. Codec libraries are selected in the Library tab, not here.</p>
            </InfoBox>

            <label className="field">
              <span className="field__label">Build shell</span>
              <span className="field__hint">Simple choice: keep <strong>ucrt64</strong>. Detailed meaning: this selects which MSYS2 compiler environment builds FFmpeg. <strong>ucrt64</strong> is the modern Windows default, <strong>mingw64</strong> is older MinGW-w64, and <strong>clang64</strong> uses Clang/LLVM.</span>
              <select className="field__input" value={buildToolSettings.windowsShellProfileName} onChange={(event) => { updateBuildToolSettings({ windowsShellProfileName: event.target.value }); updateFfmpegBuildSettings({ windowsShellProfileName: event.target.value }); }}>
                <option value="ucrt64">ucrt64 — recommended modern Windows build</option>
                <option value="mingw64">mingw64 — traditional MinGW-w64 build</option>
                <option value="clang64">clang64 — LLVM/Clang build</option>
              </select>
            </label>
            <label className="field">
              <span className="field__label">Build tools</span>
              <span className="field__hint">Simple choice: keep the recommended list. Detailed meaning: one MSYS2 pacman package per line, installed into the private MSYS2 folder.</span>
              <textarea className="field__textarea" rows={12} value={msys2PackageText} onChange={(event) => { setMsys2PackageText(event.target.value); setToolchainPreparationPlanReview(null); }} />
            </label>
          </section>
        )}

        {activeTabId === "library" && (
          <section className="tab-page">
            <PageHeader title="FFmpeg Libraries" text="Show what FFmpeg already includes, then choose extra external libraries." />
            <InfoBox title="Plain explanation">
              <p>The checked items at the top are FFmpeg's own built-in parts. They are shown because an empty-looking Library page can feel suspicious.</p>
              <p>External libraries are different: FFmpeg does not use them by default. Selecting one adds packages, configure flags, and sometimes license obligations.</p>
            </InfoBox>
            <InfoBox title="Detailed rule">
              <p>Official FFmpeg documentation says native decoders and encoders are enabled by default, while decoders/encoders requiring external libraries must be enabled manually with the corresponding <code>--enable-lib...</code> option.</p>
              <p>Locked checked rows mean “included by the normal FFmpeg source build.” Unchecked rows mean “extra external library; read the explanation before selecting.”</p>
            </InfoBox>
            <LibraryPresetSelector presets={libraryPresets} selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} />
            <LibraryList catalog={initialApplicationState.defaultLibraryCatalog} selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds} onToggleLibrary={toggleLibrary} />
          </section>
        )}

        {activeTabId === "options" && (
          <section className="tab-page">
            <PageHeader title="FFmpeg Options" text="Choose the general build settings that will be combined with the selected libraries." />
            <InfoBox title="Plain explanation">
              <p>This page controls how FFmpeg is built after the source and libraries are chosen.</p>
              <p>Use the defaults for a first build. Change license boundary only when the selected libraries require it, and leave advanced flags empty unless a needed option is missing from the checkboxes.</p>
            </InfoBox>
            <div className="field field--readonly">
              <span className="field__label">License boundary</span>
              <span className="field__hint">This is now automatic. It follows the selected libraries, so the user cannot accidentally choose LGPL while a GPL or nonfree library is selected.</span>
              <div className="readonly-value">{licenseBoundaryLabel(ffmpegBuildSettings.licenseProfileName)}</div>
              <span className="field__hint">No GPL/nonfree libraries selected → LGPL local. GPL libraries selected → GPL local. Nonfree libraries or <code>--enable-nonfree</code> → nonfree local.</span>
            </div>
            <label className="field">
              <span className="field__label">Build jobs</span>
              <span className="field__hint">How many compile jobs <code>make</code> may run at the same time. Higher is faster but uses more CPU and memory. Use 1 for safest behavior; use your CPU core count minus one for speed.</span>
              <input className="field__input" type="number" min="1" max="256" value={ffmpegBuildSettings.parallelJobCount} onChange={(event) => updateFfmpegBuildSettings({ parallelJobCount: Number(event.target.value) })} />
            </label>
            <ConfigureOptionList catalog={initialApplicationState.defaultConfigureOptionCatalog} selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds} onToggleOption={toggleConfigureOption} />
            <label className="field field--advanced">
              <span className="field__label">Advanced FFmpeg configure flags</span>
              <span className="field__hint">Optional escape hatch. Leave empty unless a needed FFmpeg configure flag is missing from the checkboxes above. One <code>./configure</code> flag per line. Review will show every manual flag before backend confirmation.</span>
              <textarea className="field__textarea" rows={5} value={extraConfigureFlagText} onChange={(event) => { setExtraConfigureFlagText(event.target.value); setFfmpegBuildPlanReview(null); }} placeholder="Only if missing above: --extra-cflags=..." />
            </label>
          </section>
        )}

        {activeTabId === "prep" && (
          <section className="tab-page">
            <PageHeader title="Prep" text="Review and run the private build tools plan before choosing FFmpeg libraries." />
            <InfoBox title="Why this step exists">
              <p>Installing or reinstalling build tools can reset FFmpeg plans. This page keeps the build-tools plan separate, so the FFmpeg library and option choices come after the toolchain is ready.</p>
            </InfoBox>
            {toolchainPreparationPlanReview ? (
              <ApprovalPanel
                title="Build tools plan"
                actionName={toolchainPreparationPlanReview.plan.actionName}
                planHash={toolchainPreparationPlanReview.plan.planHash}
                expectedConsentText={toolchainPreparationPlanReview.expectedConsentText}
                operations={toolchainPreparationPlanReview.plan.operations}
                warnings={toolchainPreparationPlanReview.plan.warnings}
                isExecutable={toolchainPreparationPlanReview.plan.isExecutable}
                onRequestBackendConfirmation={approveToolchainPreparationPlan}
              />
            ) : <EmptyReview text="No build tools plan has been added yet. Go to Build Tools and press Add Build Plan and Continue to Prep." />}
            {!toolchainPreparationPlanReview && (
              <>
                <div className="status-box">
                  <span className="status-box__label">Current build tools status</span>
                  <strong className="status-box__value">{approvedActionPhase === "toolchain" ? approvedActionStatus : "idle"}</strong>
                </div>
                <button className="button button--danger" type="button" disabled={!canCancelToolchain} onClick={cancelApprovedAction}>Cancel Running Actions</button>
                {canCancelToolchain && securityLogEntries.length > 0 && (
                  <div className="recent-log">
                    <span className="recent-log__label">Recent activity</span>
                    {securityLogEntries.slice(-3).map((entry, index) => (
                      <p className={`log-list__entry log-list__entry--${entry.level}`} key={`prep-recent-${index}`}><strong>{entry.level}</strong><time className="log-list__time">{entry.timestamp}</time><span>{entry.message}</span></p>
                    ))}
                  </div>
                )}
              </>
            )}
          </section>
        )}

        {activeTabId === "buildFfmpeg" && (
          <section className="tab-page">
            <PageHeader title="Build FFmpeg" text="Review the FFmpeg plan, request backend confirmation, and watch the build status." />
            {ffmpegBuildPlanReview ? (
              <ApprovalPanel
                title="FFmpeg build plan"
                actionName={ffmpegBuildPlanReview.plan.actionName}
                planHash={ffmpegBuildPlanReview.plan.planHash}
                expectedConsentText={ffmpegBuildPlanReview.expectedConsentText}
                operations={ffmpegBuildPlanReview.plan.operations}
                warnings={ffmpegBuildPlanReview.plan.warnings}
                isExecutable={ffmpegBuildPlanReview.plan.isExecutable}
                selectedLibraries={ffmpegBuildPlanReview.plan.selectedLibraries}
                requiredMsys2PackageNames={ffmpegBuildPlanReview.plan.requiredMsys2PackageNames}
                generatedConfigureFlags={ffmpegBuildPlanReview.plan.generatedConfigureFlags}
                selectedConfigureOptions={ffmpegBuildPlanReview.plan.selectedConfigureOptions}
                generatedOptionFlags={ffmpegBuildPlanReview.plan.generatedOptionFlags}
                extraConfigureFlags={ffmpegBuildPlanReview.plan.extraConfigureFlags}
                finalConfigureFlags={ffmpegBuildPlanReview.plan.configureFlags}
                onRequestBackendConfirmation={approveFfmpegBuildPlan}
              />
            ) : <EmptyReview text="No FFmpeg plan has been reviewed yet. Go to FFmpeg Options and press Review the Plan." />}
            {!ffmpegBuildPlanReview && (
              <>
                <div className="status-box">
                  <span className="status-box__label">Current FFmpeg build status</span>
                  <strong className="status-box__value">{approvedActionPhase === "ffmpeg" ? approvedActionStatus : "idle"}</strong>
                </div>
                <button className="button button--danger" type="button" disabled={!canCancelFfmpeg} onClick={cancelApprovedAction}>Cancel Running Actions</button>
                {canCancelFfmpeg && securityLogEntries.length > 0 && (
                  <div className="recent-log">
                    <span className="recent-log__label">Recent activity</span>
                    {securityLogEntries.slice(-3).map((entry, index) => (
                      <p className={`log-list__entry log-list__entry--${entry.level}`} key={`ffmpeg-recent-${index}`}><strong>{entry.level}</strong><time className="log-list__time">{entry.timestamp}</time><span>{entry.message}</span></p>
                    ))}
                  </div>
                )}
              </>
            )}
          </section>
        )}

        {activeTabId === "result" && (
          <section className="tab-page">
            <PageHeader title="Result" text="Final FFmpeg output files and what was included in the build." />
            <ResultPanel result={buildResult} errorText={buildResultError} onRefresh={refreshBuildResult} onOpenFolder={openResultFolder} />
          </section>
        )}

        {activeTabId === "logs" && (
          <section className="tab-page">
            <PageHeader title="Logs" text="Security and progress messages from backend actions." />
            <div className="log-list" aria-live="polite">
              {securityLogEntries.length === 0 ? <p className="empty-text">No approved action has run yet.</p> : [...securityLogEntries].reverse().map((entry, index) => (
                <p className={`log-list__entry log-list__entry--${entry.level}`} key={`${entry.level}-${index}-${entry.message}`}><strong>{entry.level}</strong><time className="log-list__time">{entry.timestamp}</time><span>{entry.message}</span></p>
              ))}
            </div>
          </section>
        )}
      </section>
      <div className="tab-bottom-bar">
        {activeTabId === "source" && (
          <button className="button button--primary" type="button" onClick={() => setActiveTabId("buildTools")}>Choose Build Tools</button>
        )}
        {activeTabId === "buildTools" && (
          <>
            <button className="button button--primary" type="button" onClick={addBuildToolsPlanAndContinueToPrep}>Add Build Plan and Continue to Prep</button>
            <button className="button" type="button" onClick={restoreRecommendedToolchainPackages}>Restore Recommended List</button>
          </>
        )}
        {activeTabId === "prep" && (
          <button className="button button--primary" type="button" onClick={() => setActiveTabId("library")}>Choose FFmpeg Libraries</button>
        )}
        {activeTabId === "library" && (
          <button className="button button--primary" type="button" onClick={() => setActiveTabId("options")}>Continue to FFmpeg Options</button>
        )}
        {activeTabId === "options" && (
          <>
            <button className="button button--primary" type="button" onClick={reviewFfmpegPlans}>Review the Plan</button>
            <button className="button" type="button" onClick={restoreRecommendedExtraFlags}>Restore Recommended Options</button>
          </>
        )}
        {activeTabId === "prep" && toolchainPreparationPlanReview && (
          <button className="button button--primary" type="button" disabled={!toolchainPreparationPlanReview.plan.isExecutable} onClick={approveToolchainPreparationPlan}>Request Backend Confirmation</button>
        )}
        {activeTabId === "buildFfmpeg" && ffmpegBuildPlanReview && (
          <button className="button button--primary" type="button" disabled={!ffmpegBuildPlanReview.plan.isExecutable} onClick={approveFfmpegBuildPlan}>Request Backend Confirmation</button>
        )}
      </div>
      </div>
    </main>
  );
}

function ResultPanel(props: { result: BuildResult | null; errorText: string; onRefresh: () => Promise<void>; onOpenFolder: () => Promise<void> }) {
  const result = props.result;
  return (
    <section className="result-panel">
      <div className="actions">
        <button className="button button--primary" type="button" onClick={props.onOpenFolder}>Open Result Folder</button>
        <button className="button" type="button" onClick={props.onRefresh}>Refresh Result</button>
      </div>
      {props.errorText && <p className="empty-text">{props.errorText}</p>}
      {!props.errorText && !result && <p className="empty-text">No build result has been found yet.</p>}
      {result && (
        <>
          <dl className="metadata">
            <dt>Result folder</dt><dd className="metadata__hash">{result.artifactsDirectory}</dd>
            <dt>Latest report</dt><dd className="metadata__hash">{result.reportPath || "No build report found yet."}</dd>
          </dl>
          <section className="result-files">
            <h2 className="review-list__title">Files</h2>
            {result.files.length === 0 ? <p className="empty-text">No FFmpeg executable has been copied to the result folder yet.</p> : result.files.map((file) => (
              <div className="result-file" key={file.path}>
                <strong>{file.name}</strong>
                <span>{formatBytes(file.sizeBytes)}</span>
                <span className="metadata__hash">{file.path}</span>
                {file.sha256Hash && <span className="metadata__hash">SHA-256: {file.sha256Hash}</span>}
              </div>
            ))}
          </section>
          {result.selectedLibraries.length > 0 && <ReviewList title="Libraries included by this build plan" items={result.selectedLibraries} />}
          {result.selectedConfigureOptions.length > 0 && <ReviewList title="FFmpeg options included by this build plan" items={result.selectedConfigureOptions} />}
          {result.configureFlags.length > 0 && <ReviewList title="Final configure flags" items={result.configureFlags} />}
          {result.requiredMsys2PackageNames.length > 0 && <ReviewList title="Library packages installed for this build" items={result.requiredMsys2PackageNames} />}
        </>
      )}
    </section>
  );
}

function formatBytes(sizeBytes: number): string {
  if (sizeBytes < 1024) {
    return `${sizeBytes} B`;
  }
  if (sizeBytes < 1024 * 1024) {
    return `${(sizeBytes / 1024).toFixed(1)} KB`;
  }
  if (sizeBytes < 1024 * 1024 * 1024) {
    return `${(sizeBytes / 1024 / 1024).toFixed(1)} MB`;
  }
  return `${(sizeBytes / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

function ExternalLinkButton(props: { label: string; url: string; onOpen: (urlToOpen: string) => Promise<void> }) {
  return (
    <button className="button button--link" type="button" onClick={() => props.onOpen(props.url)}>
      {props.label}
    </button>
  );
}

function PageHeader(props: { title: string; text: string }) {
  return (
    <header className="page-header">
      <h1 className="page-header__title">{props.title}</h1>
      <p className="page-header__text">{props.text}</p>
    </header>
  );
}

function InfoBox(props: { title: string; children: React.ReactNode }) {
  return (
    <section className="info-box">
      <h2 className="info-box__title">{props.title}</h2>
      <div className="info-box__body">{props.children}</div>
    </section>
  );
}

function LibraryPresetSelector(props: { presets: LibraryPreset[]; selectedPresetId: LibraryPresetId; onApplyPreset: (presetId: LibraryPresetId) => void }) {
  return (
    <section className="preset-panel">
      <div className="preset-panel__header">
        <h2 className="preset-panel__title">Library Preset</h2>
        <p className="preset-panel__text">Choose a starting point. You can still check or uncheck individual libraries after applying a preset.</p>
      </div>
      <div className="preset-grid">
        {props.presets.filter((preset) => preset.presetId !== "custom").map((preset) => (
          <button className={`preset-card ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <span className="preset-card__name">{preset.displayName}</span>
            <span className="preset-card__plain">{preset.plainExplanation}</span>
            <span className="preset-card__technical">{preset.technicalExplanation}</span>
          </button>
        ))}
      </div>
      {props.selectedPresetId === "custom" && <p className="preset-panel__custom">Custom selection. The checked libraries no longer match one preset exactly.</p>}
    </section>
  );
}

function LibraryList(props: { catalog: LibraryChoice[]; selectedLibraryIds: string[]; onToggleLibrary: (libraryId: string) => void }) {
  const groupedLibraries = groupLibrariesByCategory(props.catalog);
  return (
    <div className="library-list">
      {Object.entries(groupedLibraries).map(([categoryName, libraries]) => (
        <section className="library-group" key={categoryName}>
          <h2 className="library-group__title">{categoryName}</h2>
          {libraries.map((library) => {
            const isChecked = props.selectedLibraryIds.includes(library.libraryId) || library.defaultChecked;
            const isExternal = !library.defaultChecked;
            return (
              <label className={`library-row ${library.locked ? "library-row--locked" : ""}`} key={library.libraryId}>
                <input type="checkbox" checked={isChecked} disabled={library.locked} onChange={() => props.onToggleLibrary(library.libraryId)} />
                <span className="library-row__main">
                  <span className="library-row__name">{library.displayName}</span>
                  {isExternal ? (
                    <>
                      <span className="library-row__note">{library.plainExplanation || library.reviewNote}</span>
                      <span className="library-row__detail library-row__detail--why">{library.technicalExplanation}</span>
                    </>
                  ) : (
                    <>
                      <span className="library-row__note">{library.plainExplanation || library.reviewNote}</span>
                      <span className="library-row__detail">{library.technicalExplanation}</span>
                    </>
                  )}
                  {library.configureFlags.length > 0 && <span className="library-row__detail"><strong>Flags:</strong> {library.configureFlags.join(" ")}</span>}
                  {library.packageNames.length > 0 && <span className="library-row__detail"><strong>Packages:</strong> {library.packageNames.join(", ")}</span>}
                </span>
                <span className={`library-row__license library-row__license--${library.licenseEffectName}`}>{library.defaultChecked ? "included" : library.licenseEffectName}</span>
              </label>
            );
          })}
        </section>
      ))}
    </div>
  );
}

function ConfigureOptionList(props: { catalog: ConfigureOptionChoice[]; selectedOptionIds: string[]; onToggleOption: (optionId: string) => void }) {
  const groupedOptions = groupConfigureOptionsByCategory(props.catalog);
  return (
    <div className="option-list">
      {Object.entries(groupedOptions).map(([categoryName, options]) => (
        <section className="option-group" key={categoryName}>
          <h2 className="option-group__title">{categoryName}</h2>
          {options.map((option) => (
            <label className="option-row" key={option.optionId}>
              <input type="checkbox" checked={props.selectedOptionIds.includes(option.optionId)} disabled={option.locked} onChange={() => props.onToggleOption(option.optionId)} />
              <span className="option-row__main">
                <span className="option-row__name">{option.displayName}</span>
                <span className="option-row__plain">{option.plainExplanation}</span>
                <span className="option-row__technical">{option.technicalNote}</span>
                <span className="option-row__flags">{option.configureFlags.length > 0 ? `Flags: ${option.configureFlags.join(" ")}` : "Default FFmpeg behavior; no extra flag is added."}</span>
              </span>
            </label>
          ))}
        </section>
      ))}
    </div>
  );
}

function EmptyReview(props: { text: string }) {
  return <p className="empty-text">{props.text}</p>;
}

type ApprovalPanelProps = {
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

function ApprovalPanel(props: ApprovalPanelProps) {
  return (
    <section className="approval-panel">
      <h2 className="approval-panel__title">{props.title}</h2>
      <p className="approval-panel__summary">Nothing runs from this panel. This only asks the backend to show its own confirmation dialog.</p>
      <dl className="metadata">
        <dt>Action</dt><dd>{props.actionName}</dd>
        <dt>Plan hash</dt><dd className="metadata__hash">{props.planHash}</dd>
        <dt>Backend phrase</dt><dd>{props.expectedConsentText}</dd>
      </dl>
      {props.selectedLibraries && props.selectedLibraries.length > 0 && (
        <ReviewList title="Selected libraries" items={props.selectedLibraries.map((library) => `${library.displayName} | ${library.licenseEffectName} | ${library.configureFlags.join(" ")}`)} />
      )}
      {props.requiredMsys2PackageNames && props.requiredMsys2PackageNames.length > 0 && <ReviewList title="Required library packages" items={props.requiredMsys2PackageNames} />}
      {props.generatedConfigureFlags && props.generatedConfigureFlags.length > 0 && <ReviewList title="Generated library flags" items={props.generatedConfigureFlags} />}
      {props.selectedConfigureOptions && props.selectedConfigureOptions.length > 0 && <ReviewList title="Selected built-in FFmpeg options" items={props.selectedConfigureOptions.map((option) => `${option.displayName} | ${option.configureFlags.join(" ")}`)} />}
      {props.generatedOptionFlags && props.generatedOptionFlags.length > 0 && <ReviewList title="Generated option flags" items={props.generatedOptionFlags} />}
      {props.extraConfigureFlags && props.extraConfigureFlags.length > 0 && <ReviewList title="Advanced manual flags" items={props.extraConfigureFlags} />}
      {props.finalConfigureFlags && props.finalConfigureFlags.length > 0 && <ReviewList title="Final configure flags" items={props.finalConfigureFlags} />}
      <ReviewList title="Operations" items={props.operations.map((operation) => operation.summary)} />
      {props.warnings.length > 0 && <WarningList warnings={props.warnings} />}
    </section>
  );
}

function ReviewList(props: { title: string; items: string[] }) {
  return (
    <section className="review-list">
      <h3 className="review-list__title">{props.title}</h3>
      <ul className="review-list__items">
        {props.items.map((item) => <li className="review-list__item" key={item}>{item}</li>)}
      </ul>
    </section>
  );
}

function WarningList(props: { warnings: PlanWarning[] }) {
  return (
    <section className="review-list">
      <h3 className="review-list__title">Warnings</h3>
      <ul className="review-list__items">
        {props.warnings.map((warning, index) => <li className={`review-list__item review-list__item--${warning.riskLevelName}`} key={`${warning.riskLevelName}-${index}`}>{warning.message}</li>)}
      </ul>
    </section>
  );
}

function groupConfigureOptionsByCategory(catalog: ConfigureOptionChoice[]) {
  return catalog.reduce<Record<string, ConfigureOptionChoice[]>>((groups, option) => {
    const categoryName = option.categoryName || "Other";
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(option);
    return groups;
  }, {});
}

function groupLibrariesByCategory(catalog: LibraryChoice[]) {
  return catalog.reduce<Record<string, LibraryChoice[]>>((groups, library) => {
    const categoryName = library.categoryName || "Other";
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}

function createApprovalRequest(actionName: string, planHash: string, consentText: string): ApprovalRequest {
  return { approvedActionName: actionName, approvedPlanHash: planHash, consentText };
}


const baseIncludedLibraryIds = [
  "ffmpeg-program",
  "ffprobe-program",
  "libavcodec",
  "libavformat",
  "libavfilter",
  "libavutil",
  "libswscale",
  "libswresample",
  "native-codecs",
  "native-formats",
];

const libraryPresets: LibraryPreset[] = [
  {
    presetId: "default",
    displayName: "Default",
    plainExplanation: "Use FFmpeg's normal source build with its built-in programs, libraries, codecs, and formats.",
    technicalExplanation: "No external codec or multimedia package is added from this preset.",
    libraryIds: baseIncludedLibraryIds,
  },
  {
    presetId: "efficiency",
    displayName: "Maximum Audio/Video Efficiency",
    plainExplanation: "Adds audio and video codecs that are often chosen because they encode more efficiently or produce better quality than the built-in defaults.",
    technicalExplanation: "Adds H.264/H.265, VP9/AV1, AAC, Opus, MP3, fast AV1 decoding, and high-quality resampling. The preset follows feature goals first; the derived license boundary is shown later in Build FFmpeg.",
    libraryIds: [...baseIncludedLibraryIds, "x264", "x265", "libvpx", "aom", "svt-av1", "dav1d", "fdk-aac", "opus", "mp3lame", "soxr"],
  },
  {
    presetId: "compatibility",
    displayName: "Maximum Audio/Video Compatibility",
    plainExplanation: "Adds as many common audio and video codecs as possible so the build can open or create more media formats.",
    technicalExplanation: "Adds the efficiency set plus additional speech, legacy, WebM/Vorbis, AMR, AAC, and H.264/H.265 compatibility libraries. This preset follows compatibility first; the derived license boundary is shown later in Build FFmpeg.",
    libraryIds: [...baseIncludedLibraryIds, "x264", "x265", "libvpx", "aom", "svt-av1", "rav1e", "dav1d", "openh264", "fdk-aac", "vorbis", "mp3lame", "twolame", "opus", "soxr", "speex", "gsm", "ilbc", "opencore-amr", "vo-amrwbenc"],
  },
  {
    presetId: "editor",
    displayName: "Audio/Video Editor",
    plainExplanation: "Starts from maximum compatibility and adds libraries useful for editing, filtering, retiming, subtitle work, image workflows, quality checks, and high-quality video processing.",
    technicalExplanation: "Adds text rendering, subtitle rendering, image codecs, high-quality scaling/colorspace conversion, libplacebo, frei0r effects, Rubber Band audio retiming/pitch shifting, VMAF, and common processing helpers. The preset follows editor usefulness first; the derived license boundary is shown later in Build FFmpeg.",
    libraryIds: [...baseIncludedLibraryIds, "x264", "x265", "libvpx", "aom", "svt-av1", "rav1e", "dav1d", "openh264", "fdk-aac", "vorbis", "mp3lame", "twolame", "opus", "soxr", "speex", "gsm", "ilbc", "opencore-amr", "vo-amrwbenc", "libjxl", "openjpeg", "webp", "png", "zimg", "libplacebo", "vmaf", "frei0r", "rubberband", "freetype", "fontconfig", "fribidi", "harfbuzz", "ass", "xml2"],
  },
  {
    presetId: "full",
    displayName: "Full",
    plainExplanation: "Selects the broadest set this app can plan without selecting mutually exclusive alternatives together.",
    technicalExplanation: "Includes editor features plus disc/device input, streaming/network protocols, OCR, OpenSSL TLS, and FDK AAC. GnuTLS is left unchecked only because FFmpeg cannot enable GnuTLS and OpenSSL at the same time.",
    libraryIds: [...baseIncludedLibraryIds, "x264", "x265", "libvpx", "aom", "svt-av1", "rav1e", "dav1d", "openh264", "xavs2", "libjxl", "openjpeg", "webp", "png", "zimg", "libplacebo", "vmaf", "frei0r", "rubberband", "opus", "vorbis", "mp3lame", "twolame", "soxr", "speex", "gsm", "ilbc", "opencore-amr", "vo-amrwbenc", "fdk-aac", "freetype", "fontconfig", "fribidi", "harfbuzz", "ass", "bluray", "cdio", "modplug", "openal", "sdl2", "openssl", "srt", "ssh", "zmq", "rist", "xml2", "tesseract"],
  },
];

function normalizeLibrarySelection(selectedLibraryIds: string[]): string[] {
  const selectedSet = new Set<string>([...baseIncludedLibraryIds, ...selectedLibraryIds]);
  if (selectedSet.has("openssl") && selectedSet.has("gnutls")) {
    selectedSet.delete("gnutls");
  }
  return Array.from(selectedSet);
}

function matchLibraryPresetId(selectedLibraryIds: string[]): LibraryPresetId {
  const normalizedSelection = normalizeLibrarySelection(selectedLibraryIds).slice().sort();
  for (const preset of libraryPresets) {
    if (preset.presetId === "custom") {
      continue;
    }
    const normalizedPreset = normalizeLibrarySelection(preset.libraryIds).slice().sort();
    if (normalizedSelection.length === normalizedPreset.length && normalizedSelection.every((libraryId, index) => libraryId === normalizedPreset[index])) {
      return preset.presetId;
    }
  }
  return "custom";
}

function isValidLibraryPresetId(value: unknown): value is LibraryPresetId {
  return value === "default" || value === "efficiency" || value === "compatibility" || value === "editor" || value === "full" || value === "custom";
}

function removeMutuallyExclusiveLibraries(selectedLibraryIds: string[], selectedLibraryId: string): string[] {
  const exclusiveGroups: Record<string, string[]> = {
    openssl: ["gnutls"],
    gnutls: ["openssl"],
  };
  const conflicts = exclusiveGroups[selectedLibraryId] ?? [];
  if (conflicts.length === 0) {
    return selectedLibraryIds;
  }
  return selectedLibraryIds.filter((libraryId) => libraryId === selectedLibraryId || !conflicts.includes(libraryId));
}

function deriveLicenseBoundaryFromSelectedLibraries(selectedLibraryIds: string[], catalog: LibraryChoice[]): string {
  const selectedLibraries = catalog.filter((library) => selectedLibraryIds.includes(library.libraryId));
  if (selectedLibraries.some((library) => library.licenseEffectName === "nonfree")) {
    return "nonfree-local";
  }
  if (selectedLibraries.some((library) => library.licenseEffectName === "gpl")) {
    return "gpl-local";
  }
  return "lgpl-local";
}

function licenseBoundaryLabel(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return "GPL local — required by selected GPL libraries";
    case "nonfree-local":
      return "Nonfree local — required by selected nonfree items";
    case "lgpl-local":
    default:
      return "LGPL local — current selected libraries do not require GPL/nonfree";
  }
}

function splitLines(value: string): string[] {
  return value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

function normalizeLogLevel(value: string): "info" | "warn" | "error" {
  if (value === "warn" || value === "error") {
    return value;
  }
  return "info";
}



function isValidTabId(value: unknown): value is TabId {
  return value === "source" || value === "buildTools" || value === "prep" || value === "library" || value === "options" || value === "buildFfmpeg" || value === "result" || value === "logs";
}

function readSavedUiState(): SavedUiState {
  try {
    const rawValue = window.localStorage.getItem(savedUiStateKey);
    if (!rawValue) {
      return {};
    }
    const parsedValue = JSON.parse(rawValue) as SavedUiState;
    return parsedValue && typeof parsedValue === "object" ? parsedValue : {};
  } catch {
    return {};
  }
}

function saveUiState(nextState: SavedUiState) {
  try {
    window.localStorage.setItem(savedUiStateKey, JSON.stringify(nextState));
  } catch {
    // Persistence is helpful but must never block the build workflow.
  }
}

function readSavedWindowState(): SavedWindowState {
  try {
    const rawValue = window.localStorage.getItem(savedWindowStateKey);
    if (!rawValue) {
      return {};
    }
    const parsedValue = JSON.parse(rawValue) as SavedWindowState;
    return parsedValue && typeof parsedValue === "object" ? parsedValue : {};
  } catch {
    return {};
  }
}

async function restoreWindowState() {
  const savedWindowState = readSavedWindowState();
  if (Number.isFinite(savedWindowState.width) && Number.isFinite(savedWindowState.height)) {
    await WindowSetSize(Number(savedWindowState.width), Number(savedWindowState.height));
  }
  if (Number.isFinite(savedWindowState.x) && Number.isFinite(savedWindowState.y)) {
    await WindowSetPosition(Number(savedWindowState.x), Number(savedWindowState.y));
  }
}

async function saveWindowState() {
  try {
    const runtimeSize = await WindowGetSize();
    const runtimePosition = await WindowGetPosition();
    const width = readRuntimeNumber(runtimeSize, 0, "w", "width");
    const height = readRuntimeNumber(runtimeSize, 1, "h", "height");
    const x = readRuntimeNumber(runtimePosition, 0, "x", "left");
    const y = readRuntimeNumber(runtimePosition, 1, "y", "top");
    window.localStorage.setItem(savedWindowStateKey, JSON.stringify({ width, height, x, y } satisfies SavedWindowState));
  } catch {
    // Window persistence is best-effort. Never block the app if the runtime is unavailable.
  }
}

function readRuntimeNumber(value: unknown, tupleIndex: number, primaryPropertyName: string, fallbackPropertyName: string): number {
  if (Array.isArray(value)) {
    return Number(value[tupleIndex]);
  }
  if (value && typeof value === "object") {
    const recordValue = value as Record<string, unknown>;
    return Number(recordValue[primaryPropertyName] ?? recordValue[fallbackPropertyName] ?? 0);
  }
  return 0;
}

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BuilderApp />
  </React.StrictMode>,
);
