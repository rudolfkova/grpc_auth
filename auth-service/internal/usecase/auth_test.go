package usecase_test

import (
	"auth/internal/config"
	"auth/internal/domain"
	"auth/internal/repository"
	"auth/internal/usecase"
	providerMocks "auth/mocks/provider"
	repoMocks "auth/mocks/repository"
	"auth/provider"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

var (
	cfg = config.Config{
		DatabaseURL:     "",
		BindAddr:        ":50051",
		AccessTokenTTL:  time.Minute * 15,
		RefreshTokenTTL: time.Hour * 24 * 7,
		LogLevel:        "DEBUG",
		JWTSecret:       "123",
	}
)

type testUserRequest struct {
	ctx          context.Context
	email        string
	password     string
	hashPass     []byte
	appID        int
	refreshToken string
}

func TestAuthUseCase_Login_Success(t *testing.T) {
	// m := mocks.*
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
		appID:    1,
	}
	user1.hashPass, _ = bcrypt.GenerateFromPassword([]byte(user1.password), bcrypt.DefaultCost)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: 42, Email: user1.email, PassHash: user1.hashPass}, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, 42, user1.appID, "REFRESH", mock.AnythingOfType("time.Time")).
		Return(100, nil)

	tokenProv.
		On("CreateAccessToken", 42, 100, user1.appID, mock.AnythingOfType("time.Time")).
		Return("ACCESS", nil)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.NoError(t, err)
	assert.Equal(t, "ACCESS", tok.AccessToken)
	assert.Equal(t, "REFRESH", tok.RefreshToken)

	userRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_Login_WrongEmail(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "wrong_user@example.com",
		password: "password",
		appID:    1,
	}

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{}, repository.ErrUserNotFound)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.ErrorIs(t, err, repository.ErrInvalidCredentials)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_WrongPassword(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "user@example.com",
		password: "wrong_password",
		appID:    1,
	}

	realHashPass, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: 42, Email: user1.email, PassHash: realHashPass}, nil)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.ErrorIs(t, err, repository.ErrInvalidCredentials)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_UserRepoError(t *testing.T) {
	const op = "Auth.Login"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
		appID:    1,
	}

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{}, errFailed)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Login_RefreshTokenProviderError(t *testing.T) {
	const op = "Auth.Login"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
		appID:    1,
	}
	user1.hashPass, _ = bcrypt.GenerateFromPassword([]byte(user1.password), bcrypt.DefaultCost)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: 42, Email: user1.email, PassHash: user1.hashPass}, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("", errFailed)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_Login_SessionRepoError(t *testing.T) {
	const op = "Auth.Login"
	errFailed := fmt.Errorf("failed")
	emptyID := 0

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
		appID:    1,
	}
	user1.hashPass, _ = bcrypt.GenerateFromPassword([]byte(user1.password), bcrypt.DefaultCost)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: 42, Email: user1.email, PassHash: user1.hashPass}, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, 42, user1.appID, "REFRESH", mock.AnythingOfType("time.Time")).
		Return(emptyID, errFailed)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_Login_AccessTokenProviderError(t *testing.T) {
	const op = "Auth.Login"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
		appID:    1,
	}
	user1.hashPass, _ = bcrypt.GenerateFromPassword([]byte(user1.password), bcrypt.DefaultCost)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: 42, Email: user1.email, PassHash: user1.hashPass}, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, 42, user1.appID, "REFRESH", mock.AnythingOfType("time.Time")).
		Return(100, nil)

	tokenProv.
		On("CreateAccessToken", 42, 100, user1.appID, mock.AnythingOfType("time.Time")).
		Return("", errFailed)

	tok, err := uc.Login(user1.ctx, user1.email, user1.password, user1.appID)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	userRepo.AssertExpectations(t)
	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_Register_Success(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
	}
	id := 42

	var savedHash []byte
	userRepo.
		On("SaveUser", user1.ctx, user1.email, mock.MatchedBy(func(passHash []byte) bool {
			if len(passHash) >= 59 && len(passHash) <= 60 && passHash[0] == '$' {
				savedHash = passHash
				return true
			}
			return false
		})).
		Return(nil)

	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{ID: id, Email: user1.email, PassHash: savedHash}, nil)

	userID, err := uc.Register(user1.ctx, user1.email, user1.password)

	require.NoError(t, err)
	assert.Equal(t, id, userID)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Register_SaveUserError(t *testing.T) {
	const op = "Auth.Register"
	errFailed := fmt.Errorf("failed")
	emptyID := 0

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
	}

	userRepo.
		On("SaveUser", user1.ctx, user1.email, mock.MatchedBy(func(passHash []byte) bool {
			return len(passHash) >= 59 && len(passHash) <= 60 &&
				len(passHash) > 0 && passHash[0] == '$'
		})).
		Return(errFailed)

	userID, err := uc.Register(user1.ctx, user1.email, user1.password)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, emptyID, userID)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Register_UserByEmailError(t *testing.T) {
	const op = "Auth.Register"
	errFailed := fmt.Errorf("failed")
	emptyID := 0

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:      context.Background(),
		email:    "test@example.com",
		password: "password",
	}

	userRepo.
		On("SaveUser", user1.ctx, user1.email, mock.MatchedBy(func(passHash []byte) bool {
			return len(passHash) >= 59 && len(passHash) <= 60 &&
				len(passHash) > 0 && passHash[0] == '$'
		})).
		Return(nil)
	userRepo.
		On("UserByEmail", user1.ctx, user1.email).
		Return(domain.User{}, errFailed)

	userID, err := uc.Register(user1.ctx, user1.email, user1.password)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, emptyID, userID)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_IsAdmin_SeccessTrue(t *testing.T) {
	// m := mocks.*
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx: context.Background(),
	}
	id := 42

	userRepo.
		On("IsAdmin", user1.ctx, id).
		Return(true, nil)

	isAdmin, err := uc.IsAdmin(user1.ctx, id)

	require.NoError(t, err)
	assert.Equal(t, true, isAdmin)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_IsAdmin_SeccessFalse(t *testing.T) {
	// m := mocks.*
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx: context.Background(),
	}
	id := 42

	userRepo.
		On("IsAdmin", user1.ctx, id).
		Return(false, nil)

	isAdmin, err := uc.IsAdmin(user1.ctx, id)

	require.NoError(t, err)
	assert.Equal(t, false, isAdmin)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_IsAdmin_IsAdminError(t *testing.T) {
	const op = "Auth.IsAdmin"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx: context.Background(),
	}
	id := 42

	userRepo.
		On("IsAdmin", user1.ctx, id).
		Return(false, errFailed)

	isAdmin, err := uc.IsAdmin(user1.ctx, id)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, false, isAdmin)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_SeccessOK(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "REFRESH",
	}

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(true, nil)

	ok, err := uc.Logout(user1.ctx, user1.refreshToken)

	require.NoError(t, err)
	assert.Equal(t, true, ok)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_SeccessFail(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "REFRESH",
	}

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(false, nil)

	ok, err := uc.Logout(user1.ctx, user1.refreshToken)

	require.NoError(t, err)
	assert.Equal(t, false, ok)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_Logout_RevokeByRefreshTokenError(t *testing.T) {
	const op = "Auth.Logout"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "REFRESH",
	}

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(false, errFailed)

	ok, err := uc.Logout(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, false, ok)

	userRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_Success(t *testing.T) {
	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "OLD_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(time.Hour * 24),
		Status:           "active",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(true, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("NEW_REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, session.UserID, session.AppID, "NEW_REFRESH", mock.AnythingOfType("time.Time")).
		Return(200, nil)

	tokenProv.
		On("CreateAccessToken", session.UserID, 200, session.AppID, mock.AnythingOfType("time.Time")).
		Return("NEW_ACCESS", nil)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.NoError(t, err)
	assert.Equal(t, "NEW_ACCESS", tok.AccessToken)
	assert.Equal(t, "NEW_REFRESH", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_SessionByRefreshTokenError(t *testing.T) {
	const op = "Auth.RefreshToken"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "INVALID_REFRESH",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(domain.Session{}, errFailed)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_InvalidStatus(t *testing.T) {
	const op = "Auth.RefreshToken"

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "REVOKED_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(time.Hour * 24),
		Status:           "revoked",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, provider.ErrInvalidRefreshToken)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_Expired(t *testing.T) {
	const op = "Auth.RefreshToken"

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "EXPIRED_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(-time.Hour),
		Status:           "active",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, provider.ErrInvalidRefreshToken)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_CreateRefreshTokenError(t *testing.T) {
	const op = "Auth.RefreshToken"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "OLD_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(time.Hour * 24),
		Status:           "active",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(true, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("", errFailed)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_CreateSessionError(t *testing.T) {
	const op = "Auth.RefreshToken"
	errFailed := fmt.Errorf("failed")
	emptyID := 0

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "OLD_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(time.Hour * 24),
		Status:           "active",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(true, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("NEW_REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, session.UserID, session.AppID, "NEW_REFRESH", mock.AnythingOfType("time.Time")).
		Return(emptyID, errFailed)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}

func TestAuthUseCase_RefreshToken_CreateAccessTokenError(t *testing.T) {
	const op = "Auth.RefreshToken"
	errFailed := fmt.Errorf("failed")

	userRepo := new(repoMocks.UserRepository)
	sessRepo := new(repoMocks.SessionRepository)
	tokenProv := new(providerMocks.TokenProvider)

	logger := config.NewLogger(&cfg)

	uc := usecase.NewAuthUseCase(
		userRepo,
		sessRepo,
		tokenProv,
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	user1 := testUserRequest{
		ctx:          context.Background(),
		refreshToken: "OLD_REFRESH",
		appID:        1,
	}

	session := domain.Session{
		ID:               100,
		UserID:           42,
		AppID:            user1.appID,
		RefreshExpiresAt: time.Now().Add(time.Hour * 24),
		Status:           "active",
	}

	sessRepo.
		On("SessionByRefreshToken", user1.ctx, user1.refreshToken).
		Return(session, nil)

	sessRepo.
		On("RevokeByRefreshToken", user1.ctx, user1.refreshToken).
		Return(true, nil)

	tokenProv.
		On("CreateRefreshToken").
		Return("NEW_REFRESH", nil)

	sessRepo.
		On("CreateSession", user1.ctx, session.UserID, session.AppID, "NEW_REFRESH", mock.AnythingOfType("time.Time")).
		Return(200, nil)

	tokenProv.
		On("CreateAccessToken", session.UserID, 200, session.AppID, mock.AnythingOfType("time.Time")).
		Return("", errFailed)

	tok, err := uc.RefreshToken(user1.ctx, user1.refreshToken)

	require.Error(t, err)
	require.ErrorIs(t, err, errFailed)
	assert.Contains(t, err.Error(), op)
	assert.Equal(t, "", tok.AccessToken)
	assert.Equal(t, "", tok.RefreshToken)

	sessRepo.AssertExpectations(t)
	tokenProv.AssertExpectations(t)
}
