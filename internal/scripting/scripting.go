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

	"promptfulcustomffmpegbuilder/internal/workspace"
	"promptfulcustomffmpegbuilder/shared/releasesupport"
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

// mingwRepoNames are the MSYS2 mingw-family repositories that the default
// pacman.conf enables. A build targets exactly one of them (the selected shell
// profile); the rest are disabled before syncing so an unrelated repo's transient
// mirror failure cannot affect ??or appear in ??the build.
var mingwRepoNames = []string{"mingw32", "mingw64", "ucrt64", "clang64", "clangarm64"}

// msys2Mirrors are the upstream package servers pacman tries in order. The first two
// are MSYS2's own origin and CDN; the rest are long-lived independent mirrors. Listing
// several lets a per-file 404 fall through to the next server when one mirror has not
// yet propagated a just-published package (mirror/CDN skew) instead of failing the
// build. The list is fixed and ordered so the generated, hash-pinned script is stable.
var msys2Mirrors = []string{
	"https://repo.msys2.org",
	"https://mirror.msys2.org",
	"https://mirrors.tuna.tsinghua.edu.cn/msys2",
	"https://download.nus.edu.sg/mirror/msys2",
}

// msys2MirrorlistRepos pairs each pacman mirrorlist file with the repository path
// appended to every mirror base. Ordered for a deterministic script.
var msys2MirrorlistRepos = []struct {
	mirrorlistName string
	repoPath       string
}{
	{"msys", "/msys/$arch/"},
	{"mingw32", "/mingw/mingw32/"},
	{"mingw64", "/mingw/mingw64/"},
	{"ucrt64", "/mingw/ucrt64/"},
	{"clang64", "/mingw/clang64/"},
	{"clangarm64", "/mingw/clangarm64/"},
}

// mirrorlistSetupLines writes every pacman mirrorlist with the full ordered mirror set
// so pacman retries the next server on a per-file 404. Shared by the toolchain install
// and the FFmpeg library/build-dependency package install so both gain the same fallback.
func mirrorlistSetupLines() []string {
	lines := []string{}
	for _, repo := range msys2MirrorlistRepos {
		lines = append(lines, "cat > /etc/pacman.d/mirrorlist."+repo.mirrorlistName+" <<'EOF'")
		for _, mirror := range msys2Mirrors {
			lines = append(lines, "Server = "+mirror+repo.repoPath)
		}
		lines = append(lines, "EOF")
	}
	return lines
}

func selectedProfileRepoName(windowsShellProfileName string) string {
	for _, repoName := range mingwRepoNames {
		if repoName == windowsShellProfileName {
			return repoName
		}
	}
	return "ucrt64"
}

func PacmanInstallScriptLines(packageNames []string, windowsShellProfileName string) ([]string, error) {
	quotedPackageNames := make([]string, 0, len(packageNames))
	for _, packageName := range packageNames {
		if err := ValidateMsys2PackageName(packageName); err != nil {
			return nil, err
		}
		quotedPackageNames = append(quotedPackageNames, shellQuote(packageName))
	}

	// Disable every mingw-family repo except the selected profile's, so `pacman -Syy`
	// only refreshes the databases this build actually uses (the profile repo plus
	// the always-needed `msys` repo). This keeps a build that chose mingw64 from
	// touching clang64/ucrt64/etc. databases.
	selectedRepoName := selectedProfileRepoName(windowsShellProfileName)
	disableUnusedRepoLines := []string{"echo 'Limiting package databases to the selected profile and the msys repository.'"}
	for _, repoName := range mingwRepoNames {
		if repoName == selectedRepoName {
			continue
		}
		disableUnusedRepoLines = append(disableUnusedRepoLines,
			fmt.Sprintf("sed -i -E '/^\\[%s\\]/,/^Include/ s/^/#/' /etc/pacman.conf", repoName))
	}

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package servers (with fallback mirrors) for this private toolchain.'",
	}
	scriptLines = append(scriptLines, mirrorlistSetupLines()...)
	scriptLines = append(scriptLines,
		"echo 'Initializing the private MSYS2 package keyring.'",
		"pacman-key --init 2>&1",
		"pacman-key --populate msys2 2>&1",
		"echo 'Preparing pacman database signature policy.'",
		"# MSYS2 packages remain signature-checked. Repository database signatures are treated as optional, matching normal MSYS2 pacman behavior and avoiding false failures when a database .sig is not served or is cleared during refresh.",
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
	)
	scriptLines = append(scriptLines, disableUnusedRepoLines...)
	scriptLines = append(scriptLines,
		"echo 'Clearing stale or half-downloaded package databases before refresh.'",
		"rm -f /var/lib/pacman/sync/*.db /var/lib/pacman/sync/*.db.sig /var/lib/pacman/sync/*.files /var/lib/pacman/sync/*.files.sig /var/cache/pacman/pkg/*.part",
		"echo 'Refreshing the private MSYS2 package databases and keyring package.'",
		"pacman -Syy --needed --noconfirm msys2-keyring",
		"echo 'Installing the approved build-tool packages.'",
		"pacman -S --needed --noconfirm "+strings.Join(quotedPackageNames, " "),
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
	)
	return scriptLines, nil
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
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		"echo 'Using the official MSYS2 package servers (with fallback mirrors) for FFmpeg library packages.'",
	}
	scriptLines = append(scriptLines, mirrorlistSetupLines()...)
	scriptLines = append(scriptLines,
		"sed -i -E 's/^SigLevel[[:space:]]*=.*/SigLevel = Required DatabaseOptional/' /etc/pacman.conf",
		"echo 'Clearing half-downloaded package files before installing FFmpeg library packages.'",
		"rm -f /var/cache/pacman/pkg/*.part",
		// Full refresh + upgrade (-Syyu), not a bare -Syy, so the install is not a partial
		// upgrade: MSYS2 is rolling, and installing a new package against a stale set of
		// already-installed dependencies can pull a dependency version whose file the mirror
		// has already superseded (404). A full upgrade keeps every package coherent.
		"echo 'Refreshing databases and upgrading the environment before installing FFmpeg library packages.'",
		"pacman -Syyu --needed --noconfirm",
		"echo 'Installing MSYS2 packages required by the selected FFmpeg libraries.'",
		// --overwrite '*' lets pacman take ownership of untracked files in the private,
		// rebuildable prefix. A prior library prep (e.g. AviSynth+) may have written headers
		// such as avisynth/avisynth_c.h that an MSYS2 package (libx264) also ships; without
		// --overwrite, pacman aborts with "conflicting files" on a dirty/re-run prefix.
		// Install ordering re-runs the prep afterward, so the prep-provided files win in the end.
		"pacman -S --noconfirm --overwrite '*' "+strings.Join(quotedPackageNames, " "),
	)
	scriptLines = append(scriptLines, sanitizeRabbitmqPkgConfigLines()...)
	return scriptLines, nil
}

// sanitizeRabbitmqPkgConfigLines repairs the MSYS2 librabbitmq.pc. The rabbitmq-c
// CMake build composes Libs.private by looping its socket libraries as "-l<lib>"; an
// empty list element leaves a stray, name-less "-l" token (the shipped 0.15 .pc reads
// "Libs.private: ... -lws2_32 -l -lssl -lcrypto"). With the build's default
// --pkg-config-flags=--static, FFmpeg's configure link probe passes that bare "-l" to
// gcc, the link fails, and configure reports "librabbitmq >= 0.7.1 not found".
// Stripping the empty "-l" namespec from the Libs lines makes the static link valid.
// Guarded by file existence, so it is a no-op when rabbitmq-c was not installed.
func sanitizeRabbitmqPkgConfigLines() []string {
	return []string{
		`rabbitmq_pc="${MSYSTEM_PREFIX:-/ucrt64}/lib/pkgconfig/librabbitmq.pc"`,
		`if [ -f "${rabbitmq_pc}" ]; then`,
		`  echo "Repairing librabbitmq.pc: removing empty -l tokens that break the static link probe."`,
		`  sed -i -E '/^Libs/ s/ -l( |$)/\1/g' "${rabbitmq_pc}"`,
		`  echo "Patched librabbitmq.pc: $(grep -E '^Libs' "${rabbitmq_pc}")"`,
		`fi`,
	}
}

// enableLegacyOpencvDetectionLines makes FFmpeg's libopencv detection work against MSYS2's
// OpenCV 4 package on FFmpeg releases <= 8.0. Those configure scripts probe the legacy
// pkg-config module name "opencv" (OpenCV 2.x) and run a bare `check_headers
// opencv2/core/core_c.h` (no pkg-config cflags) before any pkg-config branch. MSYS2 ships the
// OpenCV 4 headers under .../include/opencv4 (off the default search path) and registers only
// opencv4.pc, so BOTH halves of the probe fail: the bare header check cannot reach core_c.h,
// and the pkg-config branches cannot resolve the "opencv" module. The probe then falls through
// to its last resort, which requires opencv/cxcore.h — an OpenCV 1.x/2.x header absent from
// OpenCV 4 — and configure aborts with "opencv not found using pkg-config" plus the core_c.h
// and cxcore.h "No such file or directory" errors. FFmpeg 8.1 probes "opencv4" directly (its
// cflags carry the include dir) and is unaffected.
//
// OpenCV 4 still ships the legacy C API libopencv targets (core_c.h + cvCreateImageHeader —
// FFmpeg's own opencv4 link check confirms the symbol), so no separate OpenCV 2.x/3.x package
// is needed (and MSYS2 ships none). Two steps make the installed OpenCV 4 satisfy the legacy
// probe: (1) copy opencv4.pc to opencv.pc so the pkg-config branch resolves under the old name,
// and (2) add the include/opencv4 dir to --extra-cflags so the leading bare check_headers
// (and every later configure compile test) can reach opencv2/core/core_c.h. With both, the
// `require libopencv ... -lopencv_core -lopencv_imgproc` branch links and libopencv enables.
// Both steps are guarded so they are no-ops when OpenCV 4 is absent (or an opencv.pc already
// exists), which also keeps them harmless on FFmpeg 8.1+.
func enableLegacyOpencvDetectionLines() []string {
	return []string{
		`opencv4_pc="${MSYSTEM_PREFIX:-/ucrt64}/lib/pkgconfig/opencv4.pc"`,
		`opencv_pc="${MSYSTEM_PREFIX:-/ucrt64}/lib/pkgconfig/opencv.pc"`,
		`if [ -f "${opencv4_pc}" ] && [ ! -f "${opencv_pc}" ]; then`,
		`  echo "Aliasing opencv4.pc to opencv.pc so FFmpeg <= 8.0 libopencv detection finds MSYS2 OpenCV 4."`,
		`  cp "${opencv4_pc}" "${opencv_pc}"`,
		`fi`,
		`opencv_incdir="${MSYSTEM_PREFIX:-/ucrt64}/include/opencv4"`,
		`if [ -d "${opencv_incdir}" ]; then`,
		`  echo "Adding ${opencv_incdir} to FFmpeg extra-cflags so the libopencv core_c.h header check passes (OpenCV 4 installs its headers under include/opencv4, off the default search path)."`,
		`  configure_flags+=("--extra-cflags=-I${opencv_incdir}")`,
		`fi`,
	}
}

// ConfigureScriptLines builds the FFmpeg configure script. privatePkgConfigDirs are the
// pkgconfig directories of any privately-installed libraries (e.g. libtls); they are
// prepended to the script's exported PKG_CONFIG_PATH/PKG_CONFIG_LIBDIR so pkg-config finds
// those isolated modules (and their private libssl/libcrypto) ahead of the shared prefix.
// ffmpegVersion is the release being built (e.g. "8.1.2"); the pre-configure pkg-config
// version floors are resolved from that release's support manifest, so an older FFmpeg gets
// its own (usually lower) floors rather than a single hardcoded set. A snapshot/unknown
// version resolves no floors and the build relies on FFmpeg configure as the backstop.
func ConfigureScriptLines(configureFlags []string, privatePkgConfigDirs []string, ffmpegVersion string) ([]string, error) {
	quotedConfigureFlags := make([]string, 0, len(configureFlags))
	for _, configureFlag := range configureFlags {
		if err := ValidateConfigureFlag(configureFlag); err != nil {
			return nil, err
		}
		quotedConfigureFlags = append(quotedConfigureFlags, shellQuote(configureFlag))
	}
	quotedConfigureFlagArray := strings.Join(quotedConfigureFlags, " ")

	pkgConfigModules := pkgConfigModulesForConfigureFlags(configureFlags, ffmpegVersion)

	// Prepend any private pkgconfig dirs to the shared prefix search path. Each is a unix
	// path under the profile prefix produced by PrivateLibraryPkgConfigDir.
	privatePkgConfigPrefix := ""
	for _, privateDir := range privatePkgConfigDirs {
		if !safePkgConfigDirPattern.MatchString(privateDir) {
			return nil, fmt.Errorf("private pkgconfig dir is unsafe: %s", privateDir)
		}
		privatePkgConfigPrefix += privateDir + ":"
	}

	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		`export PATH="${profile_prefix}/bin:/usr/bin:${PATH}"`,
		`export PKG_CONFIG_PATH="` + privatePkgConfigPrefix + `${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export PKG_CONFIG_LIBDIR="` + privatePkgConfigPrefix + `${profile_prefix}/lib/pkgconfig:${profile_prefix}/share/pkgconfig:/usr/lib/pkgconfig:/usr/share/pkgconfig"`,
		`export CPPFLAGS="-I${profile_prefix}/include"`,
		`if [ -x "${profile_prefix}/bin/pkgconf.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkgconf.exe"; elif [ -x "${profile_prefix}/bin/pkg-config.exe" ]; then export PKG_CONFIG="${profile_prefix}/bin/pkg-config.exe"; else export PKG_CONFIG="pkg-config"; fi`,
		`echo "Using pkg-config command: ${PKG_CONFIG}"`,
		`echo "Using pkg-config search paths: ${PKG_CONFIG_LIBDIR}"`,
		"configure_flags=(" + quotedConfigureFlagArray + ")",
		`svtjpegxs_disabled=0`,
		`lensfun_disabled=0`,
		`vapoursynth_disabled=0`,
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
		`vapoursynth_api_probe() {`,
		`  probe_dir="$(mktemp -d)"`,
		`  cat > "${probe_dir}/vapoursynth_api_probe.c" <<'EOF'`,
		`#include <VSScript4.h>`,
		`#include <VapourSynth4.h>`,
		`int main(void){ const VSSCRIPTAPI *api = getVSScriptAPI(VSSCRIPT_API_VERSION); (void)api; return 0; }`,
		`EOF`,
		`  echo "VapourSynth FFmpeg API probe starting."`,
		`  if gcc $("${PKG_CONFIG}" --cflags vapoursynth-script) "${probe_dir}/vapoursynth_api_probe.c" $("${PKG_CONFIG}" --libs vapoursynth-script) -o "${probe_dir}/vapoursynth_api_probe.exe" 2>&1; then`,
		`    echo "VapourSynth FFmpeg API probe passed."`,
		`    rm -rf "${probe_dir}"`,
		`    return 0`,
		`  fi`,
		`  echo "VapourSynth FFmpeg API probe failed."`,
		`  rm -rf "${probe_dir}"`,
		`  return 1`,
		`}`,
		`try_enable_vapoursynth() {`,
		`  echo "VapourSynth compatibility check starting at $(date +%T)"`,
		`  echo "VapourSynth is hidden from automatic presets. The MSYS2 package may be older than this FFmpeg source requires."`,
		`  if ! "${PKG_CONFIG}" --exists vapoursynth-script; then`,
		`    echo "WARNING: VapourSynth (vapoursynth-script) is not visible to pkg-config. Disabling --enable-vapoursynth and continuing."`,
		`    return 1`,
		`  fi`,
		`  echo "vapoursynth-script pkg-config version: $("${PKG_CONFIG}" --modversion vapoursynth-script 2>&1 || true)"`,
		`  if vapoursynth_api_probe; then`,
		`    echo "VapourSynth package exposes the FFmpeg-required API. Keeping --enable-vapoursynth."`,
		`    return 0`,
		`  fi`,
		`  echo "WARNING: VapourSynth package was found, but it does not expose the API used by this FFmpeg source (the MSYS2 package is likely older than required). Disabling --enable-vapoursynth and continuing."`,
		`  return 1`,
		`}`,
	}
	if configureFlagSelected(configureFlags, "--enable-libsvtjpegxs") {
		scriptLines = append(scriptLines, "if ! try_enable_svt_jpeg_xs; then svtjpegxs_disabled=1; remove_configure_flag --enable-libsvtjpegxs; fi")
	}
	if configureFlagSelected(configureFlags, "--enable-liblensfun") {
		scriptLines = append(scriptLines, "if ! try_enable_lensfun; then lensfun_disabled=1; remove_configure_flag --enable-liblensfun; fi")
	}
	if configureFlagSelected(configureFlags, "--enable-vapoursynth") {
		scriptLines = append(scriptLines, "if ! try_enable_vapoursynth; then vapoursynth_disabled=1; remove_configure_flag --enable-vapoursynth; fi")
	}
	if configureFlagSelected(configureFlags, "--enable-libopencv") {
		scriptLines = append(scriptLines, enableLegacyOpencvDetectionLines()...)
	}
	if len(pkgConfigModules) > 0 {
		scriptLines = append(scriptLines, "echo 'Diagnosing selected pkg-config libraries before FFmpeg configure.'")
		for _, mod := range pkgConfigModules {
			diagnosticLine := fmt.Sprintf("diagnose_pkg_config_module %s %s", shellQuote(mod.Name), shellQuote(mod.MinVersion))
			if mod.Name == "SvtJpegxs" {
				diagnosticLine = "if [ \"${svtjpegxs_disabled}\" -eq 0 ]; then " + diagnosticLine + "; else echo 'SVT JPEG XS pkg-config diagnostic skipped because --enable-libsvtjpegxs was disabled after compatibility checks.'; fi"
			}
			if mod.Name == "vapoursynth-script" {
				diagnosticLine = "if [ \"${vapoursynth_disabled}\" -eq 0 ]; then " + diagnosticLine + "; else echo 'VapourSynth pkg-config diagnostic skipped because --enable-vapoursynth was disabled after compatibility checks.'; fi"
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

func pkgConfigModulesForConfigureFlags(configureFlags []string, ffmpegVersion string) []pkgConfigModule {
	// Only list libraries that are expected to provide a pkg-config module in MSYS2.
	// Some valid FFmpeg options, such as --enable-libgsm, are probed by FFmpeg
	// through headers/libraries instead of a .pc file. Pre-checking those with
	// pkg-config incorrectly blocks valid builds.
	//
	// name is the pkg-config module name (often different from the catalog library id, e.g.
	// SvtAv1Enc vs svt-av1). libraryId is the catalog id used to look up FFmpeg's pkg-config
	// minimum for the release being built, from the per-release support manifest. The floor is
	// resolved per release (not a single hardcoded set), so an older FFmpeg gets its own lower
	// floor and an older package that satisfies it is no longer falsely rejected before
	// configure. A flag with no libraryId, an unmanifested/snapshot version, or a library the
	// release pins no minimum for, carries no floor and is only checked for existence.
	type entry struct {
		name      string
		libraryId string
	}
	moduleByFlag := map[string]entry{
		"--enable-libaom":            {name: "aom", libraryId: "aom"},
		"--enable-libass":            {name: "libass", libraryId: "ass"},
		"--enable-libbluray":         {name: "libbluray", libraryId: "bluray"},
		"--enable-libcdio":           {name: "libcdio", libraryId: "cdio"},
		"--enable-libdav1d":          {name: "dav1d", libraryId: "dav1d"},
		"--enable-libdvdnav":         {name: "dvdnav", libraryId: "dvdnav"},
		"--enable-libkvazaar":        {name: "kvazaar", libraryId: "kvazaar"},
		"--enable-libonnxruntime":    {name: "libonnxruntime", libraryId: "onnxruntime"},
		"--enable-vapoursynth":       {name: "vapoursynth-script", libraryId: "vapoursynth"},
		"--enable-libfdk-aac":        {name: "fdk-aac", libraryId: "fdk-aac"},
		"--enable-libfontconfig":     {name: "fontconfig", libraryId: "fontconfig"},
		"--enable-fontconfig":        {name: "fontconfig", libraryId: "fontconfig"},
		"--enable-libfreetype":       {name: "freetype2", libraryId: "freetype"},
		"--enable-libfribidi":        {name: "fribidi", libraryId: "fribidi"},
		"--enable-libharfbuzz":       {name: "harfbuzz", libraryId: "harfbuzz"},
		"--enable-libilbc":           {name: "libilbc", libraryId: "ilbc"},
		"--enable-liblensfun":        {name: "lensfun", libraryId: "lensfun"},
		"--enable-libjxl":            {name: "libjxl", libraryId: "libjxl"},
		"--enable-libmodplug":        {name: "libmodplug", libraryId: "modplug"},
		"--enable-libmpeghdec":       {name: "mpeghdec", libraryId: "mpeghdec"},
		"--enable-libmp3lame":        {name: "lame", libraryId: "mp3lame"},
		"--enable-libopencore-amrnb": {name: "opencore-amrnb", libraryId: "opencore-amr"},
		"--enable-libopencore-amrwb": {name: "opencore-amrwb", libraryId: "opencore-amr"},
		"--enable-libopenh264":       {name: "openh264", libraryId: "openh264"},
		"--enable-libopenjpeg":       {name: "libopenjp2", libraryId: "openjpeg"},
		"--enable-libopus":           {name: "opus", libraryId: "opus"},
		"--enable-libplacebo":        {name: "libplacebo", libraryId: "libplacebo"},
		"--enable-librav1e":          {name: "rav1e", libraryId: "rav1e"},
		"--enable-librubberband":     {name: "rubberband", libraryId: "rubberband"},
		"--enable-libsoxr":           {name: "soxr", libraryId: "soxr"},
		"--enable-libspeex":          {name: "speex", libraryId: "speex"},
		"--enable-libssh":            {name: "libssh", libraryId: "ssh"},
		"--enable-libsvtav1":         {name: "SvtAv1Enc", libraryId: "svt-av1"},
		"--enable-libtesseract":      {name: "tesseract", libraryId: "tesseract"},
		"--enable-libtwolame":        {name: "twolame", libraryId: "twolame"},
		"--enable-libvmaf":           {name: "libvmaf", libraryId: "vmaf"},
		"--enable-libvorbis":         {name: "vorbis", libraryId: "vorbis"},
		"--enable-libvpx":            {name: "vpx", libraryId: "libvpx"},
		"--enable-libwebp":           {name: "libwebp", libraryId: "webp"},
		"--enable-libx264":           {name: "x264", libraryId: "x264"},
		"--enable-libx265":           {name: "x265", libraryId: "x265"},
		"--enable-libxavs2":          {name: "xavs2", libraryId: "xavs2"},
		"--enable-libzimg":           {name: "zimg", libraryId: "zimg"},
		"--enable-openal":            {name: "openal", libraryId: "openal"},
		"--enable-openssl":           {name: "openssl", libraryId: "openssl"},
		"--enable-gnutls":            {name: "gnutls", libraryId: "gnutls"},
		"--enable-sdl2":              {name: "sdl2", libraryId: "sdl2"},
		"--enable-chromaprint":       {name: "libchromaprint", libraryId: "chromaprint"},
		"--enable-libaribcaption":    {name: "libaribcaption", libraryId: "aribcaption"},
		"--enable-libbs2b":           {name: "libbs2b", libraryId: "bs2b"},
		"--enable-libcaca":           {name: "caca", libraryId: "caca"},
		"--enable-libdvdread":        {name: "dvdread", libraryId: "dvdread"},
		"--enable-libmysofa":         {name: "libmysofa", libraryId: "mysofa"},
		"--enable-libopencolorio":    {name: "OpenColorIO", libraryId: "opencolorio"},
		"--enable-libopencv":         {name: "opencv4", libraryId: "opencv"},
		"--enable-libqrencode":       {name: "libqrencode", libraryId: "qrencode"},
		"--enable-librabbitmq":       {name: "librabbitmq", libraryId: "rabbitmq"},
		"--enable-librsvg":           {name: "librsvg-2.0", libraryId: "rsvg"},
		"--enable-libsvtjpegxs":      {name: "SvtJpegxs", libraryId: "svtjpegxs"},
		"--enable-liblc3":            {name: "lc3", libraryId: "lc3"},
		"--enable-lv2":               {name: "lilv-0", libraryId: "lv2"},
		"--enable-lcms2":             {name: "lcms2", libraryId: "lcms2"},
		"--enable-opencl":            {name: "OpenCL", libraryId: "opencl"},
		"--enable-whisper":           {name: "whisper", libraryId: "whisper"},
	}
	release, manifested := releasesupport.ResolveReleaseSupport(ffmpegVersion)
	floorFor := func(libraryId string) string {
		if !manifested || libraryId == "" {
			return ""
		}
		support, supported := release.LibrarySupportFor(libraryId)
		if !supported {
			return ""
		}
		return support.MinVersion
	}
	modules := []pkgConfigModule{}
	seen := map[string]bool{}
	for _, configureFlag := range configureFlags {
		e, exists := moduleByFlag[configureFlag]
		if !exists || seen[e.name] {
			continue
		}
		seen[e.name] = true
		modules = append(modules, pkgConfigModule{Name: e.name, MinVersion: floorFor(e.libraryId)})
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

// LibraryBuildSpec is the scripting-layer view of one non-Native library preparation.
// It is mapped from planning.LibraryPreparation by the app layer (scripting cannot
// import planning without an import cycle). The generated scripts operate only on the
// already-downloaded, already-extracted archive contents in the script working
// directory and install into the selected MSYS2 profile prefix; no URL or raw path
// from outside the workspace ever enters the shell.
type LibraryBuildSpec struct {
	LibraryId   string
	DisplayName string
	// BuildSystem selects the Internal-track source-build generator ("cmake",
	// "autotools", "make"). Empty is treated as cmake. Ignored for external imports.
	BuildSystem string
	// CFlags are extra C compiler flags exported as CFLAGS for the build (e.g. demoting a GCC-14
	// hard error to a warning for an older C library that predates it). Honored by the meson
	// generator. Each is validated against safeCompilerFlagPattern.
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
	// PkgConfigLibsLine, when set, replaces the installed .pc's entire "Libs:" line value.
	// Used when -l<name> would resolve to a same-named shared import library (.dll.a) that
	// shadows this recipe's own static archive in a shared prefix; forcing -l:lib<name>.a
	// (or an absolute archive path) makes the link pick the static archive instead. May
	// reference ${libdir}. Mutually independent of PkgConfigAppendLibs.
	PkgConfigLibsLine string
	// PrivatePrefixInstall installs the library into its own per-library prefix
	// (<profile>/PrivateLibraryInstallSubdir/<PkgConfigName>) instead of the shared MSYS2
	// prefix, and strips the installed .pc's Requires/Requires.private so no transitive
	// module re-adds a shared-prefix archive. Used to isolate a library whose archive base
	// names collide with another package's (e.g. LibreSSL libtls vs the openssl package).
	PrivatePrefixInstall     bool
	VerifyHeaderRelativePath string
	VerifyLibStem            string
	SourcePatches            []LibrarySourcePatch
	// GeneratedSourceFiles are files written into the extracted source tree before configure,
	// for recipes whose build expects a file that the release tarball omits because upstream
	// generates it from a .git checkout (e.g. libvmaf's vcs_version.h, emitted by meson vcs_tag
	// only when a .git dir is present). Supplies what no build flag or single-line patch can.
	GeneratedSourceFiles []GeneratedSourceFile
}

// GeneratedSourceFile is one file the recipe writes into the extracted source tree before
// configure. Path is relative to the source root; Lines are the file's lines, written
// verbatim. Path is validated as a safe relative path and each line must contain no single
// quote or newline, so the lines are safe to single-quote into the generated printf.
type GeneratedSourceFile struct {
	Path  string
	Lines []string
}

// PrivateLibraryInstallSubdir is the prefix-relative directory under the MSYS2 profile
// prefix that holds per-library private installs (see LibraryBuildSpec.PrivatePrefixInstall).
// A library installs into <profile>/PrivateLibraryInstallSubdir/<PkgConfigName>. Shared by
// the build-script generator (which installs there) and the app (which prepends that
// prefix's pkgconfig dir to the FFmpeg configure PKG_CONFIG_PATH); keep the two in sync.
const PrivateLibraryInstallSubdir = "opt/customffmpeg"

// PrivateLibraryPkgConfigDir returns the unix pkgconfig directory of a privately-installed
// library, given the MSYS2 profile's unix prefix (e.g. "/ucrt64") and the library's
// pkg-config module name. The configure step adds this to PKG_CONFIG_PATH.
func PrivateLibraryPkgConfigDir(profileUnixPrefix string, pkgConfigName string) string {
	return profileUnixPrefix + "/" + PrivateLibraryInstallSubdir + "/" + pkgConfigName + "/lib/pkgconfig"
}

// LibrarySourcePatch is one exact full-line replacement applied to the extracted source
// tree before configure/build, for recipes that must work around an upstream portability
// bug that no build flag can fix. File is relative to the source root; Find is matched as
// a whole line and replaced with Replace. Find/Replace must contain no single quote,
// backslash, or newline so they are safe to single-quote into the generated awk command.
type LibrarySourcePatch struct {
	File    string
	Find    string
	Replace string
}

var safeLibraryPathSegmentPattern = regexp.MustCompile(`^[A-Za-z0-9_+-]+$`)
var safeLibraryHeaderPathPattern = regexp.MustCompile(`^[A-Za-z0-9_./+-]+$`)
var safeCMakeOptionPattern = regexp.MustCompile(`^-D[A-Za-z0-9_]+=[A-Za-z0-9_./:+-]*$`)
var safeCMakeTargetPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var safeConfigureOptionPattern = regexp.MustCompile(`^--[A-Za-z0-9][A-Za-z0-9=._/+-]*$`)

// safeCompilerFlagPattern matches a single C compiler flag exported as CFLAGS (e.g.
// -Wno-error=implicit-function-declaration). No spaces or shell metacharacters, so the joined
// flags are safe to interpolate into the CFLAGS= assignment in the generated script.
var safeCompilerFlagPattern = regexp.MustCompile(`^-[A-Za-z0-9][A-Za-z0-9=._+-]*$`)
var safeMakeTargetPattern = regexp.MustCompile(`^[A-Za-z0-9_.-]+$`)
var safeMakeVariablePattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*=[A-Za-z0-9_./:= +-]*$`)
var safePkgConfigLibsLinePattern = regexp.MustCompile(`^[A-Za-z0-9_:./${} +-]+$`)

// safePkgConfigDirPattern matches a unix pkgconfig directory under the MSYS2 profile prefix
// (e.g. /ucrt64/opt/customffmpeg/libtls/lib/pkgconfig). No spaces, no shell metacharacters,
// so it is safe to interpolate into the exported PKG_CONFIG_PATH.
var safePkgConfigDirPattern = regexp.MustCompile(`^/[A-Za-z0-9_./-]+$`)

func validateLibraryBuildSpec(spec LibraryBuildSpec, requireImportSubdirs bool) error {
	if !safeLibraryPathSegmentPattern.MatchString(spec.LibraryId) {
		return fmt.Errorf("library preparation id contains unsafe characters: %s", spec.LibraryId)
	}
	if spec.VerifyLibStem != "" && !safeLibraryPathSegmentPattern.MatchString(spec.VerifyLibStem) {
		return fmt.Errorf("library preparation lib stem contains unsafe characters: %s", spec.VerifyLibStem)
	}
	if spec.VerifyHeaderRelativePath == "" {
		return errors.New("library preparation verify header path is empty")
	}
	if !safeLibraryHeaderPathPattern.MatchString(spec.VerifyHeaderRelativePath) || strings.Contains(spec.VerifyHeaderRelativePath, "..") {
		return fmt.Errorf("library preparation verify header path is unsafe: %s", spec.VerifyHeaderRelativePath)
	}
	for _, cmakeOption := range spec.CMakeOptions {
		if !safeCMakeOptionPattern.MatchString(cmakeOption) {
			return fmt.Errorf("library preparation cmake option is unsafe: %s", cmakeOption)
		}
	}
	for _, buildTarget := range spec.CMakeBuildTargets {
		if !safeCMakeTargetPattern.MatchString(buildTarget) {
			return fmt.Errorf("library preparation cmake build target is unsafe: %s", buildTarget)
		}
	}
	if spec.PkgConfigName != "" && !safeLibraryPathSegmentPattern.MatchString(spec.PkgConfigName) {
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
		if patch.File == "" || !safeLibraryHeaderPathPattern.MatchString(patch.File) || strings.Contains(patch.File, "..") {
			return fmt.Errorf("library preparation source patch file is unsafe: %q", patch.File)
		}
		if patch.Find == "" {
			return fmt.Errorf("library preparation source patch find is empty for %q", patch.File)
		}
		if strings.ContainsAny(patch.Find, "'\\\n\r") || strings.ContainsAny(patch.Replace, "'\\\n\r") {
			return fmt.Errorf("library preparation source patch contains unsafe characters for %q", patch.File)
		}
	}
	for _, generated := range spec.GeneratedSourceFiles {
		if generated.Path == "" || !safeLibraryHeaderPathPattern.MatchString(generated.Path) || strings.Contains(generated.Path, "..") {
			return fmt.Errorf("library preparation generated source file path is unsafe: %q", generated.Path)
		}
		for _, line := range generated.Lines {
			if strings.ContainsAny(line, "'\n\r") {
				return fmt.Errorf("library preparation generated source file %q has unsafe line content", generated.Path)
			}
		}
	}
	for _, appendLib := range spec.PkgConfigAppendLibs {
		if !safeLibraryPathSegmentPattern.MatchString(appendLib) {
			return fmt.Errorf("library preparation pkg-config append lib is unsafe: %s", appendLib)
		}
	}
	if spec.PkgConfigLibsLine != "" && !safePkgConfigLibsLinePattern.MatchString(spec.PkgConfigLibsLine) {
		return fmt.Errorf("library preparation pkg-config Libs override is unsafe: %s", spec.PkgConfigLibsLine)
	}
	if spec.ConfigureSubdir != "" && (!safeLibraryHeaderPathPattern.MatchString(spec.ConfigureSubdir) || strings.Contains(spec.ConfigureSubdir, "..")) {
		return fmt.Errorf("library preparation configure subdir is unsafe: %s", spec.ConfigureSubdir)
	}
	for _, configureOption := range spec.ConfigureOptions {
		if !safeConfigureOptionPattern.MatchString(configureOption) {
			return fmt.Errorf("library preparation configure option is unsafe: %s", configureOption)
		}
	}
	for _, compilerFlag := range spec.CFlags {
		if !safeCompilerFlagPattern.MatchString(compilerFlag) {
			return fmt.Errorf("library preparation cflag is unsafe: %s", compilerFlag)
		}
	}
	for _, makeTarget := range append(append([]string{}, spec.MakeBuildTargets...), spec.MakeInstallTargets...) {
		if !safeMakeTargetPattern.MatchString(makeTarget) {
			return fmt.Errorf("library preparation make target is unsafe: %s", makeTarget)
		}
	}
	for _, headerFile := range spec.MakeInstallHeaderFiles {
		if headerFile == "" || !safeLibraryHeaderPathPattern.MatchString(headerFile) || strings.Contains(headerFile, "..") {
			return fmt.Errorf("library preparation make install header file is unsafe: %q", headerFile)
		}
	}
	if spec.MakeStaticLibFile != "" && (!safeLibraryHeaderPathPattern.MatchString(spec.MakeStaticLibFile) || strings.Contains(spec.MakeStaticLibFile, "..")) {
		return fmt.Errorf("library preparation make static lib file is unsafe: %q", spec.MakeStaticLibFile)
	}
	for _, makeVariable := range spec.MakeVariables {
		if !safeMakeVariablePattern.MatchString(makeVariable) {
			return fmt.Errorf("library preparation make variable is unsafe: %q", makeVariable)
		}
	}
	if requireImportSubdirs {
		for _, subdir := range []string{spec.ImportIncludeSubdir, spec.ImportLibSubdir} {
			if subdir == "" || !safeLibraryHeaderPathPattern.MatchString(subdir) || strings.Contains(subdir, "..") {
				return fmt.Errorf("library preparation import subdir is unsafe: %q", subdir)
			}
		}
	}
	return nil
}

// InternalLibrarySourceBuildScriptLines builds an Internal-track library from its
// verified, already-extracted upstream source (the script working directory) and
// installs it into the selected MSYS2 profile prefix, then verifies the installed
// header and link library exist. It dispatches on the recipe's build system so a new
// build system is added in exactly one place without touching existing recipes.
func InternalLibrarySourceBuildScriptLines(spec LibraryBuildSpec) ([]string, error) {
	switch spec.BuildSystem {
	case "", "cmake":
		return internalCMakeBuildScriptLines(spec)
	case "configure-make":
		return internalConfigureMakeBuildScriptLines(spec)
	case "make":
		return internalMakeBuildScriptLines(spec)
	case "meson":
		return internalMesonBuildScriptLines(spec)
	default:
		return nil, fmt.Errorf("unknown internal-track build system %q for %s", spec.BuildSystem, spec.LibraryId)
	}
}

// internalMesonBuildScriptLines builds a library configured with `meson setup` and built
// with ninja (e.g. libvmaf), installing into the selected MSYS2 profile prefix. The meson
// `-Dname=value` project options reuse the spec's CMakeOptions field (same option syntax,
// same validation); --buildtype=release and --default-library=static are intrinsic to how
// this builder produces the static archive FFmpeg links. ConfigureSubdir is the source
// directory holding meson.build when it is not the repo root (libvmaf keeps it in libvmaf/).
// Like the CMake path, meson here is the native mingw tool, so a unix prefix such as /ucrt64
// would install to the literal drive-root; the prefix is converted to its Windows form with
// cygpath (and excluded from MSYS2 path mangling) so meson installs into the real prefix.
func internalMesonBuildScriptLines(spec LibraryBuildSpec) ([]string, error) {
	if err := validateLibraryBuildSpec(spec, false); err != nil {
		return nil, err
	}
	mesonOptionArray := make([]string, 0, len(spec.CMakeOptions))
	for _, mesonOption := range spec.CMakeOptions {
		mesonOptionArray = append(mesonOptionArray, shellQuote(mesonOption))
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
		mesonCFlagsExport = `CFLAGS=` + shellQuote(strings.Join(spec.CFlags, " ")) + ` `
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		// The meson path installs into the shared prefix only (PrivatePrefixInstall is gated
		// to cmake by validateLibraryBuildSpec), so install_prefix is always the shared prefix
		// and the pkg-config patch and verification helpers resolve against it.
		`install_prefix="${profile_prefix}"`,
		"echo " + shellQuote("Preparing internal-track library from source: "+spec.DisplayName),
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
	scriptLines = append(scriptLines, generatedSourceFileScriptLines(spec)...)
	scriptLines = append(scriptLines, sourcePatchScriptLines(spec)...)
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
	scriptLines = append(scriptLines, pkgConfigLibsLineOverrideLines(spec)...)
	scriptLines = append(scriptLines, pkgConfigAppendLibsLines(spec)...)
	scriptLines = append(scriptLines, stripPkgConfigRequiresLines(spec)...)
	scriptLines = append(scriptLines, libraryInstallVerificationLines(spec)...)
	scriptLines = append(scriptLines, "echo "+shellQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

func internalCMakeBuildScriptLines(spec LibraryBuildSpec) ([]string, error) {
	if err := validateLibraryBuildSpec(spec, false); err != nil {
		return nil, err
	}
	cmakeOptionArray := make([]string, 0, len(spec.CMakeOptions))
	for _, cmakeOption := range spec.CMakeOptions {
		cmakeOptionArray = append(cmakeOptionArray, shellQuote(cmakeOption))
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		"echo " + shellQuote("Preparing internal-track library from source: "+spec.DisplayName),
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
		installPrefixSetupLine(spec),
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
	scriptLines = append(scriptLines, generatedSourceFileScriptLines(spec)...)
	scriptLines = append(scriptLines, sourcePatchScriptLines(spec)...)
	scriptLines = append(scriptLines, `cmake -S "${src_dir}" -B "${build_dir}" -G Ninja -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX="${install_prefix_win}" -DCMAKE_POLICY_VERSION_MINIMUM=3.5.0 "${cmake_options[@]}" -Wno-dev 2>&1`)
	// Build either the named targets a recipe requests (e.g. a header-only project whose
	// generated headers come from a target that is not in the default build) or, by
	// default, the default target.
	if len(spec.CMakeBuildTargets) > 0 {
		for _, buildTarget := range spec.CMakeBuildTargets {
			scriptLines = append(scriptLines, `cmake --build "${build_dir}" --target `+shellQuote(buildTarget)+` 2>&1`)
		}
	} else {
		scriptLines = append(scriptLines, `cmake --build "${build_dir}" 2>&1`)
	}
	scriptLines = append(scriptLines, `cmake --install "${build_dir}" 2>&1`)
	scriptLines = append(scriptLines, pkgConfigLibsLineOverrideLines(spec)...)
	scriptLines = append(scriptLines, pkgConfigAppendLibsLines(spec)...)
	scriptLines = append(scriptLines, stripPkgConfigRequiresLines(spec)...)
	scriptLines = append(scriptLines, libraryInstallVerificationLines(spec)...)
	scriptLines = append(scriptLines, "echo "+shellQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

// pkgConfigAppendLibsLines patches the installed pkg-config module so its Libs line ends
// with the recipe's extra link libraries. Used to repair a static .pc that lists the
// C++/math runtime before its own static archives, which breaks GNU static link order.
// The pkg-config name and lib names are validated as safe path segments by
// validateLibraryBuildSpec, so they are safe to interpolate into the script. Returns no
// lines when the recipe declares no fixup.
// sourcePatchScriptLines emits shell that applies each recipe source patch to the
// extracted tree. Each patch replaces an exact full line (Find) with Replace in File
// (relative to ${src_dir}). The step fails loudly when the target line is absent, so a
// re-pinned upstream release that changed the file is caught here instead of silently
// building unpatched. Find/Replace are validated to contain no quote, backslash, or
// newline, so single-quoting them into grep/awk is safe. Returns no lines when the
// recipe declares no patches.
func sourcePatchScriptLines(spec LibraryBuildSpec) []string {
	lines := []string{}
	for _, patch := range spec.SourcePatches {
		lines = append(lines,
			`patch_file="${src_dir}/`+patch.File+`"`,
			`if ! grep -qxF -- '`+patch.Find+`' "${patch_file}"; then echo "ERROR: source patch target line not found in `+patch.File+`; the pinned upstream release may have changed."; exit 1; fi`,
			`awk -v patch_find='`+patch.Find+`' -v patch_repl='`+patch.Replace+`' '{ if ($0 == patch_find) { print patch_repl } else { print } }' "${patch_file}" > "${patch_file}.patched" && mv "${patch_file}.patched" "${patch_file}"`,
			`echo "Applied source patch to `+patch.File+`"`,
		)
	}
	return lines
}

// generatedSourceFileScriptLines emits shell that writes each recipe-declared source file
// into the extracted tree before configure, for a file a release tarball omits because
// upstream generates it from a .git checkout (e.g. libvmaf's vcs_version.h). Each file's
// parent directory is created first; lines are written verbatim via single-quoted printf
// arguments (validated to contain no single quote or newline). Returns no lines when the
// recipe declares none.
func generatedSourceFileScriptLines(spec LibraryBuildSpec) []string {
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

// pkgConfigLibsLineOverrideLines replaces the installed .pc's whole "Libs:" line value with
// the recipe's PkgConfigLibsLine. A static archive and a shared import library can share a
// base name (e.g. LibreSSL's static libcrypto.a next to a leftover OpenSSL libcrypto.dll.a in
// the same prefix); GNU ld prefers the .dll.a for a bare -lcrypto, so the link picks the wrong
// crypto and fails with undefined references. Forcing -l:lib<name>.a (or an absolute archive
// path) in the Libs line makes the link select this recipe's own static archives. The pc name
// is a validated path segment and the libs line is validated against safePkgConfigLibsLinePattern,
// so both are safe to interpolate. Returns no lines when the recipe declares no override.
// installPrefixSetupLine emits the shell that defines ${install_prefix}. For a normal
// recipe it is the shared MSYS2 profile prefix. For a PrivatePrefixInstall recipe it is a
// per-library directory under the profile prefix, wiped first so a re-run starts clean and
// the library's archives never share a directory with a same-named package's.
func installPrefixSetupLine(spec LibraryBuildSpec) string {
	if spec.PrivatePrefixInstall {
		privatePrefix := `${profile_prefix}/` + PrivateLibraryInstallSubdir + `/` + spec.PkgConfigName
		return `install_prefix="` + privatePrefix + `"; rm -rf "${install_prefix}"; mkdir -p "${install_prefix}"`
	}
	return `install_prefix="${profile_prefix}"`
}

// stripPkgConfigRequiresLines removes the Requires and Requires.private lines from a
// privately-installed library's .pc. The private install lists every needed archive by
// absolute path in its Libs line, so leaving Requires.private would only let pkg-config
// re-add bare -l<name> flags that resolve back to a same-named archive in the shared
// prefix ??exactly the collision the private install exists to avoid. No-op otherwise.
func stripPkgConfigRequiresLines(spec LibraryBuildSpec) []string {
	if !spec.PrivatePrefixInstall || spec.PkgConfigName == "" {
		return nil
	}
	return []string{
		"echo " + shellQuote("Stripping Requires/Requires.private from "+spec.PkgConfigName+".pc so its absolute-path Libs line is the sole source of link libraries."),
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`sed -i -E '/^Requires(\.private)?:/d' "${pc_file}"`,
	}
}

func pkgConfigLibsLineOverrideLines(spec LibraryBuildSpec) []string {
	if spec.PkgConfigName == "" || spec.PkgConfigLibsLine == "" {
		return nil
	}
	// '#' delimiter so the path slashes in the replacement need no escaping; the validated
	// libs line cannot contain '#'. Single-quoted via shellQuote, so ${libdir} stays literal
	// in the .pc for pkg-config to expand.
	sedProgram := "s#^Libs:.*#Libs: " + spec.PkgConfigLibsLine + "#"
	return []string{
		"echo " + shellQuote("Overriding "+spec.PkgConfigName+".pc Libs line to force this library's own static archives ahead of any same-named shared import library in the prefix."),
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`sed -i -E ` + shellQuote(sedProgram) + ` "${pc_file}"`,
		`echo "Patched pkg-config Libs: $(grep '^Libs:' "${pc_file}")"`,
	}
}

func pkgConfigAppendLibsLines(spec LibraryBuildSpec) []string {
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
		"echo " + shellQuote("Patching "+spec.PkgConfigName+".pc: moving runtime libraries ("+strings.TrimSpace(appendFlags)+") after the static archives for correct GNU static link order."),
		// CMake projects install the .pc under either lib/pkgconfig (LIBDIR) or
		// share/pkgconfig (DATAROOTDIR, e.g. mpeghdec); pkg-config searches both, so accept
		// whichever the project used instead of assuming lib/pkgconfig.
		`pc_file="${install_prefix}/lib/pkgconfig/` + spec.PkgConfigName + `.pc"`,
		`if [ ! -f "${pc_file}" ]; then pc_file="${install_prefix}/share/pkgconfig/` + spec.PkgConfigName + `.pc"; fi`,
		`if [ ! -f "${pc_file}" ]; then echo "ERROR: ` + spec.PkgConfigName + `.pc not found in lib/pkgconfig or share/pkgconfig under ${install_prefix}."; exit 1; fi`,
		`sed -i -E ` + shellQuote(sedProgram) + ` "${pc_file}"`,
		`echo "Patched pkg-config Libs: $(grep '^Libs:' "${pc_file}")"`,
	}
}

// internalConfigureMakeBuildScriptLines builds a library with a ./configure script
// (standard autotools or an x264/davs2-style custom configure in a subdirectory)
// followed by make and make install, inside the MSYS2 environment. Unlike the native
// cmake path, the whole build runs under MSYS2, so the install prefix is the unix
// /ucrt64 path (the standard MSYS2 layout); the generated .pc then matches every other
// MSYS2 package and pkg-config/gcc resolve it during FFmpeg configure.
func internalConfigureMakeBuildScriptLines(spec LibraryBuildSpec) ([]string, error) {
	if err := validateLibraryBuildSpec(spec, false); err != nil {
		return nil, err
	}
	configureOptionArray := make([]string, 0, len(spec.ConfigureOptions))
	for _, configureOption := range spec.ConfigureOptions {
		configureOptionArray = append(configureOptionArray, shellQuote(configureOption))
	}
	buildTargetArray := make([]string, 0, len(spec.MakeBuildTargets))
	for _, buildTarget := range spec.MakeBuildTargets {
		buildTargetArray = append(buildTargetArray, shellQuote(buildTarget))
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
		"echo " + shellQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`for required_tool in make gcc; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
	}
	// Write recipe-generated source files, then apply source patches, before configure. No-op
	// when none.
	scriptLines = append(scriptLines, generatedSourceFileScriptLines(spec)...)
	scriptLines = append(scriptLines, sourcePatchScriptLines(spec)...)
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
		makeVariableSuffix += " " + shellQuote(makeVariable)
	}
	buildLine := `make -j"$(nproc)"` + makeVariableSuffix
	if len(buildTargetArray) > 0 {
		buildLine += " " + strings.Join(buildTargetArray, " ")
	}
	scriptLines = append(scriptLines, buildLine+" 2>&1")
	for _, installTarget := range installTargets {
		scriptLines = append(scriptLines, `make `+shellQuote(installTarget)+makeVariableSuffix+` 2>&1`)
	}
	scriptLines = append(scriptLines, libraryInstallVerificationLines(spec)...)
	scriptLines = append(scriptLines, "echo "+shellQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

// internalMakeBuildScriptLines builds a library that ships only a plain Makefile (no
// configure, no cmake) inside the MSYS2 environment. Such projects (e.g. quirc) usually
// have no lib-only install target ??their `make install` also builds demos that pull
// extra deps ??so the generator builds just the requested targets (typically the static
// archive) and installs by copying the recipe's declared header files and static archive
// into the selected profile prefix, then verifies them. Headers are copied flat into
// include/ by basename, so a recipe's VerifyHeaderRelativePath must match that basename.
func internalMakeBuildScriptLines(spec LibraryBuildSpec) ([]string, error) {
	if err := validateLibraryBuildSpec(spec, false); err != nil {
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
		buildTargetArray = append(buildTargetArray, shellQuote(buildTarget))
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		// Only the CMake path supports PrivatePrefixInstall; here install_prefix is always the
		// shared prefix so the pkg-config patch and verification helpers resolve correctly.
		`install_prefix="${profile_prefix}"`,
		"echo " + shellQuote("Preparing internal-track library from source: "+spec.DisplayName),
		`src_dir="$(pwd)"`,
		`echo "Source directory: ${src_dir}"`,
		`for required_tool in make gcc; do`,
		`  if ! command -v "${required_tool}" >/dev/null 2>&1; then echo "ERROR: ${required_tool} is required to build this internal-track library."; exit 1; fi`,
		`done`,
	}
	// Write recipe-generated source files, then apply source patches, before the build. No-op
	// when none.
	scriptLines = append(scriptLines, generatedSourceFileScriptLines(spec)...)
	scriptLines = append(scriptLines, sourcePatchScriptLines(spec)...)
	buildLine := `make -j"$(nproc)"`
	// Command-line variable assignments (e.g. SDL_CFLAGS=) come before the targets so they
	// override and skip the makefile's own $(shell pkg-config ...) assignments.
	for _, makeVariable := range spec.MakeVariables {
		buildLine += " " + shellQuote(makeVariable)
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
	scriptLines = append(scriptLines, libraryInstallVerificationLines(spec)...)
	scriptLines = append(scriptLines, "echo "+shellQuote("Internal-track library prepared: "+spec.DisplayName))
	return scriptLines, nil
}

// ExternalLibraryImportScriptLines imports a verified, already-extracted vendor
// binary archive (the script working directory) into the selected MSYS2 profile
// prefix: headers and link libraries into include/ and lib/, runtime DLLs into bin/,
// generating a MinGW import library from a bundled DLL when needed so that the
// FFmpeg link step (-l<stem>) resolves. It then verifies the imported header.
func ExternalLibraryImportScriptLines(spec LibraryBuildSpec) ([]string, error) {
	if err := validateLibraryBuildSpec(spec, true); err != nil {
		return nil, err
	}
	scriptLines := []string{
		"#!/usr/bin/env bash",
		"set -euo pipefail",
		`profile_prefix="${MSYSTEM_PREFIX:-/ucrt64}"`,
		"echo " + shellQuote("Importing external-track library: "+spec.DisplayName),
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
	scriptLines = append(scriptLines, libraryInstallVerificationLines(spec)...)
	scriptLines = append(scriptLines, "echo "+shellQuote("External-track library imported: "+spec.DisplayName))
	return scriptLines, nil
}

// libraryInstallVerificationLines confirms the installed header (always) and link
// library (when a stem is given) actually landed in the prefix, so a silent no-op
// install or import is turned into a hard failure instead of a later configure error.
func libraryInstallVerificationLines(spec LibraryBuildSpec) []string {
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
