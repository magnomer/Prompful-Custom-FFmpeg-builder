import React from "react";
import { LLocaleTextGet } from "../i18n";
import {
  LLicenseBoundaryResolve,
  LLicenseShortGet,
  LLicenseBoundaryNormalize,
} from "./optionlicense";
import { LPresetOptionId } from "./optionpreset";

// ─── Option summary ───────────────────────────────────────────────────────────

export function PSummaryOptionRender(props: {
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
            {LLicenseShortGet(props.licenseProfileName)}
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

export function PPanelTechnicalRender() {
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

// ─── Advanced flags and jobs sections ──────────────────────────────────────────

export function PSectionFlagRender(props: {
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

export function PSectionThreadRender(props: {
  parallelJobCount: number;
  LSettingsFfmpegUpdate: (partial: Partial<LSettingsFfmpeg>) => void;
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
            props.LSettingsFfmpegUpdate({
              parallelJobCount: Number(event.target.value),
            })
          }
        />
      </label>
    </section>
  );
}
