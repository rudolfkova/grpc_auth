// Package config ...
package config

import (
	"log/slog"
	"os"
	"time"

	"github.com/phsym/console-slog"
)

// NewLogger ...
func NewLogger(config *Config) *slog.Logger {
	var lvl slog.LevelVar

	if err := lvl.UnmarshalText([]byte(config.LogLevel)); err != nil {
		lvl.Set(slog.LevelInfo)
	}

	handler := console.NewHandler(
		os.Stderr,
		&console.HandlerOptions{
			Level:      slog.LevelDebug,
			TimeFormat: time.TimeOnly,
		},
	)

	return slog.New(handler)
}
