package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVmafPrepare performs the coded preparation manipulation for vmaf on FFmpeg 4.4.8.
func LLibraryVmafPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "vmaf"
	plan.VersionSpecificGoFile = "versions/4.4.8/vmaf.go"
	plan.LSourceCompilationUse("libvmaf", "meson")
	plan.LBuildPackageRequire("meson", "ninja")
	plan.LConfigureSubdirectoryUse("libvmaf")
	plan.LCMakeOptionAdd("-Denable_tests=false", "-Denable_docs=false")
	plan.LGeneratedFileWrite("libvmaf/include/vcs_version.h", "/* auto-generated, do not edit */", "#define VMAF_VERSION \"v1.5.2\"")
	plan.LCompilerFlagAdd("-Wno-error=implicit-function-declaration", "-Wno-error=implicit-int", "-Wno-error=int-conversion", "-Wno-error=incompatible-pointer-types")
	plan.LPackageConfigurationUse("libvmaf")
	plan.LLibraryLineAppend("stdc++", "m")
	plan.LCommandVerify("libvmaf/libvmaf.h", "vmaf")
}
