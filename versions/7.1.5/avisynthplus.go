package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 7.1.5.
func LLibraryAvisynthplusPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/7.1.5/avisynthplus.go"
	plan.UseInternalSourceBuild("AviSynth+", "cmake")
	plan.AddCMakeOptions("-DHEADERS_ONLY=ON")
	plan.BuildCMakeTargets("VersionGen")
	plan.Verify("avisynth/avisynth_c.h", "")
}
