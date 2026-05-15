package consent

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type ConsentKind string

const (
	ConsentKindMsys2Download             ConsentKind = "msys2-download"
	ConsentKindFfmpegSourceDownload      ConsentKind = "ffmpeg-source-download"
	ConsentKindArchiveExtraction         ConsentKind = "archive-extraction"
	ConsentKindPacmanPackageInstallation ConsentKind = "pacman-package-installation"
	ConsentKindExternalCommandExecution  ConsentKind = "external-command-execution"
	ConsentKindWorkspaceDeletion         ConsentKind = "workspace-deletion"
)

type Consent struct {
	ConsentId          string      `json:"consentId"`
	ConsentKind        ConsentKind `json:"kind"`
	ApprovedActionName string      `json:"approvedActionName"`
	ApprovedPlanHash   string      `json:"approvedPlanHash"`
	ApprovedAtUnixTime int64       `json:"approvedAtUnixTime"`
	ConsentText        string      `json:"consentText"`
}

type Msys2DownloadConsent struct{ Consent }
type FfmpegSourceDownloadConsent struct{ Consent }
type ArchiveExtractionConsent struct{ Consent }
type PacmanInstallConsent struct{ Consent }
type CommandExecutionConsent struct{ Consent }
type WorkspaceDeletionConsent struct{ Consent }

type ApprovalRequest struct {
	ApprovedActionName string `json:"approvedActionName"`
	ApprovedPlanHash   string `json:"approvedPlanHash"`
	ConsentText        string `json:"consentText"`
}

func Msys2DownloadApproval(approval ApprovalRequest) (Msys2DownloadConsent, error) {
	approvalRecord, err := createConsent(ConsentKindMsys2Download, approval)
	return Msys2DownloadConsent{Consent: approvalRecord}, err
}

func FfmpegSourceDownloadApproval(approval ApprovalRequest) (FfmpegSourceDownloadConsent, error) {
	approvalRecord, err := createConsent(ConsentKindFfmpegSourceDownload, approval)
	return FfmpegSourceDownloadConsent{Consent: approvalRecord}, err
}

func ArchiveExtractionApproval(approval ApprovalRequest) (ArchiveExtractionConsent, error) {
	approvalRecord, err := createConsent(ConsentKindArchiveExtraction, approval)
	return ArchiveExtractionConsent{Consent: approvalRecord}, err
}

func PacmanInstallApproval(approval ApprovalRequest) (PacmanInstallConsent, error) {
	approvalRecord, err := createConsent(ConsentKindPacmanPackageInstallation, approval)
	return PacmanInstallConsent{Consent: approvalRecord}, err
}

func CommandExecutionApproval(approval ApprovalRequest) (CommandExecutionConsent, error) {
	approvalRecord, err := createConsent(ConsentKindExternalCommandExecution, approval)
	return CommandExecutionConsent{Consent: approvalRecord}, err
}

func WorkspaceDeletionApproval(approval ApprovalRequest) (WorkspaceDeletionConsent, error) {
	approvalRecord, err := createConsent(ConsentKindWorkspaceDeletion, approval)
	return WorkspaceDeletionConsent{Consent: approvalRecord}, err
}

func CheckConsent(approvalRecord Consent, expectedConsentKind ConsentKind, expectedActionName string, expectedPlanHash string) error {
	if approvalRecord.ConsentId == "" {
		return errors.New("missing user consent id")
	}
	if approvalRecord.ConsentKind != expectedConsentKind {
		return errors.New("user consent kind does not match requested operation")
	}
	if approvalRecord.ApprovedActionName != expectedActionName {
		return errors.New("user consent action does not match requested operation")
	}
	if approvalRecord.ApprovedPlanHash != expectedPlanHash {
		return errors.New("user consent plan hash does not match requested operation")
	}
	if approvalRecord.ApprovedAtUnixTime <= 0 {
		return errors.New("missing user consent timestamp")
	}
	if approvalRecord.ConsentText == "" {
		return errors.New("missing user consent text")
	}
	return nil
}

func createConsent(kind ConsentKind, approval ApprovalRequest) (Consent, error) {
	if approval.ApprovedActionName == "" {
		return Consent{}, errors.New("missing approved action name")
	}
	if approval.ApprovedPlanHash == "" {
		return Consent{}, errors.New("missing approved plan hash")
	}
	if approval.ConsentText == "" {
		return Consent{}, errors.New("missing consent text")
	}
	consentId, err := createConsentId()
	if err != nil {
		return Consent{}, err
	}
	return Consent{
		ConsentId:          consentId,
		ConsentKind:        kind,
		ApprovedActionName: approval.ApprovedActionName,
		ApprovedPlanHash:   approval.ApprovedPlanHash,
		ApprovedAtUnixTime: time.Now().UTC().Unix(),
		ConsentText:        approval.ConsentText,
	}, nil
}

func createConsentId() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
