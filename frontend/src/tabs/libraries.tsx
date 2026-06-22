import React from "react";
import { PageHeader } from "./shared";
import { t } from "../i18n";
import { libraryLicenseLabel, libraryText } from "../catalogText";

// ─── Types ───────────────────────────────────────────────────────────────────

export type LibraryPresetId = "minimal" | "default" | "efficiency" | "compatibility" | "editor" | "full" | "custom";

// Localized text (name/plain/technical) is resolved at render time via t() using
// presetId, so a locale switch updates the preset buttons. Only the structural
// libraryIds live here — they are locale-independent.
export type LibraryPreset = {
  presetId: LibraryPresetId;
  libraryIds: string[];
};

// ─── Preset data ─────────────────────────────────────────────────────────────

export const baseIncludedLibraryIds = [
  "ffmpeg-program",
  "ffprobe-program",
  "libavcodec",
  "libavformat",
  "libavfilter",
  "libavutil",
  "libswscale",
  "libswresample",
  "native-codecs",
  "native-formats",
];

export const defaultPresetLibraryIds = [
  "nvenc", "amf", "qsv",
  "x264", "x265", "libvpx", "aom", "svt-av1", "dav1d", "theora", "xvid",
  "opus", "vorbis", "mp3lame", "gsm", "speex", "opencore-amr", "vo-amrwbenc", "rubberband",
  "openjpeg", "webp", "freetype", "fontconfig", "fribidi", "harfbuzz", "ass",
  "zimg", "vmaf", "vidstab", "srt", "ssh", "zmq", "openal", "sdl2", "gme", "openmpt",
];

// Preset tiers are nested: each higher tier adds to the one below it.
// Efficiency: best-quality encode/resample. Compatibility: broadest codec,
// caption, and protocol I/O. Editor: filtering, audio plugins, color, and
// subtitle/transcription tooling. Full: packageable advanced features. Prefer
// shaderc over glslang because FFmpeg configure rejects selecting both together.
// lensfun and SVT JPEG XS are intentionally hidden from the UI and automatic presets for now.
// Backend support remains for future compatibility: old saved/manual requests are
// checked before configure, and incompatible flags are skipped with a warning.
// lensfun is hidden because current MSYS2 lensfun exposes an older API than the
// FFmpeg lensfun filter source expects.
// SVT JPEG XS is hidden because current package/repository states may not satisfy
// FFmpeg's SvtJpegxs requirement.
const efficiencyExtraLibraryIds = ["fdk-aac", "soxr"];
const compatibilityExtraLibraryIds = [...efficiencyExtraLibraryIds, "rav1e", "openh264", "ilbc", "twolame", "xevd", "shine", "codec2", "lc3", "snappy", "rsvg", "zvbi", "aribb24", "aribcaption", "rtmp"];
const editorExtraLibraryIds = [
  ...compatibilityExtraLibraryIds,
  "libjxl", "png", "libplacebo", "frei0r", "xml2",
  "mysofa", "bs2b", "lcms2",
  "shaderc",
  "cairo", "opencv", "opencolorio",
  "ladspa", "lv2", "chromaprint", "qrencode", "whisper",
];
const fullExtraLibraryIds = [
  ...editorExtraLibraryIds,
  "bluray", "dvdread", "cdio",
  "modplug",
  "openssl", "rist", "rabbitmq",
  "tesseract",
  "jack", "pulse",
  "caca", "opencl",
];

export const libraryPresets: LibraryPreset[] = [
  { presetId: "minimal", libraryIds: baseIncludedLibraryIds },
  { presetId: "default", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds] },
  { presetId: "efficiency", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...efficiencyExtraLibraryIds] },
  { presetId: "compatibility", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...compatibilityExtraLibraryIds] },
  { presetId: "editor", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...editorExtraLibraryIds] },
  { presetId: "full", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...fullExtraLibraryIds] },
];

// ─── Library selection utilities ─────────────────────────────────────────────

export function normalizeLibrarySelection(selectedLibraryIds: string[]): string[] {
  const selectedSet = new Set<string>([...baseIncludedLibraryIds, ...selectedLibraryIds]);
  if (selectedSet.has("openssl") && selectedSet.has("gnutls")) {
    selectedSet.delete("gnutls");
  }
  if (selectedSet.has("shaderc") && selectedSet.has("glslang")) {
    selectedSet.delete("glslang");
  }
  return Array.from(selectedSet);
}

export function matchLibraryPresetId(selectedLibraryIds: string[]): LibraryPresetId {
  const normalizedSelection = normalizeLibrarySelection(selectedLibraryIds).slice().sort();
  for (const preset of libraryPresets) {
    if (preset.presetId === "custom") continue;
    const normalizedPreset = normalizeLibrarySelection(preset.libraryIds).slice().sort();
    if (normalizedSelection.length === normalizedPreset.length && normalizedSelection.every((libraryId, index) => libraryId === normalizedPreset[index])) {
      return preset.presetId;
    }
  }
  return "custom";
}

export function isValidLibraryPresetId(value: unknown): value is LibraryPresetId {
  return value === "minimal" || value === "default" || value === "efficiency" || value === "compatibility" || value === "editor" || value === "full" || value === "custom";
}

export function removeMutuallyExclusiveLibraries(selectedLibraryIds: string[], selectedLibraryId: string): string[] {
  const exclusiveGroups: Record<string, string[]> = {
    openssl: ["gnutls"],
    gnutls: ["openssl"],
    shaderc: ["glslang"],
    glslang: ["shaderc"],
  };
  const conflicts = exclusiveGroups[selectedLibraryId] ?? [];
  if (conflicts.length === 0) return selectedLibraryIds;
  return selectedLibraryIds.filter((libraryId) => libraryId === selectedLibraryId || !conflicts.includes(libraryId));
}


export function deriveLicenseBoundaryFromSelectedLibraries(selectedLibraryIds: string[], catalog: LibraryChoice[]): string {
  const selectedLibraries = visibleLibraries(catalog).filter((library) => selectedLibraryIds.includes(library.libraryId));
  if (selectedLibraries.some((library) => library.licenseEffectName === "nonfree")) return "nonfree-local";
  if (selectedLibraries.some((library) => library.licenseEffectName === "gpl")) return "gpl-local";
  return "lgpl-local";
}

// ─── Library UI components ────────────────────────────────────────────────────

// Renders a catalog note with minimal markup: "\n" becomes a line break and
// "**text**" becomes bold. No other markdown is interpreted.
function renderRichNote(text: string): React.ReactNode {
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

const hiddenLibraryIds = new Set(["lensfun", "svtjpegxs"]);

// Libraries with no prebuilt MSYS2 package: configure fails unless the user
// builds and installs them manually. Flagged with a pastel-red row background.
const unbuildableLibraryIds = new Set(["vvenc", "xavs2"]);

function visibleLibraries(catalog: LibraryChoice[]): LibraryChoice[] {
  return catalog.filter((library) => !hiddenLibraryIds.has(library.libraryId));
}

function groupLibrariesByCategory(catalog: LibraryChoice[]) {
  return visibleLibraries(catalog).reduce<Record<string, LibraryChoice[]>>((groups, library) => {
    const categoryName = libraryText(library, "categoryName") || t("common.other");
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}

function LibraryPresetSelector(props: { presets: LibraryPreset[]; selectedPresetId: LibraryPresetId; onApplyPreset: (presetId: LibraryPresetId) => void; showTechnicalDetails: boolean }) {
  return (
    <section className="preset-panel">
      <div className="preset-panel__header">
        <h2 className="preset-panel__title">{t("libraries.presetSelector.title")}</h2>
        <p className="preset-panel__text">{t("libraries.presetSelector.text")}</p>
      </div>
      <div className="preset-grid">
        {props.presets.filter((preset) => preset.presetId !== "custom").map((preset) => (
          <button className={`preset-card ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <span className="preset-card__name">{t(`libraries.presets.${preset.presetId}.name`)}</span>
            <span className="preset-card__plain">{t(`libraries.presets.${preset.presetId}.plain`)}</span>
            {props.showTechnicalDetails && <span className="preset-card__technical">{t(`libraries.presets.${preset.presetId}.technical`)}</span>}
          </button>
        ))}
      </div>
      {props.selectedPresetId === "custom" && <p className="preset-panel__custom">{t("libraries.presetSelector.custom")}</p>}
    </section>
  );
}

function LibrarySelectionSummary(props: { catalog: LibraryChoice[]; selectedLibraryIds: string[]; selectedPresetId: LibraryPresetId }) {
  const normalizedSelection = normalizeLibrarySelection(props.selectedLibraryIds);
  const licenseBoundary = deriveLicenseBoundaryFromSelectedLibraries(normalizedSelection, props.catalog);
  const selectedExternalCount = visibleLibraries(props.catalog).filter((library) => normalizedSelection.includes(library.libraryId) && !library.defaultChecked).length;
  const includedCount = visibleLibraries(props.catalog).filter((library) => library.defaultChecked).length;

  return (
    <section className="library-summary" aria-label={t("libraries.summary.ariaLabel")}>
      <div className="library-summary__item">
        <span className="library-summary__label">{t("libraries.summary.preset")}</span>
        <strong className="library-summary__value">{t(`libraries.presets.${props.selectedPresetId}.name`)}</strong>
      </div>
      <div className="library-summary__item">
        <span className="library-summary__label">{t("libraries.summary.external")}</span>
        <strong className="library-summary__value">{selectedExternalCount}</strong>
      </div>
      <div className="library-summary__item">
        <span className="library-summary__label">{t("libraries.summary.included")}</span>
        <strong className="library-summary__value">{includedCount}</strong>
      </div>
      <div className="library-summary__item">
        <span className="library-summary__label">{t("libraries.summary.license")}</span>
        <strong className={`library-summary__value library-summary__license library-summary__license--${licenseBoundary}`}>{t(`libraries.summary.license.${licenseBoundary}`)}</strong>
      </div>
    </section>
  );
}

function LibraryList(props: { catalog: LibraryChoice[]; selectedLibraryIds: string[]; onToggleLibrary: (libraryId: string) => void; showTechnicalDetails: boolean }) {
  const groupedLibraries = groupLibrariesByCategory(props.catalog);
  return (
    <div className="library-list">
      {Object.entries(groupedLibraries).map(([categoryName, libraries]) => (
        <section className="library-group" key={categoryName}>
          <h2 className="library-group__title">{categoryName}</h2>
          {libraries.map((library) => {
            const isChecked = props.selectedLibraryIds.includes(library.libraryId) || library.defaultChecked;
            const isExternal = !library.defaultChecked;
            return (
              <label className={`library-row ${library.locked ? "library-row--locked" : ""} ${unbuildableLibraryIds.has(library.libraryId) ? "library-row--unbuildable" : ""}`} key={library.libraryId}>
                <input type="checkbox" checked={isChecked} disabled={library.locked} onChange={() => props.onToggleLibrary(library.libraryId)} />
                <span className="library-row__main">
                  <span className="library-row__name">{libraryText(library, "displayName")}</span>
                  <span className="library-row__note">{renderRichNote(libraryText(library, "plainExplanation") || libraryText(library, "reviewNote"))}</span>
                  {props.showTechnicalDetails && libraryText(library, "technicalExplanation") && <span className={`library-row__detail ${isExternal ? "library-row__detail--why" : ""}`}>{libraryText(library, "technicalExplanation")}</span>}
                  {props.showTechnicalDetails && library.configureFlags.length > 0 && <span className="library-row__detail"><strong>{t("libraries.row.flagsLabel")}</strong> {library.configureFlags.join(" ")}</span>}
                  {props.showTechnicalDetails && library.packageNames.length > 0 && <span className="library-row__detail"><strong>{t("libraries.row.packagesLabel")}</strong> {library.packageNames.join(", ")}</span>}
                </span>
                <span className={`library-row__license library-row__license--${library.licenseEffectName}`}>{library.defaultChecked ? t("libraries.row.included") : libraryLicenseLabel(library.licenseEffectName)}</span>
              </label>
            );
          })}
        </section>
      ))}
    </div>
  );
}

// ─── LibrariesTab ─────────────────────────────────────────────────────────────

export type LibrariesTabProps = {
  initialApplicationState: InitialApplicationState;
  ffmpegBuildSettings: FfmpegBuildSettings;
  libraryPresetId: LibraryPresetId;
  toggleLibrary: (libraryId: string) => void;
  applyLibraryPreset: (presetId: LibraryPresetId) => void;
};

export function LibrariesTab({ initialApplicationState, ffmpegBuildSettings, libraryPresetId, toggleLibrary, applyLibraryPreset }: LibrariesTabProps) {
  const [showTechnicalDetails, setShowTechnicalDetails] = React.useState(false);

  return (
    <section className="tab-page libraries-page">
      <PageHeader title={t("libraries.title")} text={t("libraries.intro")} />
      <section className="libraries-briefing">
        <p>{t("libraries.info.builtIn")}</p>
        <p>{t("libraries.info.external")}</p>
      </section>
      <LibraryPresetSelector presets={libraryPresets} selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} showTechnicalDetails={showTechnicalDetails} />
      <LibrarySelectionSummary catalog={initialApplicationState.defaultLibraryCatalog} selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds} selectedPresetId={libraryPresetId} />
      <div className="libraries-toolbar">
        <button className="button libraries-technical-toggle" type="button" aria-expanded={showTechnicalDetails} onClick={() => setShowTechnicalDetails((value) => !value)}>
          {showTechnicalDetails ? t("libraries.technical.hide") : t("libraries.technical.show")}
        </button>
      </div>
      {showTechnicalDetails && (
        <section className="libraries-technical-panel">
          <h2 className="libraries-technical-panel__title">{t("libraries.technical.title")}</h2>
          <div className="libraries-technical-details">
            <section className="libraries-technical-detail">
              <h3 className="libraries-technical-detail__title">{t("libraries.technical.builtIn.title")}</h3>
              <p className="libraries-technical-detail__text">{t("libraries.technical.builtIn.text")}</p>
            </section>
            <section className="libraries-technical-detail">
              <h3 className="libraries-technical-detail__title">{t("libraries.technical.external.title")}</h3>
              <p className="libraries-technical-detail__text">{t("libraries.technical.external.text")}</p>
            </section>
            <section className="libraries-technical-detail">
              <h3 className="libraries-technical-detail__title">{t("libraries.technical.configure.title")}</h3>
              <p className="libraries-technical-detail__text">{t("libraries.technical.configure.text")}</p>
            </section>
            <section className="libraries-technical-detail">
              <h3 className="libraries-technical-detail__title">{t("libraries.technical.license.title")}</h3>
              <p className="libraries-technical-detail__text">{t("libraries.technical.license.text")}</p>
            </section>
          </div>
        </section>
      )}
      <LibraryList catalog={initialApplicationState.defaultLibraryCatalog} selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds} onToggleLibrary={toggleLibrary} showTechnicalDetails={showTechnicalDetails} />
    </section>
  );
}
