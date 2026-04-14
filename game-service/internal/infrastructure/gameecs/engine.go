package gameecs

import (
	"log/slog"
	"sync"
	"time"

	"game/internal/domain/gameplay"
	"game/internal/domain/models"
	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"

	arkserde "github.com/mlange-42/ark-serde"
	"github.com/mlange-42/ark/ecs"
)

// EngineOptions опции NewEngine (нулевое значение — только ECS, без каталога content).
type EngineOptions struct {
	Content *content.Bundle
	Logger  *slog.Logger
	// TileFullSyncInterval — как часто в payload state отдавать полный список тайлов; между полными снимками — только tile_updates.
	// <=0: по умолчанию 1s.
	TileFullSyncInterval time.Duration
}

// Engine — адаптер доменного порта GameEngine на Ark ECS.
type Engine struct {
	mu    sync.Mutex
	world *ecs.World

	byUser map[int64]ecs.Entity

	playerMapper *ecs.Map7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite]
	tileMapper   *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
	tileFilter   *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
	systems      *SystemRegistry
	moveIntents  MoveIntentStore

	moveApplyEvery   int
	moveApplyCounter int
	diagStride       diagStrideState

	tileFullSyncEvery    time.Duration
	lastFullTileSyncAt   time.Time
	pendingTileUpdates   []gamekit.TileUpdate
}

var _ ports.GameEngine = (*Engine)(nil)

// NewEngine создаёт движок с миром Ark и зарегистрированными системами.
// snapshot — JSON от ark-serde (github.com/mlange-42/ark-serde, Serialize); пустой слайс = пустой мир.
// movementApplyEveryNTicks — применять шаг движения не чаще чем раз в N тиков симуляции; <1 трактуется как 1.
func NewEngine(snapshot []byte, movementApplyEveryNTicks int, opts EngineOptions) (*Engine, error) {
	if movementApplyEveryNTicks < 1 {
		movementApplyEveryNTicks = 1
	}
	w := ecs.NewWorld()
	playerMapper := ecs.NewMap7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite](w)
	playerFilter := ecs.NewFilter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite](w)
	tileMapper := ecs.NewMap5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid](w)
	tileFilter := ecs.NewFilter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid](w)
	tileEvery := opts.TileFullSyncInterval
	if tileEvery <= 0 {
		tileEvery = time.Second
	}
	e := &Engine{
		world:             w,
		byUser:            make(map[int64]ecs.Entity),
		playerMapper:      playerMapper,
		tileMapper:        tileMapper,
		tileFilter:        tileFilter,
		moveApplyEvery:    movementApplyEveryNTicks,
		tileFullSyncEvery: tileEvery,
	}
	reg := NewSystemRegistry(w, playerMapper, playerFilter, tileMapper, tileFilter, opts.Content, opts.Logger, e)
	e.systems = reg
	if err := e.applyArkWorldSnapshot(snapshot); err != nil {
		return nil, err
	}
	if opts.Content != nil {
		registerGameContentOps(opts.Content.Runner, e)
	}
	return e, nil
}

// ProcessTick применяет действия и возвращает доменные события.
func (e *Engine) ProcessTick(actions []models.Action) []models.Event {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.resetTileDeltaForTick()

	emit := gameplay.NewEmitter(16)

	applyMove := true
	if e.moveApplyEvery > 1 {
		applyMove = e.moveApplyCounter%e.moveApplyEvery == 0
		e.moveApplyCounter++
	}

	tickCtx := &TickContext{
		Sink:          e,
		Intents:       &e.moveIntents,
		ApplyMovement: applyMove,
		Diag:          &e.diagStride,
	}
	players, tiles := e.systems.Update(tickCtx, actions)

	now := time.Now().UTC()
	emit.Broadcast("state", e.statePayloadForTick(now, players, tiles))

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

	st := gamekit.DefaultCharacterStats()
	spr := gamekit.PlayerSprite{Name: gamekit.DefaultPlayerSprite}
	ent := e.playerMapper.NewEntity(
		&gamekit.PlayerRef{UserID: userID},
		&gamekit.GridPos{X: 0, Y: 0},
		&gamekit.Speed{MaxStep: 1},
		&gamekit.Health{HP: gamekit.DefaultPlayerHP},
		&gamekit.PlayerFace{DX: gamekit.DefaultPlayerFaceDX, DY: gamekit.DefaultPlayerFaceDY},
		&st,
		&spr,
	)
	e.byUser[userID] = ent
	return ent
}

// EnsurePlayerJoin создаёт сущность при первом заходе user_id из CharacterPlayData; если игрок уже есть — не меняет компоненты.
func (e *Engine) EnsurePlayerJoin(userID int64, d gamekit.CharacterPlayData) {
	d.Normalize()
	e.mu.Lock()
	defer e.mu.Unlock()
	if ent, ok := e.byUser[userID]; ok && e.playerMapper.HasAll(ent) {
		return
	}
	hp := d.HP
	if hp <= 0 {
		hp = gamekit.DefaultPlayerHP
	}
	faceDX, faceDY := d.FaceDX, d.FaceDY
	if faceDX == 0 && faceDY == 0 {
		faceDX, faceDY = gamekit.DefaultPlayerFaceDX, gamekit.DefaultPlayerFaceDY
	}
	st := d.Stats
	spr := gamekit.PlayerSprite{Name: d.Sprite}
	ent := e.playerMapper.NewEntity(
		&gamekit.PlayerRef{UserID: userID},
		&gamekit.GridPos{X: d.X, Y: d.Y},
		&gamekit.Speed{MaxStep: 1},
		&gamekit.Health{HP: hp},
		&gamekit.PlayerFace{DX: faceDX, DY: faceDY},
		&st,
		&spr,
	)
	e.byUser[userID] = ent
}

// PlayerCharacterData сериализует CharacterPlayData в JSON для character.data.
func (e *Engine) PlayerCharacterData(userID int64) ([]byte, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	ent, ok := e.byUser[userID]
	if !ok || !e.playerMapper.HasAll(ent) {
		return gamekit.MarshalCharacterPlayData(gamekit.NewDefaultCharacterPlayData())
	}
	_, pos, _, hp, face, st, sp := e.playerMapper.Get(ent)
	cpd := gamekit.CharacterPlayData{
		X: pos.X, Y: pos.Y,
		HP: hp.HP, FaceDX: face.DX, FaceDY: face.DY,
		Stats:  *st,
		Sprite: sp.Name,
	}
	return gamekit.MarshalCharacterPlayData(cpd)
}
