package gameecs

import (
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// SnapshotSystem строит снимок игроков и тайлов.
type SnapshotSystem struct {
	playerFilter *ecs.Filter4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health]
	tileFilter   *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
}

// NewSnapshotSystem создаёт систему снимка.
func NewSnapshotSystem(
	playerFilter *ecs.Filter4[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *SnapshotSystem {
	return &SnapshotSystem{playerFilter: playerFilter, tileFilter: tileFilter}
}

func (s *SnapshotSystem) Update(ctx *TickContext) {
	q := s.playerFilter.Query()
	defer q.Close()

	out := make([]gamekit.Player, 0, 64)
	for q.Next() {
		ref, pos, _, hp := q.Get()
		out = append(out, gamekit.Player{
			ID: ref.UserID,
			X:  pos.X,
			Y:  pos.Y,
			HP: hp.HP,
		})
	}
	ctx.Players = out

	tq := s.tileFilter.Query()
	defer tq.Close()
	tiles := make([]gamekit.Tile, 0, 64)
	for tq.Next() {
		pos, lay, face, tex, sol := tq.Get()
		tiles = append(tiles, gamekit.Tile{
			X:        pos.X,
			Y:        pos.Y,
			Layer:    lay.Z,
			Rotation: face.RotationQuarter,
			Texture:  tex.Name,
			Blocks:   sol.Blocks,
		})
	}
	ctx.Tiles = tiles
}
