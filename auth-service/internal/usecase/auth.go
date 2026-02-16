// Package usecase ...
package usecase

import (
	"auth/internal/domain"
	"auth/internal/repository"
	"auth/provider"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const (
	emptyID = 0
)

var (
	// ErrInvalidCredentials ...
	ErrInvalidCredentials = errors.New("invalid credentials")
	// ErrInvalidRefreshToken ...
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
)

// AuthUseCase ...
type AuthUseCase struct {
	Users    repository.UserRepository
	Sessions repository.SessionRepository
	Token    provider.TokenProvider

	Logger slog.Logger

	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

// NewAuthUseCase ...
func NewAuthUseCase(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	token provider.TokenProvider,
	logger slog.Logger,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration) *AuthUseCase {
	return &AuthUseCase{
		Users:           users,
		Sessions:        sessions,
		Token:           token,
		Logger:          logger,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
	}
}

// Register ...
func (a *AuthUseCase) Register(ctx context.Context, email string, password string) (userID int, err error) {
	const op = "Auth.Register"

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.Users.SaveUser(ctx, email, passHash); err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.Users.UserByEmail(ctx, email)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	return user.ID, nil
}

// Login ...
func (a *AuthUseCase) Login(ctx context.Context, email string, password string, appID int) (token domain.Token, err error) {
	const op = "Auth.Login"

	log := a.Logger.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.Users.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			a.Logger.Warn("user not found")

			return domain.Token{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
		}

		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.Logger.Info("invalid credentials")

		return domain.Token{}, fmt.Errorf("%s: %w", op, ErrInvalidCredentials)
	}

	refreshToken, err := a.Token.CreateRefreshToken()
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	refExp := time.Now().Add(a.RefreshTokenTTL)

	sessionID, err := a.Sessions.CreateSession(ctx, int(user.ID), int(appID), refreshToken, refExp)
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	accExp := time.Now().Add(a.AccessTokenTTL)

	accessToken, err := a.Token.CreateAccessToken(int(user.ID), sessionID, int(appID), accExp)
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Token{
		AccessToken:     accessToken,
		AccessExpireAt:  accExp,
		RefreshToken:    refreshToken,
		RefreshExpireAt: refExp,
	}, nil

}

// IsAdmin ...
func (a *AuthUseCase) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "Auth.IsAdmin"

	isAdmin, err := a.Users.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// Logout ...
func (a *AuthUseCase) Logout(ctx context.Context, refreshToken string) (success bool, err error) {
	const op = "Auth.Logout"

	ok, err := a.Sessions.RevokeByRefreshToken(ctx, refreshToken)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return ok, nil
}

// RefreshToken ...
func (a *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (token domain.Token, err error) {
	const op = "Auth.RefreshToken"

	session, err := a.Sessions.SessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	if session.Status != "active" || time.Now().After(session.RefreshExpiresAt) {
		return domain.Token{}, fmt.Errorf("%s: %w", op, ErrInvalidRefreshToken)
	}

	_, _ = a.Sessions.RevokeByRefreshToken(ctx, refreshToken)

	newRefreshToken, err := a.Token.CreateRefreshToken()
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}
	refExp := time.Now().Add(a.RefreshTokenTTL)

	sessionID, err := a.Sessions.CreateSession(ctx, session.UserID, session.AppID, newRefreshToken, refExp)
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	accExp := time.Now().Add(a.AccessTokenTTL)
	accessToken, err := a.Token.CreateAccessToken(session.UserID, sessionID, session.AppID, accExp)
	if err != nil {
		return domain.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return domain.Token{
		AccessToken:     accessToken,
		AccessExpireAt:  accExp,
		RefreshToken:    newRefreshToken,
		RefreshExpireAt: refExp,
	}, nil
}
