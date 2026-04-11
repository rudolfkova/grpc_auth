package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"

	"character/internal/model"
	"character/internal/repository"
)

// Character — доменная логика персонажей.
type Character struct {
	repo repository.CharacterRepository
}

// NewCharacter ...
func NewCharacter(repo repository.CharacterRepository) *Character {
	return &Character{repo: repo}
}

func validOwner(ownerUserID int64) bool {
	return ownerUserID > 0
}

// ResolveForPlay — есть строка в БД или «пустой» шаблон с тем же id (игра потом вызовет CreateCharacter при первом сохранении).
func (s *Character) ResolveForPlay(ctx context.Context, ownerUserID int64, characterID string) (c model.Character, persisted bool, err error) {
	characterID = strings.TrimSpace(characterID)
	if !validOwner(ownerUserID) {
		return model.Character{}, false, fmt.Errorf("%w: user_id must be > 0", ErrInvalidArgument)
	}
	if characterID == "" {
		return model.Character{}, false, fmt.Errorf("%w: id (character uuid) is required", ErrInvalidArgument)
	}
	c, err = s.repo.GetByIDAndOwner(ctx, characterID, ownerUserID)
	if err == nil {
		return c, true, nil
	}
	if errors.Is(err, repository.ErrNotFound) {
		return model.Character{
			ID:          characterID,
			OwnerUserID: ownerUserID,
		}, false, nil
	}
	return model.Character{}, false, err
}

func (s *Character) Get(ctx context.Context, ownerUserID int64, characterID string) (model.Character, error) {
	characterID = strings.TrimSpace(characterID)
	if !validOwner(ownerUserID) || characterID == "" {
		return model.Character{}, ErrInvalidArgument
	}
	c, err := s.repo.GetByIDAndOwner(ctx, characterID, ownerUserID)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Character{}, ErrNotFound
	}
	return c, err
}

func (s *Character) GetByDisplayName(ctx context.Context, ownerUserID int64, displayName string) (model.Character, error) {
	displayName = strings.TrimSpace(displayName)
	if !validOwner(ownerUserID) || displayName == "" {
		return model.Character{}, ErrInvalidArgument
	}
	c, err := s.repo.GetByDisplayNameAndOwner(ctx, ownerUserID, displayName)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Character{}, ErrNotFound
	}
	return c, err
}

func (s *Character) List(ctx context.Context, ownerUserID int64, limit, offset int32) ([]model.Character, error) {
	if !validOwner(ownerUserID) {
		return nil, ErrInvalidArgument
	}
	return s.repo.ListByOwner(ctx, ownerUserID, limit, offset)
}

func (s *Character) Create(ctx context.Context, ownerUserID int64, characterID, displayName, description string, data []byte, schemaVersion int32) (model.Character, error) {
	if !validOwner(ownerUserID) {
		return model.Character{}, ErrInvalidArgument
	}
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return model.Character{}, ErrInvalidArgument
	}
	characterID = strings.TrimSpace(characterID)
	if characterID == "" {
		characterID = uuid.NewString()
	}
	if schemaVersion == 0 {
		schemaVersion = 1
	}
	c := model.Character{
		ID:             characterID,
		OwnerUserID:    ownerUserID,
		DisplayName:    displayName,
		Description:    strings.TrimSpace(description),
		Data:           data,
		SchemaVersion:  schemaVersion,
	}
	if c.Data == nil {
		c.Data = []byte{}
	}
	err := s.repo.Create(ctx, c)
	if errors.Is(err, repository.ErrAlreadyExists) {
		return model.Character{}, ErrAlreadyExists
	}
	if err != nil {
		return model.Character{}, err
	}
	return s.repo.GetByIDAndOwner(ctx, characterID, ownerUserID)
}

func (s *Character) ReplaceData(ctx context.Context, ownerUserID int64, characterID string, data []byte, schemaVersion int32, expectedVersion int64) (model.Character, error) {
	characterID = strings.TrimSpace(characterID)
	if !validOwner(ownerUserID) || characterID == "" {
		return model.Character{}, ErrInvalidArgument
	}
	c, err := s.repo.ReplaceData(ctx, characterID, ownerUserID, data, schemaVersion, expectedVersion)
	if errors.Is(err, repository.ErrNotFound) {
		return model.Character{}, ErrNotFound
	}
	if errors.Is(err, repository.ErrVersionMismatch) {
		return model.Character{}, ErrVersionMismatch
	}
	return c, err
}

func (s *Character) Delete(ctx context.Context, ownerUserID int64, characterID string) error {
	characterID = strings.TrimSpace(characterID)
	if !validOwner(ownerUserID) || characterID == "" {
		return ErrInvalidArgument
	}
	err := s.repo.Delete(ctx, characterID, ownerUserID)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
