package program

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"os"
	"strings"

	"github.com/ProtonMail/go-crypto/openpgp"
)

func LSignatureDetachedVerify(signaturePath string, archivePath string, publicKeyPath string, expectedPrimaryFingerprint string, signatureLabel string, emitProgress func(string, string)) error {
	publicKeyBytes, err := os.ReadFile(publicKeyPath)
	if err != nil {
		return fmt.Errorf("could not read public signing key: %w", err)
	}
	keyRing, err := LSignatureKeyringRead(publicKeyBytes)
	if err != nil {
		return fmt.Errorf("could not read public signing key: %w", err)
	}
	if !LSignatureFingerprintCheck(keyRing, expectedPrimaryFingerprint) {
		return fmt.Errorf("downloaded signing key did not match the expected fingerprint %s", expectedPrimaryFingerprint)
	}
	archiveBytes, err := os.ReadFile(archivePath)
	if err != nil {
		return fmt.Errorf("could not read archive for signature verification: %w", err)
	}
	signatureBytes, err := os.ReadFile(signaturePath)
	if err != nil {
		return fmt.Errorf("could not read detached signature: %w", err)
	}
	signer, err := LSignatureDetachedCheck(keyRing, archiveBytes, signatureBytes)
	if err != nil {
		return fmt.Errorf("detached signature check failed: %w", err)
	}
	if signer == nil || !strings.EqualFold(hex.EncodeToString(signer.PrimaryKey.Fingerprint[:]), expectedPrimaryFingerprint) {
		return fmt.Errorf("signature was valid, but it was not made by the expected key %s", expectedPrimaryFingerprint)
	}
	if emitProgress != nil {
		emitProgress("info", signatureLabel+" verification passed without requiring system GPG.")
	}
	return nil
}

func LSignatureKeyringRead(publicKeyBytes []byte) (openpgp.EntityList, error) {
	trimmedKey := bytes.TrimSpace(publicKeyBytes)
	if bytes.HasPrefix(trimmedKey, []byte("-----BEGIN PGP PUBLIC KEY BLOCK-----")) {
		return openpgp.ReadArmoredKeyRing(bytes.NewReader(publicKeyBytes))
	}
	return openpgp.ReadKeyRing(bytes.NewReader(publicKeyBytes))
}

func LSignatureDetachedCheck(keyRing openpgp.EntityList, archiveBytes []byte, signatureBytes []byte) (*openpgp.Entity, error) {
	trimmedSignature := bytes.TrimSpace(signatureBytes)
	if bytes.HasPrefix(trimmedSignature, []byte("-----BEGIN PGP SIGNATURE-----")) {
		return openpgp.CheckArmoredDetachedSignature(keyRing, bytes.NewReader(archiveBytes), bytes.NewReader(signatureBytes), nil)
	}
	return openpgp.CheckDetachedSignature(keyRing, bytes.NewReader(archiveBytes), bytes.NewReader(signatureBytes), nil)
}

func LSignatureFingerprintCheck(keyRing openpgp.EntityList, expectedPrimaryFingerprint string) bool {
	for _, entity := range keyRing {
		if entity == nil || entity.PrimaryKey == nil {
			continue
		}
		if strings.EqualFold(hex.EncodeToString(entity.PrimaryKey.Fingerprint[:]), expectedPrimaryFingerprint) {
			return true
		}
	}
	return false
}
