package scripting

// LOpenCVScriptCreate makes FFmpeg's libopencv detection work against MSYS2's
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
func LOpenCVScriptCreate() []string {
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

// LOnnxRuntimeScriptCreate exposes MSYS2's nested ONNX Runtime header to FFmpeg
// configure. MSYS2 installs onnxruntime_c_api.h under include/onnxruntime/, but
// FFmpeg 9 probes the bare header name, so the nested dir must join extra-cflags.
func LOnnxRuntimeScriptCreate() []string {
	return []string{
		`onnx_incdir="${MSYSTEM_PREFIX:-/ucrt64}/include/onnxruntime"`,
		`if [ -d "${onnx_incdir}" ]; then`,
		`  echo "Adding ${onnx_incdir} to FFmpeg extra-cflags so the ONNX Runtime onnxruntime_c_api.h header check passes (MSYS2 installs it under include/onnxruntime, off the default search path)."`,
		`  configure_flags+=("--extra-cflags=-I${onnx_incdir}")`,
		`fi`,
	}
}

// LAmfScriptCreate guards --enable-amf against a stale MSYS2 amf-headers package on FFmpeg 9.
// AMF ships headers only (no pkg-config module), so the version floor cannot flow through the
// pkg-config preflight; it must inspect AMF/core/Version.h directly. FFmpeg 9's configure aborts
// deep in its own AMF probe with an opaque check_cpp_condition failure when the header is too old.
// The floor mirrors FFmpeg 9.0 configure exactly: it requires
//
//	(AMF_VERSION_MAJOR << 48 | AMF_VERSION_MINOR << 32 | AMF_VERSION_RELEASE << 16 | AMF_VERSION_BUILD_NUM)
//	>= 0x0001000500020000, i.e. AMF >= 1.5.2.0 (FFmpeg 8.1 required only 1.4.36.0). Caller gates this
//
// to FFmpeg 9+, so 8.x builds keep their lower floor and are unaffected. A missing or unparseable
// header is left to FFmpeg configure rather than blocked here.
func LAmfScriptCreate() []string {
	return []string{
		`amf_version_header="${profile_prefix}/include/AMF/core/Version.h"`,
		`if [ -f "${amf_version_header}" ]; then`,
		`  amf_field_read() { grep -E "^#define[[:space:]]+$1([[:space:]]|\$)" "${amf_version_header}" | grep -oE "[0-9]+" | tail -n1 || true; }`,
		`  amf_major="$(amf_field_read AMF_VERSION_MAJOR)"`,
		`  amf_minor="$(amf_field_read AMF_VERSION_MINOR)"`,
		`  amf_release="$(amf_field_read AMF_VERSION_RELEASE)"`,
		`  amf_build="$(amf_field_read AMF_VERSION_BUILD_NUM)"`,
		`  if [ -n "${amf_major}" ] && [ -n "${amf_minor}" ] && [ -n "${amf_release}" ] && [ -n "${amf_build}" ]; then`,
		`    amf_installed=$(( (amf_major << 48) | (amf_minor << 32) | (amf_release << 16) | amf_build ))`,
		`    amf_required=$(( (1 << 48) | (5 << 32) | (2 << 16) | 0 ))`,
		`    echo "AMD AMF header version: ${amf_major}.${amf_minor}.${amf_release}.${amf_build} (FFmpeg 9 requires >= 1.5.2.0)"`,
		`    if [ "${amf_installed}" -lt "${amf_required}" ]; then`,
		`      echo "ERROR: Installed AMD AMF headers are ${amf_major}.${amf_minor}.${amf_release}.${amf_build}, but FFmpeg 9 requires AMF >= 1.5.2.0 for --enable-amf."`,
		`      echo "ERROR: Update the amf-headers package for this shell profile (e.g. pacman -S --needed <profile>-amf-headers) and rebuild."`,
		`      exit 1`,
		`    fi`,
		`    echo "AMD AMF header version satisfies the FFmpeg 9 minimum. Keeping --enable-amf."`,
		`  else`,
		`    echo "WARNING: Could not parse the AMD AMF header version from ${amf_version_header}. Leaving the AMF version check to FFmpeg configure."`,
		`  fi`,
		`fi`,
	}
}

// LNvencScriptCreate guards --enable-ffnvcodec against a stale MSYS2 ffnvcodec-headers package on
// FFmpeg 9. FFmpeg 9 drops NVENC SDK generations older than 11.1 and gates them at compile time
// through nvEncodeAPI.h's NVENCAPI version macros (the same NVENCAPI_CHECK_VERSION scheme FFmpeg's
// own nvenc code uses), not through a pkg-config version requirement, so a too-old header slips past
// configure's existence-only ffnvcodec probe and aborts deep in the NVENC compile with an opaque
// error. This preflight inspects NVENCAPI_MAJOR_VERSION/NVENCAPI_MINOR_VERSION directly and requires
// >= 11.1 (NVENC SDK 11.1, ffnvcodec 11.1.5.0). Caller gates this to FFmpeg 9+, so 8.x builds keep
// their lower floor. A missing or unparseable header is left to FFmpeg configure rather than blocked
// here.
func LNvencScriptCreate() []string {
	return []string{
		`nvenc_version_header="${profile_prefix}/include/ffnvcodec/nvEncodeAPI.h"`,
		`if [ -f "${nvenc_version_header}" ]; then`,
		`  nvenc_field_read() { grep -E "^#define[[:space:]]+$1([[:space:]]|\$)" "${nvenc_version_header}" | grep -oE "[0-9]+" | tail -n1 || true; }`,
		`  nvenc_major="$(nvenc_field_read NVENCAPI_MAJOR_VERSION)"`,
		`  nvenc_minor="$(nvenc_field_read NVENCAPI_MINOR_VERSION)"`,
		`  if [ -n "${nvenc_major}" ] && [ -n "${nvenc_minor}" ]; then`,
		`    nvenc_installed=$(( (nvenc_major * 1000) + nvenc_minor ))`,
		`    nvenc_required=$(( (11 * 1000) + 1 ))`,
		`    echo "NVIDIA NVENC (ffnvcodec) header NVENCAPI version: ${nvenc_major}.${nvenc_minor} (FFmpeg 9 requires >= 11.1)"`,
		`    if [ "${nvenc_installed}" -lt "${nvenc_required}" ]; then`,
		`      echo "ERROR: Installed ffnvcodec headers are NVENCAPI ${nvenc_major}.${nvenc_minor}, but FFmpeg 9 requires NVENC SDK >= 11.1 (ffnvcodec 11.1.5.0) for --enable-ffnvcodec."`,
		`      echo "ERROR: Update the ffnvcodec-headers package for this shell profile (e.g. pacman -S --needed <profile>-ffnvcodec-headers) and rebuild."`,
		`      exit 1`,
		`    fi`,
		`    echo "ffnvcodec header version satisfies the FFmpeg 9 minimum. Keeping --enable-ffnvcodec."`,
		`  else`,
		`    echo "WARNING: Could not parse the NVENCAPI version from ${nvenc_version_header}. Leaving the ffnvcodec version check to FFmpeg configure."`,
		`  fi`,
		`fi`,
	}
}
