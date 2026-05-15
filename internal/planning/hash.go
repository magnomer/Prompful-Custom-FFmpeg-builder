package planning

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func HashPlan(planValue interface{}) (string, error) {
	planBytes, err := json.Marshal(planValue)
	if err != nil {
		return "", err
	}
	planHashBytes := sha256.Sum256(planBytes)
	return hex.EncodeToString(planHashBytes[:]), nil
}
