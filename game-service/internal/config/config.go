package config

import "time"

// Config contains game-service runtime options.
type Config struct {
	BindAddr  string        `toml:"bind_addr"`
	LogLevel  string        `toml:"log_level"`
	TickRate  time.Duration `toml:"tick_rate"`
	QueueSize int           `toml:"queue_size"`
	JWTSecret string        `toml:"jwt_secret"`
}

// NewConfig returns defaults for local/dev.
func NewConfig() *Config {
	return &Config{
		BindAddr:  ":50053",
		LogLevel:  "info",
		TickRate:  50 * time.Millisecond,
		QueueSize: 1024,
		JWTSecret: "123",
	}
}
