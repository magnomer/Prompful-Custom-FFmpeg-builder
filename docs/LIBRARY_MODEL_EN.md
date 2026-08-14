# Library choices and build options

This document explains how Promptful Custom FFmpeg Builder treats library selections, presets, configure options, conflicts, and license information.

Library entries are classified by how they are prepared and whether they can be selected for the active FFmpeg release and MSYS2 shell profile. Some entries are part of the FFmpeg source tree, some use ordinary MSYS2 packages, some are prepared from source by the builder, and some are registered so their restrictions can be shown clearly.

## Library entry types

| Type | Meaning | Build effect |
|---|---|---|
| Included FFmpeg entries | FFmpeg programs and core libraries that come from the FFmpeg source tree itself | Locked on; no extra MSYS2 package and no `--enable-lib...` flag |
| Native MSYS2 package entries | Optional FFmpeg integrations satisfied by MSYS2 packages | Add package names and FFmpeg configure flags |
| Internal source-prepared entries | Optional integrations prepared inside the private MSYS2 environment | Run a preparation recipe, then add configure flags |
| External SDK/import entries | Integrations that need an outside SDK, import library, or special external setup | Block planning unless a supported import/preparation path exists |
| Disabled or gated entries | Registered entries that are not currently selectable | Shown as unavailable, blocked, or filtered by FFmpeg version/profile |

## Included FFmpeg entries

Included FFmpeg entries are always selected because they are part of the FFmpeg source build itself. Examples include:

- `ffmpeg`
- `ffprobe`
- `avcodec`
- `avformat`
- `avfilter`
- `avdevice`
- `swscale`
- `swresample`

These entries explain the baseline that every build starts from. They do not install external packages and do not add separate `--enable-lib...` flags.

## Native package entries

Most optional libraries use the native MSYS2 package path. For these entries, the program adds the required MSYS2 package names and the matching FFmpeg configure flags.

Examples include `x264`, `x265`, `libvpx`, `aom`, `dav1d`, `opus`, `mp3lame`, `ass`, `freetype`, `zimg`, `vmaf`, `srt`, and `openssl`.

A native package entry can still be unavailable for a particular FFmpeg release or MSYS2 shell profile. The program filters the library catalog before planning so an unavailable package/profile combination is not treated as selectable.

## Internal source-prepared entries

Some libraries need source preparation rather than an package-based selection. When the program supports one of these entries, it prepares the source or import files inside the private build environment before FFmpeg configure runs.

Current implemented internal source-prepared entries include:

- `vvenc`
- `lcevc-dec`
- `davs2`
- `uavs3d`
- `xavs2`
- `avisynthplus`
- `mpeghdec`
- `quirc`
- `klvanc`
- `libtls`
- `libmfx`
- `opencv`

`libmfx` is part of this implemented group. It is the legacy Intel Media SDK path used when the selected FFmpeg release/profile calls for that backend. It is mutually exclusive with `libvpl`, the newer oneVPL path. `opencv` is also internally prepared: the builder pins OpenCV 4.11.0 because current MSYS2 OpenCV 5 removed the legacy C API that FFmpeg `--enable-libopencv` still requires. Other internal-track entries, such as `dc1394` or `pocketsphinx`, can still be blocked when the program has no supported preparation path for the selected environment.

## External or blocked entries

Some FFmpeg integrations need an external SDK, special import library, or build path that the program does not currently prepare safely. These entries may be registered in the library catalog, but the plan blocks them instead of presenting them as reliably enableable.

Current blocked entries without a supported build/import path are:

- `smbclient`
- `openvino`
- `torch`
- `pocketsphinx`
- `dc1394`
- `decklink`
- `cuda-nvcc`

This does not indicate that FFmpeg can never be built with those technologies. It means this program does not currently provide a safe, local, reviewable preparation path for them.

## Disabled and version-gated entries

Two entries are globally disabled in the standard UI:

- `tensorflow`
- `vapoursynth`

Other entries are not globally disabled, but can be hidden, blocked, or removed depending on the selected FFmpeg release and MSYS2 shell profile.

Examples:

- `libvpl` is the preferred Intel oneVPL path on newer FFmpeg release lines.
- `libmfx` is the legacy Intel Media SDK path and must not be enabled together with `libvpl`.
- `lensfun` can be blocked when the available package does not satisfy the FFmpeg release requirement.
- `onnxruntime` is supported by FFmpeg 9.0.1 on `ucrt64` and `clang64`, where the builder emits `--enable-libonnxruntime`; it remains unavailable on `mingw64` and is unsupported on the older cataloged FFmpeg lines.
- `svtjpegxs` is supported on current newer FFmpeg lines, including 9.0.1, and remains version-dependent on older releases.
- On FFmpeg 9.0.1, `shaderc` and `glslang` remain selectable MSYS2 shader-tool packages, but FFmpeg 9 removed `--enable-libshaderc` and `--enable-libglslang`, so those flags are not emitted.

## Public library presets

Library presets are starting points for selection. They do not replace the library catalog, and they do not guarantee that every selected item will remain valid for every FFmpeg release/profile combination.

The public presets are:

| Preset | Purpose | Selection rule |
|---|---|---|
| `minimal` | The locked FFmpeg-source baseline | Included FFmpeg entries only |
| `default` | A practical first build with common codecs, audio helpers, subtitle/font support, hardware acceleration, and common network/filter helpers | `minimal` + default additions |
| `efficiency` | Compression and quality-per-bit helpers | `default` + efficiency additions only |
| `compatibility` | Broader codec, subtitle, caption, image, speech, and protocol coverage | `default` + compatibility additions only |
| `editor` | Editing, filtering, color, audio-analysis, subtitle, transcription, and image-workflow additions | `default` + editor additions only |
| `full` | The broadest public preset after mutually exclusive choices are normalized | `default` + efficiency + compatibility + editor + full-only additions |
| `custom` | Shown when the current selection no longer exactly matches a preset | Not an applied preset template |

`efficiency`, `compatibility`, and `editor` are not cumulative steps. Each one starts from `default` and adds its own purpose-specific entries. `full` is the broad public union.

## Public preset additions

| Preset | Additions beyond its base |
|---|---|
| `default` | `nvenc`, `amf`, `libvpl`, `libmfx`, `x264`, `x265`, `libvpx`, `aom`, `svt-av1`, `dav1d`, `theora`, `xvid`, `opus`, `vorbis`, `mp3lame`, `gsm`, `speex`, `opencore-amr`, `vo-amrwbenc`, `rubberband`, `openjpeg`, `webp`, `freetype`, `fontconfig`, `fribidi`, `harfbuzz`, `ass`, `cairo`, `zimg`, `vmaf`, `vidstab`, `srt`, `ssh`, `zmq`, `openal`, `sdl2`, `gme`, `openmpt` |
| `efficiency` | `fdk-aac`, `soxr`, `rav1e` |
| `compatibility` | `openh264`, `xeve`, `xevd`, `oapv`, `xavs`, `ilbc`, `twolame`, `shine`, `codec2`, `lc3`, `snappy`, `rsvg`, `zvbi`, `aribb24`, `aribcaption`, `rtmp` |
| `editor` | `png`, `libjxl`, `lcms2`, `libplacebo`, `shaderc`, `frei0r`, `opencv`, `opencolorio`, `xml2`, `mysofa`, `bs2b`, `ladspa`, `lv2`, `chromaprint`, `qrencode`, `whisper` |
| `full` | all efficiency additions, all compatibility additions, all editor additions, plus `kvazaar`, `bluray`, `dvdread`, `dvdnav`, `cdio`, `modplug`, `opengl`, `openssl`, `rist`, `rabbitmq`, `tesseract`, `jack`, `pulse`, `caca`, `opencl` |

`libvpl` and `libmfx` both appear in the default preset because they represent Intel hardware acceleration across different FFmpeg release lines. Normalization keeps only the backend that is valid for the selected release/profile and never passes both to FFmpeg configure.

## Extended library mode

The Extended toggle is not a separate preset. It adds selected source-prepared entries to the broader public presets. Minimal and Default are intentionally unaffected.

| Preset | Extended additions |
|---|---|
| `efficiency` | `vvenc`, `lcevc-dec` |
| `compatibility` | `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `klvanc` |
| `editor` | `avisynthplus`, `lcevc-dec`, `quirc` |
| `full` | `vvenc`, `lcevc-dec`, `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `mpeghdec`, `quirc`, `klvanc`, plus `svtjpegxs` on 8.1.2/9.0.1 and `onnxruntime` on 9.0.1 |

Extended selections can change the derived license profile. For example, `xavs2`, `davs2`, and `avisynthplus` have GPL effects, while `mpeghdec` has a nonfree effect.

## Mutually exclusive choices

Some choices cannot be enabled together because FFmpeg rejects the combination or because the entries represent alternate bindings for the same role.

| Group | Members | Rule |
|---|---|---|
| TLS backend | `openssl`, `gnutls`, `mbedtls`, `libtls` | Choose one |
| Runtime shader compiler | `shaderc`, `glslang` | Choose at most one |
| EVC decoder binding | `xevd`, `xevdb` | Choose one profile binding |
| EVC encoder binding | `xeve`, `xeveb` | Choose one profile binding |

The UI removes conflicting selections as the user toggles entries. The planner also validates the final configure flag list and blocks mutually exclusive flags if they still reach the plan.

## Manual configure flags

The advanced configure-flags box is intended for explicit user-supplied FFmpeg configure flags. It does not bypass library catalog validation.

If a manual flag exactly matches a flag from a known library entry, the planner treats that entry as effectively selected. This lets the planner add the matching packages, license information, and preparation gates.

This recovery works only for library catalog entries that have configure flags. Entries that only add packages cannot be recovered from a manual `--enable-lib...` flag unless such a flag exists in the entry.

## Configure options

The Options page covers FFmpeg build switches that are not library entries.

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

The Options page does not expose ordinary checkboxes for `--enable-gpl`, `--enable-nonfree`, or `--enable-version3`. Those are derived from selected libraries and final flags. If the user enters those flags manually, the backend still validates them and derives the matching license profile.

The program also does not expose `disable-programs` or `disable-ffprobe` as standard options because they would contradict the locked program defaults.

## Option risk levels

Configure options use risk labels to describe build/runtime surprise, not license information.

| Risk | Meaning |
|---|---|
| High | Can substantially change build shape, portability, or performance; examples include `enable-shared` and `disable-asm` |
| Medium | Can surprise users or reduce capability/performance in specific cases; examples include `disable-network`, `disable-x86asm`, and `enable-lto` |
| Low | Ordinary defaults or relatively safe toggles |

`enable-shared` is high risk because it switches the build toward DLL output and away from the static-linking expectation of this builder. `disable-asm` is high risk because it removes most SIMD optimization.

## Option presets

Option presets are flat intent profiles. They are not nested like library presets.

Current option presets are:

- `none`
- `standard`
- `compact`
- `portable`
- `performance`
- `custom`

`standard` is the initial practical default. It selects `pkg-config-static` and `disable-doc`.

High-risk or troubleshooting options such as `enable-shared`, `disable-asm`, `disable-x86asm`, and `disable-network` are intentionally absent from standard presets.

When `disable-network` is selected together with network libraries such as SRT, libssh, librtmp, librist, ZeroMQ, or RabbitMQ-C, the planner emits a warning because those integrations will not be useful in a network-disabled FFmpeg build.

## License profile

The program derives the local license profile from selected libraries and final configure flags.

Current local profile names are:

- `lgpl-local`
- `gpl-local`
- `nonfree-local`

The effective ordering is:

```text
nonfree-local > gpl-local > lgpl-local
```

GPL libraries or `--enable-gpl` move the build to `gpl-local`. Nonfree libraries or `--enable-nonfree` move it to `nonfree-local` and produce a redistribution warning.

Libraries that require FFmpeg's version-3 license switch cause `--enable-version3` to be added automatically and produce an informational warning. The current implemented version-3 list is:

- `opencore-amr`
- `vo-amrwbenc`
- `libaribb24`
- `lensfun`

`--enable-version3` is not a separate license-profile name. It is an additional final configure flag and warning layer inside the derived local profile.

## How to read support claims

A library should be described as fully supported only when the program can approve the required package, source-preparation, or import path for the selected FFmpeg release and shell profile.

These are different claims:

- listed in the library catalog;
- visible in the UI;
- selectable in standard use;
- selected by a public preset;
- prepared from source by an implemented recipe;
- imported from an external SDK/path;
- passed to FFmpeg configure;
- finally accepted by FFmpeg configure.

The program deliberately avoids claiming support merely because upstream FFmpeg has a configure flag.
