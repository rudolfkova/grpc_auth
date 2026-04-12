package gameecs

import (
	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// SystemRegistry оркестрирует системы Ark.
type SystemRegistry struct {
	perAction []System
	postTick  []System
}

// NewSystemRegistry собирает дефолтный набор систем для одного World.
func NewSystemRegistry(
	w *ecs.World,
	playerMapper *ecs.Map6[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats],
	playerFilter *ecs.Filter6[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats],
	tileMapper *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *SystemRegistry {
	return &SystemRegistry{
		perAction: []System{
			NewMoveIntentCaptureSystem(playerMapper),
			NewDamageSystem(playerMapper),
			NewTileSpawnSystem(w, tileMapper, tileFilter),
			NewTileClearSystem(w, tileFilter),
		},
		postTick: []System{
			NewMovementApplySystem(playerMapper, playerFilter, tileFilter),
			NewSnapshotSystem(playerFilter, tileFilter),
		},
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
