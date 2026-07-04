package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLcevcdecPrepare performs the coded preparation manipulation for lcevc-dec on FFmpeg 8.1.2.
func LLibraryLcevcdecPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "8.1.2"
	plan.LibraryId = "lcevc-dec"
	plan.VersionSpecificGoFile = "versions/8.1.2/lcevc_dec.go"
	plan.LSourceCompilationUse("liblcevc-dec", "cmake")
	plan.LBuildPackageRequire("python")
	plan.LCMakeOptionAdd("-DBUILD_SHARED_LIBS=OFF", "-DVN_SDK_EXECUTABLES=OFF", "-DVN_SDK_UNIT_TESTS=OFF", "-DVN_SDK_SAMPLE_SOURCE=OFF", "-DVN_SDK_JSON_CONFIG=OFF", "-DVN_SDK_PIPELINE_VULKAN=OFF", "-DVN_SDK_DOCS=OFF", "-DVN_SDK_SYSTEM_INSTALL=OFF")
	plan.LPackageConfigurationUse("lcevc_dec")
	plan.LLibraryLineAppend("stdc++", "m")
	plan.LCommandVerify("LCEVC/lcevc_dec.h", "lcevc_dec_api")
}
