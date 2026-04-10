package game

import (
	"sync"
	"time"

	"game/internal/domain/models"

	"github.com/mlange-42/ark/ecs"
)

// Engine держит Ark World и маппинг user_id → Entity.
type Engine struct {
	mu    sync.Mutex
	world *ecs.World

	byUser map[int64]ecs.Entity

	playerMapper *ecs.Map4[PlayerRef, GridPos, Speed, Health]
	playerFilter *ecs.Filter4[PlayerRef, GridPos, Speed, Health]
}

const defaultHP = 10

// NewEngine создаёт игровой движок с пустым ECS-миром.
func NewEngine() *Engine {
	w := ecs.NewWorld()
	return &Engine{
		world:        w,
		byUser:       make(map[int64]ecs.Entity),
		playerMapper: ecs.NewMap4[PlayerRef, GridPos, Speed, Health](w),
		playerFilter: ecs.NewFilter4[PlayerRef, GridPos, Speed, Health](w),
	}
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

	// MVP: broadcast full state every tick (снимок через ECS query).
	type statePayload struct {
		Players []models.Player `json:"players"`
		TickAt  time.Time       `json:"tick_at"`
	}
	players := e.runStateSnapshotQuery()

	emit.Broadcast("state", statePayload{
		Players: players,
		TickAt:  time.Now().UTC(),
	})

	return emit.Events()
}

func (e *Engine) ensurePlayerEntity(userID int64) ecs.Entity {
	if ent, ok := e.byUser[userID]; ok {
		if e.playerMapper.HasAll(ent) {
			return ent
		}
		delete(e.byUser, userID)
	}

	ent := e.playerMapper.NewEntity(
		&PlayerRef{UserID: userID},
		&GridPos{X: 0, Y: 0},
		&Speed{MaxStep: 1},
		&Health{HP: defaultHP},
	)
	e.byUser[userID] = ent
	return ent
}
