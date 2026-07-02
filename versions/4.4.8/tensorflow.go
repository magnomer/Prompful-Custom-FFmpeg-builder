package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryTensorflowPrepare performs the coded preparation manipulation for tensorflow on FFmpeg 4.4.8.
func LLibraryTensorflowPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "tensorflow"
	plan.VersionSpecificGoFile = "versions/4.4.8/tensorflow.go"
	plan.UseExternalVendorImport("TensorFlow")
	plan.ImportFromSubdirs("include", "lib")
	plan.Verify("tensorflow/c/c_api.h", "tensorflow")
}
