package main

import "testing"

func TestArgsParse(t *testing.T) {
	parsed, err := LArgumentParse([]string{
		"--ffmpeg-version", "8.1.2",
		"--preset=full",
		"--extended",
		"--enable-libx264",
		"--disable-libx265",
		"--workspace", "D:\\PromptfulWork",
		"--jobs", "8",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.version != "8.1.2" || parsed.preset != "full" || !parsed.extended {
		t.Fatalf("bad scalar parse: %+v", parsed)
	}
	if parsed.workspace != "D:\\PromptfulWork" || parsed.jobs != 8 {
		t.Fatalf("bad workspace/jobs parse: %+v", parsed)
	}
	if len(parsed.enable) != 1 || parsed.enable[0] != "--enable-libx264" {
		t.Fatalf("bad enable parse: %v", parsed.enable)
	}
	if len(parsed.disable) != 1 || parsed.disable[0] != "--disable-libx265" {
		t.Fatalf("bad disable parse: %v", parsed.disable)
	}
}

func TestArgsParseErrors(t *testing.T) {
	if _, err := LArgumentParse([]string{"--bogus"}); err == nil {
		t.Fatalf("expected unknown-flag error")
	}
	if _, err := LArgumentParse([]string{"--jobs", "-3"}); err == nil {
		t.Fatalf("expected negative-jobs error")
	}
	if _, err := LArgumentParse([]string{"--ffmpeg-version"}); err == nil {
		t.Fatalf("expected missing-value error")
	}
}

func TestSettingsResolve(t *testing.T) {
	parsed, err := LArgumentParse([]string{
		"--ffmpeg-version", "8.1.2",
		"--preset", "minimal",
		"--enable-libx264",
		"--workspace", "D:\\PromptfulWork",
	})
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	settings, err := LSettingsFFmpegResolve(parsed)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if settings.FfmpegSourceArchiveUrl == "" {
		t.Fatalf("archive URL not resolved")
	}
	found := false
	for _, id := range settings.SelectedLibraryIds {
		if id == "x264" {
			found = true
		}
	}
	if !found {
		t.Fatalf("--enable-libx264 did not add x264: %v", settings.SelectedLibraryIds)
	}
}

func TestSettingsResolveExitCodes(t *testing.T) {
	// Missing version -> bad args (2).
	if _, err := LSettingsFFmpegResolve(LArgumentBuild{}); err == nil {
		t.Fatalf("expected missing-version error")
	} else if usage, ok := err.(LErrorUsage); !ok || usage.code != 2 {
		t.Fatalf("missing version: want code 2, got %v", err)
	}
	// Unknown version -> unsupported (4).
	if _, err := LSettingsFFmpegResolve(LArgumentBuild{version: "9.9.9"}); err == nil {
		t.Fatalf("expected unknown-version error")
	} else if usage, ok := err.(LErrorUsage); !ok || usage.code != 4 {
		t.Fatalf("unknown version: want code 4, got %v", err)
	}
}
