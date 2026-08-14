# Library coverage report

This report describes which FFmpeg library entries are registered in this version of Promptful Custom FFmpeg Builder, which entries can be selected, and which entries are restricted by FFmpeg version, MSYS2 shell profile, package availability, or missing preparation support.

A library catalog entry is not necessarily a selectable option. Some entries are registered for tracking, comparison, warning, or future preparation support, and some selectable entries may still be excluded from the final build plan for a particular FFmpeg release or MSYS2 profile.

## Summary

- Total library catalog entries: **125**.
- Official FFmpeg-source entries: **10**.
- Native MSYS2 package entries: **95** (excluding the 10 entries built directly from FFmpeg source).
- Internal-track entries: **14**; **12** currently have implemented source-preparation recipes.
- External SDK/import entries: **6**.
- Globally UI-disabled entries: **2** — `tensorflow`, `vapoursynth`.
- Blocked entries without a supported build/import path: **7** — `smbclient`, `openvino`, `torch`, `pocketsphinx`, `dc1394`, `decklink`, `cuda-nvcc`.

Selection availability depends on the selected FFmpeg release and MSYS2 shell profile. The program checks the selected entries before configure so unsupported release/profile combinations are not included in the final build plan. The counts above use the current **FFmpeg 9.0.1** catalog classification; older supported releases may classify individual entries differently.

Version/profile-restricted entries are not counted as globally disabled. In the FFmpeg 9.0.1 catalog, `lensfun` is unavailable because the current package/API combination does not satisfy this release, while `onnxruntime` is selectable on `ucrt64` and `clang64` but unavailable on `mingw64`. `svtjpegxs` is selectable on FFmpeg 9.0.1 and remains release-dependent on older lines.

## How to read this report

| Term | Meaning |
|---|---|
| Included in FFmpeg source | Built from the official FFmpeg source tree and always included |
| Native MSYS2 package | Uses an MSYS2 package; it adds an FFmpeg configure flag when that FFmpeg release exposes one |
| Internal source-prepared | The program prepares source or import files inside the private build environment |
| External SDK/import | Needs an outside SDK, import library, or preparation path that is not package-based selection |
| UI-disabled | Registered in the library catalog, but not selectable in the standard UI |
| Blocked | Registered in the library catalog, but no supported build/import path exists yet |
| Version/profile-restricted | Available only for certain FFmpeg releases or MSYS2 shell profiles |

This distinction prevents the report from overstating support. A configure flag alone is not enough; the program must also define how the library is obtained, prepared, reviewed, and passed to FFmpeg.

## Coverage by category

| Category | Entries | Notes |
|---|---:|---|
| Included in FFmpeg source | 10 | Baseline programs and core FFmpeg libraries that are always included. |
| Video encoders | 19 | Software encoders, codec libraries, and special encoder integrations. |
| Hardware encoders | 5 | Header/API integrations for common Windows hardware encoder paths. |
| Video decoders | 8 | External decoder libraries and wrapper decoders. |
| Image codecs | 11 | Still-image and image-sequence helpers. |
| Filters and processing | 17 | Scaling, filtering, color, analysis, GPU/filter helpers, and plugin paths. |
| Audio | 25 | Audio codecs, analysis, effects, and device/server integrations. |
| Subtitles and text | 11 | Font shaping, subtitle rendering, captions, teletext, OCR-related text support. |
| Disc and device input | 8 | Optical media, capture, and platform input integrations. |
| Network | 7 | Transport/protocol libraries excluding TLS backend choice. |
| Secure network (TLS) | 4 | Mutually exclusive TLS backend options. |

## Public preset coverage

Public library presets are selection templates. They are not all cumulative tiers.

| Preset | Purpose | Main additions |
|---|---|---|
| Minimal | Official FFmpeg-source baseline only | FFmpeg programs and core libraries included by default |
| Default | Practical first build for common encoding, decoding, subtitles, audio, processing, and network helpers | `nvenc`, `amf`, `libvpl`, `libmfx`, `x264`, `x265`, `libvpx`, `aom`, `svt-av1`, `dav1d`, `theora`, `xvid`, `opus`, `vorbis`, `mp3lame`, `gsm`, `speex`, `opencore-amr`, `vo-amrwbenc`, `rubberband`, `openjpeg`, `webp`, `freetype`, `fontconfig`, `fribidi`, `harfbuzz`, `ass`, `cairo`, `zimg`, `vmaf`, `vidstab`, `srt`, `ssh`, `zmq`, `openal`, `sdl2`, `gme`, `openmpt` |
| Maximum Efficiency | `default` plus compression/quality-per-bit helpers | `fdk-aac`, `soxr`, `rav1e` |
| Maximum Compatibility | `default` plus broader codec, subtitle, caption, image, speech, and protocol coverage | `openh264`, `xeve`, `xevd`, `oapv`, `xavs`, `ilbc`, `twolame`, `shine`, `codec2`, `lc3`, `snappy`, `rsvg`, `zvbi`, `aribb24`, `aribcaption`, `rtmp` |
| Audio/Video Editor | `default` plus editing, filtering, color, analysis, subtitle, transcription, and image-workflow helpers | `png`, `libjxl`, `lcms2`, `libplacebo`, `shaderc`, `frei0r`, `opencv`, `opencolorio`, `xml2`, `mysofa`, `bs2b`, `ladspa`, `lv2`, `chromaprint`, `qrencode`, `whisper` |
| Full | Broadest public selection after mutually exclusive choices are normalized | all efficiency, compatibility, and editor additions, plus `kvazaar`, `bluray`, `dvdread`, `dvdnav`, `cdio`, `modplug`, `opengl`, `openssl`, `rist`, `rabbitmq`, `tesseract`, `jack`, `pulse`, `caca`, `opencl` |

`Maximum Efficiency`, `Maximum Compatibility`, and `Audio/Video Editor` each start from `Default` and add their own purpose-specific entries. They do not inherit from one another. `Full` is the broad public union.

After a preset is applied, the program normalizes the selection for the active FFmpeg release and MSYS2 shell profile. For example, the Intel hardware-acceleration capability may list both `libvpl` and `libmfx` as possible preset entries, but the final plan keeps the valid backend for the selected release/profile and never enables both together.

## Extended library mode

The Extended toggle adds selected source-prepared entries to the broader public presets. Minimal and Default are intentionally unaffected.

| Preset | Extended additions |
|---|---|
| Maximum Efficiency | `vvenc`, `lcevc-dec` |
| Maximum Compatibility | `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `klvanc` |
| Audio/Video Editor | `avisynthplus`, `lcevc-dec`, `quirc` |
| Full | `vvenc`, `lcevc-dec`, `davs2`, `uavs3d`, `xavs2`, `avisynthplus`, `mpeghdec`, `quirc`, `klvanc`, plus `svtjpegxs` on FFmpeg 8.1.2/9.0.1 and `onnxruntime` on FFmpeg 9.0.1 |

Extended entries may change the derived license profile. For example, `xavs2`, `davs2`, and `avisynthplus` add GPL license effects, while `mpeghdec` adds a nonfree license effect.

## Mutual exclusion and normalization

Some entries are alternatives and must not be enabled together.

| Group | Members | Selection rule |
|---|---|---|
| Intel hardware acceleration | `libvpl`, `libmfx` | choose the backend valid for the selected FFmpeg release/profile |
| TLS backend | `openssl`, `gnutls`, `mbedtls`, `libtls` | choose one TLS backend |
| Runtime shader compiler | `shaderc`, `glslang` | choose at most one |
| EVC decoder binding | `xevd`, `xevdb` | choose one profile binding |
| EVC encoder binding | `xeve`, `xeveb` | choose one profile binding |

## Source-prepared and imported libraries

These entries need more than MSYS2 package selection. For supported source-prepared entries, the program downloads or imports the required files, verifies the expected SHA-256 hash where available, builds or prepares the dependency, checks the expected headers and libraries, and only then passes the matching configure flag to FFmpeg.

| ID | Display name | Track | Version pin | Method | Build system / import style | Status |
|---|---|---|---|---|---|---|
| `avisynthplus` | AviSynth+ / Scripted video processing | internal source-prepared | 3.7.5 | internal source build | CMake | Selectable when supported by the selected release/profile |
| `davs2` | libdavs2 / AVS2 decoding | internal source-prepared | 1.7 | internal source build | configure + make | Selectable when supported by the selected release/profile |
| `xavs2` | xavs2 | internal source-prepared | 1.4 | internal source build | configure + make | Selectable when supported by the selected release/profile |
| `uavs3d` | libuavs3d / AVS3 decoding | internal source-prepared | master | internal source build | CMake | Selectable when supported by the selected release/profile |
| `lcevc-dec` | liblcevc-dec / LCEVC decoding | internal source-prepared | 4.2.0 | internal source build | CMake | Selectable when supported by the selected release/profile |
| `vvenc` | vvenc (VVC/H.266) | internal source-prepared | 1.14.0 | internal source build | CMake | Selectable when supported by the selected release/profile |
| `mpeghdec` | libmpeghdec / MPEG-H audio decoding | internal source-prepared | r3.0.3 | internal source build | CMake | Selectable when supported by the selected release/profile |
| `quirc` | libquirc / QR code decoding | internal source-prepared | 1.2 | internal source build | make | Selectable when supported by the selected release/profile |
| `klvanc` | libklvanc / Broadcast metadata | internal source-prepared | vid.obe.1.6.0 | internal source build | configure + make | Selectable when supported by the selected release/profile |
| `libtls` | libtls / Secure network access | internal source-prepared | 4.3.2 | internal source build | CMake | Selectable when supported by the selected release/profile |
| `libmfx` | libmfx / Legacy Intel Media SDK | internal source-prepared | 1.35.1 | internal source build | CMake | Selectable when supported by the selected release/profile; mutually exclusive with `libvpl` |
| `opencv` | OpenCV | internal source-prepared | 4.11.0 | internal source build | CMake (`core` + `imgproc`) | Selectable; pinned to OpenCV 4.11.0 because current MSYS2 OpenCV 5 no longer provides the legacy C API required by FFmpeg `--enable-libopencv` |
| `tensorflow` | TensorFlow / AI model inference | external SDK/import | 2.16.1 | external import | archive import | UI-disabled |

## Disabled, blocked, and restricted entries

The program keeps these states separate:

| State | Entries | Meaning |
|---|---|---|
| Globally UI-disabled | `tensorflow`, `vapoursynth` | registered but not selectable in the standard UI |
| Blocked without preparation path | `smbclient`, `openvino`, `torch`, `pocketsphinx`, `dc1394`, `decklink`, `cuda-nvcc` | no supported build/import path is implemented |
| Version/profile-restricted | examples include `lensfun`, `onnxruntime`, `svtjpegxs`, `libvpl`, `libmfx` | availability depends on FFmpeg release, MSYS2 shell profile, package/API requirement, or pkg-config result; on 9.0.1, `onnxruntime` works on `ucrt64`/`clang64` while `lensfun` is unavailable |

`libmfx` is not a blocked entry. It is an implemented internal source-prepared legacy Intel backend and is normalized against `libvpl`.

## Full library library catalog

The following table lists the complete library library catalog before profile-specific runtime filtering. Package names use `<profile>` to mean the active MSYS2 shell profile prefix, for example `mingw-w64-ucrt-x86_64`, `mingw-w64-x86_64`, or `mingw-w64-clang-x86_64`.

### Included by default (official FFmpeg source)

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `ffmpeg-program` | ffmpeg.exe | included FFmpeg component | Built in | included | none | included in FFmpeg source | The main command-line media processor for conversion, filtering, recording, streaming, and packaging. |
| `ffprobe-program` | ffprobe.exe | included FFmpeg component | Built in | included | none | included in FFmpeg source | Inspects media files and reports streams, codecs, metadata, chapters, durations, and technical structure. |
| `libavutil` | libavutil | included FFmpeg component | Built in | included | none | included in FFmpeg source | Shared FFmpeg utility code used across codecs, filters, formats, and tools. |
| `libavcodec` | libavcodec | included FFmpeg component | Built in | included | none | included in FFmpeg source | FFmpeg’s built-in codec library for decoding and encoding many media formats. |
| `libavformat` | libavformat | included FFmpeg component | Built in | included | none | included in FFmpeg source | FFmpeg’s built-in container library for reading and writing formats such as MP4, MOV, MKV, WAV, and MPEG-TS. |
| `libavfilter` | libavfilter | included FFmpeg component | Built in | included | none | included in FFmpeg source | FFmpeg’s built-in filtering library for video, audio, subtitles, scaling, overlays, and effects. |
| `libswscale` | libswscale | included FFmpeg component | Built in | included | none | included in FFmpeg source | Converts image sizes and pixel formats. |
| `libswresample` | libswresample | included FFmpeg component | Built in | included | none | included in FFmpeg source | Converts audio sample rates, sample formats, and channel layouts. |
| `native-codecs` | Native FFmpeg codecs | included FFmpeg component | Built in | included | none | included in FFmpeg source | Uses FFmpeg’s built-in codec support before adding external codec libraries. |
| `native-formats` | Native formats and muxers | included FFmpeg component | Built in | included | none | included in FFmpeg source | Uses FFmpeg’s built-in readers and writers for many media containers. |

### Video encoders

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `x264` | x264 | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libx264` | `<profile>-libx264` | Supports H.264 software encoding. A leading H.264/AVC encoder with strong quality, performance, and broad compatibility. |
| `x265` | x265 | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libx265` | `<profile>-x265` | Supports HEVC/H.265 software encoding for smaller files than H.264 at similar visual quality. Useful when compression efficiency matters more than playback universality. |
| `svt-av1` | SVT-AV1 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libsvtav1` | `<profile>-svt-av1` | Produces AV1 video with practical software encoding speed. Commonly chosen for modern compression at usable throughput. |
| `libvpx` | libvpx | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libvpx` | `<profile>-libvpx` | Produces VP8 and VP9 video, commonly used for WebM and web-oriented delivery. |
| `aom` | AOM AV1 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libaom` | `<profile>-aom` | Supports AV1 encoding through the reference AV1 codec family. Strong quality and standards-oriented behavior, usually chosen when output quality matters more than encoding speed. |
| `openh264` | OpenH264 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopenh264` | `<profile>-openh264` | Provides H.264 software encoding with a simpler encoder profile. Useful when basic H.264 output is needed without x264’s full decision depth. |
| `rav1e` | rav1e | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-librav1e` | `<profile>-rav1e` | Provides AV1 software encoding with a Rust-based encoder. Useful for comparing AV1 speed, quality, and encoder behavior. |
| `xvid` | Xvid | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libxvid` | `<profile>-xvidcore` | Produces older MPEG-4 Part 2 video for legacy players and devices. |
| `theora` | libtheora | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libtheora` | `<profile>-libtheora` | Produces Ogg Theora video for older open-video compatibility. |
| `kvazaar` | Kvazaar (HEVC) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libkvazaar` | `<profile>-kvazaar` | Provides HEVC/H.265 software encoding with an open research-oriented encoder. Useful when HEVC output is needed through a non-x265 path. |
| `xeve` | XEVE (EVC) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libxeve` | `<profile>-xeve` | Produces EVC video for testing and newer codec workflows. |
| `xeveb` | XEVE base profile (EVC) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libxeveb` | `<profile>-xeve` | Enables MPEG-5 EVC encoding through the EVC base profile, the royalty-free subset of EVC. Uses the same XEVE package as the main `xeve` entry. |
| `oapv` | liboapv (APV) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-liboapv` | `<profile>-openapv` | Supports APV video encoding for specialized professional video workflows. |
| `xavs` | libxavs (AVS1) | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libxavs` | `<profile>-xavs` | Encodes video in the AVS1 (Chinese AVS) format. Useful for interoperating with AVS1 broadcast or archival material. This is GPL, so the build moves to a GPL license profile. |
| `vvenc` | vvenc (VVC/H.266) | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-libvvenc` | prepared from pinned source/import procedure | Produces VVC/H.266 video for very high compression efficiency experiments and next-generation codec testing. |
| `xavs2` | xavs2 | internal source-prepared | Selectable: source/import preparation | GPL | `--enable-libxavs2` | prepared from pinned source/import procedure | Produces AVS2 video for workflows that need this regional video standard. Useful for compatibility, testing, and specialized distribution. |

### Hardware encoders

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `nvenc` | NVIDIA NVENC | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-ffnvcodec` | `<profile>-ffnvcodec-headers` | Provides NVIDIA GPU video encoding for fast H.264, HEVC, and AV1 output on supported NVIDIA systems. |
| `libvpl` | Intel oneVPL / Quick Sync | native MSYS2 package | Selectable when supported by the selected FFmpeg release; mutually exclusive with `libmfx` | LGPL | `--enable-libvpl` | `<profile>-libvpl` | Provides Intel Quick Sync hardware acceleration through oneVPL. It is the preferred Intel backend on FFmpeg 6.0 and newer. |
| `libmfx` | libmfx / Legacy Intel Media SDK | internal source-prepared | Selectable when supported by the selected FFmpeg release; mutually exclusive with `libvpl` | LGPL | `--enable-libmfx` | prepared from pinned `mfx_dispatch` source procedure | Provides the legacy Intel Media SDK dispatcher. It is mainly useful for older FFmpeg release lines where oneVPL is not available, and it must not be enabled together with `libvpl`. |
| `amf` | AMD AMF | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-amf` | `<profile>-amf-headers` | Provides AMD GPU video encoding for faster H.264, HEVC, and AV1 output on supported Radeon systems. Useful when speed and low CPU use matter. |

### Video decoders

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `dav1d` | dav1d | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libdav1d` | `<profile>-dav1d` | Provides fast AV1 video decoding. Useful for smooth playback, inspection, and conversion of AV1 files. |
| `xevd` | XEVD (EVC dec) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libxevd` | `<profile>-xevd` | Decodes EVC video streams. Useful for reading newer Essential Video Coding material. |
| `xevdb` | XEVD base profile (EVC dec) | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libxevdb` | `<profile>-xevd` | Enables MPEG-5 EVC decoding through the EVC base profile. Uses the same XEVD package as the main `xevd` entry. |
| `davs2` | libdavs2 / AVS2 decoding | internal source-prepared | Selectable: source/import preparation | GPL | `--enable-libdavs2` | prepared from pinned source/import procedure | Decodes AVS2 video streams. Useful for regional, broadcast, or compatibility workflows involving AVS2 material. |
| `uavs3d` | libuavs3d / AVS3 decoding | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-libuavs3d` | prepared from pinned source/import procedure | Decodes AVS3 video streams. Useful for AVS3 test material, regional standards, and compatibility workflows. |
| `lcevc-dec` | liblcevc-dec / LCEVC decoding | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-liblcevc-dec` | prepared from pinned source/import procedure | Decodes LCEVC enhancement layers that improve a base video stream. Useful when content uses layered video enhancement. |
| `avisynthplus` | AviSynth+ / Scripted video processing | internal source-prepared | Selectable: source/import preparation | GPL | `--enable-avisynth` | prepared from pinned source/import procedure | Opens AviSynth script-based video processing chains directly. Useful when a workflow already depends on AviSynth filters or scripted frame processing. |

### Image codecs

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `png` | libpng | native MSYS2 package | Selectable: MSYS2 package | LGPL | none | `<profile>-libpng` | Handles PNG image input and output for lossless still images and image sequences. |
| `webp` | WebP | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libwebp` | `<profile>-libwebp` | Reads and writes WebP images used widely on websites. Useful for still images, thumbnails, and image sequences. |
| `openjpeg` | OpenJPEG | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopenjpeg` | `<profile>-openjpeg2` | Reads and writes JPEG 2000 images used in archival, cinema, and professional imaging workflows. |
| `libjxl` | JPEG XL | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libjxl` | `<profile>-libjxl` | Reads and writes JPEG XL images for modern high-quality still-image compression. Useful for image sequences and advanced image workflows. |
| `rsvg` | librsvg | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-librsvg` | `<profile>-librsvg` | Renders SVG vector graphics into normal image frames. Useful for scalable logos, overlays, and graphic assets. |
| `snappy` | Snappy | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libsnappy` | `<profile>-snappy` | Provides fast Snappy data compression for formats that use it internally. |
| `lcms2` | Little CMS 2 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-lcms2` | `<profile>-lcms2` | Applies color profile management for more accurate color conversion. Useful when preserving color appearance matters. |
| `svtjpegxs` | SVT JPEG XS | native MSYS2 package | Selectable on FFmpeg 9.0.1; release-dependent on older lines | LGPL | `--enable-libsvtjpegxs` | `<profile>-svt-jpeg-xs`, `git`, `<profile>-cmake`, `<profile>-ninja`, `<profile>-yasm` | Produces JPEG XS video for low-latency professional media workflows. |

### Filters and processing

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `zimg` | zimg | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libzimg` | `<profile>-zimg` | Improves image resizing, pixel-format conversion, and color handling. Useful for clean scaling and careful video conversion. |
| `libplacebo` | libplacebo | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libplacebo`, `--enable-vulkan` | `<profile>-libplacebo`, `<profile>-vulkan-loader`, `<profile>-vulkan-headers` | Provides high-quality GPU video processing for scaling, color conversion, tone mapping, and rendering paths. |
| `vmaf` | libvmaf | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libvmaf` | `<profile>-vmaf` | Measures perceived video quality by comparing an encoded result with a reference. Useful for evaluating compression choices. |
| `vidstab` | libvidstab | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libvidstab` | `<profile>-vid.stab` | Reduces camera shake and makes handheld footage look smoother. |
| `opencolorio` | OpenColorIO | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopencolorio` | `<profile>-opencolorio` | Applies studio-grade color management for film, animation, and post-production workflows. |
| `cairo` | Cairo | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-cairo` | `<profile>-cairo` | Provides 2D vector drawing support for generated graphics, overlays, and filter visuals. Useful when media processing needs drawn shapes or rendered graphic elements. |
| `opencl` | OpenCL | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-opencl` | `<profile>-opencl-icd`, `<profile>-opencl-headers` | Runs supported compute filters on OpenCL-capable GPUs or accelerators. Useful for offloading heavy image processing. |
| `shaderc` | libshaderc | native MSYS2 package | Selectable package/tool on FFmpeg 9.0.1 | LGPL | none on FFmpeg 9.0.1 | `<profile>-shaderc` | FFmpeg 9 removed `--enable-libshaderc`; the package can still be installed as a Vulkan shader build tool. Older FFmpeg releases may still use the configure switch. |
| `glslang` | glslang | native MSYS2 package | Selectable package/tool on FFmpeg 9.0.1 | LGPL | none on FFmpeg 9.0.1 | `<profile>-glslang` | FFmpeg 9 removed `--enable-libglslang`; the package can still be installed as a Vulkan shader build tool. Older FFmpeg releases may still use the configure switch. |
| `frei0r` | frei0r | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-frei0r` | `<profile>-frei0r-plugins` | Provides extra creative video effects and filter plugins. Useful for stylized processing beyond the standard filter set. |
| `opencv` | OpenCV | internal source-prepared | Selectable: pinned source build | LGPL | `--enable-libopencv` | OpenCV 4.11.0 source (`core` + `imgproc`) | Provides the legacy OpenCV C API required by FFmpeg. The builder uses pinned OpenCV 4.11.0 because current MSYS2 OpenCV 5 removed the required legacy headers/API and `opencv4.pc` compatibility path. |
| `ladspa` | LADSPA | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-ladspa` | `<profile>-ladspa-sdk`, `<profile>-dlfcn` | Loads LADSPA audio effect plugins for extra audio processing. Useful when existing plugin chains need to be used in media conversion. |
| `lv2` | LV2 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-lv2` | `<profile>-lilv` | Loads LV2 audio plugins for more advanced audio effects and processing chains. |
| `qrencode` | libqrencode | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libqrencode` | `<profile>-qrencode` | Generates QR codes that can be placed into images or video frames. |
| `cuda-nvcc` | CUDA NVCC (NVIDIA GPU filters) | external SDK/import | Blocked: no preparation path | LGPL | `--enable-cuda-nvcc` | no supported preparation path | Compiles FFmpeg's CUDA-accelerated filters with NVIDIA's nvcc compiler. nvcc ships only in the proprietary CUDA Toolkit, which is not available as an MSYS2 package. |
| `lensfun` | lensfun | native MSYS2 package | Unavailable in the current FFmpeg 9.0.1 catalog | LGPL | `--enable-liblensfun` | `<profile>-lensfun` | Corrects lens distortion and vignetting. FFmpeg exposes the switch, but the current package/API combination is intentionally blocked for 9.0.1. |
| `vapoursynth` | VapourSynth / Scripted video processing | native MSYS2 package | UI-disabled | LGPL | `--enable-vapoursynth` | `<profile>-vapoursynth` | Opens VapourSynth script-based video processing chains. Useful when advanced scripted filtering must feed frames into a media workflow. |

### Audio

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `opus` | Opus | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopus` | `<profile>-opus` | Produces modern Opus audio for speech, music, streaming, and low-latency communication. |
| `fdk-aac` | Fraunhofer FDK AAC | native MSYS2 package | Selectable: MSYS2 package | nonfree | `--enable-libfdk-aac`, `--enable-nonfree` | `<profile>-fdk-aac` | Produces high-quality AAC audio for MP4, mobile devices, streaming, and web playback. Often chosen when AAC quality is important. |
| `mp3lame` | LAME MP3 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libmp3lame` | `<profile>-lame` | Produces MP3 audio with broad playback compatibility. Useful when the output must work on almost any device or player. |
| `vorbis` | Vorbis | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libvorbis` | `<profile>-libvorbis` | Produces Ogg Vorbis audio with good open-format compatibility. Useful for music and general audio in Ogg workflows. |
| `soxr` | SoX Resampler | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libsoxr` | `<profile>-libsoxr` | Provides high-quality audio sample-rate conversion. Useful when changing audio rates while preserving clarity. |
| `rubberband` | Rubber Band | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-librubberband` | `<profile>-rubberband` | Changes audio speed and pitch with higher-quality time-stretching. Useful for music, dialogue timing, and pitch adjustment. |
| `chromaprint` | Chromaprint | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-chromaprint` | `<profile>-chromaprint` | Generates compact audio fingerprints for track matching and identification. Useful for recognizing audio even when filenames or metadata are missing. |
| `twolame` | TwoLAME | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libtwolame` | `<profile>-twolame` | Produces MP2 audio for broadcast, DVD, and older media workflows. |
| `speex` | Speex | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libspeex` | `<profile>-speex` | Handles older Speex speech audio. Useful for legacy voice recordings and communication archives. |
| `opencore-amr` | OpenCORE AMR | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopencore-amrnb`, `--enable-libopencore-amrwb` | `<profile>-opencore-amr` | Reads and writes AMR speech audio used by older mobile and voice systems. |
| `vo-amrwbenc` | VisualOn AMR-WB encoder | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libvo-amrwbenc` | `<profile>-vo-amrwbenc` | Produces AMR-WB wideband speech audio for mobile-style voice compatibility. |
| `gsm` | GSM | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libgsm` | `<profile>-gsm` | Handles legacy GSM speech audio. Useful for old voice recordings, telephony archives, and compatibility cases. |
| `lc3` | LC3 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-liblc3` | `<profile>-liblc3` | Handles LC3 audio used in modern low-complexity communication audio. Useful for Bluetooth- and voice-oriented compatibility. |
| `ilbc` | iLBC | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libilbc` | `<profile>-libilbc` | Supports iLBC speech audio used in older internet voice communication. Useful for voice-call recordings and compatibility workflows. |
| `whisper` | whisper.cpp | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-whisper` | `<profile>-whisper.cpp`, `<profile>-ggml` | Converts spoken audio into text using whisper.cpp speech recognition. Useful for transcription and subtitle generation. |
| `mysofa` | libmysofa | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libmysofa` | `<profile>-libmysofa` | Provides spatial audio filter data for headphone-based 3D audio rendering. |
| `bs2b` | libbs2b | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libbs2b` | `<profile>-libbs2b` | Applies headphone crossfeed to make stereo listening feel less hard-panned and more speaker-like. Useful for headphone-focused audio processing. |
| `gme` | Game Music Emu | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libgme` | `<profile>-libgme` | Reads classic game music formats from older consoles and systems. Useful for playback or conversion of chiptune-style sources. |
| `shine` | Shine MP3 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libshine` | `<profile>-shine` | Produces MP3 audio using a simple fixed-point encoder. Useful for constrained or compatibility-focused cases. |
| `codec2` | Codec 2 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libcodec2` | `<profile>-codec2` | Supports very low-bitrate speech coding for radio, voice, and communication experiments. Useful when intelligible speech must fit into extremely small bitrates. |
| `mpeghdec` | libmpeghdec / MPEG-H audio decoding | internal source-prepared | Selectable: source/import preparation | nonfree | `--enable-libmpeghdec`, `--enable-nonfree` | prepared from pinned source/import procedure | Decodes MPEG-H 3D Audio for immersive and object-based audio content. |
| `pocketsphinx` | PocketSphinx (speech recognition) | internal source-prepared | Blocked: no preparation path | LGPL | `--enable-pocketsphinx` | no supported preparation path | Adds the asr speech-recognition audio filter using CMU PocketSphinx. It cannot currently be built: FFmpeg's asr filter is incompatible with present-day PocketSphinx releases, so selecting it blocks the build. |

### Subtitles and text

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `ass` | libass | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libass` | `<profile>-libass` | Renders advanced ASS/SSA subtitles with fonts, colors, placement, outlines, and animation-like effects. Commonly used for complex subtitle presentation. |
| `freetype` | FreeType | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libfreetype` | `<profile>-freetype` | Renders font glyphs for subtitles, overlays, and text-based filters. Essential when text must appear as real shaped graphics. |
| `fontconfig` | Fontconfig | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libfontconfig` | `<profile>-fontconfig` | Finds installed fonts so subtitle and text rendering can use the right typefaces. Useful for reliable text appearance. |
| `harfbuzz` | HarfBuzz | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libharfbuzz` | `<profile>-harfbuzz` | Shapes complex text so letters combine, reorder, and position correctly. Important for subtitles in many non-Latin scripts. |
| `fribidi` | FriBidi | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libfribidi` | `<profile>-fribidi` | Handles right-to-left text direction for languages such as Arabic and Hebrew. Useful when subtitles or overlays contain bidirectional text. |
| `aribcaption` | libaribcaption | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libaribcaption` | `<profile>-libaribcaption` | Reads Japanese broadcast captions with better modern ARIB caption handling. Useful for TV recordings and broadcast transport streams. |
| `aribb24` | libaribb24 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libaribb24` | `<profile>-aribb24` | Handles ARIB B24 captions used in Japanese broadcast material. Useful for preserving or processing older broadcast subtitle data. |
| `zvbi` | libzvbi | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libzvbi` | `<profile>-zvbi` | Extracts teletext and VBI data from older broadcast sources. Useful for hidden captions, pages, and broadcast text information. |

### Disc and device input

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `bluray` | libbluray | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libbluray` | `<profile>-libbluray` | Reads Blu-ray disc structures for inspection, conversion, or extraction workflows. Useful when the source is organized as Blu-ray folders rather than a single media file. |
| `dvdread` | libdvdread | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libdvdread` | `<profile>-libdvdread` | Reads DVD media structures for extraction, inspection, or conversion. Useful when the source is a DVD folder or disc layout. |
| `dvdnav` | libdvdnav | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libdvdnav` | `<profile>-libdvdnav`, `<profile>-libdvdread` | Reads DVD-Video navigation structures such as menus, titles, chapters, and program chains. Useful for disc-style DVD workflows. |
| `openmpt` | libopenmpt | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libopenmpt` | `<profile>-libopenmpt` | Reads tracker module music with accurate playback behavior. Useful for old game, demo-scene, and tracker music formats. |
| `sdl2` | SDL2 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-sdl2` | `<profile>-SDL2` | Provides simple media output and preview support through SDL2. Useful for playback-style testing and display workflows. |
| `openal` | OpenAL | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-openal` | `<profile>-openal` | Provides additional audio device input and output paths. Useful for workflows involving live audio devices. |
| `cdio` | libcdio | native MSYS2 package | Selectable: MSYS2 package | GPL | `--enable-libcdio` | `<profile>-libcdio`, `<profile>-libcdio-paranoia` | Reads disc-based CD input for audio extraction and media inspection workflows. Useful when the source is an actual CD rather than copied audio files. |
| `modplug` | libmodplug | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libmodplug` | `<profile>-libmodplug` | Reads old tracker module music formats. Useful for converting or playing legacy scene and game-style music files. |
| `jack` | JACK | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libjack` | `<profile>-jack2` | Connects to JACK audio routing for studio-style capture and playback workflows. Useful in professional Linux audio environments. |
| `pulse` | PulseAudio | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libpulse` | `<profile>-pulseaudio` | Connects to PulseAudio for Linux desktop audio capture or playback workflows. |
| `caca` | libcaca | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libcaca` | `<profile>-libcaca` | Converts video into colored text-mode visuals. Mostly useful for experiments, terminal display, and unusual preview effects. |
| `opengl` | OpenGL | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-opengl` | `<profile>-mesa` | Adds the OpenGL output device for hardware-accelerated video display through OpenGL. Provided by the Mesa package. |
| `dc1394` | libdc1394 (IEEE 1394 camera) | internal source-prepared | Blocked: no preparation path | LGPL | `--enable-libdc1394` | no supported preparation path | Captures video from IEEE 1394 (FireWire) cameras. Blocked on Windows: no MSYS2 package, and the only Windows build needs a proprietary FireWire kernel driver plus FireWire hardware, so the result would not be portable. |
| `decklink` | DeckLink (Blackmagic capture/playback) | external SDK/import | Blocked: no preparation path | LGPL | `--enable-decklink` | no supported preparation path | Adds Blackmagic DeckLink capture and playback support. It builds against the proprietary DeckLink SDK headers, which cannot be redistributed as an MSYS2 package. |

### Network

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `srt` | SRT | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libsrt` | `<profile>-srt` | Provides SRT streaming for stable live transport over unreliable networks. |
| `rtmp` | librtmp | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-librtmp` | `<profile>-rtmpdump` | Provides RTMP streaming compatibility for older live-streaming servers and workflows. |
| `rist` | librist | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-librist` | `<profile>-librist` | Provides Reliable Internet Stream Transport for professional live streaming over unstable networks. |
| `ssh` | libssh | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libssh` | `<profile>-libssh` | Reads and writes media through SSH and SFTP-style remote access. |
| `zmq` | ZeroMQ | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libzmq` | `<profile>-zeromq` | Enables message-based runtime control for supported processing workflows. Useful for automation and interactive control. |
| `rabbitmq` | RabbitMQ-C | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-librabbitmq` | `<profile>-rabbitmq-c` | Connects media workflows to RabbitMQ message queues. Useful for automated processing systems. |
| `smbclient` | libsmbclient / SMB network file access | external SDK/import | Blocked: no preparation path | GPL | `--enable-libsmbclient` | no supported preparation path | Reads from and writes to SMB/CIFS network shares. Useful for media stored on Windows-style network folders. It cannot be built on Windows yet, so it is listed last in this section and stays unavailable until a Windows version of libsmbclient exists. |

### Secure network (TLS)

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `openssl` | OpenSSL | native MSYS2 package | Selectable: MSYS2 package | nonfree | `--enable-openssl` | `<profile>-openssl` | Provides encrypted network connections for HTTPS and other secure media protocols. |
| `gnutls` | GnuTLS | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-gnutls` | `<profile>-gnutls` | Provides encrypted network connections for HTTPS and other TLS-based media access. Useful for secure streaming and remote sources. |
| `mbedtls` | mbedTLS / Secure network access | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-mbedtls` | `<profile>-mbedtls` | Provides lightweight TLS support for encrypted network media access. Useful when a smaller security backend is preferred. |
| `libtls` | libtls / Secure network access | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-libtls` | prepared from pinned source/import procedure | Provides TLS-encrypted network communication through a compact TLS interface. Useful for secure network media access. |

### OCR

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `tesseract` | Tesseract OCR | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libtesseract` | `<profile>-tesseract-ocr` | Reads visible text from images or video frames. Useful for extracting burned-in titles, signs, captions, or document text. |

### AI support

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `onnxruntime` | ONNX Runtime / AI model inference | native MSYS2 package | Selectable on FFmpeg 9.0.1 for `ucrt64`/`clang64`; unavailable on `mingw64` | LGPL | `--enable-libonnxruntime` | `<profile>-onnxruntime` (`ucrt64`/`clang64`) | Runs FFmpeg 9 ONNX Runtime integrations. The configure script adds MSYS2's nested `include/onnxruntime` directory when this library is enabled. |
| `openvino` | OpenVINO / AI model inference | external SDK/import | Blocked: no preparation path | LGPL | `--enable-libopenvino` | no supported preparation path | Runs supported AI inference filters with Intel-oriented acceleration. Useful for model-based video or image processing. |
| `torch` | Torch / libtorch | external SDK/import | Blocked: no preparation path | LGPL | `--enable-libtorch` | no supported preparation path | Runs supported deep-learning filters through Torch-based model execution. Useful for PyTorch-style inference workflows. |
| `tensorflow` | TensorFlow / AI model inference | external SDK/import | UI-disabled | LGPL | `--enable-libtensorflow` | prepared from pinned source/import procedure | Runs supported deep-learning filters through the TensorFlow C API. Useful for model-based image or video analysis. |

### Support libraries

| ID | Display name | Track | Status | License | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `xml2` | libxml2 | native MSYS2 package | Selectable: MSYS2 package | LGPL | `--enable-libxml2` | `<profile>-libxml2` | Reads structured XML data used by some media formats, subtitles, manifests, and metadata workflows. |
| `quirc` | libquirc / QR code decoding | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-libquirc` | prepared from pinned source/import procedure | Decodes QR codes from video frames or images. Useful for automation, scanning, and visual metadata workflows. |
| `klvanc` | libklvanc / Broadcast metadata | internal source-prepared | Selectable: source/import preparation | LGPL | `--enable-libklvanc` | prepared from pinned source/import procedure | Processes vertical ancillary data used in broadcast video. Useful for metadata carried alongside video lines. |



## License summary

The program derives the final license profile from selected libraries and final configure flags instead of asking the user to manually choose `--enable-gpl`, `--enable-nonfree`, or `--enable-version3`. GPL entries change the local license profile to `gpl-local`. Nonfree entries change the local license profile to `nonfree-local` and trigger redistribution warnings. Version-3-sensitive entries add `--enable-version3` automatically where required.

| License | Count | Representative entries |
|---|---:|---|
| included | 10 | `ffmpeg-program`, `ffprobe-program`, `libavutil`, `libavcodec`, `libavformat`, `libavfilter`, `libswscale`, `libswresample`, `native-codecs`, `native-formats` |
| LGPL | 98 | `svt-av1`, `libvpx`, `aom`, `openh264`, `rav1e`, `theora`, `kvazaar`, `xeve`, `xeveb`, `oapv`, `vvenc`, `nvenc` |
| GPL | 14 | `x264`, `x265`, `xvid`, `xavs`, `xavs2`, `davs2`, `avisynthplus`, `frei0r`, `rubberband`, `zvbi`, `dvdread`, `dvdnav` |
| nonfree | 3 | `fdk-aac`, `mpeghdec`, `openssl` |



## Practical interpretation

For ordinary Windows users, the meaningful supported surface is the set that the program can select, prepare, review, and pass to FFmpeg for the chosen release/profile. That includes official FFmpeg-source entries, native MSYS2-package entries, and implemented internal source-prepared entries.

The library catalog deliberately avoids claiming that every upstream FFmpeg library can be enabled just because a configure flag exists. Entries such as `cuda-nvcc`, DeckLink, OpenVINO, Torch, TensorFlow, VapourSynth, SVT JPEG XS, ONNX Runtime, PocketSphinx, lensfun, and SMB are treated conservatively because they are blocked, disabled, version/profile-restricted, or dependent on external SDK/runtime assumptions.

`libmfx` is different from that blocked group. It is an implemented legacy Intel backend and is normalized against `libvpl`.
