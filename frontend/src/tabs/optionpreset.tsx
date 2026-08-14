import React from "react";
import { LLocaleTextGet } from "../i18n";
import optionPresetCardIcon from "../assets/option-card-icons/OptionPreset.svg";
import optionNoneIcon from "../assets/option-preset-icons/PresetNone.svg";
import optionStandardIcon from "../assets/option-preset-icons/PresetStandard.svg";
import optionCompactIcon from "../assets/option-preset-icons/PresetCompact.svg";
import optionPortableIcon from "../assets/option-preset-icons/PresetPortable.svg";
import optionPerformanceIcon from "../assets/option-preset-icons/PresetPerformance.svg";

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

export const LPresetOptionCatalog: LPresetOption[] = [
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
  for (const preset of LPresetOptionCatalog) {
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

const LPresetOptionIcons: Partial<Record<LPresetOptionId, string>> = {
  none: optionNoneIcon,
  standard: optionStandardIcon,
  compact: optionCompactIcon,
  portable: optionPortableIcon,
  performance: optionPerformanceIcon,
};

function PIconOptionRender(props: { presetId: LPresetOptionId; className?: string }) {
  const icon = LPresetOptionIcons[props.presetId];
  if (!icon) return null;
  return <img className={props.className ?? ""} src={icon} alt="" aria-hidden="true" />;
}

function LPresetSelectableList(): Array<LPresetOption & { presetId: Exclude<LPresetOptionId, "custom"> }> {
  return LPresetOptionCatalog.filter((preset): preset is LPresetOption & { presetId: Exclude<LPresetOptionId, "custom"> } => preset.presetId !== "custom");
}

export function PSelectorOptionRender(props: {
  selectedPresetId: LPresetOptionId;
  onApplyPreset: (presetId: LPresetOptionId) => void;
}) {
  return (
    <section className="preset-panel options-preset-panel">
      <div className="preset-grid options-preset-grid">
        {LPresetSelectableList().map((preset) => {
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

export function POptionPresetRender(props: {
  selectedPresetId: LPresetOptionId;
  onApplyPreset: (presetId: LPresetOptionId) => void;
}) {
  const presets = LPresetSelectableList();
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
