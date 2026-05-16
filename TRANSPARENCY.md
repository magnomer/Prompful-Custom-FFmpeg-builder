# Transparency rules for security-critical code

This project intentionally keeps security-critical names plain. A reviewer should be able to scan the code and tell what each operation can do without decoding abstract labels.

## Naming rule

Use direct names for dangerous actions:

| Area | Preferred words | Avoid hiding behind |
|---|---|---|
| Downloads | `DownloadPlan`, `DownloadMsys2WithConsent`, `DownloadFfmpegSourceWithConsent`, `AllowedHosts` | vague `Spec`, generic `process`, `transfer` |
| Signatures and hashes | `verifyMsys2DetachedSignature`, `verifyFfmpegDetachedSignature`, `ExpectedSha256Hash`, `ApprovedScriptSha256Hash` | vague `check`, `auth`, `validate` without saying what is checked |
| Extraction | `ExtractPlan`, `ExtractArchiveWithConsent`, `checkExtractTarget`, `MaximumExtractedByteCount` | vague `operation`, `materialize` |
| Command execution | `CommandPlan`, `RunCommandWithConsent`, `RunPacmanWithConsent`, `ExecutablePath` | vague `external activity`, `task`, `handler` |
| Generated scripts | `WriteScriptFile`, `ConfigureScriptLines`, `MakeScriptLines`, `PacmanInstallScriptLines` | hidden command strings or anonymous script blobs |
| Path safety | `CheckRealPathInsideWorkspace`, `CheckPathInsideWorkspace`, `RejectSymlinkComponents` | vague `validate`, `sanitize` without saying what is checked |
| Consent | `Consent`, `CheckConsent`, `ApprovalRequest`, `PlanReviewSession`, `CheckReviewApproval` | legalistic names that obscure the approval boundary |
| Libraries | `LibraryChoice`, `SelectedLibraries`, `RequiredMsys2PackageNames`, `GeneratedConfigureFlags`, `LicenseEffectName` | hidden library flags in raw manual text |
| Results and audit | `BuildResult`, `BuildResultFile`, `artifactFilesForReport`, `NewWriter`, `WriteEvent` | vague `output`, `data`, `record` |

## Security review map

A reviewer should start with these files:

1. `app.go` — public backend methods called by the UI, native confirmation, action startup, artifact copying, and result helpers.
2. `internal/reviewsession/reviewsession.go` — backend-owned review session creation, consent text hash, expiry, and approval checks.
3. `internal/consent/consent.go` — consent kinds, approval conversion, and `CheckConsent`.
4. `internal/planning/planner.go` — what actions, packages, flags, warnings, license effects, and hashes the backend plans before approval.
5. `internal/planning/types.go` — public plan/review shapes returned to the frontend.
6. `internal/download/download.go` — all normal network file downloads, host allowlists, file-size checks, conflict policies, and SHA-256 checks.
7. `internal/extraction/extraction.go` — archive format handling and archive unpacking restrictions.
8. `internal/execution/execution.go` — managed command execution, explicit MSYS2 environment construction, script verification, and stdout/stderr logs.
9. `internal/scripting/scripting.go` — generated shell scripts and their hashes.
10. `internal/workspace/workspace.go` — path containment, symlink rejection, and workspace directory layout.
11. `internal/audit/audit.go` — local JSONL evidence of approved actions.
12. `frontend/src/main.tsx` — UI flow, review display, library presets, saved state, logs, and result display. This file is not a security boundary.
13. `scripts/security-scan.go` and `scripts/security-scan.ps1` — static boundary checks.

## Code rule

Security-sensitive functions should answer one visible question in their name:

- `CheckReviewApproval`: did this request match the backend-owned review session?
- `CheckConsent`: did the user approve this exact consent kind, action name, and plan hash?
- `CheckRealPathInsideWorkspace`: after filesystem resolution, is this path still inside the workspace?
- `DownloadMsys2WithConsent`: download an MSYS2 file only after matching MSYS2 download consent.
- `DownloadFfmpegSourceWithConsent`: download an FFmpeg source/signature/key file only after matching FFmpeg source download consent.
- `ExtractArchiveWithConsent`: extract an archive only after matching archive extraction consent.
- `RunPacmanWithConsent`: run pacman installation only after matching package-install consent.
- `RunCommandWithConsent`: run configure/make only after matching command-execution consent.
- `WriteScriptFile`: write a generated script and record its SHA-256 hash.
- `copyFfmpegBuildOutputs`: copy only discovered build outputs and required DLL dependencies into the result folder.

## User-facing transparency

The UI must keep security-relevant information visible before approval:

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

The approval panel may request backend approval, but it must not present itself as the final security boundary. The backend-owned native dialog is the final approval step.

## Review checklist

Before release, search for these strings:

```text
exec.Command
http.NewRequest
http.Client
.Do(request)
os.RemoveAll
os.WriteFile
os.OpenFile
archive/tar
filepath.EvalSymlinks
CheckReviewApproval
CheckConsent
RunCommandWithConsent
RunPacmanWithConsent
WriteScriptFile
ApprovedScriptSha256Hash
```

Every hit should be explainable from the function name and nearby comments. If not, rename the function, split it into smaller pieces, or move it behind the correct reviewed boundary.

## Static scanner alignment

The static scanners are part of the transparency model. When source code legitimately needs a narrow exception, the exception should be visible in scanner configuration and explained in `SECURITY.md`.

Do not leave scanner failures ambiguous. A reviewer should be able to tell whether a hit is:

- a real boundary violation;
- a narrow, documented exception;
- dead code that should be removed;
- code that should be moved into a more specific internal package.
