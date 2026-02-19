// Package main ...
package main

import (
	"chat/internal/config"
	"flag"
	"log"
	"log/slog"

	"github.com/BurntSushi/toml"
)

var (
	configPath string
)

func init() {
	flag.StringVar(&configPath, "config-path", "config-chat.toml", "path to config file")
}

func main() {
	flag.Parse()

	cfg := config.NewConfig()
	_, err := toml.DecodeFile(configPath, cfg)
	if err != nil {
		log.Fatal(err)
	}
	logger := config.NewLogger(cfg)
	log := logger.With(
		slog.String("BindAddr:", cfg.BindAddr),
		slog.String("DatabaseURL:", cfg.DatabaseURL),
		slog.String("JWTSecret:", cfg.JWTSecret),
		slog.String("LogLevel:", cfg.LogLevel),
		slog.String("RedisAddr:", cfg.RedisAddr),
		slog.String("TestDatabaseURL:", cfg.TestDatabaseURL),
		slog.String("TestRedisAddr:", cfg.TestRedisAddr),
	)
	log.Info("init")

}
