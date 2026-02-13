// Package app ...
package app

import (
	grpcapp "auth/internal/app/grpc"
	"log/slog"
)

// App ...
type App struct {
	GRPCServer *grpcapp.App
}

// New ...
func New(log *slog.Logger, port string) *App {
	// TODO: инициализация хранилища

	// TODO: init auth service (auth)

	gRPCApp := grpcapp.New(log, port)
	return &App{
		GRPCServer: gRPCApp,
	}
}
