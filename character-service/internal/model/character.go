package model

import "time"

// Character — сохранённый персонаж (opaque data для игрового ECS/JSON).
type Character struct {
	ID             string
	OwnerUserID    int64
	DisplayName    string
	Description    string
	Data           []byte
	SchemaVersion  int32
	Version        int64
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
