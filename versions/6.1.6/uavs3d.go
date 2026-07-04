package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavs3dPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 6.1.6.
func LUavs3dPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/6.1.6/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
