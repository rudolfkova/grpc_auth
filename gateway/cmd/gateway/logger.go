package main

import (
	"gateway/internal/config"
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

// NewLogger builds slog logger from config.
func NewLogger(cfg *config.Config) *slog.Logger {
	var lvl slog.LevelVar
	if err := lvl.UnmarshalText([]byte(cfg.LogLevel)); err != nil {
		lvl.Set(slog.LevelInfo)
	}

	handler := console.NewHandler(
		os.Stderr,
		&console.HandlerOptions{
			Level:      lvl.Level(),
			TimeFormat: time.TimeOnly,
		},
	)

	return slog.New(handler)
}
