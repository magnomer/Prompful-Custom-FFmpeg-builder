# Consent boundary

This program treats the frontend as a presentation layer, not as the final authority for approving work.

The backend creates executable plans, stores in-memory review sessions, validates approval requests, opens the backend-owned native confirmation dialog, and only then creates the action-specific consent values used by the approved backend operation.

Consent is one boundary layer. It is paired with plan-hash verification, generated-script hash verification, workspace containment checks, symlink checks, archive limits, fixed command allowlists, and audit/log records.

## Where consent fits in the program

The program's mutating build flow is plan-first:

1. The backend creates an executable plan.
2. The backend stores that plan in an in-memory review session.
3. The frontend displays the plan, warnings, final configure flags, package list, license effects, and consent text.
4. The frontend sends an approval request containing:
   - review session id,
   - approved action name,
   - approved plan hash,
   - exact consent text.
5. The backend retrieves the stored plan from the review session.
6. The backend checks that the approval request still matches the stored review session:
   - session exists,
   - action name matches,
   - plan hash matches,
   - consent text hash matches,
   - session has not expired,
   - session has not already been consumed.
7. The backend recomputes the stored plan hash from the stored plan content.
8. The backend checks that the stored plan is executable.
9. The backend opens a native OS confirmation dialog.
10. The backend creates action-specific consent values only after the native dialog returns `Yes`.
11. The backend starts the approved action using the stored plan and matching consent values.

## Backend-owned native confirmation

Approval RPC calls are treated as requests to approve, not as final approval.

The backend starts mutating work only after the stored review session is valid, the stored plan is executable, the recomputed plan hash still matches, the native confirmation dialog returns `Yes`, and the target function receives its matching action-specific consent type.

The native dialog is intentionally owned by the backend rather than by the WebView frontend. This keeps a forged frontend message, injected JavaScript call, replayed local state value, WebView devtool call, or direct RPC call from becoming final proof of user intent.

## Rejected confirmation results

The native dialog is configured so the safe result is `No`.

The backend treats every result other than `Yes` as rejection, including `No`, `Cancel`, Escape, window close, dialog errors, empty results, and unexpected button strings.

A rejected or cancelled native dialog does not consume the review session. The same review session can be retried until it expires. The review session is consumed only after the backend native confirmation succeeds.

## Review session lifetime and storage

Review sessions are single-use after successful native confirmation. Reusing the same review session id after it has been consumed fails.

Current review sessions have a 30-minute lifetime. Expired sessions are rejected even when the action name, plan hash, and consent text still match.

Review sessions are stored in memory. Closing or restarting the app invalidates outstanding review sessions.

## Approved action flows

The approval request is converted into narrow consent values inside backend approval methods. A single approved plan can create several consent values that share the same approved action name, approved plan hash, and consent text.

| Approved action | Consent values created after native `Yes` |
| --- | --- |
| Private MSYS2 toolchain preparation | `Msys2DownloadConsent`, `ArchiveExtractionConsent`, `PacmanInstallConsent` |
| FFmpeg build | `FfmpegSourceDownloadConsent`, `ArchiveExtractionConsent`, `PacmanInstallConsent`, `CommandExecutionConsent` |

The MSYS2 download consent also covers approved MSYS2 signature and signing-key downloads that belong to the same toolchain preparation plan.

The FFmpeg source download consent also covers approved FFmpeg signature and signing-key downloads, and approved source downloads for prepared libraries that belong to the same FFmpeg build plan.

## Current consent types

Current active consent types are:

- `Msys2DownloadConsent`
- `FfmpegSourceDownloadConsent`
- `ArchiveExtractionConsent`
- `PacmanInstallConsent`
- `CommandExecutionConsent`

The code also defines `WorkspaceDeletionConsent`, but the current cleanup and removal paths do not use it as an active approval boundary. It is reserved/defined code, not part of the current approval flow.

The code uses typed values instead of plain booleans so each guarded operation can check the exact consent kind, action name, and plan hash it was approved for.

## Operations guarded by consent

The current active consent boundary guards the mutating build/install operations below:

- MSYS2 archive, signature, and signing-key downloads through `DownloadMsys2WithConsent`.
- FFmpeg source archive, signature, signing-key, and prepared-library source downloads through `DownloadFfmpegSourceWithConsent`.
- Approved archive extraction through `ExtractArchiveWithConsent`.
- Approved package installation through `RunPacmanWithConsent`.
- Approved configure, make, and prepared-library script execution through `RunCommandWithConsent`.

These functions check the supplied consent kind, approved action name, and approved plan hash before performing the operation.

## Operations outside the active consent boundary

Not every backend command or file operation is routed through a typed consent value.

Non-mutating or support operations are outside the current consent boundary and are constrained by fixed arguments and path checks instead. Examples include:

- opening result folders, reports, log folders, or local log files;
- reading package state from the private MSYS2 environment;
- verifying built `ffmpeg.exe` and `ffprobe.exe` with fixed inspection commands;
- stopping private MSYS2 background agents during cleanup.

Cleanup/removal paths are also outside `WorkspaceDeletionConsent` in the current code. They run as consequences of an already approved operation or an explicit backend action, and they rely on workspace containment, real-path checks, root-refusal checks, symlink checks, and narrowly constructed target paths.

## Cleanup and removal boundary

Current cleanup is action-scoped and workspace-contained, not separately consent-gated.

Examples include:

- failed toolchain preparation cleanup;
- failed FFmpeg build cleanup;
- removal of stale FFmpeg artifact files before copying the new build result;
- removal of a previous private MSYS2 toolchain directory when preparing a fresh one;
- explicit removal of the private toolchain installation directory.

These paths must not delete arbitrary user paths. They check that targets are inside the selected workspace, reject dangerous roots where applicable, avoid following unsafe symlinks where applicable, and construct deletion targets from the workspace layout rather than from free-form frontend input.

## Download boundary

Approved downloads are still checked at execution time. The download layer verifies that:

- the URL uses HTTPS;
- the destination path is inside the selected workspace;
- the real destination path stays inside the selected workspace;
- partial download paths stay inside the selected workspace;
- expected file size minimum and maximum limits are respected;
- an existing destination file is reused only when the plan allows reuse and the expected SHA-256 hash matches;
- an expected SHA-256 hash is checked when the plan provides one.

Host allowlists are warning-oriented in some download paths. A non-allowlisted host is reported as a warning, but it is not by itself always a hard block. The stronger enforcement comes from HTTPS, destination containment, size limits, and the hash/signature checks defined by the plan.

## Archive extraction boundary

Approved extraction is still checked at execution time. The extraction layer verifies that:

- the archive file is inside the selected workspace;
- the real archive file path is inside the selected workspace;
- the extraction destination is inside the selected workspace;
- the destination policy is satisfied;
- archive file count, total extracted bytes, and single-file bytes stay under limits;
- extracted paths do not escape the destination;
- unsafe symlink behavior is rejected.

Extraction is therefore not controlled by consent alone. Consent authorizes the operation class for the approved plan; extraction validation constrains what the archive can write.

## Command execution boundary

Mutating build commands are consent-gated through `RunPacmanWithConsent` and `RunCommandWithConsent`.

Command execution is additionally constrained by command-plan validation:

- executable path must be inside the workspace;
- real executable path must be inside the workspace;
- working directory must be inside the workspace;
- real working directory must be inside the workspace;
- MSYS2 root directory, when present, must be inside the workspace;
- run-log directory, when present, must be inside the workspace;
- executable basename must be in the command plan's allowlist;
- executable path must not contain shell metacharacters;
- arguments must not contain null bytes.

These checks are meant to prevent an approved build step from silently turning into an arbitrary command outside the private workspace.

## Generated script boundary

Generated scripts are part of the executable boundary.

When the backend prepares a command that runs a generated script, the command plan records:

- the script kind;
- the approved script file path;
- the approved script SHA-256 hash.

Before execution, the command layer verifies that the script file is inside the workspace, verifies that the real script path is inside the workspace, checks that the script path appears in the command arguments, reads the script content, recomputes the SHA-256 hash, and rejects execution if the content no longer matches the approved hash.

After that check, the command layer replaces the script-file argument with stdin execution. This prevents the command from silently running a different script path after the approved script content has been checked.

## Reason for this design

This design separates presentation from authority.

The frontend can display the plan and request approval, but the backend owns the stored plan, plan hash, expected consent text, native confirmation, consent creation, and final execution checks.

This also blocks approval replay: the backend stores review sessions, checks expiry, and removes a session after successful native confirmation.
