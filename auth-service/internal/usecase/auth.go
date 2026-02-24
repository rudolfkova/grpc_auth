// Package usecase ...
package usecase

import (
	"auth/internal/domain"
	"auth/internal/repository"
	tokenjwt "auth/pkg/token"
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

// AuthUseCase ...
type AuthUseCase struct {
	users    repository.UserRepository
	sessions repository.SessionRepository
	cache    repository.Cache
	token    provider.TokenProvider

	logger slog.Logger

	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

// NewAuthUseCase ...
func NewAuthUseCase(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	cache repository.Cache,
	token provider.TokenProvider,
	logger slog.Logger,
	accessTokenTTL time.Duration,
	refreshTokenTTL time.Duration) *AuthUseCase {
	return &AuthUseCase{
		users:           users,
		sessions:        sessions,
		cache:           cache,
		token:           token,
		logger:          logger,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
}

// Register ...
func (a *AuthUseCase) Register(ctx context.Context, email string, password string) (userID int, err error) {
	const op = "Auth.Register"

	log := a.logger.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("register user")

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.users.SaveUser(ctx, email, passHash); err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	user, err := a.users.UserByEmail(ctx, email)
	if err != nil {
		return emptyID, fmt.Errorf("%s: %w", op, err)
	}

	log.Info("register success")

	return user.ID, nil
}

// Login ...
func (a *AuthUseCase) Login(ctx context.Context, email string, password string, appID int) (token tokenjwt.Token, err error) {
	const op = "Auth.Login"

	log := a.logger.With(
		slog.String("op", op),
		slog.String("username", email),
	)

	log.Info("attempting to login user")

	user, err := a.users.UserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			a.logger.Warn("user not found")

			return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, repository.ErrInvalidCredentials)
		}

		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		a.logger.Info("invalid credentials")

		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, repository.ErrInvalidCredentials)
	}

	refreshToken, err := a.token.CreateRefreshToken()
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	refExp := time.Now().Add(a.refreshTokenTTL)

	sessionID, err := a.sessions.CreateSession(ctx, int(user.ID), int(appID), refreshToken, refExp)
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	accExp := time.Now().Add(a.accessTokenTTL)

	accessToken, err := a.token.CreateAccessToken(int(user.ID), sessionID, int(appID), accExp)
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokenjwt.Token{
		AccessToken:     accessToken,
		AccessExpireAt:  accExp,
		RefreshToken:    refreshToken,
		RefreshExpireAt: refExp,
	}, nil

}

// IsAdmin ...
func (a *AuthUseCase) IsAdmin(ctx context.Context, userID int) (bool, error) {
	const op = "Auth.IsAdmin"

	log := a.logger.With(
		slog.String("op", op),
		slog.String("userID", fmt.Sprint(userID)),
	)

	log.Info("check permisions")

	isAdmin, err := a.users.IsAdmin(ctx, userID)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return isAdmin, nil
}

// Logout ...
func (a *AuthUseCase) Logout(ctx context.Context, refreshToken string) (success bool, err error) {
	const op = "Auth.Logout"

	log := a.logger.With(
		slog.String("op", op),
	)

	log.Info("logout user by refreshToken")

	session, err := a.sessions.SessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		if !errors.Is(err, repository.ErrSessionNotFound) {
			return false, nil
		}

		return false, fmt.Errorf("%s: %w", op, err)
	}

	if err := a.cache.DelSession(ctx, session.ID); err != nil {
		log.Warn("session not deleted from cache")
	}

	ok, err := a.sessions.RevokeByRefreshToken(ctx, refreshToken)
	if err != nil {
		return false, fmt.Errorf("%s: %w", op, err)
	}

	return ok, nil
}

// RefreshToken ...
func (a *AuthUseCase) RefreshToken(ctx context.Context, refreshToken string) (token tokenjwt.Token, err error) {
	const op = "Auth.RefreshToken"

	log := a.logger.With(
		slog.String("op", op),
	)

	log.Info("refresh token by refreshToken")

	session, err := a.sessions.SessionByRefreshToken(ctx, refreshToken)
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	if session.Status != "active" || time.Now().After(session.RefreshExpiresAt) {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, provider.ErrInvalidRefreshToken)
	}

	_, _ = a.sessions.RevokeByRefreshToken(ctx, refreshToken)
	if err := a.cache.DelSession(ctx, session.ID); err != nil {
		log.Warn("session not deleted from cache")
	}

	newRefreshToken, err := a.token.CreateRefreshToken()
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}
	refExp := time.Now().Add(a.refreshTokenTTL)

	sessionID, err := a.sessions.CreateSession(ctx, session.UserID, session.AppID, newRefreshToken, refExp)
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	accExp := time.Now().Add(a.accessTokenTTL)
	accessToken, err := a.token.CreateAccessToken(session.UserID, sessionID, session.AppID, accExp)
	if err != nil {
		return tokenjwt.Token{}, fmt.Errorf("%s: %w", op, err)
	}

	return tokenjwt.Token{
		AccessToken:     accessToken,
		AccessExpireAt:  accExp,
		RefreshToken:    newRefreshToken,
		RefreshExpireAt: refExp,
	}, nil
}

// ValidateSession ...
func (a *AuthUseCase) ValidateSession(ctx context.Context, sessionID int) (active bool, err error) {
	const op = "Auth.ValidateSession"

	log := a.logger.With(
		slog.String("op", op),
	)

	log.Info("validate session")

	ok, session, err := a.cache.GetSession(ctx, sessionID)
	if err != nil {
		log.Warn("session not get from cache")
		ok = false
	}

	if !ok {
		session, err := a.sessions.SessionByID(ctx, sessionID)
		if err != nil {
			return false, fmt.Errorf("%s: %w", op, err)
		}
		if err = a.cache.SetSession(ctx, sessionID, session); err != nil {
			log.Warn("session not set in cache")
		}
		if isSessionActive(session) {
			return true, nil
		}
		return false, nil

	}

	log.Info("validate from cache")
	if isSessionActive(session) {
		return true, nil
	}
	return false, nil

}

func isSessionActive(s domain.Session) bool {
	return s.Status == "active" && time.Now().Before(s.RefreshExpiresAt)
}
