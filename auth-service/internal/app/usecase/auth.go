// Package usecase ...
package usecase

import "auth/internal/app/repository"

// AuthUseCase ...
type AuthUseCase struct {
	users    repository.UserRepository
	sessions repository.SessionRepository
}

// NewAuthUseCase ...
func NewAuthUseCase(users repository.UserRepository, sessions repository.SessionRepository) *AuthUseCase {
	return &AuthUseCase{
		users:    users,
		sessions: sessions,
	}
}
