package service

import (
	"context"
	"errors"
	"strings"

	"world/internal/model"
	"world/internal/repository"
)

// World сервис поверх репозитория.
type World struct {
	repo repository.WorldRepository
}

// NewWorld ...
func NewWorld(repo repository.WorldRepository) *World {
	return &World{repo: repo}
}

func (s *World) Create(ctx context.Context, w model.World) error {
	if w.ID == "" {
		return ErrInvalidArgument
	}
	w.Name = strings.TrimSpace(w.Name)
	if w.SchemaVersion == 0 {
		w.SchemaVersion = 1
	}
	err := s.repo.Create(ctx, w)
	if errors.Is(err, repository.ErrAlreadyExists) {
		return ErrAlreadyExists
	}
	return err
}

func (s *World) Get(ctx context.Context, id string) (model.World, error) {
	if id == "" {
		return model.World{}, ErrInvalidArgument
	}
	w, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return model.World{}, ErrNotFound
	}
	return w, err
}

func (s *World) GetByName(ctx context.Context, name string) (model.World, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return model.World{}, ErrInvalidArgument
	}
	w, err := s.repo.GetByName(ctx, name)
	if errors.Is(err, repository.ErrNotFound) {
		return model.World{}, ErrNotFound
	}
	return w, err
}

func (s *World) ReplaceSnapshot(ctx context.Context, id string, snapshot []byte, schemaVersion int32, expectedVersion int64) (model.World, error) {
	if id == "" {
		return model.World{}, ErrInvalidArgument
	}
	w, err := s.repo.ReplaceSnapshot(ctx, id, snapshot, schemaVersion, expectedVersion)
	if errors.Is(err, repository.ErrNotFound) {
		return model.World{}, ErrNotFound
	}
	if errors.Is(err, repository.ErrVersionMismatch) {
		return model.World{}, ErrVersionMismatch
	}
	return w, err
}

func (s *World) Delete(ctx context.Context, id string) error {
	if id == "" {
		return ErrInvalidArgument
	}
	err := s.repo.Delete(ctx, id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *World) List(ctx context.Context, limit, offset int32) ([]model.World, error) {
	return s.repo.List(ctx, limit, offset)
}
