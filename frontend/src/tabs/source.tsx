import React from "react";
import { t } from "../i18n";
import { ExternalLinkButton, PageHeader } from "./shared";

const officialLinks = {
  msys2ArchiveList: "https://repo.msys2.org/distrib/x86_64/",
  msys2InstallerDocs: "https://www.msys2.org/docs/installer/",
  ffmpegDownload: "https://www.ffmpeg.org/download.html",
  ffmpegReleaseIndex: "https://ffmpeg.org/releases/",
  ffmpegSigningKey: "https://ffmpeg.org/ffmpeg-devel.asc",
};

export type SourceTabProps = {
  buildToolSettings: BuildToolSettings;
  ffmpegBuildSettings: FfmpegBuildSettings;
  updateBuildToolSettings: (partial: Partial<BuildToolSettings>) => void;
  updateFfmpegBuildSettings: (partial: Partial<FfmpegBuildSettings>) => void;
  updateMsys2ArchiveUrl: (url: string) => void;
  chooseWorkspaceDirectory: () => Promise<void>;
  openInUserBrowser: (url: string) => Promise<void>;
};

export function SourceTab({ buildToolSettings, ffmpegBuildSettings, updateBuildToolSettings, updateFfmpegBuildSettings, updateMsys2ArchiveUrl, chooseWorkspaceDirectory, openInUserBrowser }: SourceTabProps) {
  const [isMsys2TechnicalOpen, setIsMsys2TechnicalOpen] = React.useState(false);
  const [isFfmpegTechnicalOpen, setIsFfmpegTechnicalOpen] = React.useState(false);

  return (
    <section className="tab-page source-page">
      <PageHeader title={t("source.title")} text={t("source.briefing")} />

      <section className="source-section source-section--workspace">
        <h2 className="source-section__title">{t("source.workspace.label")}</h2>
        <p className="source-explanation">{t("source.workspace.hint")}</p>
        <label className="source-url-row">
          <span className="source-field__label source-field__label--hidden">{t("source.workspace.label")}</span>
          <input className="source-field__input" value={buildToolSettings.workspaceDirectory} onChange={(event) => { updateBuildToolSettings({ workspaceDirectory: event.target.value }); updateFfmpegBuildSettings({ workspaceDirectory: event.target.value }); }} placeholder={t("source.workspace.placeholder")} />
          <button className="button source-field__button" type="button" onClick={chooseWorkspaceDirectory}>{t("actions.browse")}</button>
        </label>
      </section>

      <SourceArchiveSection
        title={t("source.msys2.title")}
        explanation={t("source.msys2.text")}
        archiveLabel={t("source.msys2Archive.label")}
        archiveValue={buildToolSettings.msys2ArchiveUrl}
        archivePlaceholder={t("source.msys2Archive.placeholder")}
        isTechnicalOpen={isMsys2TechnicalOpen}
        onToggleTechnical={() => setIsMsys2TechnicalOpen((value) => !value)}
        onArchiveChange={updateMsys2ArchiveUrl}
        links={[
          { label: t("source.links.msys2ArchiveList"), url: officialLinks.msys2ArchiveList },
          { label: t("source.links.msys2Notes"), url: officialLinks.msys2InstallerDocs },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: t("source.msys2.technical.archive.title"), text: t("source.msys2.technical.archive.text") },
          { title: t("source.msys2.technical.signature.title"), text: t("source.msys2.technical.signature.text") },
          { title: t("source.msys2.technical.sha.title"), text: t("source.msys2.technical.sha.text") },
        ]}
      >
        <SourceTechnicalField label={t("source.msys2Signature.label")} explanation={t("source.msys2Signature.hint")}>
          <input className="source-field__input" value={buildToolSettings.msys2ArchiveSignatureUrl} onChange={(event) => updateBuildToolSettings({ msys2ArchiveSignatureUrl: event.target.value })} placeholder={t("source.msys2Signature.placeholder")} />
        </SourceTechnicalField>
        <SourceTechnicalField label={t("source.msys2Sha.label")} explanation={t("source.msys2Sha.hint")}>
          <input className="source-field__input source-field__input--mono" value={buildToolSettings.msys2ArchiveSha256Hash} onChange={(event) => updateBuildToolSettings({ msys2ArchiveSha256Hash: event.target.value })} placeholder={t("source.msys2Sha.placeholder")} />
        </SourceTechnicalField>
      </SourceArchiveSection>

      <SourceArchiveSection
        title={t("source.ffmpeg.title")}
        explanation={t("source.ffmpeg.text")}
        archiveLabel={t("source.ffmpegArchive.label")}
        archiveValue={ffmpegBuildSettings.ffmpegSourceArchiveUrl}
        archivePlaceholder={t("source.ffmpegArchive.placeholder")}
        isTechnicalOpen={isFfmpegTechnicalOpen}
        onToggleTechnical={() => setIsFfmpegTechnicalOpen((value) => !value)}
        onArchiveChange={(value) => updateFfmpegBuildSettings({ ffmpegSourceArchiveUrl: value, ffmpegSourceSignatureUrl: value ? `${value}.asc` : "" })}
        links={[
          { label: t("source.links.ffmpegDownload"), url: officialLinks.ffmpegDownload },
          { label: t("source.links.ffmpegReleaseArchive"), url: officialLinks.ffmpegReleaseIndex },
          { label: t("source.links.ffmpegSigningKey"), url: officialLinks.ffmpegSigningKey },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: t("source.ffmpeg.technical.archive.title"), text: t("source.ffmpeg.technical.archive.text") },
          { title: t("source.ffmpeg.technical.signature.title"), text: t("source.ffmpeg.technical.signature.text") },
          { title: t("source.ffmpeg.technical.sha.title"), text: t("source.ffmpeg.technical.sha.text") },
        ]}
      >
        <SourceTechnicalField label={t("source.ffmpegSignature.label")} explanation={t("source.ffmpegSignature.hint")}>
          <input className="source-field__input" value={ffmpegBuildSettings.ffmpegSourceSignatureUrl} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSignatureUrl: event.target.value })} placeholder={t("source.ffmpegSignature.placeholder")} />
        </SourceTechnicalField>
        <SourceTechnicalField label={t("source.ffmpegSha.label")} explanation={t("source.ffmpegSha.hint")}>
          <input className="source-field__input source-field__input--mono" value={ffmpegBuildSettings.ffmpegSourceSha256Hash} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSha256Hash: event.target.value })} placeholder={t("source.ffmpegSha.placeholder")} />
        </SourceTechnicalField>
      </SourceArchiveSection>
    </section>
  );
}

function SourceArchiveSection(props: {
  title: string;
  explanation: string;
  archiveLabel: string;
  archiveValue: string;
  archivePlaceholder: string;
  isTechnicalOpen: boolean;
  onToggleTechnical: () => void;
  onArchiveChange: (value: string) => void;
  links: { label: string; url: string }[];
  technicalSections: { title: string; text: string }[];
  openInUserBrowser: (url: string) => Promise<void>;
  children: React.ReactNode;
}) {
  return (
    <section className="source-section source-archive-section">
      <h2 className="source-section__title">{props.title}</h2>
      <p className="source-explanation">{props.explanation}</p>
      <label className="source-url-row">
        <span className="source-field__label source-field__label--hidden">{props.archiveLabel}</span>
        <input className="source-field__input" value={props.archiveValue} onChange={(event) => props.onArchiveChange(event.target.value)} placeholder={props.archivePlaceholder} />
        <button className="button source-technical-toggle" type="button" aria-expanded={props.isTechnicalOpen} onClick={props.onToggleTechnical}>
          {props.isTechnicalOpen ? t("source.technical.hide") : t("source.technical.show")}
        </button>
      </label>
      {props.isTechnicalOpen && (
        <div className="source-technical-panel">
          <h3 className="source-technical-panel__title">{t("source.technical.detail")}</h3>
          <div className="source-link-group">
            <span className="source-link-group__label">{t("source.links.heading")}</span>
            <div className="source-link-row">
              {props.links.map((link) => <ExternalLinkButton key={link.url} label={link.label} url={link.url} onOpen={props.openInUserBrowser} />)}
            </div>
          </div>
          <div className="source-technical-details">
            {props.technicalSections.map((section) => (
              <section className="source-technical-detail" key={section.title}>
                <h4 className="source-technical-detail__title">{section.title}</h4>
                <p className="source-technical-detail__text">{section.text}</p>
              </section>
            ))}
          </div>
          <div className="source-technical-fields">{props.children}</div>
        </div>
      )}
    </section>
  );
}

function SourceTechnicalField(props: { label: string; explanation: React.ReactNode; children: React.ReactNode }) {
  return (
    <label className="source-technical-field">
      <span className="source-technical-field__label">{props.label}</span>
      <span className="source-technical-field__explanation">{props.explanation}</span>
      {props.children}
    </label>
  );
}
