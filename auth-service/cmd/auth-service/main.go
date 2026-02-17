// main ...
package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/infrastructure/sqlstore"
	"auth/internal/usecase"
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "config.toml", "path to config file")
}

func main() {
	flag.Parse()

	cfg := config.NewConfig()
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		log.Fatal(err)
	}
	logger := config.NewLogger(cfg)

	db, err := sqlstore.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err = db.Close(); err != nil {
			logger.Error("bd closed with error", slog.String("err", err.Error()))
		}
	}()

	auth := usecase.NewAuthUseCase(
		sqlstore.NewUserRepository(db),
		sqlstore.NewSessionRepository(db),
		sqlstore.NewTokenProvider([]byte(cfg.JWTSecret)),
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	application := app.New(logger, cfg.BindAddr, auth)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := application.GRPCServer.Run(); err != nil {
			logger.Error("grpc server stopped with error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()
	application.GRPCServer.Stop()
}
