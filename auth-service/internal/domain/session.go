// Package domain ...
package domain

import "time"

// Session ...
type Session struct {
	ID               int
	UserID           int
	AppID            int
	RefreshExpiresAt time.Time
	Status           string
}
