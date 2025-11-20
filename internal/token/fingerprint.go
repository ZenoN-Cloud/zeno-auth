package token

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

type Fingerprint struct {
	UserAgent      string
	IPAddress      string
	AcceptLanguage string
}

func GenerateFingerprint(userAgent, ipAddress, acceptLanguage string) string {
	ipParts := strings.Split(ipAddress, ".")
	if len(ipParts) >= 3 {
		ipAddress = strings.Join(ipParts[:3], ".") + ".0"
	}

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
