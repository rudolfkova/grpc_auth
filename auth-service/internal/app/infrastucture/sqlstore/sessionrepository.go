// Package sqlstore ...
package sqlstore

import (
	"auth/internal/app/domain"
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
	return domain.Session{}, nil
}
