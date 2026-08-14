package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 9.0.1.
func LLibraryTensorflowPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/9.0.1/tensorflow.go"
	plan.LVendorSourceUse("TensorFlow")
	plan.LSubdirectoryLoad("include", "lib")
	plan.LCommandVerify("tensorflow/c/c_api.h", "tensorflow")
}
