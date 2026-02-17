package repository

import (
	"auth/internal/domain"
	"context"
	"errors"
)

var (
	// ErrUserNotFound ...
	ErrUserNotFound = errors.New("user not found")
	// ErrUserAlreadyExists ...
	ErrUserAlreadyExists = errors.New("this email already exists in user store")
	// ErrInvalidCredentials ...
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// UserRepository ...
type UserRepository interface {
	SaveUser(ctx context.Context, email string, passHash []byte) error
	UserByEmail(ctx context.Context, email string) (domain.User, error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
}
