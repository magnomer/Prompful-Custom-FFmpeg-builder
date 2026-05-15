# Security policy for CustomFFmpeg Builder


## Transparency as a security control

Security-sensitive code must use plain names that reveal the action being performed. This is not only style: unclear names make consent bypasses, unsafe path handling, hidden downloads, and command execution harder to notice during review. See `TRANSPARENCY.md` for the naming rules and review map.

## Mandatory backend rules

1. No function may download a file unless it requires a dedicated download consent type.
2. No function may extract an archive unless it requires `ArchiveExtractionConsent`.
3. No function may install packages unless it requires `PacmanInstallConsent` or another action-specific install consent type.
4. No function may execute an external command unless it requires a dedicated execution consent type or a narrower install/build consent type.
5. No function may delete workspace content unless it requires `WorkspaceDeletionConsent`.
6. No public Wails method may directly mutate the system. Public methods must be plan-first and approval-first.
7. No arbitrary shell command text may be accepted from the frontend.
8. No source download may execute unless SHA-256 verification is configured.
9. No download may use non-HTTPS URLs in normal mode.
10. No extraction may write outside the selected workspace.
11. No symlink or hardlink from third-party archives may be restored.
12. No code outside `internal/execution` may call `exec.Command`.
13. No code outside `internal/download` may call external HTTP download APIs.
14. No managed build step may modify the system PATH.
15. No managed build step may require administrator rights by default.

## Suggested CI static checks

Search failures should block pull requests:

```text
exec.Command outside internal/execution
http.Get or http.DefaultClient.Do outside internal/download
os.RemoveAll outside internal/workspace
Approve* methods that do not verify plan hash
Download* functions without an action-specific consent parameter
Extract* functions without ArchiveExtractionConsent
Install* functions without an action-specific install consent parameter
```

## Additional hardening in this revision

The backend now includes these non-consent hardening measures:

- download destinations must remain inside the selected workspace;
- verified downloads use a conflict policy and reuse existing files only when the SHA-256 hash matches;
- partial `.part` downloads are removed after failed downloads;
- archive extraction requires a destination policy and blocks overwrites by default;
- archive extraction rejects absolute paths, path traversal, hardlinks, and symlinks;
- command execution requires the executable path to be inside the workspace;
- command execution requires an executable basename allowlist;
- Bash commands are written as approved script files and verified by SHA-256 before execution;
- direct `bash -lc` command-string execution is forbidden by the security scan;
- MSYS2 package names, FFmpeg configure flags, shell profiles, and job counts are validated before a plan is executable;
- extracted FFmpeg source lookup now requires exactly one child directory;
- artifact reports include output executable paths and SHA-256 hashes when the files exist;
- `scripts/security-scan.go` and `scripts/security-scan.ps1` check for common backend bypasses.

## Backend-owned review sessions

The backend stores every executable plan it creates during review. Approval calls pass only the review session id plus the approval request; they do not pass the executable plan back from the frontend. The backend retrieves the stored immutable plan, verifies the approved action name, approved plan hash, expected consent text hash, expiry, and one-time use, and then executes the stored plan.

## Durable audit logs

Approved actions create a run directory under `workspace/logs/<runId>/` containing `security-events.jsonl`, `stdout.log`, and `stderr.log`. These files are local-only and are intended to make approved activity inspectable after the fact.

## Extraction limits

Archive extraction enforces workspace boundaries, blocks links, normalizes archive paths with POSIX archive semantics, and caps file count, total extracted bytes, and single-file size.

## Backend-owned final approval

The frontend is not a trusted security boundary. JavaScript state, button handlers, and RPC calls can be altered or invoked directly by a hostile local process or a compromised WebView.

For that reason, approval RPC calls are not enough to start a download, extraction, package install, or command execution. After the backend receives an approval-shaped request and validates the review session and plan hash, it opens a backend-owned native confirmation dialog showing the action name and plan hash. The backend starts the action only when the native dialog returns `Yes`.

Security rule:

- A frontend message may request approval.
- Only the backend-owned native confirmation may finalize approval.
- `No`, window close, escape, dialog error, or any non-`Yes` response must stop execution.
- The native dialog defaults to `No`.

This protects the case where a forged frontend/backend message claims consent while the user actually rejects the operation.

## Library transparency

Third-party FFmpeg libraries are security- and license-relevant. They must not be hidden inside a generic configure flag field.

The backend now models libraries as first-class plan objects. A reviewed FFmpeg build plan exposes selected libraries, generated MSYS2 packages, generated configure flags, final configure flags, and license effects before backend confirmation.

## User-facing transparency revision

Workflow buttons must not jump users unexpectedly. The Install tab now adds the install plan and proceeds to Library; the final Review tab is reserved for the full pending plan and backend confirmation request.

Common FFmpeg configure flags must be exposed as named options with plain-language explanations. Raw advanced flags remain only as an escape hatch.

## Library transparency

The UI must not make the Library page look empty by default. FFmpeg's own built-in components are shown as checked, locked rows. External libraries remain unchecked unless the user selects them, because external libraries change packages, configure flags, and license status. Manual configure flags are kept as an escape hatch only and must appear in Review before backend confirmation.


## Signature verification implementation

MSYS2 signatures are verified internally with the ProtonMail Go OpenPGP implementation because the older `golang.org/x/crypto/openpgp` package cannot read modern OpenPGP public keys such as algorithm 22. The user should not need to install system GPG for the normal `.sig` verification path.
