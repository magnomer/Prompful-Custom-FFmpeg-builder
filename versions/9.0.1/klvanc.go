package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryKlvancPrepare performs the coded preparation manipulation for klvanc on FFmpeg 9.0.1.
func LLibraryKlvancPrepare(plan *shared.LPreparationPlan) {
	plan.FfmpegVersion = "9.0.1"
	plan.LibraryId = "klvanc"
	plan.VersionSpecificGoFile = "versions/9.0.1/klvanc.go"
	plan.LSourceCompilationUse("libklvanc", "configure-make")
	plan.LAutogenBeforeRun()
	plan.LMsysPackageRequire("autoconf-wrapper", "automake-wrapper", "libtool")
	plan.LConfigureOptionAdd("--enable-shared=no")
	plan.LMakeVariableAdd("SUBDIRS=src")
	plan.LPreparationModificationAdd("src/core-private.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc-lines.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LPreparationModificationAdd("src/libklvanc/vanc-packets.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.LCommandVerify("libklvanc/vanc.h", "klvanc")
}
