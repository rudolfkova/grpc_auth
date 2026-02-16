// Package config ...
package config

import "time"

// Config ...
type Config struct {
	DatabaseURL     string        `toml:"database_url"`
	BindAddr        string        `toml:"bind_addr"`
	AccessTokenTTL  time.Duration `toml:"access_token_tll"`
	RefreshTokenTTL time.Duration `toml:"refresh_token_ttl"`
	LogLevel        string        `toml:"log_level"`
	JWTSecret       string        `toml:"jwt_secret"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "info",
	}
}
