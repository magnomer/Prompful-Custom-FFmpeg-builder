package planning

func LibraryCatalogForShellProfile(windowsShellProfileName string) []LibraryChoice {
	packagePrefix := packagePrefixForShellProfile(windowsShellProfileName)
	libraries := []LibraryChoice{
		// Included by default (official FFmpeg source)
		includedLibraryChoice("ffmpeg-program", "ffmpeg.exe", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("ffprobe-program", "ffprobe.exe", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libavutil", "libavutil", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libavcodec", "libavcodec", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libavformat", "libavformat", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libavfilter", "libavfilter", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libswscale", "libswscale", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("libswresample", "libswresample", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("native-codecs", "Native FFmpeg codecs", "Included by default (official FFmpeg source)"),
		includedLibraryChoice("native-formats", "Native formats and muxers", "Included by default (official FFmpeg source)"),

		// Video encoders — common, stable encoders first; niche or fragile codec families later.
		libraryChoice("x264", "x264", "Video encoders", []string{"--enable-libx264"}, []string{packagePrefix + "-libx264"}, "gpl"),
		libraryChoice("x265", "x265", "Video encoders", []string{"--enable-libx265"}, []string{packagePrefix + "-x265"}, "gpl"),
		libraryChoice("svt-av1", "SVT-AV1", "Video encoders", []string{"--enable-libsvtav1"}, []string{packagePrefix + "-svt-av1"}, "lgpl"),
		libraryChoice("libvpx", "libvpx", "Video encoders", []string{"--enable-libvpx"}, []string{packagePrefix + "-libvpx"}, "lgpl"),
		libraryChoice("aom", "AOM AV1", "Video encoders", []string{"--enable-libaom"}, []string{packagePrefix + "-aom"}, "lgpl"),
		libraryChoice("openh264", "OpenH264", "Video encoders", []string{"--enable-libopenh264"}, []string{packagePrefix + "-openh264"}, "lgpl"),
		libraryChoice("rav1e", "rav1e", "Video encoders", []string{"--enable-librav1e"}, []string{packagePrefix + "-rav1e"}, "lgpl"),
		libraryChoice("xvid", "Xvid", "Video encoders", []string{"--enable-libxvid"}, []string{packagePrefix + "-xvidcore"}, "gpl"),
		libraryChoice("theora", "libtheora", "Video encoders", []string{"--enable-libtheora"}, []string{packagePrefix + "-libtheora"}, "lgpl"),
		libraryChoice("kvazaar", "Kvazaar (HEVC)", "Video encoders", []string{"--enable-libkvazaar"}, []string{packagePrefix + "-kvazaar"}, "lgpl"),
		libraryChoice("xeve", "XEVE (EVC)", "Video encoders", []string{"--enable-libxeve"}, []string{packagePrefix + "-xeve"}, "lgpl"),
		// xeveb is the EVC base-profile encode flag; it reuses the same MSYS2 xeve package
		// as the main-profile xeve row but enables FFmpeg's separate base-profile binding.
		libraryChoice("xeveb", "XEVE base profile (EVC)", "Video encoders", []string{"--enable-libxeveb"}, []string{packagePrefix + "-xeve"}, "lgpl"),
		libraryChoice("oapv", "liboapv (APV)", "Video encoders", []string{"--enable-liboapv"}, []string{packagePrefix + "-openapv"}, "lgpl"),
		// xavs is the AVS1 (Chinese AVS) encoder; distinct from xavs2 (AVS2). GPL, with a
		// prebuilt MSYS2 xavs package, so it is a normal Native-track row.
		libraryChoice("xavs", "libxavs (AVS1)", "Video encoders", []string{"--enable-libxavs"}, []string{packagePrefix + "-xavs"}, "gpl"),
		trackedLibraryChoice(LibraryTrackInternal, "vvenc", "vvenc (VVC/H.266)", "Video encoders", []string{"--enable-libvvenc"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "xavs2", "xavs2", "Video encoders", []string{"--enable-libxavs2"}, []string{}, "gpl"),

		// Hardware encoders — ordered by common Windows adoption and stability.
		libraryChoice("nvenc", "NVIDIA NVENC", "Hardware encoders", []string{"--enable-ffnvcodec"}, []string{packagePrefix + "-ffnvcodec-headers"}, "lgpl"),
		// Intel Hardware Acceleration (Quick Sync) has two backends, MUTUALLY EXCLUSIVE: oneVPL
		// (libvpl, --enable-libvpl) and the legacy Media SDK (libmfx, --enable-libmfx). FFmpeg
		// configure dies with "can not use libmfx and libvpl together" when both are enabled. They
		// are kept adjacent here so the UI renders them as one pick-one radio block (see
		// intelHwaccelBackendLibraryIds in the frontend), matching the EVC xeve/xeveb pair. The
		// planner blocks the both-enabled combination and the selection logic keeps oneVPL (the
		// maintained path) over the deprecated libmfx.
		//
		// libvpl/oneVPL is the modern backend. Per FFmpeg's Changelog it was added in 6.0
		// ("oneVPL support for QSV"); the release-support manifest therefore lists libvpl only on
		// the 6.x+ lines, so on 4.4/5.x it is version-unsupported and pruned, leaving libmfx as the
		// only Intel HW accel path there. minVersion is the libvpl PACKAGE pkg-config floor, not an
		// FFmpeg version.
		libraryChoice("libvpl", "Intel HW accel (oneVPL)", "Hardware encoders", []string{"--enable-libvpl"}, []string{packagePrefix + "-libvpl"}, "lgpl"),
		// libmfx is the legacy Intel Media SDK dispatcher. FFmpeg's --enable-libmfx switch exists in
		// every supported release (4.4 through 8.1 — per the Changelog it has not been removed), so
		// it is the only Intel HW accel path before 6.0 and a deprecated-but-valid alternative after.
		// MSYS2 ships no libmfx/mfx_dispatch package, so it is built from source on the Internal
		// track from Intel's open dispatcher (lu-zero/mfx_dispatch), the same as the other
		// source-built libraries; the version pin and build/verify live in library-sources.json +
		// libraryItemSpecs.
		trackedLibraryChoice(LibraryTrackInternal, "libmfx", "libmfx (legacy Intel Media SDK)", "Hardware encoders", []string{"--enable-libmfx"}, []string{}, "lgpl"),
		libraryChoice("amf", "AMD AMF", "Hardware encoders", []string{"--enable-amf"}, []string{packagePrefix + "-amf-headers"}, "lgpl"),

		// Video decoders — broadly used and stable decoders first.
		libraryChoice("dav1d", "dav1d", "Video decoders", []string{"--enable-libdav1d"}, []string{packagePrefix + "-dav1d"}, "lgpl"),
		libraryChoice("xevd", "XEVD (EVC dec)", "Video decoders", []string{"--enable-libxevd"}, []string{packagePrefix + "-xevd"}, "lgpl"),
		// xevdb is the EVC base-profile decode flag; it reuses the same MSYS2 xevd package as
		// the main-profile xevd row but enables FFmpeg's separate base-profile binding.
		libraryChoice("xevdb", "XEVD base profile (EVC dec)", "Video decoders", []string{"--enable-libxevdb"}, []string{packagePrefix + "-xevd"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "davs2", "libdavs2 / AVS2 decoding", "Video decoders", []string{"--enable-libdavs2"}, []string{}, "gpl"),
		trackedLibraryChoice(LibraryTrackInternal, "uavs3d", "libuavs3d / AVS3 decoding", "Video decoders", []string{"--enable-libuavs3d"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "lcevc-dec", "liblcevc-dec / LCEVC decoding", "Video decoders", []string{"--enable-liblcevc-dec"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "avisynthplus", "AviSynth+ / Scripted video processing", "Video decoders", []string{"--enable-avisynth"}, []string{}, "gpl"),

		// Image codecs — everyday image formats first; specialized formats later.
		libraryChoice("png", "libpng", "Image codecs", []string{}, []string{packagePrefix + "-libpng"}, "lgpl"),
		libraryChoice("webp", "WebP", "Image codecs", []string{"--enable-libwebp"}, []string{packagePrefix + "-libwebp"}, "lgpl"),
		libraryChoice("openjpeg", "OpenJPEG", "Image codecs", []string{"--enable-libopenjpeg"}, []string{packagePrefix + "-openjpeg2"}, "lgpl"),
		libraryChoice("libjxl", "JPEG XL", "Image codecs", []string{"--enable-libjxl"}, []string{packagePrefix + "-libjxl"}, "lgpl"),
		libraryChoice("rsvg", "librsvg", "Image codecs", []string{"--enable-librsvg"}, []string{packagePrefix + "-librsvg"}, "lgpl"),
		libraryChoice("snappy", "Snappy", "Image codecs", []string{"--enable-libsnappy"}, []string{packagePrefix + "-snappy"}, "lgpl"),
		libraryChoice("lcms2", "Little CMS 2", "Image codecs", []string{"--enable-lcms2"}, []string{packagePrefix + "-lcms2"}, "lgpl"),
		libraryChoice("svtjpegxs", "SVT JPEG XS", "Image codecs", []string{"--enable-libsvtjpegxs"}, []string{packagePrefix + "-svt-jpeg-xs", "git", packagePrefix + "-cmake", packagePrefix + "-ninja", packagePrefix + "-yasm"}, "lgpl"),

		// Filters and processing — widely used stable processing libraries first; fragile processing paths later.
		libraryChoice("zimg", "zimg", "Filters and processing", []string{"--enable-libzimg"}, []string{packagePrefix + "-zimg"}, "lgpl"),
		libraryChoice("libplacebo", "libplacebo", "Filters and processing", []string{"--enable-libplacebo", "--enable-vulkan"}, []string{packagePrefix + "-libplacebo", packagePrefix + "-vulkan-loader", packagePrefix + "-vulkan-headers"}, "lgpl"),
		libraryChoice("vmaf", "libvmaf", "Filters and processing", []string{"--enable-libvmaf"}, []string{packagePrefix + "-vmaf"}, "lgpl"),
		libraryChoice("vidstab", "libvidstab", "Filters and processing", []string{"--enable-libvidstab"}, []string{packagePrefix + "-vid.stab"}, "lgpl"),
		libraryChoice("opencolorio", "OpenColorIO", "Filters and processing", []string{"--enable-libopencolorio"}, []string{packagePrefix + "-opencolorio"}, "lgpl"),
		libraryChoice("cairo", "Cairo", "Filters and processing", []string{"--enable-cairo"}, []string{packagePrefix + "-cairo"}, "lgpl"),
		libraryChoice("opencl", "OpenCL", "Filters and processing", []string{"--enable-opencl"}, []string{packagePrefix + "-opencl-icd", packagePrefix + "-opencl-headers"}, "lgpl"),
		libraryChoice("shaderc", "libshaderc", "Filters and processing", []string{"--enable-libshaderc"}, []string{packagePrefix + "-shaderc"}, "lgpl"),
		libraryChoice("glslang", "glslang", "Filters and processing", []string{"--enable-libglslang"}, []string{packagePrefix + "-glslang"}, "lgpl"),
		libraryChoice("frei0r", "frei0r", "Filters and processing", []string{"--enable-frei0r"}, []string{packagePrefix + "-frei0r-plugins"}, "gpl"),
		libraryChoice("opencv", "OpenCV", "Filters and processing", []string{"--enable-libopencv"}, []string{packagePrefix + "-opencv"}, "lgpl"),
		libraryChoice("ladspa", "LADSPA", "Filters and processing", []string{"--enable-ladspa"}, []string{packagePrefix + "-ladspa-sdk", packagePrefix + "-dlfcn"}, "lgpl"),
		libraryChoice("lv2", "LV2", "Filters and processing", []string{"--enable-lv2"}, []string{packagePrefix + "-lilv"}, "lgpl"),
		libraryChoice("qrencode", "libqrencode", "Filters and processing", []string{"--enable-libqrencode"}, []string{packagePrefix + "-qrencode"}, "lgpl"),
		// cuda-nvcc enables compiling FFmpeg's CUDA filter kernels with NVIDIA's nvcc. nvcc
		// ships only in the proprietary CUDA Toolkit, which is not an MSYS2 package, so this is
		// a visible External-track row with no preparation recipe: selecting it blocks the build
		// until a CUDA Toolkit import path exists, instead of failing the configure nvcc probe.
		// Placed with the other unavailable rows at the bottom of the section.
		trackedLibraryChoice(LibraryTrackExternal, "cuda-nvcc", "CUDA NVCC (NVIDIA GPU filters)", "Filters and processing", []string{"--enable-cuda-nvcc"}, []string{}, "lgpl"),
		libraryChoice("lensfun", "lensfun", "Filters and processing", []string{"--enable-liblensfun"}, []string{packagePrefix + "-lensfun"}, "lgpl"),
		libraryChoice("vapoursynth", "VapourSynth / Scripted video processing", "Filters and processing", []string{"--enable-vapoursynth"}, []string{packagePrefix + "-vapoursynth"}, "lgpl"),

		// Audio — common modern codecs and stable audio tools first; legacy and niche speech formats later.
		libraryChoice("opus", "Opus", "Audio", []string{"--enable-libopus"}, []string{packagePrefix + "-opus"}, "lgpl"),
		libraryChoice("fdk-aac", "Fraunhofer FDK AAC", "Audio", []string{"--enable-libfdk-aac", "--enable-nonfree"}, []string{packagePrefix + "-fdk-aac"}, "nonfree"),
		libraryChoice("mp3lame", "LAME MP3", "Audio", []string{"--enable-libmp3lame"}, []string{packagePrefix + "-lame"}, "lgpl"),
		libraryChoice("vorbis", "Vorbis", "Audio", []string{"--enable-libvorbis"}, []string{packagePrefix + "-libvorbis"}, "lgpl"),
		libraryChoice("soxr", "SoX Resampler", "Audio", []string{"--enable-libsoxr"}, []string{packagePrefix + "-libsoxr"}, "lgpl"),
		libraryChoice("rubberband", "Rubber Band", "Audio", []string{"--enable-librubberband"}, []string{packagePrefix + "-rubberband"}, "gpl"),
		libraryChoice("chromaprint", "Chromaprint", "Audio", []string{"--enable-chromaprint"}, []string{packagePrefix + "-chromaprint"}, "lgpl"),
		libraryChoice("twolame", "TwoLAME", "Audio", []string{"--enable-libtwolame"}, []string{packagePrefix + "-twolame"}, "lgpl"),
		libraryChoice("speex", "Speex", "Audio", []string{"--enable-libspeex"}, []string{packagePrefix + "-speex"}, "lgpl"),
		libraryChoice("opencore-amr", "OpenCORE AMR", "Audio", []string{"--enable-libopencore-amrnb", "--enable-libopencore-amrwb"}, []string{packagePrefix + "-opencore-amr"}, "lgpl"),
		libraryChoice("vo-amrwbenc", "VisualOn AMR-WB encoder", "Audio", []string{"--enable-libvo-amrwbenc"}, []string{packagePrefix + "-vo-amrwbenc"}, "lgpl"),
		libraryChoice("gsm", "GSM", "Audio", []string{"--enable-libgsm"}, []string{packagePrefix + "-gsm"}, "lgpl"),
		libraryChoice("lc3", "LC3", "Audio", []string{"--enable-liblc3"}, []string{packagePrefix + "-liblc3"}, "lgpl"),
		libraryChoice("ilbc", "iLBC", "Audio", []string{"--enable-libilbc"}, []string{packagePrefix + "-libilbc"}, "lgpl"),
		libraryChoice("whisper", "whisper.cpp", "Audio", []string{"--enable-whisper"}, []string{packagePrefix + "-whisper.cpp", packagePrefix + "-ggml"}, "lgpl"),
		libraryChoice("mysofa", "libmysofa", "Audio", []string{"--enable-libmysofa"}, []string{packagePrefix + "-libmysofa"}, "lgpl"),
		libraryChoice("bs2b", "libbs2b", "Audio", []string{"--enable-libbs2b"}, []string{packagePrefix + "-libbs2b"}, "lgpl"),
		libraryChoice("gme", "Game Music Emu", "Audio", []string{"--enable-libgme"}, []string{packagePrefix + "-libgme"}, "lgpl"),
		libraryChoice("shine", "Shine MP3", "Audio", []string{"--enable-libshine"}, []string{packagePrefix + "-shine"}, "lgpl"),
		libraryChoice("codec2", "Codec 2", "Audio", []string{"--enable-libcodec2"}, []string{packagePrefix + "-codec2"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "mpeghdec", "libmpeghdec / MPEG-H audio decoding", "Audio", []string{"--enable-libmpeghdec", "--enable-nonfree"}, []string{}, "nonfree"),
		// PocketSphinx provides the asr speech-recognition filter, but it is blocked by an
		// upstream FFmpeg incompatibility, not merely a missing package. FFmpeg's af_asr.c
		// still uses the pre-5.0 PocketSphinx API (cmd_ln_t / cmd_ln_parse_r / ps_args, plus
		// the old sphinxbase split). PocketSphinx 5.0 removed that API and absorbed sphinxbase,
		// so against any current release (5.1.x) configure's lenient ps_init pkg-config probe
		// passes but `make` fails compiling the filter (FFmpeg trac #10520, open since 2023).
		// The only combination that compiles is the legacy 5prealpha source plus a separately
		// built sphinxbase, which is not worth a two-archive prerequisite recipe. So it stays a
		// visible Internal-track row with no preparation recipe and the planner keeps blocking
		// it. Placed last in the section with the other unavailable rows.
		trackedLibraryChoice(LibraryTrackInternal, "pocketsphinx", "PocketSphinx (speech recognition)", "Audio", []string{"--enable-pocketsphinx"}, []string{}, "lgpl"),

		// Subtitles and text — common rendering stack first; regional broadcast text later.
		libraryChoice("ass", "libass", "Subtitles and text", []string{"--enable-libass"}, []string{packagePrefix + "-libass"}, "lgpl"),
		libraryChoice("freetype", "FreeType", "Subtitles and text", []string{"--enable-libfreetype"}, []string{packagePrefix + "-freetype"}, "lgpl"),
		libraryChoice("fontconfig", "Fontconfig", "Subtitles and text", []string{"--enable-libfontconfig"}, []string{packagePrefix + "-fontconfig"}, "lgpl"),
		libraryChoice("harfbuzz", "HarfBuzz", "Subtitles and text", []string{"--enable-libharfbuzz"}, []string{packagePrefix + "-harfbuzz"}, "lgpl"),
		libraryChoice("fribidi", "FriBidi", "Subtitles and text", []string{"--enable-libfribidi"}, []string{packagePrefix + "-fribidi"}, "lgpl"),
		libraryChoice("aribcaption", "libaribcaption", "Subtitles and text", []string{"--enable-libaribcaption"}, []string{packagePrefix + "-libaribcaption"}, "lgpl"),
		libraryChoice("aribb24", "libaribb24", "Subtitles and text", []string{"--enable-libaribb24"}, []string{packagePrefix + "-aribb24"}, "lgpl"),
		libraryChoice("zvbi", "libzvbi", "Subtitles and text", []string{"--enable-libzvbi"}, []string{packagePrefix + "-zvbi"}, "gpl"),

		// Disc and device input — common disc/media access first; platform-specific devices later.
		libraryChoice("bluray", "libbluray", "Disc and device input", []string{"--enable-libbluray"}, []string{packagePrefix + "-libbluray"}, "lgpl"),
		libraryChoice("dvdread", "libdvdread", "Disc and device input", []string{"--enable-libdvdread"}, []string{packagePrefix + "-libdvdread"}, "gpl"),
		libraryChoice("dvdnav", "libdvdnav", "Disc and device input", []string{"--enable-libdvdnav"}, []string{packagePrefix + "-libdvdnav", packagePrefix + "-libdvdread"}, "gpl"),
		libraryChoice("openmpt", "libopenmpt", "Disc and device input", []string{"--enable-libopenmpt"}, []string{packagePrefix + "-libopenmpt"}, "lgpl"),
		libraryChoice("sdl2", "SDL2", "Disc and device input", []string{"--enable-sdl2"}, []string{packagePrefix + "-SDL2"}, "lgpl"),
		libraryChoice("openal", "OpenAL", "Disc and device input", []string{"--enable-openal"}, []string{packagePrefix + "-openal"}, "lgpl"),
		libraryChoice("cdio", "libcdio", "Disc and device input", []string{"--enable-libcdio"}, []string{packagePrefix + "-libcdio", packagePrefix + "-libcdio-paranoia"}, "gpl"),
		libraryChoice("modplug", "libmodplug", "Disc and device input", []string{"--enable-libmodplug"}, []string{packagePrefix + "-libmodplug"}, "lgpl"),
		libraryChoice("jack", "JACK", "Disc and device input", []string{"--enable-libjack"}, []string{packagePrefix + "-jack2"}, "lgpl"),
		libraryChoice("pulse", "PulseAudio", "Disc and device input", []string{"--enable-libpulse"}, []string{packagePrefix + "-pulseaudio"}, "lgpl"),
		libraryChoice("caca", "libcaca", "Disc and device input", []string{"--enable-libcaca"}, []string{packagePrefix + "-libcaca"}, "lgpl"),
		// opengl enables the OpenGL output/render device. The mesa MSYS2 package supplies the GL
		// headers; on Windows the device links against system opengl32/gdi32. Native-track row.
		libraryChoice("opengl", "OpenGL", "Disc and device input", []string{"--enable-opengl"}, []string{packagePrefix + "-mesa"}, "lgpl"),
		// libdc1394 is IEEE 1394 (FireWire) camera capture. There is no MSYS2 package, and
		// mainline libdc1394 has no Windows transport backend (only Linux raw1394/juju and
		// macOS). The one Windows port (indigo-astronomy fork) needs the proprietary CMU
		// 1394Camera SDK at build time plus a CMU kernel driver and FireWire hardware at
		// runtime, so it cannot produce a portable static binary. It stays a visible Internal-
		// track row with no recipe: selecting it blocks the build.
		trackedLibraryChoice(LibraryTrackInternal, "dc1394", "libdc1394 (IEEE 1394 camera)", "Disc and device input", []string{"--enable-libdc1394"}, []string{}, "lgpl"),
		// DeckLink is Blackmagic capture/playback. It builds against the proprietary DeckLink
		// SDK headers, which are not redistributable as an MSYS2 package, so it is a visible
		// External-track row with no import recipe: selecting it blocks the build until the SDK
		// headers can be imported.
		trackedLibraryChoice(LibraryTrackExternal, "decklink", "DeckLink (Blackmagic capture/playback)", "Disc and device input", []string{"--enable-decklink"}, []string{}, "lgpl"),

		// Network — streaming/media protocols first; automation messaging backends later.
		// (TLS backends live in their own "Secure network (TLS)" section at the top.)
		libraryChoice("srt", "SRT", "Network", []string{"--enable-libsrt"}, []string{packagePrefix + "-srt"}, "lgpl"),
		libraryChoice("rtmp", "librtmp", "Network", []string{"--enable-librtmp"}, []string{packagePrefix + "-rtmpdump"}, "lgpl"),
		libraryChoice("rist", "librist", "Network", []string{"--enable-librist"}, []string{packagePrefix + "-librist"}, "lgpl"),
		libraryChoice("ssh", "libssh", "Network", []string{"--enable-libssh"}, []string{packagePrefix + "-libssh"}, "lgpl"),
		libraryChoice("zmq", "ZeroMQ", "Network", []string{"--enable-libzmq"}, []string{packagePrefix + "-zeromq"}, "lgpl"),
		libraryChoice("rabbitmq", "RabbitMQ-C", "Network", []string{"--enable-librabbitmq"}, []string{packagePrefix + "-rabbitmq-c"}, "lgpl"),
		// smbclient sits last in the Network section because it has no working Windows
		// build path yet: Samba publishes no MSYS2 package and no prebuilt mingw-w64
		// libsmbclient, and Samba does not build natively on Windows (it needs a POSIX
		// environment). It is kept as an External-track entry so a future Windows
		// libsmbclient can be wired in without UI changes; until then it has no
		// preparation recipe, so selecting it leaves the build blocked by the planner.
		trackedLibraryChoice(LibraryTrackExternal, "smbclient", "libsmbclient / SMB network file access", "Network", []string{"--enable-libsmbclient"}, []string{}, "gpl"),

		// Secure network (TLS) — exactly one TLS backend is used; the UI treats these as a
		// pick-one group (priority openssl > gnutls > mbedtls > libtls). Placed right after
		// Network so the license-sensitive choice (OpenSSL is nonfree) stands on its own.
		libraryChoice("openssl", "OpenSSL", "Secure network (TLS)", []string{"--enable-openssl"}, []string{packagePrefix + "-openssl"}, "nonfree"),
		libraryChoice("gnutls", "GnuTLS", "Secure network (TLS)", []string{"--enable-gnutls"}, []string{packagePrefix + "-gnutls"}, "lgpl"),
		libraryChoice("mbedtls", "mbedTLS / Secure network access", "Secure network (TLS)", []string{"--enable-mbedtls"}, []string{packagePrefix + "-mbedtls"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "libtls", "libtls / Secure network access", "Secure network (TLS)", []string{"--enable-libtls"}, []string{}, "lgpl"),

		// OCR
		libraryChoice("tesseract", "Tesseract OCR", "OCR", []string{"--enable-libtesseract"}, []string{packagePrefix + "-tesseract-ocr"}, "lgpl"),

		// AI support — general inference backends first; unavailable or heavier backends later.
		libraryChoice("onnxruntime", "ONNX Runtime / AI model inference", "AI support", []string{"--enable-libonnxruntime"}, []string{packagePrefix + "-onnxruntime"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackExternal, "openvino", "OpenVINO / AI model inference", "AI support", []string{"--enable-libopenvino"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackExternal, "torch", "Torch / libtorch", "AI support", []string{"--enable-libtorch"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackExternal, "tensorflow", "TensorFlow / AI model inference", "AI support", []string{"--enable-libtensorflow"}, []string{}, "lgpl"),
		// Support libraries — stable general-purpose support first; specialized broadcast/QR helpers later.
		libraryChoice("xml2", "libxml2", "Support libraries", []string{"--enable-libxml2"}, []string{packagePrefix + "-libxml2"}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "quirc", "libquirc / QR code decoding", "Support libraries", []string{"--enable-libquirc"}, []string{}, "lgpl"),
		trackedLibraryChoice(LibraryTrackInternal, "klvanc", "libklvanc / Broadcast metadata", "Support libraries", []string{"--enable-libklvanc"}, []string{}, "lgpl"),
	}
	// Some packages exist only for certain shell profiles. ONNX Runtime has no
	// prebuilt mingw64 package, so it is removed from the catalog for that profile;
	// the UI then never offers it and the planner drops it from any saved selection.
	return filterLibrariesUnavailableForProfile(libraries, windowsShellProfileName)
}

// libraryProfileUnavailability lists, per library id, the shell profiles that have
// no prebuilt MSYS2 package for it. Such a library is omitted from the catalog for
// those profiles. Keep in sync with the frontend libraryUnavailableProfiles map.
var libraryProfileUnavailability = map[string][]string{
	"onnxruntime": {"mingw64"},
}

func libraryAvailableForProfile(libraryId string, windowsShellProfileName string) bool {
	for _, unavailableProfileName := range libraryProfileUnavailability[libraryId] {
		if unavailableProfileName == windowsShellProfileName {
			return false
		}
	}
	return true
}

func filterLibrariesUnavailableForProfile(libraries []LibraryChoice, windowsShellProfileName string) []LibraryChoice {
	filtered := make([]LibraryChoice, 0, len(libraries))
	for _, library := range libraries {
		if libraryAvailableForProfile(library.LibraryId, windowsShellProfileName) {
			filtered = append(filtered, library)
		}
	}
	return filtered
}

func ConfigureOptionCatalog() []ConfigureOptionChoice {
	return []ConfigureOptionChoice{
		configureOptionChoice("default-static", "Build static libraries", "Default FFmpeg source build", []string{}, "FFmpeg normally builds static libraries from source.", "Checked because this is normal FFmpeg configure behavior. No extra flag is needed unless you choose a different output type.", true, true),
		configureOptionChoice("default-programs", "Build command-line programs", "Default FFmpeg source build", []string{}, "FFmpeg normally builds command-line programs such as ffmpeg.exe and ffprobe.exe.", "Checked because programs are built in a normal source build. Disable only if you want libraries without command-line tools.", true, true),
		configureOptionChoice("default-ffmpeg", "Build ffmpeg.exe", "Default FFmpeg source build", []string{}, "Builds the main command-line converter most users run.", "Checked because ffmpeg.exe is part of a normal program build.", true, true),
		configureOptionChoice("default-ffprobe", "Build ffprobe.exe", "Default FFmpeg source build", []string{}, "Builds the media inspection tool used to read stream and container information.", "Checked because ffprobe.exe is part of a normal program build.", true, true),

		configureOptionChoice("enable-shared", "Build shared DLL libraries", "Output type", []string{"--enable-shared", "--disable-static"}, "Creates DLL-style FFmpeg libraries for other programs to load.", "FFmpeg platform documentation describes --enable-shared as the way to build FFmpeg libraries as DLLs on Windows. This changes the output type, so it is not selected by default here.", false, false),
		configureOptionChoice("disable-ffplay", "Do not build ffplay", "Programs", []string{"--disable-ffplay"}, "Skips the simple playback test program.", "Useful when SDL playback support is unnecessary. ffmpeg.exe and ffprobe.exe are unaffected.", false, false),
		configureOptionChoice("disable-autodetect", "Do not auto-use hidden system libraries", "Security and reproducibility", []string{"--disable-autodetect"}, "Makes the build less surprising by using only explicitly selected external libraries.", "Good for transparent/reproducible builds. Select this when you want the Review page to explain every external dependency.", false, false),
		configureOptionChoice("disable-network", "Remove all networking support", "Security and reproducibility", []string{"--disable-network"}, "Builds FFmpeg without any network protocols for a smaller, offline-only tool.", "Select for local-only conversion or hardened builds. Disables HTTP, HTTPS, RTMP, SRT, and every other network input/output, so streaming protocols stop working even if their libraries are selected.", false, false),
		configureOptionChoice("disable-asm", "Disable assembly optimizations", "Compatibility", []string{"--disable-asm"}, "Uses slower but simpler C code paths if assembly causes build problems.", "Normally leave unchecked because FFmpeg is faster with assembly optimizations.", false, false),
		configureOptionChoice("disable-x86asm", "Disable x86 assembly", "Compatibility", []string{"--disable-x86asm"}, "Try this when NASM/YASM-related build problems occur.", "Normally leave unchecked for performance.", false, false),
		configureOptionChoice("pkg-config-static", "Link external libraries statically", "Compatibility", []string{"--pkg-config-flags=--static"}, "Tells pkg-config to pull the full static dependency chain when linking external libraries.", "Often required for static Windows builds that use external libraries, so configure can find every transitive dependency. Has no effect on a shared/DLL build.", false, false),
		configureOptionChoice("enable-runtime-cpudetect", "Detect CPU features at run time", "Compatibility", []string{"--enable-runtime-cpudetect"}, "Builds one binary that picks CPU optimizations while running, so it works across different processors.", "Useful when sharing the build with other machines. Slightly slower than a build tuned for one specific CPU.", false, false),
		configureOptionChoice("disable-doc", "Skip documentation files", "Size and speed", []string{"--disable-doc"}, "Makes the build smaller by not building local documentation files.", "Not a normal source default. Select this when you only need binaries/libraries and do not need generated docs.", false, false),
		configureOptionChoice("enable-small", "Prefer smaller binary size", "Size and speed", []string{"--enable-small"}, "Asks FFmpeg to prefer smaller output files over speed.", "Useful for constrained environments; may reduce performance.", false, false),
		configureOptionChoice("enable-lto", "Enable link-time optimization (LTO)", "Size and speed", []string{"--enable-lto"}, "Lets the compiler optimize across files for a smaller and sometimes faster binary.", "Increases build time and memory use. Leave unchecked if linking fails or runs out of memory.", false, false),
		configureOptionChoice("disable-debug", "Remove debug build data", "Debugging", []string{"--disable-debug"}, "Makes normal-use output smaller and simpler.", "Not a normal source default. Select this for ordinary release-style local builds; leave unchecked when investigating build or runtime problems.", false, false),
		configureOptionChoice("disable-stripping", "Keep symbol information", "Debugging", []string{"--disable-stripping"}, "Keeps more build information for debugging.", "Usually unnecessary for ordinary users, but useful when diagnosing crashes or build problems.", false, false),
	}
}

// Matches the "standard" option preset in the UI: the locked program defaults
// plus static external-library linking and skipping documentation, which is the
// sensible default for this static + external-library builder.
func defaultConfigureOptionIds() []string {
	return []string{"default-static", "default-programs", "default-ffmpeg", "default-ffprobe", "pkg-config-static", "disable-doc"}
}

// defaultMsys2PackageNames returns the base toolchain package list for the given
// Windows shell profile. The prefixed compiler/build packages are generated from
// the profile so a non-ucrt64 profile installs the matching toolchain instead of
// the ucrt64 one. The unprefixed MSYS packages (base-devel, git, make, diffutils)
// are profile-independent and stay as-is.
func defaultMsys2PackageNames(windowsShellProfileName string) []string {
	packagePrefix := packagePrefixForShellProfile(windowsShellProfileName)
	return []string{
		"base-devel",
		"git",
		"make",
		"diffutils",
		packagePrefix + "-binutils",
		packagePrefix + "-crt",
		packagePrefix + "-gcc",
		packagePrefix + "-headers",
		packagePrefix + "-libmangle",
		packagePrefix + "-libwinpthread",
		packagePrefix + "-make",
		packagePrefix + "-pkgconf",
		packagePrefix + "-tools",
		packagePrefix + "-winpthreads",
		packagePrefix + "-winstorecompat",
		packagePrefix + "-cmake",
		packagePrefix + "-ninja",
		packagePrefix + "-nasm",
		packagePrefix + "-yasm",
	}
}

func libraryChoice(id string, displayName string, categoryName string, flags []string, packages []string, licenseEffectName string) LibraryChoice {
	return trackedLibraryChoice(LibraryTrackNative, id, displayName, categoryName, flags, packages, licenseEffectName)
}

// Display prose (plain/technical explanations) is not set here: it lives solely in the
// localization files (catalog.libraries.<id>.plainExplanation / .technicalExplanation),
// which the frontend reads directly. The struct fields are left empty and only act as a
// never-hit fallback; TestCatalogLibraryLocalizationCoverage proves every id is localized.
func trackedLibraryChoice(trackName LibraryTrackName, id string, displayName string, categoryName string, flags []string, packages []string, licenseEffectName string) LibraryChoice {
	return LibraryChoice{LibraryId: id, TrackName: trackName, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: flags, PackageNames: packages, OfficialWebpageUrl: officialWebpageUrlForLibrary(id), LicenseEffectName: licenseEffectName, PlainExplanation: "", TechnicalExplanation: "", DefaultChecked: false, Locked: false}
}

func officialWebpageUrlForLibrary(libraryId string) string {
	officialWebpages := map[string]string{
		"aom":          "https://aomedia.googlesource.com/aom/",
		"amf":          "https://gpuopen.com/advanced-media-framework/",
		"aribb24":      "https://github.com/nkoriyama/aribb24",
		"aribcaption":  "https://github.com/xqq/libaribcaption",
		"ass":          "https://github.com/libass/libass",
		"avisynthplus": "https://avs-plus.net/",
		"bluray":       "https://www.videolan.org/developers/libbluray.html",
		"bs2b":         "https://bs2b.sourceforge.net/",
		"caca":         "http://caca.zoy.org/wiki/libcaca",
		"cairo":        "https://www.cairographics.org/",
		"cdio":         "https://www.gnu.org/software/libcdio/",
		"chromaprint":  "https://acoustid.org/chromaprint",
		"codec2":       "https://www.rowetel.com/codec2.html",
		"cuda-nvcc":    "https://developer.nvidia.com/cuda-toolkit",
		"dc1394":       "https://sourceforge.net/projects/libdc1394/",
		"decklink":     "https://www.blackmagicdesign.com/developer/product/capture-and-playback",
		"dav1d":        "https://code.videolan.org/videolan/dav1d",
		"davs2":        "https://github.com/pkuvcl/davs2",
		"dvdnav":       "https://www.videolan.org/developers/libdvdnav.html",
		"dvdread":      "https://www.videolan.org/developers/libdvdnav.html",
		"fdk-aac":      "https://github.com/mstorsjo/fdk-aac",
		"fontconfig":   "https://www.freedesktop.org/wiki/Software/fontconfig/",
		"freetype":     "https://freetype.org/",
		"frei0r":       "https://frei0r.dyne.org/",
		"fribidi":      "https://github.com/fribidi/fribidi",
		"glslang":      "https://github.com/KhronosGroup/glslang",
		"gme":          "https://github.com/libgme/game-music-emu",
		"gnutls":       "https://www.gnutls.org/",
		"gsm":          "https://www.quut.com/gsm/",
		"harfbuzz":     "https://harfbuzz.github.io/",
		"ilbc":         "https://github.com/TimothyGu/libilbc",
		"jack":         "https://jackaudio.org/",
		"klvanc":       "https://github.com/stoth68000/libklvanc",
		"kvazaar":      "https://github.com/ultravideo/kvazaar",
		"ladspa":       "https://www.ladspa.org/",
		"lc3":          "https://github.com/google/liblc3",
		"lcms2":        "https://www.littlecms.com/",
		"lcevc-dec":    "https://github.com/v-novaltd/LCEVCdec",
		"lensfun":      "https://lensfun.github.io/",
		"libjxl":       "https://jpeg.org/jpegxl/",
		"libmfx":       "https://github.com/lu-zero/mfx_dispatch",
		"libplacebo":   "https://code.videolan.org/videolan/libplacebo",
		"libtls":       "https://www.libressl.org/",
		"libvpx":       "https://chromium.googlesource.com/webm/libvpx/",
		"lv2":          "https://lv2plug.in/",
		"mbedtls":      "https://www.trustedfirmware.org/projects/mbed-tls/",
		"modplug":      "https://modplug-xmms.sourceforge.net/",
		"mpeghdec":     "https://github.com/Fraunhofer-IIS/mpeghdec",
		"mp3lame":      "https://lame.sourceforge.io/",
		"mysofa":       "https://github.com/hoene/libmysofa",
		"nvenc":        "https://developer.nvidia.com/video-codec-sdk",
		"oapv":         "https://github.com/openapv/openapv",
		"onnxruntime":  "https://onnxruntime.ai/",
		"openal":       "https://openal-soft.org/",
		"opencolorio":  "https://opencolorio.org/",
		"opencore-amr": "https://sourceforge.net/projects/opencore-amr/",
		"opencv":       "https://opencv.org/",
		"openh264":     "https://www.openh264.org/",
		"opencl":       "https://www.khronos.org/opencl/",
		"opengl":       "https://www.khronos.org/opengl/",
		"openjpeg":     "https://www.openjpeg.org/",
		"openmpt":      "https://lib.openmpt.org/libopenmpt/",
		"openssl":      "https://www.openssl.org/",
		"openvino":     "https://www.intel.com/content/www/us/en/developer/tools/openvino-toolkit/overview.html",
		"opus":         "https://opus-codec.org/",
		"png":          "http://www.libpng.org/pub/png/libpng.html",
		"pocketsphinx": "https://github.com/cmusphinx/pocketsphinx",
		"pulse":        "https://www.freedesktop.org/wiki/Software/PulseAudio/",
		"qrencode":     "https://fukuchi.org/works/qrencode/",
		"libvpl":       "https://www.intel.com/content/www/us/en/developer/tools/oneapi/onevpl.html",
		"quirc":        "https://github.com/dlbeer/quirc",
		"rabbitmq":     "https://github.com/alanxz/rabbitmq-c",
		"rav1e":        "https://github.com/xiph/rav1e",
		"rist":         "https://code.videolan.org/rist/librist",
		"rsvg":         "https://gitlab.gnome.org/GNOME/librsvg",
		"rtmp":         "https://rtmpdump.mplayerhq.hu/",
		"rubberband":   "https://breakfastquay.com/rubberband/",
		"sdl2":         "https://www.libsdl.org/",
		"shaderc":      "https://github.com/google/shaderc",
		"shine":        "https://github.com/toots/shine",
		"smbclient":    "https://www.samba.org/",
		"snappy":       "https://github.com/google/snappy",
		"soxr":         "https://sourceforge.net/p/soxr/wiki/Home/",
		"speex":        "https://www.speex.org/",
		"srt":          "https://www.srtalliance.org/",
		"ssh":          "https://www.libssh.org/",
		"svt-av1":      "https://gitlab.com/AOMediaCodec/SVT-AV1",
		"svtjpegxs":    "https://gitlab.com/SVT-JPEG-XS/SVT-JPEG-XS",
		"tensorflow":   "https://www.tensorflow.org/",
		"tesseract":    "https://tesseract-ocr.github.io/",
		"theora":       "https://www.theora.org/",
		"torch":        "https://pytorch.org/",
		"twolame":      "https://www.twolame.org/",
		"uavs3d":       "https://github.com/uavs3/uavs3d",
		"vapoursynth":  "https://www.vapoursynth.com/",
		"vidstab":      "https://github.com/georgmartius/vid.stab",
		"vmaf":         "https://github.com/Netflix/vmaf",
		"vo-amrwbenc":  "https://sourceforge.net/projects/opencore-amr/",
		"vorbis":       "https://xiph.org/vorbis/",
		"vvenc":        "https://github.com/fraunhoferhhi/vvenc",
		"webp":         "https://developers.google.com/speed/webp",
		"whisper":      "https://github.com/ggerganov/whisper.cpp",
		"x264":         "https://www.videolan.org/developers/x264.html",
		"x265":         "https://www.videolan.org/developers/x265.html",
		"xavs":         "https://xavs.sourceforge.io/",
		"xavs2":        "https://github.com/pkuvcl/xavs2",
		"xevd":         "https://github.com/mpeg5/xevd",
		"xevdb":        "https://github.com/mpeg5/xevd",
		"xeve":         "https://github.com/mpeg5/xeve",
		"xeveb":        "https://github.com/mpeg5/xeve",
		"xml2":         "https://gitlab.gnome.org/GNOME/libxml2",
		"xvid":         "https://www.xvid.com/",
		"zimg":         "https://github.com/sekrit-twc/zimg",
		"zmq":          "https://zeromq.org/",
		"zvbi":         "https://zapping.sourceforge.net/ZVBI/index.html",
	}
	return officialWebpages[libraryId]
}

func includedLibraryChoice(id string, displayName string, categoryName string) LibraryChoice {
	return LibraryChoice{LibraryId: id, TrackName: LibraryTrackNative, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: []string{}, PackageNames: []string{}, OfficialWebpageUrl: "https://ffmpeg.org/", LicenseEffectName: "included", PlainExplanation: "", TechnicalExplanation: "", DefaultChecked: true, Locked: true}
}

func defaultLibraryIds() []string {
	ids := []string{}
	for _, library := range LibraryCatalogForShellProfile("ucrt64") {
		if library.DefaultChecked {
			ids = append(ids, library.LibraryId)
		}
	}
	return ids
}

func mergeDefaultLibraryIds(selectedLibraryIds []string) []string {
	return mergeUniqueStrings(defaultLibraryIds(), selectedLibraryIds)
}

func configureOptionChoice(id string, displayName string, categoryName string, flags []string, plainExplanation string, technicalNote string, defaultEnabled bool, locked bool) ConfigureOptionChoice {
	return ConfigureOptionChoice{OptionId: id, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: flags, PlainExplanation: plainExplanation, TechnicalNote: technicalNote, DefaultEnabled: defaultEnabled, Locked: locked, RiskLevelName: configureOptionRiskLevel(id)}
}

// configureOptionRiskLevel rates how likely an option is to break the build, hurt
// performance, or surprise the user. high: enable-shared clashes with this builder's
// static external libs; disable-asm removes most SIMD speed. medium: a real but
// recoverable downside (lost protocols, slower decode, MinGW LTO link failures).
// Everything else is low. Shown as a colored pill next to each option.
func configureOptionRiskLevel(optionId string) string {
	switch optionId {
	case "enable-shared", "disable-asm":
		return "high"
	case "disable-network", "disable-x86asm", "enable-lto":
		return "medium"
	default:
		return "low"
	}
}

func packagePrefixForShellProfile(windowsShellProfileName string) string {
	switch windowsShellProfileName {
	case "mingw64":
		return "mingw-w64-x86_64"
	case "clang64":
		return "mingw-w64-clang-x86_64"
	case "ucrt64", "":
		return "mingw-w64-ucrt-x86_64"
	default:
		return "mingw-w64-ucrt-x86_64"
	}
}
