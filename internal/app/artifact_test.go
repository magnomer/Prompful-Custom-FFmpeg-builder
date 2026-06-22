package app

import "testing"

func TestIsFfmpegSharedLibraryName(t *testing.T) {
	matching := []string{
		"libavcodec-62.dll",
		"libavutil-60.dll",
		"libavformat-62.dll",
		"libavdevice-62.dll",
		"libavfilter-11.dll",
		"libswscale-9.dll",
		"libswresample-6.dll",
		"libpostproc-58.dll",
		"avcodec-62.dll",
		"LIBAVCODEC-62.DLL",
	}
	for _, name := range matching {
		if !isFfmpegSharedLibraryName(name) {
			t.Errorf("expected %s to be recognized as an FFmpeg shared library", name)
		}
	}

	nonMatching := []string{
		"libfdk-aac-2.dll",
		"libx264-164.dll",
		"zlib1.dll",
		"libavcodec.dll",
		"libavcodec-.dll",
		"libavcodec-62a.dll",
		"avcodec-62.exe",
		"notavcodec-62.dll",
		"libavcodec-62.dll.bak",
	}
	for _, name := range nonMatching {
		if isFfmpegSharedLibraryName(name) {
			t.Errorf("expected %s to NOT be recognized as an FFmpeg shared library", name)
		}
	}
}
