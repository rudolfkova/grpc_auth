package sqlstore

import (
	"auth/internal/domain"
	"auth/internal/repository"
	"context"
	"database/sql"
	"errors"
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
	return &SessionRepository{db: db}
}

// SessionByID ...
func (r *SessionRepository) SessionByID(ctx context.Context, id int) (domain.Session, error) {
	const op = "SessionRepository.SessionByID"

	q := `SELECT id, user_id, app_id, refresh_expires_at, status
	      FROM sessions
	      WHERE id = $1`

	var s domain.Session

	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&s.ID,
		&s.UserID,
		&s.AppID,
		&s.RefreshExpiresAt,
		&s.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, fmt.Errorf("%s: %w", op, repository.ErrSessionNotFound)
		}

		return domain.Session{}, fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}

// CreateSession ...
func (r *SessionRepository) CreateSession(
	ctx context.Context,
	userID int,
	appID int,
	refreshToken string,
	refExpiresAt time.Time,
) (sessionID int, err error) {
	const op = "SessionRepository.CreateSession"

	q := `INSERT INTO sessions (user_id, app_id, refresh_token, refresh_expires_at)
	      VALUES ($1, $2, $3, $4)
	      RETURNING id`

	err = r.db.QueryRowContext(ctx, q,
		userID,
		appID,
		refreshToken,
		refExpiresAt,
	).Scan(&sessionID)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return sessionID, nil
}

// RevokeByRefreshToken ...
func (r *SessionRepository) RevokeByRefreshToken(ctx context.Context, refreshToken string) (revoked bool, err error) {
	const op = "SessionRepository.RevokeByRefreshToken"

	q := `UPDATE sessions
	      SET status = 'revoked', updated_at = now()
	      WHERE refresh_token = $1 AND status = 'active'`

	res, err := r.db.ExecContext(ctx, q, refreshToken)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return rows > 0, nil
}

// SessionByRefreshToken ...
func (r *SessionRepository) SessionByRefreshToken(ctx context.Context, refreshToken string) (session domain.Session, err error) {
	const op = "SessionRepository.SessionByRefreshToken"

	q := `SELECT id, user_id, app_id, refresh_expires_at, status
	      FROM sessions
	      WHERE refresh_token = $1`

	var s domain.Session

	err = r.db.QueryRowContext(ctx, q, refreshToken).Scan(
		&s.ID,
		&s.UserID,
		&s.AppID,
		&s.RefreshExpiresAt,
		&s.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Session{}, fmt.Errorf("%s: %w", op, repository.ErrSessionNotFound)
		}

		return domain.Session{}, fmt.Errorf("%s: %w", op, err)
	}

	return s, nil
}
