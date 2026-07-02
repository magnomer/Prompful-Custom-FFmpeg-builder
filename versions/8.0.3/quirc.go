package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryQuircPrepare performs the coded preparation manipulation for quirc on FFmpeg 8.0.3.
func LLibraryQuircPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "quirc"
	plan.VersionSpecificGoFile = "versions/8.0.3/quirc.go"
	plan.UseInternalSourceBuild("libquirc", "make")
	plan.BuildMakeTargets("libquirc.a")
	plan.AddMakeVariables("SDL_CFLAGS=", "SDL_LIBS=")
	plan.InstallMakeHeaders("lib/quirc.h")
	plan.InstallMakeStaticLibrary("libquirc.a")
	plan.Verify("quirc.h", "quirc")
}
