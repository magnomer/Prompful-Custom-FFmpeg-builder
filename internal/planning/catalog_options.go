package planning

func LCatalogOptionBuild() []LOptionChoice {
	return []LOptionChoice{
		LOptionChoiceCreate("default-static", "Build static libraries", "Default FFmpeg source build", []string{}, "FFmpeg normally builds static libraries from source.", "Checked because this is normal FFmpeg configure behavior. No extra flag is needed unless you choose a different output type.", true, true),
		LOptionChoiceCreate("default-programs", "Build command-line programs", "Default FFmpeg source build", []string{}, "FFmpeg normally builds command-line programs such as ffmpeg.exe and ffprobe.exe.", "Checked because programs are built in a normal source build. Disable only if you want libraries without command-line tools.", true, true),
		LOptionChoiceCreate("default-ffmpeg", "Build ffmpeg.exe", "Default FFmpeg source build", []string{}, "Builds the main command-line converter most users run.", "Checked because ffmpeg.exe is part of a normal program build.", true, true),
		LOptionChoiceCreate("default-ffprobe", "Build ffprobe.exe", "Default FFmpeg source build", []string{}, "Builds the media inspection tool used to read stream and container information.", "Checked because ffprobe.exe is part of a normal program build.", true, true),
		LOptionChoiceCreate("enable-shared", "Build shared DLL libraries", "Output type", []string{"--enable-shared", "--disable-static"}, "Creates DLL-style FFmpeg libraries for other programs to load.", "FFmpeg platform documentation describes --enable-shared as the way to build FFmpeg libraries as DLLs on Windows. This changes the output type, so it is not selected by default here.", false, false),
		LOptionChoiceCreate("disable-ffplay", "Do not build ffplay", "Programs", []string{"--disable-ffplay"}, "Skips the simple playback test program.", "Useful when SDL playback support is unnecessary. ffmpeg.exe and ffprobe.exe are unaffected.", false, false),
		LOptionChoiceCreate("disable-autodetect", "Do not auto-use hidden system libraries", "Security and reproducibility", []string{"--disable-autodetect"}, "Makes the build less surprising by using only explicitly selected external libraries.", "Good for transparent/reproducible builds. Select this when you want the Review page to explain every external dependency.", false, false),
		LOptionChoiceCreate("disable-network", "Remove all networking support", "Security and reproducibility", []string{"--disable-network"}, "Builds FFmpeg without any network protocols for a smaller, offline-only tool.", "Select for local-only conversion or hardened builds. Disables HTTP, HTTPS, RTMP, SRT, and every other network input/output, so streaming protocols stop working even if their libraries are selected.", false, false),
		LOptionChoiceCreate("disable-asm", "Disable assembly optimizations", "Compatibility", []string{"--disable-asm"}, "Uses slower but simpler C code paths if assembly causes build problems.", "Normally leave unchecked because FFmpeg is faster with assembly optimizations.", false, false),
		LOptionChoiceCreate("disable-x86asm", "Disable x86 assembly", "Compatibility", []string{"--disable-x86asm"}, "Try this when NASM/YASM-related build problems occur.", "Normally leave unchecked for performance.", false, false),
		LOptionChoiceCreate("pkg-config-static", "Link external libraries statically", "Compatibility", []string{"--pkg-config-flags=--static"}, "Tells pkg-config to pull the full static dependency chain when linking external libraries.", "Often required for static Windows builds that use external libraries, so configure can find every transitive dependency. Has no effect on a shared/DLL build.", false, false),
		LOptionChoiceCreate("enable-runtime-cpudetect", "Detect CPU features at run time", "Compatibility", []string{"--enable-runtime-cpudetect"}, "Builds one binary that picks CPU optimizations while running, so it works across different processors.", "Useful when sharing the build with other machines. Slightly slower than a build tuned for one specific CPU.", false, false),
		LOptionChoiceCreate("disable-doc", "Skip documentation files", "Size and speed", []string{"--disable-doc"}, "Makes the build smaller by not building local documentation files.", "Not a normal source default. Select this when you only need binaries/libraries and do not need generated docs.", false, false),
		LOptionChoiceCreate("enable-small", "Prefer smaller binary size", "Size and speed", []string{"--enable-small"}, "Asks FFmpeg to prefer smaller output files over speed.", "Useful for constrained environments; may reduce performance.", false, false),
		LOptionChoiceCreate("enable-lto", "Enable link-time optimization (LTO)", "Size and speed", []string{"--enable-lto"}, "Lets the compiler optimize across files for a smaller and sometimes faster binary.", "Increases build time and memory use. Leave unchecked if linking fails or runs out of memory.", false, false),
		LOptionChoiceCreate("disable-debug", "Remove debug build data", "Debugging", []string{"--disable-debug"}, "Makes normal-use output smaller and simpler.", "Not a normal source default. Select this for ordinary release-style local builds; leave unchecked when investigating build or runtime problems.", false, false),
		LOptionChoiceCreate("disable-stripping", "Keep symbol information", "Debugging", []string{"--disable-stripping"}, "Keeps more build information for debugging.", "Usually unnecessary for ordinary users, but useful when diagnosing crashes or build problems.", false, false),
	}
}

func LOptionDefaultGet() []string {
	return []string{"default-static", "default-programs", "default-ffmpeg", "default-ffprobe", "pkg-config-static", "disable-doc"}
}

func LOptionChoiceCreate(id string, LDisplayName string, categoryName string, flags []string, plainExplanation string, technicalNote string, defaultEnabled bool, locked bool) LOptionChoice {
	return LOptionChoice{OptionId: id, DisplayName: LDisplayName, CategoryName: categoryName, ConfigureFlags: flags, PlainExplanation: plainExplanation, TechnicalNote: technicalNote, DefaultEnabled: defaultEnabled, Locked: locked, RiskLevelName: LOptionRiskResolve(id)}
}

func LOptionRiskResolve(optionId string) string {
	risks := map[string]string{
		"disable-autodetect": "review",
		"disable-network":    "feature-loss",
		"disable-asm":        "performance-loss",
		"disable-x86asm":     "performance-loss",
		"enable-lto":         "toolchain-sensitive",
		"enable-shared":      "output-type-change",
	}
	if risk, exists := risks[optionId]; exists {
		return risk
	}
	return "normal"
}
