package utils

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type TokenClaims struct {
	UserID      string `json:"userId,omitempty"`
	OrgID       string `json:"orgId,omitempty"`
	Role        string `json:"role,omitempty"`
	Setup       bool   `json:"setup,omitempty"`
	ProvisionID string `json:"provisionId,omitempty"`
	SiteSlug    string `json:"siteSlug,omitempty"`
	SiteID      string `json:"siteId,omitempty"`
	jwt.RegisteredClaims
}

func CreateToken(userID, orgID, role, secret string, ttl time.Duration) (string, error) {
	claims := TokenClaims{
		UserID: userID,
		OrgID:  orgID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(ttl)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func ParseToken(tokenString, secret string) (*TokenClaims, error) {
	parsed, err := jwt.ParseWithClaims(tokenString, &TokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		if token.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := parsed.Claims.(*TokenClaims)
	if !ok || !parsed.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

func CreateSetupToken(provisionID, siteSlug, siteID, secret string, ttl time.Duration) (string, time.Time, error) {
	expiresAt := time.Now().Add(ttl)
	claims := TokenClaims{
		Setup:       true,
		ProvisionID: provisionID,
		SiteSlug:    siteSlug,
		SiteID:      siteID,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        primitive.NewObjectID().Hex(),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}
