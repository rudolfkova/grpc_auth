// Package domain ...
package domain

import "time"

// App ...
type App struct {
	ID               int
	UserID           int
	AppID            int
	RefreshExpiresAt time.Time
	Status           string
}
