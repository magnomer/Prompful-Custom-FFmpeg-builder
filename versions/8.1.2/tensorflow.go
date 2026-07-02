package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 8.1.2.
func LLibraryTensorflowPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/8.1.2/tensorflow.go"
	plan.UseExternalVendorImport("TensorFlow")
	plan.ImportFromSubdirs("include", "lib")
	plan.Verify("tensorflow/c/c_api.h", "tensorflow")
}
