// Package repository ...
package repository

import (
	"auth/internal/domain"
	"context"
	"errors"
	"time"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

// SessionRepository ...
type SessionRepository interface {
	SessionByID(ctx context.Context, id int) (domain.Session, error)
	CreateSession(ctx context.Context, userID int, appID int, refreshToken string, refExpiresAt time.Time) (sessionID int, err error)
	RevokeByRefreshToken(ctx context.Context, refreshToken string) (revoked bool, err error)
	SessionByRefreshToken(ctx context.Context, refreshToken string) (session domain.Session, err error)
}
