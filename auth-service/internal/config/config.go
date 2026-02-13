// Package config ...
package config

// Config ...
type Config struct {
	DatabaseURL     string `toml:"database_url"`
	BindAddr        string `toml:"bind_addr"`
	AccessTokenTTL  string `toml:"access_token_tll"`
	RefreshTokenTTL string `toml:"refresh_token_ttl"`
	LogLevel        string `toml:"log_level"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr: ":8080",
		LogLevel: "info",
	}
}
