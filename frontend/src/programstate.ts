// LProgram-level types, constants, and pure utility functions.
// No React, no Wails imports.

import type { LPresetLibraryId } from "./tabs/libraries";

// ─── Types ───────────────────────────────────────────────────────────────────

export type LTabIdentifier = "source" | "buildConfig" | "prep" | "library" | "options" | "buildFfmpeg" | "result" | "logs" | "about";

export type LStateUiSaved = {
  activeTabId?: LTabIdentifier;
  buildConfigSettings?: LSettingsToolchain;
  ffmpegBuildSettings?: LSettingsFFmpeg;
  msys2PackageText?: string;
  extraConfigureFlagText?: string;
  libraryPresetId?: LPresetLibraryId;
  extendedLibraries?: boolean;
  libraryDetailedView?: boolean;
  optionsDetailedView?: boolean;
  libraryTechnicalDetails?: boolean;
  optionsTechnicalDetails?: boolean;
  librarySectionFilters?: string[];
};

export type LStateWindowSaved = {
  width?: number;
  height?: number;
  x?: number;
  y?: number;
};

// ─── Constants ───────────────────────────────────────────────────────────────

export const LStateUiKey = "customffmpeg.builder.uiState.v1";
export const LStateWindowKey = "customffmpeg.builder.windowState.v1";

export const LSettingsBuildEmpty: LSettingsToolchain = {
  workspaceDirectory: "",
  msys2ArchiveUrl: "",
  msys2ArchiveSha256Hash: "",
  msys2ArchiveSignatureUrl: "",
  msys2PackageNames: [],
  windowsShellProfileName: "ucrt64",
};

export const LSettingsFFmpegEmpty: LSettingsFFmpeg = {
  workspaceDirectory: "",
  ffmpegSourceArchiveUrl: "",
  ffmpegSourceSignatureUrl: "",
  ffmpegSourceSha256Hash: "",
  selectedLibraryIds: [],
  selectedConfigureOptionIds: ["default-static", "default-programs", "default-ffmpeg", "default-ffprobe"],
  extraConfigureFlags: [],
  configureFlags: [],
  parallelJobCount: 1,
  windowsShellProfileName: "ucrt64",
  licenseProfileName: "lgpl-local",
};

export const LStateInitialDefault: LStateInitial = {
  hostOs: "unknown",
  kindExplanation: "",
  securityRuleSummary: "",
  namingRuleSummary: "",
  defaultBuildConfigSettings: LSettingsBuildEmpty,
  defaultFfmpegBuildSettings: LSettingsFFmpegEmpty,
  defaultLibraryCatalog: [],
  defaultLibraryPresetCatalog: [],
  defaultConfigureOptionCatalog: [],
  supportedFfmpegReleases: [],
};

// ─── Utilities ───────────────────────────────────────────────────────────────

export function LTextLineSplit(value: string): string[] {
  return value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

// MSYS2 package prefix per shell profile. Mirrors LPackageProfileResolve in
// internal/planning/library catalog.go — keep both in sync.
export const LPackagePrefixTable: Record<string, string> = {
  ucrt64: "mingw-w64-ucrt-x86_64",
  mingw64: "mingw-w64-x86_64",
  clang64: "mingw-w64-clang-x86_64",
};

// Matches any of the three known MSYS2 prefixes at the start of a package line.
// ucrt-x86_64 / clang-x86_64 must precede the bare x86_64 alternative so the
// longer match wins.
const LPackagePrefixPattern = /^mingw-w64-(?:ucrt-x86_64|clang-x86_64|x86_64)-/;

// LPackagePrefixUpdate rewrites the MSYS2 prefix of every prefixed package
// line to the target profile's prefix, leaving unprefixed packages (base-devel,
// git, ...) and any custom lines untouched. Used when the shell profile changes
// so the toolchain and library packages target the selected environment.
export function LPackagePrefixUpdate(text: string, targetProfileName: string): string {
  const targetPrefix = LPackagePrefixTable[targetProfileName] ?? LPackagePrefixTable.ucrt64;
  return text.split(/\r?\n/).map((line) => {
    const match = line.match(LPackagePrefixPattern);
    if (!match) return line;
    return targetPrefix + "-" + line.slice(match[0].length);
  }).join("\n");
}

export function LLogLevelNormalize(value: string): "info" | "warn" | "error" {
  if (value === "warn" || value === "error") return value;
  return "info";
}

export function LTabIdValidate(value: unknown): value is LTabIdentifier {
  return value === "source" || value === "buildConfig" || value === "prep" || value === "library" || value === "options" || value === "buildFfmpeg" || value === "result" || value === "logs" || value === "about";
}

export function LRequestApprovalCreate(actionName: string, planHash: string, consentText: string): LRequestApproval {
  return { approvedActionName: actionName, approvedPlanHash: planHash, consentText };
}

export function LStateUiParse(raw: string): LStateUiSaved {
  try {
    if (!raw) return {};
    const parsed = JSON.parse(raw) as LStateUiSaved;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch { return {}; }
}

function LStringValueGet(value: unknown, fallback: string): string {
  return typeof value === "string" ? value : fallback;
}

function LNumberValueGet(value: unknown, fallback: number): number {
  return typeof value === "number" && Number.isFinite(value) ? value : fallback;
}

function LStringArrayGet(value: unknown, fallback: string[]): string[] {
  return Array.isArray(value) ? value.filter((item): item is string => typeof item === "string") : fallback;
}

function LArrayValueGet<T>(value: unknown, fallback: T[] = []): T[] {
  return Array.isArray(value) ? value as T[] : fallback;
}

export function LSettingsBuildNormalize(value: unknown, fallback: LSettingsToolchain = LSettingsBuildEmpty): LSettingsToolchain {
  const source = value && typeof value === "object" ? value as Partial<LSettingsToolchain> : {};
  return {
    workspaceDirectory: LStringValueGet(source.workspaceDirectory, fallback.workspaceDirectory),
    msys2ArchiveUrl: LStringValueGet(source.msys2ArchiveUrl, fallback.msys2ArchiveUrl),
    msys2ArchiveSha256Hash: LStringValueGet(source.msys2ArchiveSha256Hash, fallback.msys2ArchiveSha256Hash),
    msys2ArchiveSignatureUrl: LStringValueGet(source.msys2ArchiveSignatureUrl, fallback.msys2ArchiveSignatureUrl),
    msys2PackageNames: LStringArrayGet(source.msys2PackageNames, fallback.msys2PackageNames),
    windowsShellProfileName: LStringValueGet(source.windowsShellProfileName, fallback.windowsShellProfileName),
  };
}

export function LSettingsFFmpegNormalize(value: unknown, fallback: LSettingsFFmpeg = LSettingsFFmpegEmpty): LSettingsFFmpeg {
  const source = value && typeof value === "object" ? value as Partial<LSettingsFFmpeg> : {};
  return {
    workspaceDirectory: LStringValueGet(source.workspaceDirectory, fallback.workspaceDirectory),
    ffmpegSourceArchiveUrl: LStringValueGet(source.ffmpegSourceArchiveUrl, fallback.ffmpegSourceArchiveUrl),
    ffmpegSourceSignatureUrl: LStringValueGet(source.ffmpegSourceSignatureUrl, fallback.ffmpegSourceSignatureUrl),
    ffmpegSourceSha256Hash: LStringValueGet(source.ffmpegSourceSha256Hash, fallback.ffmpegSourceSha256Hash),
    selectedLibraryIds: LStringArrayGet(source.selectedLibraryIds, fallback.selectedLibraryIds),
    selectedConfigureOptionIds: LStringArrayGet(source.selectedConfigureOptionIds, fallback.selectedConfigureOptionIds),
    extraConfigureFlags: LStringArrayGet(source.extraConfigureFlags, fallback.extraConfigureFlags),
    configureFlags: LStringArrayGet(source.configureFlags, fallback.configureFlags),
    parallelJobCount: LNumberValueGet(source.parallelJobCount, fallback.parallelJobCount),
    windowsShellProfileName: LStringValueGet(source.windowsShellProfileName, fallback.windowsShellProfileName),
    licenseProfileName: LStringValueGet(source.licenseProfileName, fallback.licenseProfileName),
  };
}

export function LStateInitialNormalize(value: unknown, fallback: LStateInitial = LStateInitialDefault): LStateInitial {
  const source = value && typeof value === "object" ? value as Partial<LStateInitial> : {};
  return {
    hostOs: LStringValueGet(source.hostOs, fallback.hostOs),
    kindExplanation: LStringValueGet(source.kindExplanation, fallback.kindExplanation),
    securityRuleSummary: LStringValueGet(source.securityRuleSummary, fallback.securityRuleSummary),
    namingRuleSummary: LStringValueGet(source.namingRuleSummary, fallback.namingRuleSummary),
    defaultBuildConfigSettings: LSettingsBuildNormalize(source.defaultBuildConfigSettings, fallback.defaultBuildConfigSettings),
    defaultFfmpegBuildSettings: LSettingsFFmpegNormalize(source.defaultFfmpegBuildSettings, fallback.defaultFfmpegBuildSettings),
    defaultLibraryCatalog: LArrayValueGet<LLibraryChoice>(source.defaultLibraryCatalog, fallback.defaultLibraryCatalog),
    defaultLibraryPresetCatalog: LArrayValueGet<LPresetLibraryChoice>(source.defaultLibraryPresetCatalog, fallback.defaultLibraryPresetCatalog),
    defaultConfigureOptionCatalog: LArrayValueGet<LOptionChoice>(source.defaultConfigureOptionCatalog, fallback.defaultConfigureOptionCatalog),
    supportedFfmpegReleases: LArrayValueGet<LReleaseChoice>(source.supportedFfmpegReleases, fallback.supportedFfmpegReleases),
  };
}

export function LStateWindowRead(): LStateWindowSaved {
  try {
    const raw = window.localStorage.getItem(LStateWindowKey);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as LStateWindowSaved;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch { return {}; }
}
