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
	// lSessionExpiryMonotonic is the authoritative validity deadline, read from
	// the process monotonic clock so wall-clock adjustments cannot extend or
	// shorten the real lifetime. It is unexported and unserialized on purpose:
	// a monotonic reading has no meaning outside this process, so only the
	// in-memory session held server-side carries it. The wall-clock
	// ExpiresAtUnixTime remains for display and audit.
	lSessionExpiryMonotonic time.Time
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
	// now keeps its monotonic reading; deadline (Add preserves it) is the
	// authoritative validity boundary. The Unix stamps are taken from the
	// wall-clock projection for display and audit only.
	now := time.Now()
	deadline := now.Add(lifetime)
	wallNow := now.UTC()
	consentText := "I approve action " + actionName + " with plan hash " + planHash + "."
	consentHashBytes := sha256.Sum256([]byte(consentText))
	return LSessionReview{ReviewSessionId: reviewSessionId, ActionName: actionName, PlanHash: planHash, ExpectedLConsentText: consentText, ExpectedLConsentTextHash: hex.EncodeToString(consentHashBytes[:]), CreatedAtUnixTime: wallNow.Unix(), ExpiresAtUnixTime: deadline.UTC().Unix(), lSessionExpiryMonotonic: deadline}, nil
}

// LReviewExpiryCheck reports whether the review session's validity deadline has
// passed. It measures against the monotonic deadline captured at creation, so
// system wall-clock adjustments cannot extend or shorten the real lifetime, and
// treats the exact deadline instant as expired. A session that predates the
// monotonic deadline (zero value) falls back to the wall-clock stamp. Callers
// re-run this after any user-facing wait (native dialog, CLI prompt) so a
// review that expires while confirmation is pending cannot be approved.
func LReviewExpiryCheck(reviewSession LSessionReview) error {
	if !reviewSession.lSessionExpiryMonotonic.IsZero() {
		if !time.Now().Before(reviewSession.lSessionExpiryMonotonic) {
			return errors.New("review session has expired")
		}
		return nil
	}
	if time.Now().UTC().Unix() >= reviewSession.ExpiresAtUnixTime {
		return errors.New("review session has expired")
	}
	return nil
}

func LReviewApprovalCheck(reviewSession LSessionReview, approvedActionName string, approvedPlanHash string, consentText string) error {
	if reviewSession.WasUsed {
		return errors.New("review session has already been used")
	}
	if err := LReviewExpiryCheck(reviewSession); err != nil {
		return err
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
