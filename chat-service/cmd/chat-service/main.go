// Package main ...
package main

import (
	"chat/internal/app"
	authclient "chat/internal/client/auth"
	"chat/internal/config"
	"chat/internal/service"
	"context"
	"flag"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/BurntSushi/toml"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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

	chatAPI := service.NewService()

	authConn, _ := grpc.NewClient(cfg.AuthServiceAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	authClient := authclient.New(authConn)
	app := app.New(logger, chatAPI, cfg, authClient)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := app.GRPCServer.Run(); err != nil {
			logger.Error("grpc server stopped with error", slog.String("err", err.Error()))
		}
	}()

	<-ctx.Done()
	app.GRPCServer.Stop()

}
