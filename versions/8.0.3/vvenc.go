package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVvencPrepare performs the coded preparation manipulation for vvenc on FFmpeg 8.0.3.
func LLibraryVvencPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "vvenc"
	plan.VersionSpecificGoFile = "versions/8.0.3/vvenc.go"
	plan.LSourceCompilationUse("vvenc", "cmake")
	plan.LCmakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-DVVENC_LIBRARY_ONLY=ON")
	plan.LPackageConfigurationUse("libvvenc")
	plan.LLibraryLineAppend("stdc++")
	plan.LCommandVerify("vvenc/vvenc.h", "vvenc")
}
