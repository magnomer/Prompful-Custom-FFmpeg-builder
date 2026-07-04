package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 7.1.5.
func LLibraryTensorflowPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/7.1.5/tensorflow.go"
	plan.LVendorSourceUse("TensorFlow")
	plan.LSubdirectoryLoad("include", "lib")
	plan.LCommandVerify("tensorflow/c/c_api.h", "tensorflow")
}
