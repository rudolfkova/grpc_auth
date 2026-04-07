package game

import (
	"sync"
	"time"

	"game/internal/domain/models"
)

// Engine contains only game domain state/logic.
// It knows nothing about transport, auth, or timers.
type Engine struct {
	mu    sync.Mutex
	state map[int64]models.Position
}

// NewEngine creates in-memory game engine.
func NewEngine() *Engine {
	return &Engine{state: make(map[int64]models.Position)}
}

// ProcessTick applies a batch of validated actions and returns snapshot.
func (e *Engine) ProcessTick(actions []models.Action) models.Snapshot {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, a := range actions {
		if a.Type != "move" || a.PlayerID == 0 {
			continue
		}
		pos := e.state[a.PlayerID]
		pos.X += a.DX
		pos.Y += a.DY
		e.state[a.PlayerID] = pos
	}

	out := models.Snapshot{
		Players: make(map[int64]models.Position, len(e.state)),
		TickAt:  time.Now().UTC(),
	}
	for id, pos := range e.state {
		out.Players[id] = pos
	}
	return out
}
