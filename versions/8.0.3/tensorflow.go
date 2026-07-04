package version803

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 8.0.3.
func LLibraryTensorflowPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.0.3"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/8.0.3/tensorflow.go"
	plan.LVendorSourceUse("TensorFlow")
	plan.LSubdirectoryLoad("include", "lib")
	plan.LCommandVerify("tensorflow/c/c_api.h", "tensorflow")
}
