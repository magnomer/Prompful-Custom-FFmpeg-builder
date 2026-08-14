# Pre-execution approval and confirmation procedure

Promptful Custom FFmpeg Builder creates an execution plan before running a task, displays the information that requires user review, checks that the plan is still valid, and then asks for final approval through a Windows native confirmation dialog. Pressing a button in the UI alone does not start downloads, archive extraction, package installation, or build command execution.

This document explains that approval procedure.

## Summary

Before an **initial** toolchain preparation or FFmpeg build begins, the program checks the following four conditions:

1. the operation the user intends to approve matches the operation in the approval request;
2. the plan hash shown on the review screen matches the plan stored by the backend;
3. the approval wording reviewed by the user matches the approval wording in the approval request;
4. the user selected `Yes` in the Windows native confirmation dialog opened by the backend.

Only when all four conditions are satisfied does the program create the operation-specific approval token that permits the initially approved operation to proceed. A retry after a transient FFmpeg network stall is handled separately as continuation of that same approved operation, as described below.

## What the user reviews

The review screen may show the selected source archives, packages to be installed, configure flags, generated scripts, warnings, license effects, and expected file operations. The UI displays this information, but it does not independently execute the build.

The backend stores a separate review session in memory and compares the approval request received from the UI against that session. This prevents errors such as a stale UI state, a modified WebView call, a reused approval request, or a local RPC call being applied to a different build.

## Approval process

The approval process proceeds in the following order:

1. the backend creates an executable plan;
2. the backend stores the plan in an in-memory review session;
3. the UI displays the plan, warnings, final configure flags, package list, license effects, and approval wording;
4. the UI sends an approval request containing the review session ID, operation name, plan hash, and approval wording;
5. the backend checks that the session exists, has not expired, has not already been used, and matches the request;
6. the backend recalculates the plan hash from the stored plan;
7. the backend checks that the stored plan is executable;
8. the backend opens a Windows native confirmation dialog;
9. if the user selects `Yes`, the backend creates the operation-specific approval token and starts the approved operation.

## Rejection, cancellation, and approval expiry

For every plan, the default approval result is `No`. The backend treats every result other than `Yes` as rejection. This includes `No`, `Cancel`, Escape, closing the dialog, dialog errors, empty responses, and unexpected button strings.

If the user rejects or cancels the Windows native confirmation dialog, the review session is not removed immediately. While the plan session is still valid, the same plan may be submitted again for Windows native confirmation. A plan session is treated as used only after confirmation through the Windows native dialog succeeds. Review sessions remain valid for 30 minutes, and unapproved sessions are discarded when the program is closed or restarted.

## Operations covered by approval

The current approval procedure applies to operations that can change the local environment or build environment.

| Covered operation | Operation after `Yes` |
|---|---|
| Private MSYS2 tool preparation | MSYS2 download, archive extraction, package installation |
| FFmpeg build | FFmpeg/source download, archive extraction, package installation, command execution |

During private MSYS2 tool preparation, the approval procedure also covers downloads of the MSYS2 signature file and signing key. During an FFmpeg build, the same approval procedure applies to the FFmpeg source download, FFmpeg signature file, signing key, and required library source downloads.

## Execution-time safeguards

Approval is only one of several safeguards. Operations that may affect the build environment are checked again during execution.

### Downloads

Approved downloads are still checked for HTTPS use, valid destination paths inside the workspace, temporary-file placement, unexpected file size, reuse rules, and SHA-256 validation when a hash is required. The program may also warn when a download is attempted from a host that is not on the expected download-host list, but the download is not judged by the host name alone.

For MSYS2 package transfers, generated package-install scripts configure an ordered mirror set and a bounded transfer command so a stalled mirror can fail over instead of hanging indefinitely. Exhausting transient-network retries is classified separately from a build/configuration failure.

### Archive extraction

Archive extraction checks that the archive and extraction destination are inside the selected workspace. It also limits the number of extracted files and the total extracted size. Attempts to extract to invalid paths or to create unsafe symbolic links are blocked before completion. As a result, the actual locations and contents written during extraction are controlled by extraction-time validation.

### Package installation and build commands

Package installation and build execution are performed through separate plans. During execution, the program checks the executable path, working directory, MSYS2 root, log-file location, executable name, shell metacharacters in executable paths, and null bytes in command arguments. These checks prevent an approved build step from running a build command outside the selected workspace.

### Scripts

When FFmpeg builds require shell scripts for package installation, configure, make, or selected library preparation work, the program records the script path and SHA-256 hash when the script is created. Immediately before execution, the script is checked again. If its content no longer matches the approved hash, execution is stopped.

## Cleanup and auxiliary actions

## Retry after a transient FFmpeg network stall

A network stall during an already approved FFmpeg build does not create a new plan. After the initial review session has passed backend validation and the native Windows confirmation, the GUI retains the exact backend plan and approval request for that running build. If transient package/download failures exhaust the retry budget, the run ends in a `stalled` state instead of `failed`, and the UI can show the mirror addresses that were tried.

Choosing **Retry the build** launches the same stored FFmpeg plan again with the same previously approved request. It does **not** create a second review session or display a second native approval dialog. The backend recreates the operation-specific consent values from that same approval request and executes the same plan hash. Cached and already prepared files may therefore be reused by the normal build logic.

This retry path is intentionally narrower than ordinary approval: it exists only for the same in-memory FFmpeg plan that successfully passed the original approval flow. Changing libraries, configure flags, source information, workspace settings, or other plan inputs requires generating and approving a new plan. A program restart also discards the frontend's in-memory retry reference.

## Cleanup and auxiliary actions

Some auxiliary actions are not treated as build-plan operations and therefore do not use the same approval token. Examples include opening the result folder, opening a report, opening the log folder, reading package state, inspecting the generated FFmpeg binary, and cleaning up private MSYS2 tools.

Cleanup or removal operations are still limited to the selected workspace. Cleaning failed preparation work, cleaning failed builds, removing invalid artifacts, replacing a tool with a new one, and explicitly removing private tools must remain inside the selected workspace. These operations also use path validation, drive-root refusal, symbolic-link checks, and program-constructed target paths.

## Summary

The program separates “displaying a plan and requesting approval” from “authorizing execution.” The UI can display the plan and request approval, but the backend stores the plan, verifies the plan hash, calls the Windows native confirmation dialog, creates the approval token, and performs the final execution checks.
