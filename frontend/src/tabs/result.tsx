import React from "react";
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

function fileKind(fileName: string): "exe" | "dll" | "other" {
  const lower = fileName.toLowerCase();
  if (lower.endsWith(".exe")) return "exe";
  if (lower.endsWith(".dll")) return "dll";
  return "other";
}

function formatDateTime(value: string): string {
  if (!value) return t("result.summary.latestBuild.unknown");
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function buildTypeLabel(_result: BuildResult): string {
  return t("result.summary.buildType.custom");
}

// ─── ResultPanel ─────────────────────────────────────────────────────────────

function ResultSummary(props: { result: BuildResult }) {
  const totalSizeBytes = props.result.files.reduce((sum, file) => sum + file.sizeBytes, 0);
  const stats = [
    { label: t("result.summary.totalFiles"), value: String(props.result.files.length), unit: t("result.summary.filesUnit") },
    { label: t("result.summary.totalSize"), value: formatBytes(totalSizeBytes), unit: "" },
    { label: t("result.summary.buildType"), value: buildTypeLabel(props.result), unit: "" },
    { label: t("result.summary.latestBuild"), value: formatDateTime(props.result.createdAt), unit: "" },
  ];
  return (
    <section className="result-summary" aria-label={t("result.summary.ariaLabel")}>
      {stats.map((stat) => (
        <article className="result-summary__card" key={stat.label}>
          <span className="result-summary__label">{stat.label}</span>
          <strong className="result-summary__value">{stat.value}</strong>
          {stat.unit && <span className="result-summary__unit">{stat.unit}</span>}
        </article>
      ))}
    </section>
  );
}

function ResultLocationPanel(props: { result: BuildResult; onOpenFolder: () => Promise<void>; onOpenReport: () => Promise<void> }) {
  const rows = [
    { label: t("result.metadata.folder"), value: props.result.artifactsDirectory, actionLabel: t("result.actions.open") },
    { label: t("result.metadata.latestReport"), value: props.result.reportPath || t("result.metadata.noReport"), actionLabel: t("result.actions.open") },
  ];
  return (
    <section className="result-location-panel">
      {rows.map((row, index) => (
        <div className="result-location-row" key={row.label}>
          <strong className="result-location-row__label">{row.label}</strong>
          <span className="result-location-row__value">{row.value}</span>
          {index === 0 ? (
            <button className="button button--primary result-location-row__button" type="button" onClick={props.onOpenFolder}>{row.actionLabel}</button>
          ) : props.result.reportPath ? (
            <button className="button result-location-row__button" type="button" onClick={props.onOpenReport}>{row.actionLabel}</button>
          ) : <span className="result-location-row__button-spacer" />}
        </div>
      ))}
    </section>
  );
}

function ResultFiles(props: { files: BuildResultFile[] }) {
  const [query, setQuery] = React.useState("");
  const [filterKind, setFilterKind] = React.useState<"all" | "exe" | "dll" | "other">("all");
  const [sortMode, setSortMode] = React.useState<"name" | "size-desc" | "size-asc">("name");

  const visibleFiles = React.useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const filtered = props.files.filter((file) => {
      const matchesKind = filterKind === "all" || fileKind(file.name) === filterKind;
      const matchesQuery = !normalizedQuery || file.name.toLowerCase().includes(normalizedQuery) || file.path.toLowerCase().includes(normalizedQuery) || file.sha256Hash.toLowerCase().includes(normalizedQuery);
      return matchesKind && matchesQuery;
    });
    return filtered.sort((a, b) => {
      if (sortMode === "size-desc") return b.sizeBytes - a.sizeBytes || a.name.localeCompare(b.name);
      if (sortMode === "size-asc") return a.sizeBytes - b.sizeBytes || a.name.localeCompare(b.name);
      return a.name.localeCompare(b.name);
    });
  }, [filterKind, props.files, query, sortMode]);

  return (
    <section className="result-files">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{t("result.files.title")}</h2>
        <span className="result-files__count">{t("result.files.count", { count: visibleFiles.length, total: props.files.length })}</span>
      </div>
      {props.files.length > 0 && (
        <div className="result-files__controls">
          <input
            className="result-files__search"
            type="search"
            value={query}
            placeholder={t("result.files.searchPlaceholder")}
            aria-label={t("result.files.searchAriaLabel")}
            onChange={(event) => setQuery(event.currentTarget.value)}
          />
          <select className="result-files__select" value={filterKind} aria-label={t("result.files.filterAriaLabel")} onChange={(event) => setFilterKind(event.currentTarget.value as typeof filterKind)}>
            <option value="all">{t("result.files.filter.all")}</option>
            <option value="exe">{t("result.files.filter.exe")}</option>
            <option value="dll">{t("result.files.filter.dll")}</option>
            <option value="other">{t("result.files.filter.other")}</option>
          </select>
          <select className="result-files__select" value={sortMode} aria-label={t("result.files.sortAriaLabel")} onChange={(event) => setSortMode(event.currentTarget.value as typeof sortMode)}>
            <option value="name">{t("result.files.sort.name")}</option>
            <option value="size-desc">{t("result.files.sort.sizeDesc")}</option>
            <option value="size-asc">{t("result.files.sort.sizeAsc")}</option>
          </select>
        </div>
      )}
      {props.files.length === 0 ? <p className="empty-text">{t("result.files.empty")}</p> : (
        <div className="result-file-list">
          {visibleFiles.length === 0 ? <p className="empty-text result-file-list__empty">{t("result.files.noMatches")}</p> : visibleFiles.map((file) => (
            <article className="result-file" key={file.path}>
              <span className={`result-file__type result-file__type--${fileKind(file.name)}`}>{fileKind(file.name).toUpperCase()}</span>
              <strong className="result-file__name">{file.name}</strong>
              <span className="result-file__size">{formatBytes(file.sizeBytes)}</span>
              <span className="result-file__path">{file.path}</span>
              {file.sha256Hash && <span className="result-file__hash">{t("result.files.sha256", { hash: file.sha256Hash })}</span>}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}


type ResultDetailTabId = "files" | "libraries" | "options" | "flags";

function ResultPlanItems(props: { title: string; items: string[]; emptyText: string; code?: boolean }) {
  return (
    <section className="result-plan-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{props.title}</h2>
        <span className="result-files__count">{t("result.details.count", { count: props.items.length })}</span>
      </div>
      {props.items.length === 0 ? <p className="empty-text">{props.emptyText}</p> : (
        <div className="result-file-list result-plan-list">
          {props.items.map((item, index) => (
            <article className="result-plan-item" key={`${item}-${index}`}>
              <span className="result-plan-item__index">{String(index + 1).padStart(2, "0")}</span>
              {props.code ? <code className="result-plan-item__text result-plan-item__text--code">{item}</code> : <span className="result-plan-item__text">{item}</span>}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}

function ResultDetailTabs(props: { result: BuildResult }) {
  const [activeTabId, setActiveTabId] = React.useState<ResultDetailTabId>("files");
  const libraries = React.useMemo(() => props.result.selectedLibraries.map(resultLibraryLabel), [props.result.selectedLibraries]);
  const options = React.useMemo(() => props.result.selectedConfigureOptions.map(resultOptionLabel), [props.result.selectedConfigureOptions]);
  const tabs: { id: ResultDetailTabId; label: string; count: number }[] = [
    { id: "files", label: t("result.details.tabs.files"), count: props.result.files.length },
    { id: "libraries", label: t("result.details.tabs.libraries"), count: libraries.length },
    { id: "options", label: t("result.details.tabs.options"), count: options.length },
    { id: "flags", label: t("result.details.tabs.flags"), count: props.result.configureFlags.length },
  ];

  return (
    <section className="result-details-card">
      <div className="result-details-tabs" role="tablist" aria-label={t("result.details.tabs.ariaLabel")}>
        {tabs.map((tab) => (
          <button
            className={`result-details-tab ${activeTabId === tab.id ? "result-details-tab--active" : ""}`}
            type="button"
            role="tab"
            aria-selected={activeTabId === tab.id}
            key={tab.id}
            onClick={() => setActiveTabId(tab.id)}
          >
            <span>{tab.label}</span>
            <span className="result-details-tab__count">{tab.count}</span>
          </button>
        ))}
      </div>
      <div className="result-details-body">
        {activeTabId === "files" && <ResultFiles files={props.result.files} />}
        {activeTabId === "libraries" && <ResultPlanItems title={t("result.review.libraries")} items={libraries} emptyText={t("result.details.empty.libraries")} />}
        {activeTabId === "options" && <ResultPlanItems title={t("result.review.options")} items={options} emptyText={t("result.details.empty.options")} />}
        {activeTabId === "flags" && <ResultPlanItems title={t("approval.review.finalConfigureFlags")} items={props.result.configureFlags} emptyText={t("result.details.empty.flags")} code />}
      </div>
    </section>
  );
}

function ResultPanel(props: { result: BuildResult | null; errorText: string; onOpenFolder: () => Promise<void>; onOpenReport: () => Promise<void> }) {
  const result = props.result;
  return (
    <section className="result-panel">
      {props.errorText && <p className="empty-text">{props.errorText}</p>}
      {!props.errorText && !result && <p className="empty-text">{t("result.empty")}</p>}
      {result && (
        <>
          <ResultSummary result={result} />
          <ResultLocationPanel result={result} onOpenFolder={props.onOpenFolder} onOpenReport={props.onOpenReport} />
          <ResultDetailTabs result={result} />
          {result.requiredMsys2PackageNames.length > 0 && (
            <ResultPlanItems title={t("result.review.packages")} items={result.requiredMsys2PackageNames} emptyText={t("result.details.empty.packages")} code />
          )}
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
  openResultReport: () => Promise<void>;
};

export function ResultTab({ buildResult, buildResultError, refreshBuildResult, openResultFolder, openResultReport }: ResultTabProps) {
  return (
    <section className="tab-page result-page">
      <header className="result-page__header">
        <div>
          <h1 className="page-header__title">{t("result.title")}</h1>
          <p className="page-header__text">{t("result.intro")}</p>
        </div>
        <div className="result-page__actions">
          <button className="button" type="button" onClick={refreshBuildResult}>{t("result.actions.refresh")}</button>
        </div>
      </header>
      <ResultPanel result={buildResult} errorText={buildResultError} onOpenFolder={openResultFolder} onOpenReport={openResultReport} />
    </section>
  );
}
