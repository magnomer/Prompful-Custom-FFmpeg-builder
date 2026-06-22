import React from "react";
import { createRoot } from "react-dom/client";
import "./style.css";

import { SourceTab } from "./tabs/source";
import { BuildToolsTab } from "./tabs/buildtools";
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
  const tabItems: { id: TabId; label: string; description: string }[] = [
    { id: "source",      label: t("nav.source.label"),      description: t("nav.source.description") },
    { id: "buildTools",  label: t("nav.buildTools.label"),  description: t("nav.buildTools.description") },
    { id: "prep",        label: t("nav.prep.label"),        description: t("nav.prep.description") },
    { id: "library",     label: t("nav.library.label"),     description: t("nav.library.description", { count: selectedLibraryCount }) },
    { id: "options",     label: t("nav.options.label"),     description: t("nav.options.description") },
    { id: "buildFfmpeg", label: t("nav.buildFfmpeg.label"), description: tStatus(s.approvedActionStatus) },
    { id: "result",      label: t("nav.result.label"),      description: t("nav.result.description") },
    { id: "logs",        label: t("nav.logs.label"),        description: t("nav.logs.description", { count: s.securityLogEntries.length }) },
  ];

  return (
    <main className="app-shell">
      <aside className="left-nav" aria-label={t("nav.ariaLabel")}>
        <div className="left-nav__brand">{t("app.brand")}</div>
        <nav className="left-nav__items">
          {tabItems.map((tabItem) => (
            <button className={`left-nav__item ${s.activeTabId === tabItem.id ? "left-nav__item--active" : ""}`} key={tabItem.id} type="button" onClick={() => s.setActiveTabId(tabItem.id)}>
              <span className="left-nav__label">{tabItem.label}</span>
              <span className="left-nav__description">{tabItem.description}</span>
            </button>
          ))}
        </nav>
        <div className="left-nav__bottom">
          <button className={`left-nav__item left-nav__item--about ${s.activeTabId === "about" ? "left-nav__item--active" : ""}`} type="button" onClick={() => s.setActiveTabId("about")}>
            <span className="left-nav__label">{t("nav.about.label")}</span>
            <span className="left-nav__description">{t("nav.about.description")}</span>
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
              <span>{selectedLocaleLabel}</span>
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
        <section className="tab-panel" ref={s.tabPanelRef}>
          {s.activeTabId === "source" && (
            <SourceTab
              buildToolSettings={s.buildToolSettings}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              updateBuildToolSettings={s.updateBuildToolSettings}
              updateFfmpegBuildSettings={s.updateFfmpegBuildSettings}
              updateMsys2ArchiveUrl={s.updateMsys2ArchiveUrl}
              chooseWorkspaceDirectory={s.chooseWorkspaceDirectory}
              openInUserBrowser={s.openInUserBrowser}
            />
          )}
          {s.activeTabId === "buildTools" && (
            <BuildToolsTab
              buildToolSettings={s.buildToolSettings}
              updateBuildToolSettings={s.updateBuildToolSettings}
              updateFfmpegBuildSettings={s.updateFfmpegBuildSettings}
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
              approveToolchainPreparationPlan={s.approveToolchainPreparationPlan}
              cancelApprovedAction={s.cancelApprovedAction}
            />
          )}
          {s.activeTabId === "library" && (
            <LibrariesTab
              initialApplicationState={s.initialApplicationState}
              ffmpegBuildSettings={s.ffmpegBuildSettings}
              libraryPresetId={s.libraryPresetId}
              toggleLibrary={s.toggleLibrary}
              applyLibraryPreset={s.applyLibraryPreset}
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
            />
          )}
          {s.activeTabId === "result" && (
            <ResultTab
              buildResult={s.buildResult}
              buildResultError={s.buildResultError}
              refreshBuildResult={s.refreshBuildResult}
              openResultFolder={s.openResultFolder}
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

        <div className="tab-bottom-bar">
          {s.activeTabId === "source" && (
            <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("buildTools")}>{t("actions.chooseBuildTools")}</button>
          )}
          {s.activeTabId === "buildTools" && (
            <>
              <button className="button button--primary" type="button" onClick={s.addBuildToolsPlanAndContinueToPrep}>{t("actions.addBuildPlanAndContinue")}</button>
              <button className="button" type="button" onClick={s.restoreRecommendedToolchainPackages}>{t("actions.restoreRecommendedList")}</button>
            </>
          )}
          {s.activeTabId === "prep" && s.toolchainPreparationPlanReview && (
            <button className="button button--primary" type="button" disabled={!s.toolchainPreparationPlanReview.plan.isExecutable} onClick={s.approveToolchainPreparationPlan}>{t("actions.requestBackendConfirmation")}</button>
          )}
          {s.activeTabId === "prep" && (
            <button className="button" type="button" onClick={() => s.setActiveTabId("library")}>{t("actions.chooseFfmpegLibraries")}</button>
          )}
          {s.activeTabId === "library" && (
            <button className="button button--primary" type="button" onClick={() => s.setActiveTabId("options")}>{t("actions.continueToFfmpegOptions")}</button>
          )}
          {s.activeTabId === "options" && (
            <>
              <button className="button button--primary" type="button" onClick={s.reviewFfmpegPlans}>{t("actions.reviewPlan")}</button>
              <button className="button" type="button" onClick={s.restoreRecommendedExtraFlags}>{t("actions.restoreRecommendedOptions")}</button>
            </>
          )}
          {s.activeTabId === "buildFfmpeg" && s.ffmpegBuildPlanReview && (
            <button className="button button--primary" type="button" disabled={!s.ffmpegBuildPlanReview.plan.isExecutable} onClick={s.approveFfmpegBuildPlan}>{t("actions.requestBackendConfirmation")}</button>
          )}
        </div>
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
