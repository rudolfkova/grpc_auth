// Package app ...
package app

import (
	grpcapp "auth/internal/app/grpc"
	grpcauth "auth/internal/grpc/auth"
	"log/slog"
)

// App ...
type App struct {
	GRPCServer *grpcapp.App
}

// New ...
func New(log *slog.Logger, port string, auth grpcauth.Auth) *App {
	gRPCApp := grpcapp.New(log, port, auth)
	return &App{
		GRPCServer: gRPCApp,
	}
}
