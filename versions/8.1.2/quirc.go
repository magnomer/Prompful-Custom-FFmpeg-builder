package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryQuircPrepare performs the coded preparation manipulation for quirc on FFmpeg 8.1.2.
func LLibraryQuircPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "quirc"
	plan.VersionSpecificGoFile = "versions/8.1.2/quirc.go"
	plan.UseInternalSourceBuild("libquirc", "make")
	plan.BuildMakeTargets("libquirc.a")
	plan.AddMakeVariables("SDL_CFLAGS=", "SDL_LIBS=")
	plan.InstallMakeHeaders("lib/quirc.h")
	plan.InstallMakeStaticLibrary("libquirc.a")
	plan.Verify("quirc.h", "quirc")
}
