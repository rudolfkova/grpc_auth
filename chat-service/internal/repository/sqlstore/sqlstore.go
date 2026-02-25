// Package sqlstore ...
package sqlstore

import "database/sql"

// SQLStore ...
type SQLStore struct {
	DB *sql.DB
}

// New ...
func New(databaseURL string) (*SQLStore, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	return &SQLStore{DB: db}, nil
}
