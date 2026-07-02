package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibrarySvtav1Prepare performs the coded preparation manipulation for svt-av1 on FFmpeg 7.0.3.
func LLibrarySvtav1Prepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "svt-av1"
	plan.VersionSpecificGoFile = "versions/7.0.3/svt_av1.go"
	plan.UseInternalSourceBuild("SVT-AV1", "cmake")
	plan.RequireBuildPackages("nasm")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DBUILD_APPS=OFF", "-DBUILD_TESTING=OFF")
	plan.UsePkgConfig("SvtAv1Enc")
	plan.Verify("svt-av1/EbSvtAv1Enc.h", "SvtAv1Enc")
}
