import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";

import { SourceTab } from "./tabs/source";
import { BuildConfigTab } from "./tabs/buildconfig";
import { PrepTab } from "./tabs/prep";
import { LibrariesTab } from "./tabs/libraries";
import { OptionsTab } from "./tabs/options";
import { FFmpegBuildTab } from "./tabs/ffmpegbuild";
import { ResultTab } from "./tabs/result";
import { LogsTab } from "./tabs/logs";
import { AboutTab } from "./tabs/about";
import { useBuilderState } from "./useBuilderState";
import type { TabId } from "./appstate";
import { getLocale, setLocale, t, tStatus, type LocaleCode } from "./i18n";
import sourceIcon from "./assets/tab-icons/Source.svg";
import buildConfigurationIcon from "./assets/tab-icons/BuildConfiguration.svg";
import prepIcon from "./assets/tab-icons/Prep.svg";
import ffmpegLibrariesIcon from "./assets/tab-icons/FFmpegLibraries.svg";
import ffmpegOptionsIcon from "./assets/tab-icons/FFmpegOptions.svg";
import buildFfmpegIcon from "./assets/tab-icons/BuildFFmpeg.svg";
import resultIcon from "./assets/tab-icons/Result.svg";
import logsIcon from "./assets/tab-icons/Logs.svg";
import aboutIcon from "./assets/tab-icons/About.svg";
import appIcon from "../../build/appicon.png";
import nextStepIcon from "./assets/footer-icons/NextStep.svg";

// ─── BuilderApp ───────────────────────────────────────────────────────────────

function BuilderApp() {
  const [locale, setLocaleState] = React.useState<LocaleCode>(() => getLocale());
  React.useEffect(() => {
    const updateLocale = () => setLocaleState(getLocale());
    globalThis.addEventListener("customffmpeg-locale-change", updateLocale);
    return () => globalThis.removeEventListener("customffmpeg-locale-change", updateLocale);
  }, []);
  React.useEffect(() => { document.title = t("app.brand"); }, [locale]);
  const s = useBuilderState();
  const [isLocaleMenuOpen, setIsLocaleMenuOpen] = React.useState(false);
  const localeItems: { id: LocaleCode; label: string }[] = [
    { id: "en", label: t("locale.english") },
    { id: "ko", label: t("locale.korean") },
  ];
  const selectedLocaleLabel = localeItems.find((localeItem) => localeItem.id === locale)?.label ?? t("locale.english");

  const selectedLibraryCount = s.ffmpegBuildSettings.selectedLibraryIds.length;
  const tabItems: { id: TabId; label: string; description: string; icon: string }[] = [
    { id: "source",      label: t("nav.source.label"),      description: t("nav.source.description"), icon: sourceIcon },
    { id: "buildConfig",  label: t("nav.buildConfig.label"),  description: t("nav.buildConfig.description"), icon: buildConfigurationIcon },
    { id: "prep",        label: t("nav.prep.label"),        description: t("nav.prep.description"), icon: prepIcon },
    { id: "library",     label: t("nav.library.label"),     description: t("nav.library.description", { count: selectedLibraryCount }), icon: ffmpegLibrariesIcon },
    { id: "options",     label: t("nav.options.label"),     description: t("nav.options.description"), icon: ffmpegOptionsIcon },
    { id: "buildFfmpeg", label: t("nav.buildFfmpeg.label"), description: tStatus(s.approvedActionStatus), icon: buildFfmpegIcon },
    { id: "result",      label: t("nav.result.label"),      description: t("nav.result.description"), icon: resultIcon },
    { id: "logs",        label: t("nav.logs.label"),        description: t("nav.logs.description", { count: s.securityLogEntries.length }), icon: logsIcon },
  ];

  return (
    <main className="app-shell">
      <aside className="left-nav" aria-label={t("nav.ariaLabel")}>
        <div className="left-nav__brand">
          <span className="left-nav__logo" aria-hidden="true"><img src={appIcon} alt="" /></span>
          <span className="left-nav__brand-title">{t("app.brand")}</span>
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
              <span className="left-nav__label">{t("nav.about.label")}</span>
              <span className="left-nav__description">{t("nav.about.description")}</span>
            </span>
          </button>
          <div className="left-nav__locale-menu">
            <button
              className={`left-nav__locale-button ${isLocaleMenuOpen ? "left-nav__locale-button--open" : ""}`}
              type="button"
              aria-label={t("locale.selector.ariaLabel")}
              aria-haspopup="listbox"
              aria-expanded={isLocaleMenuOpen}
              onClick={() => setIsLocaleMenuOpen((isOpen) => !isOpen)}
            >
              <span className="left-nav__locale-globe" aria-hidden="true" />
              <span className="left-nav__locale-label">{selectedLocaleLabel}</span>
              <span className="left-nav__locale-chevron" aria-hidden="true" />
            </button>
            {isLocaleMenuOpen && (
              <div className="left-nav__locale-popover" role="listbox" aria-label={t("locale.selector.ariaLabel")}>
                {localeItems.map((localeItem) => (
                  <button
                    className={`left-nav__locale-option ${locale === localeItem.id ? "left-nav__locale-option--selected" : ""}`}
                    key={localeItem.id}
                    type="button"
                    role="option"
                    aria-selected={locale === localeItem.id}
                    onClick={() => {
                      setLocale(localeItem.id);
                      setIsLocaleMenuOpen(false);
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
            <SourceTab
              buildConfigSettings={s.buildConfigSettings}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              updateBuildConfigSettings={s.updateBuildConfigSettings}
              updateFfmpegBuildSettings={s.updateFfmpegBuildSettings}
              updateMsys2ArchiveUrl={s.updateMsys2ArchiveUrl}
              chooseWorkspaceDirectory={s.chooseWorkspaceDirectory}
              openInUserBrowser={s.openInUserBrowser}
            />
          )}
          {s.activeTabId === "buildConfig" && (
            <BuildConfigTab
              buildConfigSettings={s.buildConfigSettings}
              changeShellProfile={s.changeShellProfile}
              msys2PackageText={s.msys2PackageText}
              onMsys2PackageTextChange={s.handleMsys2PackageTextChange}
            />
          )}
          {s.activeTabId === "prep" && (
            <PrepTab
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
              currentShellProfileName={s.buildConfigSettings.windowsShellProfileName}
              configuredMsys2PackageNames={s.configuredMsys2PackageNames}
              approveToolchainPreparationPlan={s.approveToolchainPreparationPlan}
              cancelApprovedAction={s.cancelApprovedAction}
              onVerifyToolchain={s.verifyToolchain}
              onReuseToolchain={() => s.setActiveTabId("library")}
              onReinstallToolchain={s.addBuildConfigPlanAndContinueToPrep}
              onGoToBuildConfig={() => s.setActiveTabId("buildConfig")}
              onClearBuildEnvironments={s.clearBuildEnvironments}
            />
          )}
          {s.activeTabId === "library" && (
            <LibrariesTab
              initialApplicationState={s.initialApplicationState}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              libraryPresetId={s.libraryPresetId}
              extendedLibraries={s.extendedLibraries}
              libraryDetailedView={s.libraryDetailedView}
              setLibraryDetailedView={s.setLibraryDetailedView}
              showTechnicalDetails={s.libraryTechnicalDetails}
              setShowTechnicalDetails={s.setLibraryTechnicalDetails}
              sectionFilters={s.librarySectionFilters}
              setSectionFilters={s.setLibrarySectionFilters}
              toggleLibrary={s.toggleLibrary}
              applyLibraryPreset={s.applyLibraryPreset}
              setExtendedLibraries={s.setExtendedLibraries}
              openInUserBrowser={s.openInUserBrowser}
            />
          )}
          {s.activeTabId === "options" && (
            <OptionsTab
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              initialApplicationState={s.initialApplicationState}
              extraConfigureFlagText={s.extraConfigureFlagText}
              onExtraFlagTextChange={s.handleExtraFlagTextChange}
              updateFfmpegBuildSettings={s.updateFfmpegBuildSettings}
              toggleConfigureOption={s.toggleConfigureOption}
              applyOptionPreset={s.applyOptionPreset}
              optionsDetailedView={s.optionsDetailedView}
              setOptionsDetailedView={s.setOptionsDetailedView}
              showTechnicalDetails={s.optionsTechnicalDetails}
              setShowTechnicalDetails={s.setOptionsTechnicalDetails}
            />
          )}
          {s.activeTabId === "buildFfmpeg" && (
            <FFmpegBuildTab
              ffmpegBuildPlanReview={s.ffmpegBuildPlanReview}
              ffmpegLogEntries={s.ffmpegLogEntries}
              approvedActionPhase={s.approvedActionPhase}
              approvedActionStatus={s.approvedActionStatus}
              ffmpegProgress={s.ffmpegProgress}
              canCancelFfmpeg={s.canCancelFfmpeg}
              approveFfmpegBuildPlan={s.approveFfmpegBuildPlan}
              cancelApprovedAction={s.cancelApprovedAction}
              onGoToOptions={() => s.setActiveTabId("options")}
            />
          )}
          {s.activeTabId === "result" && (
            <ResultTab
              buildResult={s.buildResult}
              buildResultError={s.buildResultError}
              refreshBuildResult={s.refreshBuildResult}
              openResultFolder={s.openResultFolder}
              openResultReport={s.openResultReport}
            />
          )}
          {s.activeTabId === "logs" && (
            <LogsTab
              toolchainLogEntries={s.toolchainLogEntries}
              ffmpegLogEntries={s.ffmpegLogEntries}
            />
          )}
          {s.activeTabId === "about" && (
            <AboutTab openInUserBrowser={s.openInUserBrowser} />
          )}
        </section>

        {["source", "buildConfig", "prep", "library", "options", "buildFfmpeg"].includes(s.activeTabId) && (
          <div className="tab-bottom-bar">
            <div className="tab-footer-hint">
              <span className="tab-footer-badge" aria-hidden="true"><img src={nextStepIcon} alt="" /></span>
              <span className="tab-footer-hint-text">
                <span className="tab-footer-hint-title">{t("footer.nextStep")}</span>
                <span className="tab-footer-hint-sub">{t(`footer.hint.${s.activeTabId}`)}</span>
              </span>
            </div>
            <div className="tab-footer-actions">
              {s.activeTabId === "source" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("buildConfig")}>{t("actions.chooseBuildConfig")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "buildConfig" && (
                <>
                  <button className="button button--primary" type="button" onClick={s.addBuildConfigPlanAndContinueToPrep}>{t("actions.addBuildPlanAndContinue")}<span className="button__chevron" aria-hidden="true">›</span></button>
                  <button className="button" type="button" onClick={s.restoreRecommendedToolchainPackages}>{t("actions.restoreRecommendedList")}</button>
                </>
              )}
              {s.activeTabId === "prep" && s.toolchainPreparationPlanReview && s.approvedActionPhase !== "toolchain" && (
                <button className="button button--primary" type="button" disabled={!s.toolchainPreparationPlanReview.plan.isExecutable || s.isApprovedActionRunning} onClick={s.approveToolchainPreparationPlan}>{t("actions.requestBackendConfirmation")}</button>
              )}
              {s.activeTabId === "prep" && s.installedToolchainProfiles.length > 0 && s.approvedActionPhase !== "toolchain" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("library")}>{t("actions.chooseFfmpegLibraries")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "library" && (
                <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("options")}>{t("actions.continueToFfmpegOptions")}<span className="button__chevron" aria-hidden="true">›</span></button>
              )}
              {s.activeTabId === "options" && (
                <>
                  <button className="button button--primary" type="button" onClick={s.reviewFfmpegPlans}>{t("actions.reviewPlan")}<span className="button__chevron" aria-hidden="true">›</span></button>
                  <button className="button" type="button" onClick={s.restoreRecommendedExtraFlags}>{t("actions.restoreRecommendedOptions")}</button>
                </>
              )}
              {s.activeTabId === "buildFfmpeg" && s.ffmpegBuildPlanReview && s.approvedActionPhase !== "ffmpeg" && (
                <button className="button button--primary" type="button" disabled={!s.ffmpegBuildPlanReview.plan.isExecutable || s.isApprovedActionRunning} onClick={s.approveFfmpegBuildPlan}>{t("actions.requestBackendConfirmation")}</button>
              )}
            </div>
          </div>
        )}
      </div>
    </main>
  );
}

// ─── Entry point ─────────────────────────────────────────────────────────────

createRoot(document.getElementById("root") as HTMLElement).render(
  <React.StrictMode>
    <BuilderApp />
  </React.StrictMode>,
);
