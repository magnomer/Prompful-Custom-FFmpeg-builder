package planning

import "promptfulcustomffmpegbuilder/shared/librarysources"

// Non-Native library preparation is organized in three layers so that adding or
// updating a library is a small, well-scoped change and never bakes a version into the
// generic machinery:
//
//	(1) Generic layer  - track mechanics shared by every library: download a verified
//	    archive, extract it, build (internal) or import (external) into the MSYS2
//	    prefix, then verify. Implemented in the scripting + app packages. It is
//	    version-agnostic: it does not know any URL, hash, or release name, and it
//	    auto-detects the extracted source root rather than relying on a fixed dir name.
//	(2) Item layer (libraryItemSpec) - version-INDEPENDENT facts about one library:
//	    its build system, the header/library to verify, import sub-directories, and
//	    pkg-config module. These rarely change between releases and live in Go.
//	(3) Version layer (shared/librarysources/library-sources.json) - the ONLY place a
//	    concrete release lives: version string, archive URL, SHA-256, format, host, and
//	    any version-specific build options, keyed by the FFmpeg release being built.
//	    Re-pinning a release, or supporting a new FFmpeg version, is a data edit there,
//	    never a code change.
//
// preparationForLibrary pairs one item with the version resolved for the selected FFmpeg
// release, producing the plan-facing LibraryPreparation the generic layer consumes.

// LibraryPreparationMethod names how a non-Native library is made available to FFmpeg
// configure before the build runs.
type LibraryPreparationMethod string

const (
	// PreparationMethodInternalSource builds the library from a verified upstream
	// source archive inside the private MSYS2 environment and installs it into the
	// selected profile prefix.
	PreparationMethodInternalSource LibraryPreparationMethod = "internal-source-build"
	// PreparationMethodExternalImport downloads a verified vendor binary archive and
	// imports its headers and libraries into the selected profile prefix.
	PreparationMethodExternalImport LibraryPreparationMethod = "external-vendor-import"
)

// LibrarySourcePatch is one exact full-line replacement applied to the extracted source
// tree before the build runs. It exists for recipes that must work around an upstream
// portability bug that no build flag can fix (e.g. a duplicate overload that only collides
// under Windows/LLP64). File is relative to the source root; Find is matched as a whole
// line and replaced with Replace. The generated step fails if Find is absent, so a
// re-pinned release that changed the file is caught rather than silently built unpatched.
type LibrarySourcePatch struct {
	File    string `json:"file"`
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// GeneratedSourceFile is a file written into the extracted source tree before the build,
// for a file a release tarball omits because upstream generates it from a .git checkout
// (e.g. libvmaf's vcs_version.h). Path is relative to the source root; Lines are the file's
// lines, written verbatim.
type GeneratedSourceFile struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

// LibraryBuildSystem names the build system of an Internal-track source recipe. The
// generic source-build generator dispatches on this, so adding an autotools/make
// library is implemented in one place without disturbing existing cmake recipes.
type LibraryBuildSystem string

const (
	BuildSystemCMake LibraryBuildSystem = "cmake"
	// BuildSystemConfigureMake covers projects with a ./configure script (standard
	// autotools or x264/davs2-style custom configure) followed by make + make install.
	BuildSystemConfigureMake LibraryBuildSystem = "configure-make"
	BuildSystemMake          LibraryBuildSystem = "make"
	// BuildSystemMeson covers projects configured with `meson setup` and built with ninja
	// (e.g. libvmaf). The meson `-Dname=value` options reuse the CMakeOptions field, since
	// both build systems share that option syntax and the same safe-option validation.
	BuildSystemMeson LibraryBuildSystem = "meson"
)

// libraryItemSpec is layer (2): version-independent, item-specific.
type libraryItemSpec struct {
	libraryId   string
	displayName string
	trackName   LibraryTrackName
	method      LibraryPreparationMethod

	// buildDependencyPackages lists MSYS2 mingw package suffixes (e.g. "python") that
	// must be installed into the build environment before this library's source build
	// runs, for libraries whose build system find_package()s another library at configure
	// time. They are profile-prefixed at plan time (mingw-w64-<profile>-x86_64-<suffix>),
	// so the recipe stays profile-independent. Most recipes need none.
	buildDependencyPackages []string

	// msysBuildDependencyPackages lists MSYS2 base ("msys" repo) package names installed
	// verbatim ??not profile-prefixed ??before the build runs. These are build-time tools
	// that live in /usr/bin rather than the mingw prefix, e.g. the autotools an autogen
	// recipe needs (autoconf-wrapper, automake-wrapper, libtool); modern base-devel no longer
	// pulls them. Most recipes need none.
	msysBuildDependencyPackages []string

	// cFlags are extra C compiler flags exported as CFLAGS for the build. Used to demote a
	// GCC-14 hard error back to a warning for an older C library that predates it (e.g. libvmaf
	// 1.5.2's implicit function declarations). Honored by the meson generator. Most recipes
	// need none.
	cFlags []string

	// internal source build (cmake)
	buildSystem       LibraryBuildSystem
	cmakeOptions      []string // intrinsic to the library, not to a version
	cmakeBuildTargets []string // named targets to build; empty = the default target

	// internal source build (configure-make)
	configureSubdir    string   // dir holding ./configure, relative to source root (e.g. "build/linux")
	configureOptions   []string // intrinsic configure flags (--prefix is added automatically)
	makeBuildTargets   []string // make targets for the build step; empty = default target
	makeInstallTargets []string // make targets for the install step; empty = "install"
	// runAutogen bootstraps an autotools project whose tag tarball ships no generated
	// ./configure (only configure.ac + autogen.sh). The generator runs autogen.sh (or
	// autoreconf -fiv) at the source root before ./configure.
	runAutogen bool

	// internal source build (plain make): a Makefile-only project with no lib-only install
	// target installs by copying these source-relative artifacts into the prefix ??each
	// header into include/ (by basename) and the static archive into lib/. makeVariables are
	// NAME=VALUE command-line overrides that skip the makefile's own (e.g. SDL_CFLAGS=).
	makeVariables          []string
	makeInstallHeaderFiles []string
	makeStaticLibFile      string

	// external vendor import
	importIncludeSubdir string
	importLibSubdir     string

	// pkgConfigName is the installed pkg-config module (e.g. "lcevc_dec", "libvvenc") ??	// the same module FFmpeg's require_pkg_config looks up. It is used both to validate
	// the pinned version against FFmpeg's required minimum (the preflight safety net) and,
	// when pkgConfigAppendLibs is set, to patch that .pc. pkgConfigAppendLibs lists bare
	// link libraries (e.g. "stdc++", "m") appended to the .pc's Libs line after install:
	// some CMake projects emit a static .pc that lists the C++/math runtime BEFORE their
	// own static archives, so GNU ld discards it too early and the link fails with
	// undefined std::/operator new references; appending the runtime at the END fixes the
	// static link order. Leave pkgConfigName empty for libraries FFmpeg does not
	// version-check via pkg-config (header-only or vendor-imported).
	pkgConfigName       string
	pkgConfigAppendLibs []string

	// pkgConfigLibsLine, when set, replaces the installed .pc's whole "Libs:" line value.
	// Needed when a bare -l<name> in the .pc would resolve to a same-named shared import
	// library (.dll.a) that shadows this recipe's own static archive in a shared prefix:
	// forcing -l:lib<name>.a (or an absolute archive path) makes the link pick the static
	// archive. May reference ${libdir}.
	pkgConfigLibsLine string

	// privatePrefixInstall installs this library into its own per-library prefix
	// (<profile>/opt/customffmpeg/<pkgConfigName>) instead of the shared MSYS2 prefix,
	// and the configure step prepends that prefix's pkgconfig dir to PKG_CONFIG_PATH.
	// Needed when a library ships archives whose base names collide with a different
	// package's in the shared prefix ??e.g. LibreSSL's libtls brings its own libssl.a /
	// libcrypto.a, which the openssl package (pulled in by libssh/srt/rabbitmq) also owns
	// and would otherwise win. Pair with a pkgConfigLibsLine of absolute ${libdir}/lib*.a
	// archive paths so the link binds this prefix's archives regardless of -L ordering;
	// the installed .pc's Requires/Requires.private are stripped so no bare -l<name> from a
	// transitive module re-introduces the shared-prefix archive.
	privatePrefixInstall bool

	// sourcePatches are exact full-line edits applied to the extracted source before the
	// build runs, for upstream portability bugs no build flag can fix. Most recipes need none.
	sourcePatches []LibrarySourcePatch

	// generatedSourceFiles are files written into the extracted source before the build, for a
	// file a release tarball omits because upstream generates it from a .git checkout (e.g.
	// libvmaf's vcs_version.h). Most recipes need none.
	generatedSourceFiles []GeneratedSourceFile

	// verification (both methods)
	verifyHeaderRelativePath string
	verifyLibStem            string
}

// LibraryPreparation is the flattened, plan-facing projection of an item paired with its
// resolved version. It carries no source-directory name: the generic layer auto-detects
// the extracted root, so nothing here is tied to a release beyond the resolved URL/hash.
type LibraryPreparation struct {
	LibraryId   string                   `json:"libraryId"`
	DisplayName string                   `json:"displayName"`
	TrackName   LibraryTrackName         `json:"trackName"`
	Method      LibraryPreparationMethod `json:"method"`

	BuildSystem LibraryBuildSystem `json:"buildSystem"`

	// CFlags are extra C compiler flags exported as CFLAGS for the build (e.g. demoting a
	// GCC-14 hard error to a warning for an older C library). Honored by the meson generator.
	CFlags []string `json:"cFlags,omitempty"`

	// Version is the resolved release of the library (from library-sources.json), used to
	// validate against FFmpeg's required minimum and to surface what will be built.
	Version string `json:"version"`

	// BuildDependencyPackages holds fully profile-prefixed MSYS2 package names installed
	// into the build environment before this library's source build runs. Empty for
	// libraries that build with only the base toolchain.
	BuildDependencyPackages []string `json:"buildDependencyPackages,omitempty"`

	// MsysBuildDependencyPackages holds MSYS2 base ("msys" repo) package names installed
	// verbatim (not profile-prefixed) before the build runs ??build-time tools in /usr/bin
	// such as the autotools an autogen recipe needs. Empty for most libraries.
	MsysBuildDependencyPackages []string `json:"msysBuildDependencyPackages,omitempty"`

	ArchiveUrl          string `json:"archiveUrl"`
	ArchiveSha256Hash   string `json:"archiveSha256Hash"`
	AllowedDownloadHost string `json:"allowedDownloadHost"`
	ArchiveFormat       string `json:"archiveFormat"`

	CMakeOptions      []string `json:"cmakeOptions,omitempty"`
	CMakeBuildTargets []string `json:"cmakeBuildTargets,omitempty"`

	ConfigureSubdir    string   `json:"configureSubdir,omitempty"`
	ConfigureOptions   []string `json:"configureOptions,omitempty"`
	MakeBuildTargets   []string `json:"makeBuildTargets,omitempty"`
	MakeInstallTargets []string `json:"makeInstallTargets,omitempty"`
	RunAutogen         bool     `json:"runAutogen,omitempty"`

	MakeVariables          []string `json:"makeVariables,omitempty"`
	MakeInstallHeaderFiles []string `json:"makeInstallHeaderFiles,omitempty"`
	MakeStaticLibFile      string   `json:"makeStaticLibFile,omitempty"`

	ImportIncludeSubdir string `json:"importIncludeSubdir,omitempty"`
	ImportLibSubdir     string `json:"importLibSubdir,omitempty"`

	PkgConfigName        string   `json:"pkgConfigName,omitempty"`
	PkgConfigAppendLibs  []string `json:"pkgConfigAppendLibs,omitempty"`
	PkgConfigLibsLine    string   `json:"pkgConfigLibsLine,omitempty"`
	PrivatePrefixInstall bool     `json:"privatePrefixInstall,omitempty"`

	VerifyHeaderRelativePath string `json:"verifyHeaderRelativePath"`
	VerifyLibStem            string `json:"verifyLibStem"`

	SourcePatches []LibrarySourcePatch `json:"sourcePatches,omitempty"`

	GeneratedSourceFiles []GeneratedSourceFile `json:"generatedSourceFiles,omitempty"`
}

// buildPreparation projects an item plus its resolved version source into a plan-facing
// LibraryPreparation.
func buildPreparation(item libraryItemSpec, source librarysources.LibrarySource) LibraryPreparation {
	cmakeOptions := append(append([]string{}, item.cmakeOptions...), source.ExtraCMakeOptions...)
	if len(cmakeOptions) == 0 {
		cmakeOptions = nil
	}
	return LibraryPreparation{
		LibraryId:                   item.libraryId,
		DisplayName:                 item.displayName,
		TrackName:                   item.trackName,
		Method:                      item.method,
		BuildSystem:                 item.buildSystem,
		CFlags:                      append([]string{}, item.cFlags...),
		Version:                     source.Version,
		BuildDependencyPackages:     append([]string{}, item.buildDependencyPackages...),
		MsysBuildDependencyPackages: append([]string{}, item.msysBuildDependencyPackages...),
		ArchiveUrl:                  source.Url,
		ArchiveSha256Hash:           source.Sha256,
		AllowedDownloadHost:         source.Host,
		ArchiveFormat:               source.Format,
		CMakeOptions:                cmakeOptions,
		CMakeBuildTargets:           item.cmakeBuildTargets,
		ConfigureSubdir:             item.configureSubdir,
		ConfigureOptions:            item.configureOptions,
		MakeBuildTargets:            item.makeBuildTargets,
		MakeInstallTargets:          item.makeInstallTargets,
		RunAutogen:                  item.runAutogen,
		MakeVariables:               append([]string{}, item.makeVariables...),
		MakeInstallHeaderFiles:      append([]string{}, item.makeInstallHeaderFiles...),
		MakeStaticLibFile:           item.makeStaticLibFile,
		ImportIncludeSubdir:         item.importIncludeSubdir,
		ImportLibSubdir:             item.importLibSubdir,
		PkgConfigName:               item.pkgConfigName,
		PkgConfigAppendLibs:         append([]string{}, item.pkgConfigAppendLibs...),
		PkgConfigLibsLine:           item.pkgConfigLibsLine,
		PrivatePrefixInstall:        item.privatePrefixInstall,
		VerifyHeaderRelativePath:    item.verifyHeaderRelativePath,
		VerifyLibStem:               item.verifyLibStem,
		SourcePatches:               append([]LibrarySourcePatch{}, item.sourcePatches...),
		GeneratedSourceFiles:        append([]GeneratedSourceFile{}, item.generatedSourceFiles...),
	}
}

// libraryItemSpecs holds the version-independent build/verify facts for each library that
// has an implemented recipe, keyed by library id. The concrete release for each is
// resolved from library-sources.json for the FFmpeg version being built; libraries with
// no item here (or no recorded version) stay blocked by the planner.
//
// Adding a library = add its item spec here and a version entry in library-sources.json.
// Re-pinning a release = edit only library-sources.json.
var libraryItemSpecs = map[string]libraryItemSpec{
	// AviSynth+ is header-only for FFmpeg: --enable-avisynth requires just the C
	// interface headers (avisynth/avisynth_c.h); the AviSynth runtime DLL is loaded at
	// run time, so there is no link library to verify. Built with -DHEADERS_ONLY=ON,
	// which only needs the VersionGen target run to produce the installed version.h
	// (that target is not part of the default build, hence cmakeBuildTargets). Builds
	// from a release tarball without a .git directory: Version.cmake skips git and still
	// writes version.h. FFmpeg does not version-check it via pkg-config, so no
	// pkgConfigName.
	"avisynthplus": {
		libraryId:                "avisynthplus",
		displayName:              "AviSynth+",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		cmakeOptions:             []string{"-DHEADERS_ONLY=ON"},
		cmakeBuildTargets:        []string{"VersionGen"},
		verifyHeaderRelativePath: "avisynth/avisynth_c.h",
		verifyLibStem:            "", // runtime-loaded DLL; no build-time link library
	},
	// libdavs2 (AVS2 decoder) uses an x264-style configure (in build/linux) + make, not
	// cmake. configure handles its own source-path cd internally, so the script just runs
	// ./configure from build/linux. Installs davs2.h, davs2_config.h, davs2.pc and
	// libdavs2.a via install-lib-static; FFmpeg finds it through pkg-config (davs2 >= 1.6.0).
	"davs2": {
		libraryId:                "davs2",
		displayName:              "libdavs2",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemConfigureMake,
		configureSubdir:          "build/linux",
		configureOptions:         []string{"--disable-cli", "--enable-pic"},
		makeBuildTargets:         []string{"lib-static"},
		makeInstallTargets:       []string{"install-lib-static"},
		pkgConfigName:            "davs2",
		verifyHeaderRelativePath: "davs2.h",
		verifyLibStem:            "davs2",
	},
	// libxavs2 (AVS2 encoder) is davs2's sibling from the same pkuvcl org and uses the
	// identical x264-style configure (in build/linux) + make. install-lib-static installs
	// xavs2.h, xavs2_config.h, xavs2.pc and libxavs2.a; FFmpeg finds it through pkg-config
	// (xavs2 >= 1.3.0).
	//
	// The 1.4 source (2018) passes encoder_aec_encode_one_frame (returns void) where
	// xavs2_threadpool_run expects void *(*)(void *). Pre-14 GCC accepted this with a
	// warning; GCC 14 makes -Wincompatible-pointer-types a hard error by default and the
	// build fails. The mismatch is benign (the ignored return value is never read), so demote
	// it back to a warning with --extra-cflags, the same workaround distros ship for xavs2.
	"xavs2": {
		libraryId:                "xavs2",
		displayName:              "xavs2",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemConfigureMake,
		configureSubdir:          "build/linux",
		configureOptions:         []string{"--disable-cli", "--enable-pic", "--extra-cflags=-Wno-error=incompatible-pointer-types"},
		makeBuildTargets:         []string{"lib-static"},
		makeInstallTargets:       []string{"install-lib-static"},
		pkgConfigName:            "xavs2",
		verifyHeaderRelativePath: "xavs2.h",
		verifyLibStem:            "xavs2",
	},
	"uavs3d": {
		libraryId:                "uavs3d",
		displayName:              "libuavs3d",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		pkgConfigName:            "uavs3d",
		verifyHeaderRelativePath: "uavs3d.h",
		verifyLibStem:            "uavs3d",
	},
	// LCEVCdec (V-Nova) is a CMake SDK whose core decoder needs no third-party libraries
	// once samples, tests, JSON config and the Vulkan pipeline are disabled ??fmt,
	// range-v3, xxHash, nlohmann-json, GTest, CLI11 and ffmpeg-libs are all confined to
	// those gated targets. Its configure does still require a Python3 interpreter (it runs
	// cmake/tools/version_files.py to generate version headers), which the base toolchain
	// does not ship, hence the python build dependency. conan is optional: the build sets
	// CMAKE_FIND_PACKAGE_PREFER_CONFIG and falls back to native/pkg-config discovery.
	// Built static (BUILD_SHARED_LIBS=OFF) so the .pc collapses every component archive
	// into Libs and FFmpeg links it through pkg-config (lcevc_dec >= 4.0.0). The installed
	// lcevc_dec_api archive exports LCEVC_CreateDecoder, which FFmpeg's require_pkg_config
	// link-checks.
	"lcevc-dec": {
		libraryId:               "lcevc-dec",
		displayName:             "liblcevc-dec",
		trackName:               LibraryTrackInternal,
		method:                  PreparationMethodInternalSource,
		buildSystem:             BuildSystemCMake,
		buildDependencyPackages: []string{"python"},
		cmakeOptions: []string{
			"-DBUILD_SHARED_LIBS=OFF",
			"-DVN_SDK_EXECUTABLES=OFF",
			"-DVN_SDK_UNIT_TESTS=OFF",
			"-DVN_SDK_SAMPLE_SOURCE=OFF",
			"-DVN_SDK_JSON_CONFIG=OFF",
			"-DVN_SDK_PIPELINE_VULKAN=OFF",
			"-DVN_SDK_DOCS=OFF",
			"-DVN_SDK_SYSTEM_INSTALL=OFF",
		},
		// Upstream's static lcevc_dec.pc lists -lstdc++ -lm BEFORE its own static
		// archives, which breaks GNU static link order (undefined std::/operator new).
		// Append the C++/math runtime after the archives to fix it.
		pkgConfigName:            "lcevc_dec",
		pkgConfigAppendLibs:      []string{"stdc++", "m"},
		verifyHeaderRelativePath: "LCEVC/lcevc_dec.h",
		verifyLibStem:            "lcevc_dec_api",
	},
	// vvenc (Fraunhofer VVC/H.266 encoder) is a self-contained CMake project: nlohmann_json
	// and simde are bundled in thirdparty/ and used by default, so it needs no build
	// dependencies and no git/python at configure. BUILD_SHARED_LIBS already defaults OFF;
	// VVENC_LIBRARY_ONLY=ON skips the apps/tests while still installing the lib, headers,
	// and libvvenc.pc. The static .pc keeps -lvvenc in Libs and the C++ runtime in
	// Libs.private (correct order), and the build's default --pkg-config-flags=--static
	// pulls Libs.private, so no .pc fixup is needed. FFmpeg checks libvvenc >= 1.6.1.
	"vvenc": {
		libraryId:                "vvenc",
		displayName:              "vvenc",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		cmakeOptions:             []string{"-DBUILD_SHARED_LIBS=OFF", "-DVVENC_LIBRARY_ONLY=ON"},
		pkgConfigName:            "libvvenc",
		verifyHeaderRelativePath: "vvenc/vvenc.h",
		verifyLibStem:            "vvenc",
	},
	// mpeghdec (Fraunhofer MPEG-H 3D Audio decoder) is a self-contained CMake C++ project:
	// add_library(mpeghdec) takes its type from BUILD_SHARED_LIBS, so -DBUILD_SHARED_LIBS=OFF
	// builds the static libmpeghdec.a; -Dmpeghdec_BUILD_BINARIES=OFF skips the demo apps while
	// still installing the lib, the include/mpeghdec/ headers, and mpeghdec.pc. The project is
	// C++ but its generated static mpeghdec.pc lists only "-lmpeghdec -lm" in Libs, so FFmpeg's
	// C link probe fails on undefined std::/operator new; append the C++ runtime after the
	// archive to fix the static link order. FFmpeg checks mpeghdec >= 3.0.0 via pkg-config and
	// link-probes mpeghdecoder_init. libmpeghdec is on FFmpeg's nonfree list, so the catalog row
	// also carries --enable-nonfree.
	"mpeghdec": {
		libraryId:                "mpeghdec",
		displayName:              "libmpeghdec",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		cmakeOptions:             []string{"-DBUILD_SHARED_LIBS=OFF", "-Dmpeghdec_BUILD_BINARIES=OFF"},
		pkgConfigName:            "mpeghdec",
		pkgConfigAppendLibs:      []string{"stdc++"},
		verifyHeaderRelativePath: "mpeghdec/mpeghdecoder.h",
		verifyLibStem:            "mpeghdec",
		// libFDK declares fMin/fMax(SHORT,SHORT) overloads guarded only by
		// !defined(__LP64__) && defined(__x86_64__). On Windows mingw x86_64 (LLP64) __LP64__
		// is undefined, so they compile ??but FIXP_SGL is itself typedef'd to SHORT, so they
		// duplicate the existing fMin/fMax(FIXP_SGL,FIXP_SGL) overloads and the build fails
		// with "redefinition of 'SHORT fMax(SHORT, SHORT)'". Upstream only tested LP64
		// (Linux/macOS), where the block is skipped. Disable just the SHORT-overload block by
		// neutralising its #if guard; the sibling INT overloads (INT != FIXP_DBL) are fine.
		//
		// Second patch: mpeghexport.h marks the public API __declspec(dllimport) on Windows
		// unless MPEGHDEC_STATIC is defined. We build the static archive, but the installed
		// mpeghdec.pc only carries "Cflags: -I...", so FFmpeg compiles the header in dllimport
		// mode and the configure link probe fails with "undefined reference to
		// __imp_mpeghdecoder_init". Add -DMPEGHDEC_STATIC=1 to the .pc Cflags (patched in the
		// pc.in before CMake's configure_file copies it) so consumers link the static symbols.
		sourcePatches: []LibrarySourcePatch{
			{
				File:    "src/libFDK/include/common_fix.h",
				Find:    "#if !defined(_MSC_VER) && defined(__x86_64__)",
				Replace: "#if 0 /* mpeghdec recipe patch: SHORT fMin/fMax duplicate FIXP_SGL(=short) on LLP64 mingw x86_64 */",
			},
			{
				File:    "mpeghdec.pc.in",
				Find:    `Cflags: -I"${includedir}"`,
				Replace: `Cflags: -I"${includedir}" -DMPEGHDEC_STATIC=1`,
			},
		},
	},
	// quirc (QR decoder) ships only a plain Makefile ??no configure, no cmake, no pkg-config.
	// Its `make install` also builds the camera demos (SDL/libpng/v4l), which do not build on
	// the base Windows toolchain, so the recipe builds just the static archive target
	// (libquirc.a) and installs by copying lib/quirc.h and libquirc.a into the prefix. FFmpeg
	// detects it with `require libquirc quirc.h quirc_decode -lquirc` (header + lib + symbol,
	// not pkg-config), so there is no pkgConfigName and the version preflight skips it. ISC
	// licensed, so the catalog row stays lgpl.
	"quirc": {
		libraryId:        "quirc",
		displayName:      "libquirc",
		trackName:        LibraryTrackInternal,
		method:           PreparationMethodInternalSource,
		buildSystem:      BuildSystemMake,
		makeBuildTargets: []string{"libquirc.a"},
		// quirc's global QUIRC_CFLAGS embeds $(shell pkg-config --cflags sdl); with no SDL
		// installed, pkg-config's error text leaks into the compile command and breaks the
		// build. The static lib needs no SDL, so override SDL_CFLAGS/SDL_LIBS to empty on the
		// command line (which also skips the makefile's pkg-config call).
		makeVariables:            []string{"SDL_CFLAGS=", "SDL_LIBS="},
		makeInstallHeaderFiles:   []string{"lib/quirc.h"},
		makeStaticLibFile:        "libquirc.a",
		verifyHeaderRelativePath: "quirc.h",
		verifyLibStem:            "quirc",
	},
	// libklvanc (VANC/ancillary-data parser) is an autotools project whose GitHub tag
	// tarball ships no generated ./configure ??only configure.ac + autogen.sh ??so the
	// recipe bootstraps with runAutogen (autoreconf -fiv). The autotools are installed as
	// msysBuildDependencyPackages: modern MSYS2 base-devel no longer pulls autoconf/automake/
	// libtool, and the project's own autogen.sh requires an action argument (it prints a
	// usage line and exits 1 when run bare), so autoreconf is the reliable bootstrap.
	// Built static with --enable-shared=no. The repo is recursive automake (SUBDIRS =
	// "src tools"); the sibling tools/ link -ldl and use BSD sockets (sys/socket.h), both
	// absent on the mingw-w64 toolchain, so `make` at the root fails. Override SUBDIRS=src
	// on the build and install lines to build/install just the library and skip tools/.
	// FFmpeg detects it with `require libklvanc libklvanc/vanc.h klvanc_context_create
	// -lklvanc` (header + lib + symbol, not pkg-config), so there is no pkgConfigName and
	// the version preflight skips it. BSD/MIT licensed, so the catalog row stays lgpl.
	"klvanc": {
		libraryId:                   "klvanc",
		displayName:                 "libklvanc",
		trackName:                   LibraryTrackInternal,
		method:                      PreparationMethodInternalSource,
		buildSystem:                 BuildSystemConfigureMake,
		runAutogen:                  true,
		msysBuildDependencyPackages: []string{"autoconf-wrapper", "automake-wrapper", "libtool"},
		configureOptions:            []string{"--enable-shared=no"},
		makeVariables:               []string{"SUBDIRS=src"},
		// libklvanc includes <sys/errno.h> (a BSD/glibc path) in four headers; mingw-w64
		// has no sys/errno.h ??the error constants live in the standard <errno.h>. Rewrite
		// each include so the library compiles on the Windows toolchain. vanc.h carries the
		// include twice; the patch rewrites both occurrences.
		sourcePatches: []LibrarySourcePatch{
			{File: "src/core-private.h", Find: "#include <sys/errno.h>", Replace: "#include <errno.h>"},
			{File: "src/libklvanc/vanc.h", Find: "#include <sys/errno.h>", Replace: "#include <errno.h>"},
			{File: "src/libklvanc/vanc-lines.h", Find: "#include <sys/errno.h>", Replace: "#include <errno.h>"},
			{File: "src/libklvanc/vanc-packets.h", Find: "#include <sys/errno.h>", Replace: "#include <errno.h>"},
		},
		verifyHeaderRelativePath: "libklvanc/vanc.h",
		verifyLibStem:            "klvanc",
	},
	// libtls (LibreSSL's TLS layer) is the LibreSSL portable distribution, which builds
	// libtls + libssl + libcrypto together. FFmpeg detects it via pkg-config:
	// `require_pkg_config libtls libtls tls.h tls_init` (module libtls, header tls.h, symbol
	// tls_init), with no version floor, so the version preflight has nothing to compare and
	// is a no-op. Built with the portable CMake build: BUILD_SHARED_LIBS=OFF for the static
	// archives this builder links, LIBRESSL_APPS=OFF and LIBRESSL_TESTS=OFF to skip the
	// openssl/nc command-line apps and the test suite (neither is needed and both add build
	// surface). CMake installs libtls.a/libssl.a/libcrypto.a and the matching .pc files.
	// libtls is one of the mutually exclusive TLS backends (openssl/gnutls/mbedtls/libtls);
	// the planner blocks selecting more than one. BSD/ISC licensed, so the catalog row stays lgpl.
	//
	// The decisive problem: LibreSSL's libssl.a/libcrypto.a share base names with the openssl
	// package's archives, and openssl is pulled into the shared MSYS2 prefix as a transitive
	// dependency of non-TLS libraries (libssh, srt, rabbitmq-c, curl, ...) regardless of the
	// chosen FFmpeg TLS backend. In one shared prefix the openssl archives win, so libtls.a
	// links against OpenSSL's libssl/libcrypto, which lack every LibreSSL-internal symbol
	// (libressl_*, posix_*, SSL_CTX_*_mem, sk_*, arc4random, getuid) ??FFmpeg's pkg-config
	// probe then fails with ~90 undefined references. A bare-name or -l:lib<name>.a override
	// cannot fix it because the libssl.a file on disk simply IS OpenSSL's.
	//
	// Fix: privatePrefixInstall puts LibreSSL in its own prefix so its archives never collide,
	// and pkgConfigLibsLine binds them by absolute ${libdir}/lib*.a path (ld -L ordering can no
	// longer pick the shared OpenSSL), plus the Windows system libraries libcrypto needs
	// (ws2_32 sockets, bcrypt RNG, ntdll). Order tls -> ssl -> crypto -> system is what ld
	// needs. The private .pc's Requires/Requires.private are stripped on install so no bare
	// -lssl/-lcrypto from a transitive module re-introduces the shared-prefix archive. The
	// configure step prepends the private prefix's pkgconfig dir to PKG_CONFIG_PATH so FFmpeg
	// finds this libtls.pc. srt and friends keep linking the shared OpenSSL unchanged.
	"libtls": {
		libraryId:     "libtls",
		displayName:   "libtls",
		trackName:     LibraryTrackInternal,
		method:        PreparationMethodInternalSource,
		buildSystem:   BuildSystemCMake,
		cmakeOptions:  []string{"-DBUILD_SHARED_LIBS=OFF", "-DLIBRESSL_APPS=OFF", "-DLIBRESSL_TESTS=OFF"},
		pkgConfigName: "libtls",
		// Installed privately (see privatePrefixInstall) so LibreSSL's libtls.a / libssl.a /
		// libcrypto.a never share the prefix with the openssl package's same-named archives.
		// Absolute ${libdir}/lib*.a paths bind LibreSSL's own archives regardless of any
		// earlier -L<shared>/lib on FFmpeg's link line; ${libdir} expands to the private lib.
		pkgConfigLibsLine:        "${libdir}/libtls.a ${libdir}/libssl.a ${libdir}/libcrypto.a -lws2_32 -lbcrypt -lntdll",
		privatePrefixInstall:     true,
		verifyHeaderRelativePath: "tls.h",
		verifyLibStem:            "tls",
	},
	// libvmaf is normally the MSYS2 native package (Native track), which is correct for
	// FFmpeg >= 5.1 (those probe the libvmaf 2.x `vmaf_init` API the current 3.x package
	// provides). FFmpeg 4.4/5.0 instead probe the libvmaf 1.x `compute_vmaf` API, removed at
	// libvmaf 2.0, so the native package cannot satisfy them and MSYS2 ships no 1.x package.
	// For those release lines the manifest marks vmaf sourceBuild and the planner flips it to
	// this Internal recipe, which builds the pinned libvmaf 1.5.2 (resolved per FFmpeg release
	// from library-sources.json) with meson. The meson source lives in the libvmaf/ subdir of
	// the repo; -Denable_tests/-Denable_docs=false skip the test/doc targets (the only options
	// 1.5.2 exposes). libvmaf is C++, and its static libvmaf.pc lists only -lvmaf, so the C++/
	// math runtime is appended after the archive for correct GNU static link order. FFmpeg
	// detects it with require_pkg_config libvmaf (header libvmaf/libvmaf.h, symbol compute_vmaf).
	"vmaf": {
		libraryId:               "vmaf",
		displayName:             "libvmaf",
		trackName:               LibraryTrackInternal,
		method:                  PreparationMethodInternalSource,
		buildSystem:             BuildSystemMeson,
		buildDependencyPackages: []string{"meson", "ninja"},
		configureSubdir:         "libvmaf",
		cmakeOptions:            []string{"-Denable_tests=false", "-Denable_docs=false"},
		// libvmaf.rc.c (in the WIP vmaf_rc targets, which ninja still compiles as part of `all`)
		// #includes vcs_version.h, which upstream generates with meson vcs_tag from a .git checkout
		// absent in the release tarball — so the build aborts with "vcs_version.h: No such file".
		// FFmpeg 4.4/5.0 only need the legacy libvmaf (compute_vmaf), which never includes it, but
		// the target compiles regardless; supply a static header so every target builds. The
		// VMAF_VERSION value is cosmetic (a reported version string), not part of the linked API.
		generatedSourceFiles: []GeneratedSourceFile{
			{
				Path:  "libvmaf/include/vcs_version.h",
				Lines: []string{"/* auto-generated, do not edit */", "#define VMAF_VERSION \"v1.5.2\""},
			},
		},
		// libvmaf 1.5.2 (2020) predates GCC 14, which promoted several lax-C constructs from
		// warning to hard error: it calls memset without <string.h> (implicit-function-declaration)
		// and uses vmaf_ceiln/vmaf_floorn without including their alignment.h declaration. The
		// functions are defined and their int returns match the use sites, so demote the GCC-14 C
		// hard errors back to warnings (the same workaround distros ship for old C on GCC 14).
		cFlags: []string{
			"-Wno-error=implicit-function-declaration",
			"-Wno-error=implicit-int",
			"-Wno-error=int-conversion",
			"-Wno-error=incompatible-pointer-types",
		},
		pkgConfigName:            "libvmaf",
		pkgConfigAppendLibs:      []string{"stdc++", "m"},
		verifyHeaderRelativePath: "libvmaf/libvmaf.h",
		verifyLibStem:            "vmaf",
	},
	// libmfx is Intel's open Media SDK dispatcher (lu-zero/mfx_dispatch), source-built because
	// MSYS2 ships no package for it. CMake builds the static libmfx.a (target "mfx") and installs
	// mfx/mfxvideo.h, libmfx.a, and libmfx.pc. FFmpeg finds it via pkg-config (module libmfx,
	// header mfx/mfxvideo.h, symbol MFXInit). Two fixes to the upstream CMake .pc:
	//   - Its template hardcodes "Version: 2013", which fails FFmpeg 6.1+'s "libmfx < 2.0" bound;
	//     a source patch rewrites it to 1.35 (in the required [1.28, 2.0) window, and the API
	//     version this 1.35.1 dispatcher implements).
	//   - Its Libs line lists the C++ runtime BEFORE the static archive and omits the Windows COM
	//     libraries the dispatcher needs; pkgConfigLibsLine replaces it with the archive first then
	//     -lstdc++ and the Win32 libs (ole32/uuid for DXVA COM, advapi32 for the registry probe),
	//     matching the project's own autotools .pc DLLIB. This fixes the GNU static link order.
	"libmfx": {
		libraryId:                "libmfx",
		displayName:              "libmfx",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		pkgConfigName:            "libmfx",
		pkgConfigLibsLine:        "-L${libdir} -lmfx -lstdc++ -lole32 -luuid -ladvapi32",
		verifyHeaderRelativePath: "mfx/mfxvideo.h",
		verifyLibStem:            "mfx",
		// A second patch adds src/mfx_driver_store_loader.cpp to the CMake Windows SOURCES list: the
		// upstream CMakeLists omits it, but mfx_library_iterator.cpp references MFX::DriverStoreLoader,
		// so libmfx.a is built without those symbols and FFmpeg's link test fails with undefined
		// references. The file is appended on the same line as mfx_win_reg_key.cpp (CMake splits the
		// list on whitespace) so the patch replaces a single line, which the source-patch step requires.
		sourcePatches: []LibrarySourcePatch{
			{File: "libmfx.pc.cmake", Find: "Version: 2013", Replace: "Version: 1.35"},
			{File: "CMakeLists.txt", Find: "    src/mfx_win_reg_key.cpp", Replace: "    src/mfx_win_reg_key.cpp src/mfx_driver_store_loader.cpp"},
		},
	},
	// SVT-AV1's encoder API dropped svt_av1_enc_init_handle's middle p_app_data argument at
	// v3.0.0, but FFmpeg 4.4 through 7.1 call the 3-argument form, so the MSYS2 native SVT-AV1
	// (now 3.x) makes those release lines fail to compile libavcodec/libsvtav1.c ("too many
	// arguments to svt_av1_enc_init_handle"). Those lines mark svt-av1 sourceBuild and the planner
	// flips it to this Internal recipe, which builds the pinned SVT-AV1 2.3.0 (the last 2.x,
	// 3-argument release, resolved per FFmpeg release from library-sources.json) with CMake.
	// BUILD_SHARED_LIBS=OFF yields the static libSvtAv1Enc.a; BUILD_APPS/BUILD_TESTING=OFF skip the
	// encoder app and the gtest unit tests. The bundled cpuinfo is an OBJECT library compiled into
	// libSvtAv1Enc.a, so no separate cpuinfo install is needed; the installed .pc's Libs.private
	// (-lpthread -lm) covers the rest of the static link. nasm assembles the x86 SIMD. FFmpeg finds
	// it via pkg-config (module SvtAv1Enc, header EbSvtAv1Enc.h under include/svt-av1, symbol
	// svt_av1_enc_init_handle).
	"svt-av1": {
		libraryId:                "svt-av1",
		displayName:              "SVT-AV1",
		trackName:                LibraryTrackInternal,
		method:                   PreparationMethodInternalSource,
		buildSystem:              BuildSystemCMake,
		buildDependencyPackages:  []string{"nasm"},
		cmakeOptions:             []string{"-DBUILD_SHARED_LIBS=OFF", "-DBUILD_APPS=OFF", "-DBUILD_TESTING=OFF"},
		pkgConfigName:            "SvtAv1Enc",
		verifyHeaderRelativePath: "svt-av1/EbSvtAv1Enc.h",
		verifyLibStem:            "SvtAv1Enc",
	},
	// XEVE (MPEG-5 EVC encoder) is normally the MSYS2 native package (Native track), correct
	// for FFmpeg >= 7.1: those adapted libavcodec/libxeve.c to XEVE 0.5's API, where param->fps
	// is an XEVE_RATIONAL struct (FFmpeg 7.1+ configure pins xeve >= 0.5.1). FFmpeg 7.0's
	// libxeve.c instead assigns a scalar to param->fps, so the current native package (0.5.x)
	// makes 7.0 fail to compile libavcodec/libxeve.c ("incompatible types when assigning to type
	// 'XEVE_RATIONAL' from type 'long int'"). FFmpeg's pkg-config floor (xeve >= 0.4.3) is a
	// minimum not a maximum, so configure passes and the break only surfaces at make. The 7.0
	// line marks xeve sourceBuild and the planner flips it to this Internal recipe, building the
	// pinned XEVE 0.4.3 (the last 0.4.x, scalar-fps API, resolved per FFmpeg release from
	// library-sources.json) with CMake.
	//
	// version.txt: the CMakeLists derives the project version from `git describe`, absent in a
	// release tarball, then falls back to a version.txt it FATAL_ERRORs without; supply one (the
	// same fix MSYS2 ships) so configure proceeds and the installed xeve.pc reports 0.4.3, which
	// satisfies FFmpeg's xeve >= 0.4.3 check. Keep its value in lockstep with the source pin.
	// The sourcePatch drops -Werror from the project's CMAKE_C_FLAGS (also mirroring MSYS2):
	// XEVE 0.4.3 predates GCC 14, which promoted several lax-C constructs to hard errors, and an
	// env CFLAGS -Wno-error cannot win because the project appends -Werror after it.
	//
	// The CMakeLists always builds the static libxeve.a (target "xeve") alongside a shared
	// xeve_dynamic, and installs the static archive into lib/xeve/ (a subdir), the shared import
	// library libxeve.dll.a into lib/, and headers into include/xeve/. A bare -lxeve would pick
	// the shared import library, so pkgConfigLibsLine forces the static archive. It must use the
	// -L<dir> -l:<file> form, NOT a bare absolute archive path: FFmpeg's configure link test
	// (test_ld) splits pkg-config --libs into libs (args matching -l*|*.so, placed AFTER the test
	// object) and flags (everything else, placed BEFORE it). A bare ".../libxeve.a" matches
	// neither, so it lands before the object and GNU ld discards the archive before xeve_encode is
	// referenced ("undefined reference to xeve_encode"). -l:libxeve.a matches -l*, so it is placed
	// after the object and resolves; the -l: prefix pins the exact static archive over the .dll.a,
	// and -L${libdir}/xeve points at its subdir. XEVE is pure C; -lm and -lpthread cover its math
	// and threading. FFmpeg finds it via pkg-config (module xeve, header xeve.h under include/xeve,
	// symbol xeve_encode).
	"xeve": {
		libraryId:   "xeve",
		displayName: "XEVE",
		trackName:   LibraryTrackInternal,
		method:      PreparationMethodInternalSource,
		buildSystem: BuildSystemCMake,
		generatedSourceFiles: []GeneratedSourceFile{
			{Path: "version.txt", Lines: []string{"v0.4.3"}},
		},
		sourcePatches: []LibrarySourcePatch{
			{
				File:    "CMakeLists.txt",
				Find:    `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Werror -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`,
				Replace: `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`,
			},
		},
		pkgConfigName:            "xeve",
		pkgConfigLibsLine:        "-L${libdir}/xeve -l:libxeve.a -lm -lpthread",
		verifyHeaderRelativePath: "xeve/xeve.h",
		verifyLibStem:            "xeve",
	},
	// libplacebo is normally the MSYS2 native package (Native track), correct for FFmpeg >= 6.1:
	// those guard the pl_peak_detect_params.overshoot_margin and pl_render_params.force_icc_lut
	// references in vf_libplacebo.c behind PL_API_VER, so they compile against the current
	// libplacebo (API 7.x). FFmpeg 5.1's vf_libplacebo.c uses both fields unconditionally, but
	// libplacebo removed overshoot_margin at PL_API_VER 256 and force_icc_lut later, so the
	// current package fails to compile libavfilter/vf_libplacebo.c ("no member named
	// overshoot_margin"/"force_icc_lut"). FFmpeg's pkg-config floor (libplacebo >= 4.192.0) is a
	// minimum not a maximum, so configure passes and the break only surfaces at make. The 5.1 line
	// marks libplacebo sourceBuild and the planner flips it to this Internal recipe, building the
	// pinned libplacebo 4.192.0 (the 5.1 floor, API 192: both fields present) with meson.
	//
	// The native libplacebo catalog row also carries vulkan-loader/vulkan-headers (FFmpeg's
	// --enable-vulkan and libplacebo both need them); the Native->Internal flip drops the row's
	// packages, so they are reinstated here as buildDependencyPackages alongside the shader
	// compiler (shaderc) and colour-management (lcms2) libplacebo links — they install into the
	// prefix before the build and remain installed for FFmpeg configure. demos=false avoids the
	// only git submodule (demos/3rdparty/nuklear, absent from the archive); d3d11/opengl/glslang
	// are disabled to keep the dependency surface to vulkan+shaderc+lcms2. libplacebo is C but
	// shaderc is C++, so the C++ runtime is appended after the archive for correct GNU static link
	// order. FFmpeg finds it via pkg-config (module libplacebo, header libplacebo/renderer.h).
	"libplacebo": {
		libraryId:               "libplacebo",
		displayName:             "libplacebo",
		trackName:               LibraryTrackInternal,
		method:                  PreparationMethodInternalSource,
		buildSystem:             BuildSystemMeson,
		buildDependencyPackages: []string{"meson", "ninja", "vulkan-headers", "vulkan-loader", "shaderc", "lcms2"},
		cmakeOptions: []string{
			"-Ddemos=false",
			"-Dtests=false",
			"-Dopengl=disabled",
			"-Dd3d11=disabled",
			"-Dglslang=disabled",
			"-Dshaderc=enabled",
			"-Dvulkan=enabled",
			"-Dlcms=enabled",
		},
		pkgConfigName:            "libplacebo",
		pkgConfigAppendLibs:      []string{"stdc++"},
		verifyHeaderRelativePath: "libplacebo/renderer.h",
		verifyLibStem:            "placebo",
	},
	"tensorflow": {
		libraryId:                "tensorflow",
		displayName:              "TensorFlow",
		trackName:                LibraryTrackExternal,
		method:                   PreparationMethodExternalImport,
		importIncludeSubdir:      "include",
		importLibSubdir:          "lib",
		verifyHeaderRelativePath: "tensorflow/c/c_api.h",
		verifyLibStem:            "tensorflow",
	},
}

// preparationForLibrary returns the flattened recipe for a non-Native library, pairing
// its item spec with the version resolved from library-sources.json for the FFmpeg
// release being built. Returns false when the library has no item spec or no recorded
// version (it then stays blocked by the planner).
func preparationForLibrary(library LibraryChoice, ffmpegVersion string) (LibraryPreparation, bool) {
	if library.TrackName == LibraryTrackNative {
		return LibraryPreparation{}, false
	}
	item, exists := libraryItemSpecs[library.LibraryId]
	if !exists {
		return LibraryPreparation{}, false
	}
	source, resolved := librarysources.ResolveLibrarySource(ffmpegVersion, library.LibraryId)
	if !resolved {
		return LibraryPreparation{}, false
	}
	return buildPreparation(item, source), true
}

// partitionNonNativeLibraries splits selected non-Native libraries into those that have
// an implemented recipe with a resolvable version (preparable) and those that do not
// (still blocked), for the FFmpeg release being built.
func partitionNonNativeLibraries(libraries []LibraryChoice, ffmpegVersion string) (preparable []LibraryPreparation, blocked []LibraryChoice) {
	for _, library := range libraries {
		if library.TrackName == LibraryTrackNative {
			continue
		}
		if recipe, exists := preparationForLibrary(library, ffmpegVersion); exists {
			preparable = append(preparable, recipe)
		} else {
			blocked = append(blocked, library)
		}
	}
	return preparable, blocked
}

// prefixPreparationBuildDependencyPackages turns each recipe's profile-independent build
// dependency suffixes (e.g. "python") into fully qualified MSYS2 mingw package names for
// the selected shell profile (e.g. "mingw-w64-ucrt-x86_64-python"), in place. A recipe
// with no build dependencies is left untouched.
func prefixPreparationBuildDependencyPackages(preparations []LibraryPreparation, windowsShellProfileName string) {
	packagePrefix := packagePrefixForShellProfile(windowsShellProfileName)
	for i := range preparations {
		if len(preparations[i].BuildDependencyPackages) == 0 {
			preparations[i].BuildDependencyPackages = nil
			continue
		}
		prefixed := make([]string, 0, len(preparations[i].BuildDependencyPackages))
		for _, packageSuffix := range preparations[i].BuildDependencyPackages {
			prefixed = append(prefixed, packagePrefix+"-"+packageSuffix)
		}
		preparations[i].BuildDependencyPackages = prefixed
	}
}

func preparationsForTrack(preparations []LibraryPreparation, trackName LibraryTrackName) []LibraryPreparation {
	tracked := []LibraryPreparation{}
	for _, preparation := range preparations {
		if preparation.TrackName == trackName {
			tracked = append(tracked, preparation)
		}
	}
	return tracked
}
