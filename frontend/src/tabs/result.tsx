import React from "react";
import { t } from "../i18n";
import { catalogOptionName, libraryLicenseLabel } from "../catalogText";
import { licenseBoundaryShortLabel } from "./options";
import { ClipboardSetText } from "../../wailsjs/runtime/runtime";
import emptyStateBlueIcon from "../assets/empty-card-icons/EmptyStateBlue.svg";
import emptyStatePurpleIcon from "../assets/empty-card-icons/EmptyStatePurple.svg";
import copyIcon from "../assets/button-icons/Copy.svg";

function ResultFileCopyButton(props: { hash: string }) {
  const [copied, setCopied] = React.useState(false);
  const resetTimer = React.useRef<ReturnType<typeof setTimeout> | null>(null);
  React.useEffect(() => () => { if (resetTimer.current) clearTimeout(resetTimer.current); }, []);
  async function copy() {
    await ClipboardSetText(props.hash);
    setCopied(true);
    if (resetTimer.current) clearTimeout(resetTimer.current);
    resetTimer.current = setTimeout(() => setCopied(false), 1400);
  }
  return (
    <button
      className={`result-file__copy ${copied ? "result-file__copy--copied" : ""}`}
      type="button"
      onClick={copy}
      title={copied ? t("result.files.copied") : t("result.files.copyHash")}
      aria-label={copied ? t("result.files.copied") : t("result.files.copyHash")}
    >
      <img className="result-file__copy-icon" src={copyIcon} alt="" aria-hidden="true" />
    </button>
  );
}

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
  const stats: { label: string; value: string; unit: string; danger?: boolean }[] = [
    { label: t("result.summary.totalFiles"), value: String(props.result.files.length), unit: t("result.summary.filesUnit") },
    { label: t("result.summary.totalSize"), value: formatBytes(totalSizeBytes), unit: "" },
    { label: t("result.summary.buildType"), value: buildTypeLabel(props.result), unit: "" },
    { label: t("result.summary.license"), value: licenseBoundaryShortLabel(props.result.licenseProfileName), unit: "", danger: props.result.licenseProfileName === "nonfree-local" },
    { label: t("result.summary.latestBuild"), value: formatDateTime(props.result.createdAt), unit: "" },
  ];
  return (
    <section className="result-summary" aria-label={t("result.summary.ariaLabel")}>
      {stats.map((stat) => (
        <article className="result-summary__card" key={stat.label}>
          <span className="result-summary__label">{stat.label}</span>
          <strong className={`result-summary__value ${stat.danger ? "result-summary__value--danger" : ""}`}>{stat.value}</strong>
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
              {file.sha256Hash && <ResultFileCopyButton hash={file.sha256Hash} />}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}


type ResultDetailTabId = "files" | "libraries" | "packages" | "options" | "flags";

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

function ResultDetailTabs(props: { result: BuildResult; verification: BuildVerification | null; verificationError: string; isVerifying: boolean }) {
  const [activeTabId, setActiveTabId] = React.useState<ResultDetailTabId>("files");
  const options = React.useMemo(() => props.result.selectedConfigureOptions.map(resultOptionLabel), [props.result.selectedConfigureOptions]);

  // Jump to the Libraries tab when a verification run finishes or starts, so its
  // results surface where the user is looking instead of on a hidden tab.
  React.useEffect(() => {
    if (props.isVerifying || props.verification || props.verificationError) setActiveTabId("libraries");
  }, [props.isVerifying, props.verification, props.verificationError]);
  const tabs: { id: ResultDetailTabId; label: string; count: number }[] = [
    { id: "files", label: t("result.details.tabs.files"), count: props.result.files.length },
    { id: "libraries", label: t("result.details.tabs.libraries"), count: props.result.selectedLibraries.length },
    { id: "packages", label: t("result.details.tabs.packages"), count: props.result.requiredMsys2PackageNames.length },
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
        {activeTabId === "libraries" && <ResultLibraryItems librarySpecs={props.result.selectedLibraries} verification={props.verification} verificationError={props.verificationError} isVerifying={props.isVerifying} />}
        {activeTabId === "packages" && <ResultPlanItems title={t("result.review.packages")} items={props.result.requiredMsys2PackageNames} emptyText={t("result.details.empty.packages")} code />}
        {activeTabId === "options" && <ResultPlanItems title={t("result.review.options")} items={options} emptyText={t("result.details.empty.options")} />}
        {activeTabId === "flags" && <ResultPlanItems title={t("approval.review.finalConfigureFlags")} items={props.result.configureFlags} emptyText={t("result.details.empty.flags")} code />}
      </div>
    </section>
  );
}

function ResultEmptyCard(props: { variant: "workspace" | "build"; title: string; description: string; actionLabel: string; onAction: () => void }) {
  return (
    <section className={`card card--${props.variant === "workspace" ? "blue" : "purple"} result-empty-card`}>
      <span className="card__badge" aria-hidden="true">
        <img className="card__badge-icon" src={props.variant === "workspace" ? emptyStateBlueIcon : emptyStatePurpleIcon} alt="" />
      </span>
      <div className="card__head result-empty-card__head">
        <h2 className="card__title">{props.title}</h2>
        <p className="card__desc">{props.description}</p>
      </div>
      <div className="result-empty-card__actions">
        <button className="button button--primary" type="button" onClick={props.onAction}>{props.actionLabel}</button>
      </div>
    </section>
  );
}

function resultLibraryId(raw: string): string {
  const match = raw.match(/^library:([^:]+):/);
  return match ? match[1] : "";
}

// Pill text + variant for one library's verification status. Libraries are
// checked by their configure flag, by probing for a provided component, or are
// built-in components that ship in every FFmpeg build.
function libraryVerifyPill(status: LibraryVerification): { variant: string; label: string; title: string } {
  if (status.status === "builtin") {
    return { variant: "builtin", label: t("result.verify.status.builtin"), title: t("result.verify.builtinHint") };
  }
  const byComponent = status.method === "component";
  if (status.status === "ok") {
    return {
      variant: "ok",
      label: t("result.verify.status.ok"),
      title: byComponent ? t("result.verify.foundComponent", { components: status.components.join(", ") }) : status.expectedFlags.join(" "),
    };
  }
  return {
    variant: "missing",
    label: t("result.verify.status.missing"),
    title: byComponent ? t("result.verify.missingComponent", { components: status.components.join(", ") }) : t("result.verify.missingFlags", { flags: status.missingFlags.join(" ") }),
  };
}

// Libraries detail tab. When a verification has been run, each row gains a
// Present/Missing pill and a compact caption carries the global findings.
function ResultLibraryItems(props: { librarySpecs: string[]; verification: BuildVerification | null; verificationError: string; isVerifying: boolean }) {
  const verification = props.verification;
  const statusByLibraryId = React.useMemo(() => {
    const map: Record<string, LibraryVerification> = {};
    if (verification) for (const library of verification.libraries) map[library.libraryId] = library;
    return map;
  }, [verification]);
  const showsSummary = Boolean(verification) && verification!.overall !== "unverifiable";

  return (
    <section className="result-plan-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{t("result.review.libraries")}</h2>
        <span className="result-files__count">{t("result.details.count", { count: props.librarySpecs.length })}</span>
        {showsSummary && (
          <span className={`result-verify-badge result-verify-badge--${verification!.overall}`}>
            {t("result.verify.summary", { ok: verification!.okCount, total: verification!.totalCount })}
          </span>
        )}
      </div>
      {props.librarySpecs.length === 0 ? <p className="empty-text">{t("result.details.empty.libraries")}</p> : (
        <div className="result-file-list result-plan-list">
          {props.librarySpecs.map((spec, index) => {
            const status = statusByLibraryId[resultLibraryId(spec)];
            const pill = status ? libraryVerifyPill(status) : null;
            return (
              <article className={`result-plan-item ${pill ? "result-plan-item--with-status" : ""}`} key={`${spec}-${index}`}>
                <span className="result-plan-item__index">{String(index + 1).padStart(2, "0")}</span>
                <span className="result-plan-item__text">{resultLibraryLabel(spec)}</span>
                {pill && (
                  <span className={`result-verify-pill result-verify-pill--${pill.variant}`} title={pill.title}>{pill.label}</span>
                )}
              </article>
            );
          })}
        </div>
      )}
      {props.isVerifying && <p className="result-plan-panel__note">{t("result.verify.running")}</p>}
      {props.verificationError && <p className="result-plan-panel__note result-plan-panel__note--warn">{props.verificationError}</p>}
      {verification && verification.overall === "unverifiable" && <p className="result-plan-panel__note">{verification.message}</p>}
      {verification && verification.ffmpegVersion && <p className="result-plan-panel__note">{t("result.verify.ffmpegVersion", { version: verification.ffmpegVersion })}</p>}
      {verification && verification.unexpectedEnableFlags.length > 0 && (
        <p className="result-plan-panel__note result-plan-panel__note--warn">{t("result.verify.unexpected", { flags: verification.unexpectedEnableFlags.join(" ") })}</p>
      )}
    </section>
  );
}

function ResultPanel(props: { result: BuildResult | null; errorText: string; isLoading: boolean; verification: BuildVerification | null; verificationError: string; isVerifying: boolean; hasWorkspace: boolean; onOpenFolder: () => Promise<void>; onOpenReport: () => Promise<void>; onGoToSource: () => void; onGoToBuild: () => void }) {
  const result = props.result;
  const isMissingWorkspace = !props.hasWorkspace;
  const isLoadingFirstResult = props.isLoading && !result && !props.errorText && props.hasWorkspace;
  return (
    <section className="result-panel">
      {isMissingWorkspace && (
        <ResultEmptyCard
          variant="workspace"
          title={t("result.error.chooseWorkspaceFirst")}
          description={t("result.empty.workspace.description")}
          actionLabel={t("result.actions.goToSource")}
          onAction={props.onGoToSource}
        />
      )}
      {props.errorText && !isMissingWorkspace && <p className="empty-text">{props.errorText}</p>}
      {isLoadingFirstResult && <p className="empty-text">{t("result.loading")}</p>}
      {!props.errorText && props.hasWorkspace && !result && !props.isLoading && (
        <ResultEmptyCard
          variant="build"
          title={t("result.empty")}
          description={t("result.empty.build.description")}
          actionLabel={t("result.actions.goToBuild")}
          onAction={props.onGoToBuild}
        />
      )}
      {result && (
        <>
          <ResultSummary result={result} />
          <ResultLocationPanel result={result} onOpenFolder={props.onOpenFolder} onOpenReport={props.onOpenReport} />
          <ResultDetailTabs result={result} verification={props.verification} verificationError={props.verificationError} isVerifying={props.isVerifying} />
        </>
      )}
    </section>
  );
}

// ─── ResultTab ───────────────────────────────────────────────────────────────

export type ResultTabProps = {
  buildResult: BuildResult | null;
  buildResultError: string;
  isLoadingBuildResult: boolean;
  buildVerification: BuildVerification | null;
  buildVerificationError: string;
  isVerifyingBuild: boolean;
  verifyBuildResult: () => Promise<void>;
  hasWorkspace: boolean;
  refreshBuildResult: () => Promise<void>;
  openResultFolder: () => Promise<void>;
  openResultReport: () => Promise<void>;
  onGoToSource: () => void;
  onGoToBuild: () => void;
};

export function ResultTab({ buildResult, buildResultError, isLoadingBuildResult, buildVerification, buildVerificationError, isVerifyingBuild, verifyBuildResult, hasWorkspace, refreshBuildResult, openResultFolder, openResultReport, onGoToSource, onGoToBuild }: ResultTabProps) {
  return (
    <section className="tab-page result-page">
      <header className="result-page__header">
        <div>
          <h1 className="page-header__title">{t("result.title")}</h1>
          <p className="page-header__text">{t("result.intro")}</p>
        </div>
        <div className="result-page__actions">
          {buildResult && (
            <button className="button" type="button" onClick={verifyBuildResult} disabled={isVerifyingBuild}>
              {isVerifyingBuild ? t("result.verify.running") : t("result.verify.action")}
            </button>
          )}
          <button className="button" type="button" onClick={refreshBuildResult}>{t("result.actions.refresh")}</button>
        </div>
      </header>
      <ResultPanel
        result={buildResult}
        errorText={buildResultError}
        isLoading={isLoadingBuildResult}
        verification={buildVerification}
        verificationError={buildVerificationError}
        isVerifying={isVerifyingBuild}
        hasWorkspace={hasWorkspace}
        onOpenFolder={openResultFolder}
        onOpenReport={openResultReport}
        onGoToSource={onGoToSource}
        onGoToBuild={onGoToBuild}
      />
    </section>
  );
}
