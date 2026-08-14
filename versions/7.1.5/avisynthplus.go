package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 7.1.5.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/7.1.5/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCmakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCmakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
