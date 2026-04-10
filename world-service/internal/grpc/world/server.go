package worldgrpc

import (
	"context"
	"errors"
	"log/slog"

	"world/internal/model"
	"world/internal/service"
	worldv1 "world/proto/world/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register регистрирует WorldService на gRPC-сервере.
func Register(srv *grpc.Server, svc *service.World, log *slog.Logger) {
	worldv1.RegisterWorldServiceServer(srv, &serverAPI{svc: svc, log: log})
}

type serverAPI struct {
	worldv1.UnimplementedWorldServiceServer
	svc *service.World
	log *slog.Logger
}

func (s *serverAPI) CreateWorld(ctx context.Context, req *worldv1.CreateWorldRequest) (*worldv1.CreateWorldResponse, error) {
	w := model.World{
		ID:            req.GetId(),
		Name:          req.GetName(),
		Description:   req.GetDescription(),
		Snapshot:      req.GetSnapshot(),
		SchemaVersion: req.GetSchemaVersion(),
	}
	if err := s.svc.Create(ctx, w); err != nil {
		s.log.WarnContext(ctx, "CreateWorld failed",
			"world_id", w.ID,
			"schema_version", w.SchemaVersion,
			"snapshot_bytes", len(w.Snapshot),
			"err", err,
		)
		return nil, mapErr(err)
	}
	created, err := s.svc.Get(ctx, w.ID)
	if err != nil {
		s.log.WarnContext(ctx, "CreateWorld reload failed",
			"world_id", w.ID,
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "world created",
		"world_id", created.ID,
		"name", created.Name,
		"version", created.Version,
		"schema_version", created.SchemaVersion,
		"snapshot_bytes", len(created.Snapshot),
	)
	return &worldv1.CreateWorldResponse{World: toProto(created)}, nil
}

func (s *serverAPI) GetWorld(ctx context.Context, req *worldv1.GetWorldRequest) (*worldv1.GetWorldResponse, error) {
	id := req.GetId()
	w, err := s.svc.Get(ctx, id)
	if err != nil {
		s.log.WarnContext(ctx, "GetWorld failed",
			"world_id", id,
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "world retrieved",
		"world_id", w.ID,
		"name", w.Name,
		"version", w.Version,
		"schema_version", w.SchemaVersion,
		"snapshot_bytes", len(w.Snapshot),
	)
	return &worldv1.GetWorldResponse{World: toProto(w)}, nil
}

func (s *serverAPI) GetWorldByName(ctx context.Context, req *worldv1.GetWorldByNameRequest) (*worldv1.GetWorldByNameResponse, error) {
	name := req.GetName()
	w, err := s.svc.GetByName(ctx, name)
	if err != nil {
		s.log.WarnContext(ctx, "GetWorldByName failed",
			"name", name,
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "world retrieved by name",
		"world_id", w.ID,
		"name", w.Name,
		"version", w.Version,
		"schema_version", w.SchemaVersion,
		"snapshot_bytes", len(w.Snapshot),
	)
	return &worldv1.GetWorldByNameResponse{World: toProto(w)}, nil
}

func (s *serverAPI) ReplaceWorldSnapshot(ctx context.Context, req *worldv1.ReplaceWorldSnapshotRequest) (*worldv1.ReplaceWorldSnapshotResponse, error) {
	id := req.GetId()
	w, err := s.svc.ReplaceSnapshot(ctx, id, req.GetSnapshot(), req.GetSchemaVersion(), req.GetExpectedVersion())
	if err != nil {
		s.log.WarnContext(ctx, "ReplaceWorldSnapshot failed",
			"world_id", id,
			"expected_version", req.GetExpectedVersion(),
			"schema_version", req.GetSchemaVersion(),
			"snapshot_bytes", len(req.GetSnapshot()),
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "world snapshot replaced",
		"world_id", w.ID,
		"version", w.Version,
		"schema_version", w.SchemaVersion,
		"snapshot_bytes", len(w.Snapshot),
	)
	return &worldv1.ReplaceWorldSnapshotResponse{World: toProto(w)}, nil
}

func (s *serverAPI) DeleteWorld(ctx context.Context, req *worldv1.DeleteWorldRequest) (*worldv1.DeleteWorldResponse, error) {
	id := req.GetId()
	if err := s.svc.Delete(ctx, id); err != nil {
		s.log.WarnContext(ctx, "DeleteWorld failed",
			"world_id", id,
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "world deleted", "world_id", id)
	return &worldv1.DeleteWorldResponse{}, nil
}

func (s *serverAPI) ListWorlds(ctx context.Context, req *worldv1.ListWorldsRequest) (*worldv1.ListWorldsResponse, error) {
	limit, offset := req.GetLimit(), req.GetOffset()
	list, err := s.svc.List(ctx, limit, offset)
	if err != nil {
		s.log.WarnContext(ctx, "ListWorlds failed",
			"limit", limit,
			"offset", offset,
			"err", err,
		)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "worlds listed",
		"limit", limit,
		"offset", offset,
		"count", len(list),
	)
	out := make([]*worldv1.World, 0, len(list))
	for i := range list {
		out = append(out, toProto(list[i]))
	}
	return &worldv1.ListWorldsResponse{Worlds: out}, nil
}

func toProto(w model.World) *worldv1.World {
	return &worldv1.World{
		Id:            w.ID,
		Name:          w.Name,
		Description:   w.Description,
		Snapshot:      w.Snapshot,
		SchemaVersion: w.SchemaVersion,
		Version:       w.Version,
		CreatedAtUnix: w.CreatedAt.Unix(),
		UpdatedAtUnix: w.UpdatedAt.Unix(),
	}
}

func mapErr(err error) error {
	switch {
	case errors.Is(err, service.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, service.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, service.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, service.ErrVersionMismatch):
		return status.Error(codes.Aborted, err.Error())
	default:
		return status.Error(codes.Internal, "internal error")
	}
}
