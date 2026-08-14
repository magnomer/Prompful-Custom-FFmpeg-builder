package version616

import "promptfulcustomffmpegbuilder/versions/shared"

// LPreparationCatalog lists executable version/library preparation hooks for FFmpeg 6.1.6.
var LPreparationCatalog = map[string]shared.LPreparationManipulator{
	"opencv":       LOpencvPrepare,
	"avisynthplus": LLibraryAvisynthplusPrepare,
	"davs2":        LDavs2Prepare,
	"klvanc":       LLibraryKlvancPrepare,
	"libmfx":       LLibraryLibmfxPrepare,
	"libtls":       LLibraryLibtlsPrepare,
	"svt-av1":      LSvtav1Prepare,
	"tensorflow":   LLibraryTensorflowPrepare,
	"uavs3d":       LUavsPrepare,
	"xavs2":        LXavs2Prepare,
}
