// Pure log infrastructure — types, constants, phase detectors, parse/compute functions.
// No React. Imported by logs.tsx (UI) and LStateBuilderUse.ts (progress computation).
import { LLocaleTextGet } from "../i18n";

// ─── Types ───────────────────────────────────────────────────────────────────

export type LLogSecurityEntry = {
  timestamp: string;
  level: "info" | "warn" | "error";
  message: string;
};

export type LLogSecurityPayload = {
  level: string;
  message: string;
};

export type LStatusActionPayload = {
  status: string;
};

export type LPhaseToolchainId =
  | "tc-download"
  | "tc-extract"
  | "tc-keyring"
  | "tc-syncdb"
  | "tc-install"
  | "tc-verify";

export type LPhaseFFmpegId =
  | "ff-download"
  | "ff-pkgconfig"
  | "ff-configure"
  | "ff-compile"
  | "ff-extraction";

export type LPhaseLogId = LPhaseToolchainId | LPhaseFFmpegId | "other";

export type LLogParsedEntry = LLogSecurityEntry & {
  phase: LPhaseLogId;
  compileOp?: string;
  compileTarget?: string;
  dllName?: string;
  dllAction?: "copied" | "skipped" | "system" | "found";
  dllDep?: string;
  isSystemDll?: boolean;
};

export type LPhaseLogGroup = {
  phase: LPhaseLogId;
  label: string;
  entries: LLogParsedEntry[];
  compileCount: number;
  assembleCount: number;
  copiedDlls: string[];
  systemDllCount: number;
  skippedDllCount: number;
  startTime?: string;
  endTime?: string;
};

export type LProgressLive = {
  currentPhaseLabel: string | null;
  currentPhaseId: LPhaseLogId | null;
  compileCount: number;
  assembleCount: number;
  copiedDllCount: number;
  lastMessage: string | null;
  failureMessages: string[];
  isComplete: boolean;
  hasFailed: boolean;
  phaseGroups?: LPhaseLogGroup[];
};

// ─── Constants ───────────────────────────────────────────────────────────────

export const LLogCompileOps = new Set(["CC", "CXX", "HOSTCC", "X86ASM", "WINDRES"]);
export const LLogDocsOps    = new Set(["HTML", "POD", "TXT", "TEXI", "GENTEXI"]);
export const LLogBuildOps   = new Set(["AR", "LDXX", "LD", "HOSTLD"]);
export const LLogStripOps   = new Set(["STRIP"]);
export const LLogShaderOps  = new Set(["GLSLC", "BIN2C", "GZIP", "MINIFY"]);

const LPhaseLabelKeys: Record<LPhaseLogId, string> = {
  "tc-download": "logs.phase.tcDownload",
  "tc-extract":  "logs.phase.tcExtract",
  "tc-keyring":  "logs.phase.tcKeyring",
  "tc-syncdb":   "logs.phase.tcSyncDb",
  "tc-install":  "logs.phase.tcInstall",
  "tc-verify":   "logs.phase.tcVerify",
  "ff-download":   "logs.phase.ffDownload",
  "ff-pkgconfig":  "logs.phase.ffPkgconfig",
  "ff-configure":  "logs.phase.ffConfigure",
  "ff-compile":    "logs.phase.ffCompile",
  "ff-extraction": "logs.phase.ffExtraction",
  "other":         "common.other",
};

export function LPhaseLabelGet(phase: LPhaseLogId): string {
  return LLocaleTextGet(LPhaseLabelKeys[phase] ?? "common.other");
}

export function LLogRuntimeBuild(message: string): string {
  const text = message.trim();

  let match = text.match(/^Downloading approved file from (.+)$/);
  if (match) return LLocaleTextGet("runtimeLog.downloadingApprovedFile", { source: match[1] });

  match = text.match(/^Calculated SHA-256 for (.+?):\s*([0-9a-fA-F]+)$/);
  if (match) return LLocaleTextGet("runtimeLog.sha256Calculated", { source: match[1], hash: match[2] });

  match = text.match(/^Approved FFmpeg build started\. Run:\s*(.+)$/);
  if (match) return LLocaleTextGet("run.log.ffmpegStarted", { runId: match[1] });

  if (text === "Approved FFmpeg build completed. Artifact report written.") return LLocaleTextGet("run.log.ffmpegCompleted");

  if (text === "Extracting approved archive inside workspace.") return LLocaleTextGet("runtimeLog.extractingApprovedArchive");

  match = text.match(/^Starting FFmpeg configure at (.+)$/);
  if (match) return LLocaleTextGet("runtimeLog.startingFfmpegConfigure", { time: match[1] });

  match = text.match(/^Starting FFmpeg make at (.+)$/);
  if (match) return LLocaleTextGet("runtimeLog.startingFfmpegMake", { time: match[1] });

  match = text.match(/^Starting FFmpeg artifact collection at (.+)$/);
  if (match) return LLocaleTextGet("runtimeLog.startingFfmpegArtifactCollection", { time: match[1] });

  match = text.match(/^Emptying FFmpeg artifact directory before copying the new build: (.+)$/);
  if (match) return LLocaleTextGet("runtimeLog.emptyingFfmpegArtifactDirectory", { path: match[1] });

  match = text.match(/^Removed (\d+) stale FFmpeg artifact entries before copying the new build\.$/);
  if (match) return LLocaleTextGet("runtimeLog.removedStaleFfmpegArtifacts", { count: match[1] });

  if (text.startsWith("Running approved command")) return LLocaleTextGet("runtimeLog.runningApprovedCommand");

  return message;
}

export const LPhaseToolchainOrder: LPhaseLogId[] = [
  "tc-download", "tc-extract", "tc-keyring", "tc-syncdb", "tc-install", "tc-verify",
];
export const LPhaseFFmpegOrder: LPhaseLogId[] = [
  "ff-download", "ff-pkgconfig", "ff-configure", "ff-compile", "ff-extraction",
];

export function LPipelineToolchainGet(): { id: LPhaseLogId; label: string; short: string }[] {
  return [
    { id: "tc-download", label: LLocaleTextGet("logs.phase.tcDownload"), short: LLocaleTextGet("logs.phaseShort.download") },
    { id: "tc-extract",  label: LLocaleTextGet("logs.phase.tcExtract"),  short: LLocaleTextGet("logs.phaseShort.extract") },
    { id: "tc-keyring",  label: LLocaleTextGet("logs.phase.tcKeyring"),  short: LLocaleTextGet("logs.phaseShort.keyring") },
    { id: "tc-syncdb",   label: LLocaleTextGet("logs.phase.tcSyncDb"),   short: LLocaleTextGet("logs.phaseShort.syncDbs") },
    { id: "tc-install",  label: LLocaleTextGet("logs.phase.tcInstall"),  short: LLocaleTextGet("logs.phaseShort.install") },
    { id: "tc-verify",   label: LLocaleTextGet("logs.phase.tcVerify"),   short: LLocaleTextGet("logs.phaseShort.verify") },
  ];
}

export function LPipelineFFmpegGet(): { id: LPhaseLogId; label: string; short: string }[] {
  return [
    { id: "ff-download",   label: LLocaleTextGet("logs.phase.ffDownload"),   short: LLocaleTextGet("logs.phaseShort.download") },
    { id: "ff-pkgconfig",  label: LLocaleTextGet("logs.phase.ffPkgconfig"),  short: LLocaleTextGet("logs.phaseShort.library") },
    { id: "ff-configure",  label: LLocaleTextGet("logs.phase.ffConfigure"),  short: LLocaleTextGet("logs.phaseShort.configure") },
    { id: "ff-compile",    label: LLocaleTextGet("logs.phase.ffCompile"),    short: LLocaleTextGet("logs.phaseShort.compile") },
    { id: "ff-extraction", label: LLocaleTextGet("logs.phase.ffExtraction"), short: LLocaleTextGet("logs.phaseShort.extraction") },
  ];
}

// ─── Phase detectors ─────────────────────────────────────────────────────────

export function LPhaseToolchainDetect(msg: string): LPhaseLogId {
  if (
    msg.startsWith("Approved private MSYS2 preparation started") ||
    msg.startsWith("Downloading approved file from MSYS2") ||
    msg.startsWith("Calculated SHA-256 for MSYS2") ||
    msg.startsWith("MSYS2 .sig verification")
  ) return "tc-download";
  if (
    msg.startsWith("A previous private MSYS2 folder") ||
    msg.startsWith("Stopped private MSYS2") ||
    msg.startsWith("Previous private MSYS2 folder removed") ||
    msg.startsWith("Extracting approved archive") ||
    msg.startsWith("MSYS2 archive") ||
    msg.startsWith("Running approved command")
  ) return "tc-extract";
  if (
    msg.startsWith("Initializing the private MSYS2 package keyring") ||
    msg.startsWith("Using the official MSYS2 package server") ||
    msg.startsWith("Preparing pacman database") ||
    msg.startsWith("gpg:") ||
    msg.startsWith("==>") ||
    msg.startsWith("->")
  ) return "tc-keyring";
  if (
    msg.startsWith("Clearing stale") ||
    msg.startsWith("Refreshing the private") ||
    msg.startsWith(":: Synchronizing") ||
    msg.includes(" downloading...") ||
    msg.startsWith("clangarm64") || msg.startsWith("clang64") ||
    msg.startsWith("ucrt64") || msg.startsWith("mingw64") ||
    msg.startsWith("mingw32") || msg.startsWith("msys ")
  ) return "tc-syncdb";
  if (
    msg.startsWith("Checking that the selected") ||
    msg.startsWith("The compiler check") ||
    msg.startsWith("Approved private MSYS2 environment is ready")
  ) return "tc-verify";
  if (
    msg.startsWith("Installing the approved") ||
    msg.startsWith("there is nothing to do") ||
    msg.startsWith(":: Proceed") ||
    msg.startsWith(":: There are") ||
    msg.startsWith(":: Repository") ||
    msg.startsWith("Enter a selection") ||
    msg.startsWith("Packages (") ||
    msg.startsWith("Total Download") ||
    msg.startsWith("Total Installed") ||
    msg.startsWith("Net Upgrade") ||
    msg.startsWith("resolving dependencies") ||
    msg.startsWith("looking for conflicting") ||
    msg.startsWith("checking keyring") ||
    msg.startsWith("checking package integrity") ||
    msg.startsWith("loading package") ||
    msg.startsWith("checking for file conflicts") ||
    msg.startsWith("checking available disk space") ||
    msg.startsWith(":: Processing package changes") ||
    msg.startsWith(":: Running post-transaction") ||
    msg.startsWith("installing ") ||
    msg.startsWith("reinstalling ") ||
    msg.startsWith("upgrading ") ||
    msg.startsWith("downgrading ") ||
    msg.startsWith("removing ") ||
    msg.startsWith("Optional dependencies for") ||
    msg.startsWith("updating font cache") ||
    msg.match(/^\(\d+\/\d+\)/) !== null ||
    msg.startsWith("warning: ")
  ) return "tc-install";
  return "other";
}

export function LPhaseFFmpegDetect(msg: string): LPhaseLogId {
  const n = msg.trimStart();
  const first = n.split(/\s+/)[0] ?? "";
  if (n.startsWith("Starting FFmpeg configure")) return "ff-configure";
  if (n.startsWith("Starting FFmpeg make")) return "ff-compile";
  if (n.startsWith("Starting FFmpeg artifact collection")) return "ff-extraction";
  if (n.startsWith("Emptying FFmpeg artifact directory")) return "ff-extraction";
  if (n.startsWith("Removed ") && n.includes(" stale FFmpeg artifact entries")) return "ff-extraction";
  if (LLogCompileOps.has(first)) return "ff-compile";
  if (LLogShaderOps.has(first)) return "ff-compile";
  if (LLogStripOps.has(first)) return "ff-compile";
  if (LLogBuildOps.has(first)) return "ff-compile";
  if (LLogDocsOps.has(first)) return "ff-compile";
  if (first === "GEN") return "ff-compile";
  if (n.startsWith("pkg-config check") || n.startsWith("Using pkg-config")) return "ff-pkgconfig";
  if (
    n.startsWith("reinstalling ") || n.startsWith("downgrading ") ||
    n.startsWith("upgrading ") || n.startsWith("installing ") ||
    n.startsWith("Refreshing package databases") || n.startsWith("Clearing half-downloaded") ||
    n.startsWith(":: Synchronizing") || n.startsWith(":: Processing") ||
    n.startsWith(":: Running post-transaction") || n.startsWith("checking") ||
    n.startsWith("loading package") || n.startsWith("looking for conflicting") ||
    n.startsWith("resolving dependencies") || n.startsWith("Packages (") ||
    n.startsWith("Net Upgrade") || n.startsWith("Total Installed") ||
    n.startsWith("updating font cache") || n.includes(" downloading...") ||
    n.startsWith("warning: ")
  ) return "ff-pkgconfig";
  if (
    n.startsWith("Approved FFmpeg build started") ||
    n.startsWith("Downloading approved file from FFmpeg") ||
    n.startsWith("Calculated SHA-256 for FFmpeg") ||
    n.startsWith("FFmpeg .asc verification") ||
    n.startsWith("Extracting approved archive")
  ) return "ff-download";
  if (n.startsWith("Approved FFmpeg build completed") || n.startsWith("Artifact report")) return "ff-extraction";
  if (
    n.startsWith("FFmpeg configure") || n.startsWith("Starting FFmpeg configure") ||
    n.startsWith("License:") || n.startsWith("Enabled ") || n.startsWith("External ") ||
    n.startsWith("Programs:") || n.startsWith("Libraries:") || n.startsWith("ARCH ") ||
    n.startsWith("C compiler") || n.startsWith("C library") ||
    n.startsWith("install prefix") || n.startsWith("source path") ||
    n.startsWith("static ") || n.startsWith("shared ") || n.startsWith("optimizations") ||
    n.startsWith("debug ") || n.startsWith("network") || n.startsWith("threading") ||
    n.startsWith("safe bitstream") || n.startsWith("x86 assembler") ||
    n.startsWith("standalone") || n.startsWith("runtime cpu") ||
    n.startsWith("big-endian") || n.startsWith("MMXEXT") || n.startsWith("MMX ") ||
    n.startsWith("SSE") || n.startsWith("AESNI") || n.startsWith("CLMUL") ||
    n.startsWith("AVX") || n.startsWith("XOP") || n.startsWith("FMA") ||
    n.startsWith("i686") || n.startsWith("CMOV") || n.startsWith("EBX") ||
    n.startsWith("EBP") || n.startsWith("optimize") || n.startsWith("experimental") ||
    n.startsWith("makeinfo") || n.startsWith("perl ") || n.startsWith("texi2html") ||
    n.startsWith("xmllint") || n.startsWith("pod2man")
  ) return "ff-configure";
  if (n.startsWith("PE DLL dependencies") || n.startsWith("DLL lookup index") || n.startsWith("dependency ")) return "ff-extraction";
  if (n.startsWith("Running approved command")) return "other";
  return "other";
}

// ─── Parsing and aggregation ──────────────────────────────────────────────────

export function LLogEntryParse(entry: LLogSecurityEntry, context: "toolchain" | "ffmpeg"): LLogParsedEntry {
  const msg = entry.message;
  const n = msg.trimStart();
  const phase = context === "toolchain" ? LPhaseToolchainDetect(msg) : LPhaseFFmpegDetect(msg);
  const parsed: LLogParsedEntry = { ...entry, phase };
  const compileMatch = n.match(/^(CC|CXX|HOSTCC|X86ASM|WINDRES|STRIP|AR|LDXX|LD|HOSTLD|GEN|BIN2C|GZIP|MINIFY|GLSLC|POD|HTML|TXT|TEXI|GENTEXI)\s+(.+)$/);
  if (compileMatch) { parsed.compileOp = compileMatch[1]; parsed.compileTarget = compileMatch[2]; }
  if (msg.startsWith("dependency ")) {
    const copiedMatch = msg.match(/^dependency (.+?): copied OK/);
    const skippedMatch = msg.match(/^dependency (.+?): already copied/);
    const foundMatch = msg.match(/^dependency (.+?): found at/);
    const systemMatch = msg.match(/^dependency (.+?): not in MSYS2 bin/);
    if (copiedMatch) { parsed.dllDep = copiedMatch[1]; parsed.dllAction = "copied"; }
    else if (skippedMatch) { parsed.dllDep = skippedMatch[1]; parsed.dllAction = "skipped"; }
    else if (foundMatch) { parsed.dllDep = foundMatch[1]; parsed.dllAction = "found"; }
    else if (systemMatch) { parsed.dllDep = systemMatch[1]; parsed.dllAction = "system"; parsed.isSystemDll = true; }
  }
  if (msg.startsWith("PE DLL dependencies for ")) {
    const m = msg.match(/^PE DLL dependencies for (.+?):/);
    if (m) parsed.dllName = m[1];
  }
  return parsed;
}

export function LPhaseGroupBuild(entries: LLogParsedEntry[], phaseOrder: LPhaseLogId[]): LPhaseLogGroup[] {
  const phaseMap = new Map<LPhaseLogId, LPhaseLogGroup>();
  for (const entry of entries) {
    if (!phaseMap.has(entry.phase)) {
      phaseMap.set(entry.phase, {
        phase: entry.phase, label: LPhaseLabelGet(entry.phase), entries: [],
        compileCount: 0, assembleCount: 0, copiedDlls: [], systemDllCount: 0, skippedDllCount: 0,
        startTime: entry.timestamp, endTime: entry.timestamp,
      });
    }
    const group = phaseMap.get(entry.phase)!;
    group.entries.push(entry);
    group.endTime = entry.timestamp;
    if (entry.compileOp === "CC" || entry.compileOp === "CXX" || entry.compileOp === "HOSTCC") group.compileCount++;
    if (entry.compileOp === "X86ASM") group.assembleCount++;
    if (entry.dllAction === "copied" && entry.dllDep) group.copiedDlls.push(entry.dllDep);
    if (entry.dllAction === "system") group.systemDllCount++;
    if (entry.dllAction === "skipped") group.skippedDllCount++;
  }
  return phaseOrder.filter((phase) => phaseMap.has(phase)).map((phase) => phaseMap.get(phase)!);
}

export function LProgressGet(entries: LLogSecurityEntry[], approvedActionStatus: string, context: "toolchain" | "ffmpeg"): LProgressLive {
  const phaseOrder = context === "toolchain" ? LPhaseToolchainOrder : LPhaseFFmpegOrder;
  const phaseSet = new Set<LPhaseLogId>(phaseOrder);

  if (entries.length === 0) {
    return { currentPhaseLabel: null, currentPhaseId: null, compileCount: 0, assembleCount: 0, copiedDllCount: 0, lastMessage: null, failureMessages: [], isComplete: false, hasFailed: false };
  }

  const parsed = entries.map((e) => LLogEntryParse(e, context));
  const groups = LPhaseGroupBuild(parsed, phaseOrder);

  const isComplete =
    approvedActionStatus === "completed" ||
    (context === "ffmpeg"
      ? parsed.some((e) => e.message.startsWith("Approved FFmpeg build completed"))
      : parsed.some((e) => e.message.startsWith("Approved private MSYS2 environment is ready")));

  // Failure is driven by the backend's authoritative run status, not by scraping
  // log lines. Tools like pacman emit non-fatal "error:" lines (e.g. a transient
  // mirror 404 on one repo's .db during sync) and then recover; those must not
  // flip the UI to "failed" while the run keeps going and ultimately completes.
  const hasFailed = !isComplete && approvedActionStatus === "failed";

  const IGNORED = context === "ffmpeg"
    ? ["Approved FFmpeg build completed", "Artifact report"]
    : ["Approved private MSYS2 environment is ready"];

  let currentPhaseId: LPhaseLogId | null = null;
  let currentPhaseIndex = -1;
  for (const e of parsed) {
    if (!phaseSet.has(e.phase)) continue;
    if (IGNORED.some((p) => e.message.startsWith(p))) continue;
    const idx = phaseOrder.indexOf(e.phase);
    if (idx > currentPhaseIndex) { currentPhaseId = e.phase; currentPhaseIndex = idx; }
  }

  const currentPhaseLabel = currentPhaseId ? LPhaseLabelGet(currentPhaseId) : null;

  const NOISY = ["dependency ", "PE DLL dependencies for ", "DLL lookup index"];
  let lastMessage: string | null = null;
  for (let i = parsed.length - 1; i >= 0; i--) {
    const msg = parsed[i].message;
    if (!NOISY.some((p) => msg.startsWith(p))) { lastMessage = msg; break; }
  }

  // Error-level lines so a failed run can show the cause inline instead of forcing
  // the user to open the Logs tab. The last error is usually a generic wrapper
  // ("... failed: exit status 1"), so keep the tail of error lines — the real cause
  // (e.g. pacman "error: failed retrieving file ...") sits just before it.
  const failureMessages = hasFailed
    ? parsed.filter((e) => e.level === "error").map((e) => e.message).slice(-6)
    : [];

  return {
    currentPhaseLabel,
    currentPhaseId,
    compileCount: groups.reduce((s, g) => s + g.compileCount, 0),
    assembleCount: groups.reduce((s, g) => s + g.assembleCount, 0),
    copiedDllCount: groups.reduce((s, g) => s + g.copiedDlls.length, 0),
    lastMessage,
    failureMessages,
    isComplete,
    hasFailed,
    phaseGroups: groups,
  };
}
