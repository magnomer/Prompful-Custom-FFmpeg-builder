package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryDavs2Prepare performs the coded preparation manipulation for davs2 on FFmpeg 8.0.3.
func LLibraryDavs2Prepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "davs2"
	plan.VersionSpecificGoFile = "versions/8.0.3/davs2.go"
	plan.UseInternalSourceBuild("libdavs2", "configure-make")
	plan.ConfigureInSubdir("build/linux")
	plan.AddConfigureOptions("--disable-cli", "--enable-pic")
	plan.BuildMakeTargets("lib-static")
	plan.InstallMakeTargets("install-lib-static")
	plan.UsePkgConfig("davs2")
	plan.Verify("davs2.h", "davs2")
}
