# CustomFFmpeg Builder

A Windows-only Go + Wails GUI scaffold for building FFmpeg locally from user-approved source archives.

The application itself intentionally contains:

- no FFmpeg binary,
- no FFmpeg source archive,
- no codec library,
- no multimedia library,
- no generated FFmpeg build.

It is a consent-first build orchestrator.

## Security model

The central rule is:

> Every function that downloads, extracts, installs, deletes, updates, or executes external tooling must require a dedicated user consent type in its signature.

Boolean consent values are forbidden.

The app uses this flow:

```text
Request plan
  -> show dry-run operations and warnings
  -> user approves exact plan hash
  -> backend creates action-specific consent values
  -> mutating functions verify consent and plan hash
  -> execution begins
```

## Implemented safeguards

- Plan-first public API; no direct `StartBuild(settings)` method.
- Action-specific consent structs:
  - `Msys2DownloadConsent`
  - `FfmpegSourceDownloadConsent`
  - `ArchiveExtractionConsent`
  - `PacmanInstallConsent`
  - `CommandExecutionConsent`
  - `WorkspaceDeletionConsent`
- Stable plan hashes.
- Mandatory SHA-256 verified downloads.
- HTTPS-only downloads.
- Download host allowlists.
- Safe archive extraction with path traversal checks.
- Symlink and hardlink extraction blocked.
- Controlled workspace layout.
- No system PATH modification.
- No admin-right requirement by default.
- No existing `C:\msys64` modification by default.
- External command execution isolated in `internal/execution`.
- HTTP download code isolated in `internal/download`.
- Security log events are emitted to the UI.
- Artifact report is written after FFmpeg build.

## Wails public backend API

Only plan-first methods are exposed:

```go
RequestToolchainPreparationPlan(...)
ApproveToolchainPreparationPlan(...)
RequestFfmpegBuildPlan(...)
ApproveFfmpegBuildPlan(...)
CancelApprovedAction()
```

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

## Important implementation note

This scaffold supports `.tar.xz` extraction through `github.com/ulikunitz/xz` because MSYS2 base archives are commonly distributed as `.tar.xz` files.

## Build

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend
npm install
cd ..
wails dev
```

## Verification performed in this environment

The following internal packages were type-checked successfully:

```bash
go test ./internal/consent ./internal/planning ./internal/workspace ./internal/execution ./internal/download
```

A full Wails build was not run here because the environment has no network access for downloading Go/npm dependencies.

## Backend hardening notes

This revision intentionally keeps the consent model unchanged so that a stricter consent-session design can be introduced later. Non-consent backend hardening was added around downloads, extraction, command execution, validation, and reporting.

Run the static security scan with:

```bash
go run scripts/security-scan.go
```

The scanner rejects direct command execution outside `internal/execution`, direct HTTP download primitives outside `internal/download`, and `bash -lc` shell-string execution.

## Security changes in v3

- Approval now uses backend-owned review sessions. The frontend no longer sends an executable plan back during approval.
- MSYS2 command execution now builds an explicit `PATH` from the private MSYS2 root and selected shell profile.
- Archive extraction now has file-count and byte-count limits and stricter archive path normalization.
- FFmpeg artifacts are copied into `workspace/artifacts` and hashed in the build report.
- Approved action logs are written to `workspace/logs/<runId>/`.
- The static security scanner checks network, process, deletion, write, and shell execution boundaries.
