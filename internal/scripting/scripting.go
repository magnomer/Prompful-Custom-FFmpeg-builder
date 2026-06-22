package scripting

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"customffmpegbuilder/internal/workspace"
)

type ScriptFilePlan struct {
	WorkspaceDirectory string   `json:"workspaceDirectory"`
	ScriptFilePath     string   `json:"scriptFilePath"`
	ScriptLines        []string `json:"scriptLines"`
}

type ScriptFile struct {
	ScriptFilePath   string `json:"scriptFilePath"`
	ScriptSha256Hash string `json:"scriptSha256Hash"`
}

var safeMsys2PackageNamePattern = regexp.MustCompile(`^[A-Za-z0-9_+.-]+$`)
var safeConfigureFlagPattern = regexp.MustCompile(`^--[A-Za-z0-9][A-Za-z0-9_+./:=,-]*$`)

func ValidateMsys2PackageName(packageName string) error {
	if packageName == "" {
		return errors.New("MSYS2 package name is empty")
	}
	if !safeMsys2PackageNamePattern.MatchString(packageName) {
		return fmt.Errorf("MSYS2 package name contains unsafe characters: %s", packageName)
	}
	return nil
}

func ValidateConfigureFlag(configureFlag string) error {
	if configureFlag == "" {
		return errors.New("configure flag is empty")
	}
	if !safeConfigureFlagPattern.MatchString(configureFlag) {
		return fmt.Errorf("configure flag contains unsafe characters: %s", configureFlag)
	}
	if strings.Contains(configureFlag, "..") {
		return fmt.Errorf("configure flag contains unsafe path traversal marker: %s", configureFlag)
	}
	return nil
}

func WriteScriptFile(scriptFilePlan ScriptFilePlan) (ScriptFile, error) {
	if scriptFilePlan.WorkspaceDirectory == "" || scriptFilePlan.ScriptFilePath == "" {
		return ScriptFile{}, errors.New("approved shell script paths must not be empty")
	}
	if err := workspace.CheckPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, filepath.Dir(scriptFilePlan.ScriptFilePath)); err != nil {
		return ScriptFile{}, err
	}
	if len(scriptFilePlan.ScriptLines) == 0 {
		return ScriptFile{}, errors.New("approved shell script has no lines")
	}
	scriptDirectory := filepath.Dir(scriptFilePlan.ScriptFilePath)
	if err := os.MkdirAll(scriptDirectory, 0o755); err != nil {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptFilePlan.ScriptFilePath); err != nil {
		return ScriptFile{}, err
	}
	scriptText := strings.Join(scriptFilePlan.ScriptLines, "\n") + "\n"
	scriptHash := sha256.Sum256([]byte(scriptText))
	scriptHashString := hex.EncodeToString(scriptHash[:])
	if existingInfo, err := os.Lstat(scriptFilePlan.ScriptFilePath); err == nil {
		if existingInfo.Mode()&os.ModeSymlink != 0 {
			return ScriptFile{}, errors.New("approved shell script path is a symlink")
		}
		if existingInfo.IsDir() {
			return ScriptFile{}, errors.New("approved shell script path is a directory")
		}
		existingBytes, readErr := os.ReadFile(scriptFilePlan.ScriptFilePath)
		if readErr != nil {
			return ScriptFile{}, readErr
		}
		existingHash := sha256.Sum256(existingBytes)
		if strings.EqualFold(hex.EncodeToString(existingHash[:]), scriptHashString) {
			return ScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
		}
		if removeErr := os.Remove(scriptFilePlan.ScriptFilePath); removeErr != nil {
			return ScriptFile{}, fmt.Errorf("remove stale approved shell script: %w", removeErr)
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return ScriptFile{}, err
	}
	if err := workspace.CheckRealPathInsideWorkspace(scriptFilePlan.WorkspaceDirectory, scriptDirectory); err != nil {
		return ScriptFile{}, err
	}
	outputFile, err := os.OpenFile(scriptFilePlan.ScriptFilePath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o700)
	if err != nil {
		return ScriptFile{}, err
	}
	_, writeErr := outputFile.Write([]byte(scriptText))
	closeErr := outputFile.Close()
	if writeErr != nil {
		return ScriptFile{}, writeErr
	}
	if closeErr != nil {
		return ScriptFile{}, closeErr
	}
	return ScriptFile{ScriptFilePath: scriptFilePlan.ScriptFilePath, ScriptSha256Hash: scriptHashString}, nil
}

func PacmanInstallScriptLines(packageNames []string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := ValidateMsys2PackageName(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, shellQuote(packageName))
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package server for this private toolchain.'",
		"cat > /etc/pacman.d/mirrorlist.msys <<'EOF'",
		"Server = https://repo.msys2.org/msys/$arch/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw32 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw32/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.ucrt64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/ucrt64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clang64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clang64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clangarm64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clangarm64/",
		"EOF",
		"echo 'Initializing the private MSYS2 package keyring.'",
		"pacman-key --init 2>&1",
		"pacman-key --populate msys2 2>&1",
		"echo 'Preparing pacman database signature policy.'",
		"# MSYS2 packages remain signature-checked. Repository database signatures are treated as optional, matching normal MSYS2 pacman behavior and avoiding false failures when a database .sig is not served or is cleared during refresh.",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
		"echo 'Clearing stale or half-downloaded package databases before refresh.'",
		"rm -f /var/lib/pacman/sync/*.db /var/lib/pacman/sync/*.db.sig /var/lib/pacman/sync/*.files /var/lib/pacman/sync/*.files.sig /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing the private MSYS2 package databases and keyring package.'",
		"pacman -Syy --needed --noconfirm msys2-keyring",
		"echo 'Installing the approved build-tool packages.'",
		"pacman -S --needed --noconfirm " + strings.Join(quotedPackageNames, " "),
		"echo 'Checking that the selected MSYS2 compiler can create a Windows executable.'",
		"echo 'The compiler check uses a private writable folder inside the extracted MSYS2 toolchain.'",
		`compiler_check_dir="/usr/local/var/customffmpeg-compiler-check"`,
		`rm -rf "$compiler_check_dir"`,
		`mkdir -p "$compiler_check_dir"`,
		`trap 'rm -rf "$compiler_check_dir"' EXIT`,
		`printf 'int main(void){return 0;}\n' > "$compiler_check_dir/check.c"`,
		`gcc "$compiler_check_dir/check.c" -o "$compiler_check_dir/check.exe"`,
		`test -s "$compiler_check_dir/check.exe"`,
		`chmod +x "$compiler_check_dir/check.exe"`,
		`"$compiler_check_dir/check.exe"`,
		`rm -rf "$compiler_check_dir"`,
	}, nil
}

func FfmpegLibraryPackageInstallScriptLines(packageNames []string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := ValidateMsys2PackageName(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, shellQuote(packageName))
	}
	if len(quotedPackageNames) == 0 {
		return []string{
			"#!/usr/bin/env bash",
			"set -euo pipefail",
			"echo 'No FFmpeg library packages were selected.'",
		}, nil
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package server for FFmpeg library packages.'",
		"cat > /etc/pacman.d/mirrorlist.msys <<'EOF'",
		"Server = https://repo.msys2.org/msys/$arch/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw32 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw32/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.mingw64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/mingw64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.ucrt64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/ucrt64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clang64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clang64/",
		"EOF",
		"cat > /etc/pacman.d/mirrorlist.clangarm64 <<'EOF'",
		"Server = https://repo.msys2.org/mingw/clangarm64/",
		"EOF",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
		"echo 'Clearing half-downloaded package files before installing FFmpeg library packages.'",
		"rm -f /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing package databases before installing FFmpeg library packages.'",
		"pacman -Syy --needed --noconfirm",
		"echo 'Installing MSYS2 packages required by the selected FFmpeg libraries.'",
		"pacman -S --noconfirm " + strings.Join(quotedPackageNames, " "),
	}, nil
}

func ConfigureScriptLines(configureFlags []string) ([]string, error) {
	quotedConfigureFlags := make([]string, 0, len(configureFlags))
	for _, configureFlag := range configureFlags {
		if err := ValidateConfigureFlag(configureFlag); err != nil {
			return nil, err
		}
		quotedConfigureFlags = append(quotedConfigureFlags, shellQuote(configureFlag))
	}
	quotedConfigureFlagArray := strings.Join(quotedConfigureFlags, " ")

	pkgConfigModules := pkgConfigModulesForConfigureFlags(configureFlags)

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		`export PATH="${profile_prefix}/bin:/usr/bin:${PATH}"`,
		`export PKG_CONFIG_PATH="${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export PKG_CONFIG_LIBDIR="${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export CPPFLAGS="-I${profile_prefix}/include"`,
		`if [ -x "${profile_prefix}/bin/pkgconf.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkgconf.exe"; elif [ -x "${profile_prefix}/bin/pkg-config.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkg-config.exe"; else export PKG_CONFIG="pkg-config"; fi`,
		`echo "Using pkg-config command: ${PKG_CONFIG}"`,
		`echo "Using pkg-config search paths: ${PKG_CONFIG_LIBDIR}"`,
		"configure_flags=(" + quotedConfigureFlagArray + ")",
		`svtjpegxs_disabled=0`,
		`lensfun_disabled=0`,
		`remove_configure_flag() {`,
		`  flag_to_remove="$1"`,
		`  kept_flags=()`,
		`  for configure_flag in "${configure_flags[@]}"; do`,
		`    if [ "${configure_flag}" != "${flag_to_remove}" ]; then`,
		`      kept_flags+=("${configure_flag}")`,
		`    fi`,
		`  done`,
		`  configure_flags=("${kept_flags[@]}")`,
		`}`,
		`diagnose_pkg_config_module() {`,
		`  module_name="$1"`,
		`  minimum_version="${2:-}"`,
		`  echo "pkg-config diagnostic starting: ${module_name} at $(date +%T)"`,
		`  echo "pkg-config module: ${module_name}"`,
		`  if ! "${PKG_CONFIG}" --print-errors --exists "${module_name}" 2>&1; then`,
		`    echo "pkg-config diagnostic result: ${module_name} was not found in ${PKG_CONFIG_LIBDIR}"`,
		`    echo "pkg-config diagnostic completed: ${module_name} at $(date +%T)"`,
		`    return 0`,
		`  fi`,
		`  echo "pkg-config exists: yes"`,
		`  echo "pkg-config version: $("${PKG_CONFIG}" --modversion "${module_name}" 2>&1 || true)"`,
		`  echo "pkg-config pcfiledir: $("${PKG_CONFIG}" --variable=pcfiledir "${module_name}" 2>&1 || true)"`,
		`  echo "pkg-config cflags: $("${PKG_CONFIG}" --cflags "${module_name}" 2>&1 || true)"`,
		`  echo "pkg-config libs: $("${PKG_CONFIG}" --libs "${module_name}" 2>&1 || true)"`,
		`  echo "pkg-config libs static: $("${PKG_CONFIG}" --libs --static "${module_name}" 2>&1 || true)"`,
		`  if [ -n "${minimum_version}" ]; then`,
		`    if "${PKG_CONFIG}" --print-errors --atleast-version="${minimum_version}" "${module_name}" 2>&1; then`,
		`      echo "pkg-config version check: ${module_name} >= ${minimum_version}"`,
		`    else`,
		`      echo "pkg-config version check failed: ${module_name} must be >= ${minimum_version}"`,
		`      exit 1`,
		`    fi`,
		`  fi`,
		`  probe_dir="$(mktemp -d)"`,
		`  printf 'int main(void){return 0;}
' > "${probe_dir}/pkg_probe.c"`,
		`  echo "pkg-config compile/link probe starting: ${module_name}"`,
		`  if gcc $("${PKG_CONFIG}" --cflags "${module_name}") "${probe_dir}/pkg_probe.c" $("${PKG_CONFIG}" --libs "${module_name}") -o "${probe_dir}/pkg_probe.exe" 2>&1; then`,
		`    echo "pkg-config compile/link probe passed: ${module_name}"`,
		`  else`,
		`    echo "pkg-config compile/link probe failed: ${module_name}"`,
		`  fi`,
		`  rm -rf "${probe_dir}"`,
		`  echo "pkg-config diagnostic completed: ${module_name} at $(date +%T)"`,
		`}`,
		`try_enable_svt_jpeg_xs() {`,
		`  echo "SVT JPEG XS future-compatibility check starting at $(date +%T)"`,
		`  echo "SVT JPEG XS is hidden from the UI for now. Backend support is left for future compatibility."`,
		`  echo "Trying MSYS2/package-provided SvtJpegxs first."`,
		`  if "${PKG_CONFIG}" --exists "SvtJpegxs >= 0.10.0" 2>/dev/null; then`,
		`    echo "SVT JPEG XS package is compatible: $("${PKG_CONFIG}" --modversion SvtJpegxs 2>&1 || true)"`,
		`    return 0`,
		`  fi`,
		`  if "${PKG_CONFIG}" --exists SvtJpegxs 2>/dev/null; then`,
		`    echo "SVT JPEG XS package is installed but incompatible: $("${PKG_CONFIG}" --modversion SvtJpegxs 2>&1 || true)"`,
		`    "${PKG_CONFIG}" --print-errors --exists "SvtJpegxs >= 0.10.0" 2>&1 || true`,
		`  else`,
		`    echo "SVT JPEG XS package is not visible to pkg-config."`,
		`  fi`,
		`  echo "Trying official upstream SVT-JPEG-XS source as a fallback."`,
		`  if ! command -v git >/dev/null 2>&1 || ! command -v cmake >/dev/null 2>&1 || ! command -v ninja >/dev/null 2>&1; then`,
		`    echo "WARNING: SVT JPEG XS fallback needs git, cmake, and ninja. Disabling --enable-libsvtjpegxs and continuing."`,
		`    return 1`,
		`  fi`,
		`  svt_work_root="$(pwd)/../customffmpeg-svt-jpeg-xs-upstream"`,
		`  svt_source_dir="${svt_work_root}/source"`,
		`  svt_build_dir="${svt_work_root}/build"`,
		`  rm -rf "${svt_work_root}"`,
		`  mkdir -p "${svt_work_root}"`,
		`  if ! git clone --depth 1 --recursive https://github.com/OpenVisualCloud/SVT-JPEG-XS.git "${svt_source_dir}" 2>&1; then`,
		`    echo "WARNING: SVT JPEG XS upstream source download failed. Disabling --enable-libsvtjpegxs and continuing."`,
		`    return 1`,
		`  fi`,
		`  echo "SVT JPEG XS upstream branch: $(git -C "${svt_source_dir}" branch --show-current 2>&1 || true)"`,
		`  echo "SVT JPEG XS upstream commit: $(git -C "${svt_source_dir}" rev-parse HEAD 2>&1 || true)"`,
		`  echo "SVT JPEG XS upstream describe: $(git -C "${svt_source_dir}" describe --tags --always --dirty 2>&1 || true)"`,
		`  svt_declared_version="$(sed -n 's/.*project(svt-jpegxs VERSION \([0-9.]*\).*/\1/p' "${svt_source_dir}/CMakeLists.txt" | head -n 1)"`,
		`  echo "SVT JPEG XS upstream CMake project version before build: ${svt_declared_version:-unknown}"`,
		`  echo "SVT JPEG XS API check in upstream source:"`,
		`  grep -R "svt_jpeg_xs_decoder_get_single_frame_size_with_proxy" -n "${svt_source_dir}/Source" 2>&1 || true`,
		`  grep -R "svt_jpeg_xs_encoder_init" -n "${svt_source_dir}/Source" 2>&1 || true`,
		`  if ! MSYS2_ARG_CONV_EXCL="-DCMAKE_INSTALL_PREFIX=" cmake -S "${svt_source_dir}" -B "${svt_build_dir}" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="${profile_prefix}" -DCMAKE_DLL_NAME_WITH_SOVERSION=ON -DCMAKE_POLICY_VERSION_MINIMUM=3.5.0 -DBUILD_SHARED_LIBS=ON -DBUILD_APPS=OFF -DNATIVE=OFF -Wno-dev 2>&1; then`,
		`    echo "WARNING: SVT JPEG XS upstream configure failed. Disabling --enable-libsvtjpegxs and continuing."`,
		`    return 1`,
		`  fi`,
		`  if ! cmake --build "${svt_build_dir}" 2>&1; then`,
		`    echo "WARNING: SVT JPEG XS upstream build failed. Disabling --enable-libsvtjpegxs and continuing."`,
		`    return 1`,
		`  fi`,
		`  if ! cmake --install "${svt_build_dir}" 2>&1; then`,
		`    echo "WARNING: SVT JPEG XS upstream install failed. Disabling --enable-libsvtjpegxs and continuing."`,
		`    return 1`,
		`  fi`,
		`  echo "SVT JPEG XS installed pkg-config version after source-build: $("${PKG_CONFIG}" --modversion SvtJpegxs 2>&1 || true)"`,
		`  if "${PKG_CONFIG}" --exists "SvtJpegxs >= 0.10.0" 2>/dev/null; then`,
		`    echo "SVT JPEG XS upstream source-build satisfied FFmpeg's SvtJpegxs >= 0.10.0 requirement."`,
		`    return 0`,
		`  fi`,
		`  echo "WARNING: SVT JPEG XS remained incompatible after MSYS2 package and upstream source attempts. Disabling --enable-libsvtjpegxs and continuing."`,
		`  "${PKG_CONFIG}" --print-errors --exists "SvtJpegxs >= 0.10.0" 2>&1 || true`,
		`  return 1`,
		`}`,
		`lensfun_symbol_probe() {`,
		`  symbol_name="$1"`,
		`  probe_dir="$(mktemp -d)"`,
		`  printf '#include <lensfun.h>
#include <stdint.h>
long check_symbol(void) { return (long) %s; }
int main(void){return ((intptr_t)check_symbol) == 0;}
' "${symbol_name}" > "${probe_dir}/lensfun_symbol_probe.c"`,
		`  echo "lensfun symbol/link probe: ${symbol_name}"`,
		`  if gcc $("${PKG_CONFIG}" --cflags lensfun) "${probe_dir}/lensfun_symbol_probe.c" $("${PKG_CONFIG}" --libs lensfun) -o "${probe_dir}/lensfun_symbol_probe.exe" 2>&1; then`,
		`    echo "lensfun symbol/link probe passed: ${symbol_name}"`,
		`    rm -rf "${probe_dir}"`,
		`    return 0`,
		`  fi`,
		`  echo "lensfun symbol/link probe failed: ${symbol_name}"`,
		`  rm -rf "${probe_dir}"`,
		`  return 1`,
		`}`,
		`diagnose_lensfun_details() {`,
		`  echo "lensfun detail diagnostic starting at $(date +%T)"`,
		`  echo "lensfun header search under ${profile_prefix}/include:"`,
		`  find "${profile_prefix}/include" -maxdepth 4 \( -iname '*lensfun*' -o -iname 'lensfun.h' \) -print 2>&1 || true`,
		`  if "${PKG_CONFIG}" --exists lensfun; then`,
		`    lensfun_probe_dir="$(mktemp -d)"`,
		`    printf '#include <lensfun.h>
int main(void){return 0;}
' > "${lensfun_probe_dir}/lensfun_header_probe.c"`,
		`    echo "lensfun header probe: #include <lensfun.h>"`,
		`    gcc $("${PKG_CONFIG}" --cflags lensfun) -c "${lensfun_probe_dir}/lensfun_header_probe.c" -o "${lensfun_probe_dir}/lensfun_header_probe.o" 2>&1 || true`,
		`    printf '#include <lensfun/lensfun.h>
int main(void){return 0;}
' > "${lensfun_probe_dir}/lensfun_nested_header_probe.c"`,
		`    echo "lensfun header probe: #include <lensfun/lensfun.h>"`,
		`    gcc $("${PKG_CONFIG}" --cflags lensfun) -c "${lensfun_probe_dir}/lensfun_nested_header_probe.c" -o "${lensfun_probe_dir}/lensfun_nested_header_probe.o" 2>&1 || true`,
		`    rm -rf "${lensfun_probe_dir}"`,
		`    lensfun_symbol_probe lf_db_create || true`,
		`    lensfun_symbol_probe lf_db_new || true`,
		`    lensfun_symbol_probe lf_db_load_path || true`,
		`  fi`,
		`  echo "lensfun detail diagnostic completed at $(date +%T)"`,
		`}`,
		`lensfun_ffmpeg_api_probe() {`,
		`  probe_dir="$(mktemp -d)"`,
		`  cat > "${probe_dir}/lensfun_ffmpeg_api_probe.c" <<'EOF'`,
		`#include <lensfun.h>`,
		`#include <stdint.h>`,
		`int main(void) {`,
		`    lfDatabase *db = lf_db_create();`,
		`    lfCamera *camera = lf_camera_create();`,
		`    lfLens *lens = lf_lens_create();`,
		`    const lfLens **lenses = lf_db_find_lenses(db, camera, (const char *)0, (const char *)0, 0);`,
		`    lfModifier *modifier = lf_modifier_create(lens, LF_PF_U8, 1920, 1080);`,
		`    int status = lf_db_load_path(db, "");`,
		`    lf_modifier_enable_vignetting_correction(modifier, 1.0f, 1.0f);`,
		`    lf_modifier_enable_distortion_correction(modifier);`,
		`    lf_modifier_enable_projection_transform(modifier, LF_RECTILINEAR);`,
		`    lf_modifier_enable_scaling(modifier, 1.0f);`,
		`    lf_modifier_enable_tca_correction(modifier);`,
		`    return (int)((intptr_t)db + (intptr_t)camera + (intptr_t)lens + (intptr_t)lenses + (intptr_t)modifier + status);`,
		`}`,
		`EOF`,
		`  echo "lensfun FFmpeg API probe starting."`,
		`  if gcc $("${PKG_CONFIG}" --cflags lensfun) -Werror=implicit-function-declaration -Werror=incompatible-pointer-types -Werror=int-conversion "${probe_dir}/lensfun_ffmpeg_api_probe.c" $("${PKG_CONFIG}" --libs lensfun) -o "${probe_dir}/lensfun_ffmpeg_api_probe.exe" 2>&1; then`,
		`    echo "lensfun FFmpeg API probe passed."`,
		`    rm -rf "${probe_dir}"`,
		`    return 0`,
		`  fi`,
		`  echo "lensfun FFmpeg API probe failed."`,
		`  rm -rf "${probe_dir}"`,
		`  return 1`,
		`}`,
		`try_enable_lensfun() {`,
		`  echo "lensfun compatibility check starting at $(date +%T)"`,
		`  echo "lensfun is hidden from automatic presets for now. Backend support is left for future compatibility."`,
		`  if ! "${PKG_CONFIG}" --exists lensfun; then`,
		`    echo "WARNING: lensfun is not visible to pkg-config. Disabling --enable-liblensfun and continuing."`,
		`    return 1`,
		`  fi`,
		`  echo "lensfun pkg-config version: $("${PKG_CONFIG}" --modversion lensfun 2>&1 || true)"`,
		`  if lensfun_ffmpeg_api_probe; then`,
		`    echo "lensfun package exposes the FFmpeg-required API. Keeping --enable-liblensfun."`,
		`    return 0`,
		`  fi`,
		`  echo "WARNING: lensfun package was found, but it does not expose the API used by this FFmpeg source. Disabling --enable-liblensfun and continuing."`,
		`  echo "WARNING: This backend path is kept for future compatibility; current MSYS2 lensfun is not compatible with this FFmpeg lensfun filter source."`,
		`  return 1`,
		`}`,
	}
	if configureFlagSelected(configureFlags, "--enable-libsvtjpegxs") {
		scriptLines = append(scriptLines, "if ! try_enable_svt_jpeg_xs; then svtjpegxs_disabled=1; remove_configure_flag --enable-libsvtjpegxs; fi")
	}
	if configureFlagSelected(configureFlags, "--enable-liblensfun") {
		scriptLines = append(scriptLines, "if ! try_enable_lensfun; then lensfun_disabled=1; remove_configure_flag --enable-liblensfun; fi")
	}
	if len(pkgConfigModules) > 0 {
		scriptLines = append(scriptLines, "echo 'Diagnosing selected pkg-config libraries before FFmpeg configure.'")
		for _, mod := range pkgConfigModules {
			diagnosticLine := fmt.Sprintf("diagnose_pkg_config_module %s %s", shellQuote(mod.Name), shellQuote(mod.MinVersion))
			if mod.Name == "SvtJpegxs" {
				diagnosticLine = "if [ \"${svtjpegxs_disabled}\" -eq 0 ]; then " + diagnosticLine + "; else echo 'SVT JPEG XS pkg-config diagnostic skipped because --enable-libsvtjpegxs was disabled after compatibility checks.'; fi"
			}
			scriptLines = append(scriptLines, diagnosticLine)
			if mod.Name == "lensfun" {
				diagnosticLine = "if [ \"${lensfun_disabled}\" -eq 0 ]; then " + diagnosticLine + "; else echo 'lensfun pkg-config diagnostic skipped because --enable-liblensfun was disabled after compatibility checks.'; fi"
				scriptLines[len(scriptLines)-1] = diagnosticLine
				scriptLines = append(scriptLines, "if [ \"${lensfun_disabled}\" -eq 0 ]; then diagnose_lensfun_details; else echo 'lensfun detail diagnostic skipped because --enable-liblensfun was disabled after compatibility checks.'; fi")
			}
		}
	}
	scriptLines = append(scriptLines, `echo "Starting FFmpeg configure at $(date +%T)"`)
	scriptLines = append(scriptLines, "set +e")
	scriptLines = append(scriptLines, `echo "Final FFmpeg configure flags: ${configure_flags[*]}"`)
	scriptLines = append(scriptLines, `./configure "${configure_flags[@]}"`)
	scriptLines = append(scriptLines, "configure_status=$?")
	scriptLines = append(scriptLines, "set -e")
	scriptLines = append(scriptLines, `if [ "${configure_status}" -ne 0 ]; then`)
	scriptLines = append(scriptLines, `  echo "FFmpeg configure failed with status ${configure_status}."`)
	scriptLines = append(scriptLines, `  if [ -f ffbuild/config.log ]; then`)
	scriptLines = append(scriptLines, `    echo "----- BEGIN ffbuild/config.log tail -----"`)
	scriptLines = append(scriptLines, `    tail -n 220 ffbuild/config.log`)
	scriptLines = append(scriptLines, `    echo "----- END ffbuild/config.log tail -----"`)
	scriptLines = append(scriptLines, `  else`)
	scriptLines = append(scriptLines, `    echo "ffbuild/config.log was not found after configure failure."`)
	scriptLines = append(scriptLines, `  fi`)
	scriptLines = append(scriptLines, `  exit "${configure_status}"`)
	scriptLines = append(scriptLines, `fi`)
	scriptLines = append(scriptLines, `echo "FFmpeg configure completed at $(date +%T)"`)
	return scriptLines, nil
}

// pkgConfigModule holds a pkg-config module name and an optional minimum version
// that must be satisfied before FFmpeg configure is attempted.
type pkgConfigModule struct {
	Name       string
	MinVersion string // empty means no version check beyond existence
}

func pkgConfigModulesForConfigureFlags(configureFlags []string) []pkgConfigModule {
	// Only list libraries that are expected to provide a pkg-config module in MSYS2.
	// Some valid FFmpeg options, such as --enable-libgsm, are probed by FFmpeg
	// through headers/libraries instead of a .pc file. Pre-checking those with
	// pkg-config incorrectly blocks valid builds.
	//
	// MinVersion values mirror the lower bounds enforced by FFmpeg's own configure
	// script. When FFmpeg requires a minimum version, we check it here with
	// --atleast-version so that the error is caught before configure runs and the
	// message clearly names the library and required version rather than the
	// generic "not found" that configure emits when the version constraint fails.
	type entry struct {
		name       string
		minVersion string
	}
	moduleByFlag := map[string]entry{
		"--enable-libaom":            {name: "aom"},
		"--enable-libass":            {name: "libass"},
		"--enable-libbluray":         {name: "libbluray"},
		"--enable-libcdio":           {name: "libcdio"},
		"--enable-libdav1d":          {name: "dav1d", minVersion: "1.0"},
		"--enable-libfdk-aac":        {name: "fdk-aac"},
		"--enable-libfontconfig":     {name: "fontconfig"},
		"--enable-fontconfig":        {name: "fontconfig"},
		"--enable-libfreetype":       {name: "freetype2"},
		"--enable-libfribidi":        {name: "fribidi"},
		"--enable-libharfbuzz":       {name: "harfbuzz"},
		"--enable-libilbc":           {name: "libilbc"},
		"--enable-liblensfun":        {name: "lensfun"},
		"--enable-libjxl":            {name: "libjxl", minVersion: "0.7.0"},
		"--enable-libmodplug":        {name: "libmodplug"},
		"--enable-libmp3lame":        {name: "lame"},
		"--enable-libopencore-amrnb": {name: "opencore-amrnb"},
		"--enable-libopencore-amrwb": {name: "opencore-amrwb"},
		"--enable-libopenh264":       {name: "openh264"},
		"--enable-libopenjpeg":       {name: "libopenjp2"},
		"--enable-libopus":           {name: "opus"},
		"--enable-libplacebo":        {name: "libplacebo", minVersion: "5.229.0"},
		"--enable-librav1e":          {name: "rav1e"},
		"--enable-librubberband":     {name: "rubberband"},
		"--enable-libsoxr":           {name: "soxr"},
		"--enable-libspeex":          {name: "speex"},
		"--enable-libssh":            {name: "libssh"},
		"--enable-libsvtav1":         {name: "SvtAv1Enc", minVersion: "0.9.0"},
		"--enable-libtesseract":      {name: "tesseract"},
		"--enable-libtwolame":        {name: "twolame"},
		"--enable-libvmaf":           {name: "libvmaf", minVersion: "2.0.0"},
		"--enable-libvorbis":         {name: "vorbis"},
		"--enable-libvpx":            {name: "vpx"},
		"--enable-libwebp":           {name: "libwebp"},
		"--enable-libx264":           {name: "x264"},
		"--enable-libx265":           {name: "x265"},
		"--enable-libxavs2":          {name: "xavs2"},
		"--enable-libzimg":           {name: "zimg", minVersion: "2.9"},
		"--enable-openal":            {name: "openal"},
		"--enable-openssl":           {name: "openssl"},
		"--enable-gnutls":            {name: "gnutls"},
		"--enable-sdl2":              {name: "sdl2"},
		"--enable-chromaprint":       {name: "libchromaprint"},
		"--enable-libaribcaption":    {name: "libaribcaption"},
		"--enable-libbs2b":           {name: "libbs2b"},
		"--enable-libcaca":           {name: "caca"},
		"--enable-libdvdread":        {name: "dvdread"},
		"--enable-libmysofa":         {name: "libmysofa"},
		"--enable-libopencolorio":    {name: "OpenColorIO"},
		"--enable-libopencv":         {name: "opencv4"},
		"--enable-libqrencode":       {name: "libqrencode"},
		"--enable-librabbitmq":       {name: "librabbitmq"},
		"--enable-librsvg":           {name: "librsvg-2.0"},
		"--enable-libsvtjpegxs":      {name: "SvtJpegxs", minVersion: "0.10.0"},
		"--enable-liblc3":            {name: "lc3"},
		"--enable-lv2":               {name: "lilv-0"},
		"--enable-lcms2":             {name: "lcms2"},
		"--enable-opencl":            {name: "OpenCL"},
		"--enable-whisper":           {name: "whisper"},
	}
	modules := []pkgConfigModule{}
	seen := map[string]bool{}
	for _, configureFlag := range configureFlags {
		e, exists := moduleByFlag[configureFlag]
		if !exists || seen[e.name] {
			continue
		}
		seen[e.name] = true
		modules = append(modules, pkgConfigModule{Name: e.name, MinVersion: e.minVersion})
	}
	return modules
}

func configureFlagSelected(configureFlags []string, wantedFlag string) bool {
	for _, configureFlag := range configureFlags {
		if configureFlag == wantedFlag {
			return true
		}
	}
	return false
}

func MakeScriptLines(parallelJobCount int) ([]string, error) {
	if parallelJobCount < 1 || parallelJobCount > 256 {
		return nil, errors.New("parallel job count must be between 1 and 256")
	}
	return []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`echo "Starting FFmpeg make at $(date +%T)"`,
		fmt.Sprintf("make -j%d", parallelJobCount),
		`echo "FFmpeg make completed at $(date +%T)"`,
	}, nil
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\\''") + "'"
}
