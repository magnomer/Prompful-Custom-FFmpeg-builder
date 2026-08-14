package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryMpeghdecPrepare performs the coded preparation manipulation for mpeghdec on FFmpeg 9.0.1.
func LLibraryMpeghdecPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "mpeghdec"
	plan.VersionSpecificGoFile = "versions/9.0.1/mpeghdec.go"
	plan.LSourceCompilationUse("libmpeghdec", "cmake")
	plan.LCMakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-Dmpeghdec_BUILD_BINARIES=OFF")
	plan.LPackageConfigurationUse("mpeghdec")
	plan.LLibraryLineAppend("stdc++")
	plan.LPreparationModificationAdd("src/libFDK/include/common_fix.h", "#if !defined(_MSC_VER) && defined(__x86_64__)", "#if 0 /* mpeghdec recipe patch: SHORT fMin/fMax duplicate FIXP_SGL(=short) on LLP64 mingw x86_64 */")
	plan.LPreparationModificationAdd("mpeghdec.pc.in", `Cflags: -I"${includedir}"`, `Cflags: -I"${includedir}" -DMPEGHDEC_STATIC=1`)
	plan.LCommandVerify("mpeghdec/mpeghdecoder.h", "mpeghdec")
}
