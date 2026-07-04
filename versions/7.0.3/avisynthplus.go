package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 7.0.3.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/7.0.3/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCMakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCMakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
