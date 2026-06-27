import React from "react";
import { t } from "../i18n";
import { DescriptionLines, ExternalLinkButton, PageHeader } from "./shared";
import { isDevUnlockEnabled, isSudoDevUnlockEnabled, registerVersionClick, registerDevUnlockIndicatorClick } from "../devUnlock";
import aboutIcon from "../assets/about-card-icons/AboutInfo.svg";
import workflowIcon from "../assets/about-card-icons/WhatThisProgramDoes.svg";
import boundaryIcon from "../assets/about-card-icons/WhatThisProgramDoesNotInclude.svg";

declare const __APP_VERSION__: string;

export type AboutTabProps = {
  openInUserBrowser: (url: string) => Promise<void>;
};

const officialLinks = {
  projectRepository: "https://github.com/magnomer/Prompful-Custom-FFmpeg-builder",
  ffmpegWebsite: "https://ffmpeg.org",
  msys2Website: "https://www.msys2.org",
};

export function AboutTab({ openInUserBrowser }: AboutTabProps) {
  // Hidden developer feature: twelve clicks on the version number toggle basic dev
  // unlock (UI-unavailable libraries become selectable); with basic on, twelve clicks
  // on the dev indicator below toggle the sudo tier (relaxes pick-one groups). See
  // devUnlock.ts.
  const [devUnlocked, setDevUnlocked] = React.useState(isDevUnlockEnabled());
  const [sudoDevUnlocked, setSudoDevUnlocked] = React.useState(isSudoDevUnlockEnabled());

  // Toggling basic dev off also clears sudo; keep the indicator's state in sync.
  function handleVersionClick() {
    const nextDevUnlocked = registerVersionClick();
    setDevUnlocked(nextDevUnlocked);
    if (!nextDevUnlocked) {
      setSudoDevUnlocked(false);
    }
  }

  return (
    <section className="tab-page about-page">
      <PageHeader title={t("about.title")} text={t("about.intro")} />

      <section className="card card--blue about-card about-identity-card" aria-label={t("about.version.ariaLabel")}>
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={aboutIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{t("about.identity.title")}</h2>
          <DescriptionLines text={t("about.identity.text")} />
          <div className="about-version">
            <span className="about-version__label">{t("about.version.label")}</span>
            <strong
              className="about-version__value"
              onClick={handleVersionClick}
            >{t("about.version.value", { version: __APP_VERSION__ })}</strong>
            {devUnlocked && (
              <span
                className={`about-version__dev-unlock ${sudoDevUnlocked ? "about-version__dev-unlock--sudo" : ""}`}
                onClick={() => setSudoDevUnlocked(registerDevUnlockIndicatorClick())}
              >{t(sudoDevUnlocked ? "about.version.sudoDevUnlocked" : "about.version.devUnlocked")}</span>
            )}
          </div>
        </div>
        <div className="about-link-row">
          <div className="about-link-row__left">
            <ExternalLinkButton label={t("about.links.projectRepository")} url={officialLinks.projectRepository} onOpen={openInUserBrowser} />
          </div>
          <div className="about-link-row__right">
            <ExternalLinkButton label={t("about.links.ffmpegWebsite")} url={officialLinks.ffmpegWebsite} onOpen={openInUserBrowser} />
            <ExternalLinkButton label={t("about.links.msys2Website")} url={officialLinks.msys2Website} onOpen={openInUserBrowser} />
          </div>
        </div>
      </section>

      <section className="card card--teal about-card about-flow-card" aria-labelledby="about-flow-title">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={workflowIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title" id="about-flow-title">{t("about.does.title")}</h2>
          <DescriptionLines text={t("about.does.p1")} />
        </div>
        <ol className="about-flow-steps">
          <AboutFlowStep number="1" label={t("about.flow.workspace.label")} caption={t("about.flow.workspace.caption")} />
          <AboutFlowStep number="2" label={t("about.flow.configure.label")} caption={t("about.flow.configure.caption")} />
          <AboutFlowStep number="3" label={t("about.flow.review.label")} caption={t("about.flow.review.caption")} />
          <AboutFlowStep number="4" label={t("about.flow.result.label")} caption={t("about.flow.result.caption")} />
        </ol>
      </section>

      <section className="card card--purple about-card about-boundary-card" aria-labelledby="about-boundary-title">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={boundaryIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title" id="about-boundary-title">{t("about.doesNotInclude.title")}</h2>
          <DescriptionLines text={t("about.doesNotInclude.p2")} />
        </div>
        <div className="about-boundary-items">
          <span className="about-boundary-item">{t("about.notIncluded.prebuilt")}</span>
          <span className="about-boundary-item">{t("about.notIncluded.source")}</span>
          <span className="about-boundary-item">{t("about.notIncluded.codecs")}</span>
          <span className="about-boundary-item">{t("about.notIncluded.packages")}</span>
          <span className="about-boundary-item">{t("about.notIncluded.hidden")}</span>
        </div>
      </section>

    </section>
  );
}

function AboutFlowStep(props: { number: string; label: string; caption: string }) {
  return (
    <li className="about-flow-step">
      <span className="about-flow-step__number">{props.number}</span>
      <span className="about-flow-step__label">{props.label}</span>
      <span className="about-flow-step__caption">{props.caption}</span>
    </li>
  );
}
