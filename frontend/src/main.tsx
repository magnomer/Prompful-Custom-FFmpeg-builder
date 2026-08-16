import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";

import { PSourceRender } from "./tabs/source";
import { PConfigRender } from "./tabs/buildconfig";
import { PPrepRender } from "./tabs/prep";
import { PLibraryRender } from "./tabs/libraries";
import { POptionRender } from "./tabs/options";
import { PBuildRender } from "./tabs/ffmpegbuild";
import { PResultRender } from "./tabs/result";
import { PLogRender } from "./tabs/logs";
import { PAboutRender } from "./tabs/about";
import { LStateBuilderUse } from "./useBuilderState";
import type { LTabIdentifier } from "./programstate";
import { LLocaleGet, LLocaleSet, LLocaleSynchronize, LLocaleTextGet, LLocaleStatusGet, type LLocaleCode } from "./i18n";
import sourceIcon from "./assets/tab-icons/Source.svg";
import buildConfigurationIcon from "./assets/tab-icons/BuildConfiguration.svg";
import prepIcon from "./assets/tab-icons/Prep.svg";
import ffmpegLibrariesIcon from "./assets/tab-icons/FfmpegLibraries.svg";
import ffmpegOptionsIcon from "./assets/tab-icons/FfmpegOptions.svg";
import buildFfmpegIcon from "./assets/tab-icons/BuildFfmpeg.svg";
import resultIcon from "./assets/tab-icons/Result.svg";
import logsIcon from "./assets/tab-icons/Logs.svg";
import aboutIcon from "./assets/tab-icons/About.svg";
import programIcon from "../../build/appicon.png";
import nextStepIcon from "./assets/footer-icons/NextStep.svg";

type LErrorBoundaryState = {
  errorText: string;
};

class PErrorBoundary extends React.Component<React.PropsWithChildren, LErrorBoundaryState> {
  state: LErrorBoundaryState = { errorText: "" };

  static getDerivedStateFromError(error: unknown): LErrorBoundaryState {
    return { errorText: LErrorTextFormat(error) };
  }

  componentDidCatch(error: unknown, errorInfo: React.ErrorInfo) {
    console.error("Unhandled render error", error, errorInfo.componentStack);
  }

  render() {
    if (!this.state.errorText) return this.props.children;
    return <PFatalErrorRender title={LLocaleTextGet("fatal.renderFailed")} text={this.state.errorText} />;
  }
}

function LErrorTextFormat(error: unknown): string {
  if (error instanceof Error) return `${error.name}: ${error.message}\n${error.stack ?? ""}`.trim();
  return String(error);
}

function PFatalErrorRender(props: { title: string; text: string }) {
  return (
    <main className="program-fatal">
      <section className="program-fatal__panel">
        <h1>{props.title}</h1>
        <pre>{props.text}</pre>
      </section>
    </main>
  );
}

function PApprovalConfirmationRender(props: { request: { actionName: string; planHash: string }; onResolve: (approved: boolean) => Promise<void> }) {
  const rejectButtonRef = React.useRef<HTMLButtonElement>(null);
  const [isResolving, setIsResolving] = React.useState(false);
  const resolve = React.useCallback(async (approved: boolean) => {
    if (isResolving) return;
    setIsResolving(true);
    await props.onResolve(approved);
  }, [isResolving, props]);
  React.useEffect(() => {
    rejectButtonRef.current?.focus();
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") void resolve(false);
    };
    window.addEventListener("keydown", onKeyDown);
    return () => window.removeEventListener("keydown", onKeyDown);
  }, [resolve]);
  const action = LLocaleTextGet(`approval.action.${props.request.actionName}`);
  return (
    <div className="approval-confirmation-backdrop">
      <section className="approval-confirmation" role="alertdialog" aria-modal="true" aria-labelledby="approval-confirmation-title" aria-describedby="approval-confirmation-message">
        <h2 id="approval-confirmation-title">{LLocaleTextGet("native.approval.title")}</h2>
        <p id="approval-confirmation-message">{LLocaleTextGet("native.approval.message", { action, planHash: props.request.planHash })}</p>
        <div className="approval-confirmation__actions">
          <button ref={rejectButtonRef} className="button" type="button" disabled={isResolving} onClick={() => void resolve(false)}>{LLocaleTextGet("native.approval.no")}</button>
          <button className="button button--primary" type="button" disabled={isResolving} onClick={() => void resolve(true)}>{LLocaleTextGet("native.approval.yes")}</button>
        </div>
      </section>
    </div>
  );
}

// ─── PProgramRender ───────────────────────────────────────────────────────────────

function PProgramRender() {
  const [locale, LLocaleSetState] = React.useState<LLocaleCode>(() => LLocaleGet());
  React.useEffect(() => {
    const updateLocale = () => LLocaleSetState(LLocaleGet());
    globalThis.addEventListener("customffmpeg-locale-change", updateLocale);
    return () => globalThis.removeEventListener("customffmpeg-locale-change", updateLocale);
  }, []);
  React.useEffect(() => { document.title = LLocaleTextGet("app.brand"); }, [locale]);
  const s = LStateBuilderUse();
  const [isLocaleMenuOpen, setIsLocaleMenuOpen] = React.useState(false);
  const localeItems: { id: LLocaleCode; label: string }[] = [
    { id: "en", label: LLocaleTextGet("locale.english") },
    { id: "ko", label: LLocaleTextGet("locale.korean") },
  ];
  const selectedLocaleLabel = localeItems.find((localeItem) => localeItem.id === locale)?.label ?? LLocaleTextGet("locale.english");

  const selectedLibraryCount = s.ffmpegBuildSettings.selectedLibraryIds.length;
  const tabItems: { id: LTabIdentifier; label: string; description: string; icon: string }[] = [
    { id: "source",      label: LLocaleTextGet("nav.source.label"),      description: LLocaleTextGet("nav.source.description"), icon: sourceIcon },
    { id: "buildConfig",  label: LLocaleTextGet("nav.buildConfig.label"),  description: LLocaleTextGet("nav.buildConfig.description"), icon: buildConfigurationIcon },
    { id: "prep",        label: LLocaleTextGet("nav.prep.label"),        description: LLocaleTextGet("nav.prep.description"), icon: prepIcon },
    { id: "library",     label: LLocaleTextGet("nav.library.label"),     description: LLocaleTextGet("nav.library.description", { count: selectedLibraryCount }), icon: ffmpegLibrariesIcon },
    { id: "options",     label: LLocaleTextGet("nav.options.label"),     description: LLocaleTextGet("nav.options.description"), icon: ffmpegOptionsIcon },
    { id: "buildFfmpeg", label: LLocaleTextGet("nav.buildFfmpeg.label"), description: LLocaleStatusGet(s.approvedActionStatus), icon: buildFfmpegIcon },
    { id: "result",      label: LLocaleTextGet("nav.result.label"),      description: LLocaleTextGet("nav.result.description"), icon: resultIcon },
    { id: "logs",        label: LLocaleTextGet("nav.logs.label"),        description: LLocaleTextGet("nav.logs.description", { count: s.securityLogEntries.length }), icon: logsIcon },
  ];

  return (
    <main className="program-shell">
      <aside className="left-nav" aria-label={LLocaleTextGet("nav.ariaLabel")}>
        <div className="left-nav__brand">
          <span className="left-nav__logo" aria-hidden="true"><img src={programIcon} alt="" /></span>
          <span className="left-nav__brand-title">{LLocaleTextGet("app.brand")}</span>
        </div>
        <nav className="left-nav__items">
          {tabItems.map((tabItem) => (
            <button className={`left-nav__item ${s.activeTabId === tabItem.id ? "left-nav__item--active" : ""}`} key={tabItem.id} type="button" onClick={() => s.setActiveTabId(tabItem.id)}>
              <span className="left-nav__icon" aria-hidden="true"><img src={tabItem.icon} alt="" /></span>
              <span className="left-nav__item-text">
                <span className="left-nav__label">{tabItem.label}</span>
                <span className="left-nav__description">{tabItem.description}</span>
              </span>
            </button>
          ))}
        </nav>
        <div className="left-nav__bottom">
          <button className={`left-nav__item left-nav__item--about ${s.activeTabId === "about" ? "left-nav__item--active" : ""}`} type="button" onClick={() => s.setActiveTabId("about")}>
            <span className="left-nav__icon" aria-hidden="true"><img src={aboutIcon} alt="" /></span>
            <span className="left-nav__item-text">
              <span className="left-nav__label">{LLocaleTextGet("nav.about.label")}</span>
              <span className="left-nav__description">{LLocaleTextGet("nav.about.description")}</span>
            </span>
          </button>
          <div className="left-nav__locale-menu">
            <button
              className={`left-nav__locale-button ${isLocaleMenuOpen ? "left-nav__locale-button--open" : ""}`}
              type="button"
              aria-label={LLocaleTextGet("locale.selector.ariaLabel")}
              aria-haspopup="listbox"
              aria-expanded={isLocaleMenuOpen}
              onClick={() => setIsLocaleMenuOpen((isOpen) => !isOpen)}
            >
              <span className="left-nav__locale-globe" aria-hidden="true" />
              <span className="left-nav__locale-label">{selectedLocaleLabel}</span>
              <span className="left-nav__locale-chevron" aria-hidden="true" />
            </button>
            {isLocaleMenuOpen && (
              <div className="left-nav__locale-popover" role="listbox" aria-label={LLocaleTextGet("locale.selector.ariaLabel")}>
                {localeItems.map((localeItem) => (
                  <button
                    className={`left-nav__locale-option ${locale === localeItem.id ? "left-nav__locale-option--selected" : ""}`}
                    key={localeItem.id}
                    type="button"
                    role="option"
                    aria-selected={locale === localeItem.id}
                    onClick={async () => {
                      setIsLocaleMenuOpen(false);
                      await LLocaleSet(localeItem.id);
                    }}
                  >
                    <span>{localeItem.label}</span>
                    {locale === localeItem.id && <span className="left-nav__locale-check" aria-hidden="true">✓</span>}
                  </button>
                ))}
              </div>
            )}
          </div>
        </div>
      </aside>

      <div className="tab-right-column">
        <section className={`tab-panel ${s.activeTabId === "result" ? "tab-panel--result" : ""}`} ref={s.tabPanelRef}>
          {s.activeTabId === "source" && (
            <PSourceRender
              buildConfigSettings={s.buildConfigSettings}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              supportedFfmpegReleases={s.initialProgramState.supportedFfmpegReleases}
              LSettingsToolchainUpdate={s.LSettingsToolchainUpdate}
              LSettingsFfmpegUpdate={s.LSettingsFfmpegUpdate}
              LMSYSArchiveUpdate={s.LMSYSArchiveUpdate}
              chooseWorkspaceDirectory={s.chooseWorkspaceDirectory}
              openInUserBrowser={s.openInUserBrowser}
            />
          )}
          {s.activeTabId === "buildConfig" && (
            <PConfigRender
              buildConfigSettings={s.buildConfigSettings}
              LProfileShellUpdate={s.LProfileShellUpdate}
              msys2PackageText={s.msys2PackageText}
              onMsys2PackageTextChange={s.LMSYSPackageUpdate}
            />
          )}
          {s.activeTabId === "prep" && (
            <PPrepRender
              toolchainPreparationPlanReview={s.toolchainPreparationPlanReview}
              toolchainLogEntries={s.toolchainLogEntries}
              approvedActionPhase={s.approvedActionPhase}
              approvedActionStatus={s.approvedActionStatus}
              toolchainProgress={s.toolchainProgress}
              canCancelToolchain={s.canCancelToolchain}
              toolchainStatus={s.toolchainStatus}
              installedToolchainProfiles={s.installedToolchainProfiles}
              toolchainVerification={s.toolchainVerification}
              isVerifyingToolchain={s.isVerifyingToolchain}
              isApprovedActionRunning={s.isApprovedActionRunning}
              isPlanningToolchain={s.isPlanningToolchain}
              currentShellProfileName={s.buildConfigSettings.windowsShellProfileName}
              configuredMsys2PackageNames={s.configuredMsys2PackageNames}
              approveToolchainPreparationPlan={s.approveToolchainPreparationPlan}
              LPlanToolchainCancel={s.LPlanToolchainCancel}
              cancelApprovedAction={s.cancelApprovedAction}
              LActionApprovedClear={s.LActionApprovedClear}
              onVerifyToolchain={s.verifyToolchain}
              onReuseToolchain={() => s.setActiveTabId("library")}
              onReinstallToolchain={s.addBuildConfigPlanAndContinueToPrep}
              onGoToBuildConfig={() => s.setActiveTabId("buildConfig")}
              onClearBuildEnvironments={s.clearBuildEnvironments}
            />
          )}
          {s.activeTabId === "library" && (
            <PLibraryRender
              initialProgramState={s.initialProgramState}
              libraryCatalog={s.libraryCatalog}
              libraryPresetCatalog={s.libraryPresetCatalog}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              libraryPresetId={s.libraryPresetId}
              extendedLibraries={s.extendedLibraries}
              libraryDetailedView={s.libraryDetailedView}
              setLibraryDetailedView={s.setLibraryDetailedView}
              showTechnicalDetails={s.libraryTechnicalDetails}
              setShowTechnicalDetails={s.setLibraryTechnicalDetails}
              sectionFilters={s.librarySectionFilters}
              setSectionFilters={s.setLLibrarySectionFilters}
              LLibraryToggle={s.LLibraryToggle}
              LPresetLibraryApply={s.LPresetLibraryApply}
              LLibraryExtendedUpdate={s.LLibraryExtendedUpdate}
              openInUserBrowser={s.openInUserBrowser}
            />
          )}
          {s.activeTabId === "options" && (
            <POptionRender
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              initialProgramState={s.initialProgramState}
              extraConfigureFlagText={s.extraConfigureFlagText}
              onExtraFlagTextChange={s.LFlagExtraUpdate}
              LSettingsFfmpegUpdate={s.LSettingsFfmpegUpdate}
              LOptionToggle={s.LOptionToggle}
              LPresetOptionApply={s.LPresetOptionApply}
              optionsDetailedView={s.optionsDetailedView}
              setOptionsDetailedView={s.setOptionsDetailedView}
              showTechnicalDetails={s.optionsTechnicalDetails}
              setShowTechnicalDetails={s.setOptionsTechnicalDetails}
            />
          )}
          {s.activeTabId === "buildFfmpeg" && (
            <PBuildRender
              ffmpegBuildPlanReview={s.ffmpegBuildPlanReview}
              ffmpegLogEntries={s.ffmpegLogEntries}
              approvedActionPhase={s.approvedActionPhase}
              approvedActionStatus={s.approvedActionStatus}
              ffmpegStalledAddresses={s.ffmpegStalledAddresses}
              ffmpegProgress={s.ffmpegProgress}
              canCancelFfmpeg={s.canCancelFfmpeg}
              approveFfmpegBuildPlan={s.approveFfmpegBuildPlan}
              retryFfmpegBuildPlan={s.retryFfmpegBuildPlan}
              cancelApprovedAction={s.cancelApprovedAction}
              LActionApprovedClear={s.LActionApprovedClear}
              onGoToOptions={() => s.setActiveTabId("options")}
            />
          )}
          {s.activeTabId === "result" && (
            <PResultRender
              buildResult={s.buildResult}
              buildResultError={s.buildResultError}
              isLoadingBuildResult={s.isLoadingBuildResult}
              buildVerification={s.buildVerification}
              buildVerificationError={s.buildVerificationError}
              isVerifyingBuild={s.isVerifyingBuild}
              verifyBuildResult={s.verifyBuildResult}
              hasWorkspace={Boolean(s.buildConfigSettings.workspaceDirectory || s.ffmpegBuildSettings.workspaceDirectory)}
              refreshBuildResult={s.refreshBuildResult}
              openResultFolder={s.openResultFolder}
              openResultReport={s.openResultReport}
              onGoToSource={() => s.setActiveTabId("source")}
              onGoToBuild={() => s.setActiveTabId("buildFfmpeg")}
            />
          )}
          {s.activeTabId === "logs" && (
            <PLogRender
              toolchainLogEntries={s.toolchainLogEntries}
              ffmpegLogEntries={s.ffmpegLogEntries}
              localLogRecords={s.localLogRecords}
              localLogRecordsError={s.localLogRecordsError}
              refreshLocalLogRecords={s.refreshLocalLogRecords}
              loadLocalLogRecord={s.loadLocalLogRecord}
              openLocalLogsFolder={s.openLocalLogsFolder}
              openLocalLogRecordFolder={s.openLocalLogRecordFolder}
              openLocalLogRecordFile={s.openLocalLogRecordFile}
              onGoToPrep={() => s.setActiveTabId("prep")}
              onGoToBuild={() => s.setActiveTabId("buildFfmpeg")}
            />
          )}
          {s.activeTabId === "about" && (
            <PAboutRender openInUserBrowser={s.openInUserBrowser} />
          )}
        </section>

        {["source", "buildConfig", "prep", "library", "options", "buildFfmpeg"].includes(s.activeTabId) && (
          <div className="tab-bottom-bar">
            <div className="tab-footer-hint">
              <span className="tab-footer-badge" aria-hidden="true"><img src={nextStepIcon} alt="" /></span>
              <span className="tab-footer-hint-text">
                <span className="tab-footer-hint-title">{LLocaleTextGet("footer.nextStep")}</span>
                <span className="tab-footer-hint-sub">{LLocaleTextGet(`footer.hint.${s.activeTabId}`)}</span>
              </span>
            </div>
            <div className="tab-footer-actions">
              {s.activeTabId === "source" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("buildConfig")}>{LLocaleTextGet("actions.chooseBuildConfig")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "buildConfig" && (
                <>
                  <button className="button button--primary" type="button" disabled={s.isPlanningToolchain} onClick={s.addBuildConfigPlanAndContinueToPrep}>{LLocaleTextGet("actions.addBuildPlanAndContinue")}<span className="button__chevron" aria-hidden="true">›</span></button>
                  <button className="button" type="button" onClick={s.LPackageToolchainRestore}>{LLocaleTextGet("actions.restoreRecommendedList")}</button>
                </>
              )}
              {s.activeTabId === "prep" && s.toolchainStatus?.installed && s.approvedActionPhase !== "toolchain" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("library")}>{LLocaleTextGet("actions.chooseFfmpegLibraries")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "library" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("options")}>{LLocaleTextGet("actions.continueToFfmpegOptions")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "options" && (
                <>
                  <button className="button button--primary" type="button" disabled={s.isPlanningFfmpeg} onClick={s.reviewFfmpegPlans}>{LLocaleTextGet("actions.reviewPlan")}<span className="button__chevron" aria-hidden="true">›</span></button>
                  <button className="button" type="button" onClick={s.LFlagExtraRestore}>{LLocaleTextGet("actions.restoreRecommendedOptions")}</button>
                </>
              )}
              {s.activeTabId === "buildFfmpeg" && s.ffmpegBuildPlanReview && s.approvedActionPhase !== "ffmpeg" && (
                <button className="button button--primary" type="button" disabled={!s.ffmpegBuildPlanReview.plan.isExecutable || s.isApprovedActionRunning} onClick={s.approveFfmpegBuildPlan}>{LLocaleTextGet("actions.requestBackendConfirmation")}</button>
              )}
            </div>
          </div>
        )}
      </div>
      {s.approvalConfirmationRequest && (
        <PApprovalConfirmationRender request={s.approvalConfirmationRequest} onResolve={s.resolveApprovalConfirmation} />
      )}
    </main>
  );
}

// ─── Entry point ─────────────────────────────────────────────────────────────

async function LProgramMount() {
  const rootElement = document.getElementById("root");
  if (!rootElement) return;
  const root = createRoot(rootElement);
  const PErrorFatalRender = (title: string, error: unknown) => {
    root.render(<PFatalErrorRender title={title} text={LErrorTextFormat(error)} />);
  };

  window.addEventListener("error", (event) => PErrorFatalRender(LLocaleTextGet("fatal.runtimeFailed"), event.error ?? event.message));
  window.addEventListener("unhandledrejection", (event) => PErrorFatalRender(LLocaleTextGet("fatal.asyncRuntimeFailed"), event.reason));

  try {
    await LLocaleSynchronize();
  } catch (error) {
    PErrorFatalRender(LLocaleTextGet("fatal.runtimeFailed"), error);
    return;
  }

  root.render(
    <React.StrictMode>
      <PErrorBoundary>
        <PProgramRender />
      </PErrorBoundary>
    </React.StrictMode>,
  );
}

void LProgramMount();
