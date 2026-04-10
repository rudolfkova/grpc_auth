package gameecs

import (
	"sync"
	"time"

	"game/internal/domain/gameplay"
	"game/internal/domain/models"
	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	arkserde "github.com/mlange-42/ark-serde"
	"github.com/mlange-42/ark/ecs"
)

// Engine — адаптер доменного порта GameEngine на Ark ECS.
type Engine struct {
	mu    sync.Mutex
	world *ecs.World

	byUser map[int64]ecs.Entity

	playerMapper *ecs.Map4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health]
	systems      *SystemRegistry
}

var _ ports.GameEngine = (*Engine)(nil)

// NewEngine создаёт движок с миром Ark и зарегистрированными системами.
// snapshot — JSON от ark-serde (github.com/mlange-42/ark-serde, Serialize); пустой слайс = пустой мир.
func NewEngine(snapshot []byte) (*Engine, error) {
	w := ecs.NewWorld()
	playerMapper := ecs.NewMap4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health](w)
	playerFilter := ecs.NewFilter4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health](w)
	tileMapper := ecs.NewMap5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid](w)
	tileFilter := ecs.NewFilter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid](w)
	reg := NewSystemRegistry(w, playerMapper, playerFilter, tileMapper, tileFilter)
	e := &Engine{
		world:        w,
		byUser:       make(map[int64]ecs.Entity),
		playerMapper: playerMapper,
		systems:      reg,
	}
	if err := e.applyArkWorldSnapshot(snapshot); err != nil {
		return nil, err
	}
	return e, nil
}

// ProcessTick применяет действия и возвращает доменные события.
func (e *Engine) ProcessTick(actions []models.Action) []models.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	emit := gameplay.NewEmitter(16)

	players, tiles := e.systems.Update(e, actions)

	type statePayload struct {
		Players []gamekit.Player `json:"players"`
		Tiles   []gamekit.Tile   `json:"tiles"`
		TickAt  time.Time        `json:"tick_at"`
	}
	emit.Broadcast("state", statePayload{
		Players: players,
		Tiles:   tiles,
		TickAt:  time.Now().UTC(),
	})

	return emit.Events()
}

// SerializeWorld сериализует текущий ECS-мир (ark-serde JSON).
func (e *Engine) SerializeWorld() ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	return arkserde.Serialize(e.world)
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
		&gamekit.PlayerRef{UserID: userID},
		&gamekit.GridPos{X: 0, Y: 0},
		&gamekit.Speed{MaxStep: 1},
		&gamekit.Health{HP: gamekit.DefaultPlayerHP},
	)
	e.byUser[userID] = ent
	return ent
}
