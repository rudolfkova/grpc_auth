package grpcauth

import (
	authMocks "auth/mocks/auth"
	tokenjwt "auth/pkg/token"
	authv1 "auth/proto/auth/v1"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	ctx = context.Background()
)

func TestGRPCAuth_RegisterSuccess(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.RegisterRequest{
		Email:    "user@example.org",
		Password: "password",
	}
	id := 42

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Register", ctx, req.GetEmail(), req.GetPassword()).
		Return(id, nil)

	resp, err := server.Register(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, int64(id), resp.GetUserId())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_RegisterInternalError(t *testing.T) {
	errFailed := fmt.Errorf("failed")
	internalError := status.Error(codes.Internal, "internal error")
	emptyID := 0

	auth := new(authMocks.Auth)
	req := &authv1.RegisterRequest{
		Email:    "user@example.org",
		Password: "password",
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Register", ctx, req.GetEmail(), req.GetPassword()).
		Return(emptyID, errFailed)

	resp, err := server.Register(ctx, req)

	require.ErrorIs(t, err, internalError)
	assert.Equal(t, int64(emptyID), resp.GetUserId())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_LoginSuccess(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.LoginRequest{
		Email:    "user@example.org",
		Password: "password",
		AppId:    1,
	}

	server := serverAPI{
		auth: auth,
	}

	respToken := tokenjwt.Token{
		AccessToken:     "ACCESS",
		RefreshToken:    "REFRESH",
		AccessExpireAt:  time.Now(),
		RefreshExpireAt: time.Now(),
	}

	auth.
		On("Login", ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId())).
		Return(respToken, nil)

	resp, err := server.Login(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, respToken.AccessToken, resp.GetAccessToken())
	assert.Equal(t, respToken.RefreshToken, resp.GetRefreshToken())
	assert.Equal(t, timestamppb.New(respToken.AccessExpireAt), resp.GetAccessExpiresAt())
	assert.Equal(t, timestamppb.New(respToken.RefreshExpireAt), resp.GetRefreshExpiresAt())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_LoginInternalError(t *testing.T) {
	errFailed := fmt.Errorf("failed")
	internalError := status.Error(codes.Internal, "internal error")

	auth := new(authMocks.Auth)
	req := &authv1.LoginRequest{
		Email:    "user@example.org",
		Password: "password",
		AppId:    1,
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Login", ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId())).
		Return(tokenjwt.Token{}, errFailed)

	resp, err := server.Login(ctx, req)

	require.ErrorIs(t, err, internalError)
	assert.Nil(t, resp)

	auth.AssertExpectations(t)
}

func TestGRPCAuth_IsAdminSuccessTrue(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.IsAdminRequest{
		UserId: 42,
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("IsAdmin", ctx, int(req.GetUserId())).
		Return(true, nil)

	resp, err := server.IsAdmin(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, true, resp.GetIsAdmin())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_IsAdminSuccessFalse(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.IsAdminRequest{
		UserId: 42,
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("IsAdmin", ctx, int(req.GetUserId())).
		Return(false, nil)

	resp, err := server.IsAdmin(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, false, resp.GetIsAdmin())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_IsAdminInternalError(t *testing.T) {
	errFailed := fmt.Errorf("failed")
	internalError := status.Error(codes.Internal, "internal error")

	auth := new(authMocks.Auth)
	req := &authv1.IsAdminRequest{
		UserId: 42,
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("IsAdmin", ctx, int(req.GetUserId())).
		Return(false, errFailed)

	resp, err := server.IsAdmin(ctx, req)

	require.ErrorIs(t, err, internalError)
	assert.Equal(t, false, resp.GetIsAdmin())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_LogoutSuccess(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.LogoutRequest{
		RefreshToken: "REFRESH",
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Logout", ctx, req.GetRefreshToken()).
		Return(true, nil)

	resp, err := server.Logout(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, true, resp.GetSuccess())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_LogoutSuccessFail(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.LogoutRequest{
		RefreshToken: "REFRESH",
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Logout", ctx, req.GetRefreshToken()).
		Return(false, nil)

	resp, err := server.Logout(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, false, resp.GetSuccess())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_LogoutInternalError(t *testing.T) {
	errFailed := fmt.Errorf("failed")
	internalError := status.Error(codes.Internal, "internal error")

	auth := new(authMocks.Auth)
	req := &authv1.LogoutRequest{
		RefreshToken: "REFRESH",
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("Logout", ctx, req.GetRefreshToken()).
		Return(false, errFailed)

	resp, err := server.Logout(ctx, req)

	require.ErrorIs(t, err, internalError)
	assert.Equal(t, false, resp.GetSuccess())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_RefreshTokenSuccess(t *testing.T) {
	auth := new(authMocks.Auth)
	req := &authv1.RefreshTokenRequest{
		RefreshToken: "REFRESH",
	}

	server := serverAPI{
		auth: auth,
	}

	respToken := tokenjwt.Token{
		AccessToken:     "ACCESS",
		RefreshToken:    "REFRESH",
		AccessExpireAt:  time.Now(),
		RefreshExpireAt: time.Now(),
	}

	auth.
		On("RefreshToken", ctx, req.GetRefreshToken()).
		Return(respToken, nil)

	resp, err := server.RefreshToken(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, respToken.AccessToken, resp.GetAccessToken())
	assert.Equal(t, respToken.RefreshToken, resp.GetRefreshToken())
	assert.Equal(t, timestamppb.New(respToken.AccessExpireAt), resp.GetAccessExpiresAt())
	assert.Equal(t, timestamppb.New(respToken.RefreshExpireAt), resp.GetRefreshExpiresAt())

	auth.AssertExpectations(t)
}

func TestGRPCAuth_RefreshTokenInternalError(t *testing.T) {
	errFailed := fmt.Errorf("failed")
	internalError := status.Error(codes.Internal, "internal error")

	auth := new(authMocks.Auth)
	req := &authv1.RefreshTokenRequest{
		RefreshToken: "REFRESH",
	}

	server := serverAPI{
		auth: auth,
	}

	auth.
		On("RefreshToken", ctx, req.GetRefreshToken()).
		Return(tokenjwt.Token{}, errFailed)

	resp, err := server.RefreshToken(ctx, req)

	require.ErrorIs(t, err, internalError)
	assert.Nil(t, resp)

	auth.AssertExpectations(t)
}
