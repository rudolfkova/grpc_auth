package models

import "time"

// Action is a validated domain action for one tick.
type Action struct {
	PlayerID int64
	Type     string
	DX       int
	DY       int
}

// MoveIntent is external intent payload (ws input).
type MoveIntent struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

// Player is a projection for outbound snapshots.
type Player struct {
	ID int64 `json:"id"`
	X  int   `json:"x"`
	Y  int   `json:"y"`
}

// Position is internal world coordinate.
type Position struct {
	X int
	Y int
}

// Snapshot is a world state for current tick.
type Snapshot struct {
	Players map[int64]Position
	TickAt  time.Time
}
