// Package sqlstore ...
package sqlstore

import (
	"auth/internal/domain"
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
func (r *UserRepository) SaveUser(ctx context.Context, email string, passHash []byte) {
	_ = ctx
	_ = email
	_ = passHash
}

// UserByEmail ...
func (r *UserRepository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	_ = ctx
	_ = email
	return domain.User{}, nil
}

// IsAdmin ...
func (r *UserRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	_ = ctx
	_ = userID
	return false, nil
}
