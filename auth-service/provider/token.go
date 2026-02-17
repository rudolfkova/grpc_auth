// Package provider ...
package provider

import (
	"errors"
	"time"
)

var (
	// ErrInvalidRefreshToken ...
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// TokenProvider ...
type TokenProvider interface {
	CreateAccessToken(userID int, sessionID int, appID int, exp time.Time) (accToken string, err error)
	CreateRefreshToken() (refToken string, err error)
}
