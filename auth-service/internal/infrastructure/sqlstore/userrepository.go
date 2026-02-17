// Package sqlstore ...
package sqlstore

import (
	"auth/internal/domain"
	"context"
	"database/sql"
	"fmt"
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
func (r *UserRepository) SaveUser(ctx context.Context, email string, passHash []byte) error {
	const op = "UserRepository.SaveUser"

	_ = ctx
	_ = email
	_ = passHash

	return fmt.Errorf("not implemented %s", op)
}

// UserByEmail ...
func (r *UserRepository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	const op = "UserRepository.UserByEmail"

	_ = ctx
	_ = email

	return domain.User{}, fmt.Errorf("not implemented %s", op)
}

// IsAdmin ...
func (r *UserRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "UserRepository.IsAdmin"

	_ = ctx
	_ = userID

	return false, fmt.Errorf("not implemented %s", op)
}
