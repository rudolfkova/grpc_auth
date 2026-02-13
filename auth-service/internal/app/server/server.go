// Package server ...
package server

import (
	"auth/internal/app/infrastucture/sqlstore"
	"auth/internal/app/usecase"
	"database/sql"
	"log/slog"
)

// Server ...
type Server struct {
	config      *Config
	authUseCase *usecase.AuthUseCase
	logger      *slog.Logger
}

// Start ...
func Start(config *Config) error {
	db, err := newDB(config.DatabaseURL)
	if err != nil {
		return err
	}
	userRepository := sqlstore.NewUserRepository(db)
	sessionRepository := sqlstore.NewSessionRepository(db)

	authUseCase := usecase.NewAuthUseCase(userRepository, sessionRepository)

	logger := NewLogger(config)

	New(config, logger, authUseCase)

	return nil
}

// newDB ...
func newDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return db, nil
}

// New ...
func New(config *Config, logger *slog.Logger, authUseCase *usecase.AuthUseCase) *Server {
	s := &Server{
		config:      config,
		authUseCase: authUseCase,
		logger:      logger,
	}

	return s

}
