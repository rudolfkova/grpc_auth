// Package app ...
package app

import (
	grpcapp "auth/internal/app/grpc"
	"auth/internal/usecase"
	"log/slog"
)

// App ...
type App struct {
	GRPCServer *grpcapp.App
}

// New ...
func New(log *slog.Logger, port string, auth *usecase.AuthUseCase) *App {
	// TODO: инициализация хранилища

	// TODO: init auth service (auth)

	gRPCApp := grpcapp.New(log, port, auth)
	return &App{
		GRPCServer: gRPCApp,
	}
}
