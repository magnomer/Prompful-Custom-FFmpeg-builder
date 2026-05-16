# Security model

This document describes how Custom FFmpeg Builder separates planning, approval, download, extraction, command execution, workspace access, logging, and result reporting.

The app is a local FFmpeg build tool. It prepares a private MSYS2-based build environment, downloads reviewed source/tool archives, extracts them into a workspace, installs selected packages, runs generated build scripts, and copies the resulting FFmpeg files into a result folder.

## App trust boundary

The frontend displays choices and review data. The backend owns executable plans, review sessions, native confirmation, typed consent values, mutating filesystem operations, downloads, extraction, package installation, command execution, and result reporting.

The frontend may request approval, but the backend decides whether the request matches a stored review session and whether native confirmation was accepted.

## Plan-first backend flow

The backend plans work before it performs work.

A reviewed plan contains the action name, selected shell profile, workspace layout, package names, generated scripts, selected libraries, configure options, manual flags, final configure flags, warning list, license profile, and plan hash.

The frontend receives review data derived from that plan. Approval calls send a review session id and approval request, not a replacement executable plan.

## Backend review sessions

The backend stores every executable plan it creates during review.

Approval checks compare the incoming approval request with the stored session:

- review session id,
- action name,
- plan hash,
- consent text hash,
- expiry,
- consumed state.

The backend also recomputes the stored plan hash from stored content before it creates consent values or starts the action.

## Native confirmation

The backend opens a native OS confirmation dialog after the review session check passes.

The app treats the native dialog as the final approval step. Any result other than `Yes` cancels the action.

## Typed consent values

Mutating backend operations receive narrow consent values rather than plain booleans.

Current consent types cover:

- MSYS2 download,
- FFmpeg source download,
- archive extraction,
- pacman package installation,
- build command execution,
- workspace deletion.

Each consent value carries the action kind, action name, and plan hash it applies to. The receiving function checks that the consent value matches its operation.

## Downloads

Normal HTTP download code lives in `internal/download`.

Download plans include the source URL, expected destination path, expected hash when available, size limits, conflict behavior, allowed hosts, and action-specific consent.

Normal-mode downloads use HTTPS. FFmpeg source downloads are verified with the matching `.asc` PGP signature before extraction. MSYS2 archive downloads use detached signature verification when a signature URL is supplied.

## Archive extraction

Archive extraction code lives in `internal/extraction`.

Extraction plans describe the source archive, target directory, allowed formats, byte limits, entry-count limits, and archive-extraction consent.

The extractor checks workspace containment, rejects absolute paths, rejects parent-directory traversal, rejects symlink and hardlink restoration, and rejects extraction targets outside the selected workspace.

Currently accepted source archive formats include:

- `.tar.zst`
- `.tar.xz`
- `.tar.gz`
- `.tgz`
- `.tar.bz2`
- `.tar`

MSYS2 `.tar.zst` is the preferred archive format. `.tar.xz` is accepted as a fallback. MSYS2 installer formats such as `.exe` and `.sfx.exe` are not part of the planned extraction path.

## Command execution

Managed build commands run through `internal/execution`.

A command plan contains:

- action name,
- plan hash,
- executable path,
- arguments,
- working directory,
- workspace directory,
- private MSYS2 root,
- selected Windows shell profile,
- environment variables,
- executable basename allowlist,
- script kind,
- approved script file path,
- approved script SHA-256,
- run log directory.

Before execution, the backend checks workspace containment, resolved real paths, executable basename allowlist, script kind, and script hash.

The runtime environment builds an explicit MSYS2 PATH from the private MSYS2 root and selected shell profile. Managed build steps do not modify the system PATH and do not require administrator rights by default.

## Generated scripts

Generated shell scripts are written by the scripting layer, hashed, and checked again before execution.

The plan records the approved script path and SHA-256 hash. The execution layer verifies that the script on disk still matches the approved hash before running it.

## Workspace model

Workspace-sensitive filesystem behavior is centralized around workspace-aware helpers.

The workspace layer models selected workspace paths, private MSYS2 paths, source directories, artifact directories, run logs, and result folders. Path checks are based on containment and resolved filesystem paths where needed.

Build outputs are copied into:

```text
workspace/FFmpeg/
```

The app keeps generated tools and build outputs inside the selected workspace rather than installing FFmpeg globally.

## Audit logs

Approved actions create a run directory under:

```text
workspace/logs/<runId>/
```

The directory contains local evidence such as:

- `security-events.jsonl`
- `stdout.log`
- `stderr.log`

These logs are local files intended to make approved activity inspectable after the fact.

## Artifacts and result reporting

Successful FFmpeg builds copy output files into the workspace result folder.

The result folder may include:

- `ffmpeg.exe`
- `ffprobe.exe`
- required MSYS2 DLL dependencies discovered from PE imports
- `build-report-<runId>.json`

The build report records selected libraries, selected configure options, required MSYS2 package names, generated flags, final configure flags, license profile, artifact paths, sizes, and SHA-256 hashes.

If the final artifact copy fails, the backend leaves the built FFmpeg files in the source build directory and logs their location for inspection or recovery.

## Library and license visibility

Third-party FFmpeg libraries are security- and license-relevant, so the backend models them as first-class plan objects.

A reviewed FFmpeg build plan exposes selected libraries, generated MSYS2 packages, generated configure flags, final configure flags, selected configure options, warnings, and license effects before backend confirmation.

FFmpeg's own built-in components are shown as checked and locked rows. External libraries remain unchecked until selected because they change packages, configure flags, and license status.

Manual configure flags remain available as an escape hatch. They are included in the reviewed final configure flag list before backend confirmation.

## Static scanner

`scripts/security-scan.go` is a source-tree boundary scanner. It searches for sensitive API usage and compares each hit against the intended package boundary.

The scanner currently tracks patterns such as:

- command execution,
- HTTP download primitives,
- recursive deletion,
- direct file deletion,
- file renaming,
- direct file writes,
- shell-string execution.

The scanner describes the intended architecture of the current source tree. A scanner failure means the implementation and the documented package boundary no longer match.
