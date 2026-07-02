package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLibjxlPrepare provides a pre-0.9 libjxl API for FFmpeg 5.1.9 without
// patching the FFmpeg source.
func LLibraryLibjxlPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "libjxl"
	plan.VersionSpecificGoFile = "versions/5.1.9/libjxl.go"
	plan.UseInternalSourceBuild("JPEG XL", "cmake")
	plan.RequireBuildPackages("brotli", "cmake", "highway", "lcms2", "ninja", "pkgconf")
	plan.AddCMakeOptions(
		"-DBUILD_SHARED_LIBS=OFF",
		"-DBUILD_TESTING=OFF",
		"-DJPEGXL_ENABLE_BENCHMARK=OFF",
		"-DJPEGXL_ENABLE_DEVTOOLS=OFF",
		"-DJPEGXL_ENABLE_DOXYGEN=OFF",
		"-DJPEGXL_ENABLE_EXAMPLES=OFF",
		"-DJPEGXL_ENABLE_FUZZERS=OFF",
		"-DJPEGXL_ENABLE_JNI=OFF",
		"-DJPEGXL_ENABLE_JPEGLI=OFF",
		"-DJPEGXL_ENABLE_JPEGLI_LIBJPEG=OFF",
		"-DJPEGXL_ENABLE_MANPAGES=OFF",
		"-DJPEGXL_ENABLE_OPENEXR=OFF",
		"-DJPEGXL_ENABLE_PLUGINS=OFF",
		"-DJPEGXL_ENABLE_SJPEG=OFF",
		"-DJPEGXL_ENABLE_SKCMS=OFF",
		"-DJPEGXL_ENABLE_TOOLS=OFF",
		"-DJPEGXL_FORCE_SYSTEM_BROTLI=ON",
		"-DJPEGXL_FORCE_SYSTEM_HWY=ON",
		"-DJPEGXL_FORCE_SYSTEM_LCMS2=ON",
	)
	plan.UsePkgConfig("libjxl")
	plan.UsePrivatePrefixInstall()
	plan.AppendPkgConfigCFlags("-DJXL_STATIC_DEFINE")
	plan.OverridePkgConfigLibsLine("-L${libdir} -ljxl -lm -lstdc++ -lhwy -lbrotlienc -lbrotlidec -lbrotlicommon -llcms2 -pthread")
	plan.OverridePkgConfigModuleLibsLine("libjxl_threads", "-L${libdir} -ljxl_threads -lstdc++ -pthread")
	plan.Verify("jxl/decode.h", "jxl")
}
