import React from "react";
import { LLocaleGet, LLocaleMessageGet, LLocaleTextGet, type LLocalizedMessage } from "../i18n";
import { LOptionNameGet, LLicenseLabelGet } from "../catalogText";
import { LLicenseShortGet } from "./options";
import { ClipboardSetText } from "../../wailsjs/runtime/runtime";
import emptyStateBlueIcon from "../assets/empty-card-icons/EmptyStateBlue.svg";
import emptyStatePurpleIcon from "../assets/empty-card-icons/EmptyStatePurple.svg";
import copyIcon from "../assets/button-icons/Copy.svg";

function PButtonFileRender(props: { hash: string }) {
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
      title={copied ? LLocaleTextGet("result.files.copied") : LLocaleTextGet("result.files.copyHash")}
      aria-label={copied ? LLocaleTextGet("result.files.copied") : LLocaleTextGet("result.files.copyHash")}
    >
      <img className="result-file__copy-icon" src={copyIcon} alt="" aria-hidden="true" />
    </button>
  );
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

function LTextByteFormat(sizeBytes: number): string {
  if (sizeBytes < 1024) return `${sizeBytes} ${LLocaleTextGet("result.size.byte")}`;
  if (sizeBytes < 1024 * 1024) return `${(sizeBytes / 1024).toFixed(1)} ${LLocaleTextGet("result.size.kilobyte")}`;
  if (sizeBytes < 1024 * 1024 * 1024) return `${(sizeBytes / 1024 / 1024).toFixed(1)} ${LLocaleTextGet("result.size.megabyte")}`;
  return `${(sizeBytes / 1024 / 1024 / 1024).toFixed(2)} ${LLocaleTextGet("result.size.gigabyte")}`;
}

function LTextLibraryBuild(raw: string): string {
  const match = raw.match(/^library:([^:]+):(.+)$/);
  if (!match) return raw;
  const [, libraryId, licenseEffectName] = match;
  const name = LLocaleTextGet(`catalog.libraries.${libraryId}.displayName`);
  return `${name} (${LLicenseLabelGet(licenseEffectName)})`;
}

function LOptionResultGet(raw: string): string {
  const match = raw.match(/^option:(.+)$/);
  if (!match) return raw;
  return LOptionNameGet(match[1]);
}

function LFileKindGet(fileName: string): "exe" | "dll" | "other" {
  const lower = fileName.toLowerCase();
  if (lower.endsWith(".exe")) return "exe";
  if (lower.endsWith(".dll")) return "dll";
  return "other";
}

function LDateTimeFormat(value: string): string {
  if (!value) return LLocaleTextGet("result.summary.latestBuild.unknown");
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return value;
  return new Intl.DateTimeFormat(LLocaleGet() === "ko" ? "ko-KR" : "en-US", { dateStyle: "medium", timeStyle: "short" }).format(date);
}

function LBuildTypeGet(_result: LResultState): string {
  return LLocaleTextGet("result.summary.buildType.custom");
}

// ─── PPanelResultRender ─────────────────────────────────────────────────────────────

function PSummaryResultRender(props: { result: LResultState }) {
  const totalSizeBytes = props.result.files.reduce((sum, file) => sum + file.sizeBytes, 0);
  const stats: { label: string; value: string; unit: string; danger?: boolean }[] = [
    { label: LLocaleTextGet("result.summary.version"), value: props.result.ffmpegVersion || LLocaleTextGet("result.summary.latestBuild.unknown"), unit: "" },
    { label: LLocaleTextGet("result.summary.totalFiles"), value: String(props.result.files.length), unit: LLocaleTextGet("result.summary.filesUnit") },
    { label: LLocaleTextGet("result.summary.totalSize"), value: LTextByteFormat(totalSizeBytes), unit: "" },
    { label: LLocaleTextGet("result.summary.buildType"), value: LBuildTypeGet(props.result), unit: "" },
    { label: LLocaleTextGet("result.summary.license"), value: LLicenseShortGet(props.result.licenseProfileName), unit: "", danger: props.result.licenseProfileName === "nonfree-local" },
    { label: LLocaleTextGet("result.summary.latestBuild"), value: LDateTimeFormat(props.result.createdAt), unit: "" },
  ];
  return (
    <section className="result-summary" aria-label={LLocaleTextGet("result.summary.ariaLabel")}>
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

function PPanelLocationRender(props: { result: LResultState; onOpenFolder: () => Promise<void>; onOpenReport: () => Promise<void> }) {
  const rows = [
    { label: LLocaleTextGet("result.metadata.folder"), value: props.result.artifactsDirectory, actionLabel: LLocaleTextGet("result.actions.open") },
    { label: LLocaleTextGet("result.metadata.latestReport"), value: props.result.reportPath || LLocaleTextGet("result.metadata.noReport"), actionLabel: LLocaleTextGet("result.actions.open") },
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

function PListFileRender(props: { files: LFileResult[] }) {
  const [query, setQuery] = React.useState("");
  const [filterKind, setFilterKind] = React.useState<"all" | "exe" | "dll" | "other">("all");
  const [sortMode, setSortMode] = React.useState<"name" | "size-desc" | "size-asc">("name");

  const visibleFiles = React.useMemo(() => {
    const normalizedQuery = query.trim().toLowerCase();
    const filtered = props.files.filter((file) => {
      const matchesKind = filterKind === "all" || LFileKindGet(file.name) === filterKind;
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
        <h2 className="review-list__title result-files__title">{LLocaleTextGet("result.files.title")}</h2>
        <span className="result-files__count">{LLocaleTextGet("result.files.count", { count: visibleFiles.length, total: props.files.length })}</span>
      </div>
      {props.files.length > 0 && (
        <div className="result-files__controls">
          <input
            className="result-files__search"
            type="search"
            value={query}
            placeholder={LLocaleTextGet("result.files.searchPlaceholder")}
            aria-label={LLocaleTextGet("result.files.searchAriaLabel")}
            onChange={(event) => setQuery(event.currentTarget.value)}
          />
          <select className="result-files__select" value={filterKind} aria-label={LLocaleTextGet("result.files.filterAriaLabel")} onChange={(event) => setFilterKind(event.currentTarget.value as typeof filterKind)}>
            <option value="all">{LLocaleTextGet("result.files.filter.all")}</option>
            <option value="exe">{LLocaleTextGet("result.files.filter.exe")}</option>
            <option value="dll">{LLocaleTextGet("result.files.filter.dll")}</option>
            <option value="other">{LLocaleTextGet("result.files.filter.other")}</option>
          </select>
          <select className="result-files__select" value={sortMode} aria-label={LLocaleTextGet("result.files.sortAriaLabel")} onChange={(event) => setSortMode(event.currentTarget.value as typeof sortMode)}>
            <option value="name">{LLocaleTextGet("result.files.sort.name")}</option>
            <option value="size-desc">{LLocaleTextGet("result.files.sort.sizeDesc")}</option>
            <option value="size-asc">{LLocaleTextGet("result.files.sort.sizeAsc")}</option>
          </select>
        </div>
      )}
      {props.files.length === 0 ? <p className="empty-text">{LLocaleTextGet("result.files.empty")}</p> : (
        <div className="result-file-list">
          {visibleFiles.length === 0 ? <p className="empty-text result-file-list__empty">{LLocaleTextGet("result.files.noMatches")}</p> : visibleFiles.map((file) => (
            <article className="result-file" key={file.path}>
              <span className={`result-file__type result-file__type--${LFileKindGet(file.name)}`}>{LFileKindGet(file.name).toUpperCase()}</span>
              <strong className="result-file__name">{file.name}</strong>
              <span className="result-file__size">{LTextByteFormat(file.sizeBytes)}</span>
              <span className="result-file__path">{file.path}</span>
              {file.sha256Hash && <span className="result-file__hash">{LLocaleTextGet("result.files.sha256", { hash: file.sha256Hash })}</span>}
              {file.sha256Hash && <PButtonFileRender hash={file.sha256Hash} />}
            </article>
          ))}
        </div>
      )}
    </section>
  );
}


type LResultDetailIdentifier = "files" | "libraries" | "packages" | "options" | "flags";

function PResultItemsRender(props: { title: string; items: string[]; emptyText: string; code?: boolean }) {
  return (
    <section className="result-plan-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{props.title}</h2>
        <span className="result-files__count">{LLocaleTextGet("result.details.count", { count: props.items.length })}</span>
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

function PTabDetailRender(props: { result: LResultState; verification: LVerificationState | null; verificationError: LLocalizedMessage | null; isVerifying: boolean }) {
  const [activeTabId, setActiveTabId] = React.useState<LResultDetailIdentifier>("files");
  const locale = LLocaleGet();
  const options = React.useMemo(() => props.result.selectedConfigureOptions.map(LOptionResultGet), [props.result.selectedConfigureOptions, locale]);

  // Jump to the Libraries tab when a verification run finishes or starts, so its
  // results surface where the user is looking instead of on a hidden tab.
  React.useEffect(() => {
    if (props.isVerifying || props.verification || props.verificationError) setActiveTabId("libraries");
  }, [props.isVerifying, props.verification, props.verificationError]);
  const tabs: { id: LResultDetailIdentifier; label: string; count: number }[] = [
    { id: "files", label: LLocaleTextGet("result.details.tabs.files"), count: props.result.files.length },
    { id: "libraries", label: LLocaleTextGet("result.details.tabs.libraries"), count: props.result.selectedLibraries.length },
    { id: "packages", label: LLocaleTextGet("result.details.tabs.packages"), count: props.result.requiredMsys2PackageNames.length },
    { id: "options", label: LLocaleTextGet("result.details.tabs.options"), count: options.length },
    { id: "flags", label: LLocaleTextGet("result.details.tabs.flags"), count: props.result.configureFlags.length },
  ];

  return (
    <section className="result-details-card">
      <div className="result-details-tabs" role="tablist" aria-label={LLocaleTextGet("result.details.tabs.ariaLabel")}>
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
        {activeTabId === "files" && <PListFileRender files={props.result.files} />}
        {activeTabId === "libraries" && <PResultLibraryRender librarySpecs={props.result.selectedLibraries} verification={props.verification} verificationError={props.verificationError} isVerifying={props.isVerifying} />}
        {activeTabId === "packages" && <PResultItemsRender title={LLocaleTextGet("result.review.packages")} items={props.result.requiredMsys2PackageNames} emptyText={LLocaleTextGet("result.details.empty.packages")} code />}
        {activeTabId === "options" && <PResultItemsRender title={LLocaleTextGet("result.review.options")} items={options} emptyText={LLocaleTextGet("result.details.empty.options")} />}
        {activeTabId === "flags" && <PResultItemsRender title={LLocaleTextGet("approval.review.finalConfigureFlags")} items={props.result.configureFlags} emptyText={LLocaleTextGet("result.details.empty.flags")} code />}
      </div>
    </section>
  );
}

function PCardEmptyRender(props: { variant: "workspace" | "build"; title: string; description: string; actionLabel: string; onAction: () => void }) {
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

function LLibraryResultGet(raw: string): string {
  const match = raw.match(/^library:([^:]+):/);
  return match ? match[1] : "";
}

// Pill text + variant for one library's verification status. Libraries are
// checked by their configure flag, by probing for a provided component, or are
// built-in components that ship in every FFmpeg build.
function LLibraryPillGet(status: LVerificationLibrary): { variant: string; label: string; title: string } {
  if (status.status === "builtin") {
    return { variant: "builtin", label: LLocaleTextGet("result.verify.status.builtin"), title: LLocaleTextGet("result.verify.builtinHint") };
  }
  const byComponent = status.method === "component";
  if (status.status === "ok") {
    return {
      variant: "ok",
      label: LLocaleTextGet("result.verify.status.ok"),
      title: byComponent ? LLocaleTextGet("result.verify.foundComponent", { components: status.components.join(", ") }) : status.expectedFlags.join(" "),
    };
  }
  return {
    variant: "missing",
    label: LLocaleTextGet("result.verify.status.missing"),
    title: byComponent ? LLocaleTextGet("result.verify.missingComponent", { components: status.components.join(", ") }) : LLocaleTextGet("result.verify.missingFlags", { flags: status.missingFlags.join(" ") }),
  };
}

// Libraries detail tab. When a verification has been run, each row gains a
// Present/Missing pill and a compact caption carries the global findings.
function PResultLibraryRender(props: { librarySpecs: string[]; verification: LVerificationState | null; verificationError: LLocalizedMessage | null; isVerifying: boolean }) {
  const verification = props.verification;
  const statusByLibraryId = React.useMemo(() => {
    const map: Record<string, LVerificationLibrary> = {};
    if (verification) for (const library of verification.libraries) map[library.libraryId] = library;
    return map;
  }, [verification]);
  const showsSummary = Boolean(verification) && verification!.overall !== "unverifiable";

  return (
    <section className="result-plan-panel">
      <div className="result-files__head">
        <h2 className="review-list__title result-files__title">{LLocaleTextGet("result.review.libraries")}</h2>
        <span className="result-files__count">{LLocaleTextGet("result.details.count", { count: props.librarySpecs.length })}</span>
        {showsSummary && (
          <span className={`result-verify-badge result-verify-badge--${verification!.overall}`}>
            {LLocaleTextGet("result.verify.summary", { ok: verification!.okCount, total: verification!.totalCount })}
          </span>
        )}
      </div>
      {props.librarySpecs.length === 0 ? <p className="empty-text">{LLocaleTextGet("result.details.empty.libraries")}</p> : (
        <div className="result-file-list result-plan-list">
          {props.librarySpecs.map((spec, index) => {
            const status = statusByLibraryId[LLibraryResultGet(spec)];
            const pill = status ? LLibraryPillGet(status) : null;
            return (
              <article className={`result-plan-item ${pill ? "result-plan-item--with-status" : ""}`} key={`${spec}-${index}`}>
                <span className="result-plan-item__index">{String(index + 1).padStart(2, "0")}</span>
                <span className="result-plan-item__text">{LTextLibraryBuild(spec)}</span>
                {pill && (
                  <span className={`result-verify-pill result-verify-pill--${pill.variant}`} title={pill.title}>{pill.label}</span>
                )}
              </article>
            );
          })}
        </div>
      )}
      {props.isVerifying && <p className="result-plan-panel__note">{LLocaleTextGet("result.verify.running")}</p>}
      {props.verificationError && <p className="result-plan-panel__note result-plan-panel__note--warn">{LLocaleMessageGet(props.verificationError)}</p>}
      {verification && verification.overall === "unverifiable" && <p className="result-plan-panel__note">{LLocaleMessageGet(verification)}</p>}
      {verification && verification.ffmpegVersion && <p className="result-plan-panel__note">{LLocaleTextGet("result.verify.ffmpegVersion", { version: verification.ffmpegVersion })}</p>}
      {verification && verification.unexpectedEnableFlags.length > 0 && (
        <p className="result-plan-panel__note result-plan-panel__note--warn">{LLocaleTextGet("result.verify.unexpected", { flags: verification.unexpectedEnableFlags.join(" ") })}</p>
      )}
    </section>
  );
}

function PPanelResultRender(props: { result: LResultState | null; errorText: string; isLoading: boolean; verification: LVerificationState | null; verificationError: LLocalizedMessage | null; isVerifying: boolean; hasWorkspace: boolean; onOpenFolder: () => Promise<void>; onOpenReport: () => Promise<void>; onGoToSource: () => void; onGoToBuild: () => void }) {
  const result = props.result;
  const isMissingWorkspace = !props.hasWorkspace;
  const isLoadingFirstResult = props.isLoading && !result && !props.errorText && props.hasWorkspace;
  return (
    <section className="result-panel">
      {isMissingWorkspace && (
        <PCardEmptyRender
          variant="workspace"
          title={LLocaleTextGet("result.error.chooseWorkspaceFirst")}
          description={LLocaleTextGet("result.empty.workspace.description")}
          actionLabel={LLocaleTextGet("result.actions.goToSource")}
          onAction={props.onGoToSource}
        />
      )}
      {props.errorText && !isMissingWorkspace && <p className="empty-text">{props.errorText}</p>}
      {isLoadingFirstResult && <p className="empty-text">{LLocaleTextGet("result.loading")}</p>}
      {!props.errorText && props.hasWorkspace && !result && !props.isLoading && (
        <PCardEmptyRender
          variant="build"
          title={LLocaleTextGet("result.empty")}
          description={LLocaleTextGet("result.empty.build.description")}
          actionLabel={LLocaleTextGet("result.actions.goToBuild")}
          onAction={props.onGoToBuild}
        />
      )}
      {result && (
        <>
          <PSummaryResultRender result={result} />
          <PPanelLocationRender result={result} onOpenFolder={props.onOpenFolder} onOpenReport={props.onOpenReport} />
          <PTabDetailRender result={result} verification={props.verification} verificationError={props.verificationError} isVerifying={props.isVerifying} />
        </>
      )}
    </section>
  );
}

// ─── PResultRender ───────────────────────────────────────────────────────────────

export type LResultProperties = {
  buildResult: LResultState | null;
  buildResultError: string;
  isLoadingBuildResult: boolean;
  buildVerification: LVerificationState | null;
  buildVerificationError: LLocalizedMessage | null;
  isVerifyingBuild: boolean;
  verifyBuildResult: () => Promise<void>;
  hasWorkspace: boolean;
  refreshBuildResult: () => Promise<void>;
  openResultFolder: () => Promise<void>;
  openResultReport: () => Promise<void>;
  onGoToSource: () => void;
  onGoToBuild: () => void;
};

export function PResultRender({ buildResult, buildResultError, isLoadingBuildResult, buildVerification, buildVerificationError, isVerifyingBuild, verifyBuildResult, hasWorkspace, refreshBuildResult, openResultFolder, openResultReport, onGoToSource, onGoToBuild }: LResultProperties) {
  return (
    <section className="tab-page result-page">
      <header className="result-page__header">
        <div>
          <h1 className="page-header__title">{LLocaleTextGet("result.title")}</h1>
          <p className="page-header__text">{LLocaleTextGet("result.intro")}</p>
        </div>
        <div className="result-page__actions">
          {buildResult && (
            <button className="button" type="button" onClick={verifyBuildResult} disabled={isVerifyingBuild}>
              {isVerifyingBuild ? LLocaleTextGet("result.verify.running") : LLocaleTextGet("result.verify.action")}
            </button>
          )}
          <button className="button" type="button" onClick={refreshBuildResult}>{LLocaleTextGet("result.actions.refresh")}</button>
        </div>
      </header>
      <PPanelResultRender
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
