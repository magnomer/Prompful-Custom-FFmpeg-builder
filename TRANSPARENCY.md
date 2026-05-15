# Transparency rules for security-critical code

This project intentionally keeps security-critical names plain. A reviewer should be able to scan the code and tell what each operation can do without decoding abstract labels.

## Naming rule

Use direct names for dangerous actions:

| Area | Preferred words | Avoid hiding behind |
|---|---|---|
| Downloading | `DownloadPlan`, `DownloadMsys2WithConsent`, `AllowedHosts` | vague `Spec`, generic `process`, `transfer` |
| Extraction | `ExtractPlan`, `ExtractArchiveWithConsent`, `checkExtractTarget` | vague `operation`, `materialize` |
| Command execution | `CommandPlan`, `RunCommandWithConsent`, `ExecutablePath` | vague `external activity`, `task`, `handler` |
| Path safety | `CheckRealPathInsideWorkspace`, `RejectSymlinkComponents` | vague `validate`, `sanitize` without saying what is checked |
| Consent | `Consent`, `CheckConsent`, `ApprovalRequest` | legalistic names that obscure the approval boundary |

## Security review map

A reviewer should start with these files:

1. `app.go` — public backend methods called by the UI. Public methods must plan first and run only after approval.
2. `internal/consent/consent.go` — what an approval contains and how approval is checked.
3. `internal/planning/planner.go` — what actions the backend plans before the user approves.
4. `internal/download/download.go` — all network file downloads.
5. `internal/extraction/extraction.go` — all archive unpacking.
6. `internal/execution/execution.go` — the only place that may call `exec.Command`.
7. `internal/workspace/workspace.go` — path containment, symlink, and workspace boundary checks.
8. `internal/scripting/scripting.go` — generated shell scripts and their hashes.
9. `internal/audit/audit.go` — local evidence of what ran.

## Code rule

Security-sensitive functions should answer one visible question in their name:

- `CheckConsent`: did the user approve this exact action and plan hash?
- `CheckRealPathInsideWorkspace`: after filesystem resolution, is this path still inside the workspace?
- `DownloadMsys2WithConsent`: download the MSYS2 file only after matching consent.
- `ExtractArchiveWithConsent`: extract an archive only after matching consent.
- `RunCommandWithConsent`: run a command only after matching consent.

## Review checklist

Before release, search for these strings:

```text
exec.Command
http.Client
http.Get
os.RemoveAll
os.WriteFile
os.OpenFile
archive/tar
filepath.EvalSymlinks
CheckConsent
```

Every hit should be explainable from the function name and nearby comments. If not, rename the function or split it into smaller pieces.
