import React from "react";
import { DescriptionLines, PageHeader } from "./shared";
import { t } from "../i18n";
import { libraryLicenseLabel, libraryText } from "../catalogText";
import { isDevUnlockEnabled, isSudoDevUnlockEnabled } from "../devUnlock";
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

export type LibraryPresetId = "minimal" | "default" | "efficiency" | "compatibility" | "editor" | "full" | "ai" | "streaming" | "maxtest" | "custom";

// Localized text (name/plain/technical) is resolved at render time via t() using
// presetId, so a locale switch updates the preset buttons. Only the structural
// libraryIds live here — they are locale-independent.
export type LibraryPreset = {
  presetId: LibraryPresetId;
  libraryIds: string[];
  hidden?: boolean;
  // dev presets are shown only when the hidden About-tab developer unlock is on, and
  // always render last. Their libraryIds are computed from the live catalog (see
  // maximumTestLibraryIds), so the static libraryIds array is left empty.
  dev?: boolean;
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
  // Hardware encoders: NVIDIA (nvenc), AMD (amf), and Intel Hardware Acceleration (Quick Sync).
  // Intel ships two version-split backends — libvpl (oneVPL, FFmpeg 6.0+) and libmfx (legacy
  // Media SDK, the only path before 6.0). Both are listed so the Intel HW accel capability stays
  // in every preset like the other vendors regardless of FFmpeg version: they are mutually
  // exclusive, and normalizeLibrarySelection keeps exactly one per version (libvpl on 6.0+,
  // libmfx — source-built, like vmaf on 4.4 — on older releases where libvpl is version-unsupported).
  "nvenc", "amf", "libvpl", "libmfx",
  "x264", "x265", "libvpx", "aom", "svt-av1", "dav1d", "theora", "xvid",
  "opus", "vorbis", "mp3lame", "gsm", "speex", "opencore-amr", "vo-amrwbenc", "rubberband",
  "openjpeg", "webp", "freetype", "fontconfig", "fribidi", "harfbuzz", "ass", "cairo",
  "zimg", "vmaf", "vidstab", "srt", "ssh", "zmq", "openal", "sdl2", "gme", "openmpt",
];

// Preset tiers are independent: each public broadening preset is Default + its own
// additions and does NOT inherit the others. Full is the union of all three plus
// its own broadest-only extras, so it stays a true superset. Hidden focused presets
// (ai/streaming) are likewise Default + their own focused additions only.
// fdk-aac is nonfree, so it is kept to Efficiency + Full; Compatibility and Editor
// stay license-free (LGPL/GPL). Prefer shaderc over glslang because FFmpeg configure
// rejects selecting both together (normalizeLibrarySelection drops glslang if both
// appear). The EVC pick-one pairs (xevd/xevdb, xeve/xeveb) are similarly pruned, so
// presets list only the full-profile member (xevd, xeve).
// lensfun, SVT JPEG XS, and VapourSynth are kept out of the automatic presets. Their
// run-time eligibility is decided per FFmpeg release, not by a global rule:
//   - lensfun is marked unavailable on every release line in the support manifest (FFmpeg
//     gates it by the lf_db_create symbol the available lensfun package lacks), so the
//     backend annotation hides it for every version.
//   - SVT JPEG XS stays selectable and is gated by its per-release pkg-config floor at build
//     time (the package must reach FFmpeg's required SvtJpegxs version for the chosen release).
//   - VapourSynth is build-capability disabled here (uiDisabledLibraryIds) because its runtime
//     is non-portable (needs a Python + VapourSynth runtime) regardless of FFmpeg version.
// Backend support remains in all cases: old saved/manual requests are checked before configure
// and incompatible flags are skipped with a warning.
// Efficiency: enhance Default's compression/quality per bit. fdk-aac (best AAC,
// nonfree), soxr (HQ resampler), rav1e (quality-focused AV1 encoder).
const efficiencyExtraLibraryIds = ["fdk-aac", "soxr", "rav1e"];
// Compatibility: read/write a wider range of formats. Free license only (no
// fdk-aac). Niche encoders/decoders: OpenH264, EVC (xeve/xevd full-profile), APV
// (oapv), AVS1 (xavs), speech (ilbc, codec2, lc3), MP2 (twolame), Shine MP3,
// images (snappy, rsvg), broadcast text (zvbi, aribb24, aribcaption), RTMP input.
const compatibilityExtraLibraryIds = [
  "openh264", "xeve", "xevd", "oapv", "xavs",
  "ilbc", "twolame", "shine", "codec2", "lc3",
  "snappy", "rsvg",
  "zvbi", "aribb24", "aribcaption",
  "rtmp",
];
// Editor: filters, color management, image formats, plugin hosting, analysis, and
// transcription useful to audio/video editors. Free license only.
const editorExtraLibraryIds = [
  "png", "libjxl", "lcms2",
  "libplacebo", "shaderc", "frei0r", "opencv", "opencolorio",
  "xml2", "mysofa", "bs2b",
  "ladspa", "lv2", "chromaprint", "qrencode", "whisper",
];
// Full: union of the three broadening presets plus the broadest-only extras (discs,
// devices, networks/messaging, TLS, OCR, GPU/CL, special outputs). One TLS backend
// (openssl) and one shader compiler (shaderc, from Editor) are picked; conflicting
// members are pruned by normalizeLibrarySelection.
const fullExtraLibraryIds = [
  ...efficiencyExtraLibraryIds,
  ...compatibilityExtraLibraryIds,
  ...editorExtraLibraryIds,
  "kvazaar",
  "bluray", "dvdread", "dvdnav", "cdio",
  "modplug",
  "opengl",
  "openssl", "rist", "rabbitmq",
  "tesseract",
  "jack", "pulse",
  "caca", "opencl",
];

const aiExtraLibraryIds = [
  "onnxruntime", "openvino", "torch", "tensorflow",
  "whisper", "tesseract", "opencv",
  "libplacebo", "opencl", "png", "libjxl",
];

const streamingExtraLibraryIds = [
  "fdk-aac", "soxr",
  "rtmp", "rist", "gnutls",
  "libplacebo", "opencl",
];

export const libraryPresets: LibraryPreset[] = [
  { presetId: "minimal", libraryIds: baseIncludedLibraryIds },
  { presetId: "default", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds] },
  { presetId: "efficiency", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...efficiencyExtraLibraryIds] },
  { presetId: "compatibility", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...compatibilityExtraLibraryIds] },
  { presetId: "editor", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...editorExtraLibraryIds] },
  { presetId: "full", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...fullExtraLibraryIds] },
  { presetId: "ai", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...aiExtraLibraryIds], hidden: true },
  { presetId: "streaming", libraryIds: [...baseIncludedLibraryIds, ...defaultPresetLibraryIds, ...streamingExtraLibraryIds], hidden: true },
  { presetId: "maxtest", libraryIds: [], dev: true },
];

// Source-build (Internal track) libraries added to each broadening preset when the
// "Extended" toggle is on. These have no prebuilt MSYS2 package, so they are prepared
// from source. Extended presets may therefore land on a stricter license boundary than
// their base preset (see per-id license effect in the catalog):
//   - vvenc/uavs3d/lcevc-dec: LGPL-safe
//   - xavs2/davs2/avisynthplus: flip the build to GPL
//   - mpeghdec: flips the build to nonfree (Full only, which already pulls openssl)
// libmfx is NOT listed here: it is part of the Intel Hardware Acceleration capability and lives in
// defaultPresetLibraryIds next to libvpl (see the note there), so it is in every preset like the
// other hardware encoders rather than gated behind the Extended toggle.
const extendedPresetExtraLibraryIds: Partial<Record<LibraryPresetId, string[]>> = {
  efficiency: ["vvenc", "lcevc-dec"],
  compatibility: ["davs2", "uavs3d", "xavs2", "avisynthplus", "klvanc"],
  editor: ["avisynthplus", "lcevc-dec", "quirc"],
  full: ["vvenc", "lcevc-dec", "davs2", "uavs3d", "xavs2", "avisynthplus", "mpeghdec", "quirc", "klvanc"],
};

// Effective library ids for a preset. With the Extended toggle on, the broadening
// presets gain their source-build extras; Minimal/Default/dev presets are unchanged.
export function presetLibraryIds(preset: LibraryPreset, extendedLibraries: boolean): string[] {
  if (!extendedLibraries) return preset.libraryIds;
  return [...preset.libraryIds, ...(extendedPresetExtraLibraryIds[preset.presetId] ?? [])];
}

// maximumTestLibraryIds returns every catalog library except those with no implemented
// build recipe (unimplementedBuildLibraryIds) and the build-capability-disabled UI rows
// (uiDisabledLibraryIds, e.g. tensorflow/vapoursynth). normalizeLibrarySelection
// then resolves the mutually-exclusive groups and drops any library unavailable for the
// active profile, so the result is a buildable superset.
export function maximumTestLibraryIds(catalog: LibraryChoice[], windowsShellProfileName?: string): string[] {
  const candidateIds = catalog
    .map((library) => library.libraryId)
    .filter((libraryId) => {
      const library = catalog.find((item) => item.libraryId === libraryId);
      return library && !isLibraryUiUnavailable(library, windowsShellProfileName ?? "");
    });
  return normalizeLibrarySelection(candidateIds, windowsShellProfileName, catalog);
}

// ─── Library selection utilities ─────────────────────────────────────────────

// TLS backends form a pick-one group in normal AND basic dev mode (rendered as radio
// buttons). Only the sudo dev tier relaxes this so every backend can be selected at once
// for build testing, so the mutual-exclusion pruning below is skipped only when sudo.
export const tlsBackendLibraryIds = new Set(["openssl", "gnutls", "mbedtls", "libtls"]);

// shaderc/glslang are a pick-one shader-compiler group. They are not a radio group (zero
// may be selected), but they share the TLS group's shortened-divider visual so the two
// rows read as one "choose at most one" block.
export const shaderCompilerLibraryIds = new Set(["shaderc", "glslang"]);

// xevd/xevdb and xeve/xeveb are EVC full-profile and baseline-profile bindings.
// FFmpeg configure rejects enabling both members of either pair, so they share the
// same pick-one radio + divider treatment with separate radio groups.
export const evcDecoderLibraryIds = new Set(["xevd", "xevdb"]);
export const evcEncoderLibraryIds = new Set(["xeve", "xeveb"]);

// libvpl (Intel oneVPL, --enable-libvpl) and libmfx (legacy Intel Media SDK, --enable-libmfx)
// are the two Intel Hardware Acceleration backends. FFmpeg configure rejects enabling both ("can
// not use libmfx and libvpl together"), so they are a pick-one radio group with the same divider
// treatment as the EVC pairs. They are adjacent in the catalog so the two rows read as one block.
export const intelHwaccelBackendLibraryIds = new Set(["libvpl", "libmfx"]);

export function normalizeLibrarySelection(selectedLibraryIds: string[], windowsShellProfileName?: string, catalog?: LibraryChoice[]): string[] {
  const selectedSet = new Set<string>([...baseIncludedLibraryIds, ...selectedLibraryIds]);
  // Only one TLS backend may be selected. Priority: openssl > gnutls > mbedtls > libtls.
  // Only the sudo dev tier keeps all selected backends so the TLS section can be tested
  // together; basic dev still enforces the normal pick-one rule.
  if (!isSudoDevUnlockEnabled()) {
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
  // Drop libraries that have no package for the active profile (e.g. onnxruntime
  // on mingw64), so the selection always matches what the profile can build.
  if (windowsShellProfileName) {
    for (const libraryId of [...selectedSet]) {
      if (!isLibraryAvailableForProfile(libraryId, windowsShellProfileName)) {
        selectedSet.delete(libraryId);
      }
    }
  }
  if (catalog) {
    const catalogById = new Map(catalog.map((library) => [library.libraryId, library]));
    for (const libraryId of [...selectedSet]) {
      const library = catalogById.get(libraryId);
      if (library && isLibraryUiUnavailable(library, windowsShellProfileName ?? "")) {
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

function sameLibrarySet(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((libraryId, index) => libraryId === b[index]);
}

function normalizedPresetIds(preset: LibraryPreset, windowsShellProfileName?: string, catalog?: LibraryChoice[], extendedLibraries = false): string[] {
  return normalizeLibrarySelection(presetLibraryIds(preset, extendedLibraries), windowsShellProfileName, catalog).slice().sort();
}

export function matchLibraryPresetId(selectedLibraryIds: string[], windowsShellProfileName?: string, catalog?: LibraryChoice[], extendedLibraries = false, preferredPresetId?: LibraryPresetId): LibraryPresetId {
  const normalizedSelection = normalizeLibrarySelection(selectedLibraryIds, windowsShellProfileName, catalog).slice().sort();
  const preferredPreset = libraryPresets.find((preset) => preset.presetId === preferredPresetId && preset.presetId !== "custom");
  if (preferredPreset && !preferredPreset.hidden && !preferredPreset.dev && sameLibrarySet(normalizedSelection, normalizedPresetIds(preferredPreset, windowsShellProfileName, catalog, extendedLibraries))) {
    return preferredPreset.presetId;
  }
  for (const preset of libraryPresets) {
    if (preset.presetId === "custom" || preset.hidden || preset.dev) continue;
    const normalizedPreset = normalizedPresetIds(preset, windowsShellProfileName, catalog, extendedLibraries);
    if (sameLibrarySet(normalizedSelection, normalizedPreset)) {
      return preset.presetId;
    }
  }
  // Maximum test is catalog-derived, so it can only be matched when the catalog is
  // available. Checked last so a narrower named preset always wins.
  if (catalog && isDevUnlockEnabled()) {
    const maxTest = maximumTestLibraryIds(catalog, windowsShellProfileName).slice().sort();
    if (sameLibrarySet(normalizedSelection, maxTest)) return "maxtest";
  }
  return "custom";
}

export function isValidLibraryPresetId(value: unknown): value is LibraryPresetId {
  return value === "minimal" || value === "default" || value === "efficiency" || value === "compatibility" || value === "editor" || value === "full" || value === "ai" || value === "streaming" || value === "maxtest" || value === "custom";
}

// Presets whose displayed name is never prefixed with "Extended" even when the
// extended-libraries toggle is on. Only the broadening presets get the prefix.
const nonExtendedPresetIds = new Set<LibraryPresetId>(["minimal", "default", "maxtest", "custom"]);

const libraryPresetIconById: Partial<Record<LibraryPresetId, string>> = {
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
function libraryPresetDisplayName(presetId: LibraryPresetId, extendedLibraries: boolean): string {
  const baseName = t(`libraries.presets.${presetId}.name`);
  if (!extendedLibraries || nonExtendedPresetIds.has(presetId)) return baseName;
  return t("libraries.extended.presetPrefix") + baseName;
}

function LibraryPresetIcon(props: { presetId: LibraryPresetId; className?: string }) {
  const icon = libraryPresetIconById[props.presetId];
  if (!icon) return null;
  return <img className={props.className ?? ""} src={icon} alt="" aria-hidden="true" />;
}

export function removeMutuallyExclusiveLibraries(selectedLibraryIds: string[], selectedLibraryId: string): string[] {
  // shaderc/glslang and the EVC profile bindings stay mutually exclusive always (FFmpeg
  // configure rejects both). The TLS pick-one group is relaxed only under the sudo dev
  // tier for build testing; basic dev keeps it.
  const exclusiveGroups: Record<string, string[]> = {
    ...(isSudoDevUnlockEnabled() ? {} : {
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


export function deriveLicenseBoundaryFromSelectedLibraries(selectedLibraryIds: string[], catalog: LibraryChoice[], windowsShellProfileName: string): string {
  const selectedLibraries = visibleLibraries(catalog, windowsShellProfileName).filter((library) => selectedLibraryIds.includes(library.libraryId));
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

// Build-capability disabled rows: not an FFmpeg-version question, so they are NOT in the
// per-release manifest. tensorflow has a recipe but needs a heavy external libtensorflow
// import that is not auto-provided; vapoursynth's runtime is non-portable (needs a Python +
// VapourSynth runtime) regardless of FFmpeg version. Everything that IS version-dependent
// (lensfun, svtjpegxs, onnxruntime) is now decided per FFmpeg release by the backend manifest
// annotation (versionCompatibility), not by a global list here: lensfun is manifest-unavailable
// on every line, onnxruntime is absent from every released line (master-only switch), and
// svtjpegxs is gated by its per-release pkg-config floor at build time.
const uiDisabledLibraryIds = new Set(["tensorflow", "vapoursynth"]);

// Libraries that are present in the catalog but still have no implemented
// preparation/build recipe. They remain visible for transparency, but normal users
// cannot check them until the backend can actually prepare them. The hidden About
// tab developer unlock still makes them selectable for development testing.
const unimplementedBuildLibraryIds = new Set([
  "smbclient",
  "openvino",
  "torch",
  // No MSYS2 package and no preparation recipe yet, so these block the build and need a
  // source/SDK import path before normal users can select them.
  // (libmfx was here, but it now has an Internal-track source-build recipe — mfx_dispatch —
  // so it is selectable on every FFmpeg 4.4-8.1 where the --enable-libmfx switch exists.)
  "pocketsphinx",
  "dc1394",
  "decklink",
  "cuda-nvcc",
]);

// Per-library shell profiles with no prebuilt MSYS2 package. Keep in sync with the
// backend libraryProfileUnavailability map. A library unavailable for the active
// profile is shown disabled, not preset-selectable, and dropped from the selection.
const libraryUnavailableProfiles: Record<string, string[]> = {
  onnxruntime: ["mingw64"],
};

// Libraries kept in the catalog and backend (recipe, prep, import all stay) but made
// unselectable in the UI, with a shown reason. They remain visible so the user can see
// that the program knows about them, but cannot accidentally select a fragile option.
function isLibraryAvailableForProfile(libraryId: string, windowsShellProfileName: string): boolean {
  return !(libraryUnavailableProfiles[libraryId] ?? []).includes(windowsShellProfileName);
}

function isLibrarySupportedForFfmpegVersion(library: LibraryChoice): boolean {
  return library.versionCompatibility?.supported !== false;
}

// Available means FFmpeg has the switch for the chosen release AND the package this builder
// can supply can satisfy it on that release. It is false for a manifest-unavailable row
// (e.g. lensfun, whose package FFmpeg cannot use) even though the switch nominally exists, so
// it is a stricter gate than isLibrarySupportedForFfmpegVersion. Both are per-FFmpeg-version,
// driven by the backend manifest annotation, never a global list.
function isLibraryAvailableForFfmpegVersion(library: LibraryChoice): boolean {
  return library.versionCompatibility?.available !== false;
}

function libraryTrackLabel(trackName: string): string {
  return t(`libraries.row.track.${trackName || "native"}`);
}

function visibleLibraries(catalog: LibraryChoice[], windowsShellProfileName: string): LibraryChoice[] {
  void windowsShellProfileName;
  return catalog;
}

function isLibraryUiUnavailable(library: LibraryChoice, windowsShellProfileName: string): boolean {
  return uiDisabledLibraryIds.has(library.libraryId) || unimplementedBuildLibraryIds.has(library.libraryId) || !isLibraryAvailableForProfile(library.libraryId, windowsShellProfileName) || !isLibrarySupportedForFfmpegVersion(library) || !isLibraryAvailableForFfmpegVersion(library);
}

// Whether the checkbox is actually locked. Same as UI-unavailable, except the hidden
// About-tab developer unlock makes otherwise-unavailable libraries checkable for testing.
// The unavailable styling still shows regardless of unlock.
function isLibraryCheckboxLocked(library: LibraryChoice, windowsShellProfileName: string): boolean {
  if (!isLibrarySupportedForFfmpegVersion(library)) return true;
  if (!isLibraryAvailableForFfmpegVersion(library)) return true;
  return isLibraryUiUnavailable(library, windowsShellProfileName) && !isDevUnlockEnabled();
}

// Returns the localization key suffix for an unavailable row's note, or "" when no
// note should be shown. Unimplemented-build rows show no note: the disabled styling
// already signals they cannot be selected, and a generic "Not selectable." line added
// nothing.
function libraryUiUnavailableReasonKey(library: LibraryChoice, windowsShellProfileName: string): string {
  const libraryId = library.libraryId;
  if (!isLibrarySupportedForFfmpegVersion(library)) return "ffmpegVersionUnsupported";
  if (!isLibraryAvailableForProfile(libraryId, windowsShellProfileName)) return "profileUnavailable";
  if (unimplementedBuildLibraryIds.has(libraryId)) return "";
  return libraryId;
}

function groupLibrariesByCategory(catalog: LibraryChoice[], windowsShellProfileName: string) {
  return visibleLibraries(catalog, windowsShellProfileName).reduce<Record<string, LibraryChoice[]>>((groups, library) => {
    const categoryName = libraryText(library, "categoryName") || t("common.other");
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}


function selectableLibraryPresets(): Array<LibraryPreset & { presetId: Exclude<LibraryPresetId, "custom"> }> {
  return libraryPresets.filter((preset): preset is LibraryPreset & { presetId: Exclude<LibraryPresetId, "custom"> } =>
    preset.presetId !== "custom" && !preset.hidden && (!preset.dev || isDevUnlockEnabled())
  );
}

function SimpleLibraryPresetCard(props: { selectedPresetId: LibraryPresetId; onApplyPreset: (presetId: LibraryPresetId) => void; extendedLibraries: boolean }) {
  const presets = selectableLibraryPresets();
  const selectedPreset = presets.find((preset) => preset.presetId === props.selectedPresetId);
  const selectedPresetDescription = selectedPreset
    ? t(`libraries.presets.${selectedPreset.presetId}.plain`)
    : t("libraries.presetSelector.custom");

  return (
    <section className="card card--blue libraries-simple-card libraries-simple-preset-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={libraryPresetCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{t("libraries.simple.preset.title")}</h2>
      </div>
      <div className="card__control libraries-simple-card__control">
        <select
          className="card__input"
          value={props.selectedPresetId}
          onChange={(event) => props.onApplyPreset(event.target.value as LibraryPresetId)}
        >
          {props.selectedPresetId === "custom" && <option value="custom">{t("libraries.presets.custom.name")}</option>}
          {presets.map((preset) => (
            <option value={preset.presetId} key={preset.presetId}>
              {libraryPresetDisplayName(preset.presetId, props.extendedLibraries)}
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

function SimpleLibraryCard(props: {
  catalog: LibraryChoice[];
  selectedLibraryIds: string[];
  onToggleLibrary: (libraryId: string) => void;
  windowsShellProfileName: string;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LibrarySectionFilter[];
  sectionOptions: string[];
  onSectionFiltersChange: (value: LibrarySectionFilter[]) => void;
  onOpenOfficialWebpage: (url: string) => void;
}) {
  return (
    <section className="card card--teal libraries-simple-card libraries-simple-library-card">
      <span className="card__badge" aria-hidden="true"><img className="card__badge-icon" src={librariesCardIcon} alt="" /></span>
      <div className="card__head">
        <h2 className="card__title">{t("libraries.simple.library.title")}</h2>
      </div>
      <div className="libraries-simple-library-card__body">
        <div className="libraries-simple-library-card__controls">
          <label className="libraries-search">
            <span className="visually-hidden">{t("libraries.search.label")}</span>
            <input value={props.searchQuery} onChange={(event) => props.onSearchQueryChange(event.target.value)} placeholder={t("libraries.search.placeholder")} />
          </label>
          <LibrarySectionDropdown
            selectedSections={props.sectionFilters}
            sectionOptions={props.sectionOptions}
            onChangeSections={props.onSectionFiltersChange}
          />
        </div>
        <div className="libraries-simple-library-card__results">
          <LibraryList
            catalog={props.catalog}
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

function LibraryPresetSelector(props: { presets: LibraryPreset[]; selectedPresetId: LibraryPresetId; onApplyPreset: (presetId: LibraryPresetId) => void; showTechnicalDetails: boolean; extendedLibraries: boolean }) {
  const selectedPreset = props.presets.find((preset) => preset.presetId === props.selectedPresetId && preset.presetId !== "custom") as (LibraryPreset & { presetId: Exclude<LibraryPresetId, "custom"> }) | undefined;

  return (
    <section className="preset-panel">
      <div className="preset-grid">
        {props.presets.filter((preset): preset is LibraryPreset & { presetId: Exclude<LibraryPresetId, "custom"> } => selectableLibraryPresets().some((selectablePreset) => selectablePreset.presetId === preset.presetId)).map((preset) => (
          <button className={`preset-card ${preset.dev ? "preset-card--dev" : ""} ${props.selectedPresetId === preset.presetId ? "preset-card--active" : ""}`} type="button" key={preset.presetId} onClick={() => props.onApplyPreset(preset.presetId)}>
            <LibraryPresetIcon presetId={preset.presetId} className="preset-card__icon" />
            <span>
              <span className="preset-card__name">{libraryPresetDisplayName(preset.presetId, props.extendedLibraries)}</span>
              <span className="preset-card__plain">{t(`libraries.presets.${preset.presetId}.plain`)}</span>
            </span>
            {props.selectedPresetId === preset.presetId && <span className="preset-card__check" aria-hidden="true">✓</span>}
          </button>
        ))}
      </div>
      {props.showTechnicalDetails && selectedPreset && (
        <section className="preset-technical-card" aria-label={t("libraries.technical.show")}>
          <div className="preset-technical-card__header">
            <span className="preset-technical-card__title">
              <strong>{libraryPresetDisplayName(selectedPreset.presetId, props.extendedLibraries)}</strong>
              <span>{t("libraries.technical.show")}</span>
            </span>
          </div>
          <p className="preset-technical-card__text">{t(`libraries.presets.${selectedPreset.presetId}.technical`)}</p>
        </section>
      )}
    </section>
  );
}

function LibrarySelectionSummary(props: { catalog: LibraryChoice[]; selectedLibraryIds: string[]; selectedPresetId: LibraryPresetId; windowsShellProfileName: string; extendedLibraries: boolean }) {
  const normalizedSelection = normalizeLibrarySelection(props.selectedLibraryIds, props.windowsShellProfileName, props.catalog);
  const visibleCatalog = visibleLibraries(props.catalog, props.windowsShellProfileName);
  const licenseBoundary = deriveLicenseBoundaryFromSelectedLibraries(normalizedSelection, props.catalog, props.windowsShellProfileName);
  const selectedOptionalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && !library.defaultChecked).length;
  const selectedInternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "internal").length;
  const selectedExternalCount = visibleCatalog.filter((library) => normalizedSelection.includes(library.libraryId) && library.trackName === "external").length;
  const includedCount = visibleCatalog.filter((library) => library.defaultChecked).length;

  const selectedPresetName = libraryPresetDisplayName(props.selectedPresetId, props.extendedLibraries);
  const selectionMessage = props.selectedPresetId === "custom"
    ? t("libraries.presetSelector.custom")
    : t("libraries.summary.currentPreset", { preset: selectedPresetName });

  return (
    <section className="library-summary-card" aria-label={t("libraries.summary.ariaLabel")}>
      <div className="library-summary-card__header">
        <span className="library-summary-card__status" aria-hidden="true">✓</span>
        <span className="library-summary-card__copy">
          <strong className="library-summary-card__title">{t("libraries.summary.currentTitle")}</strong>
          <span className="library-summary-card__message">{selectionMessage}</span>
        </span>
      </div>
      <div className="library-summary">
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{t("libraries.summary.license")}</span>
            <strong className={`library-summary__value library-summary__license library-summary__license--${licenseBoundary}`}>{t(`libraries.summary.license.${licenseBoundary}`)}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{t("libraries.summary.included")}</span>
            <strong className="library-summary__value">{includedCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{t("libraries.summary.optional")}</span>
            <strong className="library-summary__value">{selectedOptionalCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{t("libraries.summary.internal")}</span>
            <strong className="library-summary__value">{selectedInternalCount}</strong>
          </span>
        </div>
        <div className="library-summary__item">
          <span className="library-summary__text">
            <span className="library-summary__label">{t("libraries.summary.externalTrack")}</span>
            <strong className="library-summary__value">{selectedExternalCount}</strong>
          </span>
        </div>
      </div>
    </section>
  );
}

type LibrarySectionFilter = string;

function libraryCategoryName(library: LibraryChoice): string {
  return libraryText(library, "categoryName") || t("common.other");
}

function isDefaultLibraryCategory(categoryName: string): boolean {
  const normalizedCategoryName = categoryName.toLocaleLowerCase();
  return normalizedCategoryName.includes("included by default") || normalizedCategoryName.includes("기본 포함");
}

function librarySectionFilterLabel(categoryName: LibrarySectionFilter): string {
  if (isDefaultLibraryCategory(categoryName)) return t("libraries.categoryFilter.default");
  return categoryName;
}

function librarySectionFilterSummary(selectedSections: LibrarySectionFilter[]): string {
  if (selectedSections.length === 0) return t("libraries.categoryFilter.all");
  if (selectedSections.length === 1) return librarySectionFilterLabel(selectedSections[0]);
  return t("libraries.categoryFilter.selectedCount").replace("{count}", String(selectedSections.length));
}

function librarySectionOptions(catalog: LibraryChoice[], windowsShellProfileName: string): string[] {
  const categoryNames: string[] = [];
  for (const library of visibleLibraries(catalog, windowsShellProfileName)) {
    const categoryName = libraryCategoryName(library);
    if (!categoryNames.includes(categoryName)) categoryNames.push(categoryName);
  }
  return categoryNames;
}

function filterLibraries(catalog: LibraryChoice[], windowsShellProfileName: string, searchQuery: string, sectionFilters: LibrarySectionFilter[]): LibraryChoice[] {
  const normalizedQuery = searchQuery.trim().toLocaleLowerCase();
  return visibleLibraries(catalog, windowsShellProfileName).filter((library) => {
    const categoryName = libraryCategoryName(library);
    if (sectionFilters.length > 0 && !sectionFilters.includes(categoryName)) return false;
    if (!normalizedQuery) return true;
    const searchableText = [
      library.libraryId,
      libraryText(library, "displayName"),
      categoryName,
      libraryText(library, "plainExplanation"),
      libraryText(library, "technicalExplanation"),
      library.configureFlags.join(" "),
      library.packageNames.join(" "),
      library.officialWebpageUrl || "",
    ].join(" ").toLocaleLowerCase();
    return searchableText.includes(normalizedQuery);
  });
}


function splitLibraryDisplayName(displayName: string): { buildName: string; featureName: string } {
  const separator = " / ";
  if (!displayName.includes(separator)) {
    return { buildName: displayName, featureName: "" };
  }
  const [buildName, ...featureParts] = displayName.split(separator);
  return { buildName: buildName.trim(), featureName: cleanLibraryFeatureName(featureParts.join(separator).trim()) };
}

function cleanLibraryFeatureName(featureName: string): string {
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

function groupFilteredLibraries(libraries: LibraryChoice[]) {
  return libraries.reduce<Record<string, LibraryChoice[]>>((groups, library) => {
    const categoryName = libraryCategoryName(library);
    groups[categoryName] = groups[categoryName] || [];
    groups[categoryName].push(library);
    return groups;
  }, {});
}

function LibraryList(props: { catalog: LibraryChoice[]; selectedLibraryIds: string[]; onToggleLibrary: (libraryId: string) => void; showTechnicalDetails: boolean; windowsShellProfileName: string; searchQuery: string; sectionFilters: LibrarySectionFilter[]; onOpenOfficialWebpage: (url: string) => void }) {
  const filteredLibraries = filterLibraries(props.catalog, props.windowsShellProfileName, props.searchQuery, props.sectionFilters);
  const groupedLibraries = groupFilteredLibraries(filteredLibraries);
  return (
    <div className="library-list">
      {filteredLibraries.length === 0 && <p className="library-list__empty">{t("libraries.list.empty")}</p>}
      {Object.entries(groupedLibraries).map(([categoryName, libraries]) => (
        <section className="library-group" key={categoryName}>
          <h2 className="library-group__title">{categoryName}</h2>
          {libraries.map((library) => {
            const isUiUnavailable = isLibraryUiUnavailable(library, props.windowsShellProfileName);
            const unavailableReasonKey = libraryUiUnavailableReasonKey(library, props.windowsShellProfileName);
            const isCheckboxLocked = isLibraryCheckboxLocked(library, props.windowsShellProfileName);
            const isChecked = (props.selectedLibraryIds.includes(library.libraryId) || library.defaultChecked) && !isCheckboxLocked;
            // TLS backends, shaderc/glslang, and the EVC binding pairs each render as a
            // radio group (pick one) in normal and basic dev mode; clicking the selected
            // radio clears it (FFmpeg allows zero). Only the sudo dev tier switches them
            // back to checkboxes so several can be selected for testing. Each group needs
            // its own radio `name` so unrelated groups do not clear each other.
            const radioGroupName = isSudoDevUnlockEnabled()
              ? undefined
              : tlsBackendLibraryIds.has(library.libraryId)
                ? "tls-backend"
                : shaderCompilerLibraryIds.has(library.libraryId)
                  ? "shader-compiler"
                  : evcDecoderLibraryIds.has(library.libraryId)
                    ? "evc-decoder"
                    : evcEncoderLibraryIds.has(library.libraryId)
                      ? "evc-encoder"
                      : intelHwaccelBackendLibraryIds.has(library.libraryId)
                        ? "intel-hwaccel-backend"
                        : undefined;
            const isExclusiveRadio = radioGroupName !== undefined;
            const trackName = library.trackName || "native";
            const isNativeTrack = trackName === "native";
            const showsOfficialWebpage = !isNativeTrack && Boolean(library.officialWebpageUrl);
            const showsPackageNames = isNativeTrack;
            const packageValue = library.packageNames.length > 0 ? library.packageNames.join(", ") : t("libraries.row.ffmpegSourcePackage");
            const hasTechnicalMetadata = library.configureFlags.length > 0 || showsOfficialWebpage || showsPackageNames;
            const displayName = libraryText(library, "displayName");
            const { buildName, featureName } = splitLibraryDisplayName(displayName);
            const statusLabel = library.defaultChecked ? t("libraries.row.included") : libraryLicenseLabel(library.licenseEffectName);
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
                    {!isNativeTrack && <span className={`library-row__track library-row__track--${trackName}`}>{libraryTrackLabel(trackName)}</span>}
                  </span>
                  <span className="library-row__copy">
                    {featureName && <span className="library-row__feature">{featureName}</span>}
                    <span className="library-row__note">{renderRichNote(libraryText(library, "plainExplanation"))}</span>
                    {isUiUnavailable && unavailableReasonKey && <span className="library-row__note">{t(`libraries.row.unavailable.${unavailableReasonKey}`)}</span>}
                  </span>
                  {props.showTechnicalDetails && libraryText(library, "technicalExplanation") && <span className="library-row__detail">{libraryText(library, "technicalExplanation")}</span>}
                  {props.showTechnicalDetails && hasTechnicalMetadata &&
                    <span className="library-row__technical-metadata">
                      {library.configureFlags.length > 0 &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--flag">{t("libraries.row.flagsLabel")}</span>
                          <span className="library-row__technical-value">{library.configureFlags.join(" ")}</span>
                        </span>
                      }
                      {showsPackageNames &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--package">{t("libraries.row.packagesLabel")}</span>
                          <span className="library-row__technical-value">{packageValue}</span>
                        </span>
                      }
                      {showsOfficialWebpage &&
                        <span className="library-row__technical-line">
                          <span className="library-row__technical-badge library-row__technical-badge--official">{t("libraries.row.officialWebpageLabel")}</span>
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



function LibrarySectionDropdown(props: { selectedSections: LibrarySectionFilter[]; sectionOptions: string[]; onChangeSections: (value: LibrarySectionFilter[]) => void }) {
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

  function toggleSection(sectionName: LibrarySectionFilter) {
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
        <strong className="libraries-section-dropdown__value">{librarySectionFilterSummary(props.selectedSections)}</strong>
        <span className="libraries-section-dropdown__chevron" aria-hidden="true" />
      </button>
      {isOpen && (
        <div className="libraries-section-dropdown__menu" role="listbox" aria-multiselectable="true" aria-label={t("libraries.categoryFilter.ariaLabel")}>
          <button
            className={`libraries-section-dropdown__option ${props.selectedSections.length === 0 ? "libraries-section-dropdown__option--active" : ""}`}
            type="button"
            role="option"
            aria-selected={props.selectedSections.length === 0}
            onClick={() => props.onChangeSections([])}
          >
            <span>{t("libraries.categoryFilter.all")}</span>
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
                <span>{librarySectionFilterLabel(sectionName)}</span>
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

function LibrariesToolbar(props: {
  showTechnicalDetails: boolean;
  onToggleTechnicalDetails: () => void;
  searchQuery: string;
  onSearchQueryChange: (value: string) => void;
  sectionFilters: LibrarySectionFilter[];
  sectionOptions: string[];
  onSectionFiltersChange: (value: LibrarySectionFilter[]) => void;
  onResetFilters: () => void;
}) {
  return (
    <div className="libraries-toolbar">
      <button className="button libraries-technical-toggle" type="button" aria-expanded={props.showTechnicalDetails} onClick={props.onToggleTechnicalDetails}>
        <img className="card__btn-icon" src={technicalDetailsIcon} alt="" aria-hidden="true" />
        {props.showTechnicalDetails ? t("libraries.technical.hide") : t("libraries.technical.show")}
      </button>
      <label className="libraries-search">
        <span className="visually-hidden">{t("libraries.search.label")}</span>
        <input value={props.searchQuery} onChange={(event) => props.onSearchQueryChange(event.target.value)} placeholder={t("libraries.search.placeholder")} />
      </label>
      <LibrarySectionDropdown
        selectedSections={props.sectionFilters}
        sectionOptions={props.sectionOptions}
        onChangeSections={props.onSectionFiltersChange}
      />
      <button className="button libraries-reset" type="button" onClick={props.onResetFilters}>
        <img className="card__btn-icon" src={resetIcon} alt="" aria-hidden="true" />
        {t("libraries.filter.reset")}
      </button>
    </div>
  );
}

// ─── LibrariesTab ─────────────────────────────────────────────────────────────

export type LibrariesTabProps = {
  initialApplicationState: InitialApplicationState;
  libraryCatalog: LibraryChoice[];
  ffmpegBuildSettings: FfmpegBuildSettings;
  libraryPresetId: LibraryPresetId;
  extendedLibraries: boolean;
  libraryDetailedView: boolean;
  setLibraryDetailedView: (value: boolean) => void;
  showTechnicalDetails: boolean;
  setShowTechnicalDetails: (value: boolean) => void;
  sectionFilters: LibrarySectionFilter[];
  setSectionFilters: (value: LibrarySectionFilter[]) => void;
  toggleLibrary: (libraryId: string) => void;
  applyLibraryPreset: (presetId: LibraryPresetId) => void;
  setExtendedLibraries: (value: boolean) => void;
  openInUserBrowser: (url: string) => Promise<void>;
};

export function LibrariesTab({ initialApplicationState, libraryCatalog, ffmpegBuildSettings, libraryPresetId, extendedLibraries, libraryDetailedView, setLibraryDetailedView, showTechnicalDetails, setShowTechnicalDetails, sectionFilters, setSectionFilters, toggleLibrary, applyLibraryPreset, setExtendedLibraries, openInUserBrowser }: LibrariesTabProps) {
  const [searchQuery, setSearchQuery] = React.useState("");
  const catalog = libraryCatalog.length > 0 ? libraryCatalog : initialApplicationState.defaultLibraryCatalog;
  const sectionOptions = librarySectionOptions(catalog, ffmpegBuildSettings.windowsShellProfileName);

  React.useEffect(() => {
    const prunedSections = sectionFilters.filter((sectionName) => sectionOptions.includes(sectionName));
    if (prunedSections.length !== sectionFilters.length) setSectionFilters(prunedSections);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sectionOptions]);

  return (
    <section className="tab-page libraries-page">
      <div className="libraries-page__header-row">
        <PageHeader title={t("libraries.title")} text={t("libraries.intro")} />
        <label className="libraries-design-toggle">
          <input type="checkbox" checked={libraryDetailedView} onChange={(event) => setLibraryDetailedView(event.target.checked)} />
          <span className="libraries-design-toggle__text">{t("libraries.designToggle.label")}</span>
        </label>
      </div>

      {!libraryDetailedView && (
        <div className="libraries-simple-layout">
          <SimpleLibraryPresetCard selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} extendedLibraries={extendedLibraries} />
          <SimpleLibraryCard
            catalog={catalog}
            selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds}
            onToggleLibrary={toggleLibrary}
            windowsShellProfileName={ffmpegBuildSettings.windowsShellProfileName}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            sectionFilters={sectionFilters}
            sectionOptions={sectionOptions}
            onSectionFiltersChange={setSectionFilters}
            onOpenOfficialWebpage={openInUserBrowser}
          />
        </div>
      )}

      {libraryDetailedView && (
        <>
          <LibraryPresetSelector presets={libraryPresets} selectedPresetId={libraryPresetId} onApplyPreset={applyLibraryPreset} showTechnicalDetails={showTechnicalDetails} extendedLibraries={extendedLibraries} />
          <label className="library-extended-toggle">
            <input type="checkbox" checked={extendedLibraries} onChange={(event) => setExtendedLibraries(event.target.checked)} />
            <span className="library-extended-toggle__label">{t("libraries.extended.label")}</span>
          </label>
          <LibrarySelectionSummary catalog={catalog} selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds} selectedPresetId={libraryPresetId} windowsShellProfileName={ffmpegBuildSettings.windowsShellProfileName} extendedLibraries={extendedLibraries} />
          <LibrariesToolbar
            showTechnicalDetails={showTechnicalDetails}
            onToggleTechnicalDetails={() => setShowTechnicalDetails(!showTechnicalDetails)}
            searchQuery={searchQuery}
            onSearchQueryChange={setSearchQuery}
            sectionFilters={sectionFilters}
            sectionOptions={sectionOptions}
            onSectionFiltersChange={setSectionFilters}
            onResetFilters={() => {
              setSearchQuery("");
              setSectionFilters([]);
            }}
          />
          {showTechnicalDetails && (
            <section className="libraries-technical-panel">
              <h2 className="libraries-technical-panel__title">{t("libraries.technical.title")}</h2>
              <div className="libraries-technical-details">
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{t("libraries.technical.builtIn.title")}</h3>
                  <p className="libraries-technical-detail__text">{t("libraries.technical.builtIn.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{t("libraries.technical.sourceBuild.title")}</h3>
                  <p className="libraries-technical-detail__text">{t("libraries.technical.sourceBuild.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{t("libraries.technical.customBuild.title")}</h3>
                  <p className="libraries-technical-detail__text">{t("libraries.technical.customBuild.text")}</p>
                </section>
                <section className="libraries-technical-detail">
                  <h3 className="libraries-technical-detail__title">{t("libraries.technical.license.title")}</h3>
                  <p className="libraries-technical-detail__text">{t("libraries.technical.license.text")}</p>
                </section>
              </div>
            </section>
          )}
          <LibraryList catalog={catalog} selectedLibraryIds={ffmpegBuildSettings.selectedLibraryIds} onToggleLibrary={toggleLibrary} showTechnicalDetails={showTechnicalDetails} windowsShellProfileName={ffmpegBuildSettings.windowsShellProfileName} searchQuery={searchQuery} sectionFilters={sectionFilters} onOpenOfficialWebpage={openInUserBrowser} />
        </>
      )}
    </section>
  );
}
