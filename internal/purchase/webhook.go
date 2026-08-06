package purchase

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

var ErrInvalidWebhookSignature = errors.New("invalid webhook signature")

func SignWebhook(secret, payload []byte, timestamp time.Time) string {
	seconds := timestamp.UTC().Unix()
	message := fmt.Sprintf("%d.%s", seconds, payload)
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write([]byte(message))
	return fmt.Sprintf("t=%d,v1=%s", seconds, hex.EncodeToString(mac.Sum(nil)))
}

func VerifyWebhookSignature(secret, payload []byte, header string, now time.Time, tolerance time.Duration) error {
	var timestampRaw, signatureRaw string
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		switch key {
		case "t":
			timestampRaw = value
		case "v1":
			signatureRaw = value
		}
	}
	seconds, err := strconv.ParseInt(timestampRaw, 10, 64)
	if err != nil || signatureRaw == "" {
		return ErrInvalidWebhookSignature
	}
	timestamp := time.Unix(seconds, 0).UTC()
	if now.UTC().Sub(timestamp) > tolerance || timestamp.Sub(now.UTC()) > tolerance {
		return ErrInvalidWebhookSignature
	}
	expectedHeader := SignWebhook(secret, payload, timestamp)
	_, expectedRaw, _ := strings.Cut(expectedHeader, "v1=")
	expected, err := hex.DecodeString(expectedRaw)
	if err != nil {
		return ErrInvalidWebhookSignature
	}
	provided, err := hex.DecodeString(signatureRaw)
	if err != nil || !hmac.Equal(expected, provided) {
		return ErrInvalidWebhookSignature
	}
	return nil
}
