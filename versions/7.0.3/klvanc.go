package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryKlvancPrepare performs the coded preparation manipulation for klvanc on FFmpeg 7.0.3.
func LLibraryKlvancPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "klvanc"
	plan.VersionSpecificGoFile = "versions/7.0.3/klvanc.go"
	plan.LSourceCompilationUse("libklvanc", "configure-make")
	plan.LAutogenBeforeRun()
	plan.LMSYSPackageRequire("autoconf-wrapper", "automake-wrapper", "libtool")
	plan.LConfigureOptionAdd("--enable-shared=no")
	plan.LMakeVariableAdd("SUBDIRS=src")
	plan.LPreparationModificationAdd("src/core-private.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc-lines.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc-packets.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LCommandVerify("libklvanc/vanc.h", "klvanc")
}
