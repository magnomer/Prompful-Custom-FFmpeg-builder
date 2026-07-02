package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryUavs3dPrepare performs the coded preparation manipulation for uavs3d on FFmpeg 8.0.3.
func LLibraryUavs3dPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "uavs3d"
	plan.VersionSpecificGoFile = "versions/8.0.3/uavs3d.go"
	plan.UseInternalSourceBuild("libuavs3d", "cmake")
	plan.UsePkgConfig("uavs3d")
	plan.Verify("uavs3d.h", "uavs3d")
}
