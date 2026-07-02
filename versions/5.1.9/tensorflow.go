package version519

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 5.1.9.
func LLibraryTensorflowPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "5.1.9"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/5.1.9/tensorflow.go"
	plan.UseExternalVendorImport("TensorFlow")
	plan.ImportFromSubdirs("include", "lib")
	plan.Verify("tensorflow/c/c_api.h", "tensorflow")
}
