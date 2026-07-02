package version448

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryVmafPrepare performs the coded preparation manipulation for vmaf on FFmpeg 4.4.8.
func LLibraryVmafPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "4.4.8"
	plan.LibraryId = "vmaf"
	plan.VersionSpecificGoFile = "versions/4.4.8/vmaf.go"
	plan.UseInternalSourceBuild("libvmaf", "meson")
	plan.RequireBuildPackages("meson", "ninja")
	plan.ConfigureInSubdir("libvmaf")
	plan.AddCMakeOptions("-Denable_tests=false", "-Denable_docs=false")
	plan.WriteGeneratedSourceFile("libvmaf/include/vcs_version.h", "/* auto-generated, do not edit */", "#define VMAF_VERSION \"v1.5.2\"")
	plan.AddCFlags("-Wno-error=implicit-function-declaration", "-Wno-error=implicit-int", "-Wno-error=int-conversion", "-Wno-error=incompatible-pointer-types")
	plan.UsePkgConfig("libvmaf")
	plan.AppendPkgConfigLibs("stdc++", "m")
	plan.Verify("libvmaf/libvmaf.h", "vmaf")
}
