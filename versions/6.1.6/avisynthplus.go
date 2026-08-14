package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 6.1.6.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/6.1.6/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCmakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCmakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
