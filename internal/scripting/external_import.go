package scripting

// LScriptExternalCreate imports a verified, already-extracted vendor
// binary archive (the script working directory) into the selected MSYS2 profile
// prefix: headers and link libraries into include/ and lib/, runtime DLLs into bin/,
// generating a MinGW import library from a bundled DLL when needed so that the
// FFmpeg link step (-l<stem>) resolves. It then verifies the imported header.
func LScriptExternalCreate(spec LLibraryBuildSpec) ([]string, error) {
	if err := LLibrarySpecValidate(spec, true); err != nil {
		return nil, err
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		"echo " + LShellTextQuote("Importing external-track library: "+spec.DisplayName),
		`import_root="$(pwd)"`,
		`include_src="${import_root}/` + spec.ImportIncludeSubdir + `"`,
		`lib_src="${import_root}/` + spec.ImportLibSubdir + `"`,
		`if [ ! -d "${include_src}" ]; then echo "ERROR: vendor include directory not found: ${include_src}"; exit 1; fi`,
		`if [ ! -d "${lib_src}" ]; then echo "ERROR: vendor lib directory not found: ${lib_src}"; exit 1; fi`,
		`mkdir -p "${profile_prefix}/include" "${profile_prefix}/lib" "${profile_prefix}/bin"`,
		`echo "Copying vendor headers into the profile prefix."`,
		`cp -r "${include_src}/." "${profile_prefix}/include/"`,
		`echo "Copying vendor import/static libraries into the profile prefix."`,
		`find "${lib_src}" -maxdepth 1 -type f \( -name '*.dll.a' -o -name '*.a' -o -name '*.lib' \) -exec cp {} "${profile_prefix}/lib/" \;`,
		`echo "Copying vendor runtime DLLs into the profile prefix bin directory."`,
		`find "${lib_src}" -maxdepth 1 -type f -name '*.dll' -exec cp {} "${profile_prefix}/bin/" \;`,
		`find "${import_root}" -maxdepth 1 -type f -name '*.dll' -exec cp {} "${profile_prefix}/bin/" \;`,
	}
	if spec.VerifyLibStem != "" {
		stem := spec.VerifyLibStem
		// Produce a GNU import library lib<stem>.dll.a (the name -l<stem> prefers) so the
		// FFmpeg link step resolves. Two independent strategies, neither relying on
		// gendef's exit code (gendef returns non-zero on benign export-decoration
		// warnings, which previously aborted the whole attempt):
		//   1. gendef -> dlltool from the bundled DLL (native GNU import lib).
		//   2. fall back to the vendor MSVC import library (<stem>.lib), which is itself a
		//      COFF import archive the GNU linker can consume, copied to the .dll.a name.
		scriptLines = append(scriptLines,
			`vendor_dll="${profile_prefix}/bin/`+stem+`.dll"`,
			`vendor_lib="${profile_prefix}/lib/`+stem+`.lib"`,
			`mingw_import_lib="${profile_prefix}/lib/lib`+stem+`.dll.a"`,
			`if [ ! -f "${mingw_import_lib}" ] && [ -f "${vendor_dll}" ] && command -v gendef >/dev/null 2>&1 && command -v dlltool >/dev/null 2>&1; then`,
			`  echo "Generating a MinGW import library for `+stem+` from the vendor DLL."`,
			`  def_dir="$(mktemp -d)"`,
			`  def_file="${def_dir}/`+stem+`.def"`,
			`  ( cd "${def_dir}" && gendef "${vendor_dll}" ) >/dev/null 2>&1 || true`,
			`  if [ -s "${def_file}" ]; then`,
			`    dlltool -d "${def_file}" -D "`+stem+`.dll" -l "${mingw_import_lib}" 2>&1 || true`,
			`  else`,
			`    echo "gendef did not produce a usable .def for `+stem+`; will try the vendor import library instead."`,
			`  fi`,
			`  rm -rf "${def_dir}"`,
			`fi`,
			`if [ ! -f "${mingw_import_lib}" ] && [ -f "${vendor_lib}" ]; then`,
			`  echo "Using the vendor `+stem+`.lib as the import library (copied to lib`+stem+`.dll.a)."`,
			`  cp "${vendor_lib}" "${mingw_import_lib}"`,
			`fi`,
			`if [ -f "${mingw_import_lib}" ]; then echo "Import library ready: ${mingw_import_lib}"; else echo "WARNING: no import library for `+stem+` could be prepared; the FFmpeg link step may fail."; fi`,
		)
	}
	scriptLines = append(scriptLines, LScriptVerifyCreate(spec)...)
	scriptLines = append(scriptLines, "echo "+LShellTextQuote("External-track library imported: "+spec.DisplayName))
	return scriptLines, nil
}
