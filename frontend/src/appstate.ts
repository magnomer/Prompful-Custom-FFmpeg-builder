// App-level types, constants, and pure utility functions.
// No React, no Wails imports.

import type { LibraryPresetId } from "./tabs/libraries";

// ─── Types ───────────────────────────────────────────────────────────────────

export type TabId = "source" | "buildConfig" | "prep" | "library" | "options" | "buildFfmpeg" | "result" | "logs" | "about";

export type SavedUiState = {
  activeTabId?: TabId;
  buildConfigSettings?: BuildConfigSettings;
  ffmpegBuildSettings?: FfmpegBuildSettings;
  msys2PackageText?: string;
  extraConfigureFlagText?: string;
  libraryPresetId?: LibraryPresetId;
  extendedLibraries?: boolean;
  libraryDetailedView?: boolean;
  optionsDetailedView?: boolean;
  libraryTechnicalDetails?: boolean;
  optionsTechnicalDetails?: boolean;
  librarySectionFilters?: string[];
};

export type SavedWindowState = {
  width?: number;
  height?: number;
  x?: number;
  y?: number;
};

// ─── Constants ───────────────────────────────────────────────────────────────

export const savedUiStateKey = "customffmpeg.builder.uiState.v1";
export const savedWindowStateKey = "customffmpeg.builder.windowState.v1";

export const emptyBuildConfigSettings: BuildConfigSettings = {
  workspaceDirectory: "",
  msys2ArchiveUrl: "",
  msys2ArchiveSha256Hash: "",
  msys2ArchiveSignatureUrl: "",
  msys2PackageNames: [],
  windowsShellProfileName: "ucrt64",
};

export const emptyFfmpegBuildSettings: FfmpegBuildSettings = {
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

export const defaultInitialApplicationState: InitialApplicationState = {
  hostOs: "unknown",
  kindExplanation: "",
  securityRuleSummary: "",
  namingRuleSummary: "",
  defaultBuildConfigSettings: emptyBuildConfigSettings,
  defaultFfmpegBuildSettings: emptyFfmpegBuildSettings,
  defaultLibraryCatalog: [],
  defaultConfigureOptionCatalog: [],
  supportedFfmpegReleases: [],
};

// ─── Utilities ───────────────────────────────────────────────────────────────

export function splitLines(value: string): string[] {
  return value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

// MSYS2 package prefix per shell profile. Mirrors packagePrefixForShellProfile in
// internal/planning/catalog.go — keep both in sync.
export const msys2PackagePrefixByProfile: Record<string, string> = {
  ucrt64: "mingw-w64-ucrt-x86_64",
  mingw64: "mingw-w64-x86_64",
  clang64: "mingw-w64-clang-x86_64",
};

// Matches any of the three known MSYS2 prefixes at the start of a package line.
// ucrt-x86_64 / clang-x86_64 must precede the bare x86_64 alternative so the
// longer match wins.
const knownMsys2PrefixPattern = /^mingw-w64-(?:ucrt-x86_64|clang-x86_64|x86_64)-/;

// remapMsys2PackagePrefixes rewrites the MSYS2 prefix of every prefixed package
// line to the target profile's prefix, leaving unprefixed packages (base-devel,
// git, ...) and any custom lines untouched. Used when the shell profile changes
// so the toolchain and library packages target the selected environment.
export function remapMsys2PackagePrefixes(text: string, targetProfileName: string): string {
  const targetPrefix = msys2PackagePrefixByProfile[targetProfileName] ?? msys2PackagePrefixByProfile.ucrt64;
  return text.split(/\r?\n/).map((line) => {
    const match = line.match(knownMsys2PrefixPattern);
    if (!match) return line;
    return targetPrefix + "-" + line.slice(match[0].length);
  }).join("\n");
}

export function normalizeLogLevel(value: string): "info" | "warn" | "error" {
  if (value === "warn" || value === "error") return value;
  return "info";
}

export function isValidTabId(value: unknown): value is TabId {
  return value === "source" || value === "buildConfig" || value === "prep" || value === "library" || value === "options" || value === "buildFfmpeg" || value === "result" || value === "logs" || value === "about";
}

export function createApprovalRequest(actionName: string, planHash: string, consentText: string): ApprovalRequest {
  return { approvedActionName: actionName, approvedPlanHash: planHash, consentText };
}

export function parseSavedUiState(raw: string): SavedUiState {
  try {
    if (!raw) return {};
    const parsed = JSON.parse(raw) as SavedUiState;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch { return {}; }
}

export function readSavedWindowState(): SavedWindowState {
  try {
    const raw = window.localStorage.getItem(savedWindowStateKey);
    if (!raw) return {};
    const parsed = JSON.parse(raw) as SavedWindowState;
    return parsed && typeof parsed === "object" ? parsed : {};
  } catch { return {}; }
}
