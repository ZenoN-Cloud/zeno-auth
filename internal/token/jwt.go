package token

import (
	"context"
	"crypto/rsa"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Claims struct {
	UserID             uuid.UUID `json:"user_id"`
	OrgID              uuid.UUID `json:"org_id"`
	Roles              []string  `json:"roles"`
	OrgStatus          string    `json:"org_status"`
	SubscriptionStatus string    `json:"subscription_status,omitempty"`
	TrialEndsAt        *int64    `json:"trial_ends_at,omitempty"`
	jwt.RegisteredClaims
}

type JWTManager struct {
	privateKey *rsa.PrivateKey
	publicKey  *rsa.PublicKey
}

func NewJWTManager(privateKeyPEM, publicKeyPEM string) (*JWTManager, error) {
	privateKey, err := jwt.ParseRSAPrivateKeyFromPEM([]byte(privateKeyPEM))
	if err != nil {
		return nil, err
	}

	var publicKey *rsa.PublicKey
	if publicKeyPEM != "" {
		publicKey, err = jwt.ParseRSAPublicKeyFromPEM([]byte(publicKeyPEM))
		if err != nil {
			return nil, err
		}
	} else {
		// Extract public key from private key
		publicKey = &privateKey.PublicKey
	}

	return &JWTManager{
		privateKey: privateKey,
		publicKey:  publicKey,
	}, nil
}

func (j *JWTManager) Generate(ctx context.Context, userID, orgID uuid.UUID, roles []string, ttlSeconds int) (string, error) {
	return j.GenerateWithOrgStatus(ctx, userID, orgID, roles, "created", "", nil, ttlSeconds)
}

func (j *JWTManager) GenerateWithOrgStatus(ctx context.Context, userID, orgID uuid.UUID, roles []string, orgStatus, subStatus string, trialEndsAt *int64, ttlSeconds int) (string, error) {
	select {
	case <-ctx.Done():
		return "", ctx.Err()
	default:
	}

	now := time.Now()
	jti := uuid.New().String() // Unique token ID for revocation

	claims := Claims{
		UserID:             userID,
		OrgID:              orgID,
		Roles:              roles,
		OrgStatus:          orgStatus,
		SubscriptionStatus: subStatus,
		TrialEndsAt:        trialEndsAt,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID.String(),
			ID:        jti,
			ExpiresAt: jwt.NewNumericDate(now.Add(time.Duration(ttlSeconds) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "zeno-auth",
			Audience:  []string{"zeno-frontend", "zeno-api"},
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	token.Header["kid"] = "2024-01" // Key ID for key rotation
	return token.SignedString(j.privateKey)
}

func (j *JWTManager) Validate(ctx context.Context, tokenString string) (*Claims, error) {
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	default:
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return j.publicKey, nil
	}, jwt.WithValidMethods([]string{"RS256"}))

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		// Validate issuer and audience
		if claims.Issuer != "zeno-auth" {
			return nil, jwt.ErrTokenInvalidIssuer
		}
		validAudience := false
		for _, aud := range claims.Audience {
			if aud == "zeno-frontend" || aud == "zeno-api" {
				validAudience = true
				break
			}
		}
		if !validAudience {
			return nil, jwt.ErrTokenInvalidAudience
		}
		return claims, nil
	}

	return nil, jwt.ErrTokenInvalidClaims
}

func (j *JWTManager) GetPublicKey() *rsa.PublicKey {
	return j.publicKey
}
