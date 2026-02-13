package server

import (
	"log/slog"
	"os"
)

// NewLogger ...
func NewLogger(config *Config) *slog.Logger {
	var lvl slog.LevelVar

	if err := lvl.UnmarshalText([]byte(config.LogLevel)); err != nil {
		lvl.Set(slog.LevelInfo)
	}

	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: &lvl,
	})

	return slog.New(handler)
}
