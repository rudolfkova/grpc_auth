package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"character/internal/model"
	"character/internal/repository"

	"github.com/lib/pq"
)

// CharacterRepo реализует repository.CharacterRepository.
type CharacterRepo struct {
	db *sql.DB
}

// NewCharacterRepo ...
func NewCharacterRepo(s *Store) *CharacterRepo {
	return &CharacterRepo{db: s.DB}
}

func (r *CharacterRepo) Create(ctx context.Context, c model.Character) error {
	if c.ID == "" {
		return fmt.Errorf("empty id")
	}
	if c.Data == nil {
		c.Data = []byte{}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO characters (id, owner_user_id, display_name, description, data, schema_version, version)
		VALUES ($1, $2, $3, $4, $5, $6, 1)
	`, c.ID, c.OwnerUserID, c.DisplayName, c.Description, c.Data, c.SchemaVersion)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *CharacterRepo) GetByIDAndOwner(ctx context.Context, id string, ownerUserID int64) (model.Character, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, display_name, description, data, schema_version, version, created_at, updated_at
		FROM characters WHERE id = $1 AND owner_user_id = $2
	`, id, ownerUserID)
	return scanCharacter(row)
}

func (r *CharacterRepo) GetByDisplayNameAndOwner(ctx context.Context, ownerUserID int64, displayName string) (model.Character, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, owner_user_id, display_name, description, data, schema_version, version, created_at, updated_at
		FROM characters WHERE owner_user_id = $1 AND display_name = $2
	`, ownerUserID, displayName)
	return scanCharacter(row)
}

func scanCharacter(row *sql.Row) (model.Character, error) {
	var c model.Character
	err := row.Scan(&c.ID, &c.OwnerUserID, &c.DisplayName, &c.Description, &c.Data, &c.SchemaVersion, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.Character{}, repository.ErrNotFound
	}
	if err != nil {
		return model.Character{}, err
	}
	return c, nil
}

func (r *CharacterRepo) ListByOwner(ctx context.Context, ownerUserID int64, limit, offset int32) ([]model.Character, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 500 {
		limit = 500
	}
	if offset < 0 {
		offset = 0
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_user_id, display_name, description, data, schema_version, version, created_at, updated_at
		FROM characters
		WHERE owner_user_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3
	`, ownerUserID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []model.Character
	for rows.Next() {
		var c model.Character
		if err := rows.Scan(&c.ID, &c.OwnerUserID, &c.DisplayName, &c.Description, &c.Data, &c.SchemaVersion, &c.Version, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *CharacterRepo) ReplaceData(ctx context.Context, id string, ownerUserID int64, data []byte, schemaVersion int32, expectedVersion int64) (model.Character, error) {
	if data == nil {
		data = []byte{}
	}
	var c model.Character
	err := r.db.QueryRowContext(ctx, `
		UPDATE characters
		SET data = $3,
		    schema_version = $4,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1 AND owner_user_id = $2
		  AND ($5 = 0 OR version = $5)
		RETURNING id, owner_user_id, display_name, description, data, schema_version, version, created_at, updated_at
	`, id, ownerUserID, data, schemaVersion, expectedVersion).Scan(
		&c.ID, &c.OwnerUserID, &c.DisplayName, &c.Description, &c.Data, &c.SchemaVersion, &c.Version, &c.CreatedAt, &c.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		_, getErr := r.GetByIDAndOwner(ctx, id, ownerUserID)
		if errors.Is(getErr, repository.ErrNotFound) {
			return model.Character{}, repository.ErrNotFound
		}
		return model.Character{}, repository.ErrVersionMismatch
	}
	if err != nil {
		return model.Character{}, err
	}
	return c, nil
}

func (r *CharacterRepo) Delete(ctx context.Context, id string, ownerUserID int64) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM characters WHERE id = $1 AND owner_user_id = $2`, id, ownerUserID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return repository.ErrNotFound
	}
	return nil
}
