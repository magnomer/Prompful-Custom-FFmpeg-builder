import React from "react";
import { LLocaleTextGet } from "../i18n";
import { LLicenseLabelGet, LLibraryTextGet } from "../catalogText";
import { LUnlockSudoCheck } from "../devUnlock";
import libraryMinimalIcon from "../assets/library-preset-icons/LibraryMinimal.svg";
import libraryDefaultIcon from "../assets/library-preset-icons/LibraryDefault.svg";
import libraryMaximumEfficiencyIcon from "../assets/library-preset-icons/LibraryMaximumEfficiency.svg";
import libraryMaximumCompatibilityIcon from "../assets/library-preset-icons/LibraryMaximumCompatibility.svg";
import libraryMaximumAudioVideoEditorIcon from "../assets/library-preset-icons/LibraryMaximumAudioVideoEditor.svg";
import libraryMaximumFullIcon from "../assets/library-preset-icons/LibraryMaximumFull.svg";
import libraryDevIcon from "../assets/library-preset-icons/LibraryDev.svg";
import libraryPresetCardIcon from "../assets/library-card-icons/LibraryPreset.svg";
import librariesCardIcon from "../assets/library-card-icons/Libraries.svg";
import technicalDetailsIcon from "../assets/button-icons/TechnicalDetails.svg";
import resetIcon from "../assets/button-icons/Reset.svg";
import {
  LSectionState,
  LSectionOption,
  LLibraryFilter,
  LLibraryFilterCreate,
  LLibraryUnavailableCheck,
  LLibraryReasonGet,
  LLibraryLockedCheck,
  LLibraryTrackGet,
  LLibraryNameSplit,
  LLibraryVisibleFilter,
  LSectionLabelGet,
  LSectionSummaryGet,
} from "./librarycatalog";
import {
  LPresetLibrary,
  LPresetLibraryId,
  LLibrarySelectionNormalize,
  LLicenseBoundaryGet,
  LPresetNameGet,
  LPresetLibraryList,
  LLibraryTlsBackend,
  LLibraryShaderCompiler,
  LLibraryEvcDecoder,
  LLibraryEvcEncoder,
  LLibraryIntelBackend,
} from "./librarypresets";

const LPresetIconTable: Partial<Record<LPresetLibraryId, string>> = {
  minimal: libraryMinimalIcon,
  default: libraryDefaultIcon,
  efficiency: libraryMaximumEfficiencyIcon,
  compatibility: libraryMaximumCompatibilityIcon,
  editor: libraryMaximumAudioVideoEditorIcon,
  full: libraryMaximumFullIcon,
  maxtest: libraryDevIcon,
};

function PIconPresetRender(props: { presetId: LPresetLibraryId; className?: string }) {
  const icon = LPresetIconTable[props.presetId];
  if (!icon) return null;
  return <img className={props.className ?? ""} src={icon} alt="" aria-hidden="true" />;
}

// ─── Library UI components ────────────────────────────────────────────────────

// Renders a library catalog note with minimal markup: "\n" becomes a line break and
// "**text**" becomes bold. No other markdown is interpreted.
function PNoteRichRender(text: string): React.ReactNode {
  return text.split("\n").map((line, lineIndex) => (
    <React.Fragment key={lineIndex}>
      {lineIndex > 0 && <br />}
      {line.split(/(\*\*[^*]+\*\*)/g).map((segment, segmentIndex) =>
        segment.startsWith("**") && segment.endsWith("**")
          ? <strong key={segmentIndex}>{segment.slice(2, -2)}</strong>
          : segment,
      )}
    </React.Fragment>
  ));
}

export function PCardPresetRender(props: { presets: LPresetLibrary[]; selectedPresetId: LPresetLibraryId; onApplyPreset: (presetId: LPresetLibraryId) => void; extendedLibraries: boolean }) {
  const presets = LPresetLibraryList(props.presets);
  const selectedPreset = presets.find((preset) => preset.presetId === props.selectedPresetId);
  const selectedPresetDescription = selectedPreset
    ? LLocaleTextGet(`libraries.presets.${selectedPreset.presetId}.plain`)
    : LLocaleTextGet("libraries.presetSelector.custom");

  return (
    <section className="card card--blue libraries-simple-card libraries-simple-preset-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={libraryPresetCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("libraries.simple.preset.title")}</h2>
      </div>
      <div className="card__control libraries-simple-card__control">
        <select
          className="card__input"
          value={props.selectedPresetId}
          onChange={(event) => props.onApplyPreset(event.target.value as LPresetLibraryId)}
        >
          {props.selectedPresetId === "custom" && <option value="custom">{LLocaleTextGet("libraries.presets.custom.name")}</option>}
          {presets.map((preset) => (
            <option value={preset.presetId} key={preset.presetId}>
              {LPresetNameGet(preset.presetId, props.extendedLibraries)}
            </option>
          ))}
        </select>
      </div>
      <div className="libraries-simple-preset-card__description" aria-live="polite">
        {selectedPresetDescription}
      </div>
    </section>
  );
}

export function PCardLibraryRender(props: {
  LCatalogLibrarySource: LLibraryChoice[];
  selectedLibraryIds: string[];
  onToggleLibrary: (libraryId: string) => void;
  windowsShellProfileName: string;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LSectionState[];
  sectionOptions: LSectionOption[];
  onSectionFiltersChange: (value: LSectionState[]) => void;
  onOpenOfficialWebpage: (url: string) => void;
}) {
  return (
    <section className="card card--teal libraries-simple-card libraries-simple-library-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={librariesCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{LLocaleTextGet("libraries.simple.library.title")}</h2>
      </div>
      <div className="libraries-simple-library-card__body">
        <div className="libraries-simple-library-card__controls">
          <label className="libraries-search">
            <span className="visually-hidden">{LLocaleTextGet("libraries.search.label")}</span>
            <input value={props.searchQuery} onChange={(event) => props.onSearchQueryChange(event.target.value)} placeholder={LLocaleTextGet("libraries.search.placeholder")} />
          </label>
          <PDropdownSectionRender
            selectedSections={props.sectionFilters}
            sectionOptions={props.sectionOptions}
            onChangeSections={props.onSectionFiltersChange}
          />
        </div>
        <div className="libraries-simple-library-card__results">
          <PListLibraryRender
            LCatalogLibrarySource={props.LCatalogLibrarySource}
            selectedLibraryIds={props.selectedLibraryIds}
            onToggleLibrary={props.onToggleLibrary}
            showTechnicalDetails={false}
            windowsShellProfileName={props.windowsShellProfileName}
            searchQuery={props.searchQuery}
            sectionFilters={props.sectionFilters}
            onOpenOfficialWebpage={props.onOpenOfficialWebpage}
          />
        </div>
      </div>
    </section>
  );
}

export function PSelectorPresetRender(props: { presets: LPresetLibrary[]; selectedPresetId: LPresetLibraryId; onApplyPreset: (presetId: LPresetLibraryId) => void; showTechnicalDetails: boolean; extendedLibraries: boolean }) {
  const selectedPreset = props.presets.find((preset) => preset.presetId === props.selectedPresetId && preset.presetId !== "custom") as (LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> }) | undefined;

  return (
    <section className="preset-panel">
      <div className="preset-grid">
        {props.presets.filter((preset): preset is LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> } => LPresetLibraryList(props.presets).some((selectablePreset) => selectablePreset.presetId === preset.presetId)).map((preset) => (
          <button className={`preset-card ${preset.dev ? "preset-card--dev" : ""} ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <PIconPresetRender presetId={preset.presetId} className="preset-card__icon" />
            <span>
              <span className="preset-card__name">{LPresetNameGet(preset.presetId, props.extendedLibraries)}</span>
              <span className="preset-card__plain">{LLocaleTextGet(`libraries.presets.${preset.presetId}.plain`)}</span>
            </span>
            {props.selectedPresetId === preset.presetId && <span className="preset-card__check" aria-hidden="true">✓</span>}
          </button>
        ))}
      </div>
      {props.showTechnicalDetails && selectedPreset && (
        <section className="preset-technical-card" aria-label={LLocaleTextGet("libraries.technical.show")}>
          <div className="preset-technical-card__header">
            <span className="preset-technical-card__title">
              <strong>{LPresetNameGet(selectedPreset.presetId, props.extendedLibraries)}</strong>
              <span>{LLocaleTextGet("libraries.technical.show")}</span>
            </span>
          </div>
          <p className="preset-technical-card__text">{LLocaleTextGet(`libraries.presets.${selectedPreset.presetId}.technical`)}</p>
        </section>
      )}
    </section>
  );
}

export function PSummaryLibraryRender(props: { LCatalogLibrarySource: LLibraryChoice[]; selectedLibraryIds: string[]; selectedPresetId: LPresetLibraryId; windowsShellProfileName: string; extendedLibraries: boolean }) {
  const normalizedSelection = LLibrarySelectionNormalize(props.selectedLibraryIds, props.windowsShellProfileName, props.LCatalogLibrarySource);
  const visibleCatalog = LLibraryVisibleFilter(props.LCatalogLibrarySource, props.windowsShellProfileName);
  const licenseBoundary = LLicenseBoundaryGet(normalizedSelection, props.LCatalogLibrarySource, props.windowsShellProfileName);
  const selectedOptionalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && !library.defaultChecked).length;
  const selectedInternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "internal").length;
  const selectedExternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "external").length;
  const includedCount = visibleCatalog.filter((library) => library.defaultChecked).length;

  const selectedPresetName = LPresetNameGet(props.selectedPresetId, props.extendedLibraries);
  const selectionMessage = props.selectedPresetId === "custom"
    ? LLocaleTextGet("libraries.presetSelector.custom")
    : LLocaleTextGet("libraries.summary.currentPreset", { preset: selectedPresetName });

  return (
    <section className="library-summary-card" aria-label={LLocaleTextGet("libraries.summary.ariaLabel")}>
      <div className="library-summary-card__header">
        <span className="library-summary-card__status" aria-hidden="true">✓</span>
        <span className="library-summary-card__copy">
          <strong className="library-summary-card__title">{LLocaleTextGet("libraries.summary.currentTitle")}</strong>
          <span className="library-summary-card__message">{selectionMessage}</span>
        </span>
      </div>
      <div className="library-summary">
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{LLocaleTextGet("libraries.summary.license")}</span>
            <strong className={`library-summary__value library-summary__license library-summary__license--${licenseBoundary}`}>{LLocaleTextGet(`libraries.summary.license.${licenseBoundary}`)}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{LLocaleTextGet("libraries.summary.included")}</span>
            <strong className="library-summary__value">{includedCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{LLocaleTextGet("libraries.summary.optional")}</span>
            <strong className="library-summary__value">{selectedOptionalCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{LLocaleTextGet("libraries.summary.internal")}</span>
            <strong className="library-summary__value">{selectedInternalCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{LLocaleTextGet("libraries.summary.externalTrack")}</span>
            <strong className="library-summary__value">{selectedExternalCount}</strong>
          </span>
        </div>
      </div>
    </section>
  );
}

export function PListLibraryRender(props: { LCatalogLibrarySource: LLibraryChoice[]; selectedLibraryIds: string[]; onToggleLibrary: (libraryId: string) => void; showTechnicalDetails: boolean; windowsShellProfileName: string; searchQuery: string; sectionFilters: LSectionState[]; onOpenOfficialWebpage: (url: string) => void }) {
  const filteredLibraries = LLibraryFilter(props.LCatalogLibrarySource, props.windowsShellProfileName, props.searchQuery, props.sectionFilters);
  const groupedLibraries = LLibraryFilterCreate(filteredLibraries);
  return (
    <div className="library-list">
      {filteredLibraries.length === 0 && <p className="library-list__empty">{LLocaleTextGet("libraries.list.empty")}</p>}
      {Object.entries(groupedLibraries).map(([categoryName, libraries]) => (
        <section className="library-group" key={categoryName}>
          <h2 className="library-group__title">{categoryName}</h2>
          {libraries.map((library) => {
            const isUiUnavailable = LLibraryUnavailableCheck(library, props.windowsShellProfileName);
            const unavailableReasonKey = LLibraryReasonGet(library, props.windowsShellProfileName);
            const isCheckboxLocked = LLibraryLockedCheck(library, props.windowsShellProfileName);
            const isChecked = (props.selectedLibraryIds.includes(library.libraryId) || library.defaultChecked) && !isCheckboxLocked;
            // TLS backends, shaderc/glslang, and the EVC binding pairs each render as a
            // radio group (pick one) in normal and basic dev mode; clicking the selected
            // radio clears it (FFmpeg allows zero). Only the sudo dev tier switches them
            // back to checkboxes so several can be selected for testing. Each group needs
            // its own radio `name` so unrelated groups do not clear each other.
            const radioGroupName = LUnlockSudoCheck()
              ? undefined
              : LLibraryTlsBackend.has(library.libraryId)
                ? "tls-backend"
                : LLibraryShaderCompiler.has(library.libraryId)
                  ? "shader-compiler"
                  : LLibraryEvcDecoder.has(library.libraryId)
                    ? "evc-decoder"
                    : LLibraryEvcEncoder.has(library.libraryId)
                      ? "evc-encoder"
                      : LLibraryIntelBackend.has(library.libraryId)
                        ? "intel-hwaccel-backend"
                        : undefined;
            const isExclusiveRadio = radioGroupName !== undefined;
            const trackName = library.trackName || "native";
            const isNativeTrack = trackName === "native";
            const showsOfficialWebpage = !isNativeTrack && Boolean(library.officialWebpageUrl);
            const showsPackageNames = isNativeTrack;
            const packageValue = library.packageNames.length > 0 ? library.packageNames.join(", ") : LLocaleTextGet("libraries.row.ffmpegSourcePackage");
            const hasTechnicalMetadata = library.configureFlags.length > 0 || showsOfficialWebpage || showsPackageNames;
            const displayName = LLibraryTextGet(library, "displayName");
            const { buildName, featureName } = LLibraryNameSplit(displayName);
            const statusLabel = library.defaultChecked ? LLocaleTextGet("libraries.row.included") : LLicenseLabelGet(library.licenseEffectName);
            return (
              <label className={`library-row library-row--track-${trackName} ${props.showTechnicalDetails ? "library-row--technical-open" : ""} ${library.locked ? "library-row--locked" : ""} ${isUiUnavailable ? "library-row--unavailable" : ""} ${isExclusiveRadio ? "library-row--tls" : ""}`} key={library.libraryId}>
                <input
                  type={isExclusiveRadio ? "radio" : "checkbox"}
                  name={radioGroupName}
                  checked={isChecked}
                  disabled={library.locked || isCheckboxLocked}
                  onChange={() => props.onToggleLibrary(library.libraryId)}
                  onClick={isExclusiveRadio && isChecked ? (event) => { event.preventDefault(); props.onToggleLibrary(library.libraryId); } : undefined}
                />
                <span className="library-row__main">
                  <span className="library-row__heading">
                    <span className="library-row__name">{buildName}</span>
                    {!isNativeTrack && <span className={`library-row__track library-row__track--${trackName}`}>{LLibraryTrackGet(trackName)}</span>}
                  </span>
                  <span className="library-row__copy">
                    {featureName && <span className="library-row__feature">{featureName}</span>}
                    <span className="library-row__note">{PNoteRichRender(LLibraryTextGet(library, "plainExplanation"))}</span>
                    {isUiUnavailable && unavailableReasonKey && <span className="library-row__note">{LLocaleTextGet(`libraries.row.unavailable.${unavailableReasonKey}`)}</span>}
                  </span>
                  {props.showTechnicalDetails && LLibraryTextGet(library, "technicalExplanation") && <span className="library-row__detail">{LLibraryTextGet(library, "technicalExplanation")}</span>}
                  {props.showTechnicalDetails && hasTechnicalMetadata &&
                    <span className="library-row__technical-metadata">
                      {library.configureFlags.length > 0 &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--flag">{LLocaleTextGet("libraries.row.flagsLabel")}</span>
                          <span className="library-row__technical-value">{library.configureFlags.join(" ")}</span>
                        </span>
                      }
                      {showsPackageNames &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--package">{LLocaleTextGet("libraries.row.packagesLabel")}</span>
                          <span className="library-row__technical-value">{packageValue}</span>
                        </span>
                      }
                      {showsOfficialWebpage &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--official">{LLocaleTextGet("libraries.row.officialWebpageLabel")}</span>
                          <button
                            className="library-row__technical-link"
                            type="button"
                            onClick={(event) => {
                              event.preventDefault();
                              event.stopPropagation();
                              props.onOpenOfficialWebpage(library.officialWebpageUrl);
                            }}
                          >
                            {library.officialWebpageUrl}
                          </button>
                        </span>
                      }
                    </span>
                  }
                </span>
                {!props.showTechnicalDetails &&
                  <span className="library-row__status">
                    <span className={`library-row__license library-row__license--${library.licenseEffectName}`}>{statusLabel}</span>
                  </span>
                }
                {props.showTechnicalDetails &&
                  <span className="library-row__status library-row__status--technical">
                    <span className={`library-row__license library-row__license--${library.licenseEffectName}`}>{statusLabel}</span>
                  </span>
                }
              </label>
            );
          })}
        </section>
      ))}
    </div>
  );
}

function PDropdownSectionRender(props: { selectedSections: LSectionState[]; sectionOptions: LSectionOption[]; onChangeSections: (value: LSectionState[]) => void }) {
  const [isOpen, setIsOpen] = React.useState(false);
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (!isOpen) return;
    function PDropdownClose(event: PointerEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("pointerdown", PDropdownClose);
    return () => document.removeEventListener("pointerdown", PDropdownClose);
  }, [isOpen]);

  function LSectionToggle(sectionName: LSectionState) {
    const nextSections = props.selectedSections.includes(sectionName)
      ? props.selectedSections.filter((selectedSection) => selectedSection !== sectionName)
      : [...props.selectedSections, sectionName];
    props.onChangeSections(nextSections);
  }

  return (
    <div className="libraries-section-dropdown" ref={dropdownRef}>
      <button
        className={`libraries-section-dropdown__button ${isOpen ? "libraries-section-dropdown__button--open" : ""}`}
        type="button"
        aria-haspopup="listbox"
        aria-expanded={isOpen}
        onClick={() => setIsOpen((value) => !value)}
      >
        <strong className="libraries-section-dropdown__value">{LSectionSummaryGet(props.selectedSections, props.sectionOptions)}</strong>
        <span className="libraries-section-dropdown__chevron" aria-hidden="true" />
      </button>
      {isOpen && (
        <div className="libraries-section-dropdown__menu" role="listbox" aria-multiselectable="true" aria-label={LLocaleTextGet("libraries.categoryFilter.ariaLabel")}>
          <button
            className={`libraries-section-dropdown__option ${props.selectedSections.length === 0 ? "libraries-section-dropdown__option--active" : ""}`}
            type="button"
            role="option"
            aria-selected={props.selectedSections.length === 0}
            onClick={() => props.onChangeSections([])}
          >
            <span>{LLocaleTextGet("libraries.categoryFilter.all")}</span>
            {props.selectedSections.length === 0
              ? <span className="libraries-section-dropdown__check" aria-hidden="true">✓</span>
              : <span className="libraries-section-dropdown__check-placeholder" aria-hidden="true" />
            }
          </button>
          {props.sectionOptions.map((section) => {
            const isSelected = props.selectedSections.includes(section.id);
            return (
              <button
                className={`libraries-section-dropdown__option ${isSelected ? "libraries-section-dropdown__option--active" : ""}`}
                type="button"
                role="option"
                aria-selected={isSelected}
                key={section.id}
                onClick={() => LSectionToggle(section.id)}
              >
                <span>{LSectionLabelGet(section.id, props.sectionOptions)}</span>
                {isSelected
                  ? <span className="libraries-section-dropdown__check" aria-hidden="true">✓</span>
                  : <span className="libraries-section-dropdown__check-placeholder" aria-hidden="true" />
                }
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

export function PToolbarLibraryRender(props: {
  showTechnicalDetails: boolean;
  onToggleTechnicalDetails: () => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LSectionState[];
  sectionOptions: LSectionOption[];
  onSectionFiltersChange: (value: LSectionState[]) => void;
  onResetFilters: () => void;
}) {
  return (
    <div className="libraries-toolbar">
      <button className="button libraries-technical-toggle" type="button" aria-expanded={props.showTechnicalDetails} onClick={props.onToggleTechnicalDetails}>
        <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
        {props.showTechnicalDetails ? LLocaleTextGet("libraries.technical.hide") : LLocaleTextGet("libraries.technical.show")}
      </button>
      <label className="libraries-search">
        <span className="visually-hidden">{LLocaleTextGet("libraries.search.label")}</span>
        <input value={props.searchQuery} onChange={(event) => props.onSearchQueryChange(event.target.value)} placeholder={LLocaleTextGet("libraries.search.placeholder")} />
      </label>
      <PDropdownSectionRender
        selectedSections={props.sectionFilters}
        sectionOptions={props.sectionOptions}
        onChangeSections={props.onSectionFiltersChange}
      />
      <button className="button libraries-reset" type="button" onClick={props.onResetFilters}>
        <img className="card__btn-icon" src={resetIcon} alt="" aria-hidden="true" />
        {LLocaleTextGet("libraries.filter.reset")}
      </button>
    </div>
  );
}
