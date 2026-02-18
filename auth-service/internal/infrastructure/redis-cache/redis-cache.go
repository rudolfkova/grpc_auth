// Package rediscache ...
package rediscache

import (
	"auth/internal/config"
	"auth/internal/domain"
	"context"
)

// Store ...
type Store struct{}

// NewCacheStore ...
func NewCacheStore(cfg config.Config) (*Store, error) {
	_ = cfg
	return &Store{}, nil
}

// Close ...
func (s *Store) Close() error {
	return nil
}

// SetSession ...
func (s *Store) SetSession(ctx context.Context, keyID int, value domain.Session) error {
	_ = ctx
	_ = keyID
	_ = value
	return nil
}

// GetSession ...
func (s *Store) GetSession(ctx context.Context, keyID int) (ok bool, value domain.Session, err error) {
	_ = ctx
	_ = keyID
	return false, domain.Session{}, nil
}

// DelSession ...
func (s *Store) DelSession(ctx context.Context, keyID int) error {
	_ = ctx
	_ = keyID
	return nil
}
