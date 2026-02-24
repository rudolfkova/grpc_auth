// Package app ...
package app

import (
	grpcapp "chat/internal/app/grpc"
	"chat/internal/grpc/chat"
	"log/slog"
)

// App ...
type App struct {
	GRPCServer *grpcapp.App
}

// New ...
func New(log *slog.Logger, port string, auth chat.Chat) *App {
	gRPCApp := grpcapp.New(log, port, auth)
	return &App{
		GRPCServer: gRPCApp,
	}
}
