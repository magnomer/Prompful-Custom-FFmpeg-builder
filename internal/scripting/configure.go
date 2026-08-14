package scripting

import (
	"fmt"
	"strings"

	"promptfulcustomffmpegbuilder/internal/catalogfacts"
)

// LConfigureScriptCreate builds the FFmpeg configure script. privatePkgConfigDirs are the
// pkgconfig directories of any privately-installed libraries (e.g. libtls); they are
// prepended to the script's exported PKG_CONFIG_PATH/PKG_CONFIG_LIBDIR so pkg-config finds
// those isolated modules (and their private libssl/libcrypto) ahead of the shared prefix.
// ffmpegVersion is the release being built (e.g. "8.1.2"); the pre-configure pkg-config
// version floors are resolved from that release's support release-support manifest, so an older FFmpeg gets
// its own (usually lower) floors rather than a single hardcoded set. A snapshot/unknown
// version resolves no floors and the build relies on FFmpeg configure as the backstop.
func LConfigureScriptCreate(configureFlags []string, privatePkgConfigDirs []string, ffmpegVersion string) ([]string, error) {
	quotedConfigureFlags := make([]string, 0, len(configureFlags))
	for _, configureFlag := range configureFlags {
		if err := LFlagConfigureValidate(configureFlag); err != nil {
			return nil, err
		}
		quotedConfigureFlags = append(quotedConfigureFlags, LShellTextQuote(configureFlag))
	}
	quotedConfigureFlagArray := strings.Join(quotedConfigureFlags, " ")

	LModulePkgconfigs := LPackageModuleList(configureFlags, ffmpegVersion)

	// Prepend any private pkgconfig dirs to the shared prefix search path. Each is a unix
	// path under the profile prefix produced by LPrivateDirectoryGet.
	privatePkgConfigPrefix := ""
	for _, privateDir := range privatePkgConfigDirs {
		if !LPackageDirectoryPattern.MatchString(privateDir) {
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
	if LFlagConfigureCheck(configureFlags, "--enable-libsvtjpegxs") {
		scriptLines = append(scriptLines, "if ! try_enable_svt_jpeg_xs; then svtjpegxs_disabled=1; remove_configure_flag --enable-libsvtjpegxs; fi")
	}
	if LFlagConfigureCheck(configureFlags, "--enable-liblensfun") {
		scriptLines = append(scriptLines, "if ! try_enable_lensfun; then lensfun_disabled=1; remove_configure_flag --enable-liblensfun; fi")
	}
	if LFlagConfigureCheck(configureFlags, "--enable-vapoursynth") {
		scriptLines = append(scriptLines, "if ! try_enable_vapoursynth; then vapoursynth_disabled=1; remove_configure_flag --enable-vapoursynth; fi")
	}
	if LFlagConfigureCheck(configureFlags, "--enable-libopencv") {
		scriptLines = append(scriptLines, LOpencvScriptCreate()...)
	}
	if LFlagConfigureCheck(configureFlags, "--enable-libonnxruntime") {
		scriptLines = append(scriptLines, LOnnxRuntimeScriptCreate()...)
	}
	// The --enable-amf flag is version-agnostic, but only FFmpeg 9+ raises the AMF header floor to
	// 1.5.2.0. Scope the header preflight to FFmpeg 9+ so 8.x builds keep their lower floor. An
	// unknown/snapshot version resolves no line and is left to FFmpeg configure.
	if LFlagConfigureCheck(configureFlags, "--enable-amf") {
		if ffmpegMajor, _, ok := catalogfacts.LReleaseLineSplit(catalogfacts.LReleaseKeyGet(ffmpegVersion)); ok && ffmpegMajor >= 9 {
			scriptLines = append(scriptLines, LAmfScriptCreate()...)
		}
	}
	// The --enable-ffnvcodec flag is version-agnostic, but only FFmpeg 9+ raises the NVENC SDK floor
	// to 11.1 (NVENCAPI 11.1). Scope the header preflight to FFmpeg 9+ so 8.x builds keep their lower
	// floor. An unknown/snapshot version resolves no line and is left to FFmpeg configure.
	if LFlagConfigureCheck(configureFlags, "--enable-ffnvcodec") {
		if ffmpegMajor, _, ok := catalogfacts.LReleaseLineSplit(catalogfacts.LReleaseKeyGet(ffmpegVersion)); ok && ffmpegMajor >= 9 {
			scriptLines = append(scriptLines, LNvencScriptCreate()...)
		}
	}
	if len(LModulePkgconfigs) > 0 {
		scriptLines = append(scriptLines, "echo 'Diagnosing selected pkg-config libraries before FFmpeg configure.'")
		for _, mod := range LModulePkgconfigs {
			diagnosticLine := fmt.Sprintf("diagnose_pkg_config_module %s %s", LShellTextQuote(mod.Name), LShellTextQuote(mod.MinVersion))
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

func LFlagConfigureCheck(configureFlags []string, wantedFlag string) bool {
	for _, configureFlag := range configureFlags {
		if configureFlag == wantedFlag {
			return true
		}
	}
	return false
}
