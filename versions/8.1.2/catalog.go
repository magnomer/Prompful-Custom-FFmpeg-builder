package version812

import "promptfulcustomffmpegbuilder/versions/shared"

// LLibraryPreparationList lists executable version/library preparation hooks for FFmpeg 8.1.2.
var LLibraryPreparationList = map[string]shared.LibraryPreparationManipulator{
	"avisynthplus": LLibraryAvisynthplusPrepare,
	"davs2":        LLibraryDavs2Prepare,
	"klvanc":       LLibraryKlvancPrepare,
	"lcevc-dec":    LLibraryLcevcdecPrepare,
	"libmfx":       LLibraryLibmfxPrepare,
	"libtls":       LLibraryLibtlsPrepare,
	"mpeghdec":     LLibraryMpeghdecPrepare,
	"quirc":        LLibraryQuircPrepare,
	"tensorflow":   LLibraryTensorflowPrepare,
	"uavs3d":       LLibraryUavs3dPrepare,
	"vvenc":        LLibraryVvencPrepare,
	"xavs2":        LLibraryXavs2Prepare,
}
