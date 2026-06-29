package extraction

import (
	"archive/tar"
	"archive/zip"
	"compress/bzip2"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
	"promptfulcustomffmpegbuilder/internal/consent"
	"promptfulcustomffmpegbuilder/internal/workspace"
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
	ArchiveFormatZip   ArchiveFormat = "zip"

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
	if extractPlan.ArchiveFormat == ArchiveFormatZip {
		return extractZipArchive(ctx, extractPlan)
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
		targetPath, err := safeExtractTargetPath(extractPlan.DestinationDirectory, header.Name)
		if err != nil {
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
			// Archive links are never materialized. Creating a symlink could redirect a later
			// entry's write outside the destination (symlink traversal), so the link is skipped
			// entirely; every real-file write above is still containment-checked, so not creating
			// the link removes the traversal vector completely. A link whose target stays inside
			// the destination — e.g. a source tarball's relative helper symlink like libvmaf's
			// ptools/Makefile.Linux64 -> Makefile.Linux — is skipped silently and the build, which
			// does not use it, proceeds. A link whose target escapes the destination is the classic
			// hostile setup and is rejected so a malicious archive still fails loudly.
			if !linkTargetWithinDestination(extractPlan.DestinationDirectory, header.Name, header.Linkname, header.Typeflag == tar.TypeLink) {
				return fmt.Errorf("archive link target escapes extraction root: %s -> %s", header.Name, header.Linkname)
			}
		default:
			// Ignore metadata-only entries such as pax headers.
		}
	}
}

// extractZipArchive extracts a .zip archive (used for vendor binary archives that
// ship as .zip on Windows) under the same safety bounds as the tar path: zip-slip
// protection via cleanEntryName + checkExtractTarget + workspace containment checks,
// per-file and total size limits, file-count limit, no symlinks, and the file-mode
// policy. archive/zip needs random access, so this is a separate path from the
// streaming tar reader rather than another openArchiveReader case.
func extractZipArchive(ctx context.Context, extractPlan ExtractPlan) error {
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
		return err
	}
	if err := os.MkdirAll(extractPlan.DestinationDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
		return err
	}
	zipReader, err := zip.OpenReader(extractPlan.ArchiveFilePath)
	if err != nil {
		return err
	}
	defer zipReader.Close()

	totalExtractedBytes := int64(0)
	fileCount := 0
	for _, zipEntry := range zipReader.File {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		if zipEntry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("archive links are blocked for safety: %s", zipEntry.Name)
		}
		targetPath, err := safeExtractTargetPath(extractPlan.DestinationDirectory, zipEntry.Name)
		if err != nil {
			return err
		}
		if err := workspace.CheckPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetPath); err != nil {
			return err
		}
		if zipEntry.FileInfo().IsDir() {
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetPath); err != nil {
				return err
			}
			continue
		}
		fileCount++
		if fileCount > extractPlan.MaximumFileCount {
			return errors.New("archive contains more files than allowed")
		}
		entrySize := int64(zipEntry.UncompressedSize64)
		if entrySize < 0 {
			return fmt.Errorf("archive entry has invalid size: %s", zipEntry.Name)
		}
		if entrySize > extractPlan.MaximumSingleFileByteCount {
			return fmt.Errorf("archive entry exceeds single file size limit: %s", zipEntry.Name)
		}
		totalExtractedBytes += entrySize
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
		if err := workspace.CheckRealPathInsideWorkspace(extractPlan.WorkspaceDirectory, targetDirectory); err != nil {
			return err
		}
		if extractPlan.ExtractDestinationPolicy != OverwriteExtractedFiles {
			if _, statError := os.Lstat(targetPath); statError == nil {
				return fmt.Errorf("archive extraction would overwrite existing file: %s", zipEntry.Name)
			} else if !errors.Is(statError, os.ErrNotExist) {
				return statError
			}
		}
		if err := writeZipEntry(zipEntry, targetPath, entrySize, extractPlan); err != nil {
			return err
		}
	}
	return nil
}

func writeZipEntry(zipEntry *zip.File, targetPath string, entrySize int64, extractPlan ExtractPlan) error {
	entryReader, err := zipEntry.Open()
	if err != nil {
		return err
	}
	defer entryReader.Close()
	fileMode := safeFileMode(zipEntry.Mode(), extractPlan.ExtractedFileModePolicy)
	outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		if extractPlan.ExtractDestinationPolicy == OverwriteExtractedFiles && errors.Is(err, os.ErrExist) {
			if fileInfo, lstatErr := os.Lstat(targetPath); lstatErr != nil {
				return lstatErr
			} else if fileInfo.Mode()&os.ModeSymlink != 0 {
				return errors.New("archive extraction refuses to overwrite symlink")
			}
			outputFile, err = os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, fileMode)
		}
		if err != nil {
			return err
		}
	}
	_, copyErr := io.CopyN(outputFile, entryReader, entrySize)
	closeErr := outputFile.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
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

// linkTargetWithinDestination reports whether an archive link entry resolves to a path inside the
// extraction destination. The link is never created either way; this only decides whether to skip
// it silently (benign, in-tree target) or reject the whole archive as hostile (the target escapes
// the destination — the classic symlink-traversal setup). Hardlink targets are archive-root
// relative; symlink targets are relative to the entry's own directory, and an absolute symlink
// target is treated as root-relative so it fails the containment check below. Names use forward
// slashes (tar), so path (not filepath) is used to resolve them before the OS-path containment check.
func linkTargetWithinDestination(destinationDirectory string, entryName string, linkName string, isHardlink bool) bool {
	if linkName == "" {
		return false
	}
	archiveRelativeTarget := linkName
	if !isHardlink && !path.IsAbs(linkName) {
		archiveRelativeTarget = path.Join(path.Dir(entryName), linkName)
	}
	_, err := safeExtractTargetPath(destinationDirectory, archiveRelativeTarget)
	return err == nil
}

func safeExtractTargetPath(destinationDirectory string, headerName string) (string, error) {
	cleanName, err := cleanEntryName(headerName)
	if err != nil {
		return "", err
	}
	localName := filepath.FromSlash(cleanName)
	if !filepath.IsLocal(localName) {
		return "", fmt.Errorf("unsafe archive path: %s", headerName)
	}
	targetPath := filepath.Join(destinationDirectory, localName)
	if err := checkExtractTarget(destinationDirectory, targetPath, headerName); err != nil {
		return "", err
	}
	return targetPath, nil
}

func cleanEntryName(headerName string) (string, error) {
	if headerName == "" {
		return "", errors.New("archive entry has empty path")
	}
	normalizedName := strings.ReplaceAll(headerName, "\\", "/")
	cleanName := path.Clean(normalizedName)
	if cleanName == "." || !fs.ValidPath(cleanName) || path.IsAbs(normalizedName) || hasWindowsDrivePrefix(normalizedName) || strings.HasPrefix(cleanName, "../") || strings.Contains(cleanName, "/../") {
		return "", fmt.Errorf("unsafe archive path: %s", headerName)
	}
	return cleanName, nil
}

func hasWindowsDrivePrefix(name string) bool {
	return len(name) >= 2 && name[1] == ':' && ((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z'))
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
