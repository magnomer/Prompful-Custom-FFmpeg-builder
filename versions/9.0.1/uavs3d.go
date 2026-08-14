package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavs3dPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 9.0.1.
func LUavs3dPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/9.0.1/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
