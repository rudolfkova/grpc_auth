package gameecs

import (
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// SnapshotSystem строит снимок игроков и тайлов.
type SnapshotSystem struct {
	playerFilter *ecs.Filter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite]
	tileFilter   *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
}

// NewSnapshotSystem создаёт систему снимка.
func NewSnapshotSystem(
	playerFilter *ecs.Filter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *SnapshotSystem {
	return &SnapshotSystem{playerFilter: playerFilter, tileFilter: tileFilter}
}

func (s *SnapshotSystem) Update(ctx *TickContext) {
	q := s.playerFilter.Query()
	defer q.Close()

	out := make([]gamekit.Player, 0, 64)
	for q.Next() {
		ref, pos, _, hp, face, st, sp := q.Get()
		fdx, fdy := face.DX, face.DY
		if fdx == 0 && fdy == 0 {
			fdx, fdy = gamekit.DefaultPlayerFaceDX, gamekit.DefaultPlayerFaceDY
		}
		stats := *st
		if stats.IsUnset() {
			stats = gamekit.DefaultCharacterStats()
		}
		sprite := strings.TrimSpace(sp.Name)
		if sprite == "" {
			sprite = gamekit.DefaultPlayerSprite
		}
		out = append(out, gamekit.Player{
			ID:     ref.UserID,
			X:      pos.X,
			Y:      pos.Y,
			HP:     hp.HP,
			FaceDX: fdx,
			FaceDY: fdy,
			Stats:  stats,
			Sprite: sprite,
		})
	}
	ctx.Players = out

	tq := s.tileFilter.Query()
	defer tq.Close()
	tiles := make([]gamekit.Tile, 0, 64)
	for tq.Next() {
		pos, lay, face, tex, sol := tq.Get()
		t := gamekit.Tile{
			X:        pos.X,
			Y:        pos.Y,
			Layer:    lay.Z,
			Rotation: face.RotationQuarter,
			Texture:  tex.Name,
			Blocks:   sol.Blocks,
		}
		if len(tex.InstanceArgs) > 0 {
			t.InstanceArgs = tex.InstanceArgs
		}
		tiles = append(tiles, t)
	}
	ctx.Tiles = tiles
}
