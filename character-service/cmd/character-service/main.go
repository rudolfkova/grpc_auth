package main

import (
	"context"
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"character/internal/config"
	charactergrpc "character/internal/grpc/character"
	"character/internal/interceptor"
	"character/internal/repository/sqlstore"
	"character/internal/service"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "config-character.toml", "path to config file")
}

func main() {
	flag.Parse()

	cfg := config.NewConfig()
	if _, err := toml.DecodeFile(configPath, cfg); err != nil {
		log.Fatalf("config: %v", err)
	}
	logger := config.NewLogger(cfg)

	store, err := sqlstore.New(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db: %v", err)
	}
	defer func() { _ = store.DB.Close() }()

	charSvc := service.NewCharacter(sqlstore.NewCharacterRepo(store))

	opts := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(interceptor.ServiceAuthUnary(cfg.ServiceToken)),
	}
	srv := grpc.NewServer(opts...)
	charactergrpc.Register(srv, charSvc, logger)
	reflection.Register(srv)

	ln, err := net.Listen("tcp", cfg.BindAddr)
	if err != nil {
		log.Fatalf("listen: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	logger.Info("character-service listening", "addr", cfg.BindAddr)
	if err := srv.Serve(ln); err != nil {
		log.Fatalf("grpc: %v", err)
	}
}
