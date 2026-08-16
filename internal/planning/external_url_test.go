package planning

import (
	"strings"
	"testing"
)

func TestLExternalWebURLValidate(t *testing.T) {
	for _, acceptedURL := range []string{"https://example.com/project", "http://example.com"} {
		if err := LExternalWebURLValidate(acceptedURL); err != nil {
			t.Errorf("expected %q to be accepted: %v", acceptedURL, err)
		}
	}
	for _, rejectedURL := range []string{"file:///tmp/report", "custom-handler:action", "javascript:alert(1)", "https:///missing-host", " https://example.com"} {
		if err := LExternalWebURLValidate(rejectedURL); err == nil {
			t.Errorf("expected %q to be rejected", rejectedURL)
		}
	}
}

func TestLCatalogEmbeddedValidateRejectsNonWebOfficialURL(t *testing.T) {
	file := LCatalogEmbeddedFile{
		DomainName: LCatalogDomainLibraries,
		PathName:   "libraries/example.json",
		BaseName:   "example",
		RawContent: []byte(`{"recordKind":"ffmpeg-version-aware-library","libraryId":"example","ffmpegVersions":{"1.0":{"ffmpegVersion":"1.0","officialWebpageUrl":"custom-handler:action"}}}`),
	}
	report := LCatalogEmbeddedValidate(LCatalogEmbedded{LibraryFiles: []LCatalogEmbeddedFile{file}})
	found := false
	for _, issue := range report.Issues {
		if strings.Contains(issue.Message, "officialWebpageUrl is invalid") {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected catalog validation to report the unsafe official webpage URL")
	}
}
