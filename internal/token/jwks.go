package token

import (
	"context"
	"encoding/base64"
	"math/big"
)

type JWK struct {
	Kty string `json:"kty"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	N   string `json:"n"`
	E   string `json:"e"`
}

type JWKS struct {
	Keys []JWK `json:"keys"`
}

func (j *JWTManager) GetJWKS(ctx context.Context) (*JWKS, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	return &JWKS{
		Keys: []JWK{
			{
				Kty: "RSA",
				Use: "sig",
				Kid: "zeno-auth-key",
				N:   j.encodeBase64BigInt(j.publicKey.N),
				E:   j.encodeBase64BigInt(big.NewInt(int64(j.publicKey.E))),
			},
		},
	}, nil
}

func (j *JWTManager) encodeBase64BigInt(n *big.Int) string {
	return base64.RawURLEncoding.EncodeToString(n.Bytes())
}