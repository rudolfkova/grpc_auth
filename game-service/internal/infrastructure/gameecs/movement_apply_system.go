package gameecs

import (
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// MovementApplySystem один раз за тик сдвигает каждого игрока по текущему интенту из MoveIntentStore.
type MovementApplySystem struct {
	mapper       *ecs.Map5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace]
	playerFilter *ecs.Filter5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace]
	tileFilter   *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid]
}

// NewMovementApplySystem ...
func NewMovementApplySystem(
	mapper *ecs.Map5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace],
	playerFilter *ecs.Filter5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace],
	tileFilter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
) *MovementApplySystem {
	return &MovementApplySystem{mapper: mapper, playerFilter: playerFilter, tileFilter: tileFilter}
}

func (s *MovementApplySystem) Update(ctx *TickContext) {
	if ctx.Intents == nil || !ctx.ApplyMovement {
		return
	}
	q := s.playerFilter.Query()
	defer q.Close()
	for q.Next() {
		ref, _, _, _, _ := q.Get()
		dx, dy := ctx.Intents.Get(ref.UserID)
		if dx == 0 && dy == 0 {
			continue
		}
		adx, ady, split := dx, dy, false
		if ctx.Diag != nil {
			adx, ady, split = ctx.Diag.pick(ref.UserID, dx, dy)
		}
		moved := s.applyStep(q.Entity(), adx, ady)
		if ctx.Diag != nil {
			ctx.Diag.onApplied(ref.UserID, split, moved)
		}
	}
}

func (s *MovementApplySystem) applyStep(ent ecs.Entity, dx, dy int) bool {
	if !s.mapper.HasAll(ent) {
		return false
	}
	_, pos, speed, _, face := s.mapper.Get(ent)
	step := speed.MaxStep
	if step <= 0 {
		return false
	}
	if dx < -step || dx > step || dy < -step || dy > step {
		return false
	}
	tx, ty := pos.X+dx, pos.Y+dy
	if s.blockedCell(tx, ty) {
		return false
	}
	pos.X = tx
	pos.Y = ty
	if dx != 0 || dy != 0 {
		face.DX, face.DY = signInt(dx), signInt(dy)
	}
	return true
}

func (s *MovementApplySystem) blockedCell(x, y int) bool {
	if s.tileFilter == nil {
		return false
	}
	q := s.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, _, _, _, solid := q.Get()
		if pos.X == x && pos.Y == y && solid.Blocks {
			return true
		}
	}
	return false
}
