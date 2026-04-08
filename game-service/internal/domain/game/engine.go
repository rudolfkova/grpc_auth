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
	state map[int64]actorState
}

type actorState struct {
	Pos models.Position
	HP  int
}

const defaultHP = 10

// NewEngine creates in-memory game engine.
func NewEngine() *Engine {
	return &Engine{state: make(map[int64]actorState)}
}

// ProcessTick applies a batch of actions and returns a stream of domain events.
// The engine can emit both broadcast and user-targeted events.
func (e *Engine) ProcessTick(actions []models.Action) []models.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	emit := newEmitter(16)

	for _, a := range actions {
		if a.PlayerID == 0 {
			continue
		}
		e.dispatchAction(a, emit)
	}

	// MVP: broadcast full state every tick.
	type statePayload struct {
		Players []models.Player `json:"players"`
		TickAt  time.Time       `json:"tick_at"`
	}
	players := make([]models.Player, 0, len(e.state))
	for id, st := range e.state {
		players = append(players, models.Player{
			ID: id,
			X:  st.Pos.X,
			Y:  st.Pos.Y,
			HP: st.HP,
		})
	}

	emit.Broadcast("state", statePayload{
		Players: players,
		TickAt:  time.Now().UTC(),
	})

	return emit.Events()
}

func (e *Engine) ensureActor(id int64) actorState {
	st, ok := e.state[id]
	if !ok {
		st = actorState{HP: defaultHP}
		e.state[id] = st
	}
	return st
}
