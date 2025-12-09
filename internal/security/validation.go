package security

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"strings"
)

// ValidateSignature checks the GitHub X-Hub-Signature-256 header against the request body.
func ValidateSignature(signatureHeader string, payloadBody []byte, secret string) bool {
	const prefix = "sha256="
	
	if !strings.HasPrefix(signatureHeader, prefix) {
		log.Println("Validation Error: Signature header missing 'sha256=' prefix.")
		return false
	}
	
	actualSignatureHex := strings.TrimPrefix(signatureHeader, prefix)
	
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payloadBody)
	
	actualSignature, err := hex.DecodeString(actualSignatureHex)
	if err != nil {
		log.Printf("Validation Error: Failed to decode signature header: %v", err)
		return false
	}

	// Use constant-time comparison for security
	return hmac.Equal(actualSignature, mac.Sum(nil))
}
