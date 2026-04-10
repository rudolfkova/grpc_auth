package gameecs

import (
	"sync"
	"time"

	"game/internal/domain/gameplay"
	"game/internal/domain/models"
	"game/internal/domain/ports"
	"game/internal/domain/world"

	"github.com/mlange-42/ark/ecs"
)

// Engine — адаптер доменного порта GameEngine на Ark ECS.
type Engine struct {
	mu    sync.Mutex
	world *ecs.World

	byUser map[int64]ecs.Entity

	playerMapper *ecs.Map4[world.PlayerRef, world.GridPos, world.Speed, world.Health]
	systems      *SystemRegistry
}

var _ ports.GameEngine = (*Engine)(nil)

// NewEngine создаёт движок с миром Ark и зарегистрированными системами.
func NewEngine() *Engine {
	w := ecs.NewWorld()
	mapper := ecs.NewMap4[world.PlayerRef, world.GridPos, world.Speed, world.Health](w)
	filter := ecs.NewFilter4[world.PlayerRef, world.GridPos, world.Speed, world.Health](w)

	return &Engine{
		world:        w,
		byUser:       make(map[int64]ecs.Entity),
		playerMapper: mapper,
		systems:      NewSystemRegistry(mapper, filter),
	}
}

// ProcessTick применяет действия и возвращает доменные события.
func (e *Engine) ProcessTick(actions []models.Action) []models.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	emit := gameplay.NewEmitter(16)

	players := e.systems.Update(e, actions)

	type statePayload struct {
		Players []models.Player `json:"players"`
		TickAt  time.Time       `json:"tick_at"`
	}
	emit.Broadcast("state", statePayload{
		Players: players,
		TickAt:  time.Now().UTC(),
	})

	return emit.Events()
}

// EnsurePlayerEntity возвращает сущность игрока по user_id, создавая при необходимости.
func (e *Engine) EnsurePlayerEntity(userID int64) ecs.Entity {
	if ent, ok := e.byUser[userID]; ok {
		if e.playerMapper.HasAll(ent) {
			return ent
		}
		delete(e.byUser, userID)
	}

	ent := e.playerMapper.NewEntity(
		&world.PlayerRef{UserID: userID},
		&world.GridPos{X: 0, Y: 0},
		&world.Speed{MaxStep: 1},
		&world.Health{HP: world.DefaultPlayerHP},
	)
	e.byUser[userID] = ent
	return ent
}
