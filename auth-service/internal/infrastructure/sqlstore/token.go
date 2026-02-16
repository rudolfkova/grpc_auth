// Package sqlstore ...
package sqlstore

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	refreshTokenBytes = 32
)

// TokenProvider ...
type TokenProvider struct {
	jwtSecret []byte
}

// NewTokenProvider ...
func NewTokenProvider(jwtSecret []byte) TokenProvider {
	return TokenProvider{
		jwtSecret: jwtSecret,
	}
}

// AccessClaims ...
type AccessClaims struct {
	UserID    int `json:"user_id"`
	SessionID int `json:"session_id"`
	AppID     int `json:"app_id"`
	jwt.RegisteredClaims
}

// CreateAccessToken ...
func (p TokenProvider) CreateAccessToken(userID int, sessionID int, appID int, accExp time.Time) (accToken string, err error) {
	const op = "TokenProvider.CreateAccessToken"

	claims := AccessClaims{
		UserID:    userID,
		SessionID: sessionID,
		AppID:     appID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(accExp),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	accessTokenStr, err := token.SignedString(p.jwtSecret)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return accessTokenStr, nil
}

// CreateRefreshToken ...
func (p TokenProvider) CreateRefreshToken() (refToken string, err error) {
	bytes := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(bytes), nil
}
