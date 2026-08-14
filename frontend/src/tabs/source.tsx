import React from "react";
import { LLocaleTextGet } from "../i18n";
import { PTextDescriptionRender, PLinkExternalRender, PHeaderPageRender } from "./shared";
import workspaceFolderIcon from "../assets/source-card-icons/WorkspaceFolder.svg";
import buildEnvironmentIcon from "../assets/source-card-icons/BuildEnvironment.svg";
import ffmpegSourceIcon from "../assets/source-card-icons/FfmpegSource.svg";
import browseIcon from "../assets/button-icons/Browse.svg";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";

const LLinkOfficialTable = {
  msys2ArchiveList: "https://repo.msys2.org/distrib/x86_64/",
  msys2InstallerDocs: "https://www.msys2.org/docs/installer/",
  ffmpegDownload: "https://www.ffmpeg.org/download.html",
  LReleaseFfmpegIndex: "https://ffmpeg.org/releases/",
  ffmpegSigningKey: "https://ffmpeg.org/ffmpeg-devel.asc",
};

// The source-version dropdown carries no release data of its own: the supported releases (and
// their archive/.asc URLs) come from the backend via supportedFfmpegReleases. This sentinel marks
// the "no supported release matches the current archive URL" case.
const LReleaseCustomValue = "custom";

// LReleaseArchiveGet returns the version of the supported release whose archive URL
// matches the current value, or the custom sentinel when none does.
function LReleaseArchiveGet(supportedFfmpegReleases: LReleaseChoice[], archiveUrl: string): string {
  const normalized = archiveUrl.trim();
  const match = supportedFfmpegReleases.find((release) => release.archiveUrl === normalized);
  return match ? match.version : LReleaseCustomValue;
}

function LReleaseOptionGet(release: LReleaseChoice, newestReleaseVersion: string): string {
  const baseLabel = `${release.version} ${release.codename}`;
  return release.version === newestReleaseVersion ? baseLabel : `${baseLabel} (Legacy Support)`;
}

export type LSourceProperties = {
  buildConfigSettings: LSettingsToolchain;
  ffmpegBuildSettings: LSettingsFfmpeg;
  supportedFfmpegReleases: LReleaseChoice[];
  LSettingsToolchainUpdate: (partial: Partial<LSettingsToolchain>) => void;
  LSettingsFfmpegUpdate: (partial: Partial<LSettingsFfmpeg>) => void;
  LMSYSArchiveUpdate: (url: string) => void;
  chooseWorkspaceDirectory: () => Promise<void>;
  openInUserBrowser: (url: string) => Promise<void>;
};

export function PSourceRender({ buildConfigSettings, ffmpegBuildSettings, supportedFfmpegReleases, LSettingsToolchainUpdate, LSettingsFfmpegUpdate, LMSYSArchiveUpdate, chooseWorkspaceDirectory, openInUserBrowser }: LSourceProperties) {
  const [isMsys2TechnicalOpen, setIsMsys2TechnicalOpen] = React.useState(false);
  const [isFfmpegTechnicalOpen, setIsFfmpegTechnicalOpen] = React.useState(false);

  return (
    <section className="tab-page source-page">
      <PHeaderPageRender title={LLocaleTextGet("source.title")} text={LLocaleTextGet("source.briefing")} />

      <section className="card card--blue">
        <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={workspaceFolderIcon} alt="" /></span>
        <div className="card__head">
          <h2 className="card__title">{LLocaleTextGet("source.workspace.label")}</h2>
          <PTextDescriptionRender text={LLocaleTextGet("source.workspace.hint")} />
        </div>
        <label className="card__control">
          <span className="card__label-hidden">{LLocaleTextGet("source.workspace.label")}</span>
          <input className="card__input" value={buildConfigSettings.workspaceDirectory} onChange={(event) => { LSettingsToolchainUpdate({ workspaceDirectory: event.target.value }); LSettingsFfmpegUpdate({ workspaceDirectory: event.target.value }); }} placeholder={LLocaleTextGet("source.workspace.placeholder")} />
          <button className="button card__action-btn" type="button" onClick={chooseWorkspaceDirectory}><img className="card__btn-icon" src={browseIcon} alt="" aria-hidden="true" />{LLocaleTextGet("actions.browse")}</button>
        </label>
      </section>

      <PSourceArchiveRender
        variant="teal"
        icon={buildEnvironmentIcon}
        title={LLocaleTextGet("source.msys2.title")}
        explanation={LLocaleTextGet("source.msys2.text")}
        archiveLabel={LLocaleTextGet("source.msys2Archive.label")}
        archiveValue={buildConfigSettings.msys2ArchiveUrl}
        archivePlaceholder={LLocaleTextGet("source.msys2Archive.placeholder")}
        isTechnicalOpen={isMsys2TechnicalOpen}
        onToggleTechnical={() => setIsMsys2TechnicalOpen((value) => !value)}
        onArchiveChange={LMSYSArchiveUpdate}
        intro={LLocaleTextGet("source.msys2.technical.intro")}
        links={[
          { label: LLocaleTextGet("source.links.msys2ArchiveList"), url: LLinkOfficialTable.msys2ArchiveList },
          { label: LLocaleTextGet("source.links.msys2Notes"), url: LLinkOfficialTable.msys2InstallerDocs },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: LLocaleTextGet("source.msys2.technical.archive.title"), text: LLocaleTextGet("source.msys2.technical.archive.text") },
        ]}
      >
        <PSourceTechnicalRender label={LLocaleTextGet("source.msys2Signature.label")} explanation={LLocaleTextGet("source.msys2Signature.hint")}>
          <input className="card__input" value={buildConfigSettings.msys2ArchiveSignatureUrl} onChange={(event) => LSettingsToolchainUpdate({ msys2ArchiveSignatureUrl: event.target.value })} placeholder={LLocaleTextGet("source.msys2Signature.placeholder")} />
        </PSourceTechnicalRender>
        <PSourceTechnicalRender label={LLocaleTextGet("source.msys2Sha.label")} explanation={LLocaleTextGet("source.msys2.technical.sha.text")}>
          <input className="card__input card__input--mono" value={buildConfigSettings.msys2ArchiveSha256Hash} onChange={(event) => LSettingsToolchainUpdate({ msys2ArchiveSha256Hash: event.target.value })} placeholder={LLocaleTextGet("source.msys2Sha.placeholder")} />
        </PSourceTechnicalRender>
      </PSourceArchiveRender>

      <PSourceArchiveRender
        variant="purple"
        icon={ffmpegSourceIcon}
        title={LLocaleTextGet("source.ffmpeg.title")}
        explanation={LLocaleTextGet("source.ffmpeg.text")}
        archiveLabel={LLocaleTextGet("source.ffmpegArchive.label")}
        archiveValue={ffmpegBuildSettings.ffmpegSourceArchiveUrl}
        archivePlaceholder={LLocaleTextGet("source.ffmpegArchive.placeholder")}
        isTechnicalOpen={isFfmpegTechnicalOpen}
        onToggleTechnical={() => setIsFfmpegTechnicalOpen((value) => !value)}
        onArchiveChange={(value) => {
          const matched = supportedFfmpegReleases.find((release) => release.archiveUrl === value.trim());
          LSettingsFfmpegUpdate({ ffmpegSourceArchiveUrl: value, ffmpegSourceSignatureUrl: matched ? matched.signatureUrl : "" });
        }}
        versionSelect={(
          <label className="card__control">
            <span className="card__label-hidden">{LLocaleTextGet("source.ffmpegVersion.label")}</span>
            <select
              className="card__input"
              value={LReleaseArchiveGet(supportedFfmpegReleases, ffmpegBuildSettings.ffmpegSourceArchiveUrl)}
              onChange={(event) => {
                const selected = event.target.value;
                const release = supportedFfmpegReleases.find((candidate) => candidate.version === selected);
                if (!release) {
                  LSettingsFfmpegUpdate({ ffmpegSourceArchiveUrl: "", ffmpegSourceSignatureUrl: "" });
                  return;
                }
                LSettingsFfmpegUpdate({ ffmpegSourceArchiveUrl: release.archiveUrl, ffmpegSourceSignatureUrl: release.signatureUrl });
              }}
            >
              {supportedFfmpegReleases.map((release) => (
                <option key={release.version} value={release.version}>{LReleaseOptionGet(release, supportedFfmpegReleases[0]?.version ?? "")}</option>
              ))}
              <option value={LReleaseCustomValue}>{LLocaleTextGet("source.ffmpegVersion.custom")}</option>
            </select>
          </label>
        )}
        links={[
          { label: LLocaleTextGet("source.links.ffmpegDownload"), url: LLinkOfficialTable.ffmpegDownload },
          { label: LLocaleTextGet("source.links.ffmpegReleaseArchive"), url: LLinkOfficialTable.LReleaseFfmpegIndex },
          { label: LLocaleTextGet("source.links.ffmpegSigningKey"), url: LLinkOfficialTable.ffmpegSigningKey },
        ]}
        openInUserBrowser={openInUserBrowser}
        technicalSections={[
          { title: LLocaleTextGet("source.ffmpeg.technical.archive.title"), text: LLocaleTextGet("source.ffmpeg.technical.archive.text") },
        ]}
      >
        <PSourceTechnicalRender label={LLocaleTextGet("source.ffmpegSignature.label")} explanation={LLocaleTextGet("source.ffmpegSignature.hint")}>
          <input className="card__input" value={ffmpegBuildSettings.ffmpegSourceSignatureUrl} onChange={(event) => LSettingsFfmpegUpdate({ ffmpegSourceSignatureUrl: event.target.value })} placeholder={LLocaleTextGet("source.ffmpegSignature.placeholder")} />
        </PSourceTechnicalRender>
        <PSourceTechnicalRender label={LLocaleTextGet("source.ffmpegSha.label")} explanation={LLocaleTextGet("source.ffmpeg.technical.sha.text")}>
          <input className="card__input card__input--mono" value={ffmpegBuildSettings.ffmpegSourceSha256Hash} onChange={(event) => LSettingsFfmpegUpdate({ ffmpegSourceSha256Hash: event.target.value })} placeholder={LLocaleTextGet("source.ffmpegSha.placeholder")} />
        </PSourceTechnicalRender>
      </PSourceArchiveRender>
    </section>
  );
}


function PSourceArchiveRender(props: {
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
        <PTextDescriptionRender text={props.explanation} />
      </div>
      {props.versionSelect}
      <label className="card__control">
        <span className="card__label-hidden">{props.archiveLabel}</span>
        <input className="card__input" value={props.archiveValue} onChange={(event) => props.onArchiveChange(event.target.value)} placeholder={props.archivePlaceholder} />
        <button className="button card__toggle" type="button" aria-expanded={props.isTechnicalOpen} onClick={props.onToggleTechnical}>
          <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
          {props.isTechnicalOpen ? LLocaleTextGet("source.technical.hide") : LLocaleTextGet("source.technical.show")}
        </button>
      </label>
      {props.isTechnicalOpen && (
        <div className="card__panel">
          <h3 className="card__panel-title">{LLocaleTextGet("source.technical.detail")}</h3>
          {props.intro && <PTextDescriptionRender text={props.intro} className="card__intro" />}
          <div className="card__links">
            <span className="card__links-label">{LLocaleTextGet("source.links.heading")}</span>
            <div className="card__link-row">
              {props.links.map((link) => <PLinkExternalRender key={link.url} label={link.label} url={link.url} onOpen={props.openInUserBrowser} />)}
            </div>
          </div>
          <div className="card__details">
            {props.technicalSections.map((section) => (
              <section className="card__detail" key={section.title}>
                <h4 className="card__detail-title">{section.title}</h4>
                <PTextDescriptionRender text={section.text} className="card__detail-text" groupSentences />
              </section>
            ))}
          </div>
          <div className="card__fields">{props.children}</div>
        </div>
      )}
    </section>
  );
}

function PSourceTechnicalRender(props: { label: string; explanation: string; children: React.ReactNode }) {
  return (
    <label className="card__field">
      <span className="card__field-label">{props.label}</span>
      <PTextDescriptionRender text={props.explanation} className="card__field-desc" groupSentences />
      {props.children}
    </label>
  );
}
