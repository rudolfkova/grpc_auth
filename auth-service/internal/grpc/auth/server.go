// Package grpcauth ...
package grpcauth

import (
	"auth/internal/domain"
	"auth/internal/repository"
	authv1 "auth/proto/auth/v1"
	"auth/provider"
	"context"
	"errors"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Auth ...
type Auth interface {
	Register(ctx context.Context, email string, password string) (userID int, err error)
	Login(ctx context.Context, email string, password string, appID int) (token domain.Token, err error)
	IsAdmin(ctx context.Context, userID int) (isAdmin bool, err error)
	Logout(ctx context.Context, refreshToken string) (success bool, err error)
	RefreshToken(ctx context.Context, refreshToken string) (token domain.Token, err error)
	ValidateSession(ctx context.Context, sessionID int) (active bool, err error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServiceServer
	auth Auth
}

// Register ...
func Register(gRPCServer *grpc.Server, auth Auth) {
	authv1.RegisterAuthServiceServer(gRPCServer, &serverAPI{auth: auth})
}

// Ниже бизнес логика сервиса, rpc методы.

// Register ...
func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	if err := ValidateRegisterRequest(req); err != nil {
		return nil, status.Error(codes.InvalidArgument, "internal error")
	}
	
	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		if errors.Is(err, repository.ErrUserAlreadyExists) {
			return nil, status.Error(codes.AlreadyExists, "user with this email already exists")
		}
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid user")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RegisterResponse{
		UserId: int64(userID),
	}, nil
}

// Login ...
func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), int(req.GetAppId()))
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "wrong email or password")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LoginResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		AccessExpiresAt:  timestamppb.New(token.AccessExpireAt),
		RefreshExpiresAt: timestamppb.New(token.RefreshExpireAt),
	}, nil
}

// IsAdmin ...
func (s *serverAPI) IsAdmin(ctx context.Context, req *authv1.IsAdminRequest) (*authv1.IsAdminResponse, error) {
	isAdmin, err := s.auth.IsAdmin(ctx, int(req.GetUserId()))
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.IsAdminResponse{
		IsAdmin: isAdmin,
	}, nil
}

// Logout ...
func (s *serverAPI) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	success, err := s.auth.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid user")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.LogoutResponse{
		Success: success,
	}, nil
}

// RefreshToken ...
func (s *serverAPI) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	token, err := s.auth.RefreshToken(ctx, req.RefreshToken)
	if err != nil {
		if errors.Is(err, repository.ErrInvalidCredentials) {
			return nil, status.Error(codes.Unauthenticated, "invalid user")
		}
		if errors.Is(err, provider.ErrInvalidRefreshToken) {
			return nil, status.Error(codes.Unauthenticated, "invalid user")
		}
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		AccessExpiresAt:  timestamppb.New(token.AccessExpireAt),
		RefreshExpiresAt: timestamppb.New(token.RefreshExpireAt),
	}, nil
}

// ValidateSession ...
func (s *serverAPI) ValidateSession(ctx context.Context, req *authv1.ValidateSessionRequest) (*authv1.ValidateSessionResponse, error) {
	active, err := s.auth.ValidateSession(ctx, int(req.GetSessionId()))
	if err != nil {
		// если хочешь заморочиться, можешь маппить ошибки по‑красивому
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.ValidateSessionResponse{
		Active: active,
	}, nil
}
