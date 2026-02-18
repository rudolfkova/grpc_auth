// Package rediscache ...
package rediscache

import (
	"auth/internal/config"
	"auth/internal/domain"
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
func (s *Store) SetSession(keyID int, value domain.Session) error {
	_ = keyID
	_ = value
	return nil
}

// GetSession ...
func (s *Store) GetSession(keyID int) (value domain.Session, err error) {
	_ = keyID
	return domain.Session{}, nil
}

// DelSession ...
func (s *Store) DelSession(keyID int) error {
	_ = keyID
	return nil
}
