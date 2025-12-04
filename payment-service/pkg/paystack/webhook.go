package paystack

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
)

// ValidateWebhookSignature validates Paystack webhook signature
func ValidateWebhookSignature(signature string, payload []byte, secretKey string) bool {
	mac := hmac.New(sha512.New, []byte(secretKey))
	mac.Write(payload)
	expectedSignature := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}