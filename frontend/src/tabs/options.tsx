import React from "react";
import { t } from "../i18n";
import { configureOptionText } from "../catalogText";
import { PageHeader } from "./shared";

// ─── Preset data ───────────────────────────────────────────────────────────────

export type OptionPresetId =
  | "none"
  | "standard"
  | "compact"
  | "portable"
  | "performance"
  | "custom";

export type OptionPreset = {
  presetId: OptionPresetId;
  optionIds: string[];
};

// Locked program defaults are always present; presets layer build-tuning toggles
// on top. High-risk and troubleshooting toggles (enable-shared, disable-asm,
// disable-x86asm, disable-network, etc.) are intentionally in no preset and stay
// reachable only by hand, which keeps every preset safe.
const baseConfigureOptionIds = [
  "default-static",
  "default-programs",
  "default-ffmpeg",
  "default-ffprobe",
];

export const optionPresets: OptionPreset[] = [
  { presetId: "none", optionIds: baseConfigureOptionIds },
  {
    presetId: "standard",
    optionIds: [...baseConfigureOptionIds, "pkg-config-static", "disable-doc"],
  },
  {
    presetId: "compact",
    optionIds: [
      ...baseConfigureOptionIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
    ],
  },
  {
    presetId: "portable",
    optionIds: [
      ...baseConfigureOptionIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
      "enable-runtime-cpudetect",
    ],
  },
  {
    presetId: "performance",
    optionIds: [
      ...baseConfigureOptionIds,
      "pkg-config-static",
      "disable-doc",
      "disable-debug",
      "enable-lto",
    ],
  },
];

export function matchOptionPresetId(
  selectedOptionIds: string[],
): OptionPresetId {
  const normalizedSelection = Array.from(new Set(selectedOptionIds)).sort();
  for (const preset of optionPresets) {
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

function selectableOptionPresets(): Array<OptionPreset & { presetId: Exclude<OptionPresetId, "custom"> }> {
  return optionPresets.filter((preset): preset is OptionPreset & { presetId: Exclude<OptionPresetId, "custom"> } => preset.presetId !== "custom");
}

function OptionPresetSelector(props: {
  selectedPresetId: OptionPresetId;
  onApplyPreset: (presetId: OptionPresetId) => void;
}) {
  return (
    <section className="preset-panel options-preset-panel">
      <div className="preset-grid options-preset-grid">
        {selectableOptionPresets().map((preset) => {
          const active = props.selectedPresetId === preset.presetId;
          return (
            <button
              className={`preset-card options-preset-card ${active ? "preset-card--active" : ""}`}
              type="button"
              key={preset.presetId}
              onClick={() => props.onApplyPreset(preset.presetId)}
            >
              <span>
                <span className="preset-card__name">
                  {t(`options.presets.${preset.presetId}.name`)}
                </span>
                <span className="preset-card__plain">
                  {t(`options.presets.${preset.presetId}.plain`)}
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
          {t("options.presetSelector.custom")}
        </p>
      )}
    </section>
  );
}

function SimpleOptionPresetCard(props: {
  selectedPresetId: OptionPresetId;
  onApplyPreset: (presetId: OptionPresetId) => void;
}) {
  const presets = selectableOptionPresets();
  const selectedPreset = presets.find((preset) => preset.presetId === props.selectedPresetId);
  const selectedPresetDescription = selectedPreset
    ? t(`options.presets.${selectedPreset.presetId}.plain`)
    : t("options.presetSelector.custom");

  return (
    <section className="card card--blue options-simple-card options-simple-preset-card">
      <span className="card__badge" aria-hidden="true" />
      <div className="card__head">
        <h2 className="card__title">{t("options.simple.preset.title")}</h2>
      </div>
      <div className="card__control options-simple-card__control">
        <select
          className="card__input"
          value={props.selectedPresetId}
          onChange={(event) => props.onApplyPreset(event.target.value as OptionPresetId)}
        >
          {props.selectedPresetId === "custom" && <option value="custom">{t("options.presets.custom.name")}</option>}
          {presets.map((preset) => (
            <option value={preset.presetId} key={preset.presetId}>
              {t(`options.presets.${preset.presetId}.name`)}
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
function renderOptionRichText(text: string): React.ReactNode {
  return text.split("\n").map((line, lineIndex) => (
    <React.Fragment key={lineIndex}>
      {lineIndex > 0 && <br />}
      {line
        .split(/(\*\*[^*]+\*\*)/g)
        .map((segment, segmentIndex) =>
          segment.startsWith("**") && segment.endsWith("**") ? (
            <strong key={segmentIndex}>
              {renderRedEmphasis(segment.slice(2, -2))}
            </strong>
          ) : (
            <React.Fragment key={segmentIndex}>
              {renderRedEmphasis(segment)}
            </React.Fragment>
          ),
        )}
    </React.Fragment>
  ));
}

function renderRedEmphasis(text: string): React.ReactNode {
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

function optionRiskLabel(riskLevelName: string): string {
  return t(`options.row.risk.${normalizeOptionRiskName(riskLevelName)}`);
}

function normalizeOptionRiskName(riskLevelName: string): string {
  return riskLevelName === "high" || riskLevelName === "medium"
    ? riskLevelName
    : "low";
}

function groupConfigureOptionsByCategory(catalog: ConfigureOptionChoice[]) {
  return catalog.reduce<Record<string, ConfigureOptionChoice[]>>(
    (groups, option) => {
      const categoryName =
        configureOptionText(option, "categoryName") || t("common.other");
      groups[categoryName] = groups[categoryName] || [];
      groups[categoryName].push(option);
      return groups;
    },
    {},
  );
}

export function licenseBoundaryLabel(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return t("options.licenseBoundary.gpl");
    case "nonfree-local":
      return t("options.licenseBoundary.nonfree");
    case "lgpl-local":
    default:
      return t("options.licenseBoundary.lgpl");
  }
}

function licenseBoundaryShortLabel(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
      return t("options.summary.license.gpl-local");
    case "nonfree-local":
      return t("options.summary.license.nonfree-local");
    case "lgpl-local":
    default:
      return t("options.summary.license.lgpl-local");
  }
}

function normalizeLicenseBoundaryName(licenseProfileName: string): string {
  switch (licenseProfileName) {
    case "gpl-local":
    case "nonfree-local":
      return licenseProfileName;
    default:
      return "lgpl-local";
  }
}

// ─── Option summary ───────────────────────────────────────────────────────────

function OptionSummary(props: {
  licenseProfileName: string;
  selectedOptionCount: number;
  selectedPresetId: OptionPresetId;
}) {
  const licenseBoundary = normalizeLicenseBoundaryName(
    props.licenseProfileName,
  );
  return (
    <section
      className="option-summary-card"
      aria-label={t("options.summary.ariaLabel")}
    >
      <div className="option-summary-card__header">
        <span className="option-summary-card__status" aria-hidden="true">
          ✓
        </span>
        <div className="option-summary-card__copy">
          <span className="option-summary-card__title">
            {t("options.summary.currentSelection")}
          </span>
          <span className="option-summary-card__message">
            {t(`options.presets.${props.selectedPresetId}.name`)}
          </span>
        </div>
      </div>
      <div className="option-summary">
        <div className="option-summary__item">
          <span className="option-summary__label">
            {t("options.summary.license")}
          </span>
          <strong
            className={`option-summary__value option-summary__license option-summary__license--${licenseBoundary}`}
            title={licenseBoundaryLabel(props.licenseProfileName)}
          >
            {licenseBoundaryShortLabel(props.licenseProfileName)}
          </strong>
        </div>
        <div className="option-summary__item">
          <span className="option-summary__label">
            {t("options.summary.selected")}
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

function OptionTechnicalPanel() {
  return (
    <section className="option-technical-panel">
      <h2 className="option-technical-panel__title">
        {t("options.technical.title")}
      </h2>
      <div className="option-technical-details">
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {t("options.technical.license.title")}
          </h3>
          <p className="option-technical-detail__text">
            {t("options.license.hint")}
          </p>
          <p className="option-technical-detail__text">
            {t("options.license.rule.lgpl")}{" "}
            <strong>{t("libraries.summary.license.lgpl-local")}</strong>.<br />
            {t("options.license.rule.gpl")}{" "}
            <strong>{t("libraries.summary.license.gpl-local")}</strong>.<br />
            {t("options.license.rule.nonfree.prefix")}{" "}
            <code>--enable-nonfree</code>
            {t("options.license.rule.nonfree.suffix")}{" "}
            <strong>{t("libraries.summary.license.nonfree-local")}</strong>.
          </p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {t("options.technical.configure.title")}
          </h3>
          <p className="option-technical-detail__text">
            {t("options.technical.configure.text")}
          </p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">
            {t("options.technical.advanced.title")}
          </h3>
          <p className="option-technical-detail__text">
            {t("options.technical.advanced.text")}
          </p>
        </section>
      </div>
    </section>
  );
}

function optionMatchesSearch(
  option: ConfigureOptionChoice,
  query: string,
): boolean {
  const normalizedQuery = query.trim().toLowerCase();
  if (!normalizedQuery) {
    return true;
  }
  return [
    configureOptionText(option, "displayName"),
    configureOptionText(option, "plainExplanation"),
    configureOptionText(option, "technicalNote"),
    configureOptionText(option, "categoryName"),
    ...option.configureFlags,
  ].some((value) => value.toLowerCase().includes(normalizedQuery));
}

function configureOptionCategoryNames(
  catalog: ConfigureOptionChoice[],
): string[] {
  return Array.from(
    new Set(
      catalog.map(
        (option) =>
          configureOptionText(option, "categoryName") || t("common.other"),
      ),
    ),
  );
}

function OptionsCategoryDropdown(props: {
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
          {props.selectedCategoryName || t("options.category.all")}
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
            <span>{t("options.category.all")}</span>
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

function ConfigureOptionList(props: {
  catalog: ConfigureOptionChoice[];
  selectedOptionIds: string[];
  onToggleOption: (optionId: string) => void;
  showTechnicalDetails: boolean;
  searchQuery: string;
  selectedCategoryName: string;
}) {
  const filteredOptions = props.catalog.filter(
    (option) =>
      (!props.selectedCategoryName ||
        (configureOptionText(option, "categoryName") || t("common.other")) ===
          props.selectedCategoryName) &&
      optionMatchesSearch(option, props.searchQuery),
  );
  const groupedOptions = groupConfigureOptionsByCategory(filteredOptions);
  return (
    <div className="option-list">
      {Object.keys(groupedOptions).length === 0 && (
        <section className="option-empty">{t("options.empty")}</section>
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
                  {configureOptionText(option, "displayName")}
                </span>
                <span className="option-row__plain">
                  {renderOptionRichText(
                    configureOptionText(option, "plainExplanation"),
                  )}
                </span>
                {props.showTechnicalDetails &&
                  configureOptionText(option, "technicalNote") && (
                    <span className="option-row__detail">
                      {configureOptionText(option, "technicalNote")}
                    </span>
                  )}
                {props.showTechnicalDetails && (
                  <span className="option-row__detail option-row__detail--flags">
                    {option.configureFlags.length > 0
                      ? t("options.configure.flags", {
                          flags: option.configureFlags.join(" "),
                        })
                      : t("options.configure.defaultBehavior")}
                  </span>
                )}
              </span>
              <span
                className={`option-row__risk option-row__risk--${normalizeOptionRiskName(option.riskLevelName)}`}
                title={t("options.row.riskLabel")}
              >
                {optionRiskLabel(option.riskLevelName)}
              </span>
            </label>
          ))}
        </section>
      ))}
    </div>
  );
}

function AdvancedConfigureFlagsSection(props: {
  extraConfigureFlagText: string;
  onExtraFlagTextChange: (text: string) => void;
}) {
  return (
    <section className="options-section options-section--advanced">
      <h2 className="options-section__title">
        {t("options.advancedFlags.label")}
      </h2>
      <label className="field options-field options-section__body">
        <span className="field__hint">
          {t("options.advancedFlags.hint.prefix")} <code>./configure</code>
          {t("options.advancedFlags.hint.suffix")}
        </span>
        <textarea
          className="field__textarea"
          rows={5}
          value={props.extraConfigureFlagText}
          onChange={(event) => props.onExtraFlagTextChange(event.target.value)}
          placeholder={t("options.advancedFlags.placeholder")}
        />
      </label>
    </section>
  );
}

function ThreadsSection(props: {
  parallelJobCount: number;
  updateFfmpegBuildSettings: (partial: Partial<FfmpegBuildSettings>) => void;
}) {
  return (
    <section className="options-section">
      <h2 className="options-section__title">{t("options.jobs.label")}</h2>
      <label className="field options-field options-section__body">
        <span className="field__hint">{t("options.jobs.hint")}</span>
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

function SimpleOptionsCard(props: {
  catalog: ConfigureOptionChoice[];
  selectedOptionIds: string[];
  onToggleOption: (optionId: string) => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  selectedCategoryName: string;
  onSelectCategory: (categoryName: string) => void;
  categoryDropdownOpen: boolean;
  onToggleCategoryDropdown: () => void;
  optionCategoryNames: string[];
}) {
  return (
    <section className="card card--teal options-simple-card options-simple-options-card">
      <span className="card__badge" aria-hidden="true" />
      <div className="card__head">
        <h2 className="card__title">{t("options.simple.options.title")}</h2>
      </div>
      <div className="options-simple-options-card__body">
        <div className="options-simple-options-card__controls">
          <label className="options-search">
            <span className="visually-hidden">{t("options.search.label")}</span>
            <input
              type="search"
              value={props.searchQuery}
              onChange={(event) => props.onSearchQueryChange(event.target.value)}
              placeholder={t("options.search.placeholder")}
            />
          </label>
          <OptionsCategoryDropdown
            categories={props.optionCategoryNames}
            selectedCategoryName={props.selectedCategoryName}
            open={props.categoryDropdownOpen}
            onToggleOpen={props.onToggleCategoryDropdown}
            onSelectCategory={props.onSelectCategory}
          />
        </div>
        <div className="options-simple-options-card__results">
          <ConfigureOptionList
            catalog={props.catalog}
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

// ─── OptionsTab ───────────────────────────────────────────────────────────────

export type OptionsTabProps = {
  ffmpegBuildSettings: FfmpegBuildSettings;
  initialApplicationState: InitialApplicationState;
  extraConfigureFlagText: string;
  onExtraFlagTextChange: (text: string) => void;
  updateFfmpegBuildSettings: (partial: Partial<FfmpegBuildSettings>) => void;
  toggleConfigureOption: (optionId: string) => void;
  applyOptionPreset: (presetId: OptionPresetId) => void;
  optionsDetailedView: boolean;
  setOptionsDetailedView: (value: boolean) => void;
  showTechnicalDetails: boolean;
  setShowTechnicalDetails: (value: boolean) => void;
};

export function OptionsTab({
  ffmpegBuildSettings,
  initialApplicationState,
  extraConfigureFlagText,
  onExtraFlagTextChange,
  updateFfmpegBuildSettings,
  toggleConfigureOption,
  applyOptionPreset,
  optionsDetailedView,
  setOptionsDetailedView,
  showTechnicalDetails,
  setShowTechnicalDetails,
}: OptionsTabProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [selectedCategoryName, setSelectedCategoryName] = React.useState("");
  const [categoryDropdownOpen, setCategoryDropdownOpen] = React.useState(false);
  const selectedPresetId = matchOptionPresetId(
    ffmpegBuildSettings.selectedConfigureOptionIds,
  );
  const optionCategoryNames = React.useMemo(
    () =>
      configureOptionCategoryNames(
        initialApplicationState.defaultConfigureOptionCatalog,
      ),
    [initialApplicationState.defaultConfigureOptionCatalog],
  );

  return (
    <section className="tab-page options-page">
      <div className="options-page__header-row">
        <PageHeader title={t("options.title")} text={t("options.intro")} />
        <label className="options-design-toggle">
          <input
            type="checkbox"
            checked={optionsDetailedView}
            onChange={(event) => setOptionsDetailedView(event.target.checked)}
          />
          <span className="options-design-toggle__text">
            {t("options.designToggle.label")}
          </span>
        </label>
      </div>

      {!optionsDetailedView && (
        <div className="options-simple-layout">
          <SimpleOptionPresetCard
            selectedPresetId={selectedPresetId}
            onApplyPreset={applyOptionPreset}
          />
          <SimpleOptionsCard
            catalog={initialApplicationState.defaultConfigureOptionCatalog}
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
            optionCategoryNames={optionCategoryNames}
          />
        </div>
      )}

      {optionsDetailedView && (
        <>
          <OptionPresetSelector
            selectedPresetId={selectedPresetId}
            onApplyPreset={applyOptionPreset}
          />

          <OptionSummary
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
              {showTechnicalDetails
                ? t("options.technical.hide")
                : t("options.technical.show")}
            </button>
            <label className="options-search">
              <span className="visually-hidden">{t("options.search.label")}</span>
              <input
                type="search"
                value={searchQuery}
                onChange={(event) => setSearchQuery(event.target.value)}
                placeholder={t("options.search.placeholder")}
              />
            </label>
            <OptionsCategoryDropdown
              categories={optionCategoryNames}
              selectedCategoryName={selectedCategoryName}
              open={categoryDropdownOpen}
              onToggleOpen={() => setCategoryDropdownOpen((value) => !value)}
              onSelectCategory={(categoryName) => {
                setSelectedCategoryName(categoryName);
                setCategoryDropdownOpen(false);
              }}
            />
          </div>

          {showTechnicalDetails && <OptionTechnicalPanel />}

          <ConfigureOptionList
            catalog={initialApplicationState.defaultConfigureOptionCatalog}
            selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds}
            onToggleOption={toggleConfigureOption}
            showTechnicalDetails={showTechnicalDetails}
            searchQuery={searchQuery}
            selectedCategoryName={selectedCategoryName}
          />

          <AdvancedConfigureFlagsSection
            extraConfigureFlagText={extraConfigureFlagText}
            onExtraFlagTextChange={onExtraFlagTextChange}
          />

          <ThreadsSection
            parallelJobCount={ffmpegBuildSettings.parallelJobCount}
            updateFfmpegBuildSettings={updateFfmpegBuildSettings}
          />
        </>
      )}
    </section>
  );
}