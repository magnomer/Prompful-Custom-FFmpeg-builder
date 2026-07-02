package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryMpeghdecPrepare performs the coded preparation manipulation for mpeghdec on FFmpeg 8.1.2.
func LLibraryMpeghdecPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "mpeghdec"
	plan.VersionSpecificGoFile = "versions/8.1.2/mpeghdec.go"
	plan.UseInternalSourceBuild("libmpeghdec", "cmake")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-Dmpeghdec_BUILD_BINARIES=OFF")
	plan.UsePkgConfig("mpeghdec")
	plan.AppendPkgConfigLibs("stdc++")
	plan.PatchSource("src/libFDK/include/common_fix.h", "#if !defined(_MSC_VER) && defined(__x86_64__)", "#if 0 /* mpeghdec recipe patch: SHORT fMin/fMax duplicate FIXP_SGL(=short) on LLP64 mingw x86_64 */")
	plan.PatchSource("mpeghdec.pc.in", `Cflags: -I"${includedir}"`, `Cflags: -I"${includedir}" -DMPEGHDEC_STATIC=1`)
	plan.Verify("mpeghdec/mpeghdecoder.h", "mpeghdec")
}
