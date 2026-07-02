import React from "react";
import { LLocaleTextGet } from "../i18n";
import { LOptionTextGet } from "../catalogText";
import { PHeaderPageRender } from "./shared";
import optionPresetCardIcon from "../assets/option-card-icons/OptionPreset.svg";
import optionsCardIcon from "../assets/option-card-icons/Options.svg";
import optionNoneIcon from "../assets/option-preset-icons/PresetNone.svg";
import optionStandardIcon from "../assets/option-preset-icons/PresetStandard.svg";
import optionCompactIcon from "../assets/option-preset-icons/PresetCompact.svg";
import optionPortableIcon from "../assets/option-preset-icons/PresetPortable.svg";
import optionPerformanceIcon from "../assets/option-preset-icons/PresetPerformance.svg";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";
import resetIcon from "../assets/button-icons/Reset.svg";

// ─── Preset data ───────────────────────────────────────────────────────────────

export type LPresetOptionId =
  | "none"
  | "standard"
  | "compact"
  | "portable"
  | "performance"
  | "custom";

export type LPresetOption = {
  presetId: LPresetOptionId;
  optionIds: string[];
};

// Locked program defaults are always present; presets layer build-tuning toggles
// on top. High-risk and troubleshooting toggles (enable-shared, disable-asm,
// disable-x86asm, disable-network, etc.) are intentionally in no preset and stay
// reachable only by hand, which keeps every preset safe.
const LOptionBaseIds = [
  "default-static",
  "default-programs",
  "default-ffmpeg",
  "default-ffprobe",
];

export const LPresetOptionList: LPresetOption[] = [
  { presetId: "none", optionIds: LOptionBaseIds },
  {
    presetId: "standard",
    optionIds: [...LOptionBaseIds, "pkg-config-static", "disable-doc"],
  },
  {
    presetId: "compact",
    optionIds: [
      ...LOptionBaseIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
    ],
  },
  {
    presetId: "portable",
    optionIds: [
      ...LOptionBaseIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
      "enable-runtime-cpudetect",
    ],
  },
  {
    presetId: "performance",
    optionIds: [
      ...LOptionBaseIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
      "enable-lto",
    ],
  },
];

export function LPresetOptionMatch(
  selectedOptionIds: string[],
): LPresetOptionId {
  const normalizedSelection = Array.from(new Set(selectedOptionIds)).sort();
  for (const preset of LPresetOptionList) {
    const normalizedPreset = Array.from(new Set(preset.optionIds)).sort();
    if (
      normalizedSelection.length === normalizedPreset.length &&
      normalizedSelection.every(
        (optionId, index) => optionId === normalizedPreset[index],
      )
    ) {
      return preset.presetId;
    }
  }
  return "custom";
}

const LOptionPresetIconMap: Partial<Record<LPresetOptionId, string>> = {
  none: optionNoneIcon,
  standard: optionStandardIcon,
  compact: optionCompactIcon,
  portable: optionPortableIcon,
  performance: optionPerformanceIcon,
};

function PIconOptionRender(props: { presetId: LPresetOptionId; className?: string }) {
  const icon = LOptionPresetIconMap[props.presetId];
  if (!icon) return null;
  return <img className={props.className ?? ""} src={icon} alt="" aria-hidden="true" />;
}

function LPresetOptionSelectableList(): Array<LPresetOption & { presetId: Exclude<LPresetOptionId, "custom"> }> {
  return LPresetOptionList.filter((preset): preset is LPresetOption & { presetId: Exclude<LPresetOptionId, "custom"> } => preset.presetId !== "custom");
}

function PSelectorOptionRender(props: {
  selectedPresetId: LPresetOptionId;
  onApplyPreset: (presetId: LPresetOptionId) => void;
}) {
  return (
    <section className="preset-panel options-preset-panel">
      <div className="preset-grid options-preset-grid">
        {LPresetOptionSelectableList().map((preset) => {
          const active = props.selectedPresetId === preset.presetId;
          return (
            <button
              className={`preset-card options-preset-card ${active ? "preset-card--active" : ""}`}
              type="button"
              key={preset.presetId}
              onClick={() => props.onApplyPreset(preset.presetId)}
            >
              <PIconOptionRender presetId={preset.presetId} className="preset-card__icon" />
              <span>
                <span className="preset-card__name">
                  {LLocaleTextGet(`options.presets.${preset.presetId}.name`)}
                </span>
                <span className="preset-card__plain">
                  {LLocaleTextGet(`options.presets.${preset.presetId}.plain`)}
                </span>
              </span>
              {active && (
                <span className="preset-card__check" aria-hidden="true">
                  ✓
                </span>
              )}
            </button>
          );
        })}
      </div>
      {props.selectedPresetId === "custom" && (
        <p className="preset-panel__custom">
          <span className="preset-panel__custom-icon" aria-hidden="true">
            ✓
          </span>
          {LLocaleTextGet("options.presetSelector.custom")}
        </p>
      )}
    </section>
  );
}

function PCardOptionPresetRender(props: {
  selectedPresetId: LPresetOptionId;
  onApplyPreset: (presetId: LPresetOptionId) => void;
}) {
  const presets = LPresetOptionSelectableList();
  const selectedPreset = presets.find((preset) => preset.presetId === props.selectedPresetId);
  const selectedPresetDescription = selectedPreset
    ? LLocaleTextGet(`options.presets.${selectedPreset.presetId}.plain`)
    : LLocaleTextGet("options.presetSelector.custom");

  return (
    <section className="card card--blue options-simple-card options-simple-preset-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={optionPresetCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("options.simple.preset.title")}</h2>
      </div>
      <div className="card__control options-simple-card__control">
        <select
          className="card__input"
          value={props.selectedPresetId}
          onChange={(event) => props.onApplyPreset(event.target.value as LPresetOptionId)}
        >
          {props.selectedPresetId === "custom" && <option value="custom">{LLocaleTextGet("options.presets.custom.name")}</option>}
          {presets.map((preset) => (
            <option value={preset.presetId} key={preset.presetId}>
              {LLocaleTextGet(`options.presets.${preset.presetId}.name`)}
            </option>
          ))}
        </select>
      </div>
      <div className="options-simple-preset-card__description" aria-live="polite">
        {selectedPresetDescription}
      </div>
    </section>
  );
}

// Renders an option's plain text: "\n" becomes a line break, "**text**" becomes
// bold, and "<red>text</red>" becomes a red-colored span (used for the "Risk"
// keyword on high-risk options). No other markup is interpreted.
function PTextOptionRender(text: string): React.ReactNode {
  return text.split("\n").map((line, lineIndex) => (
    <React.Fragment key={lineIndex}>
      {lineIndex > 0 && <br />}
      {line
        .split(/(\*\*[^*]+\*\*)/g)
        .map((segment, segmentIndex) =>
          segment.startsWith("**") && segment.endsWith("**") ? (
            <strong key={segmentIndex}>
              {PTextEmphasisRender(segment.slice(2, -2))}
            </strong>
          ) : (
            <React.Fragment key={segmentIndex}>
              {PTextEmphasisRender(segment)}
            </React.Fragment>
          ),
        )}
    </React.Fragment>
  ));
}

function PTextEmphasisRender(text: string): React.ReactNode {
  return text.split(/(<red>[^<]+<\/red>)/g).map((segment, index) =>
    segment.startsWith("<red>") && segment.endsWith("</red>") ? (
      <span key={index} className="option-row__risk-word">
        {segment.slice(5, -6)}
      </span>
    ) : (
      segment
    ),
  );
}

function LOptionRiskResolve(riskLevelName: string): string {
  return LLocaleTextGet(`options.row.risk.${LOptionRiskNormalize(riskLevelName)}`);
}

function LOptionRiskNormalize(riskLevelName: string): string {
  return riskLevelName === "high" || riskLevelName === "medium"
    ? riskLevelName
    : "low";
}

function LOptionCategoryGroup(LCatalogLibrarySource: LOptionChoice[]) {
  return LCatalogLibrarySource.reduce<Record<string, LOptionChoice[]>>(
    (groups, option) => {
      const categoryName =
        LOptionTextGet(option, "categoryName") || LLocaleTextGet("common.other");
      groups[categoryName] = groups[categoryName] || [];
      groups[categoryName].push(option);
      return groups;
    },
    {},
  );
}

export function LLicenseBoundaryResolve(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return LLocaleTextGet("options.licenseBoundary.gpl");
    case "nonfree-local":
      return LLocaleTextGet("options.licenseBoundary.nonfree");
    case "lgpl-local":
    default:
      return LLocaleTextGet("options.licenseBoundary.lgpl");
  }
}

export function LLicenseBoundaryLabelShortGet(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return LLocaleTextGet("options.summary.license.gpl-local");
    case "nonfree-local":
      return LLocaleTextGet("options.summary.license.nonfree-local");
    case "lgpl-local":
    default:
      return LLocaleTextGet("options.summary.license.lgpl-local");
  }
}

function LLicenseBoundaryNormalize(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
    case "nonfree-local":
      return licenseProfileName;
    default:
      return "lgpl-local";
  }
}

// ─── Option summary ───────────────────────────────────────────────────────────

function PSummaryOptionRender(props: {
  licenseProfileName: string;
  selectedOptionCount: number;
  selectedPresetId: LPresetOptionId;
}) {
  const licenseBoundary = LLicenseBoundaryNormalize(
    props.licenseProfileName,
  );
  return (
    <section
      className="option-summary-card"
      aria-label={LLocaleTextGet("options.summary.ariaLabel")}
    >
      <div className="option-summary-card__header">
        <span className="option-summary-card__status" aria-hidden="true">
          ✓
        </span>
        <div className="option-summary-card__copy">
          <span className="option-summary-card__title">
            {LLocaleTextGet("options.summary.currentSelection")}
          </span>
          <span className="option-summary-card__message">
            {LLocaleTextGet(`options.presets.${props.selectedPresetId}.name`)}
          </span>
        </div>
      </div>
      <div className="option-summary">
        <div className="option-summary__item">
          <span className="option-summary__label">
            {LLocaleTextGet("options.summary.license")}
          </span>
          <strong
            className={`option-summary__value option-summary__license option-summary__license--${licenseBoundary}`}
            title={LLicenseBoundaryResolve(props.licenseProfileName)}
          >
            {LLicenseBoundaryLabelShortGet(props.licenseProfileName)}
          </strong>
        </div>
        <div className="option-summary__item">
          <span className="option-summary__label">
            {LLocaleTextGet("options.summary.selected")}
          </span>
          <strong className="option-summary__value">
            {props.selectedOptionCount}
          </strong>
        </div>
      </div>
    </section>
  );
}

// ─── Technical panel ──────────────────────────────────────────────────────────

function PPanelTechnicalRender() {
  return (
    <section className="option-technical-panel">
      <h2 className="option-technical-panel__title">
        {LLocaleTextGet("options.technical.title")}
      </h2>
      <div className="option-technical-details">
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {LLocaleTextGet("options.technical.license.title")}
          </h3>
          <p className="option-technical-detail__text">
            {LLocaleTextGet("options.license.hint")}
          </p>
          <p className="option-technical-detail__text">
            {LLocaleTextGet("options.license.rule.lgpl")}{" "}
            <strong>{LLocaleTextGet("libraries.summary.license.lgpl-local")}</strong>.<br />
            {LLocaleTextGet("options.license.rule.gpl")}{" "}
            <strong>{LLocaleTextGet("libraries.summary.license.gpl-local")}</strong>.<br />
            {LLocaleTextGet("options.license.rule.nonfree.prefix")}{" "}
            <code>--enable-nonfree</code>
            {LLocaleTextGet("options.license.rule.nonfree.suffix")}{" "}
            <strong>{LLocaleTextGet("libraries.summary.license.nonfree-local")}</strong>.
          </p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {LLocaleTextGet("options.technical.configure.title")}
          </h3>
          <p className="option-technical-detail__text">
            {LLocaleTextGet("options.technical.configure.text")}
          </p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {LLocaleTextGet("options.technical.advanced.title")}
          </h3>
          <p className="option-technical-detail__text">
            {LLocaleTextGet("options.technical.advanced.text")}
          </p>
        </section>
      </div>
    </section>
  );
}

function LOptionSearchMatch(
  option: LOptionChoice,
  query: string,
): boolean {
  const normalizedQuery = query.trim().toLowerCase();
  if (!normalizedQuery) {
    return true;
  }
  return [
    LOptionTextGet(option, "displayName"),
    LOptionTextGet(option, "plainExplanation"),
    LOptionTextGet(option, "technicalNote"),
    LOptionTextGet(option, "categoryName"),
    ...option.configureFlags,
  ].some((value) => value.toLowerCase().includes(normalizedQuery));
}

function LOptionCategoryNamesGet(
  LCatalogLibrarySource: LOptionChoice[],
): string[] {
  return Array.from(
    new Set(
      LCatalogLibrarySource.map(
        (option) =>
          LOptionTextGet(option, "categoryName") || LLocaleTextGet("common.other"),
      ),
    ),
  );
}

function PDropdownCategoryRender(props: {
  categories: string[];
  selectedCategoryName: string;
  open: boolean;
  onToggleOpen: () => void;
  onSelectCategory: (categoryName: string) => void;
}) {
  const buttonClass = `options-category-dropdown__button ${props.open ? "options-category-dropdown__button--open" : ""}`;
  return (
    <div className="options-category-dropdown">
      <button
        className={buttonClass}
        type="button"
        aria-haspopup="listbox"
        aria-expanded={props.open}
        onClick={props.onToggleOpen}
      >
        <span className="options-category-dropdown__value">
          {props.selectedCategoryName || LLocaleTextGet("options.category.all")}
        </span>
        <span
          className="options-category-dropdown__chevron"
          aria-hidden="true"
        />
      </button>
      {props.open && (
        <div className="options-category-dropdown__menu" role="listbox">
          <button
            className={`options-category-dropdown__option ${!props.selectedCategoryName ? "options-category-dropdown__option--active" : ""}`}
            type="button"
            role="option"
            aria-selected={!props.selectedCategoryName}
            onClick={() => props.onSelectCategory("")}
          >
            <span>{LLocaleTextGet("options.category.all")}</span>
            {!props.selectedCategoryName && (
              <span
                className="options-category-dropdown__check"
                aria-hidden="true"
              >
                ✓
              </span>
            )}
          </button>
          {props.categories.map((categoryName) => {
            const active = props.selectedCategoryName === categoryName;
            return (
              <button
                className={`options-category-dropdown__option ${active ? "options-category-dropdown__option--active" : ""}`}
                type="button"
                role="option"
                aria-selected={active}
                key={categoryName}
                onClick={() => props.onSelectCategory(categoryName)}
              >
                <span>{categoryName}</span>
                {active && (
                  <span
                    className="options-category-dropdown__check"
                    aria-hidden="true"
                  >
                    ✓
                  </span>
                )}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

// ─── Option list ──────────────────────────────────────────────────────────────

function PListOptionRender(props: {
  LCatalogLibrarySource: LOptionChoice[];
  selectedOptionIds: string[];
  onToggleOption: (optionId: string) => void;
  showTechnicalDetails: boolean;
  searchQuery: string;
  selectedCategoryName: string;
}) {
  const filteredOptions = props.LCatalogLibrarySource.filter(
    (option) =>
      (!props.selectedCategoryName ||
        (LOptionTextGet(option, "categoryName") || LLocaleTextGet("common.other")) ===
          props.selectedCategoryName) &&
      LOptionSearchMatch(option, props.searchQuery),
  );
  const groupedOptions = LOptionCategoryGroup(filteredOptions);
  return (
    <div className="option-list">
      {Object.keys(groupedOptions).length === 0 && (
        <section className="option-empty">{LLocaleTextGet("options.empty")}</section>
      )}
      {Object.entries(groupedOptions).map(([categoryName, options]) => (
        <section className="option-group" key={categoryName}>
          <h2 className="option-group__title">{categoryName}</h2>
          {options.map((option) => (
            <label className="option-row" key={option.optionId}>
              <input
                type="checkbox"
                checked={props.selectedOptionIds.includes(option.optionId)}
                disabled={option.locked}
                onChange={() => props.onToggleOption(option.optionId)}
              />
              <span className="option-row__main">
                <span className="option-row__name">
                  {LOptionTextGet(option, "displayName")}
                </span>
                <span className="option-row__plain">
                  {PTextOptionRender(
                    LOptionTextGet(option, "plainExplanation"),
                  )}
                </span>
                {props.showTechnicalDetails &&
                  LOptionTextGet(option, "technicalNote") && (
                    <span className="option-row__detail">
                      {LOptionTextGet(option, "technicalNote")}
                    </span>
                  )}
                {props.showTechnicalDetails && (
                  <span className="option-row__detail option-row__detail--flags">
                    {option.configureFlags.length > 0
                      ? LLocaleTextGet("options.configure.flags", {
                          flags: option.configureFlags.join(" "),
                        })
                      : LLocaleTextGet("options.configure.defaultBehavior")}
                  </span>
                )}
              </span>
              <span
                className={`option-row__risk option-row__risk--${LOptionRiskNormalize(option.riskLevelName)}`}
                title={LLocaleTextGet("options.row.riskLabel")}
              >
                {LOptionRiskResolve(option.riskLevelName)}
              </span>
            </label>
          ))}
        </section>
      ))}
    </div>
  );
}

function PSectionFlagRender(props: {
  extraConfigureFlagText: string;
  onExtraFlagTextChange: (text: string) => void;
}) {
  return (
    <section className="options-section options-section--advanced">
      <h2 className="options-section__title">
        {LLocaleTextGet("options.advancedFlags.label")}
      </h2>
      <label className="field options-field options-section__body">
        <span className="field__hint">
          {LLocaleTextGet("options.advancedFlags.hint.prefix")} <code>./configure</code>
          {LLocaleTextGet("options.advancedFlags.hint.suffix")}
        </span>
        <textarea
          className="field__textarea"
          rows={5}
          value={props.extraConfigureFlagText}
          onChange={(event) => props.onExtraFlagTextChange(event.target.value)}
          placeholder={LLocaleTextGet("options.advancedFlags.placeholder")}
        />
      </label>
    </section>
  );
}

function PSectionThreadRender(props: {
  parallelJobCount: number;
  updateFfmpegBuildSettings: (partial: Partial<LSettingsFFmpeg>) => void;
}) {
  return (
    <section className="options-section">
      <h2 className="options-section__title">{LLocaleTextGet("options.jobs.label")}</h2>
      <label className="field options-field options-section__body">
        <span className="field__hint">{LLocaleTextGet("options.jobs.hint")}</span>
        <input
          className="field__input"
          type="number"
          min="1"
          max="256"
          value={props.parallelJobCount}
          onChange={(event) =>
            props.updateFfmpegBuildSettings({
              parallelJobCount: Number(event.target.value),
            })
          }
        />
      </label>
    </section>
  );
}

function PCardOptionRender(props: {
  LCatalogLibrarySource: LOptionChoice[];
  selectedOptionIds: string[];
  onToggleOption: (optionId: string) => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  selectedCategoryName: string;
  onSelectCategory: (categoryName: string) => void;
  categoryDropdownOpen: boolean;
  onToggleCategoryDropdown: () => void;
  LOptionCategoryNameGets: string[];
}) {
  return (
    <section className="card card--teal options-simple-card options-simple-options-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={optionsCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("options.simple.options.title")}</h2>
      </div>
      <div className="options-simple-options-card__body">
        <div className="options-simple-options-card__controls">
          <label className="options-search">
            <span className="visually-hidden">{LLocaleTextGet("options.search.label")}</span>
            <input
              type="search"
              value={props.searchQuery}
              onChange={(event) => props.onSearchQueryChange(event.target.value)}
              placeholder={LLocaleTextGet("options.search.placeholder")}
            />
          </label>
          <PDropdownCategoryRender
            categories={props.LOptionCategoryNameGets}
            selectedCategoryName={props.selectedCategoryName}
            open={props.categoryDropdownOpen}
            onToggleOpen={props.onToggleCategoryDropdown}
            onSelectCategory={props.onSelectCategory}
          />
        </div>
        <div className="options-simple-options-card__results">
          <PListOptionRender
            LCatalogLibrarySource={props.LCatalogLibrarySource}
            selectedOptionIds={props.selectedOptionIds}
            onToggleOption={props.onToggleOption}
            showTechnicalDetails={false}
            searchQuery={props.searchQuery}
            selectedCategoryName={props.selectedCategoryName}
          />
        </div>
      </div>
    </section>
  );
}

// ─── POptionRender ───────────────────────────────────────────────────────────────

export type POptionProps = {
  ffmpegBuildSettings: LSettingsFFmpeg;
  initialProgramState: LStateInitial;
  extraConfigureFlagText: string;
  onExtraFlagTextChange: (text: string) => void;
  updateFfmpegBuildSettings: (partial: Partial<LSettingsFFmpeg>) => void;
  toggleConfigureOption: (optionId: string) => void;
  applyOptionPreset: (presetId: LPresetOptionId) => void;
  optionsDetailedView: boolean;
  setOptionsDetailedView: (value: boolean) => void;
  showTechnicalDetails: boolean;
  setShowTechnicalDetails: (value: boolean) => void;
};

export function POptionRender({
  ffmpegBuildSettings,
  initialProgramState,
  extraConfigureFlagText,
  onExtraFlagTextChange,
  updateFfmpegBuildSettings,
  toggleConfigureOption,
  applyOptionPreset,
  optionsDetailedView,
  setOptionsDetailedView,
  showTechnicalDetails,
  setShowTechnicalDetails,
}: POptionProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [selectedCategoryName, setSelectedCategoryName] = React.useState("");
  const [categoryDropdownOpen, setCategoryDropdownOpen] = React.useState(false);
  const selectedPresetId = LPresetOptionMatch(
    ffmpegBuildSettings.selectedConfigureOptionIds,
  );
  const LOptionCategoryNameGets = React.useMemo(
    () =>
      LOptionCategoryNamesGet(
        initialProgramState.defaultConfigureOptionCatalog,
      ),
    [initialProgramState.defaultConfigureOptionCatalog],
  );

  return (
    <section className="tab-page options-page">
      <div className="options-page__header-row">
        <PHeaderPageRender title={LLocaleTextGet("options.title")} text={LLocaleTextGet("options.intro")} />
        <label className="options-design-toggle">
          <input
            type="checkbox"
            checked={optionsDetailedView}
            onChange={(event) => setOptionsDetailedView(event.target.checked)}
          />
          <span className="options-design-toggle__text">
            {LLocaleTextGet("options.designToggle.label")}
          </span>
        </label>
      </div>

      {!optionsDetailedView && (
        <div className="options-simple-layout">
          <PCardOptionPresetRender
            selectedPresetId={selectedPresetId}
            onApplyPreset={applyOptionPreset}
          />
          <PCardOptionRender
            LCatalogLibrarySource={initialProgramState.defaultConfigureOptionCatalog}
            selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds}
            onToggleOption={toggleConfigureOption}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            selectedCategoryName={selectedCategoryName}
            onSelectCategory={(categoryName) => {
              setSelectedCategoryName(categoryName);
              setCategoryDropdownOpen(false);
            }}
            categoryDropdownOpen={categoryDropdownOpen}
            onToggleCategoryDropdown={() => setCategoryDropdownOpen((value) => !value)}
            LOptionCategoryNameGets={LOptionCategoryNameGets}
          />
        </div>
      )}

      {optionsDetailedView && (
        <>
          <PSelectorOptionRender
            selectedPresetId={selectedPresetId}
            onApplyPreset={applyOptionPreset}
          />

          <PSummaryOptionRender
            licenseProfileName={ffmpegBuildSettings.licenseProfileName}
            selectedOptionCount={
              ffmpegBuildSettings.selectedConfigureOptionIds.length
            }
            selectedPresetId={selectedPresetId}
          />

          <div className="options-toolbar">
            <button
              className="button options-technical-toggle"
              type="button"
              aria-expanded={showTechnicalDetails}
              onClick={() => setShowTechnicalDetails(!showTechnicalDetails)}
            >
              <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
              {showTechnicalDetails
                ? LLocaleTextGet("options.technical.hide")
                : LLocaleTextGet("options.technical.show")}
            </button>
            <label className="options-search">
              <span className="visually-hidden">{LLocaleTextGet("options.search.label")}</span>
              <input
                type="search"
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder={LLocaleTextGet("options.search.placeholder")}
              />
            </label>
            <PDropdownCategoryRender
              categories={LOptionCategoryNameGets}
              selectedCategoryName={selectedCategoryName}
              open={categoryDropdownOpen}
              onToggleOpen={() => setCategoryDropdownOpen((value) => !value)}
              onSelectCategory={(categoryName) => {
                setSelectedCategoryName(categoryName);
                setCategoryDropdownOpen(false);
              }}
            />
            <button
              className="button options-reset"
              type="button"
              onClick={() => {
                setSearchQuery("");
                setSelectedCategoryName("");
                setCategoryDropdownOpen(false);
              }}
            >
              <img className="card__btn-icon" src={resetIcon} alt="" aria-hidden="true" />
              {LLocaleTextGet("options.filter.reset")}
            </button>
          </div>

          {showTechnicalDetails && <PPanelTechnicalRender />}

          <PListOptionRender
            LCatalogLibrarySource={initialProgramState.defaultConfigureOptionCatalog}
            selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds}
            onToggleOption={toggleConfigureOption}
            showTechnicalDetails={showTechnicalDetails}
            searchQuery={searchQuery}
            selectedCategoryName={selectedCategoryName}
          />

          <PSectionFlagRender
            extraConfigureFlagText={extraConfigureFlagText}
            onExtraFlagTextChange={onExtraFlagTextChange}
          />

          <PSectionThreadRender
            parallelJobCount={ffmpegBuildSettings.parallelJobCount}
            updateFfmpegBuildSettings={updateFfmpegBuildSettings}
          />
        </>
      )}
    </section>
  );
}