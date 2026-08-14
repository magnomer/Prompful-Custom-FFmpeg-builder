package planning

import "strings"

func (resolver LCatalogResolver) LCatalogVersionResolve(requestedVersion string) (string, bool) {
	requestedVersion = strings.TrimSpace(requestedVersion)
	if requestedVersion == "" {
		return "", false
	}
	if _, exists := resolver.VersionRecords[requestedVersion]; exists {
		return requestedVersion, true
	}
	lineKey := LReleaseKeyGet(requestedVersion)
	if lineKey == "" {
		return "", false
	}
	if recommendedVersion, found := resolver.LReleaseRecommendResolve(lineKey); found {
		return recommendedVersion, true
	}
	requestedMajor, requestedMinor, ok := LReleaseLineSplit(lineKey)
	if !ok {
		return "", false
	}
	highestLineKey, highestVersion, found := resolver.LReleaseHighestResolve()
	if !found {
		return "", false
	}
	highestMajor, highestMinor, ok := LReleaseLineSplit(highestLineKey)
	if !ok {
		return "", false
	}
	if requestedMajor > highestMajor || (requestedMajor == highestMajor && requestedMinor > highestMinor) {
		return highestVersion, true
	}
	return "", false
}

func (resolver LCatalogResolver) LReleaseRecommendResolve(releaseLineKey string) (string, bool) {
	bestVersion := ""
	for versionId, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		if LCatalogFieldGet(ffmpeg, "releaseLine") != releaseLineKey {
			continue
		}
		if bestVersion == "" || LVersionSemverCompare(versionId, bestVersion) > 0 {
			bestVersion = versionId
		}
	}
	return bestVersion, bestVersion != ""
}

func (resolver LCatalogResolver) LReleaseHighestResolve() (string, string, bool) {
	bestLineKey := ""
	bestVersion := ""
	bestMajor, bestMinor := -1, -1
	for versionId, versionRecord := range resolver.VersionRecords {
		ffmpeg, _ := versionRecord["ffmpeg"].(map[string]any)
		lineKey := LCatalogFieldGet(ffmpeg, "releaseLine")
		major, minor, ok := LReleaseLineSplit(lineKey)
		if !ok {
			continue
		}
		if major > bestMajor || (major == bestMajor && minor > bestMinor) || (major == bestMajor && minor == bestMinor && LVersionSemverCompare(versionId, bestVersion) > 0) {
			bestMajor, bestMinor = major, minor
			bestLineKey, bestVersion = lineKey, versionId
		}
	}
	return bestLineKey, bestVersion, bestVersion != ""
}

func LArchiveURLResolve(requestedVersion string, resolvedCatalogVersion string, ffmpegObject map[string]any) string {
	requestedVersion = strings.TrimSpace(requestedVersion)
	resolvedCatalogVersion = strings.TrimSpace(resolvedCatalogVersion)
	if requestedVersion == "" || requestedVersion == resolvedCatalogVersion {
		return LCatalogFieldGet(ffmpegObject, "archiveUrl")
	}
	if LReleaseKeyGet(requestedVersion) == "" {
		return ""
	}
	return LReleaseArchiveResolve(requestedVersion)
}

func LSignatureURLResolve(requestedVersion string, resolvedCatalogVersion string, ffmpegObject map[string]any, archiveUrl string) string {
	requestedVersion = strings.TrimSpace(requestedVersion)
	resolvedCatalogVersion = strings.TrimSpace(resolvedCatalogVersion)
	if requestedVersion == "" || requestedVersion == resolvedCatalogVersion {
		return LCatalogFieldGet(ffmpegObject, "signatureUrl")
	}
	if archiveUrl == "" {
		return ""
	}
	return archiveUrl + ".asc"
}
