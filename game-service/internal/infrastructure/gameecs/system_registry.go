package gameecs

import (
	"log/slog"

	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"

	"github.com/mlange-42/ark/ecs"
)

// SystemRegistry оркестрирует системы Ark.
type SystemRegistry struct {
	perAction []System
	postTick  []System
}

// Pipeline names фиксируют порядок выполнения систем (часть поведенческого контракта тика).
const (
	PipelineMoveIntentCapture = "move_intent_capture"
	PipelineDamage            = "damage"
	PipelineInventoryMove     = "inventory_move"
	PipelinePickup            = "pickup"
	PipelineDropItem          = "drop_item"
	PipelineTileSpawn         = "tile_spawn"
	PipelineTileClear         = "tile_clear"
	PipelineInteract          = "interact"
	PipelineMovementApply     = "movement_apply"
	PipelineSnapshot          = "snapshot"
)

type namedSystem struct {
	name   string
	system System
}

// NewSystemRegistry собирает дефолтный набор систем для одного World.
func NewSystemRegistry(
	w *ecs.World,
	playerMapper *ecs.Map7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite],
	playerFilter *ecs.Filter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite],
	tileMapper *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	contentBundle *content.Bundle,
	interactLog *slog.Logger,
	engine *Engine,
) *SystemRegistry {
	perAction := []namedSystem{
		{name: PipelineMoveIntentCapture, system: NewMoveIntentCaptureSystem(playerMapper)},
		{name: PipelineDamage, system: NewDamageSystem(playerMapper)},
		{name: PipelineInventoryMove, system: NewInventoryMoveSystem(engine)},
		{name: PipelinePickup, system: NewPickupSystem(contentBundle, engine)},
		{name: PipelineDropItem, system: NewDropItemSystem(engine)},
		{name: PipelineTileSpawn, system: NewTileSpawnSystem(w, tileMapper, tileFilter, engine)},
		{name: PipelineTileClear, system: NewTileClearSystem(w, tileFilter, engine)},
		{name: PipelineInteract, system: NewInteractSystem(contentBundle, interactLog, engine)},
	}
	postTick := []namedSystem{
		{name: PipelineMovementApply, system: NewMovementApplySystem(playerMapper, playerFilter, tileFilter)},
		{name: PipelineSnapshot, system: NewSnapshotSystem(playerFilter, tileFilter, func(ent ecs.Entity) gamekit.PlayerInventory {
			return engine.playerInventorySnapshotLocked(ent)
		})},
	}

	return &SystemRegistry{
		perAction: flattenSystems(perAction),
		postTick:  flattenSystems(postTick),
	}
}

func flattenSystems(in []namedSystem) []System {
	out := make([]System, 0, len(in))
	for _, s := range in {
		out = append(out, s.system)
	}
	return out
}

func (r *SystemRegistry) PerActionPipelineNames() []string {
	return []string{
		PipelineMoveIntentCapture,
		PipelineDamage,
		PipelineInventoryMove,
		PipelinePickup,
		PipelineDropItem,
		PipelineTileSpawn,
		PipelineTileClear,
		PipelineInteract,
	}
}

func (r *SystemRegistry) PostTickPipelineNames() []string {
	return []string{
		PipelineMovementApply,
		PipelineSnapshot,
	}
}

// Update выполняет per-action системы для каждого действия, затем post-tick (в т.ч. один шаг движения за тик).
func (r *SystemRegistry) Update(ctx *TickContext, actions []models.Action) ([]gamekit.Player, []gamekit.Tile) {
	ctx.Players = nil
	ctx.Tiles = nil

	for _, a := range actions {
		if a.PlayerID == 0 {
			continue
		}
		ctx.CurrentAction = a
		for _, sys := range r.perAction {
			sys.Update(ctx)
		}
	}

	ctx.CurrentAction = models.Action{}
	for _, sys := range r.postTick {
		sys.Update(ctx)
	}

	return ctx.Players, ctx.Tiles
}
