package ports

import "game/internal/domain/models"

// GameEngine — входной порт прикладного слоя: симуляция одного тика по действиям игроков.
// Реализация живёт в infrastructure (ECS), приложение зависит только от этого интерфейса.
type GameEngine interface {
	ProcessTick(actions []models.Action) []models.Event
}
