// Package sqlstore ...
package sqlstore

import (
	"auth/internal/app/domain"
	"context"
	"database/sql"
)

// UserRepository ...
type UserRepository struct {
	db *sql.DB
}

// NewUserRepository ...
func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

// SaveUser ...
func (r *UserRepository) SaveUser(ctx context.Context, email string, passHash []byte) {}

// UserByEmail ...
func (r *UserRepository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	return domain.User{}, nil
}

// IsAdmin ...
func (r *UserRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	return false, nil
}
