package reviewsession

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

type LSessionReview struct {
	ReviewSessionId          string `json:"reviewSessionId"`
	ActionName               string `json:"actionName"`
	PlanHash                 string `json:"planHash"`
	ExpectedLConsentText     string `json:"expectedLConsentText"`
	ExpectedLConsentTextHash string `json:"expectedLConsentTextHash"`
	CreatedAtUnixTime        int64  `json:"createdAtUnixTime"`
	ExpiresAtUnixTime        int64  `json:"expiresAtUnixTime"`
	WasUsed                  bool   `json:"wasUsed"`
}

func LSessionReviewCreate(actionName string, planHash string, lifetime time.Duration) (LSessionReview, error) {
	if actionName == "" {
		return LSessionReview{}, errors.New("review session action name is empty")
	}
	if planHash == "" {
		return LSessionReview{}, errors.New("review session plan hash is empty")
	}
	reviewSessionId, err := LReviewIdentifierCreate()
	if err != nil {
		return LSessionReview{}, err
	}
	now := time.Now().UTC()
	consentText := "I approve action " + actionName + " with plan hash " + planHash + "."
	consentHashBytes := sha256.Sum256([]byte(consentText))
	return LSessionReview{ReviewSessionId: reviewSessionId, ActionName: actionName, PlanHash: planHash, ExpectedLConsentText: consentText, ExpectedLConsentTextHash: hex.EncodeToString(consentHashBytes[:]), CreatedAtUnixTime: now.Unix(), ExpiresAtUnixTime: now.Add(lifetime).Unix()}, nil
}

func LReviewApprovalCheck(reviewSession LSessionReview, approvedActionName string, approvedPlanHash string, consentText string) error {
	if reviewSession.WasUsed {
		return errors.New("review session has already been used")
	}
	if time.Now().UTC().Unix() > reviewSession.ExpiresAtUnixTime {
		return errors.New("review session has expired")
	}
	if approvedActionName != reviewSession.ActionName {
		return errors.New("approved action name does not match review session")
	}
	if approvedPlanHash != reviewSession.PlanHash {
		return errors.New("approved plan hash does not match review session")
	}
	consentHashBytes := sha256.Sum256([]byte(consentText))
	if hex.EncodeToString(consentHashBytes[:]) != reviewSession.ExpectedLConsentTextHash {
		return errors.New("consent text does not match reviewed approval text")
	}
	return nil
}

func LReviewIdentifierCreate() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
