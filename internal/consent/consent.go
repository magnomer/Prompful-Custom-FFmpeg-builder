package consent

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strconv"
	"strings"
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
	LConsentSignature  string       `json:"consentSignature"`
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

// lConsentSecret is a process-scoped HMAC key generated once at startup. Every
// consent record this backend mints is signed with it, so a record crafted
// elsewhere (for example a structure passed straight into a Wails-bound worker
// method by the WebView) cannot present a valid signature and fails
// LConsentCheck. The key never leaves the backend process.
var lConsentSecret = lConsentSecretCreate()

func lConsentSecretCreate() []byte {
	secret := make([]byte, 32)
	if _, err := rand.Read(secret); err != nil {
		panic("consent: unable to initialize signing secret: " + err.Error())
	}
	return secret
}

// LConsentSignatureCreate computes the authentication tag over a consent
// record's authoritative fields. The signature field itself is excluded so the
// same call both produces and re-derives the tag for verification.
func LConsentSignatureCreate(approvalRecord LConsent) string {
	message := strings.Join([]string{
		approvalRecord.LConsentId,
		string(approvalRecord.LConsentKind),
		approvalRecord.ApprovedActionName,
		approvalRecord.ApprovedPlanHash,
		strconv.FormatInt(approvalRecord.ApprovedAtUnixTime, 10),
		approvalRecord.LConsentText,
	}, "\x00")
	signer := hmac.New(sha256.New, lConsentSecret)
	signer.Write([]byte(message))
	return hex.EncodeToString(signer.Sum(nil))
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
	expectedSignature := LConsentSignatureCreate(approvalRecord)
	if approvalRecord.LConsentSignature == "" || !hmac.Equal([]byte(approvalRecord.LConsentSignature), []byte(expectedSignature)) {
		return errors.New("user consent signature is not authentic")
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
	approvalRecord := LConsent{
		LConsentId:         consentId,
		LConsentKind:       kind,
		ApprovedActionName: approval.ApprovedActionName,
		ApprovedPlanHash:   approval.ApprovedPlanHash,
		ApprovedAtUnixTime: time.Now().UTC().Unix(),
		LConsentText:       approval.LConsentText,
	}
	approvalRecord.LConsentSignature = LConsentSignatureCreate(approvalRecord)
	return approvalRecord, nil
}

func LConsentIdCreate() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
