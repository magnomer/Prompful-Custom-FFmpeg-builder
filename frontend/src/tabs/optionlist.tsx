import React from "react";
import { LLocaleTextGet } from "../i18n";
import { LOptionTextGet } from "../catalogText";
import optionsCardIcon from "../assets/option-card-icons/Options.svg";

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

function LOptionGroupCreate(LCatalogLibrarySource: LOptionChoice[]) {
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

export function LOptionCategoryList(
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

export function PDropdownCategoryRender(props: {
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

export function PListOptionRender(props: {
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
  const groupedOptions = LOptionGroupCreate(filteredOptions);
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

export function PCardOptionRender(props: {
  LCatalogLibrarySource: LOptionChoice[];
  selectedOptionIds: string[];
  onToggleOption: (optionId: string) => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  selectedCategoryName: string;
  onSelectCategory: (categoryName: string) => void;
  categoryDropdownOpen: boolean;
  onToggleCategoryDropdown: () => void;
  LOptionCategoryNames: string[];
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
            categories={props.LOptionCategoryNames}
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
