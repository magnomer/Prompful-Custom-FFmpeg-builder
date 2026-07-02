package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 8.0.3.
func LLibraryAvisynthplusPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/8.0.3/avisynthplus.go"
	plan.UseInternalSourceBuild("AviSynth+", "cmake")
	plan.AddCMakeOptions("-DHEADERS_ONLY=ON")
	plan.BuildCMakeTargets("VersionGen")
	plan.Verify("avisynth/avisynth_c.h", "")
}
