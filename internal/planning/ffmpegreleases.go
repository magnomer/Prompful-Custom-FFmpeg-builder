package planning

import "sort"

// LReleaseChoice is one selectable FFmpeg source release: its version, FFmpeg codename, and
// the canonical archive and detached-signature (.asc) URLs the UI fills when it is chosen. The
// UI stores none of this itself; it requests this list and writes only the chosen URLs into the
// FFmpeg build settings.
type LReleaseChoice struct {
	Version      string `json:"version"`
	Codename     string `json:"codename"`
	ArchiveUrl   string `json:"archiveUrl"`
	SignatureUrl string `json:"signatureUrl"`
}

// LReleaseArchiveResolve returns the canonical ffmpeg.org source archive URL for a version.
func LReleaseArchiveResolve(version string) string {
	return "https://www.ffmpeg.org/releases/ffmpeg-" + version + ".tar.xz"
}

// LReleaseSupportedGet lists every recommended FFmpeg source release, newest first, with
// the archive and matching .asc signature URLs. This is the single source of truth for the
// source-version dropdown.
func LReleaseSupportedGet() []LReleaseChoice {
	versions := LReleaseRecommendList()
	choices := make([]LReleaseChoice, 0, len(versions))
	for _, version := range versions {
		archiveUrl := LReleaseArchiveResolve(version)
		choices = append(choices, LReleaseChoice{
			Version:      version,
			Codename:     LReleaseCodenameResolve(version),
			ArchiveUrl:   archiveUrl,
			SignatureUrl: archiveUrl + ".asc",
		})
	}
	return choices
}

// LReleaseLineResolve returns the "major.minor" line key of a dotted-numeric version
// (e.g. "8.1.1" -> "8.1"), or "" when the version is not dotted-numeric with at least a
// major and minor component (e.g. a git snapshot), so callers skip release advisories. It
// uses the embedded catalog release-line parser so planning no longer links the retired release-support package.
func LReleaseLineResolve(version string) string {
	return LReleaseKeyGet(version)
}

// LReleaseRecommendList lists every recommended patch release, newest first.
// Used for the supported-versions advisory message and to find the newest known release.
func LReleaseRecommendList() []string {
	releases := LReleaseRawGet()
	sort.Slice(releases, func(left, right int) bool {
		comparison, _ := LVersionCompare(releases[left], releases[right])
		return comparison > 0
	})
	return releases
}

func LReleaseRawGet() []string {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return []string{}
	}
	releases := []string{}
	for versionId := range resolver.VersionRecords {
		releases = append(releases, versionId)
	}
	return releases
}

func LReleaseCodenameResolve(version string) string {
	resolver, _, err := LCatalogResolverLoad()
	if err != nil {
		return ""
	}
	versionRecord, exists := resolver.VersionRecords[version]
	if !exists {
		return ""
	}
	ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
	return LCatalogFieldGet(ffmpeg, "codename")
}

func LReleaseHighestGet() string {
	releases := LReleaseRecommendList()
	if len(releases) == 0 {
		return ""
	}
	return releases[0]
}
