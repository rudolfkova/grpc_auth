package gameecs

import (
	"fmt"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	arkserde "github.com/mlange-42/ark-serde"
	"github.com/mlange-42/ark/ecs"
)

// applyArkWorldSnapshot десериализует снимок через ark-serde в уже подготовленный пустой World
// (компоненты зарегистрированы через NewMap5 для игроков). После загрузки восстанавливается индекс byUser.
func (e *Engine) applyArkWorldSnapshot(snapshot []byte) error {
	if len(snapshot) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if err := arkserde.Deserialize(snapshot, e.world); err != nil {
		return fmt.Errorf("ark Deserialize: %w", err)
	}
	e.migrateLegacyPlayerStatsLocked()
	return e.rebuildPlayerIndexLocked()
}

// migrateLegacyPlayerStatsLocked добавляет CharacterStats к игрокам из старых снимков мира (5 компонентов).
func (e *Engine) migrateLegacyPlayerStatsLocked() {
	f5 := ecs.NewFilter5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace](e.world)
	q := f5.Query()
	defer q.Close()
	for q.Next() {
		ent := q.Entity()
		if e.playerMapper.HasAll(ent) {
			continue
		}
		ref, pos, sp, hp, face := q.Get()
		r, p, s, h, f := *ref, *pos, *sp, *hp, *face
		st := gamekit.DefaultCharacterStats()
		e.playerMapper.Add(ent, &r, &p, &s, &h, &f, &st)
	}
}

func (e *Engine) rebuildPlayerIndexLocked() error {
	clear(e.byUser)

	filt := ecs.NewFilter6[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats](e.world)
	q := filt.Query()
	defer q.Close()

	for q.Next() {
		ref, _, _, _, _, _ := q.Get()
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
