import React from "react";
import { LLocaleTextGet } from "../i18n";
import { PTextDescriptionRender, PHeaderPageRender } from "./shared";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";
import buildToolsIcon from "../assets/build-config-card-icons/BuildTools.svg";
import buildShellIcon from "../assets/build-config-card-icons/BuildShell.svg";

export type LConfigurationProperties = {
  buildConfigSettings: LSettingsToolchain;
  LProfileShellUpdate: (profileName: string) => void;
  msys2PackageText: string;
  onMsys2PackageTextChange: (text: string) => void;
};

function PPanelTechnicalRender(props: { sections: { title: string; text: string }[]; tools?: string[] }) {
  return (
    <div className="card__panel">
      <h2 className="card__panel-title">{LLocaleTextGet("buildConfig.technical.title")}</h2>
      <div className="card__details">
        {props.sections.map((section) => (
          <section className="card__detail" key={section.title}>
            <h3 className="card__detail-title">{section.title}</h3>
            <PTextDescriptionRender text={section.text} className="card__detail-text" groupSentences />
          </section>
        ))}
        {props.tools && props.tools.length > 0 && (
          <section className="card__detail">
            <h3 className="card__detail-title">{LLocaleTextGet("buildConfig.technical.tools.title")}</h3>
            <dl className="tool-glossary">
              {props.tools.map((toolId) => (
                <div className="tool-glossary__row" key={toolId}>
                  <dt className="tool-glossary__name">{toolId}</dt>
                  <dd className="tool-glossary__text">{LLocaleTextGet(`buildConfig.technical.tool.${toolId}`)}</dd>
                </div>
              ))}
            </dl>
          </section>
        )}
      </div>
    </div>
  );
}

// Base package names (profile prefix stripped) of the default toolchain, in the
// same order as the recommended list. Explanations are looked up per id.
const LToolchainGlossaryIds = [
  "base-devel", "git", "make", "diffutils", "binutils", "crt", "gcc", "headers",
  "libmangle", "libwinpthread", "pkgconf", "tools", "winpthreads", "winstorecompat",
  "cmake", "ninja", "nasm", "yasm",
];

export function PConfigRender({ buildConfigSettings, LProfileShellUpdate, msys2PackageText, onMsys2PackageTextChange }: LConfigurationProperties) {
  const [isShellTechnicalOpen, setIsShellTechnicalOpen] = React.useState(false);
  const [isPackagesTechnicalOpen, setIsPackagesTechnicalOpen] = React.useState(false);

  const shellSections = [
    { title: LLocaleTextGet("buildConfig.technical.shell.ucrt64.title"), text: LLocaleTextGet("buildConfig.technical.shell.ucrt64.text") },
    { title: LLocaleTextGet("buildConfig.technical.shell.mingw64.title"), text: LLocaleTextGet("buildConfig.technical.shell.mingw64.text") },
    { title: LLocaleTextGet("buildConfig.technical.shell.clang64.title"), text: LLocaleTextGet("buildConfig.technical.shell.clang64.text") },
  ];
  const packageSections = [
    { title: LLocaleTextGet("buildConfig.technical.packages.title"), text: LLocaleTextGet("buildConfig.technical.packages.text") },
    { title: LLocaleTextGet("buildConfig.technical.scope.title"), text: LLocaleTextGet("buildConfig.technical.scope.text") },
  ];

  return (
    <section className="tab-page build-config-page">
      <PHeaderPageRender title={LLocaleTextGet("buildConfig.title")} text={LLocaleTextGet("buildConfig.intro")} />

      <section className="card card--blue">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={buildShellIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{LLocaleTextGet("buildConfig.shell.label")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("buildConfig.shell.hint")} />
        </div>
        <div className="card__control">
          <select className="card__input" value={buildConfigSettings.windowsShellProfileName} onChange={(event) => LProfileShellUpdate(event.target.value)}>
            <option value="ucrt64">{LLocaleTextGet("buildConfig.shell.ucrt64")}</option>
            <option value="mingw64">{LLocaleTextGet("buildConfig.shell.mingw64")}</option>
            <option value="clang64">{LLocaleTextGet("buildConfig.shell.clang64")}</option>
          </select>
          <button className="button card__toggle" type="button" aria-expanded={isShellTechnicalOpen} onClick={() => setIsShellTechnicalOpen((value) => !value)}>
            <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
            {isShellTechnicalOpen ? LLocaleTextGet("buildConfig.technical.hide") : LLocaleTextGet("buildConfig.technical.show")}
          </button>
        </div>
        {isShellTechnicalOpen && <PPanelTechnicalRender sections={shellSections} />}
      </section>

      <section className="card card--teal">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={buildToolsIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{LLocaleTextGet("buildConfig.packages.label")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("buildConfig.packages.hint")} />
        </div>
        <div className="card__control">
          <textarea className="card__input build-config-packages" rows={12} value={msys2PackageText} onChange={(event) => onMsys2PackageTextChange(event.target.value)} />
        </div>
        <button className="button card__toggle card__toggle--block" type="button" aria-expanded={isPackagesTechnicalOpen} onClick={() => setIsPackagesTechnicalOpen((value) => !value)}>
          <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
          {isPackagesTechnicalOpen ? LLocaleTextGet("buildConfig.technical.hide") : LLocaleTextGet("buildConfig.technical.show")}
        </button>
        {isPackagesTechnicalOpen && <PPanelTechnicalRender sections={packageSections} tools={LToolchainGlossaryIds} />}
      </section>
    </section>
  );
}
