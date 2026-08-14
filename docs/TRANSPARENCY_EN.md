# Build Process Transparency

Promptful Custom FFmpeg Builder does not bundle FFmpeg, codec libraries, or a build environment. Before a build starts, it prepares a local build plan in the selected workspace, presents that plan for review, runs only the operations that are approved through the confirmation process, and leaves local logs and reports for later inspection.

This document describes what information is shown during the build process, how that information is checked, and what records are kept locally.

## Information available for review

For operations that require review, the plan may show:

- the operation being approved;
- download URLs and verification information;
- archive extraction targets;
- MSYS2 packages to install;
- selected libraries and generated configure flags;
- user-supplied configure flags;
- generated scripts and script hashes;
- warnings and their severity;
- license;
- whether the plan modifies PATH, needs administrator rights, uses an existing MSYS2 installation, or deletes files.

The UI displays this information. Before execution, the backend validates the reviewed operation against its stored review session and requests the user's approval through a Windows native confirmation dialog.

## Plan review before execution

This program uses a plan-first workflow. It prepares a concrete plan first, then runs the build only after the user has reviewed and approved that plan. It does not start a command first and explain the result afterward.

The review step identifies the relevant consequences before any workspace-changing operation starts: which packages will be installed, which source archives will be downloaded, which configure flags will be used, which scripts will be generated, and which license profile will apply to the planned build.

## Local workspace

The program works inside the workspace selected by the user. Downloads, extracted files, generated scripts, logs, reports, private MSYS2 files, and copied build artifacts are checked so they remain inside that workspace.

Approval does not remove path validation. Approval means that the user approved the planned operation; workspace checks still restrict where that operation can write, extract, copy, or delete files.

## Downloads and source verification

Download steps are shown in the reviewed plan. The program records the URL, destination, and available verification information. During download, it checks HTTPS use, workspace-contained destinations, file-size limits, redirect limits, optional SHA-256 verification, and signature/key downloads where the plan includes them. MSYS2 package-install scripts also configure an ordered set of package mirrors and a bounded `curl` transfer command so a stalled mirror can fail over instead of leaving the build hanging indefinitely.

A previously downloaded file may be reused only when reuse is allowed by the plan and, when an expected hash exists, the existing file matches that hash.

## Generated scripts

FFmpeg builds require shell scripts for package installation, configure, make, and some prepared-library work. The program writes those scripts inside the selected workspace and records their SHA-256 hashes.

Before a generated script is executed, the execution layer verifies that the file still matches the approved hash. This keeps generated scripts within the reviewed build plan instead of treating them as untracked execution details.

Some compatibility handling happens inside generated configure scripts. For example, selected flags such as `--enable-libsvtjpegxs`, `--enable-liblensfun`, and `--enable-vapoursynth` may be probed and removed with a warning if the installed package does not expose the API expected by the selected FFmpeg source. FFmpeg 9 handling also adds the nested MSYS2 ONNX Runtime include directory when `--enable-libonnxruntime` is selected, rejects AMF headers older than 1.5.2.0, and rejects ffnvcodec/NVENC headers older than NVENC SDK 11.1 before configure. This is specific compatibility handling, not a general promise that every unavailable library can be repaired automatically.

## Execution plans

Package installation and build operations are run through managed execution plans. The program constrains executable paths, working directories, MSYS2 roots, log directories, executable names, and generated-script hashes.

Transient network exhaustion has its own terminal state. When the execution layer identifies only retryable network failures after exhausting its command retry budget, it emits a `stalled` event with the mirror/host addresses that were tried. The run is no longer treated as active, but it is also not recorded as an ordinary build failure. For an FFmpeg build, the UI can relaunch the same previously approved plan so cache-resumable work can continue.

Auxiliary actions are limited and fixed-purpose. Opening folders, opening reports, opening saved logs, checking installed packages, probing a produced FFmpeg binary, or stopping private MSYS2 helper processes are not treated as arbitrary shell execution.

## Logs, reports, and approval records

The program leaves several kinds of local records:

| Record | Purpose |
|---|---|
| Live process logs | Show stdout/stderr while a task is running |
| Saved local log records | Keep previous local logs available for later review, including the distinct `stalled` terminal state |
| Build report JSON | Records selected libraries, flags, license profile, output files, sizes, and SHA-256 hashes |
| Approval record JSONL | Records local approval and major operation events |

These files are stored locally. The build report focuses on the finished artifacts and configuration. The approval record focuses on the approved operation and major execution events.

## UI and backend responsibilities

The UI displays information, supports user selection, provides localization, remembers relevant UI state, and presents the review screen. The backend performs the build-related operations: plan creation, review-session validation, native confirmation, approval-token creation, downloads, extraction, package installation, command execution, report writing, log writing, and workspace checks.

This separation allows the UI to be rearranged or localized without changing what the backend is allowed to execute.

## Scope of transparency

The transparency described here applies to the build plan produced by this program. It does not make every upstream FFmpeg option selectable or supported.

No build information is sent to an external service. Logs, reports, and approval records are local files created so the user can inspect the build process.
