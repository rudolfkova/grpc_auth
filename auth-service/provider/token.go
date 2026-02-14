// Package provider ...
package provider

import "time"

// TokenProvider ...
type TokenProvider interface {
	CreateAccessToken(userID int, sessionID int, appID int, exp time.Time) (accToken string, err error)
	CreateRefreshToken() (refToken string, err error)
}
