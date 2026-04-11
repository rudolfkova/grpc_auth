package config

// Config character-service.
type Config struct {
	BindAddr     string `toml:"bind_addr"`
	DatabaseURL  string `toml:"database_url"`
	LogLevel     string `toml:"log_level"`
	ServiceToken string `toml:"service_token"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:    ":50055",
		LogLevel:    "INFO",
		DatabaseURL: "postgres://messenger:messenger@localhost:5432/character_db?sslmode=disable",
	}
}
