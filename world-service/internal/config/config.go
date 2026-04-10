package config

// Config мира-сервиса.
type Config struct {
	BindAddr     string `toml:"bind_addr"`
	DatabaseURL  string `toml:"database_url"`
	LogLevel     string `toml:"log_level"`
	// ServiceToken — если непустой, требуется metadata "x-service-token: <token>" (или Bearer в authorization).
	ServiceToken string `toml:"service_token"`
}

// NewConfig ...
func NewConfig() *Config {
	return &Config{
		BindAddr:    ":50054",
		LogLevel:    "INFO",
		DatabaseURL: "postgres://messenger:messenger@localhost:5432/world_db?sslmode=disable",
	}
}
