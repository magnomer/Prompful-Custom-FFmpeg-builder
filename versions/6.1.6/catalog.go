package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryPreparationList lists executable version/library preparation hooks for FFmpeg 6.1.6.
var LLibraryPreparationList = map[string]shared.LibraryPreparationManipulator{
	"avisynthplus": LLibraryAvisynthplusPrepare,
	"davs2":        LLibraryDavs2Prepare,
	"klvanc":       LLibraryKlvancPrepare,
	"libmfx":       LLibraryLibmfxPrepare,
	"libtls":       LLibraryLibtlsPrepare,
	"svt-av1":      LLibrarySvtav1Prepare,
	"tensorflow":   LLibraryTensorflowPrepare,
	"uavs3d":       LLibraryUavs3dPrepare,
	"xavs2":        LLibraryXavs2Prepare,
}
