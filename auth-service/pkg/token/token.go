// Package tokenjwt ...
package tokenjwt

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Token ...
type Token struct {
	AccessToken     string
	RefreshToken    string
	AccessExpireAt  time.Time
	RefreshExpireAt time.Time
}

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

// UserAccessDate ...
type UserAccessDate struct {
	UserID    int
	SessionID int
	AppID     int
	AccessExp float64
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

// DecodJWT ...
func (p TokenProvider) DecodJWT(accToken string) (claims *UserAccessDate, err error) {
	token, err := jwt.Parse(accToken, func(token *jwt.Token) (interface{}, error) {
		return []byte(p.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	claimMap := token.Claims.(jwt.MapClaims)
	userid := claimMap["user_id"].(float64)
	exp := claimMap["exp"].(float64)
	sessionid := claimMap["session_id"].(float64)
	appid := claimMap["app_id"].(float64)

	return &UserAccessDate{
		UserID:    int(userid),
		SessionID: int(sessionid),
		AppID:     int(appid),
		AccessExp: exp,
	}, nil
}
