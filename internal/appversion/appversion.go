// Package appversion carries the application's own release version, distinct
// from any FFmpeg release version.
package appversion

// Version is the release version compiled into the Promptful executables.
// The package test verifies that it matches the repository manifests.
var Version = "7.0.1280"
