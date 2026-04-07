package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"game/internal/config"
	"game/internal/game"

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
	engine := game.NewEngine(logger, cfg.TickRate, cfg.QueueSize)

	// Seed one action to validate end-to-end loop on startup.
	engine.EnqueueAction(game.Action{PlayerID: 1, Type: "move", DX: 1, DY: 0})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go engine.Run(ctx)

	for {
		select {
		case <-ctx.Done():
			logger.Info("shutdown signal received")
			return
		case s := <-engine.Events():
			logger.Debug("snapshot", slog.Int("players", len(s.Players)))
		}
	}
}
