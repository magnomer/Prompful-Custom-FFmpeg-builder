package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryXevePrepare performs the coded preparation manipulation for xeve on FFmpeg 7.0.3.
func LLibraryXevePrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "xeve"
	plan.VersionSpecificGoFile = "versions/7.0.3/xeve.go"
	plan.UseInternalSourceBuild("XEVE", "cmake")
	plan.WriteGeneratedSourceFile("version.txt", "v0.4.3")
	plan.PatchSource("CMakeLists.txt", `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Werror -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`, `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`)
	plan.UsePkgConfig("xeve")
	plan.OverridePkgConfigLibsLine("-L${libdir}/xeve -l:libxeve.a -lm -lpthread")
	plan.Verify("xeve/xeve.h", "xeve")
}
