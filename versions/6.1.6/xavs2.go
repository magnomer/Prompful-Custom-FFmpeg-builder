package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryXavs2Prepare performs the coded preparation manipulation for xavs2 on FFmpeg 6.1.6.
func LLibraryXavs2Prepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "xavs2"
	plan.VersionSpecificGoFile = "versions/6.1.6/xavs2.go"
	plan.UseInternalSourceBuild("xavs2", "configure-make")
	plan.ConfigureInSubdir("build/linux")
	plan.AddConfigureOptions("--disable-cli", "--enable-pic", "--extra-cflags=-Wno-error=incompatible-pointer-types")
	plan.AddMakeVariables("STRIP=")
	plan.BuildMakeTargets("lib-static")
	plan.InstallMakeTargets("install-lib-static")
	plan.UsePkgConfig("xavs2")
	plan.Verify("xavs2.h", "xavs2")
}
