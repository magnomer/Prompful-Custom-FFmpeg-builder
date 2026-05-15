package planning

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"customffmpegbuilder/internal/scripting"
)

func DefaultBuildToolSettings() BuildToolSettings {
	return BuildToolSettings{
		WorkspaceDirectory:       filepath.Join(defaultUserDataDirectory(), "CustomFFmpegBuilder", "workspace"),
		Msys2ArchiveUrl:          "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst",
		Msys2ArchiveSha256Hash:   "",
		Msys2ArchiveSignatureUrl: "https://repo.msys2.org/distrib/msys2-x86_64-latest.tar.zst.sig",
		Msys2PackageNames:        defaultMsys2PackageNames(),
		WindowsShellProfileName:  "ucrt64",
	}
}

func DefaultFfmpegBuildSettings() FfmpegBuildSettings {
	return FfmpegBuildSettings{
		WorkspaceDirectory:         filepath.Join(defaultUserDataDirectory(), "CustomFFmpegBuilder", "workspace"),
		FfmpegSourceArchiveUrl:     "",
		FfmpegSourceSignatureUrl:   "",
		FfmpegSourceSha256Hash:     "",
		SelectedLibraryIds:         defaultLibraryIds(),
		SelectedConfigureOptionIds: defaultConfigureOptionIds(),
		ExtraConfigureFlags:        []string{},
		ConfigureFlags:             []string{},
		ParallelJobCount:           maxInt(1, runtime.NumCPU()-1),
		WindowsShellProfileName:    "ucrt64",
		LicenseProfileName:         "lgpl-local",
	}
}

func LibraryCatalogForShellProfile(windowsShellProfileName string) []LibraryChoice {
	packagePrefix := packagePrefixForShellProfile(windowsShellProfileName)
	return []LibraryChoice{
		includedLibraryChoice("ffmpeg-program", "ffmpeg.exe", "Included by default (official FFmpeg source)", "The main command-line program users normally run.", "Built by default in a normal FFmpeg source build. No external package or --enable-lib flag is needed. Only disabled by selecting --disable-programs."),
		includedLibraryChoice("ffprobe-program", "ffprobe.exe", "Included by default (official FFmpeg source)", "The media inspection tool used to check files and streams.", "Built by default in a normal FFmpeg source build. No external package or --enable-lib flag is needed. Disabled by selecting --disable-programs or --disable-ffprobe."),
		includedLibraryChoice("libavcodec", "libavcodec", "Included by default (official FFmpeg source)", "FFmpeg's built-in codec library for encoding and decoding media.", "This is one of FFmpeg's own libraries. Native decoders and encoders are enabled by default; external-library codecs still require explicit --enable-lib... flags."),
		includedLibraryChoice("libavformat", "libavformat", "Included by default (official FFmpeg source)", "FFmpeg's built-in library for reading and writing media containers such as MP4, MOV, MKV, and WAV.", "This is one of FFmpeg's own libraries and does not require an external MSYS2 package."),
		includedLibraryChoice("libavfilter", "libavfilter", "Included by default (official FFmpeg source)", "FFmpeg's built-in filtering library for scaling, trimming, overlaying, subtitles, audio filters, and more.", "This is one of FFmpeg's own libraries. Some filters can become more capable when external libraries are selected."),
		includedLibraryChoice("libavutil", "libavutil", "Included by default (official FFmpeg source)", "Shared utility code used by the rest of FFmpeg.", "This is one of FFmpeg's own libraries and is required by normal FFmpeg builds."),
		includedLibraryChoice("libswscale", "libswscale", "Included by default (official FFmpeg source)", "Built-in image scaling and pixel-format conversion.", "This is one of FFmpeg's own libraries and does not require an external package."),
		includedLibraryChoice("libswresample", "libswresample", "Included by default (official FFmpeg source)", "Built-in audio resampling and sample-format conversion.", "This is one of FFmpeg's own libraries and does not require an external package."),
		includedLibraryChoice("native-codecs", "Native FFmpeg codecs", "Included by default (official FFmpeg source)", "FFmpeg includes many native decoders and encoders before you add external codec libraries.", "Native FFmpeg decoders and encoders are enabled in a normal source build. External codec libraries are separate choices and use their own --enable-lib... flags."),
		includedLibraryChoice("native-formats", "Native formats and muxers", "Included by default (official FFmpeg source)", "FFmpeg includes many built-in readers and writers for media containers.", "These are part of FFmpeg itself. External libraries can add support or improve specific formats, but the base format layer is not an external package."),
		libraryChoice("x264", "x264", "Video encoders", []string{"--enable-libx264"}, []string{packagePrefix + "-x264"}, "gpl", "Adds H.264 encoding. Good compatibility, but it changes the build to GPL."),
		libraryChoice("x265", "x265", "Video encoders", []string{"--enable-libx265"}, []string{packagePrefix + "-x265"}, "gpl", "Adds HEVC/H.265 encoding. Smaller files than H.264, but GPL."),
		libraryChoice("libvpx", "libvpx", "Video encoders", []string{"--enable-libvpx"}, []string{packagePrefix + "-libvpx"}, "lgpl", "Adds VP8/VP9 support for WebM files."),
		libraryChoice("aom", "AOM AV1", "Video encoders", []string{"--enable-libaom"}, []string{packagePrefix + "-aom"}, "lgpl", "Adds AV1 encoding/decoding through libaom."),
		libraryChoice("svt-av1", "SVT-AV1", "Video encoders", []string{"--enable-libsvtav1"}, []string{packagePrefix + "-svt-av1"}, "lgpl", "Adds a fast AV1 encoder."),
		libraryChoice("rav1e", "rav1e", "Video encoders", []string{"--enable-librav1e"}, []string{packagePrefix + "-rav1e"}, "lgpl", "Adds the rav1e AV1 encoder."),
		libraryChoice("openh264", "OpenH264", "Video encoders", []string{"--enable-libopenh264"}, []string{packagePrefix + "-openh264"}, "lgpl", "Adds Cisco OpenH264 support."),
		libraryChoice("xavs2", "xavs2", "Video encoders", []string{"--enable-libxavs2"}, []string{packagePrefix + "-xavs2"}, "gpl", "Adds AVS2 video encoding. GPL effect."),
		libraryChoice("dav1d", "dav1d", "Video decoders", []string{"--enable-libdav1d"}, []string{packagePrefix + "-dav1d"}, "lgpl", "Adds a fast AV1 decoder."),
		libraryChoice("libjxl", "JPEG XL", "Image codecs", []string{"--enable-libjxl"}, []string{packagePrefix + "-libjxl"}, "lgpl", "Adds JPEG XL image support."),
		libraryChoice("openjpeg", "OpenJPEG", "Image codecs", []string{"--enable-libopenjpeg"}, []string{packagePrefix + "-openjpeg2"}, "lgpl", "Adds JPEG 2000 support."),
		libraryChoice("webp", "WebP", "Image codecs", []string{"--enable-libwebp"}, []string{packagePrefix + "-libwebp"}, "lgpl", "Adds WebP image support."),
		libraryChoice("png", "libpng", "Image codecs", []string{}, []string{packagePrefix + "-libpng"}, "lgpl", "PNG support is usually native; this installs libpng for filters/tools that need it."),
		libraryChoice("zimg", "zimg", "Filters and processing", []string{"--enable-libzimg"}, []string{packagePrefix + "-zimg"}, "lgpl", "Adds high-quality resizing, colorspace, and bit-depth conversion filters."),
		libraryChoice("libplacebo", "libplacebo", "Filters and processing", []string{"--enable-libplacebo", "--enable-vulkan"}, []string{packagePrefix + "-libplacebo", packagePrefix + "-vulkan-loader", packagePrefix + "-vulkan-headers"}, "lgpl", "Adds GPU-oriented video processing filters."),
		libraryChoice("vmaf", "libvmaf", "Filters and processing", []string{"--enable-libvmaf"}, []string{packagePrefix + "-vmaf"}, "lgpl", "Adds Netflix VMAF video quality measurement filters."),
		libraryChoice("frei0r", "frei0r", "Filters and processing", []string{"--enable-frei0r"}, []string{packagePrefix + "-frei0r-plugins"}, "gpl", "Adds frei0r video effects. GPL effect."),
		libraryChoice("rubberband", "Rubber Band", "Audio", []string{"--enable-librubberband"}, []string{packagePrefix + "-rubberband"}, "gpl", "Adds high quality audio time-stretching/pitch-shifting. GPL effect."),
		libraryChoice("opus", "Opus", "Audio", []string{"--enable-libopus"}, []string{packagePrefix + "-opus"}, "lgpl", "Adds Opus audio encoding/decoding."),
		libraryChoice("vorbis", "Vorbis", "Audio", []string{"--enable-libvorbis"}, []string{packagePrefix + "-libvorbis"}, "lgpl", "Adds Vorbis audio encoding."),
		libraryChoice("mp3lame", "LAME MP3", "Audio", []string{"--enable-libmp3lame"}, []string{packagePrefix + "-lame"}, "lgpl", "Adds MP3 encoding."),
		libraryChoice("twolame", "TwoLAME", "Audio", []string{"--enable-libtwolame"}, []string{packagePrefix + "-twolame"}, "lgpl", "Adds MP2 audio encoding."),
		libraryChoice("soxr", "SoX Resampler", "Audio", []string{"--enable-libsoxr"}, []string{packagePrefix + "-libsoxr"}, "lgpl", "Adds high-quality audio resampling."),
		libraryChoice("speex", "Speex", "Audio", []string{"--enable-libspeex"}, []string{packagePrefix + "-speex"}, "lgpl", "Adds Speex speech codec support."),
		libraryChoice("gsm", "GSM", "Audio", []string{"--enable-libgsm"}, []string{packagePrefix + "-gsm"}, "lgpl", "Adds GSM audio codec support."),
		libraryChoice("ilbc", "iLBC", "Audio", []string{"--enable-libilbc"}, []string{packagePrefix + "-libilbc"}, "lgpl", "Adds iLBC speech codec support."),
		libraryChoice("opencore-amr", "OpenCORE AMR", "Audio", []string{"--enable-libopencore-amrnb", "--enable-libopencore-amrwb"}, []string{packagePrefix + "-opencore-amr"}, "lgpl", "Adds AMR-NB and AMR-WB decoding support for mobile-phone voice recordings. This requires FFmpeg's version-3 license switch."),
		libraryChoice("vo-amrwbenc", "VisualOn AMR-WB encoder", "Audio", []string{"--enable-libvo-amrwbenc"}, []string{packagePrefix + "-vo-amrwbenc"}, "lgpl", "Adds AMR-WB encoding support for narrow telephony workflows. This requires FFmpeg's version-3 license switch."),
		libraryChoice("fdk-aac", "Fraunhofer FDK AAC", "Audio", []string{"--enable-libfdk-aac", "--enable-nonfree"}, []string{packagePrefix + "-fdk-aac"}, "nonfree", "Adds a high-quality AAC encoder. Makes the FFmpeg build nonfree."),
		libraryChoice("freetype", "FreeType", "Subtitles and text", []string{"--enable-libfreetype"}, []string{packagePrefix + "-freetype"}, "lgpl", "Adds font rendering for subtitles and text filters."),
		libraryChoice("fontconfig", "Fontconfig", "Subtitles and text", []string{"--enable-fontconfig"}, []string{packagePrefix + "-fontconfig"}, "lgpl", "Adds font discovery support."),
		libraryChoice("fribidi", "FriBidi", "Subtitles and text", []string{"--enable-libfribidi"}, []string{packagePrefix + "-fribidi"}, "lgpl", "Adds bidirectional text support."),
		libraryChoice("harfbuzz", "HarfBuzz", "Subtitles and text", []string{"--enable-libharfbuzz"}, []string{packagePrefix + "-harfbuzz"}, "lgpl", "Adds advanced text shaping support."),
		libraryChoice("ass", "libass", "Subtitles and text", []string{"--enable-libass"}, []string{packagePrefix + "-libass"}, "lgpl", "Adds advanced subtitle rendering."),
		libraryChoice("bluray", "libbluray", "Disc and device input", []string{"--enable-libbluray"}, []string{packagePrefix + "-libbluray"}, "lgpl", "Adds Blu-ray reading support."),
		libraryChoice("cdio", "libcdio", "Disc and device input", []string{"--enable-libcdio"}, []string{packagePrefix + "-libcdio"}, "gpl", "Adds CD input support. GPL effect."),
		libraryChoice("modplug", "libmodplug", "Disc and device input", []string{"--enable-libmodplug"}, []string{packagePrefix + "-libmodplug"}, "lgpl", "Adds tracker/module audio file support."),
		libraryChoice("openal", "OpenAL", "Disc and device input", []string{"--enable-openal"}, []string{packagePrefix + "-openal"}, "lgpl", "Adds OpenAL audio input support."),
		libraryChoice("sdl2", "SDL2", "Disc and device input", []string{"--enable-sdl2"}, []string{packagePrefix + "-SDL2"}, "lgpl", "Adds SDL2 support, mainly useful for ffplay."),
		libraryChoice("openssl", "OpenSSL", "Network", []string{"--enable-openssl"}, []string{packagePrefix + "-openssl"}, "nonfree", "Adds HTTPS/TLS network support through OpenSSL. Choose this instead of GnuTLS when you specifically need OpenSSL-compatible TLS behavior."),
		libraryChoice("gnutls", "GnuTLS", "Network", []string{"--enable-gnutls"}, []string{packagePrefix + "-gnutls"}, "lgpl", "Adds HTTPS/TLS network support through GnuTLS. Choose this instead of OpenSSL when you want TLS support without OpenSSL's redistribution concerns."),
		libraryChoice("srt", "SRT", "Network", []string{"--enable-libsrt"}, []string{packagePrefix + "-srt"}, "lgpl", "Adds Secure Reliable Transport protocol support."),
		libraryChoice("ssh", "libssh", "Network", []string{"--enable-libssh"}, []string{packagePrefix + "-libssh"}, "lgpl", "Adds SSH protocol support."),
		libraryChoice("zmq", "ZeroMQ", "Network", []string{"--enable-libzmq"}, []string{packagePrefix + "-zeromq"}, "lgpl", "Adds ZeroMQ messaging support."),
		libraryChoice("rist", "librist", "Network", []string{"--enable-librist"}, []string{packagePrefix + "-librist"}, "lgpl", "Adds RIST streaming protocol support."),
		libraryChoice("xml2", "libxml2", "Network", []string{"--enable-libxml2"}, []string{packagePrefix + "-libxml2"}, "lgpl", "Adds XML parsing support used by some formats/protocols."),
		libraryChoice("tesseract", "Tesseract OCR", "OCR", []string{"--enable-libtesseract"}, []string{packagePrefix + "-tesseract-ocr"}, "lgpl", "Adds OCR filter support through Tesseract."),
	}
}

func ConfigureOptionCatalog() []ConfigureOptionChoice {
	return []ConfigureOptionChoice{
		configureOptionChoice("default-static", "Build static libraries", "Default FFmpeg source build", []string{}, "FFmpeg normally builds static libraries from source.", "Checked because this is normal FFmpeg configure behavior. No extra flag is needed unless you choose a different output type.", true, true),
		configureOptionChoice("default-programs", "Build command-line programs", "Default FFmpeg source build", []string{}, "FFmpeg normally builds command-line programs such as ffmpeg.exe and ffprobe.exe.", "Checked because programs are built in a normal source build. Disable only if you want libraries without command-line tools.", true, true),
		configureOptionChoice("default-ffmpeg", "Build ffmpeg.exe", "Default FFmpeg source build", []string{}, "Builds the main command-line converter most users run.", "Checked because ffmpeg.exe is part of a normal program build.", true, true),
		configureOptionChoice("default-ffprobe", "Build ffprobe.exe", "Default FFmpeg source build", []string{}, "Builds the media inspection tool used to read stream and container information.", "Checked because ffprobe.exe is part of a normal program build.", true, true),

		configureOptionChoice("disable-doc", "Skip documentation files", "Optional changes", []string{"--disable-doc"}, "Makes the build smaller by not building local documentation files.", "Not a normal source default. Select this when you only need binaries/libraries and do not need generated docs.", false, false),
		configureOptionChoice("disable-debug", "Remove debug build data", "Optional changes", []string{"--disable-debug"}, "Makes normal-use output smaller and simpler.", "Not a normal source default. Select this for ordinary release-style local builds; leave unchecked when investigating build or runtime problems.", false, false),
		configureOptionChoice("enable-shared", "Build shared DLL libraries", "Output type", []string{"--enable-shared", "--disable-static"}, "Creates DLL-style FFmpeg libraries for other programs to load.", "FFmpeg platform documentation describes --enable-shared as the way to build FFmpeg libraries as DLLs on Windows. This changes the output type, so it is not selected by default here.", false, false),
		configureOptionChoice("disable-programs", "Do not build command-line programs", "Programs", []string{"--disable-programs"}, "Builds FFmpeg libraries only, without ffmpeg.exe, ffprobe.exe, or ffplay.exe.", "Select only when another application will link to the libraries and you do not need command-line tools.", false, false),
		configureOptionChoice("disable-ffplay", "Do not build ffplay", "Programs", []string{"--disable-ffplay"}, "Skips the simple playback test program.", "Useful when SDL playback support is unnecessary. ffmpeg.exe and ffprobe.exe are unaffected.", false, false),
		configureOptionChoice("disable-ffprobe", "Do not build ffprobe", "Programs", []string{"--disable-ffprobe"}, "Skips the media inspection tool.", "Leave unchecked if you want ffprobe.exe for checking streams, metadata, and containers.", false, false),
		configureOptionChoice("disable-autodetect", "Do not auto-use hidden system libraries", "Security and reproducibility", []string{"--disable-autodetect"}, "Makes the build less surprising by using only explicitly selected external libraries.", "Good for transparent/reproducible builds. Select this when you want the Review page to explain every external dependency.", false, false),
		configureOptionChoice("enable-version3", "Allow LGPL/GPL version 3 code", "License", []string{"--enable-version3"}, "Allows components that move the build to version 3 license terms.", "Select only when a chosen library or feature requires version 3 terms.", false, false),
		configureOptionChoice("disable-asm", "Disable assembly optimizations", "Compatibility", []string{"--disable-asm"}, "Uses slower but simpler C code paths if assembly causes build problems.", "Normally leave unchecked because FFmpeg is faster with assembly optimizations.", false, false),
		configureOptionChoice("disable-x86asm", "Disable x86 assembly", "Compatibility", []string{"--disable-x86asm"}, "Try this when NASM/YASM-related build problems occur.", "Normally leave unchecked for performance.", false, false),
		configureOptionChoice("enable-small", "Prefer smaller binary size", "Size and speed", []string{"--enable-small"}, "Asks FFmpeg to prefer smaller output files over speed.", "Useful for constrained environments; may reduce performance.", false, false),
		configureOptionChoice("disable-stripping", "Keep symbol information", "Debugging", []string{"--disable-stripping"}, "Keeps more build information for debugging.", "Usually unnecessary for ordinary users, but useful when diagnosing crashes or build problems.", false, false),
	}
}

func defaultConfigureOptionIds() []string {
	return []string{"default-static", "default-programs", "default-ffmpeg", "default-ffprobe"}
}

func PlanToolchainSetup(buildToolSettings BuildToolSettings) (ToolchainPreparationPlan, error) {
	buildToolSettings = cleanBuildToolSettings(buildToolSettings)
	warnings := validateCommonWindowsWorkspace(buildToolSettings.WorkspaceDirectory)
	isExecutable := !hasBlockedWarnings(warnings)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "This project profile is Windows-only."})
		isExecutable = false
	}
	if buildToolSettings.Msys2ArchiveUrl == "" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "MSYS2 archive URL is empty. Use an official MSYS2 tar archive URL before approval. .tar.zst is recommended, and .tar.xz is accepted as a fallback."})
		isExecutable = false
	} else if strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".sig") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".exe") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".sfx.exe") {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Use an MSYS2 tar archive URL here. The official .exe installer is valid MSYS2, but this app does not run installers; it verifies and extracts tar archives inside the selected workspace. Use .tar.zst when possible, or .tar.xz as a fallback. Put the matching .sig URL in the signature field."})
		isExecutable = false
	} else if !(strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".tar.zst") || strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveUrl), ".tar.xz")) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "MSYS2 archive URL must end with .tar.zst or .tar.xz. This app uses tar archives so it can verify and extract files itself without running an installer."})
		isExecutable = false
	}
	if buildToolSettings.Msys2ArchiveSignatureUrl == "" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelWarning, Message: "No MSYS2 .sig URL was supplied. The app can calculate SHA-256, but signature verification is the better official authenticity check."})
	} else if !strings.HasSuffix(strings.ToLower(buildToolSettings.Msys2ArchiveSignatureUrl), ".sig") {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "MSYS2 signature URL must end with .sig."})
		isExecutable = false
	} else if buildToolSettings.Msys2ArchiveUrl != "" && buildToolSettings.Msys2ArchiveSignatureUrl != buildToolSettings.Msys2ArchiveUrl+".sig" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelWarning, Message: "MSYS2 signature URL does not exactly match the archive URL plus .sig. This may be intentional, but usually the signature URL should be the archive URL followed by .sig."})
	}
	if buildToolSettings.Msys2ArchiveSha256Hash != "" && !isSha256Hex(buildToolSettings.Msys2ArchiveSha256Hash) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "MSYS2 SHA-256 must be exactly 64 hexadecimal characters. If you pasted a .sig file, remove it; .sig is a signature file, not a hash."})
		isExecutable = false
	}
	for _, packageName := range buildToolSettings.Msys2PackageNames {
		if err := scripting.ValidateMsys2PackageName(packageName); err != nil {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: err.Error()})
			isExecutable = false
		}
	}
	if !isSupportedWindowsShellProfileName(buildToolSettings.WindowsShellProfileName) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Windows shell profile must be ucrt64, mingw64, or clang64."})
		isExecutable = false
	}

	plan := ToolchainPreparationPlan{
		ActionName:                 "prepare-private-msys2-toolchain",
		WorkspaceDirectory:         buildToolSettings.WorkspaceDirectory,
		Msys2RootDirectory:         filepath.Join(buildToolSettings.WorkspaceDirectory, "toolchains", "msys2"),
		Msys2ArchiveUrl:            buildToolSettings.Msys2ArchiveUrl,
		Msys2ArchiveSha256Hash:     buildToolSettings.Msys2ArchiveSha256Hash,
		Msys2ArchiveSignatureUrl:   buildToolSettings.Msys2ArchiveSignatureUrl,
		Msys2PackageNames:          buildToolSettings.Msys2PackageNames,
		WindowsShellProfileName:    buildToolSettings.WindowsShellProfileName,
		WillModifySystemPath:       false,
		WillRequireAdminRights:     false,
		WillUseExistingMsys2:       false,
		WillDeleteFiles:            false,
		DownloadConflictPolicyName: "reuse-if-hash-matches",
		ExtractDestinationPolicy:   "must-not-exist",
		Operations: []PlanOperation{
			{OperationName: "create-workspace-directories", Summary: "Create directories inside the selected workspace only."},
			{OperationName: "download-msys2-archive", Summary: "Download the approved MSYS2 archive from the approved URL."},
			{OperationName: "verify-msys2-signature", Summary: "Verify the downloaded MSYS2 archive with its official .sig file using the built-in verifier."},
			{OperationName: "record-msys2-sha256", Summary: "Calculate and log the archive SHA-256 for the audit record."},
			{OperationName: "extract-private-msys2", Summary: "Extract MSYS2 into the private workspace toolchain directory."},
			{OperationName: "install-approved-pacman-packages", Summary: "Install only the package names listed in this plan."},
		},
		Warnings:     warnings,
		IsExecutable: isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := HashPlan(planWithoutHash)
	if err != nil {
		return ToolchainPreparationPlan{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func PlanFfmpegBuild(ffmpegBuildSettings FfmpegBuildSettings) (FfmpegBuildPlan, error) {
	ffmpegBuildSettings = cleanFfmpegBuildSettings(ffmpegBuildSettings)
	warnings := validateCommonWindowsWorkspace(ffmpegBuildSettings.WorkspaceDirectory)
	isExecutable := !hasBlockedWarnings(warnings)
	selectedLibraries, unknownLibraryIds := selectLibraries(ffmpegBuildSettings.WindowsShellProfileName, ffmpegBuildSettings.SelectedLibraryIds)
	for _, unknownLibraryId := range unknownLibraryIds {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Unknown library id: " + unknownLibraryId})
		isExecutable = false
	}
	libraryPackages := uniquePackagesFromLibraries(selectedLibraries)
	libraryFlags := uniqueFlagsFromLibraries(selectedLibraries)
	selectedConfigureOptions, unknownConfigureOptionIds := selectConfigureOptions(ffmpegBuildSettings.SelectedConfigureOptionIds)
	for _, unknownConfigureOptionId := range unknownConfigureOptionIds {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Unknown FFmpeg option id: " + unknownConfigureOptionId})
		isExecutable = false
	}
	optionFlags := uniqueFlagsFromConfigureOptions(selectedConfigureOptions)
	finalConfigureFlags := mergeUniqueStrings(libraryFlags, optionFlags)
	finalConfigureFlags = mergeUniqueStrings(finalConfigureFlags, ffmpegBuildSettings.ExtraConfigureFlags)
	// ExtraConfigureFlags may include --enable-lib* flags that match catalog
	// library entries whose packages were not included via checkbox selection.
	// Resolve those flags back to their catalog entries and add missing packages
	// so that pacman installs them before configure runs.
	extraLibraries := librariesForConfigureFlags(ffmpegBuildSettings.WindowsShellProfileName, ffmpegBuildSettings.ExtraConfigureFlags, selectedLibraries)
	libraryPackages = mergeUniqueStrings(libraryPackages, uniquePackagesFromLibraries(extraLibraries))
	derivedLicenseProfileName := deriveLicenseProfileName(selectedLibraries, finalConfigureFlags)
	finalConfigureFlags = addLicenseFlags(finalConfigureFlags, derivedLicenseProfileName, selectedLibraries)

	if runtime.GOOS != "windows" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "This project profile is Windows-only."})
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceArchiveUrl == "" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "FFmpeg source archive URL is empty. Paste an official fixed release archive URL before approval."})
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSignatureUrl == "" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "FFmpeg source signature URL is empty. FFmpeg releases are verified through the matching .asc PGP signature."})
		isExecutable = false
	} else if !strings.HasSuffix(strings.ToLower(ffmpegBuildSettings.FfmpegSourceSignatureUrl), ".asc") {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "FFmpeg signature URL must end in .asc. Do not paste the PGP signature text; use the URL of the matching .asc file."})
		isExecutable = false
	}
	if ffmpegBuildSettings.FfmpegSourceSha256Hash == "" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelInfo, Message: "No FFmpeg SHA-256 was supplied. This is normal for the official release page: the app will verify the .asc PGP signature and calculate SHA-256 for the log."})
	} else if !isSha256Hex(ffmpegBuildSettings.FfmpegSourceSha256Hash) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "FFmpeg SHA-256 must be exactly 64 hexadecimal characters. If you have a .asc or .sig file, do not paste it into this field; it is a signature file, not a hash."})
		isExecutable = false
	}
	if len(finalConfigureFlags) == 0 {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelWarning, Message: "No configure flags were selected."})
	}
	for _, configureFlag := range finalConfigureFlags {
		if err := scripting.ValidateConfigureFlag(configureFlag); err != nil {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: err.Error()})
			isExecutable = false
		}
	}
	for _, packageName := range libraryPackages {
		if err := scripting.ValidateMsys2PackageName(packageName); err != nil {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: err.Error()})
			isExecutable = false
		}
	}
	configureConflictWarnings, hasConfigureConflicts := validateConfigureFlagConflicts(finalConfigureFlags)
	warnings = append(warnings, configureConflictWarnings...)
	if hasConfigureConflicts {
		isExecutable = false
	}
	if !isSupportedWindowsShellProfileName(ffmpegBuildSettings.WindowsShellProfileName) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Windows shell profile must be ucrt64, mingw64, or clang64."})
		isExecutable = false
	}
	if ffmpegBuildSettings.ParallelJobCount > 256 {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Parallel job count must not exceed 256."})
		isExecutable = false
	}
	licenseWarnings, licenseBlocked := validateLicenseProfile(derivedLicenseProfileName, selectedLibraries, finalConfigureFlags)
	warnings = append(warnings, licenseWarnings...)
	if selectedLibrariesRequireVersion3(selectedLibraries) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelInfo, Message: "FFmpeg version-3 license switch was added because a selected AMR library requires --enable-version3."})
	}
	if licenseBlocked {
		isExecutable = false
	}

	plan := FfmpegBuildPlan{
		ActionName:                 "build-ffmpeg-from-approved-source",
		WorkspaceDirectory:         ffmpegBuildSettings.WorkspaceDirectory,
		Msys2RootDirectory:         filepath.Join(ffmpegBuildSettings.WorkspaceDirectory, "toolchains", "msys2"),
		FfmpegSourceArchiveUrl:     ffmpegBuildSettings.FfmpegSourceArchiveUrl,
		FfmpegSourceSignatureUrl:   ffmpegBuildSettings.FfmpegSourceSignatureUrl,
		FfmpegSourceSha256Hash:     ffmpegBuildSettings.FfmpegSourceSha256Hash,
		SelectedLibraryIds:         ffmpegBuildSettings.SelectedLibraryIds,
		SelectedLibraries:          selectedLibraries,
		RequiredMsys2PackageNames:  libraryPackages,
		GeneratedConfigureFlags:    libraryFlags,
		SelectedConfigureOptions:   selectedConfigureOptions,
		GeneratedOptionFlags:       optionFlags,
		ExtraConfigureFlags:        ffmpegBuildSettings.ExtraConfigureFlags,
		ConfigureFlags:             finalConfigureFlags,
		ParallelJobCount:           ffmpegBuildSettings.ParallelJobCount,
		WindowsShellProfileName:    ffmpegBuildSettings.WindowsShellProfileName,
		LicenseProfileName:         derivedLicenseProfileName,
		WillModifySystemPath:       false,
		WillRequireAdminRights:     false,
		WillUseExistingMsys2:       false,
		WillDeleteFiles:            false,
		DownloadConflictPolicyName: "reuse-if-hash-matches",
		ExtractDestinationPolicy:   "must-not-exist",
		Operations: []PlanOperation{
			{OperationName: "download-ffmpeg-source", Summary: "Download the approved FFmpeg source archive."},
			{OperationName: "verify-ffmpeg-source-signature", Summary: "Verify the FFmpeg source archive with the matching .asc PGP signature before extraction."},
			{OperationName: "extract-ffmpeg-source", Summary: "Extract source into the private workspace."},
			{OperationName: "review-selected-libraries", Summary: "Show selected FFmpeg libraries, generated package names, generated configure flags, and license effects."},
			{OperationName: "install-selected-library-packages", Summary: "Install only the MSYS2 packages required by the selected FFmpeg libraries before configure runs."},
			{OperationName: "run-approved-configure-script", Summary: "Run FFmpeg configure with exactly the approved final flags."},
			{OperationName: "run-approved-make-command", Summary: "Run make with the approved parallel job count."},
			{OperationName: "create-artifact-report", Summary: "Write a build report with source hashes, libraries, flags, and artifact paths."},
		},
		Warnings:     warnings,
		IsExecutable: isExecutable,
	}

	planWithoutHash := plan
	planWithoutHash.PlanHash = ""
	planHash, err := HashPlan(planWithoutHash)
	if err != nil {
		return FfmpegBuildPlan{}, err
	}
	plan.PlanHash = planHash
	return plan, nil
}

func CheckPlanCanRun(isExecutable bool) error {
	if !isExecutable {
		return errors.New("plan is blocked and cannot be executed")
	}
	return nil
}

func cleanBuildToolSettings(buildToolSettings BuildToolSettings) BuildToolSettings {
	defaults := DefaultBuildToolSettings()
	if buildToolSettings.WorkspaceDirectory == "" {
		buildToolSettings.WorkspaceDirectory = defaults.WorkspaceDirectory
	}
	if buildToolSettings.Msys2ArchiveSignatureUrl == "" && buildToolSettings.Msys2ArchiveUrl != "" {
		buildToolSettings.Msys2ArchiveSignatureUrl = buildToolSettings.Msys2ArchiveUrl + ".sig"
	}
	if buildToolSettings.WindowsShellProfileName == "" {
		buildToolSettings.WindowsShellProfileName = defaults.WindowsShellProfileName
	}
	if len(buildToolSettings.Msys2PackageNames) == 0 {
		buildToolSettings.Msys2PackageNames = defaults.Msys2PackageNames
	}
	return buildToolSettings
}

func cleanFfmpegBuildSettings(ffmpegBuildSettings FfmpegBuildSettings) FfmpegBuildSettings {
	defaults := DefaultFfmpegBuildSettings()
	if ffmpegBuildSettings.WorkspaceDirectory == "" {
		ffmpegBuildSettings.WorkspaceDirectory = defaults.WorkspaceDirectory
	}
	if ffmpegBuildSettings.FfmpegSourceSignatureUrl == "" && ffmpegBuildSettings.FfmpegSourceArchiveUrl != "" {
		ffmpegBuildSettings.FfmpegSourceSignatureUrl = ffmpegBuildSettings.FfmpegSourceArchiveUrl + ".asc"
	}
	if ffmpegBuildSettings.WindowsShellProfileName == "" {
		ffmpegBuildSettings.WindowsShellProfileName = defaults.WindowsShellProfileName
	}
	if ffmpegBuildSettings.ParallelJobCount < 1 {
		ffmpegBuildSettings.ParallelJobCount = defaults.ParallelJobCount
	}
	ffmpegBuildSettings.SelectedLibraryIds = mergeDefaultLibraryIds(ffmpegBuildSettings.SelectedLibraryIds)
	if len(ffmpegBuildSettings.SelectedConfigureOptionIds) == 0 {
		ffmpegBuildSettings.SelectedConfigureOptionIds = defaults.SelectedConfigureOptionIds
	}
	if len(ffmpegBuildSettings.ExtraConfigureFlags) == 0 && len(ffmpegBuildSettings.ConfigureFlags) > 0 {
		ffmpegBuildSettings.ExtraConfigureFlags = ffmpegBuildSettings.ConfigureFlags
	}
	if ffmpegBuildSettings.LicenseProfileName == "" {
		ffmpegBuildSettings.LicenseProfileName = defaults.LicenseProfileName
	}
	return ffmpegBuildSettings
}

func validateCommonWindowsWorkspace(workspaceDirectory string) []PlanWarning {
	warnings := []PlanWarning{}
	if workspaceDirectory == "" {
		return append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Workspace directory is empty."})
	}
	if filepath.IsAbs(workspaceDirectory) == false {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Workspace directory must be an absolute path."})
	}
	if containsSpace(workspaceDirectory) {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelWarning, Message: "Workspace path contains a space. Some FFmpeg dependency builds may fail with spaces in paths."})
	}
	return warnings
}

func hasBlockedWarnings(planWarnings []PlanWarning) bool {
	for _, planWarning := range planWarnings {
		if planWarning.RiskLevel == RiskLevelBlocked {
			return true
		}
	}
	return false
}

func isSupportedWindowsShellProfileName(windowsShellProfileName string) bool {
	switch windowsShellProfileName {
	case "ucrt64", "mingw64", "clang64":
		return true
	default:
		return false
	}
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
		"xavs2":        "Check this when you need AVS2 video encoding. It is not default because AVS2 is uncommon outside specific regional or archival workflows and it changes the resulting build to GPL.",
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
		"modplug":      "Check this when you need tracker/module audio formats such as MOD/XM/S3M. It is not default because those formats are niche.",
		"openal":       "Check this when you need OpenAL audio input. It is not default because it is a device/input feature, not needed for normal file conversion.",
		"sdl2":         "Check this when you want ffplay or SDL-based playback support. It is not default because this builder is mainly for ffmpeg/ffprobe workflows, and playback UI support adds a separate dependency.",
		"openssl":      "Check this when FFmpeg must use HTTPS/TLS through OpenSSL. Do not select it together with GnuTLS: FFmpeg's configure script rejects both TLS backends at the same time. It is not default because it can make redistribution more sensitive and many local conversions do not need network TLS.",
		"gnutls":       "Check this when FFmpeg must use HTTPS/TLS through GnuTLS. Do not select it together with OpenSSL: FFmpeg's configure script rejects both TLS backends at the same time. It is not default because local file conversion does not need network TLS support.",
		"srt":          "Check this when you need Secure Reliable Transport streaming. It is not default because SRT is for live/remote streaming workflows, not ordinary local conversion.",
		"ssh":          "Check this when FFmpeg must read or write through SSH/SFTP-style protocols. It is not default because most users use local files or HTTPS instead.",
		"zmq":          "Check this when you need ZeroMQ filter/control messaging. It is not default because it is for automation/control workflows rather than normal conversion.",
		"rist":         "Check this when you need RIST broadcast/streaming transport. It is not default because it is a specialized live-video transport.",
		"xml2":         "Check this when a format, manifest, or protocol needs XML parsing. It is not default because common local conversion usually does not need XML support.",
		"tesseract":    "Check this when you need OCR from video/images. It is not default because OCR is a heavy, specialized feature and requires extra language data for useful results.",
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
	return ConfigureOptionChoice{OptionId: id, DisplayName: displayName, CategoryName: categoryName, ConfigureFlags: flags, PlainExplanation: plainExplanation, TechnicalNote: technicalNote, DefaultEnabled: defaultEnabled, Locked: locked}
}

func selectConfigureOptions(selectedOptionIds []string) ([]ConfigureOptionChoice, []string) {
	catalog := ConfigureOptionCatalog()
	catalogById := map[string]ConfigureOptionChoice{}
	for _, option := range catalog {
		catalogById[option.OptionId] = option
	}
	selectedOptions := []ConfigureOptionChoice{}
	unknownOptionIds := []string{}
	seen := map[string]bool{}
	for _, selectedOptionId := range selectedOptionIds {
		if selectedOptionId == "" || seen[selectedOptionId] {
			continue
		}
		seen[selectedOptionId] = true
		option, found := catalogById[selectedOptionId]
		if !found {
			unknownOptionIds = append(unknownOptionIds, selectedOptionId)
			continue
		}
		selectedOptions = append(selectedOptions, option)
	}
	return selectedOptions, unknownOptionIds
}

func uniqueFlagsFromConfigureOptions(options []ConfigureOptionChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, option := range options {
		for _, flag := range option.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
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

func selectLibraries(windowsShellProfileName string, selectedLibraryIds []string) ([]LibraryChoice, []string) {
	catalog := LibraryCatalogForShellProfile(windowsShellProfileName)
	catalogById := map[string]LibraryChoice{}
	for _, library := range catalog {
		catalogById[library.LibraryId] = library
	}
	selectedLibraries := []LibraryChoice{}
	unknownLibraryIds := []string{}
	seen := map[string]bool{}
	for _, selectedLibraryId := range selectedLibraryIds {
		if selectedLibraryId == "" || seen[selectedLibraryId] {
			continue
		}
		seen[selectedLibraryId] = true
		library, found := catalogById[selectedLibraryId]
		if !found {
			unknownLibraryIds = append(unknownLibraryIds, selectedLibraryId)
			continue
		}
		selectedLibraries = append(selectedLibraries, library)
	}
	return selectedLibraries, unknownLibraryIds
}

// librariesForConfigureFlags returns catalog library entries whose configure
// flags overlap with the given flag list, excluding any already in skip.
// It is used to resolve ExtraConfigureFlags back to their MSYS2 packages.
func librariesForConfigureFlags(windowsShellProfileName string, flags []string, skip []LibraryChoice) []LibraryChoice {
	flagSet := map[string]bool{}
	for _, f := range flags {
		flagSet[f] = true
	}
	skipIds := map[string]bool{}
	for _, lib := range skip {
		skipIds[lib.LibraryId] = true
	}
	result := []LibraryChoice{}
	seen := map[string]bool{}
	for _, lib := range LibraryCatalogForShellProfile(windowsShellProfileName) {
		if skipIds[lib.LibraryId] || seen[lib.LibraryId] {
			continue
		}
		for _, f := range lib.ConfigureFlags {
			if flagSet[f] {
				seen[lib.LibraryId] = true
				result = append(result, lib)
				break
			}
		}
	}
	return result
}

func uniquePackagesFromLibraries(libraries []LibraryChoice) []string {
	packages := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		for _, packageName := range library.PackageNames {
			if !seen[packageName] {
				packages = append(packages, packageName)
				seen[packageName] = true
			}
		}
	}
	sort.Strings(packages)
	return packages
}

func uniqueFlagsFromLibraries(libraries []LibraryChoice) []string {
	flags := []string{}
	seen := map[string]bool{}
	for _, library := range libraries {
		for _, flag := range library.ConfigureFlags {
			if !seen[flag] {
				flags = append(flags, flag)
				seen[flag] = true
			}
		}
	}
	return flags
}

func mergeUniqueStrings(first []string, second []string) []string {
	merged := []string{}
	seen := map[string]bool{}
	for _, value := range append(first, second...) {
		if value == "" || seen[value] {
			continue
		}
		merged = append(merged, value)
		seen[value] = true
	}
	return merged
}

func addLicenseFlags(configureFlags []string, licenseProfileName string, libraries []LibraryChoice) []string {
	needsGpl := false
	needsNonfree := licenseProfileName == "nonfree-local"
	needsVersion3 := false
	for _, library := range libraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
		if libraryRequiresVersion3(library.LibraryId) {
			needsVersion3 = true
		}
	}
	if licenseProfileName == "gpl-local" {
		needsGpl = true
	}
	if needsGpl {
		configureFlags = mergeUniqueStrings([]string{"--enable-gpl"}, configureFlags)
	}
	if needsVersion3 {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-version3"})
	}
	if needsNonfree {
		configureFlags = mergeUniqueStrings(configureFlags, []string{"--enable-nonfree"})
	}
	return configureFlags
}

func libraryRequiresVersion3(libraryId string) bool {
	switch libraryId {
	case "opencore-amr", "vo-amrwbenc":
		return true
	default:
		return false
	}
}

func deriveLicenseProfileName(selectedLibraries []LibraryChoice, configureFlags []string) string {
	needsGpl := false
	needsNonfree := false
	for _, library := range selectedLibraries {
		switch library.LicenseEffectName {
		case "gpl":
			needsGpl = true
		case "nonfree":
			needsNonfree = true
		}
	}
	for _, configureFlag := range configureFlags {
		switch configureFlag {
		case "--enable-nonfree":
			needsNonfree = true
		case "--enable-gpl":
			needsGpl = true
		}
	}
	if needsNonfree {
		return "nonfree-local"
	}
	if needsGpl {
		return "gpl-local"
	}
	return "lgpl-local"
}

func selectedLibrariesRequireVersion3(selectedLibraries []LibraryChoice) bool {
	for _, library := range selectedLibraries {
		if libraryRequiresVersion3(library.LibraryId) {
			return true
		}
	}
	return false
}

func validateConfigureFlagConflicts(finalConfigureFlags []string) ([]PlanWarning, bool) {
	flagSet := map[string]bool{}
	for _, configureFlag := range finalConfigureFlags {
		flagSet[configureFlag] = true
	}
	warnings := []PlanWarning{}
	blocked := false
	if flagSet["--enable-gnutls"] && flagSet["--enable-openssl"] {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "Choose one TLS backend: OpenSSL or GnuTLS. FFmpeg cannot configure both --enable-openssl and --enable-gnutls at the same time."})
		blocked = true
	}
	return warnings, blocked
}

func validateLicenseProfile(licenseProfileName string, selectedLibraries []LibraryChoice, finalConfigureFlags []string) ([]PlanWarning, bool) {
	warnings := []PlanWarning{}
	blocked := false
	switch licenseProfileName {
	case "lgpl-local", "gpl-local", "nonfree-local":
	default:
		return []PlanWarning{{RiskLevel: RiskLevelBlocked, Message: "Derived license boundary must be lgpl-local, gpl-local, or nonfree-local."}}, true
	}
	if licenseProfileName == "gpl-local" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelInfo, Message: "License boundary was set to GPL local because selected libraries or flags require --enable-gpl."})
	}
	if licenseProfileName == "nonfree-local" {
		warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelWarning, Message: "License boundary was set to nonfree local because selected libraries or flags require --enable-nonfree. Do not redistribute this build unless you have reviewed the license obligations."})
	}
	for _, library := range selectedLibraries {
		if library.LicenseEffectName == "nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: fmt.Sprintf("Library %s requires nonfree-local license boundary.", library.DisplayName)})
			blocked = true
		}
		if library.LicenseEffectName == "gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: fmt.Sprintf("Library %s requires GPL-compatible license boundary.", library.DisplayName)})
			blocked = true
		}
	}
	for _, configureFlag := range finalConfigureFlags {
		if configureFlag == "--enable-nonfree" && licenseProfileName != "nonfree-local" {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "--enable-nonfree requires nonfree-local license boundary."})
			blocked = true
		}
		if configureFlag == "--enable-gpl" && licenseProfileName == "lgpl-local" {
			warnings = append(warnings, PlanWarning{RiskLevel: RiskLevelBlocked, Message: "--enable-gpl requires GPL-compatible license boundary."})
			blocked = true
		}
	}
	return warnings, blocked
}

func defaultUserDataDirectory() string {
	if localAppData := getenv("LOCALAPPDATA"); localAppData != "" {
		return localAppData
	}
	return "."
}

var getenv = os.Getenv

func containsSpace(value string) bool {
	for _, valueRune := range value {
		if valueRune == ' ' || valueRune == '\t' || valueRune == '\n' || valueRune == '\r' {
			return true
		}
	}
	return false
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}

func isSha256Hex(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, r := range value {
		if !((r >= '0' && r <= '9') || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')) {
			return false
		}
	}
	return true
}
