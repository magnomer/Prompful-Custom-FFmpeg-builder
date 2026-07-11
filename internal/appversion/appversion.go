// Package appversion carries the application's own release version (distinct
// from any FFmpeg release version). The single source of truth is version.json
// at the repo root; build.ps1 stamps Version at compile time via
// -ldflags "-X promptfulcustomffmpegbuilder/internal/appversion.Version=<v>".
//
// A plain `go build` that does not pass the ldflag leaves Version at "dev".
package appversion

// Version is the release version of the Promptful executables. It is overwritten
// at link time from version.json; "dev" marks an unstamped local build.
var Version = "dev"
