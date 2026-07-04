package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryQuircPrepare performs the coded preparation manipulation for quirc on FFmpeg 7.1.5.
func LLibraryQuircPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "quirc"
	plan.VersionSpecificGoFile = "versions/7.1.5/quirc.go"
	plan.LSourceCompilationUse("libquirc", "make")
	plan.LMakeTargetAdd("libquirc.a")
	plan.LMakeVariableAdd("SDL_CFLAGS=", "SDL_LIBS=")
	plan.LHeaderInstall("lib/quirc.h")
	plan.LLibraryStaticInstall("libquirc.a")
	plan.LCommandVerify("quirc.h", "quirc")
}
