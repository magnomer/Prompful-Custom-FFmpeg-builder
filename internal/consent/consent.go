package consent

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"
)

type LConsentKind string

const (
	LConsentKindMsys      LConsentKind = "msys2-download"
	LConsentKindFfmpeg    LConsentKind = "ffmpeg-source-download"
	LArchiveConsentKind   LConsentKind = "archive-extraction"
	LConsentKindPacman    LConsentKind = "pacman-package-installation"
	LConsentKindCommand   LConsentKind = "external-command-execution"
	LConsentKindWorkspace LConsentKind = "workspace-deletion"
)

type LConsent struct {
	LConsentId         string       `json:"consentId"`
	LConsentKind       LConsentKind `json:"kind"`
	ApprovedActionName string       `json:"approvedActionName"`
	ApprovedPlanHash   string       `json:"approvedPlanHash"`
	ApprovedAtUnixTime int64        `json:"approvedAtUnixTime"`
	LConsentText       string       `json:"consentText"`
}

type LConsentMsys struct{ LConsent }
type LConsentFfmpeg struct{ LConsent }
type LArchiveConsentState struct{ LConsent }
type LConsentPacman struct{ LConsent }
type LConsentCommand struct{ LConsent }
type LConsentWorkspace struct{ LConsent }

type LRequestApproval struct {
	ApprovedActionName string `json:"approvedActionName"`
	ApprovedPlanHash   string `json:"approvedPlanHash"`
	LConsentText       string `json:"consentText"`
}

func LConsentMsysCreate(approval LRequestApproval) (LConsentMsys, error) {
	approvalRecord, err := LConsentCreate(LConsentKindMsys, approval)
	return LConsentMsys{LConsent: approvalRecord}, err
}

func LConsentFfmpegCreate(approval LRequestApproval) (LConsentFfmpeg, error) {
	approvalRecord, err := LConsentCreate(LConsentKindFfmpeg, approval)
	return LConsentFfmpeg{LConsent: approvalRecord}, err
}

func LConsentArchiveCreate(approval LRequestApproval) (LArchiveConsentState, error) {
	approvalRecord, err := LConsentCreate(LArchiveConsentKind, approval)
	return LArchiveConsentState{LConsent: approvalRecord}, err
}

func LConsentPacmanCreate(approval LRequestApproval) (LConsentPacman, error) {
	approvalRecord, err := LConsentCreate(LConsentKindPacman, approval)
	return LConsentPacman{LConsent: approvalRecord}, err
}

func LConsentCommandCreate(approval LRequestApproval) (LConsentCommand, error) {
	approvalRecord, err := LConsentCreate(LConsentKindCommand, approval)
	return LConsentCommand{LConsent: approvalRecord}, err
}

func LConsentWorkspaceCreate(approval LRequestApproval) (LConsentWorkspace, error) {
	approvalRecord, err := LConsentCreate(LConsentKindWorkspace, approval)
	return LConsentWorkspace{LConsent: approvalRecord}, err
}

func LConsentCheck(approvalRecord LConsent, expectedLConsentKind LConsentKind, expectedActionName string, expectedPlanHash string) error {
	if approvalRecord.LConsentId == "" {
		return errors.New("missing user consent id")
	}
	if approvalRecord.LConsentKind != expectedLConsentKind {
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
	if approvalRecord.LConsentText == "" {
		return errors.New("missing user consent text")
	}
	return nil
}

func LConsentCreate(kind LConsentKind, approval LRequestApproval) (LConsent, error) {
	if approval.ApprovedActionName == "" {
		return LConsent{}, errors.New("missing approved action name")
	}
	if approval.ApprovedPlanHash == "" {
		return LConsent{}, errors.New("missing approved plan hash")
	}
	if approval.LConsentText == "" {
		return LConsent{}, errors.New("missing consent text")
	}
	consentId, err := LConsentIdCreate()
	if err != nil {
		return LConsent{}, err
	}
	return LConsent{
		LConsentId:         consentId,
		LConsentKind:       kind,
		ApprovedActionName: approval.ApprovedActionName,
		ApprovedPlanHash:   approval.ApprovedPlanHash,
		ApprovedAtUnixTime: time.Now().UTC().Unix(),
		LConsentText:       approval.LConsentText,
	}, nil
}

func LConsentIdCreate() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
