package repository

import (
	"auth/internal/domain"
	"context"
)

// Cache ...
type Cache interface {
	SetSession(ctx context.Context, keyID int, value domain.Session) error
	GetSession(ctx context.Context, keyID int) (ok bool, value domain.Session, err error)
	DelSession(ctx context.Context, keyID int) error
}
