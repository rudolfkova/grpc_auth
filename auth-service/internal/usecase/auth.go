// Package usecase ...
package usecase

import "auth/internal/repository"

// AuthUseCase ...
type AuthUseCase struct {
	Users    repository.UserRepository
	Sessions repository.SessionRepository
}

// NewAuthUseCase ...
func NewAuthUseCase(users repository.UserRepository, sessions repository.SessionRepository) *AuthUseCase {
	return &AuthUseCase{
		Users:    users,
		Sessions: sessions,
	}
}
