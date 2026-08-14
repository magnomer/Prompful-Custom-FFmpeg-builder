package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavsPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 7.1.5.
func LUavsPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/7.1.5/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
