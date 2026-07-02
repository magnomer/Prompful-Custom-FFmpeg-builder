package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVvencPrepare performs the coded preparation manipulation for vvenc on FFmpeg 8.0.3.
func LLibraryVvencPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "vvenc"
	plan.VersionSpecificGoFile = "versions/8.0.3/vvenc.go"
	plan.UseInternalSourceBuild("vvenc", "cmake")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DVVENC_LIBRARY_ONLY=ON")
	plan.UsePkgConfig("libvvenc")
	plan.AppendPkgConfigLibs("stdc++")
	plan.Verify("vvenc/vvenc.h", "vvenc")
}
