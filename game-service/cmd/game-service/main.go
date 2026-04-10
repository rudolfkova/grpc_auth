package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	gameapp "game/internal/app/game"
	"game/internal/config"
	"game/internal/infrastructure/gameecs"
	"game/internal/infrastructure/worldclient"
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
	config.ApplyEnvOverrides(cfg)

	logger := config.NewLogger(cfg)
	logger.Info("game-service starting", "addr", cfg.BindAddr, "world_id", cfg.WorldID)

	var snapshot []byte
	if cfg.WorldID != "" {
		if cfg.WorldServiceAddr == "" {
			log.Fatalf("world_id is set but world_service_addr is empty (set in config or %s)", config.EnvWorldServiceAddr)
		}
		fetchCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		fetched, err := worldclient.GetWorld(fetchCtx, cfg.WorldServiceAddr, cfg.WorldServiceToken, cfg.WorldID)
		cancel()
		if err != nil {
			log.Fatalf("load world from world-service: %v", err)
		}
		logger.Info("world snapshot fetched",
			"world_id", cfg.WorldID,
			"version", fetched.Version,
			"schema_version", fetched.SchemaVersion,
			"snapshot_bytes", len(fetched.Snapshot),
		)
		snapshot = fetched.Snapshot
	}

	engine, err := gameecs.NewEngine(snapshot)
	if err != nil {
		log.Fatalf("engine: %v", err)
	}
	app := gameapp.NewService(logger, engine, cfg.TickRate, cfg.QueueSize,
		cfg.WorldServiceAddr, cfg.WorldServiceToken, cfg.SaveWorldAdminUserID)
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
