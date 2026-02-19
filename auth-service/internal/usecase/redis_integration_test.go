//go:build integration
// +build integration

package usecase_test

import (
	"auth/internal/config"
	rediscache "auth/internal/infrastructure/redis-cache"
	"auth/internal/infrastructure/sqlstore"
	"auth/internal/infrastructure/tokengen"
	"auth/internal/usecase"
	"context"
	"database/sql"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

// testDBIntegration и intCfg уже определены в auth_integration_test.go в этом же пакете.

func skipIfNoRedisAddr(t *testing.T) {
	t.Helper()

	if strings.TrimSpace(intCfg.RedisAddr) == "" {
		t.Skip("redis_addr is empty in config.toml; skipping redis integration tests")
	}
}

// testRedisDBIntegration открывает подключение к БД и Redis, собирает полноценный AuthUseCase.
func testRedisDBIntegration(t *testing.T) (*sql.DB, *usecase.AuthUseCase, func()) {
	t.Helper()

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)

	cache, err := rediscache.NewCacheStore(intCfg)
	require.NoError(t, err)

	logger := config.NewLogger(&intCfg)

	uc := usecase.NewAuthUseCase(
		sqlstore.NewUserRepository(db),
		sqlstore.NewSessionRepository(db),
		cache,
		tokengen.NewTokenProvider([]byte(intCfg.JWTSecret)),
		*logger,
		intCfg.AccessTokenTTL,
		intCfg.RefreshTokenTTL,
	)

	cleanup := func() {
		teardown("users", "sessions")
		_ = cache.Close()
	}

	return db, uc, cleanup
}

// TestAuthUseCase_ValidateSession_RedisIntegration проверяет,
// что ValidateSession работает поверх реального Redis и БД:
// 1) активная сессия -> true,
// 2) после Logout та же сессия -> false.
func TestAuthUseCase_ValidateSession_RedisIntegration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)
	skipIfNoRedisAddr(t)

	db, uc, cleanup := testRedisDBIntegration(t)
	defer cleanup()

	ctx := context.Background()

	email := makeTestEmail("redis-validate")
	password := "password"
	appID := 1

	// Регистрируем и логиним пользователя, получаем refresh и sessionID.
	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	tok, err := uc.Login(ctx, email, password, appID)
	require.NoError(t, err)
	require.NotEmpty(t, tok.RefreshToken)

	// Находим session_id по refresh_token в БД.
	var sessionID int
	err = db.QueryRow(`SELECT id FROM sessions WHERE refresh_token = $1`, tok.RefreshToken).Scan(&sessionID)
	require.NoError(t, err)

	// Первый вызов ValidateSession — может идти через БД, должен вернуть true.
	active, err := uc.ValidateSession(ctx, sessionID)
	require.NoError(t, err)
	require.True(t, active)

	// Второй вызов ValidateSession — должен сработать и с кэшем, по поведению остаётся true.
	active, err = uc.ValidateSession(ctx, sessionID)
	require.NoError(t, err)
	require.True(t, active)

	// После Logout та же сессия должна считаться неактивной.
	ok, err := uc.Logout(ctx, tok.RefreshToken)
	require.NoError(t, err)
	require.True(t, ok)

	active, err = uc.ValidateSession(ctx, sessionID)
	require.NoError(t, err)
	require.False(t, active)
}
