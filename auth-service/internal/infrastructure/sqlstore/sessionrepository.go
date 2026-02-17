// Package sqlstore ...
package sqlstore

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"fmt"
	"time"
)

const (
	emptyID = 0
)

// SessionRepository ...
type SessionRepository struct {
	db *sql.DB
}

// NewSessionRepository ...
func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{
		db: db,
	}
}

// AppByID ...
func (r *SessionRepository) AppByID(ctx context.Context, id int) (domain.Session, error) {
	const op = "SessionRepository.AppByID"

	_ = ctx
	_ = id

	return domain.Session{}, fmt.Errorf("not implemented %s", op)
}

// CreateSession ...
func (r *SessionRepository) CreateSession(ctx context.Context, userID int, appID int, refreshToken string, refExpiresAt time.Time) (sessionID int, err error) {
	const op = "SessionRepository.CreateSession"

	_ = ctx
	_ = userID
	_ = appID
	_ = refreshToken
	_ = refExpiresAt

	return emptyID, fmt.Errorf("not implemented %s", op)
}

// RevokeByRefreshToken ...
func (r *SessionRepository) RevokeByRefreshToken(ctx context.Context, refreshToken string) (revoked bool, err error) {
	const op = "SessionRepository.RevokeByRefreshToken"

	_ = ctx
	_ = refreshToken

	return false, fmt.Errorf("not implemented %s", op)
}

// SessionByRefreshToken ...
func (r *SessionRepository) SessionByRefreshToken(ctx context.Context, refreshToken string) (session domain.Session, err error) {
	const op = "SessionRepository.SessionByRefreshToken"

	_ = ctx
	_ = refreshToken

	return domain.Session{}, fmt.Errorf("not implemented %s", op)
}
