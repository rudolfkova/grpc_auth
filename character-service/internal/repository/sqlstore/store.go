package sqlstore

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// Store — пул соединений Postgres.
type Store struct {
	DB *sql.DB
}

// New открывает БД.
func New(databaseURL string) (*Store, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Store{DB: db}, nil
}
