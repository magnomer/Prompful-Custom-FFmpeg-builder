package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 5.1.9.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/5.1.9/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCmakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCmakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
