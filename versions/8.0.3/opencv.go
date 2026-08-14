package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LOpencvPrepare performs the coded preparation manipulation for opencv on FFmpeg 8.0.3.
// OpenCV is source-built (pinned 4.x) because FFmpeg's --enable-libopencv needs the legacy
// C API (opencv2/core/core_c.h + cvCreateImageHeader), which MSYS2's current OpenCV 5 package
// no longer ships. The build installs the standard opencv4 layout (opencv4.pc + headers under
// include/opencv4 + unversioned libopencv_core/imgproc) so the existing LOpenCVScriptCreate
// FFmpeg-configure shim keeps satisfying the legacy probe. cmake flags live in the source pin.
func LOpencvPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "opencv"
	plan.VersionSpecificGoFile = "versions/8.0.3/opencv.go"
	plan.LSourceCompilationUse("OpenCV", "cmake")
	plan.LPackageConfigurationUse("opencv4")
	plan.LCommandVerify("opencv4/opencv2/core/core_c.h", "opencv_core")
}
