package gameecs

import (
	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// PlayerEntitySink — спавн/получение ECS-сущности игрока по user_id (внутренний контракт систем).
type PlayerEntitySink interface {
	EnsurePlayerEntity(userID int64) ecs.Entity
}

// MoveIntentStore — последнее намерение движения по user_id; между тиками сохраняется.
// Сообщения move только обновляют store; один физический шаг за тик делает MovementApplySystem.
type MoveIntentStore struct {
	byUser map[int64]struct{ DX, DY int }
}

// Set записывает желаемую дельту за один игровой тик (уже ограниченную Speed.MaxStep при записи).
// (0,0) сбрасывает интент (клиент «отпустил» направление).
func (s *MoveIntentStore) Set(uid int64, dx, dy int) {
	if s.byUser == nil {
		s.byUser = make(map[int64]struct{ DX, DY int })
	}
	if dx == 0 && dy == 0 {
		delete(s.byUser, uid)
		return
	}
	s.byUser[uid] = struct{ DX, DY int }{DX: dx, DY: dy}
}

func (s *MoveIntentStore) Get(uid int64) (dx, dy int) {
	v, ok := s.byUser[uid]
	if !ok {
		return 0, 0
	}
	return v.DX, v.DY
}

// TickContext передаётся в System.Update.
type TickContext struct {
	Sink          PlayerEntitySink
	Intents       *MoveIntentStore
	// ApplyMovement — в этом тике выполнять шаг по move-интенту (см. movement_apply_every_n_ticks в конфиге).
	ApplyMovement bool
	// Diag — чередование осей для диагонального интента (лесенка); nil = без разбиения.
	Diag *diagStrideState
	CurrentAction models.Action
	Players       []gamekit.Player
	Tiles         []gamekit.Tile
}
