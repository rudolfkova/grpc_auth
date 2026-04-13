// Package gamekit — общие ECS-компоненты и контракт WebSocket game-service (сервер + клиент на Go/Ark).
package gamekit

import "encoding/json"

// PlayerRef связывает сущность с user_id из JWT / транспорта.
type PlayerRef struct {
	UserID int64
}

// GridPos — целочисленная позиция на сетке (игроки и тайлы).
type GridPos struct {
	X, Y int
}

// Speed ограничивает шаг за одно действие move по каждой оси: [-MaxStep, MaxStep].
type Speed struct {
	MaxStep int
}

// Health — текущие HP.
type Health struct {
	HP int
}

// DefaultPlayerHP — стартовое здоровье при спавне игрока на сервере.
const DefaultPlayerHP = 10

// DefaultPlayerFaceDX, DefaultPlayerFaceDY — взгляд при спавне и в state, если в ECS ещё (0,0).
const (
	DefaultPlayerFaceDX = 1
	DefaultPlayerFaceDY = 0
)

// PlayerFace — направление «взгляда» / последнего успешного шага для анимаций (state.face_dx, face_dy).
// Значения в {-1, 0, 1} по осям; при спавне по умолчанию (1, 0) — на восток по экрану.
type PlayerFace struct {
	DX int `json:"face_dx"`
	DY int `json:"face_dy"`
}

// PlayerSprite — id набора ходьбы (state.sprite / character.data.sprite); клиент грузит anim/<Name>/...
type PlayerSprite struct {
	Name string `json:"name"`
}

// TileTexture — имя текстуры на клиенте (ассет).
// InstanceArgs — опциональный JSON-объект на экземпляре тайла (мержится в interact; wire: instance_args).
// Договорённость с клиентом: texture == item_def_id из каталога для резолва клика по клетке.
type TileTexture struct {
	Name          string          `json:"name"`
	InstanceArgs  json.RawMessage `json:"instance_args,omitempty"`
}

// TileSolid — Blocks=true: в клетку нельзя войти (сервер проверяет при move).
type TileSolid struct {
	Blocks bool `json:"blocks"`
}

// TileLayer — индекс слоя в клетке (x,y). Несколько сущностей с разными Z могут стоять в одной клетке.
type TileLayer struct {
	Z int `json:"layer"`
}

// TileFacing — ориентация тайла: поворот на плоскости, в четвертях оборота по часовой стрелке (0..3 → 0°, 90°, 180°, 270°).
type TileFacing struct {
	RotationQuarter int `json:"rotation"`
}
