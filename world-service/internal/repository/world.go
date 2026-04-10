package repository

import (
	"context"
	"errors"

	"world/internal/model"
)

var (
	ErrNotFound         = errors.New("world not found")
	ErrAlreadyExists    = errors.New("world already exists")
	ErrVersionMismatch  = errors.New("world version mismatch")
)

// WorldRepository хранит снапшоты миров.
type WorldRepository interface {
	Create(ctx context.Context, w model.World) error
	GetByID(ctx context.Context, id string) (model.World, error)
	GetByName(ctx context.Context, name string) (model.World, error)
	ReplaceSnapshot(ctx context.Context, id string, snapshot []byte, schemaVersion int32, expectedVersion int64) (model.World, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int32) ([]model.World, error)
}
