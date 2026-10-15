package ternwire

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"regexp"
	"strconv"
	"time"
)

var signaturePattern = regexp.MustCompile(`^\s*t=([1-9][0-9]{0,11})\s*,\s*v1=([a-f0-9]{64})\s*$`)

// VerifyWebhook verifies raw request bytes with a five-minute timestamp tolerance.
func VerifyWebhook(header string, body []byte, secret string) bool {
	return VerifyWebhookWithTolerance(header, body, secret, 5*time.Minute)
}

func VerifyWebhookWithTolerance(header string, body []byte, secret string, tolerance time.Duration) bool {
	parts := signaturePattern.FindStringSubmatch(header)
	if len(parts) != 3 || secret == "" || tolerance < 0 {
		return false
	}
	timestamp, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return false
	}
	difference := time.Now().Unix() - timestamp
	if difference < 0 {
		difference = -difference
	}
	if difference > int64(tolerance/time.Second) {
		return false
	}
	digest, err := hex.DecodeString(parts[2])
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(parts[1] + "."))
	_, _ = mac.Write(body)
	return hmac.Equal(mac.Sum(nil), digest)
}
