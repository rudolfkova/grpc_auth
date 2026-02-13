// Package grpcauth ...
package grpcauth

import (
	"auth/internal/domain"
	authv1 "auth/proto/auth/v1"
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Auth ...
type Auth interface {
	Register(ctx context.Context, email string, password string) (userID int64, err error)
	Login(ctx context.Context, email string, password string, appID int32) (token domain.Token, err error)
	IsAdmin(ctx context.Context, userID int64) (isAdmin bool, err error)
	Logout(ctx context.Context, refreshToken string) (success bool, err error)
	RefreshToken(ctx context.Context, refreshToken string) (token domain.Token, err error)
}

type serverAPI struct {
	authv1.UnimplementedAuthServiceServer
	auth Auth
}

// Register ...
func Register(gRPCServer *grpc.Server) {
	authv1.RegisterAuthServiceServer(gRPCServer, &serverAPI{})
}

// Ниже бизнес логика сервиса, rpc методы.

// Register ...
func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	userID, err := s.auth.Register(ctx, req.GetEmail(), req.GetPassword())
	if err != nil {
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RegisterResponse{
		UserId: userID,
	}, nil
}

// Login ...
func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	token, err := s.auth.Login(ctx, req.GetEmail(), req.GetPassword(), req.GetAppId())
	if err != nil {
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
	isAdmin, err := s.auth.IsAdmin(ctx, req.GetUserId())
	if err != nil {
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
		return nil, status.Error(codes.Internal, "internal error")
	}

	return &authv1.RefreshTokenResponse{
		AccessToken:      token.AccessToken,
		RefreshToken:     token.RefreshToken,
		AccessExpiresAt:  timestamppb.New(token.AccessExpireAt),
		RefreshExpiresAt: timestamppb.New(token.RefreshExpireAt),
	}, nil
}
