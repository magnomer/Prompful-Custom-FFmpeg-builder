package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVvencPrepare performs the coded preparation manipulation for vvenc on FFmpeg 9.0.1.
func LLibraryVvencPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "vvenc"
	plan.VersionSpecificGoFile = "versions/9.0.1/vvenc.go"
	plan.LSourceCompilationUse("vvenc", "cmake")
	plan.LCMakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-DVVENC_LIBRARY_ONLY=ON")
	plan.LPackageConfigurationUse("libvvenc")
	plan.LLibraryLineAppend("stdc++")
	plan.LCommandVerify("vvenc/vvenc.h", "vvenc")
}
