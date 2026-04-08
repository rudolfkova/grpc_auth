package models

import (
	"encoding/json"
	"time"
)

// Action is a validated domain action for one tick.
type Action struct {
	PlayerID int64
	Type     string
	Payload  json.RawMessage
}

// MoveIntent is external intent payload (ws input).
type MoveIntent struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

// HitIntent applies damage to a target player.
type HitIntent struct {
	TargetID int64 `json:"target_id"`
	Damage   int   `json:"damage"`
}

// Player is a projection for outbound snapshots.
type Player struct {
	ID int64 `json:"id"`
	X  int   `json:"x"`
	Y  int   `json:"y"`
	HP int   `json:"hp"`
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

// Envelope is the transport-neutral message wrapper for game WS.
// Port/adapter layers can send/receive it without parsing payload.
type Envelope struct {
	Service string          `json:"service"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Event is a domain event produced by the engine.
// RecipientUserID == 0 means broadcast.
type Event struct {
	RecipientUserID int64
	Type            string
	Payload         any
}

// Outbound is an application-level message ready to deliver via a port.
// RecipientUserID == 0 means broadcast.
type Outbound struct {
	RecipientUserID int64
	Message         Envelope
}
