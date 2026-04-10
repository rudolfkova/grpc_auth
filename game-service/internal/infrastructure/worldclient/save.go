package worldclient

import (
	"context"
	"fmt"
	"strings"

	worldv1 "world/proto/world/v1"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// SaveWorldByNameResult после успешного сохранения в world-service.
type SaveWorldByNameResult struct {
	WorldID string
	Version int64
}

// SaveWorldByName пишет снимок: при отсутствии мира с таким name — CreateWorld, иначе ReplaceWorldSnapshot (expected_version=0).
func SaveWorldByName(ctx context.Context, addr, token, name, description string, snapshot []byte, schemaVersion int32) (*SaveWorldByNameResult, error) {
	name = strings.TrimSpace(name)
	if addr == "" {
		return nil, fmt.Errorf("world service address is empty")
	}
	if name == "" {
		return nil, fmt.Errorf("world name is empty")
	}
	if snapshot == nil {
		snapshot = []byte{}
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

	getResp, err := cli.GetWorldByName(ctx, &worldv1.GetWorldByNameRequest{Name: name})
	if err == nil {
		w := getResp.GetWorld()
		if w == nil || w.GetId() == "" {
			return nil, fmt.Errorf("get world by name %q: empty world", name)
		}
		rep, err := cli.ReplaceWorldSnapshot(ctx, &worldv1.ReplaceWorldSnapshotRequest{
			Id:              w.GetId(),
			Snapshot:        snapshot,
			SchemaVersion:   schemaVersion,
			ExpectedVersion: 0,
		})
		if err != nil {
			return nil, fmt.Errorf("replace snapshot %q: %w", w.GetId(), err)
		}
		out := rep.GetWorld()
		if out == nil {
			return nil, fmt.Errorf("replace snapshot %q: empty response", w.GetId())
		}
		return &SaveWorldByNameResult{WorldID: out.GetId(), Version: out.GetVersion()}, nil
	}

	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		id := uuid.NewString()
		cr, err := cli.CreateWorld(ctx, &worldv1.CreateWorldRequest{
			Id:            id,
			Name:          name,
			Description:   description,
			Snapshot:      snapshot,
			SchemaVersion: schemaVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("create world %q: %w", name, err)
		}
		w := cr.GetWorld()
		if w == nil {
			return nil, fmt.Errorf("create world %q: empty response", name)
		}
		return &SaveWorldByNameResult{WorldID: w.GetId(), Version: w.GetVersion()}, nil
	}

	return nil, fmt.Errorf("get world by name %q: %w", name, err)
}
