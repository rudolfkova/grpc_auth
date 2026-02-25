// Package app ...
package app

import (
	grpcapp "chat/internal/app/grpc"
	authclient "chat/internal/client/auth"
	"chat/internal/config"
	"chat/internal/grpc/chat"
	"chat/internal/grpc/hub"
	"log/slog"
)

// App ...
type App struct {
	GRPCServer *grpcapp.App
}

// New ...
func New(log *slog.Logger, auth chat.Chat, cfg *config.Config, authClient *authclient.Client, hub *hub.Hub) *App {
	gRPCApp := grpcapp.New(log, cfg.BindAddr, auth, cfg, authClient, hub)
	return &App{
		GRPCServer: gRPCApp,
	}
}
