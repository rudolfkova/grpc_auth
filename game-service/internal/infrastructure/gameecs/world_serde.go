package gameecs

import (
	"fmt"

	"game/internal/domain/world"

	arkserde "github.com/mlange-42/ark-serde"
	"github.com/mlange-42/ark/ecs"
)

// applyArkWorldSnapshot десериализует снимок через ark-serde в уже подготовленный пустой World
// (компоненты зарегистрированы через NewMap4). После загрузки восстанавливается индекс byUser.
func (e *Engine) applyArkWorldSnapshot(snapshot []byte) error {
	if len(snapshot) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if err := arkserde.Deserialize(snapshot, e.world); err != nil {
		return fmt.Errorf("ark Deserialize: %w", err)
	}
	return e.rebuildPlayerIndexLocked()
}

func (e *Engine) rebuildPlayerIndexLocked() error {
	clear(e.byUser)

	filt := ecs.NewFilter4[world.PlayerRef, world.GridPos, world.Speed, world.Health](e.world)
	q := filt.Query()
	defer q.Close()

	for q.Next() {
		ref, _, _, _ := q.Get()
		if ref.UserID == 0 {
			return fmt.Errorf("world snapshot: entity with PlayerRef.UserID == 0")
		}
		if _, dup := e.byUser[ref.UserID]; dup {
			return fmt.Errorf("world snapshot: duplicate PlayerRef.UserID %d", ref.UserID)
		}
		e.byUser[ref.UserID] = q.Entity()
	}
	return nil
}
