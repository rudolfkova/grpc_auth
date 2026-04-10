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
	TypeState     = "state"
	TypeReject    = "reject"
	TypeError     = "error"
)

// Envelope — обёртка WebSocket JSON (клиент ↔ game-service).
type Envelope struct {
	Service string          `json:"service"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// MoveIntent — payload для TypeMove.
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

// Player — элемент массива players в payload события TypeState.
type Player struct {
	ID int64 `json:"id"`
	X  int   `json:"x"`
	Y  int   `json:"y"`
	HP int   `json:"hp"`
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

// StatePayload — форма payload у TypeState (для json.Unmarshal на клиенте).
type StatePayload struct {
	Players []Player  `json:"players"`
	Tiles   []Tile    `json:"tiles"`
	TickAt  time.Time `json:"tick_at"`
}

// NormalizeTileRotationQuarter приводит произвольное целое к диапазону 0..3 (четверти оборота по часовой стрелке).
func NormalizeTileRotationQuarter(r int) int {
	return ((r % 4) + 4) % 4
}
