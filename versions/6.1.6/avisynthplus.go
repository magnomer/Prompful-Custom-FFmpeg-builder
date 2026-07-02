package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 6.1.6.
func LLibraryAvisynthplusPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/6.1.6/avisynthplus.go"
	plan.UseInternalSourceBuild("AviSynth+", "cmake")
	plan.AddCMakeOptions("-DHEADERS_ONLY=ON")
	plan.BuildCMakeTargets("VersionGen")
	plan.Verify("avisynth/avisynth_c.h", "")
}
