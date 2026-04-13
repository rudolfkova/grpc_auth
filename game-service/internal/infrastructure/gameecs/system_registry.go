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
	return &SystemRegistry{
		perAction: []System{
			NewMoveIntentCaptureSystem(playerMapper),
			NewDamageSystem(playerMapper),
			NewTileSpawnSystem(w, tileMapper, tileFilter),
			NewTileClearSystem(w, tileFilter),
			NewInteractSystem(contentBundle, interactLog, engine),
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
