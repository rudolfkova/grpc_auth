package worldclient

import (
	"context"
	"strings"

	"game/internal/domain/ports"
)

// StoreAdapter — реализация ports.WorldStore поверх gRPC worldclient.
type StoreAdapter struct {
	addr  string
	token string
}

func NewStoreAdapter(addr, token string) *StoreAdapter {
	return &StoreAdapter{
		addr:  strings.TrimSpace(addr),
		token: token,
	}
}

func (a *StoreAdapter) SaveWorldByName(ctx context.Context, req ports.SaveWorldRequest) (*ports.SaveWorldResponse, error) {
	res, err := SaveWorldByName(ctx, a.addr, a.token, req.Name, req.Description, req.Snapshot, req.SchemaVersion)
	if err != nil {
		return nil, err
	}
	return &ports.SaveWorldResponse{
		WorldID: res.WorldID,
		Version: res.Version,
	}, nil
}
