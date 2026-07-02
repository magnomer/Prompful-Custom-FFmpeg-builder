package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLcevcdecPrepare performs the coded preparation manipulation for lcevc-dec on FFmpeg 7.1.5.
func LLibraryLcevcdecPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.1.5"
	plan.LibraryId = "lcevc-dec"
	plan.VersionSpecificGoFile = "versions/7.1.5/lcevc_dec.go"
	plan.UseInternalSourceBuild("liblcevc-dec", "cmake")
	plan.RequireBuildPackages("python")
	plan.AddCMakeOptions("-DBUILD_SHARED_LIBS=OFF", "-DVN_SDK_EXECUTABLES=OFF", "-DVN_SDK_UNIT_TESTS=OFF", "-DVN_SDK_SAMPLE_SOURCE=OFF", "-DVN_SDK_JSON_CONFIG=OFF", "-DVN_SDK_PIPELINE_VULKAN=OFF", "-DVN_SDK_DOCS=OFF", "-DVN_SDK_SYSTEM_INSTALL=OFF")
	plan.UsePkgConfig("lcevc_dec")
	plan.AppendPkgConfigLibs("stdc++", "m")
	plan.Verify("LCEVC/lcevc_dec.h", "lcevc_dec_api")
}
