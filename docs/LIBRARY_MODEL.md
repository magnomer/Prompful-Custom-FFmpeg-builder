# Library and options model

This document describes how the program represents FFmpeg build choices in the Library and Options pages.

The program separates two kinds of choices:

1. FFmpeg components that are already part of a normal FFmpeg source build;
2. external libraries and optional configure flags that change MSYS2 packages, generated flags, output type, or license profile.

## Components included in the official FFmpeg release

Rows marked `included` are usually built-in FFmpeg components. They are checked and locked because they are part of a normal FFmpeg source build.

Included rows currently cover:

- `ffmpeg.exe`
- `ffprobe.exe`
- `libavcodec`
- `libavformat`
- `libavfilter`
- `libavutil`
- native FFmpeg codecs
- native formats and muxers
- `libswscale`
- `libswresample`

Included rows do not add MSYS2 packages and do not add `--enable-lib...` flags. They exist so the Library page shows what FFmpeg already provides before external libraries are selected.

## External libraries

Unchecked rows are external libraries. Selecting one can add:

- MSYS2 package names,
- FFmpeg `./configure` flags,
- license effects,
- review notes and warnings.

External libraries are represented as first-class plan objects instead of being hidden inside a generic configure flag field.

Each external library exposes:

- `libraryId`
- display name
- category
- generated configure flags
- required MSYS2 packages
- license effect
- plain-language explanation
- technical explanation
- default and locked state

## Shell-profile-specific packages

Most external library package names are generated from the selected Windows shell profile:

- `ucrt64` uses `mingw-w64-ucrt-x86_64-...`
- `mingw64` uses `mingw-w64-x86_64-...`
- `clang64` uses `mingw-w64-clang-x86_64-...`

A few entries such as `xavs2` and `vvenc` are real FFmpeg configure options that have no prebuilt MSYS2 package. They are modeled as Internal-track source-build libraries, so they do not contribute MSYS2 package names; the builder prepares them from a verified upstream source archive before FFmpeg configure runs. A non-native library with no implemented preparation recipe stays blocked through the non-native preparation gate.

## Hidden libraries

`lensfun`, `svtjpegxs`, and `vapoursynth` stay in the backend catalog but are hidden from the UI and from automatic presets. They are kept for future compatibility: an old saved or manual request that names one is still honored, but the planner skips the incompatible flag with a warning when the installed package does not match the FFmpeg source. `vapoursynth` is hidden because the MSYS2 package is older than the API the FFmpeg source requires, and even when it builds the result needs a Python + VapourSynth runtime, so it is not portable.

## Profile-specific availability

Some MSYS2 packages exist for only certain shell profiles. `onnxruntime` has no prebuilt `mingw64` package, so `LibraryCatalogForShellProfile` omits it for that profile and the UI hides it. The selection is re-normalized when the profile changes, so switching to `mingw64` drops a previously selected `onnxruntime`, and the planner ignores it as an unknown id if it ever reaches the backend. The frontend `libraryUnavailableProfiles` map mirrors the backend `libraryProfileUnavailability` map.

## Library presets

The Library page includes presets as starting points. Each higher tier adds to the one below it:

- `minimal` — included FFmpeg components only;
- `default` — common encoders, decoders, hardware acceleration, audio, and subtitle helpers (the selection checked on first load);
- `efficiency` — `default` plus best-quality encode and resample helpers;
- `compatibility` — `efficiency` plus broader codec, caption, and protocol I/O;
- `editor` — `compatibility` plus filtering, audio plugins, color, and subtitle/transcription tooling;
- `full` — `editor` plus packageable advanced features, except mutually exclusive choices omitted by preset logic;
- `custom` — shown when the checked libraries no longer match a preset exactly.

Applying a preset still leaves individual external libraries editable. Locked included rows remain selected.

## Mutually exclusive choices

The UI prevents known mutually exclusive selections. The current mutually exclusive groups are:

- `openssl` / `gnutls` / `mbedtls` — only one TLS backend stays selected (FFmpeg configure rejects more than one);
- `shaderc` / `glslang` — only one runtime shader compiler stays selected (FFmpeg configure rejects both);
- `xevd` / `xevdb` — only one EVC decoder binding stays selected (FFmpeg configure rejects both);
- `xeve` / `xeveb` — only one EVC encoder binding stays selected (FFmpeg configure rejects both).

Selecting one member of a group clears the other, and the planner also emits a blocking warning if both ever reach the final flag list.

## Advanced flags

The manual flags text area is an escape hatch for flags that are not yet represented by a named checkbox.

Manual flags are still part of the reviewed plan. They are merged into the final configure flag list, validated, and shown again before backend confirmation.

## Manual `--enable-lib...` recovery

If a manual advanced flag matches a cataloged external library, the backend resolves that flag back to the library catalog and adds the missing MSYS2 package names before configure runs.

This keeps a known `--enable-lib...` flag from enabling a library while accidentally skipping the package installation that the named checkbox would have provided.

## Configure option catalog

The Options page exposes named configure choices separately from external libraries. Each option carries a category, generated configure flags, a plain-language explanation, a technical note, and a risk level.

Locked defaults (always selected) are:

- `default-static`
- `default-programs`
- `default-ffmpeg`
- `default-ffprobe`

Optional toggles, grouped by category, are:

- Output type — `enable-shared`
- Programs — `disable-ffplay`
- Security and reproducibility — `disable-autodetect`, `disable-network`
- Compatibility — `disable-asm`, `disable-x86asm`, `pkg-config-static`, `enable-runtime-cpudetect`
- Size and speed — `disable-doc`, `enable-small`, `enable-lto`
- Debugging — `disable-debug`, `disable-stripping`

GPL/nonfree/version-3 behavior is derived automatically from the selected libraries, so there are no manual `enable-gpl`, `enable-nonfree`, or `enable-version3` toggles. There are also no `disable-programs` or `disable-ffprobe` toggles, because they would contradict the locked program defaults.

## Option risk levels

Each option shows a colored risk pill, mirroring the library license tags:

- `high` — `enable-shared` (switches to DLLs and disables static, which breaks linking against the static external libraries this builder installs) and `disable-asm` (removes nearly all SIMD optimization). Their descriptions carry an inline red risk warning.
- `medium` — `disable-network`, `disable-x86asm`, `enable-lto`.
- `low` — every other option, including the locked defaults.

## Option presets

The Options page includes flat intent profiles (not nested like the library presets):

- `none` — no extra options, plain FFmpeg defaults;
- `standard` — `pkg-config-static` and `disable-doc` (the selection applied on first load);
- `compact` — `standard` plus `disable-debug`;
- `portable` — `compact` plus `enable-runtime-cpudetect`;
- `performance` — `compact` plus `enable-lto`;
- `custom` — shown when the selected options no longer match a preset.

High-risk and troubleshooting toggles (`enable-shared`, `disable-asm`, `disable-x86asm`, `disable-network`) are intentionally in no preset, so every preset stays safe. The default selected options match `standard`, because static external-library linking and skipping documentation are the sensible defaults for this static build.

When `disable-network` is selected together with a network library (`SRT`, `libssh`, `librtmp`, `librist`, `ZeroMQ`, `RabbitMQ-C`), the planner emits a warning that those libraries will not work.

## License profile derivation

The backend derives the license boundary from selected libraries and final configure flags.

Current profile names are:

- `lgpl-local`
- `gpl-local`
- `nonfree-local`

GPL libraries or `--enable-gpl` move the build to `gpl-local`. Nonfree libraries or `--enable-nonfree` move the build to `nonfree-local` and generate a redistribution warning.

Libraries that require FFmpeg's version-3 license switch (`opencore-amr`, `vo-amrwbenc`, `libaribb24`, `lensfun`) cause `--enable-version3` to be added automatically and produce an informational warning.
