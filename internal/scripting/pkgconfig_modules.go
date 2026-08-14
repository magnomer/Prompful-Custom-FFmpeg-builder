package scripting

import (
	"promptfulcustomffmpegbuilder/internal/catalogfacts"
)

// LModulePkgconfig holds a pkg-config module name and an optional minimum version
// that must be satisfied before FFmpeg configure is attempted.
type LModulePkgconfig struct {
	Name       string
	MinVersion string // empty means no version check beyond existence
}

func LPackageModuleList(configureFlags []string, ffmpegVersion string) []LModulePkgconfig {
	// Only list libraries that are expected to provide a pkg-config module in MSYS2.
	// Some valid FFmpeg options, such as --enable-libgsm, are probed by FFmpeg
	// through headers/libraries instead of a .pc file. Pre-checking those with
	// pkg-config incorrectly blocks valid builds.
	//
	// name is the pkg-config module name (often different from the library catalog library id, e.g.
	// SvtAv1Enc vs svt-av1). LLibraryId is the library catalog id used to look up FFmpeg's pkg-config
	// minimum for the release being built, from the per-release support release-support manifest. The floor is
	// resolved per release (not a single hardcoded set), so an older FFmpeg gets its own lower
	// floor and an older package that satisfies it is no longer falsely rejected before
	// configure. A flag with no LLibraryId, an unsupported/snapshot version, or a library the
	// release pins no minimum for, carries no floor and is only checked for existence.
	type LScriptEntry struct {
		name       string
		LLibraryId string
	}
	moduleByFlag := map[string]LScriptEntry{
		"--enable-libaom":            {name: "aom", LLibraryId: "aom"},
		"--enable-libass":            {name: "libass", LLibraryId: "ass"},
		"--enable-libbluray":         {name: "libbluray", LLibraryId: "bluray"},
		"--enable-libcdio":           {name: "libcdio", LLibraryId: "cdio"},
		"--enable-libdav1d":          {name: "dav1d", LLibraryId: "dav1d"},
		"--enable-libdvdnav":         {name: "dvdnav", LLibraryId: "dvdnav"},
		"--enable-libkvazaar":        {name: "kvazaar", LLibraryId: "kvazaar"},
		"--enable-libonnxruntime":    {name: "libonnxruntime", LLibraryId: "onnxruntime"},
		"--enable-vapoursynth":       {name: "vapoursynth-script", LLibraryId: "vapoursynth"},
		"--enable-libfdk-aac":        {name: "fdk-aac", LLibraryId: "fdk-aac"},
		"--enable-libfontconfig":     {name: "fontconfig", LLibraryId: "fontconfig"},
		"--enable-fontconfig":        {name: "fontconfig", LLibraryId: "fontconfig"},
		"--enable-libfreetype":       {name: "freetype2", LLibraryId: "freetype"},
		"--enable-libfribidi":        {name: "fribidi", LLibraryId: "fribidi"},
		"--enable-libharfbuzz":       {name: "harfbuzz", LLibraryId: "harfbuzz"},
		"--enable-libilbc":           {name: "libilbc", LLibraryId: "ilbc"},
		"--enable-liblensfun":        {name: "lensfun", LLibraryId: "lensfun"},
		"--enable-libjxl":            {name: "libjxl", LLibraryId: "libjxl"},
		"--enable-libmodplug":        {name: "libmodplug", LLibraryId: "modplug"},
		"--enable-libmpeghdec":       {name: "mpeghdec", LLibraryId: "mpeghdec"},
		"--enable-libmp3lame":        {name: "lame", LLibraryId: "mp3lame"},
		"--enable-libopencore-amrnb": {name: "opencore-amrnb", LLibraryId: "opencore-amr"},
		"--enable-libopencore-amrwb": {name: "opencore-amrwb", LLibraryId: "opencore-amr"},
		"--enable-libopenh264":       {name: "openh264", LLibraryId: "openh264"},
		"--enable-libopenjpeg":       {name: "libopenjp2", LLibraryId: "openjpeg"},
		"--enable-libopus":           {name: "opus", LLibraryId: "opus"},
		"--enable-libplacebo":        {name: "libplacebo", LLibraryId: "libplacebo"},
		"--enable-librav1e":          {name: "rav1e", LLibraryId: "rav1e"},
		"--enable-librubberband":     {name: "rubberband", LLibraryId: "rubberband"},
		"--enable-libsoxr":           {name: "soxr", LLibraryId: "soxr"},
		"--enable-libspeex":          {name: "speex", LLibraryId: "speex"},
		"--enable-libssh":            {name: "libssh", LLibraryId: "ssh"},
		"--enable-libsvtav1":         {name: "SvtAv1Enc", LLibraryId: "svt-av1"},
		"--enable-libtesseract":      {name: "tesseract", LLibraryId: "tesseract"},
		"--enable-libtwolame":        {name: "twolame", LLibraryId: "twolame"},
		"--enable-libvmaf":           {name: "libvmaf", LLibraryId: "vmaf"},
		"--enable-libvorbis":         {name: "vorbis", LLibraryId: "vorbis"},
		"--enable-libvpx":            {name: "vpx", LLibraryId: "libvpx"},
		"--enable-libwebp":           {name: "libwebp", LLibraryId: "webp"},
		"--enable-libx264":           {name: "x264", LLibraryId: "x264"},
		"--enable-libx265":           {name: "x265", LLibraryId: "x265"},
		"--enable-libxavs2":          {name: "xavs2", LLibraryId: "xavs2"},
		"--enable-libzimg":           {name: "zimg", LLibraryId: "zimg"},
		"--enable-openal":            {name: "openal", LLibraryId: "openal"},
		"--enable-openssl":           {name: "openssl", LLibraryId: "openssl"},
		"--enable-gnutls":            {name: "gnutls", LLibraryId: "gnutls"},
		"--enable-sdl2":              {name: "sdl2", LLibraryId: "sdl2"},
		"--enable-chromaprint":       {name: "libchromaprint", LLibraryId: "chromaprint"},
		"--enable-libaribcaption":    {name: "libaribcaption", LLibraryId: "aribcaption"},
		"--enable-libbs2b":           {name: "libbs2b", LLibraryId: "bs2b"},
		"--enable-libcaca":           {name: "caca", LLibraryId: "caca"},
		"--enable-libdvdread":        {name: "dvdread", LLibraryId: "dvdread"},
		"--enable-libmysofa":         {name: "libmysofa", LLibraryId: "mysofa"},
		"--enable-libopencolorio":    {name: "OpenColorIO", LLibraryId: "opencolorio"},
		"--enable-libopencv":         {name: "opencv4", LLibraryId: "opencv"},
		"--enable-libqrencode":       {name: "libqrencode", LLibraryId: "qrencode"},
		"--enable-librabbitmq":       {name: "librabbitmq", LLibraryId: "rabbitmq"},
		"--enable-librsvg":           {name: "librsvg-2.0", LLibraryId: "rsvg"},
		"--enable-libsvtjpegxs":      {name: "SvtJpegxs", LLibraryId: "svtjpegxs"},
		"--enable-liblc3":            {name: "lc3", LLibraryId: "lc3"},
		"--enable-lv2":               {name: "lilv-0", LLibraryId: "lv2"},
		"--enable-liboapv":           {name: "oapv", LLibraryId: "oapv"},
		"--enable-lcms2":             {name: "lcms2", LLibraryId: "lcms2"},
		"--enable-opencl":            {name: "OpenCL", LLibraryId: "opencl"},
		"--enable-whisper":           {name: "whisper", LLibraryId: "whisper"},
	}
	release, releaseSupported := catalogfacts.LReleaseSupportResolve(ffmpegVersion)
	floorFor := func(LLibraryId string) string {
		if !releaseSupported || LLibraryId == "" {
			return ""
		}
		support, supported := release.LLibrarySupportGet(LLibraryId)
		if !supported {
			return ""
		}
		return support.MinVersion
	}
	modules := []LModulePkgconfig{}
	seen := map[string]bool{}
	for _, configureFlag := range configureFlags {
		e, exists := moduleByFlag[configureFlag]
		if !exists || seen[e.name] {
			continue
		}
		seen[e.name] = true
		modules = append(modules, LModulePkgconfig{Name: e.name, MinVersion: floorFor(e.LLibraryId)})
	}
	return modules
}
