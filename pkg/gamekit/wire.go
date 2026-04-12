package gamekit

import (
	"encoding/json"
	"time"
)

// Имена сервиса и типов сообщений по wire (договорённость с game-service).
const (
	ServiceGame = "game"

	TypeMove      = "move"
	TypeHit       = "hit"
	TypeSpawnTile = "spawn_tile"
	TypeClearTile = "clear_tile"
	TypeSaveWorld = "save_world"
	TypeState     = "state"
	TypeReject    = "reject"
	TypeError     = "error"
	// TypeSaveWorldResult — ответ на save_world (только инициатору).
	TypeSaveWorldResult = "save_world_result"
)

// SnapshotSchemaVersion — schema_version для снимка ark-serde, который пишет game-service в world-service.
const SnapshotSchemaVersion int32 = 1

// Envelope — обёртка WebSocket JSON (клиент ↔ game-service).
type Envelope struct {
	Service string          `json:"service"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// MoveIntent — payload для TypeMove: обновляет «удерживаемое» направление (после clamp к Speed.MaxStep).
// Физический шаг по сетке выполняется не чаще одного раза за игровой тик на сервере; частые сообщения только обновляют интент.
// Отправьте dx=0, dy=0, чтобы сбросить движение.
type MoveIntent struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

// HitIntent — payload для TypeHit.
type HitIntent struct {
	TargetID int64 `json:"target_id"`
	Damage   int   `json:"damage"`
}

// TileSpawnIntent — payload для TypeSpawnTile.
// Layer по умолчанию 0; Rotation — четверти оборота по часовой стрелке (любое целое нормализуется к 0..3).
type TileSpawnIntent struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Layer    int    `json:"layer"`
	Rotation int    `json:"rotation"`
	Texture  string `json:"texture"`
	Blocks   bool   `json:"blocks"`
}

// TileClearIntent — payload для TypeClearTile: удалить все тайлы в клетке (x,y) на указанном слое.
type TileClearIntent struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Layer int `json:"layer"`
}

// SaveWorldIntent — payload для TypeSaveWorld (только разрешённый admin user_id на сервере).
type SaveWorldIntent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SaveWorldResultPayload — payload для TypeSaveWorldResult.
type SaveWorldResultPayload struct {
	Ok      bool   `json:"ok"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	WorldID string `json:"world_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version int64  `json:"version,omitempty"`
}

// Player — проекция игрока в payload события TypeState (см. StatePayload.Players).
// Тот же тип используется в game-service при Broadcast; дублировать поля в других пакетах не нужно.
type Player struct {
	ID     int64          `json:"id"`
	X      int            `json:"x"`
	Y      int            `json:"y"`
	HP     int            `json:"hp"`
	FaceDX int            `json:"face_dx"`
	FaceDY int            `json:"face_dy"`
	Stats  CharacterStats `json:"stats"`
	// Sprite — id листа ходьбы (как CharacterPlayData.Sprite); клиент: data/anim/<sprite>/<sprite>.png.
	Sprite string `json:"sprite"`
}

// Tile — элемент массива tiles в payload события TypeState.
type Tile struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Layer    int    `json:"layer"`
	Rotation int    `json:"rotation"`
	Texture  string `json:"texture"`
	Blocks   bool   `json:"blocks"`
}

// StatePayload — полный JSON payload у TypeState (сервер шлёт это же из gamekit; клиент Unmarshal сюда).
type StatePayload struct {
	Players []Player  `json:"players"`
	Tiles   []Tile    `json:"tiles"`
	TickAt  time.Time `json:"tick_at"`
}

// NormalizeTileRotationQuarter приводит произвольное целое к диапазону 0..3 (четверти оборота по часовой стрелке).
func NormalizeTileRotationQuarter(r int) int {
	return ((r % 4) + 4) % 4
}
