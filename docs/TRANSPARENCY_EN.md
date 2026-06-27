# Transparency model

This document explains how Promptful Custom FFmpeg Builder keeps sensitive build behavior visible in the source tree, in the review flow, and in the files written during a build.

The program builds FFmpeg by downloading source archives, unpacking them, installing MSYS2 packages, generating shell scripts, running build commands, copying build artifacts, and writing logs and reports. Those actions are intentionally named and placed so that a source reviewer can find the risky operations without first learning a private vocabulary.

Transparency in this project does not mean that every operation is safe merely because it is visible. It means the program tries to make important operations explicit, reviewable, and locally auditable.

## 1. What the transparency model covers

The transparency model has four layers:

| Layer | What it exposes | Where it appears |
|---|---|---|
| Source layout | Which package owns downloads, extraction, planning, command execution, scripts, workspace checks, and audit writes | Go packages under `internal/`, frontend files under `frontend/src/`, and scanner scripts under `scripts/` |
| Review data | What the program plans to download, install, prepare, configure, run, and copy | Backend review objects displayed by the frontend before approval |
| Runtime evidence | What happened during an approved action | live logs, saved log records, generated reports, and local audit JSONL |
| Boundary checks | Which sensitive primitives are expected in which source areas | `scripts/security-scan.go` and `scripts/security-scan.ps1` |

The frontend is part of the transparency layer because it displays plans, warnings, logs, and reports. It is not the security boundary by itself. Final approval and consent checks are backend-owned.

## 2. Main source areas

The current source tree is split by responsibility rather than by a single large backend file.

| Area | Main files | Responsibility |
|---|---|---|
| App orchestration | `internal/app/*.go` | Wails-facing methods, review-session storage, native confirmation, toolchain preparation, FFmpeg build execution, library preparation, artifact copying, log records, verification helpers, and UI/window state files |
| Consent | `internal/consent/consent.go` | typed consent values and consent checks for approved action plans |
| Review sessions | `internal/reviewsession/reviewsession.go` | short-lived backend-owned review sessions, consent text hashing, expiry, and approval matching |
| Planning | `internal/planning/*.go` | toolchain and FFmpeg build plans, library catalog, package lists, configure flags, warnings, license derivation, conflict validation, and plan hashes |
| Library-source metadata | `shared/librarysources/*` | source/import metadata for libraries prepared outside normal MSYS2 package selection |
| Download | `internal/download/download.go` | HTTPS downloads, host allowlist warnings, destination checks, file-size limits, conflict policy, and optional SHA-256 checks |
| Signature verification | `internal/app/signature.go` | detached signature verification for MSYS2 and FFmpeg source downloads |
| Extraction | `internal/extraction/extraction.go` | archive extraction and extraction limits |
| Execution | `internal/execution/execution.go` | managed build-command execution, explicit MSYS2 environment construction, script-hash verification, and stdout/stderr capture |
| Scripting | `internal/scripting/scripting.go` | generated shell script text, script files, and script SHA-256 hashes |
| Workspace | `internal/workspace/workspace.go` | workspace layout, path containment, real-path checks, and symlink rejection |
| Audit | `internal/audit/audit.go` | local JSONL action evidence for approved runs |
| Frontend | `frontend/src/**` | user-facing pages, presets, review panels, logs, result display, localization, and saved UI state |
| Static scanner | `scripts/security-scan.go`, `scripts/security-scan.ps1` | text-based checks for sensitive source primitives |

`internal/app/*.go` is the application boundary between the UI and the backend. It is intentionally broader than the lower-level packages because it coordinates review sessions, native confirmation, action startup, cancellation, reports, logs, and user-facing helper actions.

## 3. Naming model

Security-sensitive functions and types use operation-oriented names. A reader should be able to tell what the function does from its name.

| Area | Names used in the code | What the names expose |
|---|---|---|
| Downloads | `DownloadPlan`, `DownloadMsys2WithConsent`, `DownloadFfmpegSourceWithConsent`, `AllowedHosts` | what is downloaded, where it may come from, and which consent kind is required |
| Signatures and hashes | `verifyMsys2DetachedSignature`, `verifyFfmpegDetachedSignature`, `ExpectedSha256Hash`, `ApprovedScriptSha256Hash` | which signature, file hash, or script hash is being checked |
| Extraction | `ExtractPlan`, `ExtractArchiveWithConsent`, `checkExtractTarget`, `MaximumExtractedByteCount` | which archive is extracted, where it is extracted, and what limits apply |
| Command execution | `CommandPlan`, `RunCommandWithConsent`, `RunPacmanWithConsent`, `ExecutablePath` | which command plan is approved and which executable path is used |
| Generated scripts | `WriteScriptFile`, `ConfigureScriptLines`, `MakeScriptLines`, `PacmanInstallScriptLines` | which shell script text is generated before execution |
| Path safety | `CheckRealPathInsideWorkspace`, `CheckPathInsideWorkspace`, `RejectSymlinkComponents` | which workspace-containment or symlink check is being applied |
| Consent and review | `Consent`, `CheckConsent`, `ApprovalRequest`, `PlanReviewSession`, `CheckReviewApproval` | where UI review data becomes backend-owned approval state |
| Libraries | `LibraryChoice`, `SelectedLibraries`, `RequiredMsys2PackageNames`, `GeneratedConfigureFlags`, `LicenseEffectName` | which selected libraries affect packages, configure flags, and license profile |
| Results and audit | `BuildResult`, `BuildResultFile`, `artifactFilesForReport`, `NewWriter`, `WriteEvent` | which files are produced and which action events are recorded |

The project avoids using vague wrappers for operations such as downloading, extraction, command execution, deletion, or file writes. When such operations are used, their function names and package locations should make the operation visible.

## 4. Review and approval visibility

Before a mutating build action starts, the program creates a plan and shows it to the user. The review data is meant to expose the important effects of the action before backend approval.

A review can include:

- action name;
- plan hash;
- expected consent text;
- operation list;
- warnings grouped by severity;
- download URLs and verification expectations;
- MSYS2 package names;
- selected library integrations;
- selected configure options;
- generated library configure flags;
- extra manual flags;
- final configure flags;
- generated scripts and script hashes where applicable;
- derived license profile;
- whether the plan modifies PATH, requires admin rights, uses an existing MSYS2 installation, or deletes files.

The frontend displays this information and sends an approval request. The backend then checks the request against a backend-owned review session. A native Windows confirmation dialog is the final approval step before the backend creates typed consent values and starts the action.

The review session is not merely a frontend state object. The backend owns the session, the plan hash, and the consent text hash. The frontend cannot grant consent by keeping a checkbox or local state alone.

## 5. Consent-gated commands and support commands

The project separates mutating build/install commands from fixed-purpose support commands.

Mutating build/install commands use typed consent and the managed execution layer. Examples include:

- installing required MSYS2 packages with `RunPacmanWithConsent`;
- running generated FFmpeg `configure` scripts with `RunCommandWithConsent`;
- running generated `make` scripts with `RunCommandWithConsent`;
- running library-preparation commands that are part of an approved build plan.

Some commands are support actions rather than build-plan commands. They are not described as `CommandExecutionConsent` operations. Examples include:

- opening the result folder;
- opening a generated report;
- opening the log folder;
- opening a saved local log file;
- querying installed MSYS2 packages with fixed `pacman` arguments;
- verifying a built FFmpeg binary with fixed arguments such as `-version`, `-encoders`, or related probes;
- stopping private MSYS2 helper processes during cleanup.

Those support actions should still remain visible in `internal/app/*.go`, use fixed arguments where possible, and be constrained by app-owned paths or workspace checks. They should not be casually mixed with arbitrary user-provided shell commands.

## 6. Generated-script transparency

FFmpeg builds require shell scripts. The program does not treat those scripts as invisible implementation details.

The scripting layer generates script text for package installation, configure, make, and selected preparation tasks. The app writes generated scripts into app-owned workspace locations, records their SHA-256 hashes, and passes the expected hash into the managed execution layer.

Before executing a generated script, the execution layer verifies that the script file still matches the approved SHA-256 hash. For build scripts, execution is routed through a controlled MSYS2 Bash environment. This makes the generated script content, script path, and script hash part of the reviewable build boundary.

Some compatibility fallbacks are implemented inside generated scripts rather than inside the planner. For example, selected flags such as `--enable-libsvtjpegxs`, `--enable-liblensfun`, and `--enable-vapoursynth` can be probed in the generated configure script and removed with a warning if the installed package does not expose the API expected by the selected FFmpeg source. That is a script-time compatibility path, not a general planner rule for every unavailable library.

## 7. Download, extraction, and workspace visibility

Downloads are implemented in the download layer and are paired with plan data. The download boundary includes:

- HTTPS-only source URLs;
- destination checks inside the workspace;
- file-size limits;
- redirect limits;
- optional SHA-256 verification;
- reuse only when an existing file matches the expected hash where a hash is supplied;
- host allowlist warnings for known source families.

Signature and key downloads are treated as part of the approved source/toolchain download action. For example, FFmpeg source archive downloads, `.asc` downloads, and signing-key downloads use the FFmpeg source-download consent boundary.

Extraction is handled by the extraction layer. It checks the destination, archive format, extracted-byte limits, and path traversal risks. Workspace checks and symlink rejection are a second layer beside typed consent: they restrict where downloaded, extracted, generated, copied, and deleted files can land.

## 8. Logs, reports, and audit records

The program has several evidence layers. They should not be treated as the same thing.

| Evidence layer | Purpose |
|---|---|
| Live process logs | show stdout/stderr while an action is running |
| Saved local log records | preserve previous local logs and categorize them for later review |
| Build report JSON | records build result metadata such as selected libraries, flags, license profile, output files, sizes, and SHA-256 hashes |
| Audit JSONL | records local action evidence for approved runs |

Audit files are local evidence files. They are not a remote telemetry channel. Their purpose is to record what action was approved, which plan was used, and which major operation events happened locally.

The result report and audit log are complementary. The report focuses on the finished build artifacts and build configuration. The audit log focuses on approved action execution events.

## 9. Frontend and backend responsibility

The frontend is responsible for display, selection, localization, saved UI state, and review presentation. It can request backend actions, but it does not own the security decision.

The backend owns:

- plan creation;
- review-session creation and expiry;
- consent text hashing;
- plan-hash validation;
- native confirmation;
- typed consent creation;
- download, extraction, package installation, command execution, and artifact/report writes;
- workspace path checks;
- cancellation and cleanup behavior.

This split matters because the UI can become stale, localized, or rearranged without changing what the backend is allowed to execute. The backend must still validate the approved plan and consent boundary.

## 10. Static boundary scanner

The static scanner is a small text-based source review tool. It searches Go source files for sensitive primitives such as:

- process startup;
- HTTP client/request use;
- recursive deletion;
- file deletion;
- file renaming;
- direct file writes;
- shell-string execution;
- unsafe Go operations.

The scanner compares each hit against an allowlist of source locations that are expected to own that operation. This is useful because it can show when a sensitive primitive appears in a surprising location.

However, the scanner is not a formal proof system. Its allowlist must be kept in sync with the actual source layout. The current source tree has split app orchestration into `internal/app/*.go`, while the scanner rules still use older path assumptions in several places. Therefore, a scanner mismatch can mean either:

1. a real source-boundary problem; or
2. scanner rule drift after an intentional source-layout change.

At the time this document was revised, the scanner should be treated as an intended boundary-review tool whose allowlist needs maintenance before a passing result can be used as evidence that the source tree fully matches the documented boundary.

## 11. How to read boundary changes

When reviewing a source change, the most important transparency questions are:

1. Does a new download, extraction, command, deletion, write, or rename use the package that normally owns that operation?
2. If it is a mutating build/install operation, is it tied to a reviewed plan and typed consent?
3. If it is a support command, are the arguments fixed and the target paths constrained?
4. Are generated scripts written, hashed, and verified before execution?
5. Are workspace containment and symlink checks still applied around filesystem changes?
6. Does the review UI expose the packages, libraries, flags, license effects, warnings, and deletion/admin/PATH indicators that matter to the user?
7. Are logs, reports, and audit records still local and understandable?
8. If the static scanner reports a mismatch, is it a real boundary expansion or a stale scanner rule?

A source change that introduces a sensitive primitive in a new place should either move the operation into the existing boundary package, or update both the scanner and the transparency documentation so the new boundary is explicit.
