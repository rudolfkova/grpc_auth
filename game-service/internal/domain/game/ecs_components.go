package game

// ECS component types for github.com/mlange-42/ark.
// Один архетип игрока: PlayerRef + GridPos + Speed + Health.

// PlayerRef связывает сущность с user_id из JWT / транспорта.
type PlayerRef struct {
	UserID int64
}

// GridPos — целочисленная позиция на сетке мира.
type GridPos struct {
	X, Y int
}

// Speed ограничивает величину шага по каждой оси за одно действие move (-MaxStep..MaxStep).
// По умолчанию 1 (как раньше dx,dy ∈ {-1,0,1}).
type Speed struct {
	MaxStep int
}

// Health — текущие HP.
type Health struct {
	HP int
}
