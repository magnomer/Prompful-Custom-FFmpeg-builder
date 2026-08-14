package planning

import "testing"

func TestCatalogResolverFfmpeg519LibjxlUsesInternalPreparation(t *testing.T) {
	resolvedPlan, err := LCatalogEmbeddedResolve(LCatalogResolutionSettings{
		FfmpegVersion:           "5.1.9",
		WindowsShellProfileName: "ucrt64",
		SelectedLibraryIds:      []string{"libjxl"},
	})
	if err != nil {
		t.Fatalf("resolve FFmpeg 5.1.9 libjxl: %v", err)
	}

	if !LStringSliceCheck(resolvedPlan.RequiredWorkIds, "ffmpeg-5.1.9-libjxl-work") {
		t.Fatalf("required work ids = %v, want libjxl preparation work", resolvedPlan.RequiredWorkIds)
	}
	if LStringSliceCheck(resolvedPlan.RequiredPackageNames, "mingw-w64-ucrt-x86_64-libjxl") {
		t.Fatalf("required packages = %v, should not use native libjxl for FFmpeg 5.1.9", resolvedPlan.RequiredPackageNames)
	}
}

func LStringSliceCheck(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
