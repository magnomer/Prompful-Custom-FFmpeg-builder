# Library and options model

This document describes how the program models FFmpeg library choices, source-prepared integrations, configure options, presets, conflicts, and license effects.

It is an implementation model document, not a user-facing feature list. A row can be present in the catalog without being normally selectable, and a row can be selectable without being available through an ordinary MSYS2 package.

## 1. Core principle

The Library page does not treat every checkbox as the same kind of thing. Each catalog row has a role in the build model:

| Layer | Meaning | Normal build effect |
|---|---|---|
| Included FFmpeg rows | FFmpeg programs and core libraries that come from the FFmpeg source tree itself | Locked on; no extra MSYS2 package and no `--enable-lib...` flag |
| Native package rows | Optional FFmpeg integrations that are satisfied by MSYS2 packages | Add MSYS2 package names and FFmpeg configure flags |
| Internal-track rows | Optional integrations that the builder prepares inside the private MSYS2 environment | Run an internal preparation recipe before FFmpeg configure, then add configure flags |
| External-track rows | Optional integrations that need an outside SDK/import path or an external preparation model | Block normal planning unless an implemented import/preparation recipe exists |
| UI-unavailable rows | Catalog rows kept visible for transparency but locked in normal UI use | Not selectable by normal users; excluded from presets and maximum-test selection |
| Profile-unavailable rows | Rows not available for the active MSYS2 shell profile | Dropped/disabled for that profile |

The word `external` therefore has a precise meaning in the code: it is a `LibraryTrackExternal` row. It should not be used casually to describe every optional non-included library.

## 2. Included FFmpeg rows

Rows marked as included are shown so the Library page can make clear what a normal FFmpeg source build already provides. They are checked and locked.

The included rows are:

- `ffmpeg-program`
- `ffprobe-program`
- `libavcodec`
- `libavformat`
- `libavfilter`
- `libavutil`
- `libswscale`
- `libswresample`
- `native-codecs`
- `native-formats`

Included rows do not add MSYS2 package names and do not add FFmpeg `./configure` flags. They are still part of the selected-library list so presets, summaries, and reports can show a complete picture of the build.

## 3. Library tracks

Every catalog row has a `trackName`.

### Native track

`native` is the default track. Native rows use packages from the selected MSYS2 shell profile.

For example, under `ucrt64`, package names generally use this prefix:

```text
mingw-w64-ucrt-x86_64-
```

The corresponding prefixes for other supported profiles are:

| Shell profile | Package prefix |
|---|---|
| `ucrt64` | `mingw-w64-ucrt-x86_64-` |
| `mingw64` | `mingw-w64-x86_64-` |
| `clang64` | `mingw-w64-clang-x86_64-` |

Native rows may add one or more package names and one or more configure flags. Some native rows add a package only and no configure flag; `png` is one example.

### Internal track

`internal` rows are integrations that are not satisfied by simply adding a normal target-library package. The builder has to prepare them before FFmpeg configure runs.

Currently modeled internal-track rows include:

- `vvenc`
- `xavs2`
- `davs2`
- `uavs3d`
- `lcevc-dec`
- `avisynthplus`
- `mpeghdec`
- `pocketsphinx`
- `dc1394`
- `libtls`
- `quirc`
- `klvanc`

A selected internal-track row is allowed only when the backend has a preparation recipe for it. If no recipe exists, the planner emits a blocking warning instead of approving a configure flag that cannot be backed by a prepared library.

Implemented preparation recipes currently exist for:

- `avisynthplus`
- `davs2`
- `xavs2`
- `uavs3d`
- `lcevc-dec`
- `vvenc`
- `mpeghdec`
- `quirc`
- `klvanc`
- `libtls`

The frontend also keeps some internal rows locked in normal use if the backend or user experience is not ready for ordinary selection.

### External track

`external` rows represent libraries that need an import path, vendor SDK, outside build product, or other non-native preparation model.

Currently modeled external-track rows include:

- `libmfx`
- `cuda-nvcc`
- `decklink`
- `smbclient`
- `openvino`
- `torch`
- `tensorflow`

These rows are not the same thing as ordinary optional FFmpeg libraries. A normal user should not be allowed to select an external-track row unless the program has a concrete, verified preparation/import path for it.

`tensorflow` has a preparation definition in the backend, but it is still UI-disabled. A preparation definition by itself does not mean a row is ready for normal user selection.

## 4. What a library row contains

A library row can contain:

- `libraryId`
- track name
- display name
- category name
- FFmpeg configure flags
- MSYS2 package names
- license effect
- official webpage URL
- plain explanation
- technical explanation
- default/locked state

The build planner uses this data to derive package installation, source-preparation requirements, configure flags, warnings, and license profile. The frontend uses the same catalog to render the Library page, preset summaries, disabled reasons, and technical details.

## 5. UI availability and catalog availability are different

A row can be present in the catalog while still not being normally selectable.

### UI-disabled rows

The following rows are deliberately visible but locked in the normal UI:

- `lensfun`
- `svtjpegxs`
- `vapoursynth`
- `tensorflow`
- `onnxruntime`

They are not hidden from the library list. They remain visible so users can see that the program knows about them and why they are not ordinary selections.

These rows are excluded from automatic presets and from the developer maximum-test preset candidate set.

### Unimplemented-build rows

The following rows are present in the catalog but do not yet have a normal implemented build/preparation path:

- `smbclient`
- `openvino`
- `torch`
- `libmfx`
- `pocketsphinx`
- `dc1394`
- `decklink`
- `cuda-nvcc`

Normal users cannot check these rows. Developer unlock can make them selectable for testing, but the backend still blocks planning if a selected non-native row has no implemented preparation recipe.

### Profile-unavailable rows

Some libraries are unavailable only for particular shell profiles. The frontend and backend both model this with a profile-unavailability map.

Currently:

| Library | Unavailable profile |
|---|---|
| `onnxruntime` | `mingw64` |

In the backend catalog, profile-unavailable rows are omitted for the active shell profile. In the frontend, they are shown as unavailable according to the matching UI map.

`onnxruntime` also has a broader UI-disabled reason: the relevant FFmpeg ONNX Runtime DNN backend support is not a normal official-release-source path for this program yet. It should not be described as only a missing MSYS2 package issue.

## 6. Special configure-script compatibility fallbacks

Some fragile rows are locked in the UI but still have backend catalog entries for future compatibility and for old/manual selections.

For these rows, fallback handling is not performed by the planner:

- `svtjpegxs`
- `lensfun`
- `vapoursynth`

If one of their configure flags reaches the generated FFmpeg configure script, the script runs a targeted compatibility check before the final `./configure` call. When the installed package does not expose the API required by the selected FFmpeg source, the script removes that specific configure flag, prints a warning, skips the related pkg-config diagnostic, and continues.

This is a narrow generated-script fallback. It is not a general rule for all disabled, unavailable, or incompatible libraries. For example, `tensorflow` and `onnxruntime` should not be described as having the same fallback unless a matching probe-and-remove path exists.

## 7. Library presets

Library presets are frontend selection templates. They are not the authoritative backend catalog.

### Public broadening presets

The public presets are:

- `minimal`
- `default`
- `efficiency`
- `compatibility`
- `editor`
- `full`
- `custom`

The public broadening presets are nested. Higher public tiers build on the lower tiers:

| Preset | Meaning |
|---|---|
| `minimal` | Included FFmpeg rows only |
| `default` | Common codecs, audio helpers, subtitle/font helpers, hardware acceleration, and common protocol/filter support |
| `efficiency` | `default` plus quality/efficiency helpers such as high-quality AAC/resampling choices |
| `compatibility` | broader codec, subtitle, caption, and protocol coverage |
| `editor` | editing/filtering/color/audio-analysis/subtitle/transcription-oriented additions |
| `full` | broadest normal packageable selection after mutually exclusive choices are normalized |
| `custom` | synthetic state shown when the current selection no longer exactly matches a preset |

Applying a preset still leaves ordinary selectable libraries editable. Locked included rows remain selected.

### Extended preset mode

The Extended toggle is not a separate preset ID. It modifies selected broadening presets by adding internal/source-prepared rows.

Current extended additions are:

| Preset | Extended additions |
|---|---|
| `efficiency` | `vvenc`, `uavs3d`, `lcevc-dec` |
| `compatibility` | `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `xavs2` |
| `editor` | `avisynthplus`, `lcevc-dec` |
| `full` | `vvenc`, `xavs2`, `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `mpeghdec` |

Extended presets can change the derived license boundary. For example, `xavs2`, `davs2`, and `avisynthplus` have GPL effects, while `mpeghdec` has a nonfree effect.

### Hidden focused presets

The frontend also defines focused presets that are hidden from the normal preset UI:

- `ai`
- `streaming`

These are intentionally not nested broadening presets. They are `default` plus their own focused additions. They do not inherit `efficiency`, `compatibility`, `editor`, or `full`.

### Developer maximum-test preset

The developer preset is:

- `maxtest`

It is shown only when the developer unlock is enabled. Its candidate set is derived from the live catalog and excludes:

- UI-disabled rows;
- rows with no implemented build/preparation path;
- rows unavailable for the active shell profile.

The selection is then normalized so mutually exclusive groups are reduced to a buildable set. Under the stronger sudo developer tier, the frontend can relax the TLS UI pruning for testing, but backend configure-conflict validation still matters if conflicting flags reach the final plan.

## 8. Mutually exclusive choices

Some libraries cannot be enabled together because FFmpeg configure rejects the combination or because they represent alternate bindings for the same role.

Current groups are:

| Group | Members | Rule |
|---|---|---|
| TLS backend | `openssl`, `gnutls`, `mbedtls`, `libtls` | normally choose one |
| Runtime shader compiler | `shaderc`, `glslang` | choose at most one |
| EVC decoder binding | `xevd`, `xevdb` | choose one profile binding |
| EVC encoder binding | `xeve`, `xeveb` | choose one profile binding |

The UI removes conflicting selections as the user toggles rows. The planner also validates the final configure flag list and emits a blocking warning if mutually exclusive flags reach the backend.

## 9. Planning behavior

The FFmpeg build planner receives selected library IDs, configure option IDs, extra configure flags, source settings, workspace settings, shell profile, and job count.

The planner then:

1. cleans the settings;
2. resolves selected library IDs against the active catalog;
3. derives native package installation from selected native rows;
4. resolves source-preparation/import requirements for selected internal/external rows;
5. recovers known catalog rows from matching manual configure flags;
6. merges generated library flags, option flags, manual flags, and derived license flags;
7. validates conflicts and risky combinations;
8. derives the local license profile;
9. produces the reviewable plan.

Native packages are installed before FFmpeg configure. Internal/external preparations run before their configure flags are treated as approved for the final configure script.

## 10. Manual configure-flag recovery

The advanced flags box is an escape hatch. It is not outside the model.

If an extra configure flag exactly matches a configure flag from a catalog row that is not already selected, the planner treats that row as an effective library. This lets the planner add native packages, license effects, and preparation gates that the matching checkbox would have provided.

This recovery is catalog-wide. It is not limited to external-track rows.

The recovery has an important limit: it only works for catalog rows with configure flags. Rows that add packages without configure flags cannot be recovered from a manual `--enable-lib...` flag unless such a flag exists in the catalog row.

## 11. Configure option catalog

The Options page is separate from the Library page. It models FFmpeg build switches that are not library rows.

Each configure option has:

- option ID;
- category;
- configure flags;
- plain explanation;
- technical explanation;
- risk level;
- default/locked state.

Locked defaults are:

- `default-static`
- `default-programs`
- `default-ffmpeg`
- `default-ffprobe`

Optional options include:

- Output type: `enable-shared`
- Programs: `disable-ffplay`
- Security/reproducibility: `disable-autodetect`, `disable-network`
- Compatibility: `disable-asm`, `disable-x86asm`, `pkg-config-static`, `enable-runtime-cpudetect`
- Size/speed: `disable-doc`, `enable-small`, `enable-lto`
- Debugging: `disable-debug`, `disable-stripping`

The Options page does not expose ordinary checkbox toggles for `--enable-gpl`, `--enable-nonfree`, or `--enable-version3`. Those are derived from selected libraries and final flags. If the user enters these flags manually through the advanced field, the backend still validates them and derives the corresponding license boundary.

The program also does not expose `disable-programs` or `disable-ffprobe` as normal options because those would contradict the locked program defaults.

## 12. Option risk levels

Configure options use a separate risk-level model. The visual treatment resembles library labels, but the meaning is different: option risk describes build/runtime surprise, not license effect.

Current risk interpretation:

| Risk | Meaning |
|---|---|
| High | Can substantially change build shape, portability, or performance; examples include `enable-shared` and `disable-asm` |
| Medium | Can reasonably surprise users or reduce capability/performance in specific cases; examples include `disable-network`, `disable-x86asm`, and `enable-lto` |
| Low | Ordinary defaults or relatively safe toggles |

`enable-shared` is treated as high risk because it switches the build toward DLL output and away from the static-linking expectation of this builder. `disable-asm` is high risk because it removes most SIMD optimization.

## 13. Option presets

Option presets are flat intent profiles. They are not nested in the same way as library presets.

Current option presets are:

- `none`
- `standard`
- `compact`
- `portable`
- `performance`
- `custom`

`standard` is the initial practical default. It selects `pkg-config-static` and `disable-doc`.

High-risk or troubleshooting options such as `enable-shared`, `disable-asm`, `disable-x86asm`, and `disable-network` are intentionally absent from ordinary presets.

When `disable-network` is selected together with network libraries such as SRT, libssh, librtmp, librist, ZeroMQ, or RabbitMQ-C, the planner emits a warning because those integrations will not be useful in a network-disabled FFmpeg build.

## 14. License profile derivation

The backend derives the license boundary from selected libraries and the final configure flags.

Current local profile names are:

- `lgpl-local`
- `gpl-local`
- `nonfree-local`

The effective ordering is:

```text
nonfree-local > gpl-local > lgpl-local
```

GPL libraries or `--enable-gpl` move the build to `gpl-local`. Nonfree libraries or `--enable-nonfree` move the build to `nonfree-local` and produce a redistribution warning.

Libraries that require FFmpeg's version-3 license switch cause `--enable-version3` to be added automatically and produce an informational warning. The current implemented version-3 list is:

- `opencore-amr`
- `vo-amrwbenc`
- `libaribb24`
- `lensfun`

`--enable-version3` is not a separate license-profile name. It is an additional final configure flag and warning layer inside the derived local profile.

## 15. Known boundary notes

The catalog is broader than the normal user-selectable set. This is intentional. It lets the UI show known-but-not-ready libraries, keeps old/manual paths recognizable, and gives developer mode a way to test future integrations.

However, support claims must distinguish clearly between:

- listed in the catalog;
- visible in the UI;
- selectable by normal users;
- selected by a public preset;
- prepared from source by an implemented recipe;
- imported from an external SDK/path;
- passed to FFmpeg configure;
- finally accepted by FFmpeg configure.

A library should be described as fully covered only when the appropriate package/preparation/import path exists and the normal plan can approve it without blocked warnings.
