package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 8.1.2.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/8.1.2/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCmakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCmakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
