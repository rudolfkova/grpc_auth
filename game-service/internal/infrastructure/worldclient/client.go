// Package worldclient — gRPC-клиент к world-service (GetWorld).
package worldclient

import (
	"context"
	"fmt"

	worldv1 "world/proto/world/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

// Fetched снимок и метаданные строки в world-service.
type Fetched struct {
	Snapshot      []byte
	SchemaVersion int32
	Version       int64
}

// GetWorld загружает мир по id. token может быть пустым, если на сервере не задан service_token.
func GetWorld(ctx context.Context, addr, token, worldID string) (*Fetched, error) {
	if addr == "" {
		return nil, fmt.Errorf("world service address is empty")
	}
	if worldID == "" {
		return nil, fmt.Errorf("world id is empty")
	}

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("grpc dial: %w", err)
	}
	defer func() { _ = conn.Close() }()

	if token != "" {
		ctx = metadata.AppendToOutgoingContext(ctx, "x-service-token", token)
	}

	cli := worldv1.NewWorldServiceClient(conn)
	resp, err := cli.GetWorld(ctx, &worldv1.GetWorldRequest{Id: worldID})
	if err != nil {
		return nil, fmt.Errorf("get world %q: %w", worldID, err)
	}
	w := resp.GetWorld()
	if w == nil {
		return nil, fmt.Errorf("world %q: empty response", worldID)
	}
	snap := w.GetSnapshot()
	if snap == nil {
		snap = []byte{}
	}
	return &Fetched{
		Snapshot:      snap,
		SchemaVersion: w.GetSchemaVersion(),
		Version:       w.GetVersion(),
	}, nil
}
