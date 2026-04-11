package repository

import (
	"context"
	"errors"

	"character/internal/model"
)

var (
	ErrNotFound        = errors.New("character not found")
	ErrAlreadyExists   = errors.New("character already exists")
	ErrVersionMismatch = errors.New("character version mismatch")
)

// CharacterRepository — доступ к персонажам с фильтром по владельцу.
type CharacterRepository interface {
	Create(ctx context.Context, c model.Character) error
	GetByIDAndOwner(ctx context.Context, id string, ownerUserID int64) (model.Character, error)
	GetByDisplayNameAndOwner(ctx context.Context, ownerUserID int64, displayName string) (model.Character, error)
	ListByOwner(ctx context.Context, ownerUserID int64, limit, offset int32) ([]model.Character, error)
	ReplaceData(ctx context.Context, id string, ownerUserID int64, data []byte, schemaVersion int32, expectedVersion int64) (model.Character, error)
	Delete(ctx context.Context, id string, ownerUserID int64) error
}
