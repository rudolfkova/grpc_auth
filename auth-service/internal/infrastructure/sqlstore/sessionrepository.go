// Package sqlstore ...
package sqlstore

import (
	"auth/internal/domain"
	"context"
	"database/sql"
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

// SessionByID ...
func (r *SessionRepository) SessionByID(ctx context.Context, id int) (domain.Session, error) {
	_ = ctx
	_ = id
	return domain.Session{}, nil
}
