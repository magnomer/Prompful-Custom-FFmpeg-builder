package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 4.4.8.
func LLibraryTensorflowPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/4.4.8/tensorflow.go"
	plan.LVendorSourceUse("TensorFlow")
	plan.LSubdirectoryLoad("include", "lib")
	plan.LCommandVerify("tensorflow/c/c_api.h", "tensorflow")
}
