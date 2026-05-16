# Consent boundary

This app treats the frontend as a presentation layer, not as the final authority for approving work.

The backend creates executable plans, stores review sessions, validates approval requests, opens the native confirmation dialog, and only then creates the action-specific consent values used by download, extraction, installation, command execution, and cleanup code.

## Where consent fits in the app

The app's build flow is plan-first:

1. The backend creates an executable plan.
2. The backend stores that plan in a review session.
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
8. The backend opens a native OS confirmation dialog.
9. The backend creates action-specific consent values after the native dialog returns `Yes`.
10. The backend starts the approved action using the stored plan and matching consent values.

## Backend-owned native confirmation

Approval RPC calls are treated as requests to approve, not as final approval.

The backend starts mutating work only after the stored review session is valid, the stored plan is executable, the recomputed plan hash still matches, the native confirmation dialog returns `Yes`, and the target function receives its matching action-specific consent type.

This applies to downloads, archive extraction, package installation, command execution, and workspace cleanup.

## Rejected confirmation results

The native dialog is configured so the safe result is `No`.

The backend treats every result other than `Yes` as rejection, including `No`, `Cancel`, Escape, window close, dialog errors, empty results, and unexpected button strings.

## One-time review sessions

Review sessions are consumed before the approved action starts. Reusing the same review session id fails.

Current review sessions have a 30-minute lifetime. Expired sessions are rejected even when the action name, plan hash, and consent text still match.

## Action-specific consent values

The approval request is converted into narrow consent values inside backend approval methods.

Current consent types are:

- `Msys2DownloadConsent`
- `FfmpegSourceDownloadConsent`
- `ArchiveExtractionConsent`
- `PacmanInstallConsent`
- `CommandExecutionConsent`
- `WorkspaceDeletionConsent`

The code uses these typed values instead of plain booleans so each mutating operation can check the exact action kind, action name, and plan hash it was approved for.

## Reason for this design

This design keeps a forged frontend message, injected JavaScript call, replayed local state value, WebView devtool call, or direct RPC call from becoming final proof of user intent.

It also blocks approval replay: the backend stores review sessions, checks expiry, and removes a session when it is accepted.
