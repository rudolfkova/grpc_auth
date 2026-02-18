package repository

import "auth/internal/domain"

// Cache ...
type Cache interface {
	SetSession(keyID int, value domain.Session) error
	GetSession(keyID int) (value domain.Session, err error)
	DelSession(keyID int) error
}
