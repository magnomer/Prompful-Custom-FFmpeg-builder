package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LUavsPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 5.1.9.
func LUavsPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/5.1.9/uavs3d.go"
	plan.LSourceCompilationUse("libuavs3d", "cmake")
	plan.LPackageConfigurationUse("uavs3d")
	plan.LCommandVerify("uavs3d.h", "uavs3d")
}
