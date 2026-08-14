package scripting

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

// LLibraryBuildSpec is the scripting-layer view of one non-Native library preparation.
// It is mapped from planning.LLibraryPreparation by the program layer (scripting cannot
// import planning without an import cycle). The generated scripts operate only on the
// already-downloaded, already-extracted archive contents in the script working
// directory and install into the selected MSYS2 profile prefix; no URL or raw path
// from outside the workspace ever enters the shell.
type LLibraryBuildSpec struct {
	LibraryId   string
	DisplayName string
	// BuildSystem selects the Internal-track source-build generator ("cmake",
	// "autotools", "make"). Empty is treated as cmake. Ignored for external imports.
	BuildSystem string
	// CFlags are extra C compiler flags exported as CFLAGS for the build (e.g. demoting a GCC-14
	// hard error to a warning for an older C library that predates it). Honored by the meson
	// generator. Each is validated against LCompilerFlagPattern.
	CFlags             []string
	CMakeOptions       []string
	CMakeBuildTargets  []string
	ConfigureSubdir    string
	ConfigureOptions   []string
	MakeBuildTargets   []string
	MakeInstallTargets []string
	// RunAutogen bootstraps an autotools project that ships no generated ./configure
	// (only configure.ac + autogen.sh, as GitHub tag tarballs do). When set, the
	// configure-make generator runs autoreconf -fiv at the source root before
	// ./configure, falling back to the project's autogen.sh only if autoreconf is
	// unavailable. The autotools (autoconf/automake/libtool) come from the base-devel
	// toolchain already installed.
	RunAutogen bool
	// MakeVariables are NAME=VALUE assignments passed on the make command line (e.g.
	// "SDL_CFLAGS="). A command-line assignment overrides ??and skips evaluation of ??the
	// makefile's own assignment, which is how a recipe neutralises an optional
	// $(shell pkg-config ...) probe whose error text would otherwise poison CFLAGS.
	MakeVariables []string
	// MakeInstallHeaderFiles and MakeStaticLibFile drive the plain-"make" build system's
	// custom install. A bare Makefile project (e.g. quirc) often has no lib-only install
	// target ??its `make install` also builds demos that pull extra deps ??so the generator
	// installs by copying these source-relative artifacts into the prefix instead: each
	// header file into include/ (by basename) and the static archive into lib/.
	MakeInstallHeaderFiles []string
	MakeStaticLibFile      string
	ImportIncludeSubdir    string
	ImportLibSubdir        string
	PkgConfigName          string
	PkgConfigAppendLibs    []string
	PkgConfigAppendCFlags  []string
	// PkgConfigLibsLine, when set, replaces the installed .pc's entire "Libs:" line value.
	// Used when -l<name> would resolve to a same-named shared import library (.dll.a) that
	// shadows this recipe's own static archive in a shared prefix; forcing -l:lib<name>.a
	// (or an absolute archive path) makes the link pick the static archive instead. May
	// reference ${libdir}. Mutually independent of PkgConfigAppendLibs.
	PkgConfigLibsLine string
	// PkgConfigLibsLinePatches applies the same Libs-line override to additional
	// installed modules, for libraries whose FFmpeg configure checks more than one .pc.
	PkgConfigLibsLinePatches []LLibraryPatchEntry
	// PrivatePrefixInstall installs the library into its own per-library prefix
	// (<profile>/LPrivateInstallSubdirectory/<PkgConfigName>) instead of the shared MSYS2
	// prefix, and strips the installed .pc's Requires/Requires.private so no transitive
	// module re-adds a shared-prefix archive. Used to isolate a library whose archive base
	// names collide with another package's (e.g. LibreSSL libtls vs the openssl package).
	PrivatePrefixInstall     bool
	VerifyHeaderRelativePath string
	VerifyLibStem            string
	SourcePatches            []LSourcePatch
	// GeneratedSourceFiles are files written into the extracted source tree before configure,
	// for recipes whose build expects a file that the release tarball omits because upstream
	// generates it from a .git checkout (e.g. libvmaf's vcs_version.h, emitted by meson vcs_tag
	// only when a .git dir is present). Supplies what no build flag or single-line patch can.
	GeneratedSourceFiles []LFileGenerated
}

// LFileGenerated is one file the recipe writes into the extracted source tree before
// configure. Path is relative to the source root; Lines are the file's lines, written
// verbatim. Path is validated as a safe relative path and each line must contain no single
// quote or newline, so the lines are safe to single-quote into the generated printf.
type LFileGenerated struct {
	Path  string
	Lines []string
}

type LLibraryPatchEntry struct {
	Module   string
	LibsLine string
}

// LPrivateInstallSubdirectory is the prefix-relative directory under the MSYS2 profile
// prefix that holds per-library private installs (see LLibraryBuildSpec.PrivatePrefixInstall).
// A library installs into <profile>/LPrivateInstallSubdirectory/<PkgConfigName>. Shared by
// the build-script generator (which installs there) and the program (which prepends that
// prefix's pkgconfig dir to the FFmpeg configure PKG_CONFIG_PATH); keep the two in sync.
const LPrivateInstallSubdirectory = "opt/customffmpeg"

// LPrivateDirectoryGet returns the unix pkgconfig directory of a privately-installed
// library, given the MSYS2 profile's unix prefix (e.g. "/ucrt64") and the library's
// pkg-config module name. The configure step adds this to PKG_CONFIG_PATH.
func LPrivateDirectoryGet(profileUnixPrefix string, LNamePkgconfig string) string {
	return profileUnixPrefix + "/" + LPrivateInstallSubdirectory + "/" + LNamePkgconfig + "/lib/pkgconfig"
}

// LSourcePatch is one exact full-line replacement applied to the extracted source
// tree before configure/build, for recipes that must work around an upstream portability
// bug that no build flag can fix. File is relative to the source root; Find is matched as
// a whole line and replaced with Replace. Find/Replace must contain no single quote,
// backslash, or newline so they are safe to single-quote into the generated awk command.
type LSourcePatch struct {
	File    string
	Find    string
	Replace string
}

var LLibraryPathPattern = regexp.MustCompile(`^[A-Za-z0-9_+-]+$`)
var LLibraryHeaderPattern = regexp.MustCompile(`^[A-Za-z0-9_./+-]+$`)
var LCMakeOptionPattern = regexp.MustCompile(`^-D[A-Za-z0-9_]+=[A-Za-z0-9_./:+-]*$`)
var LCMakeTargetPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var LConfigureOptionPattern = regexp.MustCompile(`^--[A-Za-z0-9][A-Za-z0-9=._/+-]*$`)

// LCompilerFlagPattern matches a single C compiler flag exported as CFLAGS (e.g.
// -Wno-error=implicit-function-declaration). No spaces or shell metacharacters, so the joined
// flags are safe to LTextInterpolate into the CFLAGS= assignment in the generated script.
var LCompilerFlagPattern = regexp.MustCompile(`^-[A-Za-z0-9][A-Za-z0-9=._+-]*$`)
var LMakeTargetPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var LMakeVariablePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=[A-Za-z0-9_./:= +-]*$`)
var LPackageLinePattern = regexp.MustCompile(`^[A-Za-z0-9_:./${} +-]+$`)

// LPackageDirectoryPattern matches a unix pkgconfig directory under the MSYS2 profile prefix
// (e.g. /ucrt64/opt/customffmpeg/libtls/lib/pkgconfig). No spaces, no shell metacharacters,
// so it is safe to LTextInterpolate into the exported PKG_CONFIG_PATH.
var LPackageDirectoryPattern = regexp.MustCompile(`^/[A-Za-z0-9_./-]+$`)

func LLibrarySpecValidate(spec LLibraryBuildSpec, requireImportSubdirs bool) error {
	if !LLibraryPathPattern.MatchString(spec.LibraryId) {
		return fmt.Errorf("library preparation id contains unsafe characters: %s", spec.LibraryId)
	}
	if spec.VerifyLibStem != "" && !LLibraryPathPattern.MatchString(spec.VerifyLibStem) {
		return fmt.Errorf("library preparation lib stem contains unsafe characters: %s", spec.VerifyLibStem)
	}
	if spec.VerifyHeaderRelativePath == "" {
		return errors.New("library preparation verify header path is empty")
	}
	if !LLibraryHeaderPattern.MatchString(spec.VerifyHeaderRelativePath) || strings.Contains(spec.VerifyHeaderRelativePath, "..") {
		return fmt.Errorf("library preparation verify header path is unsafe: %s", spec.VerifyHeaderRelativePath)
	}
	for _, cmakeOption := range spec.CMakeOptions {
		if !LCMakeOptionPattern.MatchString(cmakeOption) {
			return fmt.Errorf("library preparation cmake option is unsafe: %s", cmakeOption)
		}
	}
	for _, buildTarget := range spec.CMakeBuildTargets {
		if !LCMakeTargetPattern.MatchString(buildTarget) {
			return fmt.Errorf("library preparation cmake build target is unsafe: %s", buildTarget)
		}
	}
	if spec.PkgConfigName != "" && !LLibraryPathPattern.MatchString(spec.PkgConfigName) {
		return fmt.Errorf("library preparation pkg-config name contains unsafe characters: %s", spec.PkgConfigName)
	}
	if spec.PrivatePrefixInstall {
		// A private install is keyed by its pkg-config module name (the private prefix
		// directory and the .pc patched there) and is only wired into the CMake generator.
		if spec.PkgConfigName == "" {
			return errors.New("library preparation private-prefix install requires a pkg-config name")
		}
		if spec.BuildSystem != "" && spec.BuildSystem != "cmake" {
			return fmt.Errorf("library preparation private-prefix install is only supported for the cmake build system, not %q", spec.BuildSystem)
		}
	}
	for _, patch := range spec.SourcePatches {
		if patch.File == "" || !LLibraryHeaderPattern.MatchString(patch.File) || strings.Contains(patch.File, "..") {
			return fmt.Errorf("library preparation source patch file is unsafe: %q", patch.File)
		}
		if patch.Find == "" {
			return fmt.Errorf("library preparation source patch find is empty for %q", patch.File)
		}
		if strings.ContainsAny(patch.Find, "\n\r") || strings.ContainsAny(patch.Replace, "\n\r") {
			return fmt.Errorf("library preparation source patch contains unsafe characters for %q", patch.File)
		}
	}
	for _, generated := range spec.GeneratedSourceFiles {
		if generated.Path == "" || !LLibraryHeaderPattern.MatchString(generated.Path) || strings.Contains(generated.Path, "..") {
			return fmt.Errorf("library preparation generated source file path is unsafe: %q", generated.Path)
		}
		for _, line := range generated.Lines {
			if strings.ContainsAny(line, "'\n\r") {
				return fmt.Errorf("library preparation generated source file %q has unsafe line content", generated.Path)
			}
		}
	}
	for _, appendLib := range spec.PkgConfigAppendLibs {
		if !LLibraryPathPattern.MatchString(appendLib) {
			return fmt.Errorf("library preparation pkg-config append lib is unsafe: %s", appendLib)
		}
	}
	for _, appendCFlag := range spec.PkgConfigAppendCFlags {
		if !LCompilerFlagPattern.MatchString(appendCFlag) {
			return fmt.Errorf("library preparation pkg-config append cflag is unsafe: %s", appendCFlag)
		}
	}
	if spec.PkgConfigLibsLine != "" && !LPackageLinePattern.MatchString(spec.PkgConfigLibsLine) {
		return fmt.Errorf("library preparation pkg-config Libs override is unsafe: %s", spec.PkgConfigLibsLine)
	}
	for _, patch := range spec.PkgConfigLibsLinePatches {
		if patch.Module == "" || !LLibraryPathPattern.MatchString(patch.Module) {
			return fmt.Errorf("library preparation pkg-config override module is unsafe: %s", patch.Module)
		}
		if patch.LibsLine == "" || !LPackageLinePattern.MatchString(patch.LibsLine) {
			return fmt.Errorf("library preparation pkg-config Libs override is unsafe for %s: %s", patch.Module, patch.LibsLine)
		}
	}
	if spec.ConfigureSubdir != "" && (!LLibraryHeaderPattern.MatchString(spec.ConfigureSubdir) || strings.Contains(spec.ConfigureSubdir, "..")) {
		return fmt.Errorf("library preparation configure subdir is unsafe: %s", spec.ConfigureSubdir)
	}
	for _, configureOption := range spec.ConfigureOptions {
		if !LConfigureOptionPattern.MatchString(configureOption) {
			return fmt.Errorf("library preparation configure option is unsafe: %s", configureOption)
		}
	}
	for _, compilerFlag := range spec.CFlags {
		if !LCompilerFlagPattern.MatchString(compilerFlag) {
			return fmt.Errorf("library preparation cflag is unsafe: %s", compilerFlag)
		}
	}
	for _, makeTarget := range append(append([]string{}, spec.MakeBuildTargets...), spec.MakeInstallTargets...) {
		if !LMakeTargetPattern.MatchString(makeTarget) {
			return fmt.Errorf("library preparation make target is unsafe: %s", makeTarget)
		}
	}
	for _, headerFile := range spec.MakeInstallHeaderFiles {
		if headerFile == "" || !LLibraryHeaderPattern.MatchString(headerFile) || strings.Contains(headerFile, "..") {
			return fmt.Errorf("library preparation make install header file is unsafe: %q", headerFile)
		}
	}
	if spec.MakeStaticLibFile != "" && (!LLibraryHeaderPattern.MatchString(spec.MakeStaticLibFile) || strings.Contains(spec.MakeStaticLibFile, "..")) {
		return fmt.Errorf("library preparation make static lib file is unsafe: %q", spec.MakeStaticLibFile)
	}
	for _, makeVariable := range spec.MakeVariables {
		if !LMakeVariablePattern.MatchString(makeVariable) {
			return fmt.Errorf("library preparation make variable is unsafe: %q", makeVariable)
		}
	}
	if requireImportSubdirs {
		for _, subdir := range []string{spec.ImportIncludeSubdir, spec.ImportLibSubdir} {
			if subdir == "" || !LLibraryHeaderPattern.MatchString(subdir) || strings.Contains(subdir, "..") {
				return fmt.Errorf("library preparation import subdir is unsafe: %q", subdir)
			}
		}
	}
	return nil
}
