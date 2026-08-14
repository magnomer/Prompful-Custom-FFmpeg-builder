package scripting

import (
	"strings"
)

// LAppendScriptCreate patches the installed pkg-config module so its Libs line ends
// with the recipe's extra link libraries. Used to repair a static .pc that lists the
// C++/math runtime before its own static archives, which breaks GNU static link order.
// The pkg-config name and lib names are validated as safe path segments by
// LLibrarySpecValidate, so they are safe to LTextInterpolate into the script. Returns no
// lines when the recipe declares no fixup.
// LScriptPatchCreate emits shell that applies each recipe source patch to the
// extracted tree. Each patch replaces an exact full line (Find) with Replace in File
// (relative to ${src_dir}). The step fails loudly when the target line is absent, so a
// re-pinned upstream release that changed the file is caught here instead of silently
// building unpatched. Find/Replace are shell-quoted before being passed to grep/awk.
// Returns no lines when the recipe declares no patches.
func LScriptPatchCreate(spec LLibraryBuildSpec) []string {
	lines := []string{}
	for _, patch := range spec.SourcePatches {
		lines = append(lines,
			`patch_file="${src_dir}/`+patch.File+`"`,
			`if ! grep -qxF -- `+LShellTextQuote(patch.Find)+` "${patch_file}"; then echo "ERROR: source patch target line not found in `+patch.File+`; the pinned upstream release may have changed."; exit 1; fi`,
			`awk -v patch_find=`+LShellTextQuote(patch.Find)+` -v patch_repl=`+LShellTextQuote(patch.Replace)+` '{ if ($0 == patch_find) { print patch_repl } else { print } }' "${patch_file}" > "${patch_file}.patched" && mv "${patch_file}.patched" "${patch_file}"`,
			`echo "Applied source patch to `+patch.File+`"`,
		)
	}
	return lines
}

// LSourceScriptCreate emits shell that writes each recipe-declared source file
// into the extracted tree before configure, for a file a release tarball omits because
// upstream generates it from a .git checkout (e.g. libvmaf's vcs_version.h). Each file's
// parent directory is created first; lines are written verbatim via single-quoted printf
// arguments (validated to contain no single quote or newline). Returns no lines when the
// recipe declares none.
func LSourceScriptCreate(spec LLibraryBuildSpec) []string {
	lines := []string{}
	for _, generated := range spec.GeneratedSourceFiles {
		target := `"${src_dir}/` + generated.Path + `"`
		quoted := make([]string, 0, len(generated.Lines))
		for _, line := range generated.Lines {
			quoted = append(quoted, `'`+line+`'`)
		}
		lines = append(lines,
			`mkdir -p "$(dirname `+target+`)"`,
			`printf '%s\n' `+strings.Join(quoted, " ")+` > `+target,
			`echo "Wrote generated source file: `+generated.Path+`"`,
		)
	}
	return lines
}

// LOverrideScriptCreate replaces the installed .pc's whole "Libs:" line value with
// the recipe's PkgConfigLibsLine. A static archive and a shared import library can share a
// base name (e.g. LibreSSL's static libcrypto.a next to a leftover OpenSSL libcrypto.dll.a in
// the same prefix); GNU ld prefers the .dll.a for a bare -lcrypto, so the link picks the wrong
// crypto and fails with undefined references. Forcing -l:lib<name>.a (or an absolute archive
// path) in the Libs line makes the link select this recipe's own static archives. The pc name
// is a validated path segment and the libs line is validated against LPackageLinePattern,
// so both are safe to LTextInterpolate. Returns no lines when the recipe declares no override.
// LScriptPrefixCreate emits the shell that defines ${install_prefix}. For a normal
// recipe it is the shared MSYS2 profile prefix. For a PrivatePrefixInstall recipe it is a
// per-library directory under the profile prefix, wiped first so a re-run starts clean and
// the library's archives never share a directory with a same-named package's.
func LScriptPrefixCreate(spec LLibraryBuildSpec) string {
	if spec.PrivatePrefixInstall {
		privatePrefix := `${profile_prefix}/` + LPrivateInstallSubdirectory + `/` + spec.PkgConfigName
		return `install_prefix="` + privatePrefix + `"; rm -rf "${install_prefix}"; mkdir -p "${install_prefix}"`
	}
	return `install_prefix="${profile_prefix}"`
}

// LStripScriptCreate removes the Requires and Requires.private lines from a
// privately-installed library's .pc. The private install lists every needed archive by
// absolute path in its Libs line, so leaving Requires.private would only let pkg-config
// re-add bare -l<name> flags that resolve back to a same-named archive in the shared
// prefix ??exactly the collision the private install exists to avoid. No-op otherwise.
func LStripScriptCreate(spec LLibraryBuildSpec) []string {
	if !spec.PrivatePrefixInstall || spec.PkgConfigName == "" {
		return nil
	}
	return []string{
		"echo " + LShellTextQuote("Stripping Requires/Requires.private from "+spec.PkgConfigName+".pc so its absolute-path Libs line is the sole source of link libraries."),
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`sed -i -E '/^Requires(\.private)?:/d' "${pc_file}"`,
	}
}

func LOverrideScriptCreate(spec LLibraryBuildSpec) []string {
	patches := []LLibraryPatchEntry{}
	if spec.PkgConfigName != "" && spec.PkgConfigLibsLine != "" {
		patches = append(patches, LLibraryPatchEntry{Module: spec.PkgConfigName, LibsLine: spec.PkgConfigLibsLine})
	}
	patches = append(patches, spec.PkgConfigLibsLinePatches...)
	if len(patches) == 0 {
		return nil
	}
	lines := []string{}
	for _, patch := range patches {
		// '#' delimiter so the path slashes in the replacement need no escaping; the validated
		// libs line cannot contain '#'. Single-quoted via LShellTextQuote, so ${libdir} stays literal
		// in the .pc for pkg-config to expand.
		sedProgram := "s#^Libs:.*#Libs: " + patch.LibsLine + "#"
		lines = append(lines,
			"echo "+LShellTextQuote("Overriding "+patch.Module+".pc Libs line to force this library's own static archives ahead of any same-named shared import library in the prefix."),
			`pc_file="${install_prefix}/lib/pkgconfig/`+patch.Module+`.pc"`,
			`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/`+patch.Module+`.pc"; fi`,
			`if [ ! -f "${pc_file}" ]; then echo "ERROR: `+patch.Module+`.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
			`sed -i -E `+LShellTextQuote(sedProgram)+` "${pc_file}"`,
			`echo "Patched pkg-config Libs: $(grep '^Libs:' "${pc_file}")"`,
		)
	}
	return lines
}

func LAppendScriptCreate(spec LLibraryBuildSpec) []string {
	if spec.PkgConfigName == "" || len(spec.PkgConfigAppendLibs) == 0 {
		return nil
	}
	stripCommands := ""
	appendFlags := ""
	for _, libName := range spec.PkgConfigAppendLibs {
		// Escape the regex-special '+' (validated names allow only [A-Za-z0-9_+-]).
		escaped := strings.ReplaceAll(libName, "+", `\+`)
		// Remove any existing occurrence first so the appended copy is the only one;
		// otherwise a consumer that de-duplicates link libraries (FFmpeg does) could keep
		// an earlier copy and restore the broken order.
		stripCommands += "s/ -l" + escaped + "( |$)/\\1/g; "
		appendFlags += " -l" + libName
	}
	sedProgram := "/^Libs:/{ " + stripCommands + "s/$/" + appendFlags + "/ }"
	return []string{
		"echo " + LShellTextQuote("Patching "+spec.PkgConfigName+".pc: moving runtime libraries ("+strings.TrimSpace(appendFlags)+") after the static archives for correct GNU static link order."),
		// CMake projects install the .pc under either lib/pkgconfig (LIBDIR) or
		// share/pkgconfig (DATAROOTDIR, e.g. mpeghdec); pkg-config searches both, so accept
		// whichever the project used instead of assuming lib/pkgconfig.
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`sed -i -E ` + LShellTextQuote(sedProgram) + ` "${pc_file}"`,
		`echo "Patched pkg-config Libs: $(grep '^Libs:' "${pc_file}")"`,
	}
}

func LCompilerFlagCreate(spec LLibraryBuildSpec) []string {
	if spec.PkgConfigName == "" || len(spec.PkgConfigAppendCFlags) == 0 {
		return nil
	}
	appendFlags := strings.Join(spec.PkgConfigAppendCFlags, " ")
	lines := []string{
		"echo " + LShellTextQuote("Patching "+spec.PkgConfigName+".pc Cflags: appending "+appendFlags+"."),
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`if ! grep -q '^Cflags:' "${pc_file}"; then printf '%s\n' 'Cflags:' >> "${pc_file}"; fi`,
	}
	for _, appendCFlag := range spec.PkgConfigAppendCFlags {
		sedProgram := `/^Cflags:/ { /(^|[[:space:]])` + appendCFlag + `([[:space:]]|$)/! s|$| ` + appendCFlag + `|; }`
		lines = append(lines, `sed -i -E `+LShellTextQuote(sedProgram)+` "${pc_file}"`)
	}
	lines = append(lines, `echo "Patched pkg-config Cflags: $(grep '^Cflags:' "${pc_file}")"`)
	return lines
}

// LScriptVerifyCreate confirms the installed header (always) and link
// library (when a stem is given) actually landed in the prefix, so a silent no-op
// install or import is turned into a hard failure instead of a later configure error.
func LScriptVerifyCreate(spec LLibraryBuildSpec) []string {
	lines := []string{
		`installed_header="${install_prefix}/include/` + spec.VerifyHeaderRelativePath + `"`,
		`if [ ! -f "${installed_header}" ]; then echo "ERROR: expected header was not installed: ${installed_header}"; exit 1; fi`,
		`echo "Verified installed header: ${installed_header}"`,
	}
	if spec.VerifyLibStem != "" {
		stem := spec.VerifyLibStem
		lines = append(lines,
			`if ls "${install_prefix}/lib/lib`+stem+`".* >/dev/null 2>&1 || ls "${install_prefix}/lib/`+stem+`".* >/dev/null 2>&1; then`,
			`  echo "Verified installed link library for `+stem+`."`,
			`else`,
			`  echo "ERROR: expected link library for `+stem+` was not found in ${install_prefix}/lib."; exit 1;`,
			`fi`,
		)
	}
	return lines
}
