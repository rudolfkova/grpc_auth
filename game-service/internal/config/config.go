package config

import (
	"os"
	"time"
)

// EnvWorldID — имя переменной окружения с id мира для лобби (оркестратор задаёт на инстанс).
const EnvWorldID = "WORLD_ID"

// EnvWorldServiceAddr — gRPC адрес world-service (host:port).
const EnvWorldServiceAddr = "WORLD_SERVICE_ADDR"

// EnvWorldServiceToken — секрет для metadata x-service-token (если включён на world-service).
const EnvWorldServiceToken = "WORLD_SERVICE_TOKEN"

// Config contains game-service runtime options.
type Config struct {
	BindAddr  string        `toml:"bind_addr"`
	LogLevel  string        `toml:"log_level"`
	TickRate  time.Duration `toml:"tick_rate"`
	QueueSize int           `toml:"queue_size"`
	JWTSecret string        `toml:"jwt_secret"`
	// WorldID — id мира в world-service; из TOML или из EnvWorldID, если переменная задана.
	WorldID string `toml:"world_id"`
	// WorldServiceAddr — host:port gRPC world-service (нужен, если WorldID непустой).
	WorldServiceAddr string `toml:"world_service_addr"`
	WorldServiceToken string `toml:"world_service_token"`
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

// ApplyEnvOverrides подставляет значения из окружения поверх уже загруженного TOML.
// Переменная WORLD_ID: если задана в окружении инстанса (в т.ч. пустая строка), перезаписывает WorldID.
func ApplyEnvOverrides(cfg *Config) {
	if v, ok := os.LookupEnv(EnvWorldID); ok {
		cfg.WorldID = v
	}
	if v, ok := os.LookupEnv(EnvWorldServiceAddr); ok {
		cfg.WorldServiceAddr = v
	}
	if v, ok := os.LookupEnv(EnvWorldServiceToken); ok {
		cfg.WorldServiceToken = v
	}
}
