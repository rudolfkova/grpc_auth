package ports

import (
	"game/internal/domain/models"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// GameEngine — входной порт прикладного слоя: симуляция одного тика по действиям игроков.
// Реализация живёт в infrastructure (ECS), приложение зависит только от этого интерфейса.
type GameEngine interface {
	ProcessTick(actions []models.Action) []models.Event
	// SerializeWorld — снимок ark-serde текущего ECS-мира (для сохранения в world-service).
	SerializeWorld() ([]byte, error)
	// EnsurePlayerJoin — создать игрока из character.data при первом подключении user_id; повторные вызовы не сбрасывают состояние.
	EnsurePlayerJoin(userID int64, d gamekit.CharacterPlayData)
	// PlayerCharacterData — JSON v1 для opaque character.data (позиция, HP, взгляд).
	PlayerCharacterData(userID int64) ([]byte, error)
}
