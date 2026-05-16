# Library and options model

This document describes how the app represents FFmpeg build choices in the Library and Options pages.

The app separates two kinds of choices:

1. included FFmpeg components that are already part of a normal FFmpeg source build;
2. external libraries and optional configure flags that change MSYS2 packages, generated flags, output type, or license profile.

## Included FFmpeg components

Rows marked `included` are built-in FFmpeg components. They are checked and locked because they are part of a normal FFmpeg source build.

Included rows currently cover:

- `ffmpeg.exe`
- `ffprobe.exe`
- `libavcodec`
- `libavformat`
- `libavfilter`
- `libavutil`
- `libswscale`
- `libswresample`
- native FFmpeg codecs
- native formats and muxers

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

The `xavs2` entry is special in the current code: it uses `mingw-w64-x86_64-xavs2` and produces a warning unless the selected shell profile is `mingw64`.

## Library presets

The Library page includes presets as starting points:

- `default` — included FFmpeg components only;
- `efficiency` — common efficient encoders, decoders, and audio helpers;
- `compatibility` — broader codec compatibility, including speech and legacy audio libraries;
- `editor` — subtitle, text, image, filtering, scaling, and quality-analysis helpers;
- `full` — broad catalog selection except mutually exclusive choices omitted by preset logic;
- `custom` — shown when the checked libraries no longer match a preset exactly.

Applying a preset still leaves individual external libraries editable. Locked included rows remain selected.

## Mutually exclusive choices

The UI prevents known mutually exclusive selections. The current mutually exclusive group is:

- `openssl`
- `gnutls`

Only one of these TLS library choices remains selected at a time.

## Advanced flags

The manual flags text area is an escape hatch for flags that are not yet represented by a named checkbox.

Manual flags are still part of the reviewed plan. They are merged into the final configure flag list, validated, and shown again before backend confirmation.

## Manual `--enable-lib...` recovery

If a manual advanced flag matches a cataloged external library, the backend resolves that flag back to the library catalog and adds the missing MSYS2 package names before configure runs.

This keeps a known `--enable-lib...` flag from enabling a library while accidentally skipping the package installation that the named checkbox would have provided.

## Configure option catalog

The Options page exposes named configure choices separately from external libraries.

Locked defaults currently include:

- `default-static`
- `default-programs`
- `default-ffmpeg`
- `default-ffprobe`

Optional choices currently include:

- `disable-doc`
- `disable-debug`
- `enable-shared`
- `disable-programs`
- `disable-ffplay`
- `disable-ffprobe`
- `disable-autodetect`
- `enable-version3`
- `disable-asm`
- `disable-x86asm`
- `enable-small`
- `disable-stripping`

## License profile derivation

The backend derives the license boundary from selected libraries and final configure flags.

Current profile names are:

- `lgpl-local`
- `gpl-local`
- `nonfree-local`

GPL libraries or `--enable-gpl` move the build to `gpl-local`. Nonfree libraries or `--enable-nonfree` move the build to `nonfree-local` and generate a redistribution warning.

AMR libraries that require FFmpeg's version-3 license switch cause `--enable-version3` to be added automatically and produce an informational warning.
