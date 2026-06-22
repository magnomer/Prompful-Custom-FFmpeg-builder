package planning

func LibraryCatalogForShellProfile(windowsShellProfileName string) []LibraryChoice {
	packagePrefix := packagePrefixForShellProfile(windowsShellProfileName)
	return []LibraryChoice{
		includedLibraryChoice("ffmpeg-program", "ffmpeg.exe", "Included by default (official FFmpeg source)", "The main command-line program users normally run.", "Built by default in a normal FFmpeg source build. No external package or --enable-lib flag is needed. Only disabled by selecting --disable-programs."),
		includedLibraryChoice("ffprobe-program", "ffprobe.exe", "Included by default (official FFmpeg source)", "The media inspection tool used to check files and streams.", "Built by default in a normal FFmpeg source build. No external package or --enable-lib flag is needed. Disabled by selecting --disable-programs or --disable-ffprobe."),
		includedLibraryChoice("libavcodec", "libavcodec", "Included by default (official FFmpeg source)", "FFmpeg's built-in codec library for encoding and decoding media.", "This is one of FFmpeg's own libraries. Native decoders and encoders are enabled by default; external-library codecs still require explicit --enable-lib... flags."),
		includedLibraryChoice("libavformat", "libavformat", "Included by default (official FFmpeg source)", "FFmpeg's built-in library for reading and writing media containers such as MP4, MOV, MKV, and WAV.", "This is one of FFmpeg's own libraries and does not require an external MSYS2 package."),
		includedLibraryChoice("libavfilter", "libavfilter", "Included by default (official FFmpeg source)", "FFmpeg's built-in filtering library for scaling, trimming, overlaying, subtitles, audio filters, and more.", "This is one of FFmpeg's own libraries. Some filters can become more capable when external libraries are selected."),
		includedLibraryChoice("libavutil", "libavutil", "Included by default (official FFmpeg source)", "Shared utility code used by the rest of FFmpeg.", "This is one of FFmpeg's own libraries and is required by normal FFmpeg builds."),
		includedLibraryChoice("native-codecs", "Native FFmpeg codecs", "Included by default (official FFmpeg source)", "FFmpeg includes many native decoders and encoders before you add external codec libraries.", "Native FFmpeg decoders and encoders are enabled in a normal source build. External codec libraries are separate choices and use their own --enable-lib... flags."),
		includedLibraryChoice("native-formats", "Native formats and muxers", "Included by default (official FFmpeg source)", "FFmpeg includes many built-in readers and writers for media containers.", "These are part of FFmpeg itself. External libraries can add support or improve specific formats, but the base format layer is not an external package."),
		includedLibraryChoice("libswscale", "libswscale", "Included by default (official FFmpeg source)", "Built-in image scaling and pixel-format conversion.", "This is one of FFmpeg's own libraries and does not require an external package."),
		includedLibraryChoice("libswresample", "libswresample", "Included by default (official FFmpeg source)", "Built-in audio resampling and sample-format conversion.", "This is one of FFmpeg's own libraries and does not require an external package."),

		// Video encoders — ranked by practical usefulness: H.264/HEVC/AV1 first, legacy/niche last.
		libraryChoice("x264", "x264", "Video encoders", []string{"--enable-libx264"}, []string{packagePrefix + "-libx264"}, "gpl", "Adds H.264 encoding. Good compatibility, but it changes the build to GPL."),
		libraryChoice("x265", "x265", "Video encoders", []string{"--enable-libx265"}, []string{packagePrefix + "-x265"}, "gpl", "Adds HEVC/H.265 encoding. Smaller files than H.264, but GPL."),
		libraryChoice("svt-av1", "SVT-AV1", "Video encoders", []string{"--enable-libsvtav1"}, []string{packagePrefix + "-svt-av1"}, "lgpl", "Adds a fast AV1 encoder."),
		libraryChoice("libvpx", "libvpx", "Video encoders", []string{"--enable-libvpx"}, []string{packagePrefix + "-libvpx"}, "lgpl", "Adds VP8/VP9 support for WebM files."),
		libraryChoice("aom", "AOM AV1", "Video encoders", []string{"--enable-libaom"}, []string{packagePrefix + "-aom"}, "lgpl", "Adds AV1 encoding/decoding through libaom."),
		libraryChoice("openh264", "OpenH264", "Video encoders", []string{"--enable-libopenh264"}, []string{packagePrefix + "-openh264"}, "lgpl", "Adds Cisco OpenH264 support."),
		libraryChoice("rav1e", "rav1e", "Video encoders", []string{"--enable-librav1e"}, []string{packagePrefix + "-rav1e"}, "lgpl", "Adds the rav1e AV1 encoder."),
		libraryChoice("xvid", "Xvid", "Video encoders", []string{"--enable-libxvid"}, []string{packagePrefix + "-xvidcore"}, "gpl", "Adds Xvid MPEG-4 Part 2 video encoding. GPL effect."),
		libraryChoice("theora", "libtheora", "Video encoders", []string{"--enable-libtheora"}, []string{packagePrefix + "-libtheora"}, "lgpl", "Adds Ogg Theora video encoding and decoding."),
		libraryChoice("xeve", "XEVE (EVC)", "Video encoders", []string{"--enable-libxeve"}, []string{packagePrefix + "-xeve"}, "lgpl", "Adds Essential Video Coding (EVC) encoding for broadcast and research workflows."),
		libraryChoice("oapv", "liboapv (APV)", "Video encoders", []string{"--enable-liboapv"}, []string{packagePrefix + "-openapv"}, "lgpl", "Adds AP Video (APV) professional intraframe codec support through liboapv."),
		libraryChoice("vvenc", "vvenc (VVC/H.266)", "Video encoders", []string{"--enable-libvvenc"}, []string{}, "lgpl", "Adds Versatile Video Coding (VVC/H.266) encoding.\n**No prebuilt MSYS2 package exists**, so you must build and install vvenc into the selected MSYS2 prefix yourself or FFmpeg configure will fail for this library."),
		libraryChoice("xavs2", "xavs2", "Video encoders", []string{"--enable-libxavs2"}, []string{}, "gpl", "Adds AVS2 video encoding. GPL effect.\n**No prebuilt MSYS2 package exists**, so you must build and install xavs2 into the selected MSYS2 prefix yourself or FFmpeg configure will fail for this library."),

		// Hardware encoders — NVENC/QSV/AMF. OpenCL moved to Filters and processing (it is GPU compute, not an encoder).
		libraryChoice("nvenc", "NVIDIA NVENC", "Hardware encoders", []string{"--enable-ffnvcodec"}, []string{packagePrefix + "-ffnvcodec-headers"}, "lgpl", "Enables NVIDIA GPU video encoding (h264_nvenc, hevc_nvenc, av1_nvenc) and CUVID decoding. NVIDIA libraries load at runtime from the driver; the build stays LGPL. Requires a compatible NVIDIA GPU and driver."),
		libraryChoice("qsv", "Intel QSV (oneVPL)", "Hardware encoders", []string{"--enable-libvpl"}, []string{packagePrefix + "-libvpl"}, "lgpl", "Enables Intel Quick Sync hardware encoding and decoding (h264_qsv, hevc_qsv, av1_qsv, vp9_qsv). Requires an Intel GPU with Quick Sync support and the Intel oneVPL runtime."),
		libraryChoice("amf", "AMD AMF", "Hardware encoders", []string{"--enable-amf"}, []string{packagePrefix + "-amf-headers"}, "lgpl", "Enables AMD GPU video encoding (h264_amf, hevc_amf, av1_amf). Requires AMF headers plus a compatible AMD GPU and driver."),

		// Video decoders
		libraryChoice("dav1d", "dav1d", "Video decoders", []string{"--enable-libdav1d"}, []string{packagePrefix + "-dav1d"}, "lgpl", "Adds a fast AV1 decoder."),
		libraryChoice("xevd", "XEVD (EVC dec)", "Video decoders", []string{"--enable-libxevd"}, []string{packagePrefix + "-xevd"}, "lgpl", "Adds Essential Video Coding (EVC) decoding through xevd."),

		// Image codecs
		libraryChoice("webp", "WebP", "Image codecs", []string{"--enable-libwebp"}, []string{packagePrefix + "-libwebp"}, "lgpl", "Adds WebP image support."),
		libraryChoice("openjpeg", "OpenJPEG", "Image codecs", []string{"--enable-libopenjpeg"}, []string{packagePrefix + "-openjpeg2"}, "lgpl", "Adds JPEG 2000 support."),
		libraryChoice("libjxl", "JPEG XL", "Image codecs", []string{"--enable-libjxl"}, []string{packagePrefix + "-libjxl"}, "lgpl", "Adds JPEG XL image support."),
		libraryChoice("rsvg", "librsvg", "Image codecs", []string{"--enable-librsvg"}, []string{packagePrefix + "-librsvg"}, "lgpl", "Adds SVG rasterization so vector overlays and images can be rendered."),
		libraryChoice("png", "libpng", "Image codecs", []string{}, []string{packagePrefix + "-libpng"}, "lgpl", "PNG support is usually native; this installs libpng for filters/tools that need it."),
		libraryChoice("snappy", "Snappy", "Image codecs", []string{"--enable-libsnappy"}, []string{packagePrefix + "-snappy"}, "lgpl", "Adds Snappy compression used by the HAP GPU video codec."),
		libraryChoice("svtjpegxs", "SVT JPEG XS", "Image codecs", []string{"--enable-libsvtjpegxs"}, []string{packagePrefix + "-svt-jpeg-xs", "git", packagePrefix + "-cmake", packagePrefix + "-ninja", packagePrefix + "-yasm"}, "lgpl", "Hidden from the UI for now. Left in the backend for future compatibility: when explicitly enabled, the builder tries the MSYS2 package first, then the official upstream source, and skips --enable-libsvtjpegxs with a warning if neither is compatible."),

		// Filters and processing — zimg/libplacebo/VMAF/vidstab first; OpenCL relocated here; glslang below shaderc.
		libraryChoice("zimg", "zimg", "Filters and processing", []string{"--enable-libzimg"}, []string{packagePrefix + "-zimg"}, "lgpl", "Adds high-quality resizing, colorspace, and bit-depth conversion filters."),
		libraryChoice("libplacebo", "libplacebo", "Filters and processing", []string{"--enable-libplacebo", "--enable-vulkan"}, []string{packagePrefix + "-libplacebo", packagePrefix + "-vulkan-loader", packagePrefix + "-vulkan-headers"}, "lgpl", "Adds GPU-oriented video processing filters."),
		libraryChoice("vmaf", "libvmaf", "Filters and processing", []string{"--enable-libvmaf"}, []string{packagePrefix + "-vmaf"}, "lgpl", "Adds Netflix VMAF video quality measurement filters."),
		libraryChoice("vidstab", "libvidstab", "Filters and processing", []string{"--enable-libvidstab"}, []string{packagePrefix + "-vid.stab"}, "lgpl", "Adds video stabilization filter support."),
		libraryChoice("shaderc", "libshaderc", "Filters and processing", []string{"--enable-libshaderc"}, []string{packagePrefix + "-shaderc"}, "lgpl", "Adds runtime Vulkan shader compilation used by GPU filters such as libplacebo."),
		libraryChoice("lcms2", "Little CMS 2", "Filters and processing", []string{"--enable-lcms2"}, []string{packagePrefix + "-lcms2"}, "lgpl", "Adds ICC color-management support for accurate color handling."),
		libraryChoice("opencolorio", "OpenColorIO", "Filters and processing", []string{"--enable-libopencolorio"}, []string{packagePrefix + "-opencolorio"}, "lgpl", "Adds OpenColorIO color-management and color-transform support."),
		libraryChoice("cairo", "Cairo", "Filters and processing", []string{"--enable-cairo"}, []string{packagePrefix + "-cairo"}, "lgpl", "Adds 2D vector graphics rendering support for graphics-oriented filters."),
		libraryChoice("qrencode", "libqrencode", "Filters and processing", []string{"--enable-libqrencode"}, []string{packagePrefix + "-qrencode"}, "lgpl", "Adds QR code generation for the qrencode source filter."),
		libraryChoice("frei0r", "frei0r", "Filters and processing", []string{"--enable-frei0r"}, []string{packagePrefix + "-frei0r-plugins"}, "gpl", "Adds frei0r video effects. GPL effect."),
		libraryChoice("ladspa", "LADSPA", "Filters and processing", []string{"--enable-ladspa"}, []string{packagePrefix + "-ladspa-sdk", packagePrefix + "-dlfcn"}, "lgpl", "Adds hosting of LADSPA audio plugins through the ladspa filter."),
		libraryChoice("lv2", "LV2", "Filters and processing", []string{"--enable-lv2"}, []string{packagePrefix + "-lilv"}, "lgpl", "Adds hosting of LV2 audio plugins through the lv2 filter."),
		libraryChoice("opencv", "OpenCV", "Filters and processing", []string{"--enable-libopencv"}, []string{packagePrefix + "-opencv"}, "lgpl", "Adds OpenCV-backed video filtering and computer-vision processing support."),
		libraryChoice("opencl", "OpenCL", "Filters and processing", []string{"--enable-opencl"}, []string{packagePrefix + "-opencl-icd", packagePrefix + "-opencl-headers"}, "lgpl", "Adds OpenCL GPU acceleration for filters that support it. Requires an OpenCL runtime from your GPU driver."),
		libraryChoice("glslang", "glslang", "Filters and processing", []string{"--enable-libglslang"}, []string{packagePrefix + "-glslang"}, "lgpl", "Adds runtime GLSL/HLSL to SPIR-V shader compilation support."),
		libraryChoice("lensfun", "lensfun", "Filters and processing", []string{"--enable-liblensfun"}, []string{packagePrefix + "-lensfun"}, "lgpl", "Hidden from the UI for now. Left in the backend for future compatibility: when explicitly enabled, the builder checks whether the installed lensfun exposes the API required by this FFmpeg source and skips --enable-liblensfun with a warning if incompatible."),

		// Audio — Opus/FDK-AAC/SoX/LAME/Rubber Band first; legacy and niche codecs last.
		libraryChoice("opus", "Opus", "Audio", []string{"--enable-libopus"}, []string{packagePrefix + "-opus"}, "lgpl", "Adds Opus audio encoding/decoding."),
		libraryChoice("fdk-aac", "Fraunhofer FDK AAC", "Audio", []string{"--enable-libfdk-aac", "--enable-nonfree"}, []string{packagePrefix + "-fdk-aac"}, "nonfree", "Adds a high-quality AAC encoder. Makes the FFmpeg build nonfree."),
		libraryChoice("soxr", "SoX Resampler", "Audio", []string{"--enable-libsoxr"}, []string{packagePrefix + "-libsoxr"}, "lgpl", "Adds high-quality audio resampling."),
		libraryChoice("mp3lame", "LAME MP3", "Audio", []string{"--enable-libmp3lame"}, []string{packagePrefix + "-lame"}, "lgpl", "Adds MP3 encoding."),
		libraryChoice("rubberband", "Rubber Band", "Audio", []string{"--enable-librubberband"}, []string{packagePrefix + "-rubberband"}, "gpl", "Adds high quality audio time-stretching/pitch-shifting. GPL effect."),
		libraryChoice("vorbis", "Vorbis", "Audio", []string{"--enable-libvorbis"}, []string{packagePrefix + "-libvorbis"}, "lgpl", "Adds Vorbis audio encoding."),
		libraryChoice("twolame", "TwoLAME", "Audio", []string{"--enable-libtwolame"}, []string{packagePrefix + "-twolame"}, "lgpl", "Adds MP2 audio encoding."),
		libraryChoice("speex", "Speex", "Audio", []string{"--enable-libspeex"}, []string{packagePrefix + "-speex"}, "lgpl", "Adds Speex speech codec support."),
		libraryChoice("lc3", "LC3", "Audio", []string{"--enable-liblc3"}, []string{packagePrefix + "-liblc3"}, "lgpl", "Adds Low Complexity Communication Codec (LC3) audio support used by Bluetooth LE Audio."),
		libraryChoice("opencore-amr", "OpenCORE AMR", "Audio", []string{"--enable-libopencore-amrnb", "--enable-libopencore-amrwb"}, []string{packagePrefix + "-opencore-amr"}, "lgpl", "Adds AMR-NB and AMR-WB decoding support for mobile-phone voice recordings. This requires FFmpeg's version-3 license switch."),
		libraryChoice("vo-amrwbenc", "VisualOn AMR-WB encoder", "Audio", []string{"--enable-libvo-amrwbenc"}, []string{packagePrefix + "-vo-amrwbenc"}, "lgpl", "Adds AMR-WB encoding support for narrow telephony workflows. This requires FFmpeg's version-3 license switch."),
		libraryChoice("gsm", "GSM", "Audio", []string{"--enable-libgsm"}, []string{packagePrefix + "-gsm"}, "lgpl", "Adds GSM audio codec support."),
		libraryChoice("ilbc", "iLBC", "Audio", []string{"--enable-libilbc"}, []string{packagePrefix + "-libilbc"}, "lgpl", "Adds iLBC speech codec support."),
		libraryChoice("mysofa", "libmysofa", "Audio", []string{"--enable-libmysofa"}, []string{packagePrefix + "-libmysofa"}, "lgpl", "Adds SOFA HRTF loading for the sofalizer binaural spatial-audio filter."),
		libraryChoice("bs2b", "libbs2b", "Audio", []string{"--enable-libbs2b"}, []string{packagePrefix + "-libbs2b"}, "lgpl", "Adds Bauer stereophonic-to-binaural processing for more natural headphone listening."),
		libraryChoice("chromaprint", "Chromaprint", "Audio", []string{"--enable-chromaprint"}, []string{packagePrefix + "-chromaprint"}, "lgpl", "Adds audio fingerprinting used for acoustic identification."),
		libraryChoice("whisper", "whisper.cpp", "Audio", []string{"--enable-whisper"}, []string{packagePrefix + "-whisper.cpp", packagePrefix + "-ggml"}, "lgpl", "Adds the whisper speech-to-text filter for generating transcripts from audio."),
		libraryChoice("gme", "Game Music Emu", "Audio", []string{"--enable-libgme"}, []string{packagePrefix + "-libgme"}, "lgpl", "Adds support for game music formats such as NSF, SPC, GYM, and other chiptune files."),
		libraryChoice("shine", "Shine MP3", "Audio", []string{"--enable-libshine"}, []string{packagePrefix + "-shine"}, "lgpl", "Adds a fast fixed-point MP3 encoder, useful where floating-point speed is limited."),
		libraryChoice("codec2", "Codec 2", "Audio", []string{"--enable-libcodec2"}, []string{packagePrefix + "-codec2"}, "lgpl", "Adds Codec 2 very-low-bitrate speech encoding and decoding."),

		// Subtitles and text — libass/FreeType/Fontconfig/HarfBuzz/FriBidi first; regional captions last.
		libraryChoice("ass", "libass", "Subtitles and text", []string{"--enable-libass"}, []string{packagePrefix + "-libass"}, "lgpl", "Adds advanced subtitle rendering."),
		libraryChoice("freetype", "FreeType", "Subtitles and text", []string{"--enable-libfreetype"}, []string{packagePrefix + "-freetype"}, "lgpl", "Adds font rendering for subtitles and text filters."),
		libraryChoice("fontconfig", "Fontconfig", "Subtitles and text", []string{"--enable-libfontconfig"}, []string{packagePrefix + "-fontconfig"}, "lgpl", "Adds font discovery support."),
		libraryChoice("harfbuzz", "HarfBuzz", "Subtitles and text", []string{"--enable-libharfbuzz"}, []string{packagePrefix + "-harfbuzz"}, "lgpl", "Adds advanced text shaping support."),
		libraryChoice("fribidi", "FriBidi", "Subtitles and text", []string{"--enable-libfribidi"}, []string{packagePrefix + "-fribidi"}, "lgpl", "Adds bidirectional text support."),
		libraryChoice("aribcaption", "libaribcaption", "Subtitles and text", []string{"--enable-libaribcaption"}, []string{packagePrefix + "-libaribcaption"}, "lgpl", "Adds an alternative ARIB STD-B24 caption decoder/renderer with richer styling."),
		libraryChoice("aribb24", "libaribb24", "Subtitles and text", []string{"--enable-libaribb24"}, []string{packagePrefix + "-aribb24"}, "lgpl", "Adds Japanese ARIB STD-B24 caption decoding. Requires FFmpeg's version-3 license switch."),
		libraryChoice("zvbi", "libzvbi", "Subtitles and text", []string{"--enable-libzvbi"}, []string{packagePrefix + "-zvbi"}, "gpl", "Adds teletext decoding from DVB/analog VBI data. GPL effect."),

		// Disc and device input
		libraryChoice("bluray", "libbluray", "Disc and device input", []string{"--enable-libbluray"}, []string{packagePrefix + "-libbluray"}, "lgpl", "Adds Blu-ray reading support."),
		libraryChoice("openmpt", "libopenmpt", "Disc and device input", []string{"--enable-libopenmpt"}, []string{packagePrefix + "-libopenmpt"}, "lgpl", "Adds modern tracker/module audio support (MOD, XM, S3M, IT, and others). More actively maintained than libmodplug."),
		libraryChoice("openal", "OpenAL", "Disc and device input", []string{"--enable-openal"}, []string{packagePrefix + "-openal"}, "lgpl", "Adds OpenAL audio input support."),
		libraryChoice("sdl2", "SDL2", "Disc and device input", []string{"--enable-sdl2"}, []string{packagePrefix + "-SDL2"}, "lgpl", "Adds SDL2 support, mainly useful for ffplay."),
		libraryChoice("dvdread", "libdvdread", "Disc and device input", []string{"--enable-libdvdread"}, []string{packagePrefix + "-libdvdread"}, "gpl", "Adds DVD-Video disc structure reading support. GPL effect."),
		libraryChoice("cdio", "libcdio", "Disc and device input", []string{"--enable-libcdio"}, []string{packagePrefix + "-libcdio", packagePrefix + "-libcdio-paranoia"}, "gpl", "Adds CD input support. GPL effect."),
		libraryChoice("modplug", "libmodplug", "Disc and device input", []string{"--enable-libmodplug"}, []string{packagePrefix + "-libmodplug"}, "lgpl", "Adds tracker/module audio file support."),
		libraryChoice("jack", "JACK", "Disc and device input", []string{"--enable-libjack"}, []string{packagePrefix + "-jack2"}, "lgpl", "Adds JACK low-latency audio input/output support."),
		libraryChoice("pulse", "PulseAudio", "Disc and device input", []string{"--enable-libpulse"}, []string{packagePrefix + "-pulseaudio"}, "lgpl", "Adds PulseAudio input support."),
		libraryChoice("caca", "libcaca", "Disc and device input", []string{"--enable-libcaca"}, []string{packagePrefix + "-libcaca"}, "lgpl", "Adds colored ASCII-art video output through the caca output device."),

		// Network — TLS first, then streaming/transport protocols. libxml2 moved to Support libraries.
		libraryChoice("openssl", "OpenSSL", "Network", []string{"--enable-openssl"}, []string{packagePrefix + "-openssl"}, "nonfree", "Adds HTTPS/TLS network support through OpenSSL. Choose this instead of GnuTLS when you specifically need OpenSSL-compatible TLS behavior."),
		libraryChoice("gnutls", "GnuTLS", "Network", []string{"--enable-gnutls"}, []string{packagePrefix + "-gnutls"}, "lgpl", "Adds HTTPS/TLS network support through GnuTLS. Choose this instead of OpenSSL when you want TLS support without OpenSSL's redistribution concerns."),
		libraryChoice("srt", "SRT", "Network", []string{"--enable-libsrt"}, []string{packagePrefix + "-srt"}, "lgpl", "Adds Secure Reliable Transport protocol support."),
		libraryChoice("ssh", "libssh", "Network", []string{"--enable-libssh"}, []string{packagePrefix + "-libssh"}, "lgpl", "Adds SSH protocol support."),
		libraryChoice("rtmp", "librtmp", "Network", []string{"--enable-librtmp"}, []string{packagePrefix + "-rtmpdump"}, "lgpl", "Adds RTMP/RTMPE streaming through librtmp instead of FFmpeg's native RTMP."),
		libraryChoice("rist", "librist", "Network", []string{"--enable-librist"}, []string{packagePrefix + "-librist"}, "lgpl", "Adds RIST streaming protocol support."),
		libraryChoice("zmq", "ZeroMQ", "Network", []string{"--enable-libzmq"}, []string{packagePrefix + "-zeromq"}, "lgpl", "Adds ZeroMQ messaging support."),
		libraryChoice("rabbitmq", "RabbitMQ-C", "Network", []string{"--enable-librabbitmq"}, []string{packagePrefix + "-rabbitmq-c"}, "lgpl", "Adds RabbitMQ/AMQP messaging protocol support."),

		// OCR
		libraryChoice("tesseract", "Tesseract OCR", "OCR", []string{"--enable-libtesseract"}, []string{packagePrefix + "-tesseract-ocr"}, "lgpl", "Adds OCR filter support through Tesseract."),

		// Support libraries — internal dependencies pulled in for specific formats/workflows, not user-facing codecs.
		libraryChoice("xml2", "libxml2", "Support libraries", []string{"--enable-libxml2"}, []string{packagePrefix + "-libxml2"}, "lgpl", "Adds XML parsing support used by some formats/protocols."),
	}
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

func defaultMsys2PackageNames() []string {
	return []string{
		"base-devel",
		"git",
		"make",
		"diffutils",
		"mingw-w64-ucrt-x86_64-binutils",
		"mingw-w64-ucrt-x86_64-crt",
		"mingw-w64-ucrt-x86_64-gcc",
		"mingw-w64-ucrt-x86_64-headers",
		"mingw-w64-ucrt-x86_64-libmangle",
		"mingw-w64-ucrt-x86_64-libwinpthread",
		"mingw-w64-ucrt-x86_64-make",
		"mingw-w64-ucrt-x86_64-pkgconf",
		"mingw-w64-ucrt-x86_64-tools",
		"mingw-w64-ucrt-x86_64-winpthreads",
		"mingw-w64-ucrt-x86_64-winstorecompat",
		"mingw-w64-ucrt-x86_64-cmake",
		"mingw-w64-ucrt-x86_64-ninja",
		"mingw-w64-ucrt-x86_64-nasm",
		"mingw-w64-ucrt-x86_64-yasm",
	}
}

func libraryChoice(id string, displayName string, categoryName string, flags []string, packages []string, licenseEffectName string, reviewNote string) LibraryChoice {
	plainExplanation := reviewNote
	technicalExplanation := libraryDefaultReason(id, licenseEffectName)
	return LibraryChoice{LibraryId: id, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: flags, PackageNames: packages, LicenseEffectName: licenseEffectName, ReviewNote: reviewNote, PlainExplanation: plainExplanation, TechnicalExplanation: technicalExplanation, DefaultChecked: false, Locked: false}
}

func libraryDefaultReason(libraryId string, licenseEffectName string) string {
	libraryReasons := map[string]string{
		"x264":         "Most users can already decode H.264 with FFmpeg's native code. Check this only when you need to encode H.264 with the widely compatible x264 encoder. It is not default because it changes the resulting build to GPL.",
		"x265":         "Most users can already handle many HEVC files with FFmpeg's native code. Check this when you need x265 HEVC encoding for smaller files at slower encode speed. It is not default because it changes the resulting build to GPL.",
		"libvpx":       "Check this when you need WebM VP8/VP9 encoding or libvpx-specific behavior. It is not default because it adds an external codec package; ordinary MP4 workflows do not need it.",
		"aom":          "Check this when you need standards-focused AV1 encoding/decoding through libaom. It is not default because AV1 encoding is slow and many users do not need to create AV1 files.",
		"svt-av1":      "Check this when you need faster AV1 encoding than libaom on modern CPUs. It is not default because it adds a large external encoder and is unnecessary unless you are creating AV1 video.",
		"rav1e":        "Check this when you specifically want the rav1e AV1 encoder. It is not default because it is a specialized AV1 encoder and most users do not need multiple AV1 encoders.",
		"openh264":     "Check this when you specifically need Cisco OpenH264 behavior or compatibility. It is not default because x264 is usually the preferred H.264 encoder when GPL is acceptable, and native FFmpeg already covers many H.264 tasks.",
		"xavs2":        "Check this when you need AVS2 video encoding. It is not default because AVS2 is uncommon outside specific regional or archival workflows, it changes the resulting build to GPL, and MSYS2 ships no prebuilt xavs2 package — you must provide the library yourself or the build will fail at FFmpeg configure.",
		"dav1d":        "Check this when you need fast AV1 playback/decoding. It is not default because native FFmpeg can still be built without it, and users who do not decode AV1 do not need the extra package.",
		"libjxl":       "Check this when you need JPEG XL image support. It is not default because JPEG XL is not needed for common video workflows and requires an external image codec package.",
		"openjpeg":     "Check this when you need JPEG 2000 files, often used in cinema, archive, or medical/media workflows. It is not default because JPEG 2000 is not common for ordinary video conversion.",
		"webp":         "Check this when you need to read or write WebP images or animated WebP. It is not default because normal video workflows usually use video containers rather than WebP.",
		"png":          "Check this only when you need libpng for related tools or filters. It is not default because basic PNG image handling is commonly available through FFmpeg's native code.",
		"zimg":         "Check this when you need high-quality resizing, color conversion, or bit-depth conversion. It is not default because normal scaling works without it, while zimg adds a specialized processing dependency.",
		"libplacebo":   "Check this when you need advanced GPU-style video rendering, HDR tone mapping, or shader-based processing. It is not default because it is specialized and adds a graphics-processing dependency.",
		"vmaf":         "Check this when you need objective video quality scoring with VMAF. It is not default because it is for testing/comparison workflows, not normal playback or conversion.",
		"frei0r":       "Check this when you need frei0r video effects filters. It is not default because these effects are optional creative filters and the dependency changes the resulting build to GPL.",
		"rubberband":   "Check this when you need high-quality audio time-stretching or pitch shifting. It is not default because it is only needed for special audio processing and changes the resulting build to GPL.",
		"opus":         "Check this when you need Opus audio for WebM, streaming, voice, or low-bitrate audio. It is not default because it adds an external codec package; simple AAC/MP3-style workflows may not need it.",
		"vorbis":       "Check this when you need Vorbis audio, usually for Ogg/WebM-style workflows. It is not default because Vorbis is less common in ordinary MP4 workflows.",
		"mp3lame":      "Check this when you need reliable MP3 encoding. It is not default because MP3 output is only needed for audio export workflows and adds an external encoder package.",
		"twolame":      "Check this when you need MP2 audio encoding for broadcast or legacy workflows. It is not default because MP2 is uncommon for ordinary users.",
		"soxr":         "Check this when you need very high-quality audio resampling. It is not default because FFmpeg can resample audio without it, and many users do not need the extra quality/speed tradeoff.",
		"speex":        "Check this when you need the older Speex speech codec. It is not default because Opus is usually preferred for modern speech/audio workflows.",
		"gsm":          "Check this when you need GSM audio compatibility for old telephony files. It is not default because GSM is a legacy/specialized codec.",
		"ilbc":         "Check this when you need iLBC speech codec compatibility. It is not default because iLBC is a specialized voice codec.",
		"opencore-amr": "Check this when you need AMR-NB or AMR-WB support for phone/voice recordings. It is not default because AMR is a specialized mobile/telephony format, and FFmpeg requires --enable-version3 for this library.",
		"vo-amrwbenc":  "Check this when you need AMR-WB encoding for narrow telephony workflows. It is not default because the format is specialized, and FFmpeg requires --enable-version3 for this library.",
		"fdk-aac":      "Check this when you specifically need the Fraunhofer AAC encoder. It is not default because it makes the resulting FFmpeg build nonfree and limits redistribution.",
		"freetype":     "Check this when you need text drawing, subtitles, or filters that render fonts. It is not default because pure audio/video conversion may not render text.",
		"fontconfig":   "Check this when subtitle/text rendering should find installed fonts by name. It is not default because it is only useful with font-rendering workflows.",
		"fribidi":      "Check this when you need correct right-to-left or bidirectional text rendering. It is not default because it is only needed for languages/scripts that require bidi handling.",
		"harfbuzz":     "Check this when you need high-quality shaping for complex scripts in subtitles or drawn text. It is not default because basic Latin text does not need it.",
		"ass":          "Check this when you need styled ASS/SSA subtitle rendering, common for fansubs and complex subtitles. It is not default because simple subtitle copy/conversion may not need rendered subtitles.",
		"bluray":       "Check this when you need to read Blu-ray structures. It is not default because many users work with ordinary files rather than disc layouts.",
		"cdio":         "Check this when you need direct CD input support. It is not default because CD input is uncommon and it changes the resulting build to GPL.",
		"dvdread":      "Check this when you need DVD-Video disc structure reading support. It is not default because DVD input is uncommon for ordinary file conversion and changes the resulting build to GPL.",
		"modplug":      "Check this when you need tracker/module audio formats such as MOD/XM/S3M. It is not default because those formats are niche.",
		"openal":       "Check this when you need OpenAL audio input. It is not default because it is a device/input feature, not needed for normal file conversion.",
		"jack":         "Check this when you need JACK audio input/output for low-latency routing workflows. It is not default because it is a specialized device feature, especially on Windows.",
		"pulse":        "Check this when you need PulseAudio input support. It is not default because it is a device/audio-server feature rather than normal file conversion support.",
		"sdl2":         "Check this when you want ffplay or SDL-based playback support. It is not default because this builder is mainly for ffmpeg/ffprobe workflows, and playback UI support adds a separate dependency.",
		"openssl":      "Check this when FFmpeg must use HTTPS/TLS through OpenSSL. Do not select it together with GnuTLS: FFmpeg's configure script rejects both TLS backends at the same time. It is not default because it can make redistribution more sensitive and many local conversions do not need network TLS.",
		"gnutls":       "Check this when FFmpeg must use HTTPS/TLS through GnuTLS. Do not select it together with OpenSSL: FFmpeg's configure script rejects both TLS backends at the same time. It is not default because local file conversion does not need network TLS support.",
		"srt":          "Check this when you need Secure Reliable Transport streaming. It is not default because SRT is for live/remote streaming workflows, not ordinary local conversion.",
		"ssh":          "Check this when FFmpeg must read or write through SSH/SFTP-style protocols. It is not default because most users use local files or HTTPS instead.",
		"zmq":          "Check this when you need ZeroMQ filter/control messaging. It is not default because it is for automation/control workflows rather than normal conversion.",
		"rabbitmq":     "Check this when you need AMQP/RabbitMQ messaging protocol support. It is not default because it is a niche network/messaging workflow, not ordinary media conversion.",
		"rist":         "Check this when you need RIST broadcast/streaming transport. It is not default because it is a specialized live-video transport.",
		"xml2":         "Check this when a format, manifest, or protocol needs XML parsing. It is not default because common local conversion usually does not need XML support.",
		"tesseract":    "Check this when you need OCR from video/images. It is not default because OCR is a heavy, specialized feature and requires extra language data for useful results.",
		"vvenc":        "Check this when you need VVC/H.266 encoding. It is not default because VVC is an emerging codec with very limited decoder support outside research tools, and MSYS2 ships no prebuilt vvenc package — you must provide the library yourself or the build will fail at FFmpeg configure.",
		"xeve":         "Check this when you need EVC encoding for broadcast or research. It is not default because EVC is only used in specialized workflows.",
		"oapv":         "Check this when you need AP Video (APV) professional intraframe codec support. It is not default because APV is a new, specialized format aimed at high-quality production workflows rather than ordinary delivery.",
		"nvenc":        "Check this when you need NVIDIA GPU-accelerated video encoding (h264_nvenc, hevc_nvenc, av1_nvenc) or CUVID hardware decoding. It is not default because it requires a compatible NVIDIA GPU and driver.",
		"amf":          "Check this when you need AMD GPU-accelerated video encoding (h264_amf, hevc_amf, av1_amf). It is not default because it requires a compatible AMD GPU and driver.",
		"qsv":          "Check this when you need Intel Quick Sync hardware encoding or decoding (h264_qsv, hevc_qsv, av1_qsv, vp9_qsv). It is not default because it requires an Intel GPU with Quick Sync support and the Intel oneVPL runtime installed.",
		"xvid":         "Check this when you need Xvid MPEG-4 Part 2 encoding for compatibility with older devices or workflows. It is not default because MPEG-4 Part 2 has been largely superseded by H.264, and it changes the resulting build to GPL.",
		"theora":       "Check this when you need Ogg Theora video support for OGG containers. It is not default because Theora is an older format with limited modern use.",
		"vidstab":      "Check this when you need video stabilization filters to remove camera shake. It is not default because it is a specialized post-processing feature.",
		"cairo":        "Check this when you need graphics-oriented filter support based on Cairo drawing. It is not default because it is a specialized rendering dependency rather than a codec.",
		"glslang":      "Check this when you need runtime shader compilation for Vulkan/shader-based processing paths. It is not default because many builds can use precompiled shaders or do not use GPU shader filters.",
		"opencv":       "Check this when you need OpenCV-backed video filtering or computer-vision processing. It is not default because OpenCV is a large dependency for specialized workflows.",
		"opencolorio":  "Check this when you need OpenColorIO color transforms for VFX, animation, or color-managed production workflows. It is not default because ordinary conversions rarely need full OCIO pipelines.",
		"lcms2":        "Check this when you need ICC profile color-management support. It is not default because ordinary conversions often do not require explicit ICC profile handling.",
		"gme":          "Check this when you need to read game music formats such as NSF, SPC, GYM, or other chiptune files. It is not default because game music formats are a niche use case.",
		"openmpt":      "Check this when you need to play tracker/module audio files (MOD, XM, S3M, IT). Prefer this over libmodplug for newer files. It is not default because module audio is a niche format.",
	}
	if reason, ok := libraryReasons[libraryId]; ok {
		return reason
	}
	switch licenseEffectName {
	case "gpl":
		return "Check this only when you need this specific feature. It is not default because it changes the resulting FFmpeg build to GPL."
	case "nonfree":
		return "Check this only for local/private builds that need this specific feature. It is not default because it makes the resulting FFmpeg build nonfree and limits redistribution."
	default:
		return "Check this only when you need this specific feature. It is not default because it adds another external package to the build."
	}
}

func includedLibraryChoice(id string, displayName string, categoryName string, plainExplanation string, technicalExplanation string) LibraryChoice {
	return LibraryChoice{LibraryId: id, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: []string{}, PackageNames: []string{}, LicenseEffectName: "included", ReviewNote: plainExplanation, PlainExplanation: plainExplanation, TechnicalExplanation: technicalExplanation, DefaultChecked: true, Locked: true}
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
