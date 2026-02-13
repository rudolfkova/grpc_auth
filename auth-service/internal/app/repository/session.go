// Package repository ...
package repository

import (
	"auth/internal/app/domain"
	"context"
)

// SessionRepository ...
type SessionRepository interface {
	SessionByID(ctx context.Context, id int) (domain.Session, error)
}
