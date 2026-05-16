# Consent boundary

The frontend is untrusted.

A malicious frontend, injected JavaScript, WebView devtool access, replayed local state, or direct RPC caller may send an approval-shaped message. Backend code must never treat that incoming approval request as final proof of user intent.

## Current approval path

1. The backend creates an immutable executable plan.
2. The backend creates a review session for that plan.
3. The frontend displays the plan, warnings, final flags, packages, license effects, and the exact consent text.
4. The frontend may send an approval request containing:
   - review session id,
   - approved action name,
   - approved plan hash,
   - exact consent text.
5. The backend retrieves the stored plan by review session id.
6. The backend checks that the review session is still valid:
   - session exists,
   - action name matches,
   - plan hash matches,
   - consent text hash matches,
   - session has not expired,
   - session has not already been consumed.
7. The backend recomputes the stored plan hash from the stored plan content.
8. The backend opens a native OS confirmation dialog.
9. The backend creates action-specific consent values only after the native dialog returns `Yes`.
10. The backend starts the approved action with the stored plan and matching consent values.

## Required native confirmation rule

Approval RPC calls are only requests to approve. They are not final approval.

The backend must start a download, extraction, package installation, command execution, or cleanup only when all of these are true:

- the review session check passes;
- the stored plan is executable;
- the stored plan hash still matches the stored plan content;
- the backend-owned native dialog returns exactly `Yes`;
- the mutating function receives its matching action-specific consent type.

## Rejection rule

The backend must abort when the native dialog returns anything except `Yes`.

This includes:

- `No`
- `Cancel`
- Escape
- window close
- dialog error
- empty result
- any unexpected button string

The native dialog must default to `No`.

## One-time review sessions

A review session is consumed before the action starts. Reusing the same review session id must fail.

Current review sessions are created with a 30-minute lifetime. Expired sessions must be rejected even when the action name, plan hash, and consent text are otherwise correct.

## Action-specific consent values

The approval request is converted into narrower consent values only inside backend approval methods.

Current consent types are:

- `Msys2DownloadConsent`
- `FfmpegSourceDownloadConsent`
- `ArchiveExtractionConsent`
- `PacmanInstallConsent`
- `CommandExecutionConsent`
- `WorkspaceDeletionConsent`

Boolean consent values are forbidden.

## Why this exists

This blocks the attack where a forged frontend or direct RPC message says approval was granted even though the user did not approve the backend-owned confirmation dialog.

It also blocks approval replay: a caller cannot reuse an old review session because the backend stores sessions, checks expiry, and removes a session when it is accepted.
