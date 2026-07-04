package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavs3dPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 4.4.8.
func LUavs3dPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/4.4.8/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
