# Code structure and naming

This document describes the naming and source-layout style used by Custom FFmpeg Builder.

Security-sensitive code uses direct names so ordinary source browsing reveals what each operation can do. The code avoids hiding downloads, extraction, command execution, path checks, consent checks, and artifact writes behind vague labels.

## Naming model

The code uses operation-oriented names for security-sensitive behavior:

| Area | Names used in the code | What the names expose |
|---|---|---|
| Downloads | `DownloadPlan`, `DownloadMsys2WithConsent`, `DownloadFfmpegSourceWithConsent`, `AllowedHosts` | what is downloaded, where it may come from, and which consent is required |
| Signatures and hashes | `verifyMsys2DetachedSignature`, `verifyFfmpegDetachedSignature`, `ExpectedSha256Hash`, `ApprovedScriptSha256Hash` | which signature or hash is being checked |
| Extraction | `ExtractPlan`, `ExtractArchiveWithConsent`, `checkExtractTarget`, `MaximumExtractedByteCount` | which archive is extracted and what limits apply |
| Command execution | `CommandPlan`, `RunCommandWithConsent`, `RunPacmanWithConsent`, `ExecutablePath` | which executable path and command plan are used |
| Generated scripts | `WriteScriptFile`, `ConfigureScriptLines`, `MakeScriptLines`, `PacmanInstallScriptLines` | which script text is generated and later verified |
| Path safety | `CheckRealPathInsideWorkspace`, `CheckPathInsideWorkspace`, `RejectSymlinkComponents` | which workspace-containment check is being applied |
| Consent | `Consent`, `CheckConsent`, `ApprovalRequest`, `PlanReviewSession`, `CheckReviewApproval` | where approval data becomes typed backend consent |
| Libraries | `LibraryChoice`, `SelectedLibraries`, `RequiredMsys2PackageNames`, `GeneratedConfigureFlags`, `LicenseEffectName` | which selected libraries affect packages, flags, and license profile |
| Results and audit | `BuildResult`, `BuildResultFile`, `artifactFilesForReport`, `NewWriter`, `WriteEvent` | which files are produced and which events are recorded |

## Main code paths

The main app code paths are:

1. `app.go` — public backend methods called by the UI, native confirmation, action startup, artifact copying, and result helpers.
2. `internal/reviewsession/reviewsession.go` — backend-owned review session creation, consent text hash, expiry, and approval checks.
3. `internal/consent/consent.go` — consent kinds, approval conversion, and `CheckConsent`.
4. `internal/planning/planner.go` — action planning, packages, flags, warnings, license effects, and hashes.
5. `internal/planning/types.go` — public plan and review shapes returned to the frontend.
6. `internal/download/download.go` — normal network file downloads, host allowlists, file-size checks, conflict policies, and SHA-256 checks.
7. `internal/extraction/extraction.go` — archive format handling and archive unpacking restrictions.
8. `internal/execution/execution.go` — managed command execution, explicit MSYS2 environment construction, script verification, and stdout/stderr logs.
9. `internal/scripting/scripting.go` — generated shell scripts and their hashes.
10. `internal/workspace/workspace.go` — path containment, symlink rejection, and workspace directory layout.
11. `internal/audit/audit.go` — local JSONL evidence of approved actions.
12. `frontend/src/main.tsx` — UI flow, review display, library presets, saved state, logs, and result display. This file is a display/request layer rather than a security boundary.
13. `scripts/security-scan.go` and `scripts/security-scan.ps1` — static boundary checks for sensitive source patterns.

## Function-name shape

Security-sensitive functions are named around the operation they perform:

- `CheckReviewApproval` checks whether an approval request matches the backend-owned review session.
- `CheckConsent` checks whether the user approved the exact consent kind, action name, and plan hash.
- `CheckRealPathInsideWorkspace` checks whether a resolved filesystem path is still inside the workspace.
- `DownloadMsys2WithConsent` downloads an MSYS2 file after matching MSYS2 download consent.
- `DownloadFfmpegSourceWithConsent` downloads an FFmpeg source, signature, or key file after matching FFmpeg source download consent.
- `ExtractArchiveWithConsent` extracts an archive after matching archive extraction consent.
- `RunPacmanWithConsent` runs pacman installation after matching package-install consent.
- `RunCommandWithConsent` runs configure or make after matching command-execution consent.
- `WriteScriptFile` writes a generated script and records its SHA-256 hash.
- `copyFfmpegBuildOutputs` copies discovered build outputs and required DLL dependencies into the result folder.

## User-facing review data

The UI keeps security-relevant information visible before backend approval:

- action name;
- plan hash;
- expected consent text;
- operation list;
- warnings grouped by severity;
- MSYS2 package names;
- selected external libraries;
- selected configure options;
- generated library flags;
- extra manual flags;
- final configure flags;
- derived license profile;
- whether the plan modifies PATH, requires admin rights, uses existing MSYS2, or deletes files.

The approval panel sends approval requests to the backend. The backend-owned native dialog is the final approval step.

## Static boundary scanner

The static scanners are part of the code-structure model. They search the source tree for sensitive primitives and compare each hit with the package boundary used by the app.

The scanned patterns include command execution, HTTP download primitives, direct deletion, recursive deletion, direct file writes, file renaming, archive handling, path resolution, review approval, consent checks, generated script writes, and script hash verification.

When the scanner reports a hit, it means the source tree contains a sensitive primitive outside the currently described boundary, or that the scanner boundary needs to be updated to match an intentional source layout change.
