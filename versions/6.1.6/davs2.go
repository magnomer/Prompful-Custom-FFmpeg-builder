package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LDavs2Prepare performs the coded preparation manipulation for davs2 on FFmpeg 6.1.6.
func LDavs2Prepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "davs2"
	plan.VersionSpecificGoFile = "versions/6.1.6/davs2.go"
	plan.LSourceCompilationUse("libdavs2", "configure-make")
	plan.LConfigureSubdirectoryUse("build/linux")
	plan.LConfigureOptionAdd("--disable-cli", "--enable-pic")
	plan.LMakeTargetAdd("lib-static")
	plan.LMakeTargetInstall("install-lib-static")
	plan.LPackageConfigurationUse("davs2")
	plan.LCommandVerify("davs2.h", "davs2")
}
