package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 9.0.1.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/9.0.1/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCmakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCmakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
