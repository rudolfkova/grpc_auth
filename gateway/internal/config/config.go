// Package config ...
package config

// Config ...
type Config struct {
	BindAddr        string `toml:"bind_addr"`
	AuthServiceAddr string `toml:"auth_service_addr"`
	ChatServiceAddr string `toml:"chat_service_addr"`
	JWTSecret       string `toml:"jwt_secret"`
	LogLevel        string `toml:"log_level"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:        ":8080",
		AuthServiceAddr: "localhost:50051",
		ChatServiceAddr: "localhost:50052",
		LogLevel:        "DEBUG",
	}
}
