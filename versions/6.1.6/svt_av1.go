package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LSvtav1Prepare performs the coded preparation manipulation for svt-av1 on FFmpeg 6.1.6.
func LSvtav1Prepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "svt-av1"
	plan.VersionSpecificGoFile = "versions/6.1.6/svt_av1.go"
	plan.LSourceCompilationUse("SVT-AV1", "cmake")
	plan.LBuildPackageRequire("nasm")
	plan.LCMakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-DBUILD_APPS=OFF", "-DBUILD_TESTING=OFF")
	plan.LPackageConfigurationUse("SvtAv1Enc")
	plan.LCommandVerify("svt-av1/EbSvtAv1Enc.h", "SvtAv1Enc")
}
