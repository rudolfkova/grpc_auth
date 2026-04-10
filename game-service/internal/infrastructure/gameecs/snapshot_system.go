package gameecs

import (
	"game/internal/domain/models"
	"game/internal/domain/world"

	"github.com/mlange-42/ark/ecs"
)

// SnapshotSystem строит снимок всех игроков через Ark Filter.
type SnapshotSystem struct {
	filter *ecs.Filter4[world.PlayerRef, world.GridPos, world.Speed, world.Health]
}

// NewSnapshotSystem создаёт систему снимка.
func NewSnapshotSystem(filter *ecs.Filter4[world.PlayerRef, world.GridPos, world.Speed, world.Health]) *SnapshotSystem {
	return &SnapshotSystem{filter: filter}
}

func (s *SnapshotSystem) Update(ctx *TickContext) {
	q := s.filter.Query()
	defer q.Close()

	out := make([]models.Player, 0, 64)
	for q.Next() {
		ref, pos, _, hp := q.Get()
		out = append(out, models.Player{
			ID: ref.UserID,
			X:  pos.X,
			Y:  pos.Y,
			HP: hp.HP,
		})
	}
	ctx.Players = out
}
