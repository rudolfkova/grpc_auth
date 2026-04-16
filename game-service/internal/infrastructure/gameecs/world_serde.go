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
	e.migratePlayerGearLocked()
	return e.rebuildPlayerIndexLocked()
}

// migrateLegacyPlayerSpriteLocked — снимок с CharacterStats, но без PlayerSprite (старый формат после stats).
func (e *Engine) migrateLegacyPlayerSpriteLocked() {
	f6 := ecs.NewFilter6[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats](e.world)
	q := f6.Query()
	type spriteRow struct {
		ent   ecs.Entity
		ref   gamekit.PlayerRef
		pos   gamekit.GridPos
		sp    gamekit.Speed
		hp    gamekit.Health
		face  gamekit.PlayerFace
		stats gamekit.CharacterStats
	}
	var rows []spriteRow
	for q.Next() {
		ent := q.Entity()
		if e.playerMapper.HasAll(ent) {
			continue
		}
		ref, pos, sp, hp, face, st := q.Get()
		rows = append(rows, spriteRow{
			ent: ent, ref: *ref, pos: *pos, sp: *sp, hp: *hp, face: *face, stats: *st,
		})
	}
	q.Close()
	for _, rw := range rows {
		spr := gamekit.PlayerSprite{Name: gamekit.DefaultPlayerSprite}
		r, p, s, h, f, stv := rw.ref, rw.pos, rw.sp, rw.hp, rw.face, rw.stats
		e.playerMapper.Add(rw.ent, &r, &p, &s, &h, &f, &stv, &spr)
	}
}

// migrateLegacyPlayerStatsLocked — снимок только с пятью компонентами игрока (без stats и sprite).
func (e *Engine) migrateLegacyPlayerStatsLocked() {
	f5 := ecs.NewFilter5[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace](e.world)
	q := f5.Query()
	type statsRow struct {
		ent  ecs.Entity
		ref  gamekit.PlayerRef
		pos  gamekit.GridPos
		sp   gamekit.Speed
		hp   gamekit.Health
		face gamekit.PlayerFace
	}
	var rows []statsRow
	for q.Next() {
		ent := q.Entity()
		if e.playerMapper.HasAll(ent) {
			continue
		}
		ref, pos, sp, hp, face := q.Get()
		rows = append(rows, statsRow{ent: ent, ref: *ref, pos: *pos, sp: *sp, hp: *hp, face: *face})
	}
	q.Close()
	for _, rw := range rows {
		r, p, s, h, f := rw.ref, rw.pos, rw.sp, rw.hp, rw.face
		st := gamekit.DefaultCharacterStats()
		spr := gamekit.PlayerSprite{Name: gamekit.DefaultPlayerSprite}
		e.playerMapper.Add(rw.ent, &r, &p, &s, &h, &f, &st, &spr)
	}
}

func (e *Engine) migratePlayerGearLocked() {
	filt := ecs.NewFilter7[gamekit.PlayerRef, gamekit.GridPos, gamekit.Speed, gamekit.Health, gamekit.PlayerFace, gamekit.CharacterStats, gamekit.PlayerSprite](e.world)
	q := filt.Query()
	var needGear []ecs.Entity
	for q.Next() {
		ent := q.Entity()
		if e.playerGearMapper.HasAll(ent) {
			continue
		}
		needGear = append(needGear, ent)
	}
	q.Close()
	for _, ent := range needGear {
		inv := gamekit.DefaultPlayerInventory()
		e.playerGearMapper.Add(ent, &inv)
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
