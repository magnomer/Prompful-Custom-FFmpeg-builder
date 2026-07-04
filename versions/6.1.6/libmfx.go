package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryLibmfxPrepare performs the coded preparation manipulation for libmfx on FFmpeg 6.1.6.
func LLibraryLibmfxPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "6.1.6"
	plan.LibraryId = "libmfx"
	plan.VersionSpecificGoFile = "versions/6.1.6/libmfx.go"
	plan.LSourceCompilationUse("libmfx", "cmake")
	plan.LPackageConfigurationUse("libmfx")
	plan.LLibraryLineOverride("-L${libdir} -lmfx -lstdc++ -lole32 -luuid -ladvapi32")
	plan.LPreparationModificationAdd("libmfx.pc.cmake", "Version: 2013", "Version: 1.35")
	plan.LPreparationModificationAdd("CMakeLists.txt", "    src/mfx_win_reg_key.cpp", "    src/mfx_win_reg_key.cpp src/mfx_driver_store_loader.cpp")
	plan.LCommandVerify("mfx/mfxvideo.h", "mfx")
}
