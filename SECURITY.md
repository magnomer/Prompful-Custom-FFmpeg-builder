# Security policy for CustomFFmpeg Builder

## Transparency as a security control

Security-sensitive code must use plain names that reveal the action being performed. This is not only style: unclear names make consent bypasses, unsafe path handling, hidden downloads, and command execution harder to notice during review. See `TRANSPARENCY.md` for the naming rules and review map.

## Mandatory backend rules

1. No function may download a file unless it requires a dedicated download consent type.
2. No function may extract an archive unless it requires `ArchiveExtractionConsent`.
3. No function may install packages unless it requires `PacmanInstallConsent` or another action-specific install consent type.
4. No managed build function may execute an external command unless it requires a dedicated execution consent type or a narrower install/build consent type.
5. No function may delete workspace content unless it requires `WorkspaceDeletionConsent` or is part of a reviewed failure-cleanup path that stays inside the workspace.
6. No public Wails method may directly start toolchain preparation or FFmpeg build work. Public mutation methods must be plan-first and approval-first.
7. No arbitrary shell command text may be accepted from the frontend.
8. FFmpeg source downloads must be verified with the matching `.asc` PGP signature before extraction.
9. MSYS2 archive downloads should be verified with the matching `.sig` detached signature when a signature URL is supplied.
10. No download may use non-HTTPS URLs in normal mode.
11. No extraction may write outside the selected workspace.
12. No symlink or hardlink from third-party archives may be restored.
13. Managed build command execution must go through `internal/execution`.
14. HTTP download code must go through `internal/download`.
15. No managed build step may modify the system PATH.
16. No managed build step may require administrator rights by default.
17. Generated shell scripts must be written, hashed, and rechecked before execution.
18. Build outputs copied to the result folder must stay inside the selected workspace.

## Current backend-owned review sessions

The backend stores every executable plan it creates during review. Approval calls pass only the review session id and approval request; they do not pass the executable plan back from the frontend.

The backend retrieves the stored immutable plan, verifies:

- approved action name,
- approved plan hash,
- expected consent text hash,
- expiry,
- one-time use,
- recomputed stored plan hash,
- executable status.

Only after those checks does the backend open the native confirmation dialog.

## Backend-owned final approval

The frontend is not a trusted security boundary. JavaScript state, button handlers, saved local storage, and RPC calls can be altered or invoked directly by a hostile local process or a compromised WebView.

For that reason, approval RPC calls are not enough to start a download, extraction, package install, or command execution. After the backend receives an approval-shaped request and validates the review session and plan hash, it opens a backend-owned native confirmation dialog showing the action name and plan hash.

Security rule:

- A frontend message may request approval.
- Only the backend-owned native confirmation may finalize approval.
- `No`, window close, escape, dialog error, or any non-`Yes` response must stop execution.
- The native dialog defaults to `No`.

This protects the case where a forged frontend/backend message claims consent while the user actually rejects the operation.

## Downloads and authenticity checks

All normal downloads must use HTTPS and an allowed host list.

Current download scopes:

- MSYS2 archive and `.sig`: `github.com`, `repo.msys2.org`, `mirror.msys2.org`
- MSYS2 installer signing key: `keyserver.ubuntu.com`
- FFmpeg source archive, `.asc`, and release signing key: `ffmpeg.org`

Current download hardening includes:

- destination path must remain inside the selected workspace;
- partial `.part` downloads are removed after failed downloads;
- existing files are reused only when the expected SHA-256 hash matches;
- file-size minimum and maximum bounds are checked;
- SHA-256 is calculated and logged.

MSYS2 signatures are verified internally with the ProtonMail Go OpenPGP implementation because the older `golang.org/x/crypto/openpgp` package cannot read modern OpenPGP public keys such as algorithm 22. The user should not need to install system GPG for the normal `.sig` verification path.

FFmpeg release archives are verified with their matching `.asc` PGP signature and the downloaded FFmpeg release signing key.

## Extraction limits

Archive extraction enforces workspace boundaries, blocks links, normalizes archive paths with POSIX archive semantics, and caps:

- file count,
- total extracted bytes,
- single-file size.

Current format support includes:

- `.tar.zst`
- `.tar.xz`
- `.tar.gz`
- `.tgz`
- `.tar.bz2`
- `.tar`

MSYS2 `.tar.zst` is the recommended default. `.tar.xz` is accepted as a fallback. MSYS2 installer formats such as `.exe` and `.sfx.exe` are rejected by planning because this app does not run installers.

## Command execution

Managed build commands run through `internal/execution`.

The command plan must include:

- action name,
- plan hash,
- executable path,
- argument values,
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

The runtime environment builds an explicit MSYS2 PATH from the private MSYS2 root and selected shell profile instead of modifying the system PATH.

## Durable audit logs

Approved actions create a run directory under:

```text
workspace/logs/<runId>/
```

The directory contains local evidence such as:

- `security-events.jsonl`
- `stdout.log`
- `stderr.log`

These files are local-only and are intended to make approved activity inspectable after the fact.

## Artifacts and result reporting

Successful FFmpeg builds copy output files into:

```text
workspace/FFmpeg/
```

The result folder may include:

- `ffmpeg.exe`
- `ffprobe.exe`
- required MSYS2 DLL dependencies discovered from PE imports
- `build-report-<runId>.json`

The build report records selected libraries, selected configure options, required MSYS2 package names, generated flags, final configure flags, license profile, artifact paths, sizes, and SHA-256 hashes.

If the final artifact copy fails, the backend does not remove the built FFmpeg files from the source build directory; it logs their location so the user can inspect or recover them.

## Library transparency

Third-party FFmpeg libraries are security- and license-relevant. They must not be hidden inside a generic configure flag field.

The backend models libraries as first-class plan objects. A reviewed FFmpeg build plan exposes selected libraries, generated MSYS2 packages, generated configure flags, final configure flags, selected configure options, and license effects before backend confirmation.

FFmpeg's own built-in components are shown as checked, locked rows. External libraries remain unchecked unless the user selects them, because external libraries change packages, configure flags, and license status. Manual configure flags are kept as an escape hatch only and must appear in Review before backend confirmation.

## Suggested CI static checks

Search failures should block pull requests unless a reviewed exception is explicitly documented:

```text
exec.Command outside internal/execution
http.NewRequest, http.Client, or request.Do outside internal/download
os.RemoveAll outside controlled workspace cleanup code
Approve* methods that do not verify review session and plan hash
Download* functions without an action-specific consent parameter
Extract* functions without ArchiveExtractionConsent
Install* functions without an action-specific install consent parameter
Generated scripts that are executed without SHA-256 verification
bash -lc shell-string execution
```

## Static scanner alignment note

`scripts/security-scan.go` is intentionally strict. Keep it aligned with the source tree before release.

The current scanner declares `exec.Command` outside `internal/execution` and `bash -lc` shell-string execution as violations. If app-level helpers such as result-folder opening or cleanup agent shutdown remain in `app.go`, they must either be moved behind reviewed internal boundaries or represented as explicit scanner exceptions with narrow justification.
