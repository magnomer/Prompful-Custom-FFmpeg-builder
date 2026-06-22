import React from "react";
import { t } from "../i18n";
import { configureOptionText } from "../catalogText";
import { PageHeader } from "./shared";

// ─── Option presets ───────────────────────────────────────────────────────────

export type OptionPresetId = "none" | "standard" | "compact" | "portable" | "performance" | "custom";

export type OptionPreset = {
  presetId: OptionPresetId;
  optionIds: string[];
};

// Locked program defaults are always present; presets layer build-tuning toggles
// on top. High-risk and troubleshooting toggles (enable-shared, disable-asm,
// disable-x86asm, disable-network, etc.) are intentionally in no preset and stay
// reachable only by hand, which keeps every preset safe.
const baseConfigureOptionIds = ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"];

export const optionPresets: OptionPreset[] = [
  { presetId: "none", optionIds: baseConfigureOptionIds },
  { presetId: "standard", optionIds: [...baseConfigureOptionIds, "pkg-config-static", "disable-doc"] },
  { presetId: "compact", optionIds: [...baseConfigureOptionIds, "pkg-config-static", "disable-doc", "disable-debug"] },
  { presetId: "portable", optionIds: [...baseConfigureOptionIds, "pkg-config-static", "disable-doc", "disable-debug", "enable-runtime-cpudetect"] },
  { presetId: "performance", optionIds: [...baseConfigureOptionIds, "pkg-config-static", "disable-doc", "disable-debug", "enable-lto"] },
];

export function matchOptionPresetId(selectedOptionIds: string[]): OptionPresetId {
  const normalizedSelection = Array.from(new Set(selectedOptionIds)).sort();
  for (const preset of optionPresets) {
    const normalizedPreset = Array.from(new Set(preset.optionIds)).sort();
    if (normalizedSelection.length === normalizedPreset.length && normalizedSelection.every((optionId, index) => optionId === normalizedPreset[index])) {
      return preset.presetId;
    }
  }
  return "custom";
}

function OptionPresetSelector(props: { selectedPresetId: OptionPresetId; onApplyPreset: (presetId: OptionPresetId) => void }) {
  return (
    <section className="preset-panel">
      <div className="preset-panel__header">
        <h2 className="preset-panel__title">{t("options.presetSelector.title")}</h2>
        <p className="preset-panel__text">{t("options.presetSelector.text")}</p>
      </div>
      <div className="preset-grid">
        {optionPresets.map((preset) => (
          <button className={`preset-card ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <span className="preset-card__name">{t(`options.presets.${preset.presetId}.name`)}</span>
            <span className="preset-card__plain">{t(`options.presets.${preset.presetId}.plain`)}</span>
          </button>
        ))}
      </div>
      {props.selectedPresetId === "custom" && <p className="preset-panel__custom">{t("options.presetSelector.custom")}</p>}
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
      {line.split(/(\*\*[^*]+\*\*)/g).map((segment, segmentIndex) =>
        segment.startsWith("**") && segment.endsWith("**")
          ? <strong key={segmentIndex}>{renderRedEmphasis(segment.slice(2, -2))}</strong>
          : <React.Fragment key={segmentIndex}>{renderRedEmphasis(segment)}</React.Fragment>,
      )}
    </React.Fragment>
  ));
}

function renderRedEmphasis(text: string): React.ReactNode {
  return text.split(/(<red>[^<]+<\/red>)/g).map((segment, index) =>
    segment.startsWith("<red>") && segment.endsWith("</red>")
      ? <span key={index} className="option-row__risk-word">{segment.slice(5, -6)}</span>
      : segment,
  );
}

function optionRiskLabel(riskLevelName: string): string {
  return t(`options.row.risk.${normalizeOptionRiskName(riskLevelName)}`);
}

function normalizeOptionRiskName(riskLevelName: string): string {
  return riskLevelName === "high" || riskLevelName === "medium" ? riskLevelName : "low";
}

function groupConfigureOptionsByCategory(catalog: ConfigureOptionChoice[]) {
  return catalog.reduce<Record<string, ConfigureOptionChoice[]>>((groups, option) => {
    const categoryName = configureOptionText(option, "categoryName") || t("common.other");
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(option);
    return groups;
  }, {});
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

function OptionSummary(props: { licenseProfileName: string; selectedOptionCount: number }) {
  const licenseBoundary = normalizeLicenseBoundaryName(props.licenseProfileName);
  return (
    <section className="option-summary" aria-label={t("options.summary.ariaLabel")}>
      <div className="option-summary__item">
        <span className="option-summary__label">{t("options.summary.license")}</span>
        <strong className={`option-summary__value option-summary__license option-summary__license--${licenseBoundary}`} title={licenseBoundaryLabel(props.licenseProfileName)}>{licenseBoundaryShortLabel(props.licenseProfileName)}</strong>
      </div>
      <div className="option-summary__item">
        <span className="option-summary__label">{t("options.summary.selected")}</span>
        <strong className="option-summary__value">{props.selectedOptionCount}</strong>
      </div>
    </section>
  );
}

// ─── Technical panel ──────────────────────────────────────────────────────────

function OptionTechnicalPanel() {
  return (
    <section className="option-technical-panel">
      <h2 className="option-technical-panel__title">{t("options.technical.title")}</h2>
      <div className="option-technical-details">
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">{t("options.technical.license.title")}</h3>
          <p className="option-technical-detail__text">{t("options.license.hint")}</p>
          <p className="option-technical-detail__text">
            {t("options.license.rule.lgpl")} <strong>{t("libraries.summary.license.lgpl-local")}</strong>.<br />
            {t("options.license.rule.gpl")} <strong>{t("libraries.summary.license.gpl-local")}</strong>.<br />
            {t("options.license.rule.nonfree.prefix")} <code>--enable-nonfree</code>{t("options.license.rule.nonfree.suffix")} <strong>{t("libraries.summary.license.nonfree-local")}</strong>.
          </p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">{t("options.technical.configure.title")}</h3>
          <p className="option-technical-detail__text">{t("options.technical.configure.text")}</p>
        </section>
        <section className="option-technical-detail">
          <h3 className="option-technical-detail__title">{t("options.technical.advanced.title")}</h3>
          <p className="option-technical-detail__text">{t("options.technical.advanced.text")}</p>
        </section>
      </div>
    </section>
  );
}

// ─── Option list ──────────────────────────────────────────────────────────────

function ConfigureOptionList(props: { catalog: ConfigureOptionChoice[]; selectedOptionIds: string[]; onToggleOption: (optionId: string) => void; showTechnicalDetails: boolean }) {
  const groupedOptions = groupConfigureOptionsByCategory(props.catalog);
  return (
    <div className="option-list">
      {Object.entries(groupedOptions).map(([categoryName, options]) => (
        <section className="option-group" key={categoryName}>
          <h2 className="option-group__title">{categoryName}</h2>
          {options.map((option) => (
            <label className="option-row" key={option.optionId}>
              <input type="checkbox" checked={props.selectedOptionIds.includes(option.optionId)} disabled={option.locked} onChange={() => props.onToggleOption(option.optionId)} />
              <span className="option-row__main">
                <span className="option-row__name">{configureOptionText(option, "displayName")}</span>
                <span className="option-row__plain">{renderOptionRichText(configureOptionText(option, "plainExplanation"))}</span>
                {props.showTechnicalDetails && configureOptionText(option, "technicalNote") && <span className="option-row__detail">{configureOptionText(option, "technicalNote")}</span>}
                {props.showTechnicalDetails && <span className="option-row__detail option-row__detail--flags">{option.configureFlags.length > 0 ? t("options.configure.flags", { flags: option.configureFlags.join(" ") }) : t("options.configure.defaultBehavior")}</span>}
              </span>
              <span className={`option-row__risk option-row__risk--${normalizeOptionRiskName(option.riskLevelName)}`} title={t("options.row.riskLabel")}>{optionRiskLabel(option.riskLevelName)}</span>
            </label>
          ))}
        </section>
      ))}
    </div>
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
};

export function OptionsTab({ ffmpegBuildSettings, initialApplicationState, extraConfigureFlagText, onExtraFlagTextChange, updateFfmpegBuildSettings, toggleConfigureOption, applyOptionPreset }: OptionsTabProps) {
  const [showTechnicalDetails, setShowTechnicalDetails] = React.useState(false);

  return (
    <section className="tab-page options-page">
      <PageHeader title={t("options.title")} text={t("options.intro")} />

      <section className="options-briefing">
        <p>{t("options.info.controls")}</p>
        <p>{t("options.info.defaults")}</p>
      </section>

      <OptionPresetSelector
        selectedPresetId={matchOptionPresetId(ffmpegBuildSettings.selectedConfigureOptionIds)}
        onApplyPreset={applyOptionPreset}
      />

      <OptionSummary
        licenseProfileName={ffmpegBuildSettings.licenseProfileName}
        selectedOptionCount={ffmpegBuildSettings.selectedConfigureOptionIds.length}
      />

      <section className="options-section">
        <label className="field options-field">
          <span className="field__label">{t("options.jobs.label")}</span>
          <span className="field__hint">{t("options.jobs.hint")}</span>
          <input className="field__input" type="number" min="1" max="256" value={ffmpegBuildSettings.parallelJobCount} onChange={(event) => updateFfmpegBuildSettings({ parallelJobCount: Number(event.target.value) })} />
        </label>
      </section>

      <div className="options-toolbar">
        <button className="button options-technical-toggle" type="button" aria-expanded={showTechnicalDetails} onClick={() => setShowTechnicalDetails((value) => !value)}>
          {showTechnicalDetails ? t("options.technical.hide") : t("options.technical.show")}
        </button>
      </div>

      {showTechnicalDetails && <OptionTechnicalPanel />}

      <ConfigureOptionList catalog={initialApplicationState.defaultConfigureOptionCatalog} selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds} onToggleOption={toggleConfigureOption} showTechnicalDetails={showTechnicalDetails} />

      <section className="options-section options-section--advanced">
        <label className="field options-field">
          <span className="field__label">{t("options.advancedFlags.label")}</span>
          <span className="field__hint">{t("options.advancedFlags.hint.prefix")} <code>./configure</code>{t("options.advancedFlags.hint.suffix")}</span>
          <textarea className="field__textarea" rows={5} value={extraConfigureFlagText} onChange={(event) => onExtraFlagTextChange(event.target.value)} placeholder={t("options.advancedFlags.placeholder")} />
        </label>
      </section>
    </section>
  );
}
