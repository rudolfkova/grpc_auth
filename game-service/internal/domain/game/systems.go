package game

import (
	"game/internal/domain/models"

	"github.com/mlange-42/ark/ecs"
)

// runStateSnapshotQuery обходит все сущности-игроков через Ark Filter (query).
// Вызывается в конце тика после применения действий; порядок сущностей не гарантирован.
func (e *Engine) runStateSnapshotQuery() []models.Player {
	q := e.playerFilter.Query()
	defer q.Close()

	out := make([]models.Player, 0, len(e.byUser))
	for q.Next() {
		ref, pos, _, hp := q.Get()
		out = append(out, models.Player{
			ID: ref.UserID,
			X:  pos.X,
			Y:  pos.Y,
			HP: hp.HP,
		})
	}
	return out
}

// runMovementStep применяет одно намерение движения к сущности игрока (логика «системы движения»).
// dx, dy должны укладываться в [-Speed.MaxStep, Speed.MaxStep] по каждой оси; иначе шаг отбрасывается.
// Вызывается из dispatch по порядку действий в тике, чтобы сохранить семантику interleaving.
func (e *Engine) runMovementStep(ent ecs.Entity, dx, dy int) {
	if !e.playerMapper.HasAll(ent) {
		return
	}
	_, pos, speed, _ := e.playerMapper.Get(ent)
	step := speed.MaxStep
	if step <= 0 {
		return
	}
	if dx < -step || dx > step || dy < -step || dy > step {
		return
	}
	pos.X += dx
	pos.Y += dy
}

// runDamageStep наносит урон по HP цели (логика «системы урона»).
func (e *Engine) runDamageStep(targetEnt ecs.Entity, damage int) {
	if damage <= 0 || !e.playerMapper.HasAll(targetEnt) {
		return
	}
	_, _, _, hp := e.playerMapper.Get(targetEnt)
	hp.HP -= damage
	if hp.HP < 0 {
		hp.HP = 0
	}
}
