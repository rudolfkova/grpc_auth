package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	gameapp "game/internal/app/game"
	"game/internal/config"
	domain "game/internal/domain/game"
	gamews "game/internal/ports/ws/game"

	"github.com/BurntSushi/toml"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "config-game.toml", "path to config file")
}

func main() {
	flag.Parse()

	cfg := config.NewConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		log.Fatalf("config decode: %v", err)
	}

	logger := config.NewLogger(cfg)
	engine := domain.NewEngine()
	app := gameapp.NewService(engine, cfg.TickRate, cfg.QueueSize)
	wsHandler := gamews.NewHandler(logger, cfg.JWTSecret, app)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go app.Run(ctx)

	mux := http.NewServeMux()
	mux.Handle("/ws/game", wsHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	if err := gamews.Serve(ctx, logger, cfg.BindAddr, mux); err != nil {
		log.Fatalf("game-service server: %v", err)
	}
}
