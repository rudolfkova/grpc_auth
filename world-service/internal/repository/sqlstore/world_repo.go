package sqlstore

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"world/internal/model"
	"world/internal/repository"

	"github.com/lib/pq"
)

// WorldRepo реализует repository.WorldRepository.
type WorldRepo struct {
	db *sql.DB
}

// NewWorldRepo ...
func NewWorldRepo(s *Store) *WorldRepo {
	return &WorldRepo{db: s.DB}
}

func (r *WorldRepo) Create(ctx context.Context, w model.World) error {
	if w.ID == "" {
		return fmt.Errorf("empty id")
	}
	if w.Snapshot == nil {
		w.Snapshot = []byte{}
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO worlds (id, name, description, snapshot, schema_version, version)
		VALUES ($1, $2, $3, $4, $5, 1)
	`, w.ID, w.Name, w.Description, w.Snapshot, w.SchemaVersion)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return repository.ErrAlreadyExists
		}
		return err
	}
	return nil
}

func (r *WorldRepo) GetByID(ctx context.Context, id string) (model.World, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, description, snapshot, schema_version, version, created_at, updated_at
		FROM worlds WHERE id = $1
	`, id)
	var w model.World
	err := row.Scan(&w.ID, &w.Name, &w.Description, &w.Snapshot, &w.SchemaVersion, &w.Version, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return model.World{}, repository.ErrNotFound
	}
	if err != nil {
		return model.World{}, err
	}
	return w, nil
}

func (r *WorldRepo) ReplaceSnapshot(ctx context.Context, id string, snapshot []byte, schemaVersion int32, expectedVersion int64) (model.World, error) {
	if snapshot == nil {
		snapshot = []byte{}
	}
	var w model.World
	err := r.db.QueryRowContext(ctx, `
		UPDATE worlds
		SET snapshot = $2,
		    schema_version = $3,
		    version = version + 1,
		    updated_at = now()
		WHERE id = $1
		  AND ($4 = 0 OR version = $4)
		RETURNING id, name, description, snapshot, schema_version, version, created_at, updated_at
	`, id, snapshot, schemaVersion, expectedVersion).Scan(
		&w.ID, &w.Name, &w.Description, &w.Snapshot, &w.SchemaVersion, &w.Version, &w.CreatedAt, &w.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		// либо нет строки, либо версия не совпала
		_, getErr := r.GetByID(ctx, id)
		if errors.Is(getErr, repository.ErrNotFound) {
			return model.World{}, repository.ErrNotFound
		}
		return model.World{}, repository.ErrVersionMismatch
	}
	if err != nil {
		return model.World{}, err
	}
	return w, nil
}

func (r *WorldRepo) Delete(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM worlds WHERE id = $1`, id)
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

func (r *WorldRepo) List(ctx context.Context, limit, offset int32) ([]model.World, error) {
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
		SELECT id, name, description, snapshot, schema_version, version, created_at, updated_at
		FROM worlds
		ORDER BY id
		LIMIT $1 OFFSET $2
	`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []model.World
	for rows.Next() {
		var w model.World
		if err := rows.Scan(&w.ID, &w.Name, &w.Description, &w.Snapshot, &w.SchemaVersion, &w.Version, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}
