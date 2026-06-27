package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"promptfulcustomffmpegbuilder/internal/planning"
	"promptfulcustomffmpegbuilder/shared/librarysources"
)

// requirePkgConfigMinVersionPattern extracts the minimum version FFmpeg's configure
// demands for a pkg-config module, from a line like:
//
//	require_pkg_config libvvenc "libvvenc >= 1.6.1" "vvenc/vvenc.h" vvenc_get_version
//
// The module name inside the quoted constraint is what we match on.
func requiredPkgConfigMinVersion(configureText string, pkgConfigModule string) (string, bool) {
	pattern := regexp.MustCompile(`"` + regexp.QuoteMeta(pkgConfigModule) + `\s*>=\s*([0-9][0-9.]*)"`)
	match := pattern.FindStringSubmatch(configureText)
	if match == nil {
		return "", false
	}
	return match[1], true
}

// validateLibraryVersionsAgainstFfmpeg is a preflight that catches the case where the
// version pinned for a prepared library is older than the minimum the selected FFmpeg
// release requires (a newer FFmpeg can raise a library's required floor). It reads the
// already-extracted FFmpeg configure, compares each prepared library's resolved version
// against that library's require_pkg_config minimum, and fails early ??before any library
// is downloaded or built ??with an actionable message, instead of letting it surface as a
// cryptic "X >= V not found" at the end of FFmpeg configure.
//
// It only constrains libraries that declare a pkg-config module and pin a dotted-numeric
// version; header-only (avisynth) and vendor-imported (tensorflow) libraries, and moving
// refs like "master", are skipped because there is nothing decidable to compare.
func (app *App) validateLibraryVersionsAgainstFfmpeg(plan planning.FfmpegBuildPlan, ffmpegSourceDirectory string, emitProgress func(string, string)) error {
	if len(plan.LibraryPreparations) == 0 {
		return nil
	}
	configureBytes, err := os.ReadFile(filepath.Join(ffmpegSourceDirectory, "configure"))
	if err != nil {
		// Without configure we cannot preflight; skip rather than block, since the real
		// FFmpeg configure step still gates the build.
		emitProgress("warn", localize("run.log.libraryVersionCheckSkipped", nil))
		return nil
	}
	configureText := string(configureBytes)
	for _, preparation := range plan.LibraryPreparations {
		if preparation.PkgConfigName == "" {
			continue
		}
		requiredMinimum, found := requiredPkgConfigMinVersion(configureText, preparation.PkgConfigName)
		if !found {
			continue
		}
		comparison, decidable := librarysources.CompareVersions(preparation.Version, requiredMinimum)
		if !decidable {
			continue
		}
		if comparison < 0 {
			return fmt.Errorf("%s", localize("run.failure.libraryVersionTooOld", map[string]string{
				"library":  preparation.DisplayName,
				"pinned":   preparation.Version,
				"required": requiredMinimum,
				"module":   preparation.PkgConfigName,
			}))
		}
		emitProgress("info", localize("run.log.libraryVersionOk", map[string]string{
			"library":  preparation.DisplayName,
			"pinned":   preparation.Version,
			"required": requiredMinimum,
		}))
	}
	return nil
}
