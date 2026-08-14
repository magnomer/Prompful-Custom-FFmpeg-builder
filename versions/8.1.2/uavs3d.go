package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavsPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 8.1.2.
func LUavsPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/8.1.2/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
