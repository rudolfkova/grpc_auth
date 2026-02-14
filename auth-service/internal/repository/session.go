// Package repository ...
package repository

import (
	"auth/internal/domain"
	"context"
	"time"
)

// SessionRepository ...
type SessionRepository interface {
	AppByID(ctx context.Context, id int) (domain.App, error)
	CreateSession(ctx context.Context, userID int, appID int, refreshToken string, refExpiresAt time.Time) (sessionID int, err error)
	RevokeByRefreshToken(ctx context.Context, refreshToken string) (revoked bool, err error)
	SessionByRefreshToken(ctx context.Context, refreshToken string) (session domain.App, err error)
}
