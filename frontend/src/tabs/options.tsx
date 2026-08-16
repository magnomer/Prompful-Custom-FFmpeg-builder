import React from "react";
import { LLocaleGet, LLocaleTextGet } from "../i18n";
import { PHeaderPageRender } from "./shared";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";
import resetIcon from "../assets/button-icons/Reset.svg";
import {
  LPresetOptionId,
  LPresetOptionMatch,
  PSelectorOptionRender,
  POptionPresetRender,
} from "./optionpreset";
import {
  LOptionCategoryList,
  PDropdownCategoryRender,
  PListOptionRender,
  PCardOptionRender,
} from "./optionlist";
import {
  PSummaryOptionRender,
  PPanelTechnicalRender,
  PSectionFlagRender,
  PSectionThreadRender,
} from "./optionsection";

export type { LPresetOptionId, LPresetOption } from "./optionpreset";
export { LPresetOptionCatalog, LPresetOptionMatch } from "./optionpreset";
export { LLicenseBoundaryResolve, LLicenseShortGet } from "./optionlicense";

// ─── POptionRender ───────────────────────────────────────────────────────────────

export type LOptionProperties = {
  ffmpegBuildSettings: LSettingsFfmpeg;
  initialProgramState: LStateInitial;
  extraConfigureFlagText: string;
  onExtraFlagTextChange: (text: string) => void;
  LSettingsFfmpegUpdate: (partial: Partial<LSettingsFfmpeg>) => void;
  LOptionToggle: (optionId: string) => void;
  LPresetOptionApply: (presetId: LPresetOptionId) => void;
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
  LSettingsFfmpegUpdate,
  LOptionToggle,
  LPresetOptionApply,
  optionsDetailedView,
  setOptionsDetailedView,
  showTechnicalDetails,
  setShowTechnicalDetails,
}: LOptionProperties) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const [selectedCategoryId, setSelectedCategoryId] = React.useState("");
  const [categoryDropdownOpen, setCategoryDropdownOpen] = React.useState(false);
  const selectedPresetId = LPresetOptionMatch(
    ffmpegBuildSettings.selectedConfigureOptionIds,
  );
  const locale = LLocaleGet();
  const optionCategories = React.useMemo(
    () =>
      LOptionCategoryList(
        initialProgramState.defaultConfigureOptionCatalog,
      ),
    [initialProgramState.defaultConfigureOptionCatalog, locale],
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
          <POptionPresetRender
            selectedPresetId={selectedPresetId}
            onApplyPreset={LPresetOptionApply}
          />
          <PCardOptionRender
            LCatalogLibrarySource={initialProgramState.defaultConfigureOptionCatalog}
            selectedOptionIds={ffmpegBuildSettings.selectedConfigureOptionIds}
            onToggleOption={LOptionToggle}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            selectedCategoryId={selectedCategoryId}
            onSelectCategory={(categoryId) => {
              setSelectedCategoryId(categoryId);
              setCategoryDropdownOpen(false);
            }}
            categoryDropdownOpen={categoryDropdownOpen}
            onToggleCategoryDropdown={() => setCategoryDropdownOpen((value) => !value)}
            optionCategories={optionCategories}
          />
        </div>
      )}

      {optionsDetailedView && (
        <>
          <PSelectorOptionRender
            selectedPresetId={selectedPresetId}
            onApplyPreset={LPresetOptionApply}
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
              categories={optionCategories}
              selectedCategoryId={selectedCategoryId}
              open={categoryDropdownOpen}
              onToggleOpen={() => setCategoryDropdownOpen((value) => !value)}
              onSelectCategory={(categoryId) => {
                setSelectedCategoryId(categoryId);
                setCategoryDropdownOpen(false);
              }}
            />
            <button
              className="button options-reset"
              type="button"
              onClick={() => {
                setSearchQuery("");
                setSelectedCategoryId("");
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
            onToggleOption={LOptionToggle}
            showTechnicalDetails={showTechnicalDetails}
            searchQuery={searchQuery}
            selectedCategoryId={selectedCategoryId}
          />

          <PSectionFlagRender
            extraConfigureFlagText={extraConfigureFlagText}
            onExtraFlagTextChange={onExtraFlagTextChange}
          />

          <PSectionThreadRender
            parallelJobCount={ffmpegBuildSettings.parallelJobCount}
            LSettingsFfmpegUpdate={LSettingsFfmpegUpdate}
          />
        </>
      )}
    </section>
  );
}
