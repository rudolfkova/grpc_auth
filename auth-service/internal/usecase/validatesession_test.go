package usecase_test

import (
	"auth/internal/config"
	"auth/internal/domain"
	"auth/internal/usecase"
	providerMocks "auth/mocks/provider"
	repoMocks "auth/mocks/repository"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestAuthUseCase_ValidateSession_CacheHitActive ...
func TestAuthUseCase_ValidateSession_CacheHitActive(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 100

	session := domain.Session{
		ID:               sessionID,
		UserID:           42,
		AppID:            1,
		RefreshExpiresAt: time.Now().Add(time.Hour),
		Status:           "active",
	}

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(true, session, nil)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.NoError(t, err)
	assert.True(t, active)
}

// TestAuthUseCase_ValidateSession_CacheHitInactive ...
func TestAuthUseCase_ValidateSession_CacheHitInactive(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 101

	session := domain.Session{
		ID:               sessionID,
		UserID:           42,
		AppID:            1,
		RefreshExpiresAt: time.Now().Add(time.Hour),
		Status:           "revoked",
	}

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(true, session, nil)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.NoError(t, err)
	assert.False(t, active)
}

// TestAuthUseCase_ValidateSession_CacheMissDBActive ...
func TestAuthUseCase_ValidateSession_CacheMissDBActive(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 102

	session := domain.Session{
		ID:               sessionID,
		UserID:           42,
		AppID:            1,
		RefreshExpiresAt: time.Now().Add(time.Hour),
		Status:           "active",
	}

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(false, domain.Session{}, nil)

	sessRepo.
		On("SessionByID", ctx, sessionID).
		Return(session, nil)

	cacheRepo.
		On("SetSession", ctx, sessionID, session).
		Return(nil)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.NoError(t, err)
	assert.True(t, active)

	cacheRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
}

// TestAuthUseCase_ValidateSession_CacheMissDBInactive ...
func TestAuthUseCase_ValidateSession_CacheMissDBInactive(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 103

	session := domain.Session{
		ID:               sessionID,
		UserID:           42,
		AppID:            1,
		RefreshExpiresAt: time.Now().Add(-time.Hour),
		Status:           "active",
	}

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(false, domain.Session{}, nil)

	sessRepo.
		On("SessionByID", ctx, sessionID).
		Return(session, nil)

	cacheRepo.
		On("SetSession", ctx, sessionID, session).
		Return(nil)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.NoError(t, err)
	assert.False(t, active)

	cacheRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
}

// TestAuthUseCase_ValidateSession_CacheErrorFallbackToDB ...
func TestAuthUseCase_ValidateSession_CacheErrorFallbackToDB(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 104

	session := domain.Session{
		ID:               sessionID,
		UserID:           42,
		AppID:            1,
		RefreshExpiresAt: time.Now().Add(time.Hour),
		Status:           "active",
	}

	cacheErr := fmt.Errorf("redis down")

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(false, domain.Session{}, cacheErr)

	sessRepo.
		On("SessionByID", ctx, sessionID).
		Return(session, nil)

	cacheRepo.
		On("SetSession", ctx, sessionID, session).
		Return(nil)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.NoError(t, err)
	assert.True(t, active)

	cacheRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
}

// TestAuthUseCase_ValidateSession_DBError ...
func TestAuthUseCase_ValidateSession_DBError(t *testing.T) {
	const op = "Auth.ValidateSession"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	cacheRepo := new(repoMocks.Cache)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		cacheRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	ctx := context.Background()
	sessionID := 105

	cacheRepo.
		On("GetSession", ctx, sessionID).
		Return(false, domain.Session{}, nil)

	sessRepo.
		On("SessionByID", ctx, sessionID).
		Return(domain.Session{}, errFailed)

	active, err := uc.ValidateSession(ctx, sessionID)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.False(t, active)

	sessRepo.AssertExpectations(t)
}
