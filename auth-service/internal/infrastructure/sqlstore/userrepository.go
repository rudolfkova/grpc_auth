// Package sqlstore ...
package sqlstore

import (
	"auth/internal/domain"
	"auth/internal/repository"
	"context"
	"database/sql"
	"errors"
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

	q := `INSERT INTO users (email, password_hash) VALUES ($1, $2)`

	_, err := r.db.ExecContext(ctx, q, email, string(passHash))
	if err != nil {
		// тут можешь распарсить pg error и вернуть repository.ErrUserAlreadyExists,
		// если ошибка про unique constraint по email
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// UserByEmail ...
func (r *UserRepository) UserByEmail(ctx context.Context, email string) (domain.User, error) {
	const op = "UserRepository.UserByEmail"

	q := `SELECT id, email, password_hash FROM users WHERE email = $1`

	var u domain.User
	var passHash string

	err := r.db.QueryRowContext(ctx, q, email).Scan(
		&u.ID,
		&u.Email,
		&passHash,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}

		return domain.User{}, fmt.Errorf("%s: %w", op, err)
	}

	u.PassHash = []byte(passHash)

	return u, nil
}

// IsAdmin ...
func (r *UserRepository) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "UserRepository.IsAdmin"

	q := `SELECT is_admin FROM users WHERE id = $1`

	var isAdmin bool

	err := r.db.QueryRowContext(ctx, q, userID).Scan(&isAdmin)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return false, fmt.Errorf("%s: %w", op, repository.ErrUserNotFound)
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}
