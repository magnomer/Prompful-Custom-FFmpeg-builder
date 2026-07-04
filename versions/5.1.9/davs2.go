package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LDavs2Prepare performs the coded preparation manipulation for davs2 on FFmpeg 5.1.9.
func LDavs2Prepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "davs2"
	plan.VersionSpecificGoFile = "versions/5.1.9/davs2.go"
	plan.LSourceCompilationUse("libdavs2", "configure-make")
	plan.LConfigureSubdirectoryUse("build/linux")
	plan.LConfigureOptionAdd("--disable-cli", "--enable-pic")
	plan.LMakeTargetAdd("lib-static")
	plan.LMakeTargetInstall("install-lib-static")
	plan.LPackageConfigurationUse("davs2")
	plan.LCommandVerify("davs2.h", "davs2")
}
