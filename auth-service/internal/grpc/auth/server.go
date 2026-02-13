// Package grpcauth ...
package grpcauth

import (
	authv1 "auth/proto/auth/v1"
	"context"

	"google.golang.org/grpc"
)

type serverAPI struct {
	authv1.UnimplementedAuthServiceServer
}

// Register ...
func Register(gRPCServer *grpc.Server) {
	authv1.RegisterAuthServiceServer(gRPCServer, &serverAPI{})
}

// Register ...
func (s *serverAPI) Register(ctx context.Context, req *authv1.RegisterRequest) (*authv1.RegisterResponse, error) {
	_ = ctx
	_ = req
	panic("implement me")
}

// Login ...
func (s *serverAPI) Login(ctx context.Context, req *authv1.LoginRequest) (*authv1.LoginResponse, error) {
	_ = ctx
	_ = req
	panic("implement me")
}

// IsAdmin ...
func (s *serverAPI) IsAdmin(ctx context.Context, req *authv1.IsAdminRequest) (*authv1.IsAdminResponse, error) {
	_ = ctx
	_ = req
	panic("implement me")
}

// Logout ...
func (s *serverAPI) Logout(ctx context.Context, req *authv1.LogoutRequest) (*authv1.LogoutResponse, error) {
	_ = ctx
	_ = req
	panic("implement me")
}

// RefreshToken ...
func (s *serverAPI) RefreshToken(ctx context.Context, req *authv1.RefreshTokenRequest) (*authv1.RefreshTokenResponse, error) {
	_ = ctx
	_ = req
	panic("implement me")
}
