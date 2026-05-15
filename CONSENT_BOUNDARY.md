# Consent boundary

The frontend is untrusted.

A malicious frontend, injected JavaScript, WebView devtool access, or direct RPC caller may send an approval-shaped message. Therefore, backend code must never treat the incoming approval request as final proof of user intent.

## Required approval path

1. Backend creates a plan and review session.
2. Frontend displays the plan.
3. Frontend may send an approval request.
4. Backend checks the review session and plan hash.
5. Backend opens a native OS confirmation dialog.
6. Backend starts the action only if the native dialog returns `Yes`.

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

## Why this exists

This blocks the attack where a forged message says approval was granted even though the user pressed No in the visible frontend.
