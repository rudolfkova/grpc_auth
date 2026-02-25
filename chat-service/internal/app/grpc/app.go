// Package grpcapp ...
package grpcapp

import (
	authclient "chat/internal/client/auth"
	"chat/internal/config"
	"chat/internal/grpc/chat"
	"chat/internal/grpc/hub"
	"chat/internal/interceptor"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"
)

// App ...
type App struct {
	logger     *slog.Logger
	gRPCServer *grpc.Server
	port       string
}

// New ...
func New(log *slog.Logger, port string, auth chat.Chat, cfg *config.Config, authClient *authclient.Client, hub *hub.Hub) *App {
	gRPCServer := grpc.NewServer(
		grpc.UnaryInterceptor(
			interceptor.AuthInterceptor(cfg.JWTSecret, authClient),
		),
		grpc.StreamInterceptor(
			interceptor.AuthStreamInterceptor(cfg.JWTSecret, authClient),
		),
	)
	chat.Register(gRPCServer, auth, hub, log)

	return &App{
		logger:     log,
		gRPCServer: gRPCServer,
		port:       port,
	}
}

// Run ...
func (a *App) Run() error {
	const op = "grpcapp.Run"

	log := a.logger.With(
		slog.String("op", op),
		slog.String("port", a.port),
	)

	l, err := net.Listen("tcp", a.port)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("grpc server started", slog.String("addr", l.Addr().String()))

	if err := a.gRPCServer.Serve(l); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// MustRun ...
func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

// Stop ...
func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.logger.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.String("port", a.port))

	a.gRPCServer.GracefulStop()
}
