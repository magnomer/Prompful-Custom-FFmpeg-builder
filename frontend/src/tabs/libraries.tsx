import React from "react";
import { PTextDescriptionRender, PHeaderPageRender } from "./shared";
import { LLocaleTextGet } from "../i18n";
import { LLibraryLicenseLabelGet, LLibraryTextGet } from "../catalogText";
import { LUnlockBasicCheck, LUnlockSudoCheck } from "../devUnlock";
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

// ─── Types ───────────────────────────────────────────────────────────────────

export type LPresetLibraryId = "minimal" | "default" | "efficiency" | "compatibility" | "editor" | "full" | "ai" | "streaming" | "maxtest" | "custom";

// Localized text (name/plain/technical) is resolved at render time via LLocaleTextGet() using
// presetId, so a locale switch updates the preset buttons. The structural library lists
// are supplied by the backend V5 preset catalog.
export type LPresetLibrary = {
  presetId: LPresetLibraryId;
  libraryIds: string[];
  extendedLibraryIds?: string[];
  hidden?: boolean;
  // dev presets are shown only when the hidden About-tab developer unlock is on.
  dev?: boolean;
};

// ─── Preset data ─────────────────────────────────────────────────────────────

export function LPresetLibraryListSanitize(presets: LPresetLibraryChoice[] | undefined): LPresetLibrary[] {
  return (presets ?? [])
    .filter((preset): preset is LPresetLibraryChoice & { presetId: LPresetLibraryId } => LPresetLibraryValidate(preset.presetId))
    .map((preset) => ({
      presetId: preset.presetId,
      libraryIds: Array.isArray(preset.libraryIds) ? preset.libraryIds : [],
      extendedLibraryIds: Array.isArray(preset.extendedLibraryIds) ? preset.extendedLibraryIds : undefined,
      hidden: Boolean(preset.hidden),
      dev: Boolean(preset.dev),
    }));
}

export function LPresetLibraryResolve(preset: LPresetLibrary, extendedLibraries: boolean): string[] {
  if (extendedLibraries && preset.extendedLibraryIds && preset.extendedLibraryIds.length > 0) return preset.extendedLibraryIds;
  return preset.libraryIds;
}

// LLibraryMaximumTestIdsGet returns every selectable library from the backend-resolved
// catalog. LLibrarySelectionNormalize then resolves mutually-exclusive groups and drops
// rows that the backend marked unavailable for the active FFmpeg version/profile.
export function LLibraryMaximumTestIdsGet(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName?: string): string[] {
  const candidateIds = LCatalogLibrarySource
    .map((library) => library.libraryId)
    .filter((libraryId) => {
      const library = LCatalogLibrarySource.find((item) => item.libraryId === libraryId);
      return library && !LLibraryUiUnavailableCheck(library, windowsShellProfileName ?? "");
    });
  return LLibrarySelectionNormalize(candidateIds, windowsShellProfileName, LCatalogLibrarySource);
}

// ─── Library selection utilities ─────────────────────────────────────────────

// TLS backends form a pick-one group in normal AND basic dev mode (rendered as radio
// buttons). Only the sudo dev tier relaxes this so every backend can be selected at once
// for build testing, so the mutual-exclusion pruning below is skipped only when sudo.
export const LLibraryTlsBackendIds = new Set(["openssl", "gnutls", "mbedtls", "libtls"]);

// shaderc/glslang are a pick-one shader-compiler group. They are not a radio group (zero
// may be selected), but they share the TLS group's shortened-divider visual so the two
// rows read as one "choose at most one" block.
export const LLibraryShaderCompilerIds = new Set(["shaderc", "glslang"]);

// xevd/xevdb and xeve/xeveb are EVC full-profile and baseline-profile bindings.
// FFmpeg configure rejects enabling both members of either pair, so they share the
// same pick-one radio + divider treatment with separate radio groups.
export const LLibraryEvcDecoderIds = new Set(["xevd", "xevdb"]);
export const LLibraryEvcEncoderIds = new Set(["xeve", "xeveb"]);

// libvpl (Intel oneVPL, --enable-libvpl) and libmfx (legacy Intel Media SDK, --enable-libmfx)
// are the two Intel Hardware Acceleration backends. FFmpeg configure rejects enabling both ("can
// not use libmfx and libvpl together"), so they are a pick-one radio group with the same divider
// treatment as the EVC pairs. They are adjacent in the library catalog so the two rows read as one block.
export const LLibraryIntelBackendIds = new Set(["libvpl", "libmfx"]);

export function LLibrarySelectionNormalize(selectedLibraryIds: string[], windowsShellProfileName?: string, LCatalogLibrarySource?: LLibraryChoice[]): string[] {
  const baseLibraryIds = LCatalogLibrarySource ? LCatalogLibrarySource.filter((library) => library.defaultChecked).map((library) => library.libraryId) : [];
  const selectedSet = new Set<string>([...baseLibraryIds, ...selectedLibraryIds]);
  // Only one TLS backend may be selected. Priority: openssl > gnutls > mbedtls > libtls.
  // Only the sudo dev tier keeps all selected backends so the TLS section can be tested
  // together; basic dev still enforces the normal pick-one rule.
  if (!LUnlockSudoCheck()) {
    if (selectedSet.has("openssl")) {
      selectedSet.delete("gnutls");
      selectedSet.delete("mbedtls");
      selectedSet.delete("libtls");
    } else if (selectedSet.has("gnutls")) {
      selectedSet.delete("mbedtls");
      selectedSet.delete("libtls");
    } else if (selectedSet.has("mbedtls")) {
      selectedSet.delete("libtls");
    }
  }
  if (selectedSet.has("shaderc") && selectedSet.has("glslang")) {
    selectedSet.delete("glslang");
  }
  // FFmpeg configure rejects enabling the full-profile and baseline-profile EVC bindings
  // together ("libxevd and libxevdb must not be enabled at the same time", same for the
  // encoder). Keep the full-profile binding, drop the baseline one.
  if (selectedSet.has("xevd") && selectedSet.has("xevdb")) {
    selectedSet.delete("xevdb");
  }
  if (selectedSet.has("xeve") && selectedSet.has("xeveb")) {
    selectedSet.delete("xeveb");
  }
  if (LCatalogLibrarySource) {
    const LCatalogLibrarySourceById = new Map(LCatalogLibrarySource.map((library) => [library.libraryId, library]));
    for (const libraryId of [...selectedSet]) {
      const library = LCatalogLibrarySourceById.get(libraryId);
      if (library && LLibraryUiUnavailableCheck(library, windowsShellProfileName ?? "")) {
        selectedSet.delete(libraryId);
      }
    }
  }
  // Intel Hardware Acceleration: oneVPL (libvpl / --enable-libvpl) and the legacy Media SDK
  // (libmfx) are mutually exclusive — FFmpeg configure dies if both are enabled. This runs AFTER
  // the version/profile pruning above on purpose: a preset can list both backends, and the right
  // one survives per version. On FFmpeg < 6.0 the libvpl row is version-unsupported and was just
  // pruned, so libmfx remains and gives those releases their only Intel HW accel path; on FFmpeg
  // 6.0+ both are available, so libmfx is dropped here in favour of the maintained oneVPL path
  // (matching FFmpeg's own "libmfx is deprecated, use libvpl" warning).
  if (selectedSet.has("libvpl") && selectedSet.has("libmfx")) {
    selectedSet.delete("libmfx");
  }
  return Array.from(selectedSet);
}

function LLibrarySetSameCheck(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((libraryId, index) => libraryId === b[index]);
}

function LPresetNormalizedIdsGet(preset: LPresetLibrary, windowsShellProfileName?: string, LCatalogLibrarySource?: LLibraryChoice[], extendedLibraries = false): string[] {
  return LLibrarySelectionNormalize(LPresetLibraryResolve(preset, extendedLibraries), windowsShellProfileName, LCatalogLibrarySource).slice().sort();
}

export function LPresetLibraryMatch(selectedLibraryIds: string[], presets: LPresetLibrary[], windowsShellProfileName?: string, LCatalogLibrarySource?: LLibraryChoice[], extendedLibraries = false, preferredPresetId?: LPresetLibraryId): LPresetLibraryId {
  const normalizedSelection = LLibrarySelectionNormalize(selectedLibraryIds, windowsShellProfileName, LCatalogLibrarySource).slice().sort();
  const preferredPreset = presets.find((preset) => preset.presetId === preferredPresetId && preset.presetId !== "custom");
  if (preferredPreset && !preferredPreset.hidden && !preferredPreset.dev && LLibrarySetSameCheck(normalizedSelection, LPresetNormalizedIdsGet(preferredPreset, windowsShellProfileName, LCatalogLibrarySource, extendedLibraries))) {
    return preferredPreset.presetId;
  }
  for (const preset of presets) {
    if (preset.presetId === "custom" || preset.hidden || preset.dev) continue;
    const normalizedPreset = LPresetNormalizedIdsGet(preset, windowsShellProfileName, LCatalogLibrarySource, extendedLibraries);
    if (LLibrarySetSameCheck(normalizedSelection, normalizedPreset)) {
      return preset.presetId;
    }
  }
  // Maximum test is library catalog-derived, so it can only be matched when the library catalog is
  // available. Checked last so a narrower named preset always wins.
  if (LCatalogLibrarySource && LUnlockBasicCheck()) {
    const maxTest = LLibraryMaximumTestIdsGet(LCatalogLibrarySource, windowsShellProfileName).slice().sort();
    if (LLibrarySetSameCheck(normalizedSelection, maxTest)) return "maxtest";
  }
  return "custom";
}

export function LPresetLibraryValidate(value: unknown): value is LPresetLibraryId {
  return value === "minimal" || value === "default" || value === "efficiency" || value === "compatibility" || value === "editor" || value === "full" || value === "ai" || value === "streaming" || value === "maxtest" || value === "custom";
}

// Presets whose displayed name is never prefixed with "Extended" even when the
// extended-libraries toggle is on. Only the broadening presets get the prefix.
const LPresetNonextendedIds = new Set<LPresetLibraryId>(["minimal", "default", "maxtest", "custom"]);

const LPresetLibraryIconMap: Partial<Record<LPresetLibraryId, string>> = {
  minimal: libraryMinimalIcon,
  default: libraryDefaultIcon,
  efficiency: libraryMaximumEfficiencyIcon,
  compatibility: libraryMaximumCompatibilityIcon,
  editor: libraryMaximumAudioVideoEditorIcon,
  full: libraryMaximumFullIcon,
  maxtest: libraryDevIcon,
};

// Resolves the localized preset name, prepending the "Extended" prefix when the
// extended-libraries toggle is on (descriptions are left unchanged).
function LPresetLibraryNameGet(presetId: LPresetLibraryId, extendedLibraries: boolean): string {
  const baseName = LLocaleTextGet(`libraries.presets.${presetId}.name`);
  if (!extendedLibraries || LPresetNonextendedIds.has(presetId)) return baseName;
  return LLocaleTextGet("libraries.extended.presetPrefix") + baseName;
}

function PIconPresetRender(props: { presetId: LPresetLibraryId; className?: string }) {
  const icon = LPresetLibraryIconMap[props.presetId];
  if (!icon) return null;
  return <img className={props.className ?? ""} src={icon} alt="" aria-hidden="true" />;
}

export function LLibraryExclusiveRemove(selectedLibraryIds: string[], selectedLibraryId: string): string[] {
  // shaderc/glslang and the EVC profile bindings stay mutually exclusive always (FFmpeg
  // configure rejects both). The TLS pick-one group is relaxed only under the sudo dev
  // tier for build testing; basic dev keeps it.
  const exclusiveGroups: Record<string, string[]> = {
    ...(LUnlockSudoCheck() ? {} : {
      openssl: ["gnutls", "mbedtls", "libtls"],
      gnutls: ["openssl", "mbedtls", "libtls"],
      mbedtls: ["openssl", "gnutls", "libtls"],
      libtls: ["openssl", "gnutls", "mbedtls"],
    }),
    shaderc: ["glslang"],
    glslang: ["shaderc"],
    // EVC full-profile vs baseline-profile bindings are mutually exclusive in FFmpeg configure.
    xevd: ["xevdb"],
    xevdb: ["xevd"],
    xeve: ["xeveb"],
    xeveb: ["xeve"],
    // Intel Hardware Acceleration backends: oneVPL (libvpl) and the legacy Media SDK (libmfx)
    // cannot both be enabled.
    libvpl: ["libmfx"],
    libmfx: ["libvpl"],
  };
  const conflicts = exclusiveGroups[selectedLibraryId] ?? [];
  if (conflicts.length === 0) return selectedLibraryIds;
  return selectedLibraryIds.filter((libraryId) => libraryId === selectedLibraryId || !conflicts.includes(libraryId));
}


export function LLicenseBoundaryDerive(selectedLibraryIds: string[], LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): string {
  const selectedLibraries = LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName).filter((library) => selectedLibraryIds.includes(library.libraryId));
  if (selectedLibraries.some((library) => library.licenseEffectName === "nonfree")) return "nonfree-local";
  if (selectedLibraries.some((library) => library.licenseEffectName === "gpl")) return "gpl-local";
  return "lgpl-local";
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

// Library availability is now driven by the backend resolved catalog. The frontend
// must not carry hardcoded unimplemented-library or profile-unavailable lists: those
// are version/profile facts in internal/catalogfacts/catalogdata/libraries/*.json.
function LLibraryUnavailableReasonCheck(library: LLibraryChoice, reason: string): boolean {
  return (library.unavailableReasons ?? []).includes(reason);
}

function LLibraryCurrentV4DisabledCheck(library: LLibraryChoice): boolean {
  return LLibraryUnavailableReasonCheck(library, "disabled-in-current-v4-ui") || library.supportState === "ui-disabled";
}

function LLibraryPreparationUnimplementedCheck(library: LLibraryChoice): boolean {
  return Boolean(library.preparationStatus?.required) && library.preparationStatus?.implemented !== true;
}

function LLibraryProfileAvailableCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  return !(library.unavailableProfiles ?? []).includes(windowsShellProfileName);
}

function LLibraryFFmpegSupportedCheck(library: LLibraryChoice): boolean {
  return library.versionCompatibility?.supported !== false;
}

// Available means FFmpeg has the switch for the chosen release AND the package this builder
// can supply can satisfy it on that release. It is false for a release-support-unavailable row
// (e.g. lensfun, whose package FFmpeg cannot use) even though the switch nominally exists, so
// it is a stricter gate than LLibraryFFmpegSupportedCheck. Both are per-FFmpeg-version,
// driven by the backend release-support manifest annotation, never a global list.
function LLibraryFFmpegAvailableCheck(library: LLibraryChoice): boolean {
  return library.versionCompatibility?.available !== false;
}

function LLibraryTrackLabelGet(trackName: string): string {
  return LLocaleTextGet(`libraries.row.track.${trackName || "native"}`);
}

function LLibraryVisibleFilter(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): LLibraryChoice[] {
  void windowsShellProfileName;
  return LCatalogLibrarySource;
}

function LLibraryUiUnavailableCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  return LLibraryCurrentV4DisabledCheck(library) || LLibraryPreparationUnimplementedCheck(library) || !LLibraryProfileAvailableCheck(library, windowsShellProfileName) || !LLibraryFFmpegSupportedCheck(library) || !LLibraryFFmpegAvailableCheck(library);
}

// Whether the checkbox is actually locked. Same as UI-unavailable, except the hidden
// About-tab developer unlock makes otherwise-unavailable libraries checkable for testing.
// The unavailable styling still shows regardless of unlock.
function LLibraryCheckboxLockedCheck(library: LLibraryChoice, windowsShellProfileName: string): boolean {
  if (!LLibraryFFmpegSupportedCheck(library)) return true;
  if (!LLibraryFFmpegAvailableCheck(library)) return true;
  return LLibraryUiUnavailableCheck(library, windowsShellProfileName) && !LUnlockBasicCheck();
}

// Returns the localization key suffix for an unavailable row's note, or "" when no
// note should be shown. The reason comes from the backend-resolved catalog state; the
// frontend does not own library-specific build-preparation facts.
function LLibraryUiUnavailableReasonKeyGet(library: LLibraryChoice, windowsShellProfileName: string): string {
  const libraryId = library.libraryId;
  if (!LLibraryFFmpegSupportedCheck(library)) return "ffmpegVersionUnsupported";
  if (!LLibraryProfileAvailableCheck(library, windowsShellProfileName)) return "profileUnavailable";
  if (LLibraryPreparationUnimplementedCheck(library)) return "preparationUnimplemented";
  if (LLibraryCurrentV4DisabledCheck(library)) return libraryId;
  if (!LLibraryFFmpegAvailableCheck(library)) return libraryId;
  return "";
}

function LLibraryCategoryGroup(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string) {
  return LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName).reduce<Record<string, LLibraryChoice[]>>((groups, library) => {
    const categoryName = LLibraryTextGet(library, "categoryName") || LLocaleTextGet("common.other");
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}


function LPresetLibrarySelectableList(presets: LPresetLibrary[]): Array<LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> }> {
  return presets.filter((preset): preset is LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> } =>
    preset.presetId !== "custom" && !preset.hidden && (!preset.dev || LUnlockBasicCheck())
  );
}

function PCardPresetRender(props: { presets: LPresetLibrary[]; selectedPresetId: LPresetLibraryId; onApplyPreset: (presetId: LPresetLibraryId) => void; extendedLibraries: boolean }) {
  const presets = LPresetLibrarySelectableList(props.presets);
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
              {LPresetLibraryNameGet(preset.presetId, props.extendedLibraries)}
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

function PCardLibraryRender(props: {
  LCatalogLibrarySource: LLibraryChoice[];
  selectedLibraryIds: string[];
  onToggleLibrary: (libraryId: string) => void;
  windowsShellProfileName: string;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LLibrarySectionFilter[];
  sectionOptions: string[];
  onSectionFiltersChange: (value: LLibrarySectionFilter[]) => void;
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

function PSelectorPresetRender(props: { presets: LPresetLibrary[]; selectedPresetId: LPresetLibraryId; onApplyPreset: (presetId: LPresetLibraryId) => void; showTechnicalDetails: boolean; extendedLibraries: boolean }) {
  const selectedPreset = props.presets.find((preset) => preset.presetId === props.selectedPresetId && preset.presetId !== "custom") as (LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> }) | undefined;

  return (
    <section className="preset-panel">
      <div className="preset-grid">
        {props.presets.filter((preset): preset is LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> } => LPresetLibrarySelectableList(props.presets).some((selectablePreset) => selectablePreset.presetId === preset.presetId)).map((preset) => (
          <button className={`preset-card ${preset.dev ? "preset-card--dev" : ""} ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <PIconPresetRender presetId={preset.presetId} className="preset-card__icon" />
            <span>
              <span className="preset-card__name">{LPresetLibraryNameGet(preset.presetId, props.extendedLibraries)}</span>
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
              <strong>{LPresetLibraryNameGet(selectedPreset.presetId, props.extendedLibraries)}</strong>
              <span>{LLocaleTextGet("libraries.technical.show")}</span>
            </span>
          </div>
          <p className="preset-technical-card__text">{LLocaleTextGet(`libraries.presets.${selectedPreset.presetId}.technical`)}</p>
        </section>
      )}
    </section>
  );
}

function PSummaryLibraryRender(props: { LCatalogLibrarySource: LLibraryChoice[]; selectedLibraryIds: string[]; selectedPresetId: LPresetLibraryId; windowsShellProfileName: string; extendedLibraries: boolean }) {
  const normalizedSelection = LLibrarySelectionNormalize(props.selectedLibraryIds, props.windowsShellProfileName, props.LCatalogLibrarySource);
  const visibleCatalog = LLibraryVisibleFilter(props.LCatalogLibrarySource, props.windowsShellProfileName);
  const licenseBoundary = LLicenseBoundaryDerive(normalizedSelection, props.LCatalogLibrarySource, props.windowsShellProfileName);
  const selectedOptionalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && !library.defaultChecked).length;
  const selectedInternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "internal").length;
  const selectedExternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "external").length;
  const includedCount = visibleCatalog.filter((library) => library.defaultChecked).length;

  const selectedPresetName = LPresetLibraryNameGet(props.selectedPresetId, props.extendedLibraries);
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

type LLibrarySectionFilter = string;

function LLibraryCategoryNameGet(library: LLibraryChoice): string {
  return LLibraryTextGet(library, "categoryName") || LLocaleTextGet("common.other");
}

function LLibraryCategoryDefaultCheck(categoryName: string): boolean {
  const normalizedCategoryName = categoryName.toLocaleLowerCase();
  return normalizedCategoryName.includes("included by default") || normalizedCategoryName.includes("기본 포함");
}

function LLibrarySectionFilterLabelGet(categoryName: LLibrarySectionFilter): string {
  if (LLibraryCategoryDefaultCheck(categoryName)) return LLocaleTextGet("libraries.categoryFilter.default");
  return categoryName;
}

function LLibrarySectionFilterSummaryGet(selectedSections: LLibrarySectionFilter[]): string {
  if (selectedSections.length === 0) return LLocaleTextGet("libraries.categoryFilter.all");
  if (selectedSections.length === 1) return LLibrarySectionFilterLabelGet(selectedSections[0]);
  return LLocaleTextGet("libraries.categoryFilter.selectedCount").replace("{count}", String(selectedSections.length));
}

function LLibrarySectionOptionsGet(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): string[] {
  const categoryNames: string[] = [];
  for (const library of LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName)) {
    const categoryName = LLibraryCategoryNameGet(library);
    if (!categoryNames.includes(categoryName)) categoryNames.push(categoryName);
  }
  return categoryNames;
}

function LLibraryFilter(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string, searchQuery: string, sectionFilters: LLibrarySectionFilter[]): LLibraryChoice[] {
  const normalizedQuery = searchQuery.trim().toLocaleLowerCase();
  return LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName).filter((library) => {
    const categoryName = LLibraryCategoryNameGet(library);
    if (sectionFilters.length > 0 && !sectionFilters.includes(categoryName)) return false;
    if (!normalizedQuery) return true;
    const searchableText = [
      library.libraryId,
      LLibraryTextGet(library, "displayName"),
      categoryName,
      LLibraryTextGet(library, "plainExplanation"),
      LLibraryTextGet(library, "technicalExplanation"),
      library.configureFlags.join(" "),
      library.packageNames.join(" "),
      library.officialWebpageUrl || "",
    ].join(" ").toLocaleLowerCase();
    return searchableText.includes(normalizedQuery);
  });
}


function LLibraryNameSplit(displayName: string): { buildName: string; featureName: string } {
  const separator = " / ";
  if (!displayName.includes(separator)) {
    return { buildName: displayName, featureName: "" };
  }
  const [buildName, ...featureParts] = displayName.split(separator);
  return { buildName: buildName.trim(), featureName: LLibraryFeatureNameClean(featureParts.join(separator).trim()) };
}

function LLibraryFeatureNameClean(featureName: string): string {
  const suffixes = [
    " encoding",
    " decoding",
    " support",
    " input",
    " output",
    " 인코딩",
    " 디코딩",
    " 지원",
    " 입력",
    " 출력",
  ];
  for (const suffix of suffixes) {
    if (featureName.toLocaleLowerCase().endsWith(suffix.toLocaleLowerCase())) {
      return featureName.slice(0, -suffix.length).trim();
    }
  }
  return featureName;
}

function LLibraryFilteredGroup(libraries: LLibraryChoice[]) {
  return libraries.reduce<Record<string, LLibraryChoice[]>>((groups, library) => {
    const categoryName = LLibraryCategoryNameGet(library);
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}

function PListLibraryRender(props: { LCatalogLibrarySource: LLibraryChoice[]; selectedLibraryIds: string[]; onToggleLibrary: (libraryId: string) => void; showTechnicalDetails: boolean; windowsShellProfileName: string; searchQuery: string; sectionFilters: LLibrarySectionFilter[]; onOpenOfficialWebpage: (url: string) => void }) {
  const filteredLibraries = LLibraryFilter(props.LCatalogLibrarySource, props.windowsShellProfileName, props.searchQuery, props.sectionFilters);
  const groupedLibraries = LLibraryFilteredGroup(filteredLibraries);
  return (
    <div className="library-list">
      {filteredLibraries.length === 0 && <p className="library-list__empty">{LLocaleTextGet("libraries.list.empty")}</p>}
      {Object.entries(groupedLibraries).map(([categoryName, libraries]) => (
        <section className="library-group" key={categoryName}>
          <h2 className="library-group__title">{categoryName}</h2>
          {libraries.map((library) => {
            const isUiUnavailable = LLibraryUiUnavailableCheck(library, props.windowsShellProfileName);
            const unavailableReasonKey = LLibraryUiUnavailableReasonKeyGet(library, props.windowsShellProfileName);
            const isCheckboxLocked = LLibraryCheckboxLockedCheck(library, props.windowsShellProfileName);
            const isChecked = (props.selectedLibraryIds.includes(library.libraryId) || library.defaultChecked) && !isCheckboxLocked;
            // TLS backends, shaderc/glslang, and the EVC binding pairs each render as a
            // radio group (pick one) in normal and basic dev mode; clicking the selected
            // radio clears it (FFmpeg allows zero). Only the sudo dev tier switches them
            // back to checkboxes so several can be selected for testing. Each group needs
            // its own radio `name` so unrelated groups do not clear each other.
            const radioGroupName = LUnlockSudoCheck()
              ? undefined
              : LLibraryTlsBackendIds.has(library.libraryId)
                ? "tls-backend"
                : LLibraryShaderCompilerIds.has(library.libraryId)
                  ? "shader-compiler"
                  : LLibraryEvcDecoderIds.has(library.libraryId)
                    ? "evc-decoder"
                    : LLibraryEvcEncoderIds.has(library.libraryId)
                      ? "evc-encoder"
                      : LLibraryIntelBackendIds.has(library.libraryId)
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
            const statusLabel = library.defaultChecked ? LLocaleTextGet("libraries.row.included") : LLibraryLicenseLabelGet(library.licenseEffectName);
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
                    {!isNativeTrack && <span className={`library-row__track library-row__track--${trackName}`}>{LLibraryTrackLabelGet(trackName)}</span>}
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



function PDropdownSectionRender(props: { selectedSections: LLibrarySectionFilter[]; sectionOptions: string[]; onChangeSections: (value: LLibrarySectionFilter[]) => void }) {
  const [isOpen, setIsOpen] = React.useState(false);
  const dropdownRef = React.useRef<HTMLDivElement>(null);

  React.useEffect(() => {
    if (!isOpen) return;
    function handlePointerDown(event: PointerEvent) {
      if (dropdownRef.current && !dropdownRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    document.addEventListener("pointerdown", handlePointerDown);
    return () => document.removeEventListener("pointerdown", handlePointerDown);
  }, [isOpen]);

  function toggleSection(sectionName: LLibrarySectionFilter) {
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
        <strong className="libraries-section-dropdown__value">{LLibrarySectionFilterSummaryGet(props.selectedSections)}</strong>
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
          {props.sectionOptions.map((sectionName) => {
            const isSelected = props.selectedSections.includes(sectionName);
            return (
              <button
                className={`libraries-section-dropdown__option ${isSelected ? "libraries-section-dropdown__option--active" : ""}`}
                type="button"
                role="option"
                aria-selected={isSelected}
                key={sectionName}
                onClick={() => toggleSection(sectionName)}
              >
                <span>{LLibrarySectionFilterLabelGet(sectionName)}</span>
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

function PToolbarLibraryRender(props: {
  showTechnicalDetails: boolean;
  onToggleTechnicalDetails: () => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LLibrarySectionFilter[];
  sectionOptions: string[];
  onSectionFiltersChange: (value: LLibrarySectionFilter[]) => void;
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

// ─── PLibraryRender ─────────────────────────────────────────────────────────────

export type PLibraryProps = {
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
  sectionFilters: LLibrarySectionFilter[];
  setSectionFilters: (value: LLibrarySectionFilter[]) => void;
  toggleLibrary: (libraryId: string) => void;
  applyLibraryPreset: (presetId: LPresetLibraryId) => void;
  setExtendedLibraries: (value: boolean) => void;
  openInUserBrowser: (url: string) => Promise<void>;
};

function LArrayEnsure<T>(value: T[] | null | undefined): T[] {
  return Array.isArray(value) ? value : [];
}

export function PLibraryRender({ initialProgramState, libraryCatalog, ffmpegBuildSettings, libraryPresetCatalog, libraryPresetId, extendedLibraries, libraryDetailedView, setLibraryDetailedView, showTechnicalDetails, setShowTechnicalDetails, sectionFilters, setSectionFilters, toggleLibrary, applyLibraryPreset, setExtendedLibraries, openInUserBrowser }: PLibraryProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const safeLibraryCatalog = LArrayEnsure(libraryCatalog);
  const safeInitialLibraryCatalog = LArrayEnsure(initialProgramState?.defaultLibraryCatalog);
  const safeLibraryPresetCatalog = LArrayEnsure(libraryPresetCatalog);
  const safeSelectedLibraryIds = LArrayEnsure(ffmpegBuildSettings?.selectedLibraryIds);
  const safeSectionFilters = LArrayEnsure(sectionFilters);
  const safeShellProfileName = ffmpegBuildSettings?.windowsShellProfileName ?? "ucrt64";
  const LCatalogLibrarySource = safeLibraryCatalog.length > 0 ? safeLibraryCatalog : safeInitialLibraryCatalog;
  const sectionOptions = LLibrarySectionOptionsGet(LCatalogLibrarySource, safeShellProfileName);

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
          <PCardPresetRender presets={safeLibraryPresetCatalog} selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} extendedLibraries={extendedLibraries} />
          <PCardLibraryRender
            LCatalogLibrarySource={LCatalogLibrarySource}
            selectedLibraryIds={safeSelectedLibraryIds}
            onToggleLibrary={toggleLibrary}
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
          <PSelectorPresetRender presets={safeLibraryPresetCatalog} selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} showTechnicalDetails={showTechnicalDetails} extendedLibraries={extendedLibraries} />
          <label className="library-extended-toggle">
            <input type="checkbox" checked={extendedLibraries} onChange={(event) => setExtendedLibraries(event.target.checked)} />
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
          <PListLibraryRender LCatalogLibrarySource={LCatalogLibrarySource} selectedLibraryIds={safeSelectedLibraryIds} onToggleLibrary={toggleLibrary} showTechnicalDetails={showTechnicalDetails} windowsShellProfileName={safeShellProfileName} searchQuery={searchQuery} sectionFilters={safeSectionFilters} onOpenOfficialWebpage={openInUserBrowser} />
        </>
      )}
    </section>
  );
}
