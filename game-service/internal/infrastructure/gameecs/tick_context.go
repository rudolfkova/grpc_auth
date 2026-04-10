package gameecs

import (
	"game/internal/domain/models"

	"github.com/mlange-42/ark/ecs"
)

// PlayerEntitySink — спавн/получение ECS-сущности игрока по user_id (внутренний контракт систем).
type PlayerEntitySink interface {
	EnsurePlayerEntity(userID int64) ecs.Entity
}

// TickContext передаётся в System.Update.
type TickContext struct {
	Sink          PlayerEntitySink
	CurrentAction models.Action
	Players       []models.Player
}
