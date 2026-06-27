package extraction

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"promptfulcustomffmpegbuilder/internal/consent"
)

func TestExtractTarArchiveRejectsTraversalEntries(t *testing.T) {
	for _, entryName := range []string{
		"../outside.txt",
		"safe/../../outside.txt",
		"/outside.txt",
		"C:/outside.txt",
		`..\outside.txt`,
	} {
		t.Run(entryName, func(t *testing.T) {
			workspaceDirectory := t.TempDir()
			archivePath := filepath.Join(workspaceDirectory, "archive.tar")
			destinationDirectory := filepath.Join(workspaceDirectory, "extract")
			outsidePath := filepath.Join(workspaceDirectory, "outside.txt")
			writeTarArchive(t, archivePath, map[string]string{entryName: "evil"})

			err := ExtractArchiveWithConsent(context.Background(), extractionConsent(t), ExtractPlan{
				ActionName:               "extract",
				PlanHash:                 "plan",
				ArchiveFilePath:          archivePath,
				DestinationDirectory:     destinationDirectory,
				WorkspaceDirectory:       workspaceDirectory,
				ArchiveFormat:            ArchiveFormatTar,
				ExtractDestinationPolicy: RequireNewDirectory,
			}, nil)
			if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
				t.Fatalf("expected unsafe archive path error, got %v", err)
			}
			if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
				t.Fatalf("traversal entry wrote outside destination: %v", statErr)
			}
		})
	}
}

func TestExtractZipArchiveRejectsTraversalEntries(t *testing.T) {
	for _, entryName := range []string{
		"../outside.txt",
		"safe/../../outside.txt",
		"/outside.txt",
		"C:/outside.txt",
		`..\outside.txt`,
	} {
		t.Run(entryName, func(t *testing.T) {
			workspaceDirectory := t.TempDir()
			archivePath := filepath.Join(workspaceDirectory, "archive.zip")
			destinationDirectory := filepath.Join(workspaceDirectory, "extract")
			outsidePath := filepath.Join(workspaceDirectory, "outside.txt")
			writeZipArchive(t, archivePath, map[string]string{entryName: "evil"})

			err := ExtractArchiveWithConsent(context.Background(), extractionConsent(t), ExtractPlan{
				ActionName:               "extract",
				PlanHash:                 "plan",
				ArchiveFilePath:          archivePath,
				DestinationDirectory:     destinationDirectory,
				WorkspaceDirectory:       workspaceDirectory,
				ArchiveFormat:            ArchiveFormatZip,
				ExtractDestinationPolicy: RequireNewDirectory,
			}, nil)
			if err == nil || !strings.Contains(err.Error(), "unsafe archive path") {
				t.Fatalf("expected unsafe archive path error, got %v", err)
			}
			if _, statErr := os.Stat(outsidePath); !os.IsNotExist(statErr) {
				t.Fatalf("traversal entry wrote outside destination: %v", statErr)
			}
		})
	}
}

func TestExtractArchivesAllowSafeNestedEntries(t *testing.T) {
	for _, archiveFormat := range []ArchiveFormat{ArchiveFormatTar, ArchiveFormatZip} {
		t.Run(string(archiveFormat), func(t *testing.T) {
			workspaceDirectory := t.TempDir()
			archivePath := filepath.Join(workspaceDirectory, "archive"+archiveExtension(archiveFormat))
			destinationDirectory := filepath.Join(workspaceDirectory, "extract")
			switch archiveFormat {
			case ArchiveFormatTar:
				writeTarArchive(t, archivePath, map[string]string{"project/bin/tool.txt": "ok"})
			case ArchiveFormatZip:
				writeZipArchive(t, archivePath, map[string]string{"project/bin/tool.txt": "ok"})
			default:
				t.Fatalf("unsupported test archive format: %s", archiveFormat)
			}

			err := ExtractArchiveWithConsent(context.Background(), extractionConsent(t), ExtractPlan{
				ActionName:               "extract",
				PlanHash:                 "plan",
				ArchiveFilePath:          archivePath,
				DestinationDirectory:     destinationDirectory,
				WorkspaceDirectory:       workspaceDirectory,
				ArchiveFormat:            archiveFormat,
				ExtractDestinationPolicy: RequireNewDirectory,
			}, nil)
			if err != nil {
				t.Fatalf("extract failed: %v", err)
			}
			extractedBytes, err := os.ReadFile(filepath.Join(destinationDirectory, "project", "bin", "tool.txt"))
			if err != nil {
				t.Fatalf("read extracted file: %v", err)
			}
			if string(extractedBytes) != "ok" {
				t.Fatalf("unexpected extracted content: %q", extractedBytes)
			}
		})
	}
}

func extractionConsent(t *testing.T) consent.ArchiveExtractionConsent {
	t.Helper()
	approval, err := consent.ArchiveExtractionApproval(consent.ApprovalRequest{
		ApprovedActionName: "extract",
		ApprovedPlanHash:   "plan",
		ConsentText:        "test approval",
	})
	if err != nil {
		t.Fatalf("create extraction consent: %v", err)
	}
	return approval
}

func writeTarArchive(t *testing.T, archivePath string, entries map[string]string) {
	t.Helper()
	var buffer bytes.Buffer
	tarWriter := tar.NewWriter(&buffer)
	for name, content := range entries {
		header := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(content))}
		if err := tarWriter.WriteHeader(header); err != nil {
			t.Fatalf("write tar header: %v", err)
		}
		if _, err := tarWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write tar content: %v", err)
		}
	}
	if err := tarWriter.Close(); err != nil {
		t.Fatalf("close tar writer: %v", err)
	}
	if err := os.WriteFile(archivePath, buffer.Bytes(), 0o644); err != nil {
		t.Fatalf("write tar archive: %v", err)
	}
}

func writeZipArchive(t *testing.T, archivePath string, entries map[string]string) {
	t.Helper()
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		t.Fatalf("create zip archive: %v", err)
	}
	zipWriter := zip.NewWriter(archiveFile)
	for name, content := range entries {
		entryWriter, err := zipWriter.Create(name)
		if err != nil {
			t.Fatalf("create zip entry: %v", err)
		}
		if _, err := entryWriter.Write([]byte(content)); err != nil {
			t.Fatalf("write zip content: %v", err)
		}
	}
	if err := zipWriter.Close(); err != nil {
		t.Fatalf("close zip writer: %v", err)
	}
	if err := archiveFile.Close(); err != nil {
		t.Fatalf("close zip archive: %v", err)
	}
}

func archiveExtension(archiveFormat ArchiveFormat) string {
	if archiveFormat == ArchiveFormatZip {
		return ".zip"
	}
	return ".tar"
}
