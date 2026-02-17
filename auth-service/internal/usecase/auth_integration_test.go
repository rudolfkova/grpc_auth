//go:build integration
// +build integration

package usecase_test

import (
	"auth/internal/config"
	"auth/internal/infrastructure/sqlstore"
	"auth/internal/repository"
	"auth/internal/usecase"
	"auth/provider"
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"
)

var (
	intCfg config.Config
)

// initIntegrationConfig загружает config.toml относительно расположения этого файла.
func init() {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		panic("cannot get caller info for auth_integration_test.go")
	}

	configPath := filepath.Join(
		filepath.Dir(filename),
		"..", "..", "..",
		"config.toml",
	)

	c := config.NewConfig()
	_, err := toml.DecodeFile(configPath, c)
	if err != nil {
		panic(err)
	}

	intCfg = *c
}

func skipIfNoIntegrationDSN(t *testing.T) {
	t.Helper()

	if strings.TrimSpace(intCfg.TestDatabaseURL) == "" {
		t.Skip("test_database_url is empty in config.toml; skipping integration tests")
	}
}

// testDBIntegration открывает подключение к БД и возвращает функцию очистки.
func testDBIntegration(t *testing.T, databaseURL string) (*sql.DB, func(...string)) {
	t.Helper()

	db, err := sql.Open("postgres", databaseURL)
	require.NoError(t, err)

	err = db.Ping()
	require.NoError(t, err)

	cleanup := func(tables ...string) {
		if len(tables) > 0 {
			query := "TRUNCATE " + strings.Join(tables, ", ") + " CASCADE"
			_, err := db.Exec(query)
			require.NoError(t, err)
		}

		_ = db.Close()
	}

	return db, cleanup
}

func newIntegrationUseCase(db *sql.DB) *usecase.AuthUseCase {
	userRepo := sqlstore.NewUserRepository(db)
	sessRepo := sqlstore.NewSessionRepository(db)
	tokenProv := sqlstore.NewTokenProvider([]byte(intCfg.JWTSecret))

	logger := config.NewLogger(&intCfg)

	return usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		intCfg.AccessTokenTTL,
		intCfg.RefreshTokenTTL,
	)
}

func makeTestEmail(prefix string) string {
	// чтобы тесты не дрались, если TRUNCATE вдруг не сработал
	return prefix + "-" + time.Now().Format("20060102-150405.000000") + "@example.com"
}

func expireSessionByRefreshToken(t *testing.T, db *sql.DB, refreshToken string, exp time.Time) {
	t.Helper()

	_, err := db.Exec(`UPDATE sessions SET refresh_expires_at = $1 WHERE refresh_token = $2`, exp, refreshToken)
	require.NoError(t, err)
}

func setAdminByUserID(t *testing.T, db *sql.DB, userID int, isAdmin bool) {
	t.Helper()

	_, err := db.Exec(`UPDATE users SET is_admin = $1 WHERE id = $2`, isAdmin, userID)
	require.NoError(t, err)
}

func TestAuthUseCase_Register_Login_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-login")
	password := "password"
	appID := 1

	// Register + Login через реальный sqlstore и БД.
	userID, err := uc.Register(ctx, email, password)
	require.NoError(t, err)
	require.NotZero(t, userID)

	token, err := uc.Login(ctx, email, password, appID)
	require.NoError(t, err)
	require.NotEmpty(t, token.AccessToken)
	require.NotEmpty(t, token.RefreshToken)
}

func TestAuthUseCase_Register_DuplicateEmail_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-dup")
	password := "password"

	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	_, err = uc.Register(ctx, email, password)
	require.Error(t, err)
}

func TestAuthUseCase_Login_InvalidCredentials_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-badpass")
	password := "password"
	appID := 1

	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	_, err = uc.Login(ctx, email, "wrong_password", appID)
	require.ErrorIs(t, err, repository.ErrInvalidCredentials)

	_, err = uc.Login(ctx, "no_such_user@example.com", password, appID)
	require.ErrorIs(t, err, repository.ErrInvalidCredentials)
}

func TestAuthUseCase_IsAdmin_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-admin")
	password := "password"

	userID, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	isAdmin, err := uc.IsAdmin(ctx, userID)
	require.NoError(t, err)
	require.False(t, isAdmin)

	setAdminByUserID(t, db, userID, true)

	isAdmin, err = uc.IsAdmin(ctx, userID)
	require.NoError(t, err)
	require.True(t, isAdmin)
}

func TestAuthUseCase_Logout_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()

	// нет такой сессии — ок=false, err=nil
	ok, err := uc.Logout(ctx, "no_such_refresh_token")
	require.NoError(t, err)
	require.False(t, ok)
}

func TestAuthUseCase_RefreshToken_Rotation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-refresh")
	password := "password"
	appID := 1

	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	tok1, err := uc.Login(ctx, email, password, appID)
	require.NoError(t, err)
	require.NotEmpty(t, tok1.RefreshToken)

	tok2, err := uc.RefreshToken(ctx, tok1.RefreshToken)
	require.NoError(t, err)
	require.NotEmpty(t, tok2.AccessToken)
	require.NotEmpty(t, tok2.RefreshToken)
	require.NotEqual(t, tok1.RefreshToken, tok2.RefreshToken)
}

func TestAuthUseCase_RefreshToken_RevokedOrExpired_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-revoked")
	password := "password"
	appID := 1

	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	tok, err := uc.Login(ctx, email, password, appID)
	require.NoError(t, err)

	// отзываем refresh через Logout и убеждаемся, что RefreshToken теперь не работает
	ok, err := uc.Logout(ctx, tok.RefreshToken)
	require.NoError(t, err)
	require.True(t, ok)

	_, err = uc.RefreshToken(ctx, tok.RefreshToken)
	require.ErrorIs(t, err, provider.ErrInvalidRefreshToken)

	// кейс "нет такой сессии"
	_, err = uc.RefreshToken(ctx, "no_such_refresh_token")
	require.Error(t, err)
	require.True(t, errors.Is(err, repository.ErrSessionNotFound))
}

func TestAuthUseCase_RefreshToken_ExpiredByDBMutation_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	skipIfNoIntegrationDSN(t)

	db, teardown := testDBIntegration(t, intCfg.TestDatabaseURL)
	defer teardown("users", "sessions")

	uc := newIntegrationUseCase(db)

	ctx := context.Background()
	email := makeTestEmail("integration-expired")
	password := "password"
	appID := 1

	_, err := uc.Register(ctx, email, password)
	require.NoError(t, err)

	tok, err := uc.Login(ctx, email, password, appID)
	require.NoError(t, err)

	expireSessionByRefreshToken(t, db, tok.RefreshToken, time.Now().Add(-time.Hour))

	_, err = uc.RefreshToken(ctx, tok.RefreshToken)
	require.ErrorIs(t, err, provider.ErrInvalidRefreshToken)
}

