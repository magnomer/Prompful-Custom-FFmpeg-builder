package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 5.1.9.
func LLibraryTensorflowPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/5.1.9/tensorflow.go"
	plan.LVendorSourceUse("TensorFlow")
	plan.LSubdirectoryLoad("include", "lib")
	plan.LCommandVerify("tensorflow/c/c_api.h", "tensorflow")
}
