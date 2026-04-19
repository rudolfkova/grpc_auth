package ports

import (
	"context"

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
	// JoinStateSnapshot — полный список tiles + players для немедленной синхронизации при WS connect (без ожидания глобального full-sync).
	JoinStateSnapshot() gamekit.StatePayload
}

// SaveWorldRequest — входной контракт сохранения snapshot в world-service.
type SaveWorldRequest struct {
	Name          string
	Description   string
	Snapshot      []byte
	SchemaVersion int32
}

// SaveWorldResponse — минимальный результат успешного сохранения мира.
type SaveWorldResponse struct {
	WorldID string
	Version int64
}

// WorldStore — исходящий порт приложения для операций с world-service.
type WorldStore interface {
	SaveWorldByName(ctx context.Context, req SaveWorldRequest) (*SaveWorldResponse, error)
}

// CharacterIdentity — минимум данных персонажа для persist-сессии.
type CharacterIdentity struct {
	ID          string
	DisplayName string
	Description string
}

// ResolveCharacterResponse — результат resolve на старте игровой сессии.
type ResolveCharacterResponse struct {
	Persisted bool
	Character CharacterIdentity
	Data      []byte
}

// SaveCharacterSessionRequest — данные для записи character.data после игровой сессии.
type SaveCharacterSessionRequest struct {
	UserID        int64
	Persisted     bool
	Character     CharacterIdentity
	Data          []byte
	SchemaVersion int32
}

// CharacterSessions — исходящий порт приложения для интеграции с character-service.
type CharacterSessions interface {
	ResolvePlayCharacter(ctx context.Context, userID int64, characterID string) (*ResolveCharacterResponse, error)
	SavePlaySession(ctx context.Context, req SaveCharacterSessionRequest) error
}
