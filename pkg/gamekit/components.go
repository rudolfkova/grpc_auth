// Package gamekit — общие ECS-компоненты и контракт WebSocket game-service (сервер + клиент на Go/Ark).
package gamekit

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

// TileTexture — имя текстуры на клиенте (ассет).
type TileTexture struct {
	Name string `json:"name"`
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
