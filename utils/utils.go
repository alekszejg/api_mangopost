package utils

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
)

func GetRandomState() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func ValidateSignature(signature string, secret string, rawData json.RawMessage) error {
	if signature == "" {
		return fmt.Errorf("can't validate an empty signature")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(rawData)
	expectedMAC := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(signature), []byte(expectedMAC)) {
		return fmt.Errorf("mismatch in computed & provided signatures")
	}

	return nil
}
