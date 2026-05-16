**THIS PROJECT WAS BUILT THROUGH VIBE CODING WITH HELP FROM CLAUDE AND CHATGPT**.

This is a personal project, so some bugs may still exist.




# Promptful Custom FFmpeg Builder

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

## Build

```bash
go install github.com/wailsapp/wails/v2/cmd/wails@latest
cd frontend
npm install
cd ..
wails dev
```
## Download

Download the latest release from the GitHub Releases page:

[Latest release](https://github.com/magnomer/Prompful-Custom-FFmpeg-builder/releases/latest)

## Screenshots

<img width="1186" height="1107" alt="Screenshot-01" src="https://github.com/user-attachments/assets/508ae8f6-4509-4962-894b-56a93a786029" />
<img width="1186" height="1107" alt="Screenshot-02" src="https://github.com/user-attachments/assets/e1cbcd80-4290-41b6-8976-9bf0f54085a9" />
<img width="1186" height="907" alt="Screenshot-03" src="https://github.com/user-attachments/assets/30b7a10b-9d53-4448-805a-4c9d8a020a34" />


## Security model

The central rule is:

> Every function that downloads, extracts, installs, deletes, updates, or executes external tooling must require a dedicated user consent type in its signature.

The app uses this flow:

1. Request plan
2. backend stores immutable plan in a review session
3. frontend displays dry-run operations, warnings, packages, flags, license effects, and consent text
4. user requests approval for exact action name and plan hash
4. backend validates review session, consent text hash, expiry, and plan hash
5. backend shows native OS confirmation dialog
6. backend creates action-specific consent values
7. mutating functions verify consent and plan hash
8. execution begins

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
12. copy `ffmpeg.exe`, `ffprobe.exe`, and required DLL dependencies into `workspace/FFmpeg`;
13. write `build-report-<runId>.json`.

## Library and options model

The Library page distinguishes FFmpeg components included in the official release from external libraries.

Included rows are checked and locked because they are built as part of a normal FFmpeg source build. External libraries are unchecked until selected and may add packages, configure flags, license effects, and warnings.

The Options page exposes common FFmpeg configure choices as named rows. Manual advanced flags remain an escape hatch and appear in Review before backend confirmation.

See `LIBRARY_MODEL.md` for the full model.
