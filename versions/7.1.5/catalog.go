package version715

import "promptfulcustomffmpegbuilder/versions/shared"

// LPreparationCatalog lists executable version/library preparation hooks for FFmpeg 7.1.5.
var LPreparationCatalog = map[string]shared.LPreparationManipulator{
	"avisynthplus": LLibraryAvisynthplusPrepare,
	"davs2":        LDavs2Prepare,
	"klvanc":       LLibraryKlvancPrepare,
	"lcevc-dec":    LLibraryLcevcdecPrepare,
	"libmfx":       LLibraryLibmfxPrepare,
	"libtls":       LLibraryLibtlsPrepare,
	"quirc":        LLibraryQuircPrepare,
	"svt-av1":      LSvtav1Prepare,
	"tensorflow":   LLibraryTensorflowPrepare,
	"uavs3d":       LUavs3dPrepare,
	"vvenc":        LLibraryVvencPrepare,
	"xavs2":        LXavs2Prepare,
}
