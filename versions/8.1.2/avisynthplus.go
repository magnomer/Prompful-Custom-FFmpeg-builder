package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 8.1.2.
func LLibraryAvisynthplusPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/8.1.2/avisynthplus.go"
	plan.UseInternalSourceBuild("AviSynth+", "cmake")
	plan.AddCMakeOptions("-DHEADERS_ONLY=ON")
	plan.BuildCMakeTargets("VersionGen")
	plan.Verify("avisynth/avisynth_c.h", "")
}
