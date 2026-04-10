package model

import "time"

// World — запись мира в хранилище.
type World struct {
	ID             string
	Name           string
	Description    string
	Snapshot       []byte
	SchemaVersion  int32
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
