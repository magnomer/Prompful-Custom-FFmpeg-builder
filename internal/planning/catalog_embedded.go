package planning

import (
	"encoding/json"
	"fmt"
	"strings"

	"promptfulcustomffmpegbuilder/internal/catalogfacts"
)

// LCatalogEmbeddedRoot is the shared embedded source catalog root.
const LCatalogEmbeddedRoot = "catalogdata"

// LCatalogEmbeddedFile stores one embedded JSON catalog file before schema-specific resolution.
type LCatalogEmbeddedFile struct {
	DomainName LCatalogDomainName `json:"domainName"`
	PathName   string             `json:"pathName"`
	FileName   string             `json:"fileName"`
	BaseName   string             `json:"baseName"`
	RawContent []byte             `json:"-"`
}

// LCatalogEmbedded contains all embedded catalog files loaded from the executable.
type LCatalogEmbedded struct {
	LibraryFiles []LCatalogEmbeddedFile `json:"libraryFiles"`
	VersionFiles []LCatalogEmbeddedFile `json:"versionFiles"`
	PresetFiles  []LCatalogEmbeddedFile `json:"presetFiles"`
}

// LCatalogEmbeddedLoad reads the embedded libraries, versions, and presets catalog files.
// Phase 2 only loads and validates this data; it does not change the planner's decisions.
func LCatalogEmbeddedLoad() (LCatalogEmbedded, error) {
	catalog := LCatalogEmbedded{}
	for _, domainName := range []LCatalogDomainName{LCatalogDomainLibraries, LCatalogDomainVersions, LCatalogDomainPresets} {
		files, err := LCatalogEmbeddedDomainLoad(domainName)
		if err != nil {
			return LCatalogEmbedded{}, err
		}
		switch domainName {
		case LCatalogDomainLibraries:
			catalog.LibraryFiles = files
		case LCatalogDomainVersions:
			catalog.VersionFiles = files
		case LCatalogDomainPresets:
			catalog.PresetFiles = files
		}
	}
	return catalog, nil
}

// LCatalogEmbeddedDomainLoad reads one embedded catalog domain.
func LCatalogEmbeddedDomainLoad(domainName LCatalogDomainName) ([]LCatalogEmbeddedFile, error) {
	directoryName, err := LCatalogEmbeddedDomainDirectoryName(domainName)
	if err != nil {
		return nil, err
	}
	sharedFiles, err := catalogfacts.CatalogDataDomainFilesLoad(directoryName)
	if err != nil {
		return nil, fmt.Errorf("read embedded catalog domain %q: %w", domainName, err)
	}
	files := []LCatalogEmbeddedFile{}
	for _, sharedFile := range sharedFiles {
		files = append(files, LCatalogEmbeddedFile{
			DomainName: domainName,
			PathName:   sharedFile.Path,
			FileName:   LCatalogEmbeddedFileNameRead(sharedFile.Path),
			BaseName:   sharedFile.Base,
			RawContent: sharedFile.RawContent,
		})
	}
	return files, nil
}

func LCatalogEmbeddedFileNameRead(pathName string) string {
	pathName = strings.ReplaceAll(pathName, "\\", "/")
	lastSlashIndex := strings.LastIndex(pathName, "/")
	if lastSlashIndex < 0 {
		return pathName
	}
	return pathName[lastSlashIndex+1:]
}

func LCatalogEmbeddedDomainDirectoryName(domainName LCatalogDomainName) (string, error) {
	switch domainName {
	case LCatalogDomainLibraries:
		return LCatalogEmbeddedRoot + "/libraries", nil
	case LCatalogDomainVersions:
		return LCatalogEmbeddedRoot + "/versions", nil
	case LCatalogDomainPresets:
		return LCatalogEmbeddedRoot + "/presets", nil
	default:
		return "", fmt.Errorf("unknown embedded catalog domain %q", domainName)
	}
}

func LCatalogJsonDecode[T any](file LCatalogEmbeddedFile) (T, error) {
	var value T
	decoder := json.NewDecoder(strings.NewReader(string(file.RawContent)))
	if err := decoder.Decode(&value); err != nil {
		return value, fmt.Errorf("decode embedded catalog file %q: %w", file.PathName, err)
	}
	return value, nil
}
