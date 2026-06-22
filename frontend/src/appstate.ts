// App-level types, constants, and pure utility functions.
// No React, no Wails imports.

import type { LibraryPresetId } from "./tabs/libraries";

// ─── Types ───────────────────────────────────────────────────────────────────

export type TabId = "source" | "buildTools" | "prep" | "library" | "options" | "buildFfmpeg" | "result" | "logs" | "about";

export type SavedUiState = {
  activeTabId?: TabId;
  buildToolSettings?: BuildToolSettings;
  ffmpegBuildSettings?: FfmpegBuildSettings;
  msys2PackageText?: string;
  extraConfigureFlagText?: string;
  libraryPresetId?: LibraryPresetId;
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

export const emptyBuildToolSettings: BuildToolSettings = {
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
  defaultBuildToolSettings: emptyBuildToolSettings,
  defaultFfmpegBuildSettings: emptyFfmpegBuildSettings,
  defaultLibraryCatalog: [],
  defaultConfigureOptionCatalog: [],
};

// ─── Utilities ───────────────────────────────────────────────────────────────

export function splitLines(value: string): string[] {
  return value.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

export function normalizeLogLevel(value: string): "info" | "warn" | "error" {
  if (value === "warn" || value === "error") return value;
  return "info";
}

export function isValidTabId(value: unknown): value is TabId {
  return value === "source" || value === "buildTools" || value === "prep" || value === "library" || value === "options" || value === "buildFfmpeg" || value === "result" || value === "logs" || value === "about";
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
