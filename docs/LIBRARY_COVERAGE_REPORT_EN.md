# Library coverage report

This report describes the FFmpeg library coverage of this program as represented by the current source tree. It is based on the backend catalog, frontend preset logic, source-preparation recipes, profile-availability filters, and localization descriptions.

The word “covers” is used carefully here. A library may be listed in the catalog, selectable through the normal UI, source-prepared by the builder, intentionally hidden, or blocked because the current program does not yet have a safe build/import path.

## Executive summary

- Total catalog rows: **125**.

- Locked official FFmpeg-source rows: **10**.

- Native MSYS2 package-track rows: **96**.

- Internal source-prepared rows: **12**.

- External SDK/import-track rows: **7**.

- Normal UI-selectable rows, including locked built-ins: **112**.

- Normal UI-selectable external/additional rows, excluding locked built-ins: **102**.

- UI-disabled rows: **5** — `lensfun`, `svtjpegxs`, `vapoursynth`, `tensorflow`, `onnxruntime`.

- Blocked rows with no normal build/import recipe: **8** — `smbclient`, `openvino`, `torch`, `libmfx`, `pocketsphinx`, `dc1394`, `decklink`, `cuda-nvcc`.


## What the coverage model means

The program does not treat every FFmpeg `--enable-*` option as equally supported. It separates the library catalog into several practical tracks:

| Track | Meaning in this program | Normal implication |
|---|---|---|
| Included | Official FFmpeg components that are part of a normal source build. | Locked on; no external package or `--enable-lib...` flag is added. |
| Native | A named external row that maps to MSYS2 package names and FFmpeg configure flags. | The planner installs the profile-matching MSYS2 packages before configure. |
| Internal | No suitable prebuilt package is assumed; the program prepares a pinned source archive itself. | The planner downloads/verifies/builds/imports the library before FFmpeg configure. |
| External | The library depends on an external SDK/import path or a currently unsafe/unsupported path. | Not normally selectable unless a recipe exists and the UI allows it. |

This distinction is important. The catalog is intentionally transparent: it shows not only what can be selected today, but also rows that are known, tracked, and blocked because enabling them without a safe preparation path would produce misleading or broken builds.

## Coverage by category

| Category | Catalog rows | Notes |
|---|---:|---|
| Included by default (official FFmpeg source) | 10 | Locked FFmpeg program/library/native rows. |
| Video encoders | 16 | Software encoders plus newer regional/next-generation codec bindings. |
| Hardware encoders | 4 | Header/API enablement for GPU encoder families; not a guarantee that local hardware/driver supports the encoder. |
| Video decoders | 7 | Modern and regional decoder libraries, including source-prepared AVS/LCEVC/AviSynth paths. |
| Image codecs | 8 | Image and still-picture codec/helper libraries; libpng is package-only because FFmpeg has no separate `--enable-libpng` flag row here. |
| Filters and processing | 17 | Scaling, quality metrics, GPU/shader, color, plugin, QR and scripted-processing support. |
| Audio | 22 | Audio codecs, resamplers, speech/transcription-related rows and spatial/audio-processing helpers. |
| Subtitles and text | 8 | Subtitle rendering, font shaping and broadcast caption handling. |
| Disc and device input | 14 | Disc formats, device/display/input integrations and capture SDK rows. |
| Network | 7 | Transport/protocol libraries excluding TLS backend choice. |
| Secure network (TLS) | 4 | Mutually exclusive TLS backend options. |
| OCR | 1 | OCR support through Tesseract. |
| AI support | 4 | AI/model-inference related FFmpeg hooks; currently conservative/mostly disabled or blocked. |
| Support libraries | 3 | General support libraries and metadata/QR helper libraries. |

## Preset coverage

Public library presets are cumulative broadening tiers: each higher tier includes the locked FFmpeg rows and the rows from the earlier practical tier, then adds a focused set of additional libraries.

| Preset | Meaning | Libraries added beyond previous tier |
|---|---|---|
| Minimal | Only the locked official FFmpeg components. | none |
| Default | First-run practical build: common software encoders/decoders, hardware encoder headers, major audio libraries, subtitle stack, common processing and network helpers. | `nvenc`, `amf`, `qsv`, `x264`, `x265`, `libvpx`, `aom`, `svt-av1`, `dav1d`, `theora`, `xvid`, `opus`, `vorbis`, `mp3lame`, `gsm`, `speex`, `opencore-amr`, `vo-amrwbenc`, `rubberband`, `openjpeg`, `webp`, `freetype`, `fontconfig`, `fribidi`, `harfbuzz`, `ass`, `zimg`, `vmaf`, `vidstab`, `srt`, `ssh`, `zmq`, `openal`, `sdl2`, `gme`, `openmpt` |
| Maximum Efficiency | Default plus quality/efficiency audio helpers. | `fdk-aac`, `soxr` |
| Maximum Compatibility | Efficiency plus broader codecs, caption formats, protocols, and image helpers. | `rav1e`, `openh264`, `ilbc`, `twolame`, `xevd`, `shine`, `codec2`, `lc3`, `snappy`, `rsvg`, `zvbi`, `aribb24`, `aribcaption`, `rtmp` |
| Maximum Audio/Video Editor | Compatibility plus editing, filtering, color, plugin, QR, subtitle, OCR/transcription-adjacent and image workflow libraries. | `libjxl`, `png`, `libplacebo`, `frei0r`, `xml2`, `mysofa`, `bs2b`, `lcms2`, `shaderc`, `cairo`, `opencv`, `opencolorio`, `ladspa`, `lv2`, `chromaprint`, `qrencode`, `whisper` |
| Maximum Full | Editor plus advanced packaged features, disc/device libraries, additional network/TLS choice, OCR, and display/compute helpers. | `kvazaar`, `bluray`, `dvdread`, `dvdnav`, `cdio`, `modplug`, `openssl`, `rist`, `rabbitmq`, `tesseract`, `jack`, `pulse`, `caca`, `opencl` |


### Extended library toggle

The Extended toggle adds source-prepared internal libraries to selected broadening presets. Minimal and Default are intentionally unaffected.

| Base preset with Extended enabled | Added source-prepared libraries |
|---|---|
| Maximum Efficiency | `vvenc`, `uavs3d`, `lcevc-dec` |
| Maximum Compatibility | `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `xavs2` |
| Maximum Audio/Video Editor | `avisynthplus`, `lcevc-dec` |
| Maximum Full | `vvenc`, `xavs2`, `davs2`, `uavs3d`, `lcevc-dec`, `avisynthplus`, `mpeghdec` |


## Mutual-exclusion and normalization rules

Some rows are cataloged separately because FFmpeg exposes separate bindings, but the final selection must still be normalized before configure. The UI and planner enforce these relationships:

- TLS backend: `openssl`, `gnutls`, `mbedtls`, and `libtls` are pick-one choices. Priority normalization keeps OpenSSL first, then GnuTLS, then mbedTLS, then libtls.

- Shader compiler: `shaderc` and `glslang` are pick-one choices; if both appear, `shaderc` is kept.

- EVC decoder binding: `xevd` and `xevdb` cannot be enabled together; full-profile `xevd` is kept by normalization.

- EVC encoder binding: `xeve` and `xeveb` cannot be enabled together; full-profile `xeve` is kept by normalization.


## Source-prepared and imported libraries

These rows have implemented preparation entries in `libraryItemSpecs` and pinned source/import metadata in `library-sources.json` for FFmpeg 8.1.2. Source-prepared rows are not casual configure flags: the program downloads a specific archive, verifies its SHA-256 hash, builds or imports it, verifies headers/libraries, and only then lets FFmpeg configure consume it.

| ID | Display name | Track | Version pin | Method | Build system / import style | Normal UI state |
|---|---|---|---|---|---|---|
| `avisynthplus` | AviSynth+ / Scripted video processing | internal source-prepared | 3.7.5 | internal source build | CMake | Normal selectable: prepared source/import track |
| `davs2` | libdavs2 / AVS2 decoding | internal source-prepared | 1.7 | internal source build | configure + make | Normal selectable: prepared source/import track |
| `xavs2` | xavs2 | internal source-prepared | 1.4 | internal source build | configure + make | Normal selectable: prepared source/import track |
| `uavs3d` | libuavs3d / AVS3 decoding | internal source-prepared | master | internal source build | CMake | Normal selectable: prepared source/import track |
| `lcevc-dec` | liblcevc-dec / LCEVC decoding | internal source-prepared | 4.2.0 | internal source build | CMake | Normal selectable: prepared source/import track |
| `vvenc` | vvenc (VVC/H.266) | internal source-prepared | 1.14.0 | internal source build | CMake | Normal selectable: prepared source/import track |
| `mpeghdec` | libmpeghdec / MPEG-H audio decoding | internal source-prepared | r3.0.3 | internal source build | CMake | Normal selectable: prepared source/import track |
| `quirc` | libquirc / QR code decoding | internal source-prepared | 1.2 | internal source build | make | Normal selectable: prepared source/import track |
| `klvanc` | libklvanc / Broadcast metadata | internal source-prepared | vid.obe.1.6.0 | internal source build | configure + make | Normal selectable: prepared source/import track |
| `libtls` | libtls / Secure network access | internal source-prepared | 4.3.2 | internal source build | CMake | Normal selectable: prepared source/import track |
| `tensorflow` | TensorFlow / AI model inference | external SDK/import | 2.16.1 | external import | archive import | UI-disabled |

## Intentionally disabled or blocked rows

A blocked or hidden row is not a failure of cataloging. In this project it usually means the program knows the FFmpeg option exists but refuses to present it as normally buildable until the path is compatible with the program principle: **do not patch or replace FFmpeg source merely to force compatibility, and do not ask the user to rely on an unsafe or non-portable SDK path without a dedicated preparation model.**

| ID | Display name | Track | Reason in current program |
|---|---|---|---|
| `cuda-nvcc` | CUDA NVCC (NVIDIA GPU filters) | external SDK/import | blocked because no normal build/import recipe is implemented; Compiles FFmpeg's CUDA-accelerated filters with NVIDIA's nvcc compiler. nvcc ships only in the proprietary CUDA Toolkit, which is not available as an MSYS2 package. |
| `dc1394` | libdc1394 (IEEE 1394 camera) | internal source-prepared | blocked because no normal build/import recipe is implemented; Captures video from IEEE 1394 (FireWire) cameras. Blocked on Windows: no MSYS2 package, and the only Windows build needs a proprietary FireWire kernel driver plus FireWire hardware, so the result would not be portable. |
| `decklink` | DeckLink (Blackmagic capture/playback) | external SDK/import | blocked because no normal build/import recipe is implemented; Adds Blackmagic DeckLink capture and playback support. It builds against the proprietary DeckLink SDK headers, which cannot be redistributed as an MSYS2 package. |
| `lensfun` | lensfun | native MSYS2 package | UI-disabled for normal users; Corrects lens distortion, vignetting, and other camera-lens artifacts. Useful for cleanup of footage from known lenses. |
| `libmfx` | libmfx (legacy Intel Media SDK) | external SDK/import | blocked because no normal build/import recipe is implemented; Legacy Intel hardware encode/decode through the old Media SDK dispatcher. FFmpeg removed support for this in 7.0, so it does not work in the later source this builder targets. Use Intel QSV (oneVPL) instead. |
| `onnxruntime` | ONNX Runtime / AI model inference | native MSYS2 package | UI-disabled for normal users; also unavailable for mingw64; Runs supported deep-learning filters through ONNX Runtime. Useful for model-based analysis or enhancement workflows. |
| `openvino` | OpenVINO / AI model inference | external SDK/import | blocked because no normal build/import recipe is implemented; Runs supported AI inference filters with Intel-oriented acceleration. Useful for model-based video or image processing. |
| `pocketsphinx` | PocketSphinx (speech recognition) | internal source-prepared | blocked because no normal build/import recipe is implemented; Adds the asr speech-recognition audio filter using CMU PocketSphinx. It cannot currently be built: FFmpeg's asr filter is incompatible with present-day PocketSphinx releases, so selecting it blocks the build. |
| `smbclient` | libsmbclient / SMB network file access | external SDK/import | blocked because no normal build/import recipe is implemented; Reads from and writes to SMB/CIFS network shares. Useful for media stored on Windows-style network folders. It cannot be built on Windows yet, so it is listed last in this section and stays unavailable until a Windows version of libsmbclient exists. |
| `svtjpegxs` | SVT JPEG XS | native MSYS2 package | UI-disabled for normal users; Produces JPEG XS video for low-latency professional media workflows. |
| `tensorflow` | TensorFlow / AI model inference | external SDK/import | UI-disabled for normal users; Runs supported deep-learning filters through the TensorFlow C API. Useful for model-based image or video analysis. |
| `torch` | Torch / libtorch | external SDK/import | blocked because no normal build/import recipe is implemented; Runs supported deep-learning filters through Torch-based model execution. Useful for PyTorch-style inference workflows. |
| `vapoursynth` | VapourSynth / Scripted video processing | native MSYS2 package | UI-disabled for normal users; Opens VapourSynth script-based video processing chains. Useful when advanced scripted filtering must feed frames into a media workflow. |

## Full library catalog

The following table is the complete catalog as represented in the current backend after excluding only profile-specific runtime filtering. Package names use `<profile>` to mean the active MSYS2 shell profile prefix, for example `mingw-w64-ucrt-x86_64`, `mingw-w64-x86_64`, or `mingw-w64-clang-x86_64`.

### Included by default (official FFmpeg source)

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `ffmpeg-program` | ffmpeg.exe | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | The main command-line media processor for conversion, filtering, recording, streaming, and packaging. |
| `ffprobe-program` | ffprobe.exe | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Inspects media files and reports streams, codecs, metadata, chapters, durations, and technical structure. |
| `libavutil` | libavutil | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Shared FFmpeg utility code used across codecs, filters, formats, and tools. |
| `libavcodec` | libavcodec | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | FFmpeg’s built-in codec library for decoding and encoding many media formats. |
| `libavformat` | libavformat | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | FFmpeg’s built-in container library for reading and writing formats such as MP4, MOV, MKV, WAV, and MPEG-TS. |
| `libavfilter` | libavfilter | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | FFmpeg’s built-in filtering library for video, audio, subtitles, scaling, overlays, and effects. |
| `libswscale` | libswscale | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Converts image sizes and pixel formats. |
| `libswresample` | libswresample | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Converts audio sample rates, sample formats, and channel layouts. |
| `native-codecs` | Native FFmpeg codecs | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Uses FFmpeg’s built-in codec support before adding external codec libraries. |
| `native-formats` | Native formats and muxers | included FFmpeg component | Built in / locked | included | none | included in FFmpeg source | Uses FFmpeg’s built-in readers and writers for many media containers. |

### Video encoders

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `x264` | x264 | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libx264` | `<profile>-libx264` | Supports H.264 software encoding. A leading H.264/AVC encoder with strong quality, performance, and broad compatibility. |
| `x265` | x265 | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libx265` | `<profile>-x265` | Supports HEVC/H.265 software encoding for smaller files than H.264 at similar visual quality. Useful when compression efficiency matters more than playback universality. |
| `svt-av1` | SVT-AV1 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libsvtav1` | `<profile>-svt-av1` | Produces AV1 video with practical software encoding speed. Commonly chosen for modern compression at usable throughput. |
| `libvpx` | libvpx | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvpx` | `<profile>-libvpx` | Produces VP8 and VP9 video, commonly used for WebM and web-oriented delivery. |
| `aom` | AOM AV1 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libaom` | `<profile>-aom` | Supports AV1 encoding through the reference AV1 codec family. Strong quality and standards-oriented behavior, usually chosen when output quality matters more than encoding speed. |
| `openh264` | OpenH264 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopenh264` | `<profile>-openh264` | Provides H.264 software encoding with a simpler encoder profile. Useful when basic H.264 output is needed without x264’s full decision depth. |
| `rav1e` | rav1e | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-librav1e` | `<profile>-rav1e` | Provides AV1 software encoding with a Rust-based encoder. Useful for comparing AV1 speed, quality, and encoder behavior. |
| `xvid` | Xvid | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libxvid` | `<profile>-xvidcore` | Produces older MPEG-4 Part 2 video for legacy players and devices. |
| `theora` | libtheora | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libtheora` | `<profile>-libtheora` | Produces Ogg Theora video for older open-video compatibility. |
| `kvazaar` | Kvazaar (HEVC) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libkvazaar` | `<profile>-kvazaar` | Provides HEVC/H.265 software encoding with an open research-oriented encoder. Useful when HEVC output is needed through a non-x265 path. |
| `xeve` | XEVE (EVC) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libxeve` | `<profile>-xeve` | Produces EVC video for testing and newer codec workflows. |
| `xeveb` | XEVE base profile (EVC) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libxeveb` | `<profile>-xeve` | Enables MPEG-5 EVC encoding through the EVC base profile, the royalty-free subset of EVC. Uses the same XEVE package as the main xeve row. |
| `oapv` | liboapv (APV) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-liboapv` | `<profile>-openapv` | Supports APV video encoding for specialized professional video workflows. |
| `xavs` | libxavs (AVS1) | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libxavs` | `<profile>-xavs` | Encodes video in the AVS1 (Chinese AVS) format. Useful for interoperating with AVS1 broadcast or archival material. This is GPL, so the build moves to a GPL license boundary. |
| `vvenc` | vvenc (VVC/H.266) | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-libvvenc` | prepared from pinned source/import recipe | Produces VVC/H.266 video for very high compression efficiency experiments and next-generation codec testing. |
| `xavs2` | xavs2 | internal source-prepared | Normal selectable: prepared source/import track | GPL boundary | `--enable-libxavs2` | prepared from pinned source/import recipe | Produces AVS2 video for workflows that need this regional video standard. Useful for compatibility, testing, and specialized distribution. |

### Hardware encoders

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `nvenc` | NVIDIA NVENC | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-ffnvcodec` | `<profile>-ffnvcodec-headers` | Provides NVIDIA GPU video encoding for fast H.264, HEVC, and AV1 output on supported NVIDIA systems. |
| `qsv` | Intel QSV (oneVPL) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvpl` | `<profile>-libvpl` | Provides Intel Quick Sync hardware encoding for fast video output with low CPU use. |
| `amf` | AMD AMF | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-amf` | `<profile>-amf-headers` | Provides AMD GPU video encoding for faster H.264, HEVC, and AV1 output on supported Radeon systems. Useful when speed and low CPU use matter. |
| `libmfx` | libmfx (legacy Intel Media SDK) | external SDK/import | Blocked: no normal recipe | LGPL-safe boundary | `--enable-libmfx` | no normal preparation recipe | Legacy Intel hardware encode/decode through the old Media SDK dispatcher. FFmpeg removed support for this in 7.0, so it does not work in the later source this builder targets. Use Intel QSV (oneVPL) instead. |

### Video decoders

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `dav1d` | dav1d | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libdav1d` | `<profile>-dav1d` | Provides fast AV1 video decoding. Useful for smooth playback, inspection, and conversion of AV1 files. |
| `xevd` | XEVD (EVC dec) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libxevd` | `<profile>-xevd` | Decodes EVC video streams. Useful for reading newer Essential Video Coding material. |
| `xevdb` | XEVD base profile (EVC dec) | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libxevdb` | `<profile>-xevd` | Enables MPEG-5 EVC decoding through the EVC base profile. Uses the same XEVD package as the main xevd row. |
| `davs2` | libdavs2 / AVS2 decoding | internal source-prepared | Normal selectable: prepared source/import track | GPL boundary | `--enable-libdavs2` | prepared from pinned source/import recipe | Decodes AVS2 video streams. Useful for regional, broadcast, or compatibility workflows involving AVS2 material. |
| `uavs3d` | libuavs3d / AVS3 decoding | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-libuavs3d` | prepared from pinned source/import recipe | Decodes AVS3 video streams. Useful for AVS3 test material, regional standards, and compatibility workflows. |
| `lcevc-dec` | liblcevc-dec / LCEVC decoding | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-liblcevc-dec` | prepared from pinned source/import recipe | Decodes LCEVC enhancement layers that improve a base video stream. Useful when content uses layered video enhancement. |
| `avisynthplus` | AviSynth+ / Scripted video processing | internal source-prepared | Normal selectable: prepared source/import track | GPL boundary | `--enable-avisynth` | prepared from pinned source/import recipe | Opens AviSynth script-based video processing chains directly. Useful when a workflow already depends on AviSynth filters or scripted frame processing. |

### Image codecs

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `png` | libpng | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | none | `<profile>-libpng` | Handles PNG image input and output for lossless still images and image sequences. |
| `webp` | WebP | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libwebp` | `<profile>-libwebp` | Reads and writes WebP images used widely on websites. Useful for still images, thumbnails, and image sequences. |
| `openjpeg` | OpenJPEG | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopenjpeg` | `<profile>-openjpeg2` | Reads and writes JPEG 2000 images used in archival, cinema, and professional imaging workflows. |
| `libjxl` | JPEG XL | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libjxl` | `<profile>-libjxl` | Reads and writes JPEG XL images for modern high-quality still-image compression. Useful for image sequences and advanced image workflows. |
| `rsvg` | librsvg | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-librsvg` | `<profile>-librsvg` | Renders SVG vector graphics into normal image frames. Useful for scalable logos, overlays, and graphic assets. |
| `snappy` | Snappy | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libsnappy` | `<profile>-snappy` | Provides fast Snappy data compression for formats that use it internally. |
| `lcms2` | Little CMS 2 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-lcms2` | `<profile>-lcms2` | Applies color profile management for more accurate color conversion. Useful when preserving color appearance matters. |
| `svtjpegxs` | SVT JPEG XS | native MSYS2 package | UI-disabled | LGPL-safe boundary | `--enable-libsvtjpegxs` | `<profile>-svt-jpeg-xs`, `git`, `<profile>-cmake`, `<profile>-ninja`, `<profile>-yasm` | Produces JPEG XS video for low-latency professional media workflows. |

### Filters and processing

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `zimg` | zimg | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libzimg` | `<profile>-zimg` | Improves image resizing, pixel-format conversion, and color handling. Useful for clean scaling and careful video conversion. |
| `libplacebo` | libplacebo | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libplacebo`, `--enable-vulkan` | `<profile>-libplacebo`, `<profile>-vulkan-loader`, `<profile>-vulkan-headers` | Provides high-quality GPU video processing for scaling, color conversion, tone mapping, and rendering paths. |
| `vmaf` | libvmaf | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvmaf` | `<profile>-vmaf` | Measures perceived video quality by comparing an encoded result with a reference. Useful for evaluating compression choices. |
| `vidstab` | libvidstab | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvidstab` | `<profile>-vid.stab` | Reduces camera shake and makes handheld footage look smoother. |
| `opencolorio` | OpenColorIO | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopencolorio` | `<profile>-opencolorio` | Applies studio-grade color management for film, animation, and post-production workflows. |
| `cairo` | Cairo | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-cairo` | `<profile>-cairo` | Provides 2D vector drawing support for generated graphics, overlays, and filter visuals. Useful when media processing needs drawn shapes or rendered graphic elements. |
| `opencl` | OpenCL | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-opencl` | `<profile>-opencl-icd`, `<profile>-opencl-headers` | Runs supported compute filters on OpenCL-capable GPUs or accelerators. Useful for offloading heavy image processing. |
| `shaderc` | libshaderc | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libshaderc` | `<profile>-shaderc` | Compiles shader programs used by GPU-based video filters. Useful for advanced rendering and GPU processing paths. |
| `glslang` | glslang | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libglslang` | `<profile>-glslang` | Compiles GLSL shader code for GPU-based video processing. Useful for advanced graphics and compute-style filter paths. |
| `frei0r` | frei0r | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-frei0r` | `<profile>-frei0r-plugins` | Provides extra creative video effects and filter plugins. Useful for stylized processing beyond the standard filter set. |
| `opencv` | OpenCV | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopencv` | `<profile>-opencv` | Provides computer-vision processing for image analysis and experimental visual filters. |
| `ladspa` | LADSPA | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-ladspa` | `<profile>-ladspa-sdk`, `<profile>-dlfcn` | Loads LADSPA audio effect plugins for extra audio processing. Useful when existing plugin chains need to be used in media conversion. |
| `lv2` | LV2 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-lv2` | `<profile>-lilv` | Loads LV2 audio plugins for more advanced audio effects and processing chains. |
| `qrencode` | libqrencode | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libqrencode` | `<profile>-qrencode` | Generates QR codes that can be placed into images or video frames. |
| `cuda-nvcc` | CUDA NVCC (NVIDIA GPU filters) | external SDK/import | Blocked: no normal recipe | LGPL-safe boundary | `--enable-cuda-nvcc` | no normal preparation recipe | Compiles FFmpeg's CUDA-accelerated filters with NVIDIA's nvcc compiler. nvcc ships only in the proprietary CUDA Toolkit, which is not available as an MSYS2 package. |
| `lensfun` | lensfun | native MSYS2 package | UI-disabled | LGPL-safe boundary | `--enable-liblensfun` | `<profile>-lensfun` | Corrects lens distortion, vignetting, and other camera-lens artifacts. Useful for cleanup of footage from known lenses. |
| `vapoursynth` | VapourSynth / Scripted video processing | native MSYS2 package | UI-disabled | LGPL-safe boundary | `--enable-vapoursynth` | `<profile>-vapoursynth` | Opens VapourSynth script-based video processing chains. Useful when advanced scripted filtering must feed frames into a media workflow. |

### Audio

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `opus` | Opus | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopus` | `<profile>-opus` | Produces modern Opus audio for speech, music, streaming, and low-latency communication. |
| `fdk-aac` | Fraunhofer FDK AAC | native MSYS2 package | Normal selectable: MSYS2 package track | nonfree boundary | `--enable-libfdk-aac`, `--enable-nonfree` | `<profile>-fdk-aac` | Produces high-quality AAC audio for MP4, mobile devices, streaming, and web playback. Often chosen when AAC quality is important. |
| `mp3lame` | LAME MP3 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libmp3lame` | `<profile>-lame` | Produces MP3 audio with broad playback compatibility. Useful when the output must work on almost any device or player. |
| `vorbis` | Vorbis | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvorbis` | `<profile>-libvorbis` | Produces Ogg Vorbis audio with good open-format compatibility. Useful for music and general audio in Ogg workflows. |
| `soxr` | SoX Resampler | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libsoxr` | `<profile>-libsoxr` | Provides high-quality audio sample-rate conversion. Useful when changing audio rates while preserving clarity. |
| `rubberband` | Rubber Band | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-librubberband` | `<profile>-rubberband` | Changes audio speed and pitch with higher-quality time-stretching. Useful for music, dialogue timing, and pitch adjustment. |
| `chromaprint` | Chromaprint | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-chromaprint` | `<profile>-chromaprint` | Generates compact audio fingerprints for track matching and identification. Useful for recognizing audio even when filenames or metadata are missing. |
| `twolame` | TwoLAME | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libtwolame` | `<profile>-twolame` | Produces MP2 audio for broadcast, DVD, and older media workflows. |
| `speex` | Speex | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libspeex` | `<profile>-speex` | Handles older Speex speech audio. Useful for legacy voice recordings and communication archives. |
| `opencore-amr` | OpenCORE AMR | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopencore-amrnb`, `--enable-libopencore-amrwb` | `<profile>-opencore-amr` | Reads and writes AMR speech audio used by older mobile and voice systems. |
| `vo-amrwbenc` | VisualOn AMR-WB encoder | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libvo-amrwbenc` | `<profile>-vo-amrwbenc` | Produces AMR-WB wideband speech audio for mobile-style voice compatibility. |
| `gsm` | GSM | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libgsm` | `<profile>-gsm` | Handles legacy GSM speech audio. Useful for old voice recordings, telephony archives, and compatibility cases. |
| `lc3` | LC3 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-liblc3` | `<profile>-liblc3` | Handles LC3 audio used in modern low-complexity communication audio. Useful for Bluetooth- and voice-oriented compatibility. |
| `ilbc` | iLBC | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libilbc` | `<profile>-libilbc` | Supports iLBC speech audio used in older internet voice communication. Useful for voice-call recordings and compatibility workflows. |
| `whisper` | whisper.cpp | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-whisper` | `<profile>-whisper.cpp`, `<profile>-ggml` | Converts spoken audio into text using whisper.cpp speech recognition. Useful for transcription and subtitle generation. |
| `mysofa` | libmysofa | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libmysofa` | `<profile>-libmysofa` | Provides spatial audio filter data for headphone-based 3D audio rendering. |
| `bs2b` | libbs2b | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libbs2b` | `<profile>-libbs2b` | Applies headphone crossfeed to make stereo listening feel less hard-panned and more speaker-like. Useful for headphone-focused audio processing. |
| `gme` | Game Music Emu | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libgme` | `<profile>-libgme` | Reads classic game music formats from older consoles and systems. Useful for playback or conversion of chiptune-style sources. |
| `shine` | Shine MP3 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libshine` | `<profile>-shine` | Produces MP3 audio using a simple fixed-point encoder. Useful for constrained or compatibility-focused cases. |
| `codec2` | Codec 2 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libcodec2` | `<profile>-codec2` | Supports very low-bitrate speech coding for radio, voice, and communication experiments. Useful when intelligible speech must fit into extremely small bitrates. |
| `mpeghdec` | libmpeghdec / MPEG-H audio decoding | internal source-prepared | Normal selectable: prepared source/import track | nonfree boundary | `--enable-libmpeghdec`, `--enable-nonfree` | prepared from pinned source/import recipe | Decodes MPEG-H 3D Audio for immersive and object-based audio content. |
| `pocketsphinx` | PocketSphinx (speech recognition) | internal source-prepared | Blocked: no normal recipe | LGPL-safe boundary | `--enable-pocketsphinx` | no normal preparation recipe | Adds the asr speech-recognition audio filter using CMU PocketSphinx. It cannot currently be built: FFmpeg's asr filter is incompatible with present-day PocketSphinx releases, so selecting it blocks the build. |

### Subtitles and text

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `ass` | libass | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libass` | `<profile>-libass` | Renders advanced ASS/SSA subtitles with fonts, colors, placement, outlines, and animation-like effects. Commonly used for complex subtitle presentation. |
| `freetype` | FreeType | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libfreetype` | `<profile>-freetype` | Renders font glyphs for subtitles, overlays, and text-based filters. Essential when text must appear as real shaped graphics. |
| `fontconfig` | Fontconfig | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libfontconfig` | `<profile>-fontconfig` | Finds installed fonts so subtitle and text rendering can use the right typefaces. Useful for reliable text appearance. |
| `harfbuzz` | HarfBuzz | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libharfbuzz` | `<profile>-harfbuzz` | Shapes complex text so letters combine, reorder, and position correctly. Important for subtitles in many non-Latin scripts. |
| `fribidi` | FriBidi | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libfribidi` | `<profile>-fribidi` | Handles right-to-left text direction for languages such as Arabic and Hebrew. Useful when subtitles or overlays contain bidirectional text. |
| `aribcaption` | libaribcaption | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libaribcaption` | `<profile>-libaribcaption` | Reads Japanese broadcast captions with better modern ARIB caption handling. Useful for TV recordings and broadcast transport streams. |
| `aribb24` | libaribb24 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libaribb24` | `<profile>-aribb24` | Handles ARIB B24 captions used in Japanese broadcast material. Useful for preserving or processing older broadcast subtitle data. |
| `zvbi` | libzvbi | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libzvbi` | `<profile>-zvbi` | Extracts teletext and VBI data from older broadcast sources. Useful for hidden captions, pages, and broadcast text information. |

### Disc and device input

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `bluray` | libbluray | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libbluray` | `<profile>-libbluray` | Reads Blu-ray disc structures for inspection, conversion, or extraction workflows. Useful when the source is organized as Blu-ray folders rather than a single media file. |
| `dvdread` | libdvdread | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libdvdread` | `<profile>-libdvdread` | Reads DVD media structures for extraction, inspection, or conversion. Useful when the source is a DVD folder or disc layout. |
| `dvdnav` | libdvdnav | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libdvdnav` | `<profile>-libdvdnav`, `<profile>-libdvdread` | Reads DVD-Video navigation structures such as menus, titles, chapters, and program chains. Useful for disc-style DVD workflows. |
| `openmpt` | libopenmpt | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libopenmpt` | `<profile>-libopenmpt` | Reads tracker module music with accurate playback behavior. Useful for old game, demo-scene, and tracker music formats. |
| `sdl2` | SDL2 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-sdl2` | `<profile>-SDL2` | Provides simple media output and preview support through SDL2. Useful for playback-style testing and display workflows. |
| `openal` | OpenAL | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-openal` | `<profile>-openal` | Provides additional audio device input and output paths. Useful for workflows involving live audio devices. |
| `cdio` | libcdio | native MSYS2 package | Normal selectable: MSYS2 package track | GPL boundary | `--enable-libcdio` | `<profile>-libcdio`, `<profile>-libcdio-paranoia` | Reads disc-based CD input for audio extraction and media inspection workflows. Useful when the source is an actual CD rather than copied audio files. |
| `modplug` | libmodplug | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libmodplug` | `<profile>-libmodplug` | Reads old tracker module music formats. Useful for converting or playing legacy scene and game-style music files. |
| `jack` | JACK | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libjack` | `<profile>-jack2` | Connects to JACK audio routing for studio-style capture and playback workflows. Useful in professional Linux audio environments. |
| `pulse` | PulseAudio | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libpulse` | `<profile>-pulseaudio` | Connects to PulseAudio for Linux desktop audio capture or playback workflows. |
| `caca` | libcaca | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libcaca` | `<profile>-libcaca` | Converts video into colored text-mode visuals. Mostly useful for experiments, terminal display, and unusual preview effects. |
| `opengl` | OpenGL | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-opengl` | `<profile>-mesa` | Adds the OpenGL output device for hardware-accelerated video display through OpenGL. Provided by the Mesa package. |
| `dc1394` | libdc1394 (IEEE 1394 camera) | internal source-prepared | Blocked: no normal recipe | LGPL-safe boundary | `--enable-libdc1394` | no normal preparation recipe | Captures video from IEEE 1394 (FireWire) cameras. Blocked on Windows: no MSYS2 package, and the only Windows build needs a proprietary FireWire kernel driver plus FireWire hardware, so the result would not be portable. |
| `decklink` | DeckLink (Blackmagic capture/playback) | external SDK/import | Blocked: no normal recipe | LGPL-safe boundary | `--enable-decklink` | no normal preparation recipe | Adds Blackmagic DeckLink capture and playback support. It builds against the proprietary DeckLink SDK headers, which cannot be redistributed as an MSYS2 package. |

### Network

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `srt` | SRT | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libsrt` | `<profile>-srt` | Provides SRT streaming for stable live transport over unreliable networks. |
| `rtmp` | librtmp | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-librtmp` | `<profile>-rtmpdump` | Provides RTMP streaming compatibility for older live-streaming servers and workflows. |
| `rist` | librist | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-librist` | `<profile>-librist` | Provides Reliable Internet Stream Transport for professional live streaming over unstable networks. |
| `ssh` | libssh | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libssh` | `<profile>-libssh` | Reads and writes media through SSH and SFTP-style remote access. |
| `zmq` | ZeroMQ | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libzmq` | `<profile>-zeromq` | Enables message-based runtime control for supported processing workflows. Useful for automation and interactive control. |
| `rabbitmq` | RabbitMQ-C | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-librabbitmq` | `<profile>-rabbitmq-c` | Connects media workflows to RabbitMQ message queues. Useful for automated processing systems. |
| `smbclient` | libsmbclient / SMB network file access | external SDK/import | Blocked: no normal recipe | GPL boundary | `--enable-libsmbclient` | no normal preparation recipe | Reads from and writes to SMB/CIFS network shares. Useful for media stored on Windows-style network folders. It cannot be built on Windows yet, so it is listed last in this section and stays unavailable until a Windows version of libsmbclient exists. |

### Secure network (TLS)

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `openssl` | OpenSSL | native MSYS2 package | Normal selectable: MSYS2 package track | nonfree boundary | `--enable-openssl` | `<profile>-openssl` | Provides encrypted network connections for HTTPS and other secure media protocols. |
| `gnutls` | GnuTLS | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-gnutls` | `<profile>-gnutls` | Provides encrypted network connections for HTTPS and other TLS-based media access. Useful for secure streaming and remote sources. |
| `mbedtls` | mbedTLS / Secure network access | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-mbedtls` | `<profile>-mbedtls` | Provides lightweight TLS support for encrypted network media access. Useful when a smaller security backend is preferred. |
| `libtls` | libtls / Secure network access | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-libtls` | prepared from pinned source/import recipe | Provides TLS-encrypted network communication through a compact TLS interface. Useful for secure network media access. |

### OCR

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `tesseract` | Tesseract OCR | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libtesseract` | `<profile>-tesseract-ocr` | Reads visible text from images or video frames. Useful for extracting burned-in titles, signs, captions, or document text. |

### AI support

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `onnxruntime` | ONNX Runtime / AI model inference | native MSYS2 package | UI-disabled; profile-limited: not in mingw64 | LGPL-safe boundary | `--enable-libonnxruntime` | `<profile>-onnxruntime` | Runs supported deep-learning filters through ONNX Runtime. Useful for model-based analysis or enhancement workflows. |
| `openvino` | OpenVINO / AI model inference | external SDK/import | Blocked: no normal recipe | LGPL-safe boundary | `--enable-libopenvino` | no normal preparation recipe | Runs supported AI inference filters with Intel-oriented acceleration. Useful for model-based video or image processing. |
| `torch` | Torch / libtorch | external SDK/import | Blocked: no normal recipe | LGPL-safe boundary | `--enable-libtorch` | no normal preparation recipe | Runs supported deep-learning filters through Torch-based model execution. Useful for PyTorch-style inference workflows. |
| `tensorflow` | TensorFlow / AI model inference | external SDK/import | UI-disabled | LGPL-safe boundary | `--enable-libtensorflow` | prepared from pinned source/import recipe | Runs supported deep-learning filters through the TensorFlow C API. Useful for model-based image or video analysis. |

### Support libraries

| ID | Display name | Track | Normal state | License effect | Configure flags | Packages / preparation | Purpose |
|---|---|---|---|---|---|---|---|
| `xml2` | libxml2 | native MSYS2 package | Normal selectable: MSYS2 package track | LGPL-safe boundary | `--enable-libxml2` | `<profile>-libxml2` | Reads structured XML data used by some media formats, subtitles, manifests, and metadata workflows. |
| `quirc` | libquirc / QR code decoding | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-libquirc` | prepared from pinned source/import recipe | Decodes QR codes from video frames or images. Useful for automation, scanning, and visual metadata workflows. |
| `klvanc` | libklvanc / Broadcast metadata | internal source-prepared | Normal selectable: prepared source/import track | LGPL-safe boundary | `--enable-libklvanc` | prepared from pinned source/import recipe | Processes vertical ancillary data used in broadcast video. Useful for metadata carried alongside video lines. |

## License-boundary summary

The program derives the final license profile from selected libraries and final configure flags instead of asking the user to manually choose `--enable-gpl`, `--enable-nonfree`, or `--enable-version3`. GPL rows move the plan to the GPL-local boundary. Nonfree rows move it to the nonfree-local boundary and trigger redistribution warnings. Version-3-sensitive rows add `--enable-version3` automatically where required.

| License effect | Count | Representative rows |
|---|---:|---|
| included | 10 | `ffmpeg-program`, `ffprobe-program`, `libavutil`, `libavcodec`, `libavformat`, `libavfilter`, `libswscale`, `libswresample`, `native-codecs`, `native-formats` |
| LGPL-safe boundary | 98 | `svt-av1`, `libvpx`, `aom`, `openh264`, `rav1e`, `theora`, `kvazaar`, `xeve`, `xeveb`, `oapv`, `vvenc`, `nvenc` |
| GPL boundary | 14 | `x264`, `x265`, `xvid`, `xavs`, `xavs2`, `davs2`, `avisynthplus`, `frei0r`, `rubberband`, `zvbi`, `dvdread`, `dvdnav` |
| nonfree boundary | 3 | `fdk-aac`, `mpeghdec`, `openssl` |

## Practical interpretation

For ordinary Windows users, the real supported surface is the normal selectable set: locked FFmpeg rows, native MSYS2-package rows, and implemented internal source-prepared rows. The catalog deliberately avoids pretending that every upstream FFmpeg library can be turned on just because a configure flag exists. Rows such as legacy `libmfx`, `cuda-nvcc`, DeckLink, OpenVINO, Torch, TensorFlow, VapourSynth, SVT JPEG XS and ONNX Runtime are treated conservatively because the current program either lacks a safe preparation path, lacks stable compatibility with the targeted FFmpeg source, or would require external SDK/runtime assumptions that violate the program’s local, reviewable build model.
