import React from "react";
import { PageHeader, ReviewList } from "./shared";
import { t } from "../i18n";
import { catalogOptionName, libraryLicenseLabel } from "../catalogText";

// ─── Helpers ─────────────────────────────────────────────────────────────────

function formatBytes(sizeBytes: number): string {
  if (sizeBytes < 1024) return `${sizeBytes} ${t("result.size.byte")}`;
  if (sizeBytes < 1024 * 1024) return `${(sizeBytes / 1024).toFixed(1)} ${t("result.size.kilobyte")}`;
  if (sizeBytes < 1024 * 1024 * 1024) return `${(sizeBytes / 1024 / 1024).toFixed(1)} ${t("result.size.megabyte")}`;
  return `${(sizeBytes / 1024 / 1024 / 1024).toFixed(2)} ${t("result.size.gigabyte")}`;
}

function resultLibraryLabel(raw: string): string {
  const match = raw.match(/^library:([^:]+):(.+)$/);
  if (!match) return raw;
  const [, libraryId, licenseEffectName] = match;
  const name = t(`catalog.libraries.${libraryId}.displayName`);
  return `${name} (${libraryLicenseLabel(licenseEffectName)})`;
}

function resultOptionLabel(raw: string): string {
  const match = raw.match(/^option:(.+)$/);
  if (!match) return raw;
  return catalogOptionName(match[1]);
}

// ─── ResultPanel ─────────────────────────────────────────────────────────────

function ResultPanel(props: { result: BuildResult | null; errorText: string; onRefresh: () => Promise<void>; onOpenFolder: () => Promise<void> }) {
  const result = props.result;
  return (
    <section className="result-panel">
      <div className="actions">
        <button className="button button--primary" type="button" onClick={props.onOpenFolder}>{t("result.actions.openFolder")}</button>
        <button className="button" type="button" onClick={props.onRefresh}>{t("result.actions.refresh")}</button>
      </div>
      {props.errorText && <p className="empty-text">{props.errorText}</p>}
      {!props.errorText && !result && <p className="empty-text">{t("result.empty")}</p>}
      {result && (
        <>
          <dl className="metadata">
            <dt>{t("result.metadata.folder")}</dt><dd className="metadata__hash">{result.artifactsDirectory}</dd>
            <dt>{t("result.metadata.latestReport")}</dt><dd className="metadata__hash">{result.reportPath || t("result.metadata.noReport")}</dd>
          </dl>
          <section className="result-files">
            <h2 className="review-list__title">{t("result.files.title")}</h2>
            {result.files.length === 0 ? <p className="empty-text">{t("result.files.empty")}</p> : result.files.map((file) => (
              <div className="result-file" key={file.path}>
                <strong>{file.name}</strong>
                <span>{formatBytes(file.sizeBytes)}</span>
                <span className="metadata__hash">{file.path}</span>
                {file.sha256Hash && <span className="metadata__hash">{t("result.files.sha256", { hash: file.sha256Hash })}</span>}
              </div>
            ))}
          </section>
          {result.selectedLibraries.length > 0 && <ReviewList title={t("result.review.libraries")} items={result.selectedLibraries.map(resultLibraryLabel)} />}
          {result.selectedConfigureOptions.length > 0 && <ReviewList title={t("result.review.options")} items={result.selectedConfigureOptions.map(resultOptionLabel)} />}
          {result.configureFlags.length > 0 && <ReviewList title={t("approval.review.finalConfigureFlags")} items={result.configureFlags} />}
          {result.requiredMsys2PackageNames.length > 0 && <ReviewList title={t("result.review.packages")} items={result.requiredMsys2PackageNames} />}
        </>
      )}
    </section>
  );
}

// ─── ResultTab ───────────────────────────────────────────────────────────────

export type ResultTabProps = {
  buildResult: BuildResult | null;
  buildResultError: string;
  refreshBuildResult: () => Promise<void>;
  openResultFolder: () => Promise<void>;
};

export function ResultTab({ buildResult, buildResultError, refreshBuildResult, openResultFolder }: ResultTabProps) {
  return (
    <section className="tab-page">
      <PageHeader title={t("result.title")} text={t("result.intro")} />
      <ResultPanel result={buildResult} errorText={buildResultError} onRefresh={refreshBuildResult} onOpenFolder={openResultFolder} />
    </section>
  );
}
