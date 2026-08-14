import React from "react";
import { PHeaderPageRender } from "./shared";
import { LLocaleTextGet } from "../i18n";
import { LSectionState, LSectionOptionsGet } from "./librarycatalog";
import { LPresetLibraryId, LPresetLibrary } from "./librarypresets";
import {
  PCardPresetRender,
  PCardLibraryRender,
  PSelectorPresetRender,
  PSummaryLibraryRender,
  PToolbarLibraryRender,
  PListLibraryRender,
} from "./librarycomponents";

// Preset and selection logic, and the catalog/section helpers, live in dedicated modules.
// They are re-exported here so existing importers of "./tabs/libraries" keep working.
export type { LPresetLibraryId, LPresetLibrary } from "./librarypresets";
export {
  LPresetLibraryClean,
  LPresetLibraryResolve,
  LLibrarySelectionNormalize,
  LPresetLibraryMatch,
  LPresetLibraryValidate,
  LLicenseBoundaryGet,
  LLibraryExclusiveRemove,
  LLibraryTestGet,
} from "./librarypresets";

// ─── PLibraryRender ─────────────────────────────────────────────────────────────

export type LLibraryProperties = {
  initialProgramState: LStateInitial;
  libraryCatalog: LLibraryChoice[];
  ffmpegBuildSettings: LSettingsFFmpeg;
  libraryPresetCatalog: LPresetLibrary[];
  libraryPresetId: LPresetLibraryId;
  extendedLibraries: boolean;
  libraryDetailedView: boolean;
  setLibraryDetailedView: (value: boolean) => void;
  showTechnicalDetails: boolean;
  setShowTechnicalDetails: (value: boolean) => void;
  sectionFilters: LSectionState[];
  setSectionFilters: (value: LSectionState[]) => void;
  LLibraryToggle: (libraryId: string) => void;
  LPresetLibraryApply: (presetId: LPresetLibraryId) => void;
  LLibraryExtendedUpdate: (value: boolean) => void;
  openInUserBrowser: (url: string) => Promise<void>;
};

function LArrayEnsure<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

export function PLibraryRender({ initialProgramState, libraryCatalog, ffmpegBuildSettings, libraryPresetCatalog, libraryPresetId, extendedLibraries, libraryDetailedView, setLibraryDetailedView, showTechnicalDetails, setShowTechnicalDetails, sectionFilters, setSectionFilters, LLibraryToggle, LPresetLibraryApply, LLibraryExtendedUpdate, openInUserBrowser }: LLibraryProperties) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const safeLibraryCatalog = LArrayEnsure(libraryCatalog);
  const safeInitialLibraryCatalog = LArrayEnsure(initialProgramState?.defaultLibraryCatalog);
  const safeLibraryPresetCatalog = LArrayEnsure(libraryPresetCatalog);
  const safeSelectedLibraryIds = LArrayEnsure(ffmpegBuildSettings?.selectedLibraryIds);
  const safeSectionFilters = LArrayEnsure(sectionFilters);
  const safeShellProfileName = ffmpegBuildSettings?.windowsShellProfileName ?? "ucrt64";
  const LCatalogLibrarySource = safeLibraryCatalog.length > 0 ? safeLibraryCatalog : safeInitialLibraryCatalog;
  const sectionOptions = LSectionOptionsGet(LCatalogLibrarySource, safeShellProfileName);

  React.useEffect(() => {
    const prunedSections = safeSectionFilters.filter((sectionName) => sectionOptions.includes(sectionName));
    if (prunedSections.length !== safeSectionFilters.length) setSectionFilters(prunedSections);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sectionOptions]);

  return (
    <section className="tab-page libraries-page">
      <div className="libraries-page__header-row">
        <PHeaderPageRender title={LLocaleTextGet("libraries.title")} text={LLocaleTextGet("libraries.intro")} />
        <label className="libraries-design-toggle">
          <input type="checkbox" checked={libraryDetailedView} onChange={(event) => setLibraryDetailedView(event.target.checked)} />
          <span className="libraries-design-toggle__text">{LLocaleTextGet("libraries.designToggle.label")}</span>
        </label>
      </div>

      {!libraryDetailedView && (
        <div className="libraries-simple-layout">
          <PCardPresetRender presets={safeLibraryPresetCatalog} selectedPresetId={libraryPresetId} onApplyPreset={LPresetLibraryApply} extendedLibraries={extendedLibraries} />
          <PCardLibraryRender
            LCatalogLibrarySource={LCatalogLibrarySource}
            selectedLibraryIds={safeSelectedLibraryIds}
            onToggleLibrary={LLibraryToggle}
            windowsShellProfileName={safeShellProfileName}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            sectionFilters={safeSectionFilters}
            sectionOptions={sectionOptions}
            onSectionFiltersChange={setSectionFilters}
            onOpenOfficialWebpage={openInUserBrowser}
          />
        </div>
      )}

      {libraryDetailedView && (
        <>
          <PSelectorPresetRender presets={safeLibraryPresetCatalog} selectedPresetId={libraryPresetId} onApplyPreset={LPresetLibraryApply} showTechnicalDetails={showTechnicalDetails} extendedLibraries={extendedLibraries} />
          <label className="library-extended-toggle">
            <input type="checkbox" checked={extendedLibraries} onChange={(event) => LLibraryExtendedUpdate(event.target.checked)} />
            <span className="library-extended-toggle__label">{LLocaleTextGet("libraries.extended.label")}</span>
          </label>
          <PSummaryLibraryRender LCatalogLibrarySource={LCatalogLibrarySource} selectedLibraryIds={safeSelectedLibraryIds} selectedPresetId={libraryPresetId} windowsShellProfileName={safeShellProfileName} extendedLibraries={extendedLibraries} />
          <PToolbarLibraryRender
            showTechnicalDetails={showTechnicalDetails}
            onToggleTechnicalDetails={() => setShowTechnicalDetails(!showTechnicalDetails)}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            sectionFilters={safeSectionFilters}
            sectionOptions={sectionOptions}
            onSectionFiltersChange={setSectionFilters}
            onResetFilters={() => {
              setSearchQuery("");
              setSectionFilters([]);
            }}
          />
          {showTechnicalDetails && (
            <section className="libraries-technical-panel">
              <h2 className="libraries-technical-panel__title">{LLocaleTextGet("libraries.technical.title")}</h2>
              <div className="libraries-technical-details">
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{LLocaleTextGet("libraries.technical.builtIn.title")}</h3>
                  <p className="libraries-technical-detail__text">{LLocaleTextGet("libraries.technical.builtIn.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{LLocaleTextGet("libraries.technical.sourceBuild.title")}</h3>
                  <p className="libraries-technical-detail__text">{LLocaleTextGet("libraries.technical.sourceBuild.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{LLocaleTextGet("libraries.technical.customBuild.title")}</h3>
                  <p className="libraries-technical-detail__text">{LLocaleTextGet("libraries.technical.customBuild.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{LLocaleTextGet("libraries.technical.license.title")}</h3>
                  <p className="libraries-technical-detail__text">{LLocaleTextGet("libraries.technical.license.text")}</p>
                </section>
              </div>
            </section>
          )}
          <PListLibraryRender LCatalogLibrarySource={LCatalogLibrarySource} selectedLibraryIds={safeSelectedLibraryIds} onToggleLibrary={LLibraryToggle} showTechnicalDetails={showTechnicalDetails} windowsShellProfileName={safeShellProfileName} searchQuery={searchQuery} sectionFilters={safeSectionFilters} onOpenOfficialWebpage={openInUserBrowser} />
        </>
      )}
    </section>
  );
}
