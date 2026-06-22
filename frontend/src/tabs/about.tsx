import React from "react";
import { t } from "../i18n";
import { PageHeader, ExternalLinkButton } from "./shared";

declare const __APP_VERSION__: string;

export type AboutTabProps = {
  openInUserBrowser: (url: string) => Promise<void>;
};

export function AboutTab({ openInUserBrowser }: AboutTabProps) {
  const [showTechnicalDetails, setShowTechnicalDetails] = React.useState(false);

  return (
    <section className="tab-page about-page">
      <PageHeader title={t("about.title")} text={t("about.intro")} />

      <section className="about-summary">
        <section className="about-version-card" aria-label={t("about.version.ariaLabel")}>
          <span className="about-version-card__label">{t("about.version.label")}</span>
          <strong className="about-version-card__value">{t("about.version.value", { version: __APP_VERSION__ })}</strong>
        </section>

        <section className="about-flow" aria-labelledby="about-flow-title">
          <div className="about-flow__header">
            <h2 className="about-flow__title" id="about-flow-title">{t("about.does.title")}</h2>
            <p className="about-flow__text">{t("about.does.p1")}</p>
          </div>
          <ol className="about-flow__steps">
            <li className="about-flow__step">
              <span className="about-flow__number">1</span>
              <span className="about-flow__label">{t("about.flow.workspace.label")}</span>
              <span className="about-flow__caption">{t("about.flow.workspace.caption")}</span>
            </li>
            <li className="about-flow__step">
              <span className="about-flow__number">2</span>
              <span className="about-flow__label">{t("about.flow.configure.label")}</span>
              <span className="about-flow__caption">{t("about.flow.configure.caption")}</span>
            </li>
            <li className="about-flow__step">
              <span className="about-flow__number">3</span>
              <span className="about-flow__label">{t("about.flow.review.label")}</span>
              <span className="about-flow__caption">{t("about.flow.review.caption")}</span>
            </li>
            <li className="about-flow__step">
              <span className="about-flow__number">4</span>
              <span className="about-flow__label">{t("about.flow.result.label")}</span>
              <span className="about-flow__caption">{t("about.flow.result.caption")}</span>
            </li>
          </ol>
        </section>

        <section className="about-boundary" aria-labelledby="about-boundary-title">
          <div className="about-boundary__header">
            <h2 className="about-boundary__title" id="about-boundary-title">{t("about.doesNotInclude.title")}</h2>
            <p className="about-boundary__text">{t("about.doesNotInclude.p2")}</p>
          </div>
          <div className="about-boundary__items">
            <span className="about-boundary__item">{t("about.notIncluded.prebuilt")}</span>
            <span className="about-boundary__item">{t("about.notIncluded.source")}</span>
            <span className="about-boundary__item">{t("about.notIncluded.codecs")}</span>
            <span className="about-boundary__item">{t("about.notIncluded.packages")}</span>
            <span className="about-boundary__item">{t("about.notIncluded.hidden")}</span>
          </div>
        </section>
      </section>

      <div className="about-technical">
        <button className="button about-technical-toggle" type="button" aria-expanded={showTechnicalDetails} onClick={() => setShowTechnicalDetails((value) => !value)}>
          {showTechnicalDetails ? t("about.technical.hide") : t("about.technical.show")}
        </button>
        {showTechnicalDetails && (
          <section className="about-technical-panel">
            <h2 className="about-technical-panel__title">{t("about.technical.title")}</h2>
            <div className="about-technical-details">
              <AboutTechnicalDetail title={t("about.approval.title")}>
                <p>{t("about.approval.p1")}</p>
                <p>{t("about.approval.p2")}</p>
                <p>{t("about.approval.p3.prefix")} <code>logs/</code> {t("about.approval.p3.suffix")}</p>
              </AboutTechnicalDetail>

              <AboutTechnicalDetail title={t("about.downloads.title")}>
                <p>{t("about.downloads.p1")}</p>
                <p>{t("about.downloads.p2.prefix")} <code>.sig</code>{t("about.downloads.p2.middle")} <code>.asc</code>{t("about.downloads.p2.suffix")}</p>
                <p>{t("about.downloads.p3")}</p>
              </AboutTechnicalDetail>

              <AboutTechnicalDetail title={t("about.license.title")}>
                <p>{t("about.license.p1")}</p>
                <p>{t("about.license.p2")}</p>
                <p>{t("about.license.p3")}</p>
              </AboutTechnicalDetail>
            </div>
          </section>
        )}
      </div>

      <div className="about-links">
        <ExternalLinkButton label={t("about.links.ffmpegWebsite")} url="https://ffmpeg.org" onOpen={openInUserBrowser} />
        <ExternalLinkButton label={t("about.links.ffmpegLicense")} url="https://ffmpeg.org/legal.html" onOpen={openInUserBrowser} />
        <ExternalLinkButton label={t("about.links.msys2Website")} url="https://www.msys2.org" onOpen={openInUserBrowser} />
      </div>
    </section>
  );
}

function AboutTechnicalDetail(props: { title: string; children: React.ReactNode }) {
  return (
    <section className="about-technical-detail">
      <h3 className="about-technical-detail__title">{props.title}</h3>
      <div className="about-technical-detail__text">{props.children}</div>
    </section>
  );
}
