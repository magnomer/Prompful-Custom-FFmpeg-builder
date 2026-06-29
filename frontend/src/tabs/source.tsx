import React from "react";
import { t } from "../i18n";
import { DescriptionLines, ExternalLinkButton, PageHeader } from "./shared";
import workspaceFolderIcon from "../assets/source-card-icons/WorkspaceFolder.svg";
import buildEnvironmentIcon from "../assets/source-card-icons/BuildEnvironment.svg";
import ffmpegSourceIcon from "../assets/source-card-icons/FFmpegSource.svg";
import browseIcon from "../assets/button-icons/Browse.svg";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";

const officialLinks = {
  msys2ArchiveList: "https://repo.msys2.org/distrib/x86_64/",
  msys2InstallerDocs: "https://www.msys2.org/docs/installer/",
  ffmpegDownload: "https://www.ffmpeg.org/download.html",
  ffmpegReleaseIndex: "https://ffmpeg.org/releases/",
  ffmpegSigningKey: "https://ffmpeg.org/ffmpeg-devel.asc",
};

// The source-version dropdown carries no release data of its own: the supported releases (and
// their archive/.asc URLs) come from the backend via supportedFfmpegReleases. This sentinel marks
// the "no supported release matches the current archive URL" case.
const ffmpegCustomReleaseValue = "custom";

// ffmpegReleaseValueForArchiveUrl returns the version of the supported release whose archive URL
// matches the current value, or the custom sentinel when none does.
function ffmpegReleaseValueForArchiveUrl(supportedFfmpegReleases: FfmpegReleaseChoice[], archiveUrl: string): string {
  const normalized = archiveUrl.trim();
  const match = supportedFfmpegReleases.find((release) => release.archiveUrl === normalized);
  return match ? match.version : ffmpegCustomReleaseValue;
}

function ffmpegReleaseOptionLabel(release: FfmpegReleaseChoice, newestReleaseVersion: string): string {
  const baseLabel = `${release.version} ${release.codename}`;
  return release.version === newestReleaseVersion ? baseLabel : `${baseLabel} (Legacy Support)`;
}

export type SourceTabProps = {
  buildConfigSettings: BuildConfigSettings;
  ffmpegBuildSettings: FfmpegBuildSettings;
  supportedFfmpegReleases: FfmpegReleaseChoice[];
  updateBuildConfigSettings: (partial: Partial<BuildConfigSettings>) => void;
  updateFfmpegBuildSettings: (partial: Partial<FfmpegBuildSettings>) => void;
  updateMsys2ArchiveUrl: (url: string) => void;
  chooseWorkspaceDirectory: () => Promise<void>;
  openInUserBrowser: (url: string) => Promise<void>;
};

export function SourceTab({ buildConfigSettings, ffmpegBuildSettings, supportedFfmpegReleases, updateBuildConfigSettings, updateFfmpegBuildSettings, updateMsys2ArchiveUrl, chooseWorkspaceDirectory, openInUserBrowser }: SourceTabProps) {
  const [isMsys2TechnicalOpen, setIsMsys2TechnicalOpen] = React.useState(false);
  const [isFfmpegTechnicalOpen, setIsFfmpegTechnicalOpen] = React.useState(false);

  return (
    <section className="tab-page source-page">
      <PageHeader title={t("source.title")} text={t("source.briefing")} />

      <section className="card card--blue">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={workspaceFolderIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{t("source.workspace.label")}</h2>
          <DescriptionLines text={t("source.workspace.hint")} />
        </div>
        <label className="card__control">
          <span className="card__label-hidden">{t("source.workspace.label")}</span>
          <input className="card__input" value={buildConfigSettings.workspaceDirectory} onChange={(event) => { updateBuildConfigSettings({ workspaceDirectory: event.target.value }); updateFfmpegBuildSettings({ workspaceDirectory: event.target.value }); }} placeholder={t("source.workspace.placeholder")} />
          <button className="button card__action-btn" type="button" onClick={chooseWorkspaceDirectory}><img className="card__btn-icon" src={browseIcon} alt="" aria-hidden="true" />{t("actions.browse")}</button>
        </label>
      </section>

      <SourceArchiveSection
        variant="teal"
        icon={buildEnvironmentIcon}
        title={t("source.msys2.title")}
        explanation={t("source.msys2.text")}
        archiveLabel={t("source.msys2Archive.label")}
        archiveValue={buildConfigSettings.msys2ArchiveUrl}
        archivePlaceholder={t("source.msys2Archive.placeholder")}
        isTechnicalOpen={isMsys2TechnicalOpen}
        onToggleTechnical={() => setIsMsys2TechnicalOpen((value) => !value)}
        onArchiveChange={updateMsys2ArchiveUrl}
        intro={t("source.msys2.technical.intro")}
        links={[
          { label: t("source.links.msys2ArchiveList"), url: officialLinks.msys2ArchiveList },
          { label: t("source.links.msys2Notes"), url: officialLinks.msys2InstallerDocs },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: t("source.msys2.technical.archive.title"), text: t("source.msys2.technical.archive.text") },
        ]}
      >
        <SourceTechnicalField label={t("source.msys2Signature.label")} explanation={t("source.msys2Signature.hint")}>
          <input className="card__input" value={buildConfigSettings.msys2ArchiveSignatureUrl} onChange={(event) => updateBuildConfigSettings({ msys2ArchiveSignatureUrl: event.target.value })} placeholder={t("source.msys2Signature.placeholder")} />
        </SourceTechnicalField>
        <SourceTechnicalField label={t("source.msys2Sha.label")} explanation={t("source.msys2.technical.sha.text")}>
          <input className="card__input card__input--mono" value={buildConfigSettings.msys2ArchiveSha256Hash} onChange={(event) => updateBuildConfigSettings({ msys2ArchiveSha256Hash: event.target.value })} placeholder={t("source.msys2Sha.placeholder")} />
        </SourceTechnicalField>
      </SourceArchiveSection>

      <SourceArchiveSection
        variant="purple"
        icon={ffmpegSourceIcon}
        title={t("source.ffmpeg.title")}
        explanation={t("source.ffmpeg.text")}
        archiveLabel={t("source.ffmpegArchive.label")}
        archiveValue={ffmpegBuildSettings.ffmpegSourceArchiveUrl}
        archivePlaceholder={t("source.ffmpegArchive.placeholder")}
        isTechnicalOpen={isFfmpegTechnicalOpen}
        onToggleTechnical={() => setIsFfmpegTechnicalOpen((value) => !value)}
        onArchiveChange={(value) => {
          const matched = supportedFfmpegReleases.find((release) => release.archiveUrl === value.trim());
          updateFfmpegBuildSettings({ ffmpegSourceArchiveUrl: value, ffmpegSourceSignatureUrl: matched ? matched.signatureUrl : "" });
        }}
        versionSelect={(
          <label className="card__control">
            <span className="card__label-hidden">{t("source.ffmpegVersion.label")}</span>
            <select
              className="card__input"
              value={ffmpegReleaseValueForArchiveUrl(supportedFfmpegReleases, ffmpegBuildSettings.ffmpegSourceArchiveUrl)}
              onChange={(event) => {
                const selected = event.target.value;
                const release = supportedFfmpegReleases.find((candidate) => candidate.version === selected);
                if (!release) {
                  updateFfmpegBuildSettings({ ffmpegSourceArchiveUrl: "", ffmpegSourceSignatureUrl: "" });
                  return;
                }
                updateFfmpegBuildSettings({ ffmpegSourceArchiveUrl: release.archiveUrl, ffmpegSourceSignatureUrl: release.signatureUrl });
              }}
            >
              {supportedFfmpegReleases.map((release) => (
                <option key={release.version} value={release.version}>{ffmpegReleaseOptionLabel(release, supportedFfmpegReleases[0]?.version ?? "")}</option>
              ))}
              <option value={ffmpegCustomReleaseValue}>{t("source.ffmpegVersion.custom")}</option>
            </select>
          </label>
        )}
        links={[
          { label: t("source.links.ffmpegDownload"), url: officialLinks.ffmpegDownload },
          { label: t("source.links.ffmpegReleaseArchive"), url: officialLinks.ffmpegReleaseIndex },
          { label: t("source.links.ffmpegSigningKey"), url: officialLinks.ffmpegSigningKey },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: t("source.ffmpeg.technical.archive.title"), text: t("source.ffmpeg.technical.archive.text") },
        ]}
      >
        <SourceTechnicalField label={t("source.ffmpegSignature.label")} explanation={t("source.ffmpegSignature.hint")}>
          <input className="card__input" value={ffmpegBuildSettings.ffmpegSourceSignatureUrl} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSignatureUrl: event.target.value })} placeholder={t("source.ffmpegSignature.placeholder")} />
        </SourceTechnicalField>
        <SourceTechnicalField label={t("source.ffmpegSha.label")} explanation={t("source.ffmpeg.technical.sha.text")}>
          <input className="card__input card__input--mono" value={ffmpegBuildSettings.ffmpegSourceSha256Hash} onChange={(event) => updateFfmpegBuildSettings({ ffmpegSourceSha256Hash: event.target.value })} placeholder={t("source.ffmpegSha.placeholder")} />
        </SourceTechnicalField>
      </SourceArchiveSection>
    </section>
  );
}


function SourceArchiveSection(props: {
  variant: string;
  icon: string;
  title: string;
  explanation: string;
  archiveLabel: string;
  archiveValue: string;
  archivePlaceholder: string;
  isTechnicalOpen: boolean;
  onToggleTechnical: () => void;
  onArchiveChange: (value: string) => void;
  intro?: string;
  versionSelect?: React.ReactNode;
  links: { label: string; url: string }[];
  technicalSections: { title: string; text: string }[];
  openInUserBrowser: (url: string) => Promise<void>;
  children: React.ReactNode;
}) {
  return (
    <section className={`card card--${props.variant}`}>
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={props.icon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{props.title}</h2>
        <DescriptionLines text={props.explanation} />
      </div>
      {props.versionSelect}
      <label className="card__control">
        <span className="card__label-hidden">{props.archiveLabel}</span>
        <input className="card__input" value={props.archiveValue} onChange={(event) => props.onArchiveChange(event.target.value)} placeholder={props.archivePlaceholder} />
        <button className="button card__toggle" type="button" aria-expanded={props.isTechnicalOpen} onClick={props.onToggleTechnical}>
          <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
          {props.isTechnicalOpen ? t("source.technical.hide") : t("source.technical.show")}
        </button>
      </label>
      {props.isTechnicalOpen && (
        <div className="card__panel">
          <h3 className="card__panel-title">{t("source.technical.detail")}</h3>
          {props.intro && <DescriptionLines text={props.intro} className="card__intro" />}
          <div className="card__links">
            <span className="card__links-label">{t("source.links.heading")}</span>
            <div className="card__link-row">
              {props.links.map((link) => <ExternalLinkButton key={link.url} label={link.label} url={link.url} onOpen={props.openInUserBrowser} />)}
            </div>
          </div>
          <div className="card__details">
            {props.technicalSections.map((section) => (
              <section className="card__detail" key={section.title}>
                <h4 className="card__detail-title">{section.title}</h4>
                <DescriptionLines text={section.text} className="card__detail-text" groupSentences />
              </section>
            ))}
          </div>
          <div className="card__fields">{props.children}</div>
        </div>
      )}
    </section>
  );
}

function SourceTechnicalField(props: { label: string; explanation: string; children: React.ReactNode }) {
  return (
    <label className="card__field">
      <span className="card__field-label">{props.label}</span>
      <DescriptionLines text={props.explanation} className="card__field-desc" groupSentences />
      {props.children}
    </label>
  );
}
