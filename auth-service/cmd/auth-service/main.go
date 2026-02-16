// main ...
package main

import (
	"auth/internal/app"
	"auth/internal/config"
	"auth/internal/infrastructure/sqlstore"
	"auth/internal/usecase"
	"flag"
	"log"

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

	db, err := sqlstore.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatal(err)
	}

	logger := config.NewLogger(cfg)

	auth := usecase.NewAuthUseCase(
		sqlstore.NewUserRepository(db),
		sqlstore.NewSessionRepository(db),
		sqlstore.NewTokenProvider([]byte(cfg.JWTSecret)),
		*logger,
		cfg.AccessTokenTTL,
		cfg.RefreshTokenTTL,
	)

	application := app.New(logger, cfg.BindAddr, auth)

	application.GRPCServer.MustRun()

}
