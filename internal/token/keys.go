package token

import (
	"crypto/rsa"
	"embed"

	"github.com/golang-jwt/jwt/v5"
)

//go:embed jwt_public.pem
var publicKeyFS embed.FS

func LoadPublicKeyFromFile() (*rsa.PublicKey, error) {
	keyData, err := publicKeyFS.ReadFile("jwt_public.pem")
	if err != nil {
		return nil, err
	}

	return jwt.ParseRSAPublicKeyFromPEM(keyData)
}
