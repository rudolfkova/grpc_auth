// Package main ...
package main

import (
	"flag"
	"gateway/internal/config"
	"gateway/internal/handler"
	"gateway/internal/middleware"
	authv1 "gateway/proto/auth/v1"
	chatv1 "gateway/proto/chat/v1"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "config-gateway.toml", "path to config file")
}

func main() {
	flag.Parse()

	cfg := config.NewConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		log.Printf("config not loaded (%v), using defaults", err)
	}

	logLevel := slog.LevelInfo
	if strings.ToUpper(cfg.LogLevel) == "DEBUG" {
		logLevel = slog.LevelDebug
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel}))

	authConn, err := grpc.NewClient(
		cfg.AuthServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to auth-service: %v", err)
	}
	defer authConn.Close()

	chatConn, err := grpc.NewClient(
		cfg.ChatServiceAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		log.Fatalf("failed to connect to chat-service: %v", err)
	}
	defer chatConn.Close()

	authHandler := handler.NewAuthHandler(authv1.NewAuthServiceClient(authConn))
	chatHandler := handler.NewChatHandler(chatv1.NewChatServiceClient(chatConn))

	mux := http.NewServeMux()

	// Auth
	mux.HandleFunc("POST /auth/register", authHandler.Register)
	mux.HandleFunc("POST /auth/login", authHandler.Login)
	mux.HandleFunc("POST /auth/logout", authHandler.Logout)
	mux.HandleFunc("POST /auth/refresh", authHandler.Refresh)
	mux.HandleFunc("GET /auth/is-admin", authHandler.IsAdmin)

	// Chat
	mux.HandleFunc("POST /chat/get-or-create", chatHandler.GetOrCreateChat)
	mux.HandleFunc("GET /chat/messages", chatHandler.GetMessages)
	mux.HandleFunc("GET /chat/chats", chatHandler.GetUserChats)
	mux.HandleFunc("POST /chat/send", chatHandler.SendMessage)

	// Health
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr: cfg.BindAddr,
		Handler: middleware.CORS(
			middleware.Logger(logger, mux),
		),
	}

	logger.Info("gateway starting",
		slog.String("addr", cfg.BindAddr),
		slog.String("auth", cfg.AuthServiceAddr),
		slog.String("chat", cfg.ChatServiceAddr),
	)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("gateway: %v", err)
	}
}
