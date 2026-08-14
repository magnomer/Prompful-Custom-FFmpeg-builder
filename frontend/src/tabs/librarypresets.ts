import { LLocaleTextGet } from "../i18n";
import { LUnlockBasicCheck, LUnlockSudoCheck } from "../devUnlock";
import { LLibraryUnavailableCheck, LLibraryVisibleFilter } from "./librarycatalog";

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

export function LPresetLibraryClean(presets: LPresetLibraryChoice[] | undefined): LPresetLibrary[] {
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

// LLibraryTestGet returns every selectable library from the backend-resolved
// catalog. LLibrarySelectionNormalize then resolves mutually-exclusive groups and drops
// rows that the backend marked unavailable for the active FFmpeg version/profile.
export function LLibraryTestGet(LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName?: string): string[] {
  const candidateIds = LCatalogLibrarySource
    .map((library) => library.libraryId)
    .filter((libraryId) => {
      const library = LCatalogLibrarySource.find((item) => item.libraryId === libraryId);
      return library && !LLibraryUnavailableCheck(library, windowsShellProfileName ?? "");
    });
  return LLibrarySelectionNormalize(candidateIds, windowsShellProfileName, LCatalogLibrarySource);
}

// ─── Library selection utilities ─────────────────────────────────────────────

// TLS backends form a pick-one group in normal AND basic dev mode (rendered as radio
// buttons). Only the sudo dev tier relaxes this so every backend can be selected at once
// for build testing, so the mutual-exclusion pruning below is skipped only when sudo.
export const LLibraryTlsBackend = new Set(["openssl", "gnutls", "mbedtls", "libtls"]);

// shaderc/glslang are a pick-one shader-compiler group. They are not a radio group (zero
// may be selected), but they share the TLS group's shortened-divider visual so the two
// rows read as one "choose at most one" block.
export const LLibraryShaderCompiler = new Set(["shaderc", "glslang"]);

// xevd/xevdb and xeve/xeveb are EVC full-profile and baseline-profile bindings.
// FFmpeg configure rejects enabling both members of either pair, so they share the
// same pick-one radio + divider treatment with separate radio groups.
export const LLibraryEvcDecoder = new Set(["xevd", "xevdb"]);
export const LLibraryEvcEncoder = new Set(["xeve", "xeveb"]);

// libvpl (Intel oneVPL, --enable-libvpl) and libmfx (legacy Intel Media SDK, --enable-libmfx)
// are the two Intel Hardware Acceleration backends. FFmpeg configure rejects enabling both ("can
// not use libmfx and libvpl together"), so they are a pick-one radio group with the same divider
// treatment as the EVC pairs. They are adjacent in the library catalog so the two rows read as one block.
export const LLibraryIntelBackend = new Set(["libvpl", "libmfx"]);

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
    const LLibrarySourceTable = new Map(LCatalogLibrarySource.map((library) => [library.libraryId, library]));
    for (const libraryId of [...selectedSet]) {
      const library = LLibrarySourceTable.get(libraryId);
      if (library && LLibraryUnavailableCheck(library, windowsShellProfileName ?? "")) {
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

function LLibrarySelectionCheck(a: string[], b: string[]): boolean {
  return a.length === b.length && a.every((libraryId, index) => libraryId === b[index]);
}

function LPresetIdentifiersGet(preset: LPresetLibrary, windowsShellProfileName?: string, LCatalogLibrarySource?: LLibraryChoice[], extendedLibraries = false): string[] {
  return LLibrarySelectionNormalize(LPresetLibraryResolve(preset, extendedLibraries), windowsShellProfileName, LCatalogLibrarySource).slice().sort();
}

export function LPresetLibraryMatch(selectedLibraryIds: string[], presets: LPresetLibrary[], windowsShellProfileName?: string, LCatalogLibrarySource?: LLibraryChoice[], extendedLibraries = false, preferredPresetId?: LPresetLibraryId): LPresetLibraryId {
  const normalizedSelection = LLibrarySelectionNormalize(selectedLibraryIds, windowsShellProfileName, LCatalogLibrarySource).slice().sort();
  const preferredPreset = presets.find((preset) => preset.presetId === preferredPresetId && preset.presetId !== "custom");
  if (preferredPreset && !preferredPreset.hidden && !preferredPreset.dev && LLibrarySelectionCheck(normalizedSelection, LPresetIdentifiersGet(preferredPreset, windowsShellProfileName, LCatalogLibrarySource, extendedLibraries))) {
    return preferredPreset.presetId;
  }
  for (const preset of presets) {
    if (preset.presetId === "custom" || preset.hidden || preset.dev) continue;
    const normalizedPreset = LPresetIdentifiersGet(preset, windowsShellProfileName, LCatalogLibrarySource, extendedLibraries);
    if (LLibrarySelectionCheck(normalizedSelection, normalizedPreset)) {
      return preset.presetId;
    }
  }
  // Maximum test is library catalog-derived, so it can only be matched when the library catalog is
  // available. Checked last so a narrower named preset always wins.
  if (LCatalogLibrarySource && LUnlockBasicCheck()) {
    const maxTest = LLibraryTestGet(LCatalogLibrarySource, windowsShellProfileName).slice().sort();
    if (LLibrarySelectionCheck(normalizedSelection, maxTest)) return "maxtest";
  }
  return "custom";
}

export function LPresetLibraryValidate(value: unknown): value is LPresetLibraryId {
  return value === "minimal" || value === "default" || value === "efficiency" || value === "compatibility" || value === "editor" || value === "full" || value === "ai" || value === "streaming" || value === "maxtest" || value === "custom";
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

export function LLicenseBoundaryGet(selectedLibraryIds: string[], LCatalogLibrarySource: LLibraryChoice[], windowsShellProfileName: string): string {
  const selectedLibraries = LLibraryVisibleFilter(LCatalogLibrarySource, windowsShellProfileName).filter((library) => selectedLibraryIds.includes(library.libraryId));
  if (selectedLibraries.some((library) => library.licenseEffectName === "nonfree")) return "nonfree-local";
  if (selectedLibraries.some((library) => library.licenseEffectName === "gpl")) return "gpl-local";
  return "lgpl-local";
}

// Presets whose displayed name is never prefixed with "Extended" even when the
// extended-libraries toggle is on. Only the broadening presets get the prefix.
const LPresetNonextendedIds = new Set<LPresetLibraryId>(["minimal", "default", "maxtest", "custom"]);

// Resolves the localized preset name, prepending the "Extended" prefix when the
// extended-libraries toggle is on (descriptions are left unchanged).
export function LPresetNameGet(presetId: LPresetLibraryId, extendedLibraries: boolean): string {
  const baseName = LLocaleTextGet(`libraries.presets.${presetId}.name`);
  if (!extendedLibraries || LPresetNonextendedIds.has(presetId)) return baseName;
  return LLocaleTextGet("libraries.extended.presetPrefix") + baseName;
}

export function LPresetLibraryList(presets: LPresetLibrary[]): Array<LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> }> {
  return presets.filter((preset): preset is LPresetLibrary & { presetId: Exclude<LPresetLibraryId, "custom"> } =>
    preset.presetId !== "custom" && !preset.hidden && (!preset.dev || LUnlockBasicCheck())
  );
}
