package scripting

import (
	"fmt"
	"strings"
)

// LScriptInternalCreate builds an Internal-track library from its
// verified, already-extracted upstream source (the script working directory) and
// installs it into the selected MSYS2 profile prefix, then verifies the installed
// header and link library exist. It dispatches on the recipe's build system so a new
// build system is added in exactly one place without touching existing recipes.
func LScriptInternalCreate(spec LLibraryBuildSpec) ([]string, error) {
	switch spec.BuildSystem {
	case "", "cmake":
		return LCmakeScriptCreate(spec)
	case "configure-make":
		return LConfigureMakeCreate(spec)
	case "make":
		return LMakeScriptCreate(spec)
	case "meson":
		return LMesonScriptCreate(spec)
	default:
		return nil, fmt.Errorf("unknown internal-track build system %q for %s", spec.BuildSystem, spec.LibraryId)
	}
}

// LMesonScriptCreate builds a library configured with `meson setup` and built
// with ninja (e.g. libvmaf), installing into the selected MSYS2 profile prefix. The meson
// `-Dname=value` project options reuse the spec's CMakeOptions field (same option syntax,
// same validation); --buildtype=release and --default-library=static are intrinsic to how
// this builder produces the static archive FFmpeg links. ConfigureSubdir is the source
// directory holding meson.build when it is not the repo root (libvmaf keeps it in libvmaf/).
// Like the CMake path, meson here is the native mingw tool, so a unix prefix such as /ucrt64
// would install to the literal drive-root; the prefix is converted to its Windows form with
// cygpath (and excluded from MSYS2 path mangling) so meson installs into the real prefix.
func LMesonScriptCreate(spec LLibraryBuildSpec) ([]string, error) {
	if err := LLibrarySpecValidate(spec, false); err != nil {
		return nil, err
	}
	mesonOptionArray := make([]string, 0, len(spec.CMakeOptions))
	for _, mesonOption := range spec.CMakeOptions {
		mesonOptionArray = append(mesonOptionArray, LShellTextQuote(mesonOption))
	}
	mesonSourceDir := `"${src_dir}"`
	if spec.ConfigureSubdir != "" {
		mesonSourceDir = `"${src_dir}/` + spec.ConfigureSubdir + `"`
	}
	// Meson reads CFLAGS from the environment at setup time and appends them to the C compile
	// command. A recipe sets CFlags to demote a GCC-14 hard error back to a warning for an older
	// C library (e.g. libvmaf 1.5.2's implicit function declarations). Exported just before the
	// meson invocation so it does not leak into later steps. Each flag is validated above.
	mesonCFlagsExport := ""
	if len(spec.CFlags) > 0 {
		mesonCFlagsExport = `CFLAGS=` + LShellTextQuote(strings.Join(spec.CFlags, " ")) + ` `
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		// The meson path installs into the shared prefix only (PrivatePrefixInstall is gated
		// to cmake by LLibrarySpecValidate), so install_prefix is always the shared prefix
		// and the pkg-config patch and verification helpers resolve against it.
		`install_prefix="${profile_prefix}"`,
		"echo " + LShellTextQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`build_dir="${src_dir}/customffmpeg-internal-build"`,
		`rm -rf "${build_dir}"`,
		`for required_tool in meson ninja; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
		"meson_options=(" + strings.Join(mesonOptionArray, " ") + ")",
		`install_prefix_win="$(cygpath -m "${install_prefix}")"`,
		`echo "Meson install prefix (Windows form): ${install_prefix_win}"`,
	}
	// Write recipe-generated source files, then apply source patches, before configure. No-op
	// when none. Generated files come first so a later patch could target one if ever needed.
	scriptLines = append(scriptLines, LSourceScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptPatchCreate(spec)...)
	scriptLines = append(scriptLines,
		mesonCFlagsExport+`MSYS2_ARG_CONV_EXCL="--prefix=" meson setup "${build_dir}" `+mesonSourceDir+` --prefix="${install_prefix_win}" --buildtype=release --default-library=static "${meson_options[@]}" 2>&1`,
		// `ninja install` builds the whole `all` target before installing the install:true subset:
		// a meson library defaults to build_by_default even when install:false, so this also compiles
		// targets this builder does not install — e.g. libvmaf 1.5.2's WIP "vmaf_rc" library, whose
		// libvmaf.rc.c #includes a vcs_version.h that upstream generates from a .git checkout (absent
		// in a release tarball). The recipe supplies that header via GeneratedSourceFiles so every
		// target compiles; the installed legacy libvmaf (exports compute_vmaf) never references it.
		`ninja -C "${build_dir}" install 2>&1`,
	)
	scriptLines = append(scriptLines, LOverrideScriptCreate(spec)...)
	scriptLines = append(scriptLines, LAppendScriptCreate(spec)...)
	scriptLines = append(scriptLines, LCompilerFlagCreate(spec)...)
	scriptLines = append(scriptLines, LStripScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptVerifyCreate(spec)...)
	scriptLines = append(scriptLines, "echo "+LShellTextQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

func LCmakeScriptCreate(spec LLibraryBuildSpec) ([]string, error) {
	if err := LLibrarySpecValidate(spec, false); err != nil {
		return nil, err
	}
	cmakeOptionArray := make([]string, 0, len(spec.CMakeOptions))
	for _, cmakeOption := range spec.CMakeOptions {
		cmakeOptionArray = append(cmakeOptionArray, LShellTextQuote(cmakeOption))
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		"echo " + LShellTextQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`build_dir="${src_dir}/customffmpeg-internal-build"`,
		`rm -rf "${build_dir}"`,
		`for required_tool in cmake ninja; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
		"cmake_options=(" + strings.Join(cmakeOptionArray, " ") + ")",
		// install_prefix is the shared MSYS2 prefix by default, or a per-library private
		// prefix when the recipe isolates this library (see PrivatePrefixInstall). The
		// pkg-config patch and the install verification below both key off install_prefix.
		LScriptPrefixCreate(spec),
		// The mingw CMake is a native Windows program: a unix prefix like /ucrt64 would
		// be installed to the literal drive-root \ucrt64, not the MSYS2 prefix, so the
		// post-install verification (which uses the unix path) would not find the files.
		// Convert the prefix to its real Windows form with cygpath so CMake installs into
		// the actual install prefix.
		`install_prefix_win="$(cygpath -m "${install_prefix}")"`,
		`echo "CMake install prefix (Windows form): ${install_prefix_win}"`,
		// BUILD_SHARED_LIBS is intentionally not forced: most Internal-track decoder
		// libraries install a static lib that a static FFmpeg build links with -l<stem>.
		// A recipe that needs a shared build passes -DBUILD_SHARED_LIBS=ON via CMakeOptions.
	}
	// Write recipe-generated source files, then apply source patches, before configure (e.g.
	// to work around an upstream portability bug that no CMake flag can fix). No-op when none.
	scriptLines = append(scriptLines, LSourceScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptPatchCreate(spec)...)
	// Exported (not only passed as -D on the configure line below) so nested cmake invocations
	// the build itself spawns inherit it. OpenCV's pkg-config generation re-runs cmake in script
	// mode (cmake -P OpenCVGenPkgconfig.cmake) at build time; that child does not receive the
	// configure cache's CMAKE_POLICY_VERSION_MINIMUM, so a CMake that removed pre-3.5 policy
	// compatibility aborts on the script's old cmake_minimum_required. The env var reaches every
	// child cmake and is harmless where it is not needed.
	scriptLines = append(scriptLines, `export CMAKE_POLICY_VERSION_MINIMUM=3.5`)
	scriptLines = append(scriptLines, `cmake -S "${src_dir}" -B "${build_dir}" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="${install_prefix_win}" -DCMAKE_POLICY_VERSION_MINIMUM=3.5.0 "${cmake_options[@]}" -Wno-dev 2>&1`)
	// Build either the named targets a recipe requests (e.g. a header-only project whose
	// generated headers come from a target that is not in the default build) or, by
	// default, the default target.
	if len(spec.CMakeBuildTargets) > 0 {
		for _, buildTarget := range spec.CMakeBuildTargets {
			scriptLines = append(scriptLines, `cmake --build "${build_dir}" --target `+LShellTextQuote(buildTarget)+` 2>&1`)
		}
	} else {
		scriptLines = append(scriptLines, `cmake --build "${build_dir}" 2>&1`)
	}
	scriptLines = append(scriptLines, `cmake --install "${build_dir}" 2>&1`)
	scriptLines = append(scriptLines, LOverrideScriptCreate(spec)...)
	scriptLines = append(scriptLines, LAppendScriptCreate(spec)...)
	scriptLines = append(scriptLines, LCompilerFlagCreate(spec)...)
	scriptLines = append(scriptLines, LStripScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptVerifyCreate(spec)...)
	scriptLines = append(scriptLines, "echo "+LShellTextQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

// LConfigureMakeCreate builds a library with a ./configure script
// (standard autotools or an x264/davs2-style custom configure in a subdirectory)
// followed by make and make install, inside the MSYS2 environment. Unlike the native
// cmake path, the whole build runs under MSYS2, so the install prefix is the unix
// /ucrt64 path (the standard MSYS2 layout); the generated .pc then matches every other
// MSYS2 package and pkg-config/gcc resolve it during FFmpeg configure.
func LConfigureMakeCreate(spec LLibraryBuildSpec) ([]string, error) {
	if err := LLibrarySpecValidate(spec, false); err != nil {
		return nil, err
	}
	configureOptionArray := make([]string, 0, len(spec.ConfigureOptions))
	for _, configureOption := range spec.ConfigureOptions {
		configureOptionArray = append(configureOptionArray, LShellTextQuote(configureOption))
	}
	buildTargetArray := make([]string, 0, len(spec.MakeBuildTargets))
	for _, buildTarget := range spec.MakeBuildTargets {
		buildTargetArray = append(buildTargetArray, LShellTextQuote(buildTarget))
	}
	installTargets := spec.MakeInstallTargets
	if len(installTargets) == 0 {
		installTargets = []string{"install"}
	}

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		// Only the CMake path supports PrivatePrefixInstall; here install_prefix is always the
		// shared prefix so the pkg-config patch and verification helpers resolve correctly.
		`install_prefix="${profile_prefix}"`,
		"echo " + LShellTextQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`for required_tool in make gcc; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
	}
	// Write recipe-generated source files, then apply source patches, before configure. No-op
	// when none.
	scriptLines = append(scriptLines, LSourceScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptPatchCreate(spec)...)
	// Bootstrap autotools build files for projects that ship no generated ./configure.
	// Prefer autoreconf -fiv: it regenerates ./configure from configure.ac and does
	// nothing else, which is exactly what is wanted here (configure and make are run
	// as separate steps below). A project autogen.sh is only a fallback, because some
	// (e.g. libklvanc's vid.obe fork) require an action argument like --build and,
	// run bare, print a usage line and exit 1 ??and that --build path would also run
	// configure/make itself, conflicting with the separate steps. autoreconf comes
	// from the base-devel autoconf/automake/libtool already installed.
	if spec.RunAutogen {
		scriptLines = append(scriptLines,
			`echo "Bootstrapping autotools build files"`,
			`if autoreconf -fiv 2>&1; then :; elif [ -f ./autogen.sh ]; then NOCONFIGURE=1 sh ./autogen.sh 2>&1; else echo "ERROR: no way to bootstrap autotools (no autoreconf, no autogen.sh)" >&2; exit 1; fi`,
		)
	}
	if spec.ConfigureSubdir != "" {
		scriptLines = append(scriptLines, `cd "${src_dir}/`+spec.ConfigureSubdir+`"`, `echo "Configure directory: $(pwd)"`)
	}
	scriptLines = append(scriptLines,
		`echo "Configuring with prefix ${profile_prefix}"`,
		`./configure --prefix="${profile_prefix}" `+strings.Join(configureOptionArray, " ")+` 2>&1`,
	)
	// Command-line variable assignments (e.g. SUBDIRS=src) override the makefile's own,
	// which is how a recursive-automake recipe builds and installs just its library
	// subdirectory and skips sibling tools/ that need deps absent on the Windows toolchain.
	makeVariableSuffix := ""
	for _, makeVariable := range spec.MakeVariables {
		makeVariableSuffix += " " + LShellTextQuote(makeVariable)
	}
	buildLine := `make -j"$(nproc)"` + makeVariableSuffix
	if len(buildTargetArray) > 0 {
		buildLine += " " + strings.Join(buildTargetArray, " ")
	}
	scriptLines = append(scriptLines, buildLine+" 2>&1")
	for _, installTarget := range installTargets {
		scriptLines = append(scriptLines, `make `+LShellTextQuote(installTarget)+makeVariableSuffix+` 2>&1`)
	}
	scriptLines = append(scriptLines, LScriptVerifyCreate(spec)...)
	scriptLines = append(scriptLines, "echo "+LShellTextQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

// LMakeScriptCreate builds a library that ships only a plain Makefile (no
// configure, no cmake) inside the MSYS2 environment. Such projects (e.g. quirc) usually
// have no lib-only install target ??their `make install` also builds demos that pull
// extra deps ??so the generator builds just the requested targets (typically the static
// archive) and installs by copying the recipe's declared header files and static archive
// into the selected profile prefix, then verifies them. Headers are copied flat into
// include/ by basename, so a recipe's VerifyHeaderRelativePath must match that basename.
func LMakeScriptCreate(spec LLibraryBuildSpec) ([]string, error) {
	if err := LLibrarySpecValidate(spec, false); err != nil {
		return nil, err
	}
	if len(spec.MakeInstallHeaderFiles) == 0 {
		return nil, fmt.Errorf("make build system requires at least one install header file for %s", spec.LibraryId)
	}
	if spec.MakeStaticLibFile == "" {
		return nil, fmt.Errorf("make build system requires a static lib file for %s", spec.LibraryId)
	}
	buildTargetArray := make([]string, 0, len(spec.MakeBuildTargets))
	for _, buildTarget := range spec.MakeBuildTargets {
		buildTargetArray = append(buildTargetArray, LShellTextQuote(buildTarget))
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		// Only the CMake path supports PrivatePrefixInstall; here install_prefix is always the
		// shared prefix so the pkg-config patch and verification helpers resolve correctly.
		`install_prefix="${profile_prefix}"`,
		"echo " + LShellTextQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`for required_tool in make gcc; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
	}
	// Write recipe-generated source files, then apply source patches, before the build. No-op
	// when none.
	scriptLines = append(scriptLines, LSourceScriptCreate(spec)...)
	scriptLines = append(scriptLines, LScriptPatchCreate(spec)...)
	buildLine := `make -j"$(nproc)"`
	// Command-line variable assignments (e.g. SDL_CFLAGS=) come before the targets so they
	// override and skip the makefile's own $(shell pkg-config ...) assignments.
	for _, makeVariable := range spec.MakeVariables {
		buildLine += " " + LShellTextQuote(makeVariable)
	}
	if len(buildTargetArray) > 0 {
		buildLine += " " + strings.Join(buildTargetArray, " ")
	}
	scriptLines = append(scriptLines, buildLine+" 2>&1")
	scriptLines = append(scriptLines,
		`mkdir -p "${profile_prefix}/include" "${profile_prefix}/lib"`,
	)
	for _, headerFile := range spec.MakeInstallHeaderFiles {
		scriptLines = append(scriptLines,
			`built_header="${src_dir}/`+headerFile+`"`,
			`if [ ! -f "${built_header}" ]; then echo "ERROR: expected source header was not built: ${built_header}"; exit 1; fi`,
			`cp "${built_header}" "${profile_prefix}/include/"`,
			`echo "Installed header: $(basename "${built_header}")"`,
		)
	}
	scriptLines = append(scriptLines,
		`built_lib="${src_dir}/`+spec.MakeStaticLibFile+`"`,
		`if [ ! -f "${built_lib}" ]; then echo "ERROR: expected static library was not built: ${built_lib}"; exit 1; fi`,
		`cp "${built_lib}" "${profile_prefix}/lib/"`,
		`echo "Installed static library: $(basename "${built_lib}")"`,
	)
	scriptLines = append(scriptLines, LScriptVerifyCreate(spec)...)
	scriptLines = append(scriptLines, "echo "+LShellTextQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}
