package repository

import (
	"auth/internal/app/domain"
	"context"
)

// UserRepository ...
type UserRepository interface {
	SaveUser(ctx context.Context, email string, passHash []byte)
	UserByEmail(ctx context.Context, email string) (domain.User, error)
	IsAdmin(ctx context.Context, userID int) (bool, error)
}
