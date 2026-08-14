package version901

import "promptfulcustomffmpegbuilder/versions/shared"

// LPreparationCatalog lists executable version/library preparation hooks for FFmpeg 9.0.1.
var LPreparationCatalog = map[string]shared.LPreparationManipulator{
	"opencv":       LOpencvPrepare,
	"avisynthplus": LLibraryAvisynthplusPrepare,
	"davs2":        LDavs2Prepare,
	"klvanc":       LLibraryKlvancPrepare,
	"lcevc-dec":    LLibraryLcevcdecPrepare,
	"libmfx":       LLibraryLibmfxPrepare,
	"libtls":       LLibraryLibtlsPrepare,
	"mpeghdec":     LLibraryMpeghdecPrepare,
	"quirc":        LLibraryQuircPrepare,
	"tensorflow":   LLibraryTensorflowPrepare,
	"uavs3d":       LUavs3dPrepare,
	"vvenc":        LLibraryVvencPrepare,
	"xavs2":        LXavs2Prepare,
}
