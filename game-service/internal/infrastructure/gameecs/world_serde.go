package gameecs

import (
	"fmt"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	arkserde "github.com/mlange-42/ark-serde"
	"github.com/mlange-42/ark/ecs"
)

// applyArkWorldSnapshot десериализует снимок через ark-serde в пустой World (регистрация компонентов — NewMap7 игроков).
// После Deserialize — миграция старых снимков (sprite / stats), затем индекс byUser.
func (e *Engine) applyArkWorldSnapshot(snapshot []byte) error {
	if len(snapshot) == 0 {
		return nil
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	if err := arkserde.Deserialize(snapshot, e.world); err != nil {
		return fmt.Errorf("ark Deserialize: %w", err)
	}
	// Сначала 6→7 (есть stats, нет sprite), затем 5→7 (нет stats и sprite).
	e.migrateLegacyPlayerSpriteLocked()
	e.migrateLegacyPlayerStatsLocked()
	return e.rebuildPlayerIndexLocked()
}

// migrateLegacyPlayerSpriteLocked — снимок с CharacterStats, но без PlayerSprite (старый формат после stats).
func (e *Engine) migrateLegacyPlayerSpriteLocked() {
	f6 := ecs.NewFilter6[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats](e.world)
	q := f6.Query()
	defer q.Close()
	for q.Next() {
		ent := q.Entity()
		if e.playerMapper.HasAll(ent) {
			continue
		}
		ref, pos, sp, hp, face, st := q.Get()
		r, p, s, h, f, stv := *ref, *pos, *sp, *hp, *face, *st
		spr := gamekit.PlayerSprite{Name: gamekit.DefaultPlayerSprite}
		e.playerMapper.Add(ent, &r, &p, &s, &h, &f, &stv, &spr)
	}
}

// migrateLegacyPlayerStatsLocked — снимок только с пятью компонентами игрока (без stats и sprite).
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
		spr := gamekit.PlayerSprite{Name: gamekit.DefaultPlayerSprite}
		e.playerMapper.Add(ent, &r, &p, &s, &h, &f, &st, &spr)
	}
}

func (e *Engine) rebuildPlayerIndexLocked() error {
	clear(e.byUser)

	filt := ecs.NewFilter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite](e.world)
	q := filt.Query()
	defer q.Close()

	for q.Next() {
		ref, _, _, _, _, _, _ := q.Get()
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
