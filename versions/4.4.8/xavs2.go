package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LXavs2Prepare performs the coded preparation manipulation for xavs2 on FFmpeg 4.4.8.
func LXavs2Prepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "xavs2"
	plan.VersionSpecificGoFile = "versions/4.4.8/xavs2.go"
	plan.LSourceCompilationUse("xavs2", "configure-make")
	plan.LConfigureSubdirectoryUse("build/linux")
	plan.LConfigureOptionAdd("--disable-cli", "--enable-pic", "--extra-cflags=-Wno-error=incompatible-pointer-types")
	plan.LMakeVariableAdd("STRIP=")
	plan.LMakeTargetAdd("lib-static")
	plan.LMakeTargetInstall("install-lib-static")
	plan.LPackageConfigurationUse("xavs2")
	plan.LCommandVerify("xavs2.h", "xavs2")
}
