package version703

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryKlvancPrepare performs the coded preparation manipulation for klvanc on FFmpeg 7.0.3.
func LLibraryKlvancPrepare(plan *shared.LibraryPreparationPlan) {
	plan.FfmpegVersion = "7.0.3"
	plan.LibraryId = "klvanc"
	plan.VersionSpecificGoFile = "versions/7.0.3/klvanc.go"
	plan.UseInternalSourceBuild("libklvanc", "configure-make")
	plan.RunAutogenBeforeConfigure()
	plan.RequireMsysBuildPackages("autoconf-wrapper", "automake-wrapper", "libtool")
	plan.AddConfigureOptions("--enable-shared=no")
	plan.AddMakeVariables("SUBDIRS=src")
	plan.PatchSource("src/core-private.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.PatchSource("src/libklvanc/vanc.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.PatchSource("src/libklvanc/vanc-lines.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.PatchSource("src/libklvanc/vanc-packets.h", "#include <sys/errno.h>", "#include <errno.h>")
	plan.Verify("libklvanc/vanc.h", "klvanc")
}
