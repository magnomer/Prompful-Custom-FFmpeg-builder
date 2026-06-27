import React from "react";
import { t } from "../i18n";
import { PageHeader } from "./shared";

export type BuildToolsTabProps = {
  buildToolSettings: BuildToolSettings;
  updateBuildToolSettings: (partial: Partial<BuildToolSettings>) => void;
  updateFfmpegBuildSettings: (partial: Partial<FfmpegBuildSettings>) => void;
  msys2PackageText: string;
  onMsys2PackageTextChange: (text: string) => void;
};

export function BuildToolsTab({ buildToolSettings, updateBuildToolSettings, updateFfmpegBuildSettings, msys2PackageText, onMsys2PackageTextChange }: BuildToolsTabProps) {
  const [isTechnicalOpen, setIsTechnicalOpen] = React.useState(false);

  const technicalSections = [
    { title: t("buildTools.technical.shell.title"), text: t("buildTools.technical.shell.text") },
    { title: t("buildTools.technical.packages.title"), text: t("buildTools.technical.packages.text") },
    { title: t("buildTools.technical.scope.title"), text: t("buildTools.technical.scope.text") },
  ];

  return (
    <section className="tab-page build-tools-page">
      <PageHeader title={t("buildTools.title")} text={t("buildTools.intro")} />

      <section className="build-tools-section">
        <label className="field build-tools-field">
          <span className="field__label">{t("buildTools.shell.label")}</span>
          <span className="field__hint">{t("buildTools.shell.hint")}</span>
          <select className="field__input" value={buildToolSettings.windowsShellProfileName} onChange={(event) => { updateBuildToolSettings({ windowsShellProfileName: event.target.value }); updateFfmpegBuildSettings({ windowsShellProfileName: event.target.value }); }}>
            <option value="ucrt64">{t("buildTools.shell.ucrt64")}</option>
            <option value="mingw64">{t("buildTools.shell.mingw64")}</option>
            <option value="clang64">{t("buildTools.shell.clang64")}</option>
          </select>
        </label>
      </section>

      <section className="build-tools-section">
        <label className="field build-tools-field">
          <span className="field__label">{t("buildTools.packages.label")}</span>
          <span className="field__hint">{t("buildTools.packages.hint")}</span>
          <textarea className="field__textarea build-tools-packages" rows={12} value={msys2PackageText} onChange={(event) => onMsys2PackageTextChange(event.target.value)} />
        </label>
      </section>

      <div className="build-tools-technical">
        <button className="button build-tools-technical-toggle" type="button" aria-expanded={isTechnicalOpen} onClick={() => setIsTechnicalOpen((value) => !value)}>
          {isTechnicalOpen ? t("buildTools.technical.hide") : t("buildTools.technical.show")}
        </button>
        {isTechnicalOpen && (
          <div className="build-tools-technical-panel">
            <h2 className="build-tools-technical-panel__title">{t("buildTools.technical.title")}</h2>
            <div className="build-tools-technical-details">
              {technicalSections.map((section) => (
                <section className="build-tools-technical-detail" key={section.title}>
                  <h3 className="build-tools-technical-detail__title">{section.title}</h3>
                  <p className="build-tools-technical-detail__text">{section.text}</p>
                </section>
              ))}
            </div>
          </div>
        )}
      </div>
    </section>
  );
}
