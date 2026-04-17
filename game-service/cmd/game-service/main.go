package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	gameapp "game/internal/app/game"
	"game/internal/config"
	"game/internal/infrastructure/characterclient"
	"game/internal/infrastructure/gameecs"
	prometheusobs "game/internal/infrastructure/observability/prometheus"
	"game/internal/infrastructure/worldclient"
	gamews "game/internal/ports/ws/game"

	"github.com/BurntSushi/toml"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"
	"google.golang.org/grpc/status"
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
	telemetry := prometheusobs.New(nil)

	var snapshot []byte
	if cfg.WorldID != "" {
		if cfg.WorldServiceAddr == "" {
			log.Fatalf("world_id is set but world_service_addr is empty (set in config or %s)", config.EnvWorldServiceAddr)
		}
		// world-service после depends_on может ещё поднимать Postgres и слушать gRPC позже первого «started» контейнера.
		const (
			perTryTimeout = 12 * time.Second
			retryInterval = 2 * time.Second
			maxWait       = 90 * time.Second
		)
		deadline := time.Now().Add(maxWait)
		var fetched *worldclient.Fetched
		var err error
		for {
			fetchCtx, cancel := context.WithTimeout(context.Background(), perTryTimeout)
			callStarted := time.Now()
			fetched, err = worldclient.GetWorld(fetchCtx, cfg.WorldServiceAddr, cfg.WorldServiceToken, cfg.WorldID)
			cancel()
			if err != nil {
				telemetry.ObserveGRPCClientCall("world", "get_world", status.Code(err).String(), time.Since(callStarted))
			} else {
				telemetry.ObserveGRPCClientCall("world", "get_world", "OK", time.Since(callStarted))
			}
			if err == nil {
				break
			}
			if time.Now().After(deadline) {
				log.Fatalf("load world from world-service: %v", err)
			}
			logger.Warn("world-service not ready, retrying", "addr", cfg.WorldServiceAddr, "world_id", cfg.WorldID, "err", err)
			time.Sleep(retryInterval)
		}
		logger.Info("world snapshot fetched",
			"world_id", cfg.WorldID,
			"version", fetched.Version,
			"schema_version", fetched.SchemaVersion,
			"snapshot_bytes", len(fetched.Snapshot),
		)
		snapshot = fetched.Snapshot
	}

	moveEvery := cfg.MovementApplyEveryNTicks
	if moveEvery < 1 {
		moveEvery = 1
	}
	var contentBundle *content.Bundle
	if cat := strings.TrimSpace(cfg.ContentCatalogPath); cat != "" {
		b, err := content.LoadBundle(cat, strings.TrimSpace(cfg.ContentScriptsDir))
		if err != nil {
			log.Fatalf("content bundle: %v", err)
		}
		contentBundle = b
	}
	tileFullEvery := cfg.TileFullSyncInterval
	if tileFullEvery <= 0 {
		tileFullEvery = time.Second
	}
	engine, err := gameecs.NewEngine(snapshot, moveEvery, gameecs.EngineOptions{
		Content:              contentBundle,
		Logger:               logger,
		TileFullSyncInterval: tileFullEvery,
	})
	if err != nil {
		log.Fatalf("engine: %v", err)
	}
	var worldStore *worldclient.StoreAdapter
	if strings.TrimSpace(cfg.WorldServiceAddr) != "" {
		worldStore = worldclient.NewStoreAdapter(cfg.WorldServiceAddr, cfg.WorldServiceToken)
	}
	var characterSessions *characterclient.SessionsAdapter
	if strings.TrimSpace(cfg.CharacterServiceAddr) != "" {
		characterSessions = characterclient.NewSessionsAdapter(cfg.CharacterServiceAddr, cfg.CharacterServiceToken)
	}

	app := gameapp.NewService(
		logger,
		engine,
		cfg.TickRate,
		cfg.QueueSize,
		worldStore,
		characterSessions,
		telemetry,
		cfg.SaveWorldAdminUserID,
	)
	wsHandler := gamews.NewHandler(logger, cfg.JWTSecret, app, telemetry)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go app.Run(ctx)

	mux := http.NewServeMux()
	mux.Handle("/ws/game", wsHandler)
	mux.HandleFunc("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})
	mux.Handle("/metrics", promhttp.Handler())

	if err := gamews.Serve(ctx, logger, cfg.BindAddr, mux); err != nil {
		log.Fatalf("game-service server: %v", err)
	}
}
