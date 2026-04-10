package world

// Компоненты игрока для ECS-слоя (поля — чистая модель, без зависимости от Ark).

// PlayerRef связывает сущность с user_id из JWT / транспорта.
type PlayerRef struct {
	UserID int64
}

// GridPos — целочисленная позиция на сетке мира.
type GridPos struct {
	X, Y int
}

// Speed ограничивает величину шага по каждой оси за одно действие move (-MaxStep..MaxStep).
type Speed struct {
	MaxStep int
}

// Health — текущие HP.
type Health struct {
	HP int
}

// DefaultPlayerHP — стартовое здоровье при спавне игрока.
const DefaultPlayerHP = 10
