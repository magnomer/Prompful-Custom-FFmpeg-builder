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

type LArchiveFormat string

type LPolicyExtraction string

type LPolicyFilemode string

const (
	LArchiveTarBz2 LArchiveFormat = "tar-bz2"
	LArchiveTarGz  LArchiveFormat = "tar-gz"
	LArchiveTarXz  LArchiveFormat = "tar-xz"
	LArchiveTarZst LArchiveFormat = "tar-zst"
	LArchiveTar    LArchiveFormat = "tar"
	LArchiveZip    LArchiveFormat = "zip"

	LPolicyExtractionRequireNewDirectory LPolicyExtraction = "must-not-exist"
	LPolicyExtractionDestinationEmpty    LPolicyExtraction = "must-be-empty"
	LPolicyExtractionOverwrite           LPolicyExtraction = "overwrite-approved"

	LPolicyFilemodeExecutablePreserve LPolicyFilemode = "preserve-safe-executable-bits"
	LPolicyFilemodeRegularOnly        LPolicyFilemode = "regular-files-not-executable"
)

type LPlanExtraction struct {
	ActionName                 string            `json:"actionName"`
	PlanHash                   string            `json:"planHash"`
	ArchiveFilePath            string            `json:"archiveFilePath"`
	DestinationDirectory       string            `json:"destinationDirectory"`
	WorkspaceDirectory         string            `json:"workspaceDirectory"`
	LArchiveFormat             LArchiveFormat    `json:"archiveFormatName"`
	LPolicyExtraction          LPolicyExtraction `json:"extractionDestinationPolicyName"`
	LPolicyFilemode            LPolicyFilemode   `json:"fileModePolicyName"`
	MaximumFileCount           int               `json:"maximumFileCount"`
	MaximumExtractedByteCount  int64             `json:"maximumExtractedByteCount"`
	MaximumSingleFileByteCount int64             `json:"maximumSingleFileByteCount"`
}

type LProgressFunc func(level string, message string)

func LArchiveConsentExtract(LContext context.Context, userLConsentArchive consent.LConsentArchive, extractPlan LPlanExtraction, emitProgress LProgressFunc) error {
	if err := consent.LConsentCheck(userLConsentArchive.LConsent, consent.LConsentKindArchive, extractPlan.ActionName, extractPlan.PlanHash); err != nil {
		return err
	}
	extractPlan = LExtractionDefaultApply(extractPlan)
	if err := LPlanExtractionValidate(extractPlan); err != nil {
		return err
	}
	if emitProgress != nil {
		emitProgress("info", "Extracting approved archive inside workspace.")
	}
	if extractPlan.LArchiveFormat == LArchiveZip {
		return LArchiveZipExtract(LContext, extractPlan)
	}
	return LArchiveTarExtract(LContext, extractPlan)
}

func LExtractionDefaultApply(extractPlan LPlanExtraction) LPlanExtraction {
	if extractPlan.LPolicyExtraction == "" {
		extractPlan.LPolicyExtraction = LPolicyExtractionRequireNewDirectory
	}
	if extractPlan.LPolicyFilemode == "" {
		extractPlan.LPolicyFilemode = LPolicyFilemodeExecutablePreserve
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

func LPlanExtractionValidate(extractPlan LPlanExtraction) error {
	if extractPlan.ArchiveFilePath == "" || extractPlan.DestinationDirectory == "" || extractPlan.WorkspaceDirectory == "" {
		return errors.New("archive extraction paths must not be empty")
	}
	if err := workspace.LPathWorkspaceCheck(extractPlan.WorkspaceDirectory, extractPlan.ArchiveFilePath); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, extractPlan.ArchiveFilePath); err != nil {
		return err
	}
	if err := workspace.LPathWorkspaceCheck(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
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
	return LPolicyExtractionCheck(extractPlan)
}

func LPolicyExtractionCheck(extractPlan LPlanExtraction) error {
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
	switch extractPlan.LPolicyExtraction {
	case LPolicyExtractionRequireNewDirectory:
		return errors.New("archive extraction destination already exists")
	case LPolicyExtractionDestinationEmpty:
		entries, err := os.ReadDir(extractPlan.DestinationDirectory)
		if err != nil {
			return err
		}
		if len(entries) > 0 {
			return errors.New("archive extraction destination must be empty")
		}
		return nil
	case LPolicyExtractionOverwrite:
		return nil
	default:
		return fmt.Errorf("unknown extraction destination policy: %s", extractPlan.LPolicyExtraction)
	}
}

func LArchiveTarExtract(LContext context.Context, extractPlan LPlanExtraction) error {
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
		return err
	}
	if err := os.MkdirAll(extractPlan.DestinationDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
		return err
	}
	archiveFile, err := os.Open(extractPlan.ArchiveFilePath)
	if err != nil {
		return err
	}
	defer archiveFile.Close()

	archiveReader, closeReader, err := LArchiveReaderOpen(archiveFile, extractPlan.LArchiveFormat)
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
		case <-LContext.Done():
			return LContext.Err()
		default:
		}
		header, err := tarReader.Next()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return err
		}
		targetPath, err := LPathExtractionResolve(extractPlan.DestinationDirectory, header.Name)
		if err != nil {
			return err
		}
		if err := workspace.LPathWorkspaceCheck(extractPlan.WorkspaceDirectory, targetPath); err != nil {
			return err
		}
		switch header.Typeflag {
		case tar.TypeDir:
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetPath); err != nil {
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
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetDirectory); err != nil {
				return err
			}
			if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
				return err
			}
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetPath); err != nil {
				return err
			}
			if extractPlan.LPolicyExtraction != LPolicyExtractionOverwrite {
				if _, statError := os.Lstat(targetPath); statError == nil {
					return fmt.Errorf("archive extraction would overwrite existing file: %s", header.Name)
				} else if !errors.Is(statError, os.ErrNotExist) {
					return statError
				}
			}
			outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, LModeFileResolve(header.FileInfo().Mode(), extractPlan.LPolicyFilemode))
			if err != nil {
				if extractPlan.LPolicyExtraction == LPolicyExtractionOverwrite && errors.Is(err, os.ErrExist) {
					if fileInfo, lstatErr := os.Lstat(targetPath); lstatErr != nil {
						err = lstatErr
					} else if fileInfo.Mode()&os.ModeSymlink != 0 {
						err = errors.New("archive extraction refuses to overwrite symlink")
					} else {
						outputFile, err = os.OpenFile(targetPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, LModeFileResolve(header.FileInfo().Mode(), extractPlan.LPolicyFilemode))
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
			if !LLinkDestinationCheck(extractPlan.DestinationDirectory, header.Name, header.Linkname, header.Typeflag == tar.TypeLink) {
				return fmt.Errorf("archive link target escapes extraction root: %s -> %s", header.Name, header.Linkname)
			}
		default:
			// Ignore metadata-only entries such as pax headers.
		}
	}
}

// LArchiveZipExtract extracts a .zip archive (used for vendor binary archives that
// ship as .zip on Windows) under the same safety bounds as the tar path: zip-slip
// protection via LEntryNameClean + LTargetExtractionCheck + workspace containment checks,
// per-file and total size limits, file-count limit, no symlinks, and the file-mode
// policy. archive/zip needs random access, so this is a separate path from the
// streaming tar LReader rather than another LArchiveReaderOpen case.
func LArchiveZipExtract(LContext context.Context, extractPlan LPlanExtraction) error {
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, filepath.Dir(extractPlan.DestinationDirectory)); err != nil {
		return err
	}
	if err := os.MkdirAll(extractPlan.DestinationDirectory, 0o755); err != nil {
		return err
	}
	if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, extractPlan.DestinationDirectory); err != nil {
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
		case <-LContext.Done():
			return LContext.Err()
		default:
		}
		if zipEntry.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("archive links are blocked for safety: %s", zipEntry.Name)
		}
		targetPath, err := LPathExtractionResolve(extractPlan.DestinationDirectory, zipEntry.Name)
		if err != nil {
			return err
		}
		if err := workspace.LPathWorkspaceCheck(extractPlan.WorkspaceDirectory, targetPath); err != nil {
			return err
		}
		if zipEntry.FileInfo().IsDir() {
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, filepath.Dir(targetPath)); err != nil {
				return err
			}
			if err := os.MkdirAll(targetPath, 0o755); err != nil {
				return err
			}
			if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetPath); err != nil {
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
		if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetDirectory); err != nil {
			return err
		}
		if err := os.MkdirAll(targetDirectory, 0o755); err != nil {
			return err
		}
		if err := workspace.LPathRealCheck(extractPlan.WorkspaceDirectory, targetDirectory); err != nil {
			return err
		}
		if extractPlan.LPolicyExtraction != LPolicyExtractionOverwrite {
			if _, statError := os.Lstat(targetPath); statError == nil {
				return fmt.Errorf("archive extraction would overwrite existing file: %s", zipEntry.Name)
			} else if !errors.Is(statError, os.ErrNotExist) {
				return statError
			}
		}
		if err := LArchiveEntryWrite(zipEntry, targetPath, entrySize, extractPlan); err != nil {
			return err
		}
	}
	return nil
}

func LArchiveEntryWrite(zipEntry *zip.File, targetPath string, entrySize int64, extractPlan LPlanExtraction) error {
	entryReader, err := zipEntry.Open()
	if err != nil {
		return err
	}
	defer entryReader.Close()
	fileMode := LModeFileResolve(zipEntry.Mode(), extractPlan.LPolicyFilemode)
	outputFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_EXCL|os.O_WRONLY, fileMode)
	if err != nil {
		if extractPlan.LPolicyExtraction == LPolicyExtractionOverwrite && errors.Is(err, os.ErrExist) {
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

func LArchiveReaderOpen(archiveFile *os.File, archiveFormatName LArchiveFormat) (io.Reader, func(), error) {
	switch archiveFormatName {
	case LArchiveTarBz2:
		return bzip2.NewReader(archiveFile), nil, nil
	case LArchiveTarGz:
		gzipReader, err := gzip.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return gzipReader, func() { _ = gzipReader.Close() }, nil
	case LArchiveTarXz:
		xzReader, err := xz.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return xzReader, nil, nil
	case LArchiveTarZst:
		zstdReader, err := zstd.NewReader(archiveFile)
		if err != nil {
			return nil, nil, err
		}
		return zstdReader, func() { zstdReader.Close() }, nil
	case LArchiveTar:
		return archiveFile, nil, nil
	default:
		return nil, nil, fmt.Errorf("unsupported archive format: %s", archiveFormatName)
	}
}

// LLinkDestinationCheck reports whether an archive link entry resolves to a path inside the
// extraction destination. The link is never created either way; this only decides whether to skip
// it silently (benign, in-tree target) or reject the whole archive as hostile (the target escapes
// the destination — the classic symlink-traversal setup). Hardlink targets are archive-root
// relative; symlink targets are relative to the entry's own directory, and an absolute symlink
// target is treated as root-relative so it fails the containment check below. Names use forward
// slashes (tar), so path (not filepath) is used to resolve them before the OS-path containment check.
func LLinkDestinationCheck(destinationDirectory string, entryName string, linkName string, isHardlink bool) bool {
	if linkName == "" {
		return false
	}
	archiveRelativeTarget := linkName
	if !isHardlink && !path.IsAbs(linkName) {
		archiveRelativeTarget = path.Join(path.Dir(entryName), linkName)
	}
	_, err := LPathExtractionResolve(destinationDirectory, archiveRelativeTarget)
	return err == nil
}

func LPathExtractionResolve(destinationDirectory string, headerName string) (string, error) {
	cleanName, err := LEntryNameClean(headerName)
	if err != nil {
		return "", err
	}
	localName := filepath.FromSlash(cleanName)
	if !filepath.IsLocal(localName) {
		return "", fmt.Errorf("unsafe archive path: %s", headerName)
	}
	targetPath := filepath.Join(destinationDirectory, localName)
	if err := LTargetExtractionCheck(destinationDirectory, targetPath, headerName); err != nil {
		return "", err
	}
	return targetPath, nil
}

func LEntryNameClean(headerName string) (string, error) {
	if headerName == "" {
		return "", errors.New("archive entry has empty path")
	}
	normalizedName := strings.ReplaceAll(headerName, "\\", "/")
	cleanName := path.Clean(normalizedName)
	if cleanName == "." || !fs.ValidPath(cleanName) || path.IsAbs(normalizedName) || LPathDriveCheck(normalizedName) || strings.HasPrefix(cleanName, "../") || strings.Contains(cleanName, "/../") {
		return "", fmt.Errorf("unsafe archive path: %s", headerName)
	}
	return cleanName, nil
}

func LPathDriveCheck(name string) bool {
	return len(name) >= 2 && name[1] == ':' && ((name[0] >= 'A' && name[0] <= 'Z') || (name[0] >= 'a' && name[0] <= 'z'))
}

func LTargetExtractionCheck(destinationDirectory string, targetPath string, originalHeaderName string) error {
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

func LModeFileResolve(fileMode os.FileMode, fileModePolicyName LPolicyFilemode) os.FileMode {
	switch fileModePolicyName {
	case LPolicyFilemodeRegularOnly:
		return fileMode & 0o644
	case LPolicyFilemodeExecutablePreserve, "":
		return fileMode & 0o755
	default:
		return fileMode & 0o755
	}
}
