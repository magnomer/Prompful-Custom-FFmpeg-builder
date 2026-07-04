package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryXevePrepare performs the coded preparation manipulation for xeve on FFmpeg 7.0.3.
func LLibraryXevePrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "xeve"
	plan.VersionSpecificGoFile = "versions/7.0.3/xeve.go"
	plan.LSourceCompilationUse("XEVE", "cmake")
	plan.LGeneratedFileWrite("version.txt", "v0.4.3")
	plan.LPreparationModificationAdd("CMakeLists.txt", `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Werror -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`, `    set (CMAKE_C_FLAGS "${CMAKE_C_FLAGS} ${OPT_DBG} -${OPT_LV} -fomit-frame-pointer -Wall -Wno-unused-function -Wno-unused-but-set-variable -Wno-unused-variable -Wno-attributes -Wno-strict-overflow -Wno-unknown-pragmas -Wno-stringop-overflow -std=c99")`)
	plan.LPackageConfigurationUse("xeve")
	plan.LLibraryLineOverride("-L${libdir}/xeve -l:libxeve.a -lm -lpthread")
	plan.LCommandVerify("xeve/xeve.h", "xeve")
}
