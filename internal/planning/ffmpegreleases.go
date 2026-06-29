package planning

import (
	"sort"

	"promptfulcustomffmpegbuilder/shared/librarysources"
	"promptfulcustomffmpegbuilder/shared/releasesupport"
)

// supportedFfmpegReleases maps each maintained FFmpeg release line (major.minor) to its
// recommended patch release — the branch tip that carries that line's security and critical
// fixes. These are the only releases this builder vouches for. Update when FFmpeg publishes
// a new point release on a maintained branch (edit the value), or maintains a new line (add
// a key); both are data edits here.
var supportedFfmpegReleases = map[string]string{
	"8.1": "8.1.2",
	"8.0": "8.0.3",
	"7.1": "7.1.5",
	"7.0": "7.0.3",
	"6.1": "6.1.6",
	"5.1": "5.1.9",
	"4.4": "4.4.8",
}

// ffmpegReleaseCodenames maps each recommended patch release to its FFmpeg codename, shown
// alongside the version in the source-version dropdown. Keep in sync with supportedFfmpegReleases.
var ffmpegReleaseCodenames = map[string]string{
	"8.1.2":  "Hoare",
	"8.0.3":  "Huffman",
	"7.1.5":  "Péter",
	"7.0.3":  "Dijkstra",
	"6.1.6":  "Heaviside",
	"5.1.9":  "Riemann",
	"4.4.8":  "Rao",
}

// FfmpegReleaseChoice is one selectable FFmpeg source release: its version, FFmpeg codename, and
// the canonical archive and detached-signature (.asc) URLs the UI fills when it is chosen. The
// UI stores none of this itself; it requests this list and writes only the chosen URLs into the
// FFmpeg build settings.
type FfmpegReleaseChoice struct {
	Version      string `json:"version"`
	Codename     string `json:"codename"`
	ArchiveUrl   string `json:"archiveUrl"`
	SignatureUrl string `json:"signatureUrl"`
}

// ffmpegReleaseArchiveUrl returns the canonical ffmpeg.org source archive URL for a version.
func ffmpegReleaseArchiveUrl(version string) string {
	return "https://www.ffmpeg.org/releases/ffmpeg-" + version + ".tar.xz"
}

// SupportedFfmpegReleaseChoices lists every recommended FFmpeg source release, newest first, with
// the archive and matching .asc signature URLs. This is the single source of truth for the
// source-version dropdown.
func SupportedFfmpegReleaseChoices() []FfmpegReleaseChoice {
	versions := recommendedFfmpegReleasesDescending()
	choices := make([]FfmpegReleaseChoice, 0, len(versions))
	for _, version := range versions {
		archiveUrl := ffmpegReleaseArchiveUrl(version)
		choices = append(choices, FfmpegReleaseChoice{
			Version:      version,
			Codename:     ffmpegReleaseCodenames[version],
			ArchiveUrl:   archiveUrl,
			SignatureUrl: archiveUrl + ".asc",
		})
	}
	return choices
}

// ffmpegReleaseLineKey returns the "major.minor" line key of a dotted-numeric version
// (e.g. "8.1.1" -> "8.1"), or "" when the version is not dotted-numeric with at least a
// major and minor component (e.g. a git snapshot), so callers skip release advisories. It
// delegates to releasesupport.ReleaseLineKey so planning and scripting share one definition.
func ffmpegReleaseLineKey(version string) string {
	return releasesupport.ReleaseLineKey(version)
}

// recommendedFfmpegReleasesDescending lists every recommended patch release, newest first.
// Used for the supported-versions advisory message and to find the newest known release.
func recommendedFfmpegReleasesDescending() []string {
	releases := make([]string, 0, len(supportedFfmpegReleases))
	for _, recommended := range supportedFfmpegReleases {
		releases = append(releases, recommended)
	}
	sort.Slice(releases, func(left, right int) bool {
		comparison, _ := librarysources.CompareVersions(releases[left], releases[right])
		return comparison > 0
	})
	return releases
}

func highestRecommendedFfmpegRelease() string {
	releases := recommendedFfmpegReleasesDescending()
	if len(releases) == 0 {
		return ""
	}
	return releases[0]
}
