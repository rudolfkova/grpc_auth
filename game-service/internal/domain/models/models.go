package models

import (
	"encoding/json"
	"time"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// Action is a validated domain action for one tick.
type Action struct {
	PlayerID int64
	Type     string
	Payload  json.RawMessage
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
	Message         gamekit.Envelope
}
