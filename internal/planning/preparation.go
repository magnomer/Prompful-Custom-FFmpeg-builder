package planning

import "promptfulcustomffmpegbuilder/versions/shared"

// Non-Native library preparation is split deliberately:
//
//	(1) Source pins live in the embedded library source catalog. They identify the
//	    archive to download and verify; they do not describe build manipulation.
//	(2) Version/library manipulation is executable Go code under /versions/x.x.x/.
//	    Those hooks change the build plan by calling explicit operations such as
//	    LCmakeOptionAdd, LPreparationModificationAdd, LLibraryLineOverride, or LInstallPrivateUse.
//	(3) This package only dispatches the hook and projects the result into the
//	    script-facing LLibraryPreparation structure.

// LLibraryMethod names how a non-Native library is made available to FFmpeg
// configure before the build runs.
type LLibraryMethod string

const (
	// LLibraryInternalMethod builds the library from a verified upstream
	// source archive inside the private MSYS2 environment and installs it into the
	// selected profile prefix.
	LLibraryInternalMethod LLibraryMethod = "internal-source-build"
	// LLibraryExternalMethod downloads a verified vendor binary archive and
	// imports its headers and libraries into the selected profile prefix.
	LLibraryExternalMethod LLibraryMethod = "external-vendor-import"
)

// LSourcePatch is one exact full-line replacement applied to the extracted source
// tree before the build runs. It exists for recipes that must work around an upstream
// portability bug that no build flag can fix (e.g. a duplicate overload that only collides
// under Windows/LLP64). File is relative to the source root; Find is matched as a whole
// line and replaced with Replace. The generated step fails if Find is absent, so a
// re-pinned release that changed the file is caught rather than silently built unpatched.
type LSourcePatch struct {
	File    string `json:"file"`
	Find    string `json:"find"`
	Replace string `json:"replace"`
}

// LFileGenerated is a file written into the extracted source tree before the build,
// for a file a release tarball omits because upstream generates it from a .git checkout
// (e.g. libvmaf's vcs_version.h). Path is relative to the source root; Lines are the file's
// lines, written verbatim.
type LFileGenerated struct {
	Path  string   `json:"path"`
	Lines []string `json:"lines"`
}

// LLibraryPatchEntry replaces the Libs line in an installed pkg-config module
// that is not the recipe's primary module.
type LLibraryPatchEntry struct {
	Module   string `json:"module"`
	LibsLine string `json:"libsLine"`
}

// LLibraryBuildsystem names the build system of an Internal-track source recipe. The
// generic source-build generator dispatches on this, so adding an autotools/make
// library is implemented in one place without disturbing existing cmake recipes.
type LLibraryBuildsystem string

const (
	LBuildSystemCmake LLibraryBuildsystem = "cmake"
	// LConfigureMakeSystem covers projects with a ./configure script (standard
	// autotools or x264/davs2-style custom configure) followed by make + make install.
	LConfigureMakeSystem LLibraryBuildsystem = "configure-make"
	LMakeBuildSystem     LLibraryBuildsystem = "make"
	// LBuildSystemMeson covers projects configured with `meson setup` and built with ninja
	// (e.g. libvmaf). The meson `-Dname=value` options reuse the CMakeOptions field, since
	// both build systems share that option syntax and the same safe-option validation.
	LBuildSystemMeson LLibraryBuildsystem = "meson"
)

// LLibrarySpec is layer (2): version-independent, item-specific.
type LLibrarySpec struct {
	LLibraryId   string
	LNameDisplay string
	LTrackName   LLibraryTrack
	LMethod      LLibraryMethod

	// LPackageBuildDependency lists MSYS2 mingw package suffixes (e.g. "python") that
	// must be installed into the build environment before this library's source build
	// runs, for libraries whose build system find_package()s another library at configure
	// time. They are profile-prefixed at plan time (mingw-w64-<profile>-x86_64-<suffix>),
	// so the recipe stays profile-independent. Most recipes need none.
	LPackageBuildDependency []string

	// LMsysBuildDependency lists MSYS2 base ("msys" repo) package names installed
	// verbatim ??not profile-prefixed ??before the build runs. These are build-time tools
	// that live in /usr/bin rather than the mingw prefix, e.g. the autotools an autogen
	// recipe needs (autoconf-wrapper, automake-wrapper, libtool); modern base-devel no longer
	// pulls them. Most recipes need none.
	LMsysBuildDependency []string

	// LCFlag are extra C compiler flags exported as CFLAGS for the build. Used to demote a
	// GCC-14 hard error back to a warning for an older C library that predates it (e.g. libvmaf
	// 1.5.2's implicit function declarations). Honored by the meson generator. Most recipes
	// need none.
	LCFlag []string

	// internal source build (cmake)
	LBuildSystem       LLibraryBuildsystem
	LCmakeOption       []string // intrinsic to the library, not to a version
	LCmakeBuildTargets []string // named targets to build; empty = the default target

	// internal source build (configure-make)
	LConfigureSubdirectory string   // dir holding ./configure, relative to source root (e.g. "build/linux")
	LConfigureOption       []string // intrinsic configure flags (--prefix is added automatically)
	LMakeBuildTargets   []string // make targets for the build step; empty = default target
	LMakeInstallTargets []string // make targets for the install step; empty = "install"
	// LAutogenCommand bootstraps an autotools project whose tag tarball ships no generated
	// ./configure (only configure.ac + autogen.sh). The generator runs autogen.sh (or
	// autoreconf -fiv) at the source root before ./configure.
	LAutogenCommand bool

	// internal source build (plain make): a Makefile-only project with no lib-only install
	// target installs by copying these source-relative artifacts into the prefix ??each
	// header into include/ (by basename) and the static archive into lib/. LMakeVariables are
	// NAME=VALUE command-line overrides that skip the makefile's own (e.g. SDL_CFLAGS=).
	LMakeVariables      []string
	LMakeInstallHeaders []string
	LLibraryStaticFile  string

	// external vendor import
	LSubdirectoryImportInclude string
	LSubdirectoryImportLibrary string

	// LNamePkgconfig is the installed pkg-config module (e.g. "lcevc_dec", "libvvenc") ??	// the same module FFmpeg's require_pkg_config looks up. It is used both to validate
	// the pinned version against FFmpeg's required minimum (the preflight safety net) and,
	// when LLibraryExtraLine is set, to patch that .pc. LLibraryExtraLine lists bare
	// link libraries (e.g. "stdc++", "m") appended to the .pc's Libs line after install:
	// some CMake projects emit a static .pc that lists the C++/math runtime BEFORE their
	// own static archives, so GNU ld discards it too early and the link fails with
	// undefined std::/operator new references; appending the runtime at the END fixes the
	// static link order. Leave LNamePkgconfig empty for libraries FFmpeg does not
	// version-check via pkg-config (header-only or vendor-imported).
	LNamePkgconfig    string
	LLibraryExtraLine []string
	LCompilerFlagLine []string

	// LLinePkgconfigLibs, when set, replaces the installed .pc's whole "Libs:" line value.
	// Needed when a bare -l<name> in the .pc would resolve to a same-named shared import
	// library (.dll.a) that shadows this recipe's own static archive in a shared prefix:
	// forcing -l:lib<name>.a (or an absolute archive path) makes the link pick the static
	// archive. May reference ${libdir}.
	LLinePkgconfigLibs string

	// LPrivateInstallPrefix installs this library into its own per-library prefix
	// (<profile>/opt/customffmpeg/<LNamePkgconfig>) instead of the shared MSYS2 prefix,
	// and the configure step prepends that prefix's pkgconfig dir to PKG_CONFIG_PATH.
	// Needed when a library ships archives whose base names collide with a different
	// package's in the shared prefix ??e.g. LibreSSL's libtls brings its own libssl.a /
	// libcrypto.a, which the openssl package (pulled in by libssh/srt/rabbitmq) also owns
	// and would otherwise win. Pair with a LLinePkgconfigLibs of absolute ${libdir}/lib*.a
	// archive paths so the link binds this prefix's archives regardless of -L ordering;
	// the installed .pc's Requires/Requires.private are stripped so no bare -l<name> from a
	// transitive module re-introduces the shared-prefix archive.
	LPrivateInstallPrefix bool

	// LPatchSource are exact full-line edits applied to the extracted source before the
	// build runs, for upstream portability bugs no build flag can fix. Most recipes need none.
	LPatchSource []LSourcePatch

	// LFileGeneratedSource are files written into the extracted source before the build, for a
	// file a release tarball omits because upstream generates it from a .git checkout (e.g.
	// libvmaf's vcs_version.h). Most recipes need none.
	LFileGeneratedSource []LFileGenerated

	// verification (both methods)
	LHeaderPathRelative      string
	LLibraryStemVerification string
}

// LLibraryPreparation is the flattened, plan-facing projection of an item paired with its
// resolved version. It carries no source-directory name: the generic layer auto-detects
// the extracted root, so nothing here is tied to a release beyond the resolved URL/hash.
type LLibraryPreparation struct {
	LibraryId   string         `json:"libraryId"`
	DisplayName string         `json:"displayName"`
	TrackName   LLibraryTrack  `json:"trackName"`
	Method      LLibraryMethod `json:"method"`

	BuildSystem LLibraryBuildsystem `json:"buildSystem"`

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
	LArchiveKind        string `json:"archiveFormat"`

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

	PkgConfigName            string               `json:"pkgConfigName,omitempty"`
	PkgConfigAppendLibs      []string             `json:"pkgConfigAppendLibs,omitempty"`
	PkgConfigAppendCFlags    []string             `json:"pkgConfigAppendCFlags,omitempty"`
	PkgConfigLibsLine        string               `json:"pkgConfigLibsLine,omitempty"`
	PkgConfigLibsLinePatches []LLibraryPatchEntry `json:"pkgConfigLibsLinePatches,omitempty"`
	PrivatePrefixInstall     bool                 `json:"privatePrefixInstall,omitempty"`

	VerifyHeaderRelativePath string `json:"verifyHeaderRelativePath"`
	VerifyLibStem            string `json:"verifyLibStem"`

	SourcePatches []LSourcePatch `json:"sourcePatches,omitempty"`

	GeneratedSourceFiles []LFileGenerated `json:"generatedSourceFiles,omitempty"`
}

// LPreparationBuildCreate projects an item plus its resolved version source into a plan-facing
// LLibraryPreparation.
func LPreparationBuildCreate(item LLibrarySpec, source LLibrarySourcePin) LLibraryPreparation {
	LCmakeOption := append(append([]string{}, item.LCmakeOption...), source.ExtraCMakeOptions...)
	if len(LCmakeOption) == 0 {
		LCmakeOption = nil
	}
	return LLibraryPreparation{
		LibraryId:                   item.LLibraryId,
		DisplayName:                 item.LNameDisplay,
		TrackName:                   item.LTrackName,
		Method:                      item.LMethod,
		BuildSystem:                 item.LBuildSystem,
		CFlags:                      append([]string{}, item.LCFlag...),
		Version:                     source.Version,
		BuildDependencyPackages:     append([]string{}, item.LPackageBuildDependency...),
		MsysBuildDependencyPackages: append([]string{}, item.LMsysBuildDependency...),
		ArchiveUrl:                  source.Url,
		ArchiveSha256Hash:           source.Sha256,
		AllowedDownloadHost:         source.Host,
		LArchiveKind:                source.Format,
		CMakeOptions:                LCmakeOption,
		CMakeBuildTargets:           item.LCmakeBuildTargets,
		ConfigureSubdir:             item.LConfigureSubdirectory,
		ConfigureOptions:            item.LConfigureOption,
		MakeBuildTargets:            item.LMakeBuildTargets,
		MakeInstallTargets:          item.LMakeInstallTargets,
		RunAutogen:                  item.LAutogenCommand,
		MakeVariables:               append([]string{}, item.LMakeVariables...),
		MakeInstallHeaderFiles:      append([]string{}, item.LMakeInstallHeaders...),
		MakeStaticLibFile:           item.LLibraryStaticFile,
		ImportIncludeSubdir:         item.LSubdirectoryImportInclude,
		ImportLibSubdir:             item.LSubdirectoryImportLibrary,
		PkgConfigName:               item.LNamePkgconfig,
		PkgConfigAppendLibs:         append([]string{}, item.LLibraryExtraLine...),
		PkgConfigAppendCFlags:       append([]string{}, item.LCompilerFlagLine...),
		PkgConfigLibsLine:           item.LLinePkgconfigLibs,
		PkgConfigLibsLinePatches:    nil,
		PrivatePrefixInstall:        item.LPrivateInstallPrefix,
		VerifyHeaderRelativePath:    item.LHeaderPathRelative,
		VerifyLibStem:               item.LLibraryStemVerification,
		SourcePatches:               append([]LSourcePatch{}, item.LPatchSource...),
		GeneratedSourceFiles:        append([]LFileGenerated{}, item.LFileGeneratedSource...),
	}
}

// LLibraryPreparationBuild returns the flattened preparation for a non-Native library.
// The ffmpegVersion argument is the resolved compatibility/catalog version. For
// a requested newer FFmpeg source, this can intentionally differ from the source
// archive version so the build uses the highest known compatible version/library
// hook while still compiling the requested FFmpeg source. Missing manipulation
// for that resolved compatibility version is reported as unavailable.
func LLibraryPreparationBuild(library LLibraryChoice, ffmpegVersion string) (LLibraryPreparation, bool) {
	if library.TrackName == LLibraryTrackNative {
		return LLibraryPreparation{}, false
	}
	source, sourceResolved := LSourceSpecificResolve(ffmpegVersion, library.LibraryId)
	if !sourceResolved {
		return LLibraryPreparation{}, false
	}
	registry, err := LWorkRegistryLoad()
	if err != nil {
		return LLibraryPreparation{}, false
	}
	implementation, exists := registry.LWorkLibraryResolve(ffmpegVersion, library.LibraryId)
	if !exists || implementation.Manipulator == nil {
		return LLibraryPreparation{}, false
	}
	plan := shared.LPreparationPlanCreate(ffmpegVersion, library.LibraryId, implementation.Work.GoFilePath)
	implementation.Manipulator(plan)
	return LPreparationPlanResolve(*plan, source), true
}

func LPreparationPlanResolve(plan shared.LPreparationPlan, source LLibrarySourcePin) LLibraryPreparation {
	cmakeOptions := append([]string{}, plan.CMakeOptions...)
	cmakeOptions = append(cmakeOptions, source.ExtraCMakeOptions...)
	if len(cmakeOptions) == 0 {
		cmakeOptions = nil
	}
	return LLibraryPreparation{
		LibraryId:                   plan.LibraryId,
		DisplayName:                 plan.DisplayName,
		TrackName:                   LLibraryTrack(plan.TrackName),
		Method:                      LLibraryMethod(plan.Method),
		BuildSystem:                 LLibraryBuildsystem(plan.BuildSystem),
		CFlags:                      append([]string{}, plan.CFlags...),
		Version:                     source.Version,
		BuildDependencyPackages:     append([]string{}, plan.BuildDependencyPackages...),
		MsysBuildDependencyPackages: append([]string{}, plan.MsysBuildDependencyPackages...),
		ArchiveUrl:                  source.Url,
		ArchiveSha256Hash:           source.Sha256,
		AllowedDownloadHost:         source.Host,
		LArchiveKind:                source.Format,
		CMakeOptions:                cmakeOptions,
		CMakeBuildTargets:           append([]string{}, plan.CMakeBuildTargets...),
		ConfigureSubdir:             plan.ConfigureSubdir,
		ConfigureOptions:            append([]string{}, plan.ConfigureOptions...),
		MakeBuildTargets:            append([]string{}, plan.MakeBuildTargets...),
		MakeInstallTargets:          append([]string{}, plan.MakeInstallTargets...),
		RunAutogen:                  plan.RunAutogen,
		MakeVariables:               append([]string{}, plan.MakeVariables...),
		MakeInstallHeaderFiles:      append([]string{}, plan.MakeInstallHeaderFiles...),
		MakeStaticLibFile:           plan.MakeStaticLibFile,
		ImportIncludeSubdir:         plan.ImportIncludeSubdir,
		ImportLibSubdir:             plan.ImportLibSubdir,
		PkgConfigName:               plan.PkgConfigName,
		PkgConfigAppendLibs:         append([]string{}, plan.PkgConfigAppendLibs...),
		PkgConfigAppendCFlags:       append([]string{}, plan.PkgConfigAppendCFlags...),
		PkgConfigLibsLine:           plan.PkgConfigLibsLine,
		PkgConfigLibsLinePatches:    LLineModificationRead(plan.PkgConfigLibsLinePatches),
		PrivatePrefixInstall:        plan.PrivatePrefixInstall,
		VerifyHeaderRelativePath:    plan.VerifyHeaderRelativePath,
		VerifyLibStem:               plan.VerifyLibStem,
		SourcePatches:               LSourceModificationRead(plan.SourcePatches),
		GeneratedSourceFiles:        LGeneratedFileRead(plan.GeneratedSourceFiles),
	}
}

func LSourceModificationRead(patches []shared.LSourcePatchEntry) []LSourcePatch {
	converted := make([]LSourcePatch, 0, len(patches))
	for _, patch := range patches {
		converted = append(converted, LSourcePatch{File: patch.File, Find: patch.Find, Replace: patch.Replace})
	}
	return converted
}

func LLineModificationRead(patches []shared.LPackagePatchEntry) []LLibraryPatchEntry {
	converted := make([]LLibraryPatchEntry, 0, len(patches))
	for _, patch := range patches {
		converted = append(converted, LLibraryPatchEntry{Module: patch.Module, LibsLine: patch.LibsLine})
	}
	return converted
}

func LGeneratedFileRead(files []shared.LGeneratedFile) []LFileGenerated {
	converted := make([]LFileGenerated, 0, len(files))
	for _, file := range files {
		converted = append(converted, LFileGenerated{Path: file.Path, Lines: append([]string{}, file.Lines...)})
	}
	return converted
}

// LLibraryPartitionCreate splits selected non-Native libraries into those that have
// an implemented recipe with a resolvable version (preparable) and those that do not
// (still blocked), for the FFmpeg release being built.
func LLibraryPartitionCreate(libraries []LLibraryChoice, ffmpegVersion string) (preparable []LLibraryPreparation, blocked []LLibraryChoice) {
	for _, library := range libraries {
		if library.TrackName == LLibraryTrackNative {
			continue
		}
		if recipe, exists := LLibraryPreparationBuild(library, ffmpegVersion); exists {
			preparable = append(preparable, recipe)
		} else {
			blocked = append(blocked, library)
		}
	}
	return preparable, blocked
}

// LDependencyPrefixGet turns each recipe's profile-independent build
// dependency suffixes (e.g. "python") into fully qualified MSYS2 mingw package names for
// the selected shell profile (e.g. "mingw-w64-ucrt-x86_64-python"), in place. A recipe
// with no build dependencies is left untouched.
func LDependencyPrefixGet(preparations []LLibraryPreparation, windowsShellProfileName string) {
	packagePrefix := LPackageProfileResolve(windowsShellProfileName)
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

func LLibraryTrackFilter(preparations []LLibraryPreparation, LTrackName LLibraryTrack) []LLibraryPreparation {
	tracked := []LLibraryPreparation{}
	for _, preparation := range preparations {
		if preparation.TrackName == LTrackName {
			tracked = append(tracked, preparation)
		}
	}
	return tracked
}
