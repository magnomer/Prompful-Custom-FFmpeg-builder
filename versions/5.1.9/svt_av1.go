package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibrarySvtav1Prepare performs the coded preparation manipulation for svt-av1 on FFmpeg 5.1.9.
func LLibrarySvtav1Prepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "svt-av1"
	plan.VersionSpecificGoFile = "versions/5.1.9/svt_av1.go"
	plan.UseInternalSourceBuild("SVT-AV1", "cmake")
	plan.RequireBuildPackages("nasm")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DBUILD_APPS=OFF", "-DBUILD_TESTING=OFF")
	plan.UsePkgConfig("SvtAv1Enc")
	plan.Verify("svt-av1/EbSvtAv1Enc.h", "SvtAv1Enc")
}
