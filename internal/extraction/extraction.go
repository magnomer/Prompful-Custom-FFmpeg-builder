package extraction

import (
	"archive/tar"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"strings"

	"customffmpegbuilder/internal/consent"
	"customffmpegbuilder/internal/workspace"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

type ArchiveFormat string

type ExtractDestinationPolicy string

type ExtractedFileModePolicy string

const (
	TarBz2             ArchiveFormat = "tar-bz2"
	ArchiveFormatTarGz ArchiveFormat = "tar-gz"
	TarXz              ArchiveFormat = "tar-xz"
	TarZst             ArchiveFormat = "tar-zst"
	ArchiveFormatTar   ArchiveFormat = "tar"

	RequireNewDirectory                    ExtractDestinationPolicy = "must-not-exist"
	ExtractionDestinationPolicyMustBeEmpty ExtractDestinationPolicy = "must-be-empty"
	OverwriteExtractedFiles                ExtractDestinationPolicy = "overwrite-approved"

	PreserveSafeExecutableBits              ExtractedFileModePolicy = "preserve-safe-executable-bits"
	FileModePolicyRegularFilesNotExecutable ExtractedFileModePolicy = "regular-files-not-executable"
)

type ExtractPlan struct {
	ActionName                 string                   `json:"actionName"`
	PlanHash                   string                   `json:"planHash"`
	ArchiveFilePath            string                   `json:"archiveFilePath"`
	DestinationDirectory       string                   `json:"destinationDirectory"`
	WorkspaceDirectory         string                   `json:"workspaceDirectory"`
	ArchiveFormat              ArchiveFormat            `json:"archiveFormatName"`
	ExtractDestinationPolicy   ExtractDestinationPolicy `json:"extractionDestinationPolicyName"`
	ExtractedFileModePolicy    ExtractedFileModePolicy  `json:"fileModePolicyName"`
	MaximumFileCount           int                      `json:"maximumFileCount"`
	MaximumExtractedByteCount  int64                    `json:"maximumExtractedByteCount"`
	MaximumSingleFileByteCount int64                    `json:"maximumSingleFileByteCount"`
}

type ProgressFunc func(level string, message string)

func ExtractArchiveWithConsent(ctx context.Context, userArchiveExtractionConsent consent.ArchiveExtractionConsent, extractPlan ExtractPlan, emitProgress ProgressFunc) error {
	if err := consent.CheckConsent(userArchiveExtractionConsent.Consent, consent.ConsentKindArchiveExtraction, extractPlan.ActionName, extractPlan.PlanHash); err != nil {
		return err
	}
	extractPlan = applyExtractDefaults(extractPlan)
	if err := validateExtractPlan(extractPlan); err != nil {
		return err
	}
	if emitProgress != nil {
		emitProgress("info", "Extracting approved archive inside workspace.")
	}
	return extractTarArchive(ctx, extractPlan)
}

func applyExtractDefaults(extractPlan ExtractPlan) ExtractPlan {
	if extractPlan.ExtractDestinationPolicy == "" {
		extractPlan.ExtractDestinationPolicy = RequireNewDirectory
	}
	if extractPlan.ExtractedFileModePolicy == "" {
		extractPlan.ExtractedFileModePolicy = PreserveSafeExecutableBits
	}
	if extractPlan.MaximumFileCount <= 0 {
		extractPlan.MaximumFileCount = 250000
	}
	if extractPlan.MaximumExtractedByteCount <= 0 {
		extractPlan.MaximumExtractedByteCount = 10_000_000_000
	}
	if extractPlan.MaximumSingleFileByteCount <= 0 {
		extractPlan.MaximumSingleFileByteCount = 2_000_000_000
	}
	return extractPlan
}

func validateExtractPlan(extractPlan ExtractPlan) error {
	if extractPlan.ArchiveFilePath == "" || extractPlan.DestinationDirectory == "" || extractPlan.WorkspaceDirectory == "" {
		return errors.New("archive extraction paths must not be empty")
	}
	if err := workspace.CheckPathInsideWorkspace(extractPlan.WorkspaceDirectory, extractPlan.ArchiveFilePath); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, extractPlan.ArchiveFilePath); err != nil {
		return err
	}
	if err := workspace.CheckPathInsideWorkspace(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
		return err
	}
	if extractPlan.MaximumFileCount < 1 {
		return errors.New("archive maximum file count must be positive")
	}
	if extractPlan.MaximumExtractedByteCount < 1 {
		return errors.New("archive maximum extracted byte count must be positive")
	}
	if extractPlan.MaximumSingleFileByteCount < 1 {
		return errors.New("archive maximum single file byte count must be positive")
	}
	return checkExtractDestinationPolicy(extractPlan)
}

func checkExtractDestinationPolicy(extractPlan ExtractPlan) error {
	fileInfo, statError := os.Lstat(extractPlan.DestinationDirectory)
	if errors.Is(statError, os.ErrNotExist) {
		return nil
	}
	if statError != nil {
		return statError
	}
	if fileInfo.Mode()&os.ModeSymlink != 0 {
		return errors.New("archive extraction destination must not be a symlink")
	}
	if !fileInfo.IsDir() {
		return errors.New("archive extraction destination exists and is not a directory")
	}
	switch extractPlan.ExtractDestinationPolicy {
	case RequireNewDirectory:
		return errors.New("archive extraction destination already exists")
	case ExtractionDestinationPolicyMustBeEmpty:
		entries, err := os.ReadDir(extractPlan.DestinationDirectory)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return errors.New("archive extraction destination must be empty")
		}
		return nil
	case OverwriteExtractedFiles:
		return nil
	default:
		return fmt.Errorf("unknown extraction destination policy: %s", extractPlan.ExtractDestinationPolicy)
	}
}

func extractTarArchive(ctx context.Context, extractPlan ExtractPlan) error {
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
		return err
	}
	if err := os.MkdirAll(extractPlan.DestinationDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
		return err
	}
	archiveFile, err := os.Open(extractPlan.ArchiveFilePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	archiveReader, closeReader, err := openArchiveReader(archiveFile, extractPlan.ArchiveFormat)
	if err != nil {
		return err
	}
	if closeReader != nil {
		defer closeReader()
	}

	totalExtractedBytes := int64(0)
	fileCount := 0
	tarReader := tar.NewReader(archiveReader)
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		cleanEntryName, err := cleanEntryName(header.Name)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(extractPlan.DestinationDirectory, filepath.FromSlash(cleanEntryName))
		if err := checkExtractTarget(extractPlan.DestinationDirectory, targetPath, header.Name); err != nil {
			return err
		}
		if err := workspace.CheckPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetPath); err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetPath); err != nil {
				return err
			}
		case tar.TypeReg, tar.TypeRegA:
			fileCount++
			if fileCount > extractPlan.MaximumFileCount {
				return errors.New("archive contains more files than allowed")
			}
			if header.Size < 0 {
				return fmt.Errorf("archive entry has invalid size: %s", header.Name)
			}
			if header.Size > extractPlan.MaximumSingleFileByteCount {
				return fmt.Errorf("archive entry exceeds single file size limit: %s", header.Name)
			}
			totalExtractedBytes += header.Size
			if totalExtractedBytes > extractPlan.MaximumExtractedByteCount {
				return errors.New("archive extracted byte count exceeds limit")
			}
			targetDirectory := filepath.Dir(targetPath)
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetDirectory); err != nil {
				return err
			}
			if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
				return err
			}
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetPath); err != nil {
				return err
			}
			if extractPlan.ExtractDestinationPolicy != OverwriteExtractedFiles {
				if _, statError := os.Lstat(targetPath); statError == nil {
					return fmt.Errorf("archive extraction would overwrite existing file: %s", header.Name)
				} else if !errors.Is(statError, os.ErrNotExist) {
					return statError
				}
			}
			outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, safeFileMode(header.FileInfo().Mode(), extractPlan.ExtractedFileModePolicy))
			if err != nil {
				if extractPlan.ExtractDestinationPolicy == OverwriteExtractedFiles && errors.Is(err, os.ErrExist) {
					if fileInfo, lstatErr := os.Lstat(targetPath); lstatErr != nil {
						err = lstatErr
					} else if fileInfo.Mode()&os.ModeSymlink != 0 {
						err = errors.New("archive extraction refuses to overwrite symlink")
					} else {
						outputFile, err = os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, safeFileMode(header.FileInfo().Mode(), extractPlan.ExtractedFileModePolicy))
					}
				}
				if err != nil {
					return err
				}
			}
			_, copyErr := io.CopyN(outputFile, tarReader, header.Size)
			closeErr := outputFile.Close()
			if copyErr != nil {
				return copyErr
			}
			if closeErr != nil {
				return closeErr
			}
		case tar.TypeSymlink, tar.TypeLink:
			return fmt.Errorf("archive links are blocked for safety: %s", header.Name)
		default:
			// Ignore metadata-only entries such as pax headers.
		}
	}
}

func openArchiveReader(archiveFile *os.File, archiveFormatName ArchiveFormat) (io.Reader, func(), error) {
	switch archiveFormatName {
	case TarBz2:
		return bzip2.NewReader(archiveFile), nil, nil
	case ArchiveFormatTarGz:
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return gzipReader, func() { _ = gzipReader.Close() }, nil
	case TarXz:
		xzReader, err := xz.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return xzReader, nil, nil
	case TarZst:
		zstdReader, err := zstd.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return zstdReader, func() { zstdReader.Close() }, nil
	case ArchiveFormatTar:
		return archiveFile, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported archive format: %s", archiveFormatName)
	}
}

func cleanEntryName(headerName string) (string, error) {
	if headerName == "" {
		return "", errors.New("archive entry has empty path")
	}
	normalizedName := strings.ReplaceAll(headerName, "\\", "/")
	cleanName := path.Clean(normalizedName)
	if cleanName == "." || cleanName == ".." || strings.HasPrefix(cleanName, "../") || strings.HasPrefix(cleanName, "/") || strings.Contains(cleanName, "/../") {
		return "", fmt.Errorf("unsafe archive path: %s", headerName)
	}
	return cleanName, nil
}

func checkExtractTarget(destinationDirectory string, targetPath string, originalHeaderName string) error {
	absoluteDestinationDirectory, err := filepath.Abs(destinationDirectory)
	if err != nil {
		return err
	}
	absoluteTargetPath, err := filepath.Abs(targetPath)
	if err != nil {
		return err
	}
	relativePath, err := filepath.Rel(absoluteDestinationDirectory, absoluteTargetPath)
	if err != nil {
		return err
	}
	if relativePath == "." || strings.HasPrefix(relativePath, "..") || filepath.IsAbs(relativePath) {
		return fmt.Errorf("archive path escapes destination: %s", originalHeaderName)
	}
	return nil
}

func safeFileMode(fileMode os.FileMode, fileModePolicyName ExtractedFileModePolicy) os.FileMode {
	switch fileModePolicyName {
	case FileModePolicyRegularFilesNotExecutable:
		return fileMode & 0o644
	case PreserveSafeExecutableBits, "":
		return fileMode & 0o755
	default:
		return fileMode & 0o755
	}
}
