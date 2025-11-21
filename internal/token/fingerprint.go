package token

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
)

type Fingerprint struct {
	UserAgent      string
	IPAddress      string
	AcceptLanguage string
}

func GenerateFingerprint(userAgent, ipAddress, acceptLanguage string) string {
	// Use full IP address for better security
	// Previous implementation used only first 3 octets which was too weak
	data := fmt.Sprintf("%s|%s|%s", userAgent, ipAddress, acceptLanguage)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:])
}

func ParseUserAgent(userAgent string) string {
	if len(userAgent) > 200 {
		return userAgent[:200]
	}
	return userAgent
}
