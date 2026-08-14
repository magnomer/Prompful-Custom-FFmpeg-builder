package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavsPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 4.4.8.
func LUavsPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/4.4.8/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
