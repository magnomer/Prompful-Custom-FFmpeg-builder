package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVvencPrepare performs the coded preparation manipulation for vvenc on FFmpeg 7.1.5.
func LLibraryVvencPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "vvenc"
	plan.VersionSpecificGoFile = "versions/7.1.5/vvenc.go"
	plan.UseInternalSourceBuild("vvenc", "cmake")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DVVENC_LIBRARY_ONLY=ON")
	plan.UsePkgConfig("libvvenc")
	plan.AppendPkgConfigLibs("stdc++")
	plan.Verify("vvenc/vvenc.h", "vvenc")
}
