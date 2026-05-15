package reviewsession

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"
)

type PlanReviewSession struct {
	ReviewSessionId         string `json:"reviewSessionId"`
	ActionName              string `json:"actionName"`
	PlanHash                string `json:"planHash"`
	ExpectedConsentText     string `json:"expectedConsentText"`
	ExpectedConsentTextHash string `json:"expectedConsentTextHash"`
	CreatedAtUnixTime       int64  `json:"createdAtUnixTime"`
	ExpiresAtUnixTime       int64  `json:"expiresAtUnixTime"`
	WasUsed                 bool   `json:"wasUsed"`
}

func NewPlanReviewSession(actionName string, planHash string, lifetime time.Duration) (PlanReviewSession, error) {
	if actionName == "" {
		return PlanReviewSession{}, errors.New("review session action name is empty")
	}
	if planHash == "" {
		return PlanReviewSession{}, errors.New("review session plan hash is empty")
	}
	reviewSessionId, err := createReviewSessionId()
	if err != nil {
		return PlanReviewSession{}, err
	}
	now := time.Now().UTC()
	consentText := "I approve action " + actionName + " with plan hash " + planHash + "."
	consentHashBytes := sha256.Sum256([]byte(consentText))
	return PlanReviewSession{ReviewSessionId: reviewSessionId, ActionName: actionName, PlanHash: planHash, ExpectedConsentText: consentText, ExpectedConsentTextHash: hex.EncodeToString(consentHashBytes[:]), CreatedAtUnixTime: now.Unix(), ExpiresAtUnixTime: now.Add(lifetime).Unix()}, nil
}

func CheckReviewApproval(reviewSession PlanReviewSession, approvedActionName string, approvedPlanHash string, consentText string) error {
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
	if hex.EncodeToString(consentHashBytes[:]) != reviewSession.ExpectedConsentTextHash {
		return errors.New("consent text does not match reviewed approval text")
	}
	return nil
}

func createReviewSessionId() (string, error) {
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(randomBytes), nil
}
