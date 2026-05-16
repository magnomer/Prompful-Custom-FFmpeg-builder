# CustomFFmpeg Builder

A Windows-focused Go + Wails GUI for building FFmpeg locally from user-approved source archives.

The application itself intentionally contains:

- no FFmpeg binary,
- no FFmpeg source archive,
- no codec library,
- no multimedia library,
- no generated FFmpeg build.

It is a consent-first build orchestrator. It downloads approved source/tool archives, verifies them, extracts them into a private workspace, runs approved build scripts, and copies build outputs into a local result folder.

## Current scope

This project profile is Windows-only. Planning is blocked on non-Windows hosts.

The default toolchain profile uses a private MSYS2 archive inside the selected workspace:

- default shell profile: `ucrt64`
- supported shell profiles: `ucrt64`, `mingw64`, `clang64`
- default MSYS2 archive URL: `https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst`
- default MSYS2 signature URL: `https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig`

The app accepts MSYS2 `.tar.zst` archives and `.tar.xz` archives. It intentionally does not run the MSYS2 `.exe` installer or self-extracting installer.

## Security model

The central rule is:

> Every function that downloads, extracts, installs, deletes, updates, or executes external tooling must require a dedicated user consent type in its signature.

Boolean consent values are forbidden.

The app uses this flow:

```text
Request plan
  -> backend stores immutable plan in a review session
  -> frontend displays dry-run operations, warnings, packages, flags, license effects, and consent text
  -> user requests approval for exact action name and plan hash
  -> backend validates review session, consent text hash, expiry, and plan hash
  -> backend shows native OS confirmation dialog
  -> backend creates action-specific consent values
  -> mutating functions verify consent and plan hash
  -> execution begins
```

The frontend is not a trusted security boundary. Final approval is the backend-owned native dialog.

## Implemented safeguards

- Plan-first public API; no direct `StartBuild(settings)` method.
- Backend-owned review sessions with action name, plan hash, consent text hash, expiry, and one-time use.
- Backend-owned native confirmation dialog after frontend approval request.
- Action-specific consent structs:
  - `Msys2DownloadConsent`
  - `FfmpegSourceDownloadConsent`
  - `ArchiveExtractionConsent`
  - `PacmanInstallConsent`
  - `CommandExecutionConsent`
  - `WorkspaceDeletionConsent`
- Stable plan hashes and backend recomputation before execution.
- HTTPS-only downloads in normal mode.
- Download host allowlists.
- Optional SHA-256 reuse policy: existing downloads are reused only when the expected hash matches.
- MSYS2 `.sig` detached signature verification with a built-in OpenPGP verifier.
- FFmpeg `.asc` detached signature verification with the FFmpeg release signing key.
- Safe archive extraction with path traversal checks.
- `.tar.zst`, `.tar.xz`, `.tar.gz`, `.tgz`, `.tar.bz2`, and `.tar` archive-format handling.
- Symlink and hardlink extraction blocked.
- Archive file-count, total-byte, and single-file-size limits.
- Controlled workspace layout.
- No system PATH modification by managed build steps.
- No admin-right requirement by default.
- No existing `C:\msys64` modification by default.
- External command execution isolated in `internal/execution` for managed build commands.
- HTTP download code isolated in `internal/download`.
- Generated Bash scripts are written to workspace files and verified by SHA-256 before execution.
- MSYS2 execution builds an explicit PATH from the private MSYS2 root and selected shell profile.
- Security log events are emitted to the UI.
- Approved action logs are written under `workspace/logs/<runId>/`.
- FFmpeg artifacts are copied into `workspace/FFmpeg` and hashed in the build report.
- The Result tab can show output files, hashes, final configure flags, selected libraries, selected configure options, and installed library packages.

## Workspace layout

For a selected workspace directory, the backend creates and checks these directories:

```text
workspace/
  cache/
    downloads/
  sources/
  build/
    scripts/
  prefix/
  FFmpeg/
  logs/
  toolchains/
    msys2/
```

Build outputs are copied to `workspace/FFmpeg`. Approved action logs are written to `workspace/logs/<runId>/`.

## Wails public backend API

Public backend methods are plan-first or read/open helpers:

```go
GetInitialApplicationState()
SelectWorkspace()
RequestToolchainPreparationPlan(...)
ApproveToolchainPreparationPlan(...)
RequestFfmpegBuildPlan(...)
ApproveFfmpegBuildPlan(...)
CancelApprovedAction()
GetBuildResult(...)
OpenResultFolder(...)
OpenExternalUrl(...)
```

Approval methods receive a review session id plus an approval request. They do not receive an executable plan back from the frontend.

## FFmpeg build flow

The FFmpeg build action currently performs these planned operations:

1. download the approved FFmpeg source archive;
2. download the matching `.asc` signature;
3. download the FFmpeg release signing key;
4. verify the detached signature;
5. extract the source archive into a private source directory;
6. require exactly one extracted source child directory;
7. install MSYS2 packages required by selected external libraries;
8. write and hash the approved configure script;
9. run configure through the private MSYS2 Bash executable;
10. write and hash the approved make script;
11. run make with the approved parallel job count;
12. copy `ffmpeg.exe`, `ffprobe.exe`, and required MSYS2 DLL dependencies into `workspace/FFmpeg`;
13. write `build-report-<runId>.json`.

## Library and options model

The Library page distinguishes included FFmpeg components from external libraries.

Included rows are checked and locked because they are built as part of a normal FFmpeg source build. External libraries are unchecked until selected and may add packages, configure flags, license effects, and warnings.

The Options page exposes common FFmpeg configure choices as named rows. Manual advanced flags remain an escape hatch and appear in Review before backend confirmation.

See `LIBRARY_MODEL.md` for the full model.

## Naming rule

Every name must expose its domain, type, and purpose without abbreviation.

### Go

- Packages: lowercase single concept, for example `consent`, `planning`, `execution`.
- Exported types: `PascalCase` nouns, for example `FfmpegBuildPlan`.
- Functions: verb-led names, for example `PlanFfmpegBuild`.
- Variables: lowerCamel with full domain wording, for example `userArchiveExtractionConsent`.
- Boolean fields: `Is...`, `Has...`, `Can...`, `Will...`, `Should...`.

### TypeScript

- Types and React components: `PascalCase`.
- Variables and functions: `lowerCamel`.
- Boolean variables: `is...`, `has...`, `can...`, `should...`, `will...`.

### CSS

Strict BEM only:

```css
.approval-panel {}
.approval-panel__title {}
.approval-panel__item--blocked {}
```

## Build

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend
npm install
cd ..
wails dev
```

## Verification

Run Go tests with:

```bash
go test ./...
```

Run the static security scan with:

```bash
go run scripts/security-scan.go
```

In this offline environment, `go test ./...` could not complete because Go dependencies were not available locally and network access to `proxy.golang.org` was blocked.

## Static security scan note

`scripts/security-scan.go` currently declares these strict boundaries:

- `exec.Command` only inside `internal/execution`;
- HTTP request creation and execution only inside `internal/download`;
- recursive deletion only inside `internal/workspace`;
- direct file writes only in scripting, audit, and app-level reporting code;
- raw network dialing, `os.StartProcess`, `syscall.Exec`, `unsafe`, and `bash -lc` shell-string execution forbidden.

Before release, keep the scanner and code aligned. If an app-level helper needs a narrowly reviewed exception, the scanner must document that exception explicitly instead of relying on an accidental bypass.
