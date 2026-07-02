import React from "react";
import { LLocaleTextGet } from "../i18n";
import { PTextDescriptionRender, PButtonLinkExternalRender, PHeaderPageRender } from "./shared";
import { LUnlockBasicCheck, LUnlockSudoCheck, LUnlockVersionClickRegister, LUnlockIndicatorClickRegister } from "../devUnlock";
import aboutIcon from "../assets/about-card-icons/AboutInfo.svg";
import workflowIcon from "../assets/about-card-icons/WhatThisProgramDoes.svg";
import boundaryIcon from "../assets/about-card-icons/WhatThisProgramDoesNotInclude.svg";

declare const __PROGRAM_VERSION__: string;

export type PAboutProps = {
  openInUserBrowser: (url: string) => Promise<void>;
};

const LLinkOfficialMap = {
  projectRepository: "https://github.com/magnomer/Prompful-Custom-FFmpeg-builder",
  ffmpegWebsite: "https://ffmpeg.org",
  msys2Website: "https://www.msys2.org",
};

export function PAboutRender({ openInUserBrowser }: PAboutProps) {
  // Hidden developer feature: twelve clicks on the version number toggle basic dev
  // unlock (UI-unavailable libraries become selectable); with basic on, twelve clicks
  // on the dev indicator below toggle the sudo tier (relaxes pick-one groups). See
  // devUnlock.ts.
  const [devUnlocked, setDevUnlocked] = React.useState(LUnlockBasicCheck());
  const [sudoDevUnlocked, setSudoDevUnlocked] = React.useState(LUnlockSudoCheck());

  // Toggling basic dev off also clears sudo; keep the indicator's state in sync.
  function handleVersionClick() {
    const nextDevUnlocked = LUnlockVersionClickRegister();
    setDevUnlocked(nextDevUnlocked);
    if (!nextDevUnlocked) {
      setSudoDevUnlocked(false);
    }
  }

  return (
    <section className="tab-page about-page">
      <PHeaderPageRender title={LLocaleTextGet("about.title")} text={LLocaleTextGet("about.intro")} />

      <section className="card card--blue about-card about-identity-card" aria-label={LLocaleTextGet("about.version.ariaLabel")}>
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={aboutIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{LLocaleTextGet("about.identity.title")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("about.identity.text")} />
          <div className="about-version">
            <span className="about-version__label">{LLocaleTextGet("about.version.label")}</span>
            <strong
              className="about-version__value"
              onClick={handleVersionClick}
            >{LLocaleTextGet("about.version.value", { version: __PROGRAM_VERSION__ })}</strong>
            {devUnlocked && (
              <span
                className={`about-version__dev-unlock ${sudoDevUnlocked ? "about-version__dev-unlock--sudo" : ""}`}
                onClick={() => setSudoDevUnlocked(LUnlockIndicatorClickRegister())}
              >{LLocaleTextGet(sudoDevUnlocked ? "about.version.sudoDevUnlocked" : "about.version.devUnlocked")}</span>
            )}
          </div>
        </div>
        <div className="about-link-row">
          <div className="about-link-row__left">
            <PButtonLinkExternalRender label={LLocaleTextGet("about.links.projectRepository")} url={LLinkOfficialMap.projectRepository} onOpen={openInUserBrowser} />
          </div>
          <div className="about-link-row__right">
            <PButtonLinkExternalRender label={LLocaleTextGet("about.links.ffmpegWebsite")} url={LLinkOfficialMap.ffmpegWebsite} onOpen={openInUserBrowser} />
            <PButtonLinkExternalRender label={LLocaleTextGet("about.links.msys2Website")} url={LLinkOfficialMap.msys2Website} onOpen={openInUserBrowser} />
          </div>
        </div>
      </section>

      <section className="card card--teal about-card about-flow-card" aria-labelledby="about-flow-title">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={workflowIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title" id="about-flow-title">{LLocaleTextGet("about.does.title")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("about.does.p1")} />
        </div>
        <ol className="about-flow-steps">
          <PStepAboutRender number="1" label={LLocaleTextGet("about.flow.workspace.label")} caption={LLocaleTextGet("about.flow.workspace.caption")} />
          <PStepAboutRender number="2" label={LLocaleTextGet("about.flow.configure.label")} caption={LLocaleTextGet("about.flow.configure.caption")} />
          <PStepAboutRender number="3" label={LLocaleTextGet("about.flow.review.label")} caption={LLocaleTextGet("about.flow.review.caption")} />
          <PStepAboutRender number="4" label={LLocaleTextGet("about.flow.result.label")} caption={LLocaleTextGet("about.flow.result.caption")} />
        </ol>
      </section>

      <section className="card card--purple about-card about-boundary-card" aria-labelledby="about-boundary-title">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={boundaryIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title" id="about-boundary-title">{LLocaleTextGet("about.doesNotInclude.title")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("about.doesNotInclude.p2")} />
        </div>
        <div className="about-boundary-items">
          <span className="about-boundary-item">{LLocaleTextGet("about.notIncluded.prebuilt")}</span>
          <span className="about-boundary-item">{LLocaleTextGet("about.notIncluded.source")}</span>
          <span className="about-boundary-item">{LLocaleTextGet("about.notIncluded.codecs")}</span>
          <span className="about-boundary-item">{LLocaleTextGet("about.notIncluded.packages")}</span>
          <span className="about-boundary-item">{LLocaleTextGet("about.notIncluded.hidden")}</span>
        </div>
      </section>

    </section>
  );
}

function PStepAboutRender(props: { number: string; label: string; caption: string }) {
  return (
    <li className="about-flow-step">
      <span className="about-flow-step__number">{props.number}</span>
      <span className="about-flow-step__label">{props.label}</span>
      <span className="about-flow-step__caption">{props.caption}</span>
    </li>
  );
}
