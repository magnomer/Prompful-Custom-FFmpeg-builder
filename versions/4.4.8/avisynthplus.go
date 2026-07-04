package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryAvisynthplusPrepare performs the coded preparation manipulation for avisynthplus on FFmpeg 4.4.8.
func LLibraryAvisynthplusPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "avisynthplus"
	plan.VersionSpecificGoFile = "versions/4.4.8/avisynthplus.go"
	plan.LSourceCompilationUse("AviSynth+", "cmake")
	plan.LCMakeOptionAdd("-DHEADERS_ONLY=ON")
	plan.LCMakeTargetAdd("VersionGen")
	plan.LCommandVerify("avisynth/avisynth_c.h", "")
}
