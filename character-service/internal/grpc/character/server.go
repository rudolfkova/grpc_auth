package charactergrpc

import (
	"context"
	"errors"
	"log/slog"

	"character/internal/model"
	"character/internal/service"
	characterv1 "character/proto/character/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Register регистрирует CharacterService на gRPC-сервере.
func Register(srv *grpc.Server, svc *service.Character, log *slog.Logger) {
	characterv1.RegisterCharacterServiceServer(srv, &serverAPI{svc: svc, log: log})
}

type serverAPI struct {
	characterv1.UnimplementedCharacterServiceServer
	svc *service.Character
	log *slog.Logger
}

func (s *serverAPI) ResolvePlayCharacter(ctx context.Context, req *characterv1.ResolvePlayCharacterRequest) (*characterv1.ResolvePlayCharacterResponse, error) {
	c, persisted, err := s.svc.ResolveForPlay(ctx, req.GetUserId(), req.GetId())
	if err != nil {
		s.log.WarnContext(ctx, "ResolvePlayCharacter failed", "err", err)
		return nil, mapErr(err)
	}
	s.log.InfoContext(ctx, "ResolvePlayCharacter",
		"user_id", req.GetUserId(),
		"id", req.GetId(),
		"persisted", persisted,
	)
	return &characterv1.ResolvePlayCharacterResponse{
		Persisted: persisted,
		Character: toProto(c),
	}, nil
}

func (s *serverAPI) GetCharacter(ctx context.Context, req *characterv1.GetCharacterRequest) (*characterv1.GetCharacterResponse, error) {
	c, err := s.svc.Get(ctx, req.GetUserId(), req.GetId())
	if err != nil {
		s.log.WarnContext(ctx, "GetCharacter failed", "id", req.GetId(), "err", err)
		return nil, mapErr(err)
	}
	return &characterv1.GetCharacterResponse{Character: toProto(c)}, nil
}

func (s *serverAPI) GetCharacterByDisplayName(ctx context.Context, req *characterv1.GetCharacterByDisplayNameRequest) (*characterv1.GetCharacterByDisplayNameResponse, error) {
	c, err := s.svc.GetByDisplayName(ctx, req.GetUserId(), req.GetDisplayName())
	if err != nil {
		s.log.WarnContext(ctx, "GetCharacterByDisplayName failed", "err", err)
		return nil, mapErr(err)
	}
	return &characterv1.GetCharacterByDisplayNameResponse{Character: toProto(c)}, nil
}

func (s *serverAPI) ListCharacters(ctx context.Context, req *characterv1.ListCharactersRequest) (*characterv1.ListCharactersResponse, error) {
	list, err := s.svc.List(ctx, req.GetUserId(), req.GetLimit(), req.GetOffset())
	if err != nil {
		s.log.WarnContext(ctx, "ListCharacters failed", "err", err)
		return nil, mapErr(err)
	}
	out := make([]*characterv1.Character, 0, len(list))
	for i := range list {
		out = append(out, toProto(list[i]))
	}
	return &characterv1.ListCharactersResponse{Characters: out}, nil
}

func (s *serverAPI) CreateCharacter(ctx context.Context, req *characterv1.CreateCharacterRequest) (*characterv1.CreateCharacterResponse, error) {
	c, err := s.svc.Create(ctx,
		req.GetUserId(),
		req.GetId(),
		req.GetDisplayName(),
		req.GetDescription(),
		req.GetData(),
		req.GetSchemaVersion(),
	)
	if err != nil {
		s.log.WarnContext(ctx, "CreateCharacter failed", "err", err)
		return nil, mapErr(err)
	}
	return &characterv1.CreateCharacterResponse{Character: toProto(c)}, nil
}

func (s *serverAPI) ReplaceCharacterData(ctx context.Context, req *characterv1.ReplaceCharacterDataRequest) (*characterv1.ReplaceCharacterDataResponse, error) {
	c, err := s.svc.ReplaceData(ctx,
		req.GetUserId(),
		req.GetId(),
		req.GetData(),
		req.GetSchemaVersion(),
		req.GetExpectedVersion(),
	)
	if err != nil {
		s.log.WarnContext(ctx, "ReplaceCharacterData failed", "id", req.GetId(), "err", err)
		return nil, mapErr(err)
	}
	return &characterv1.ReplaceCharacterDataResponse{Character: toProto(c)}, nil
}

func (s *serverAPI) DeleteCharacter(ctx context.Context, req *characterv1.DeleteCharacterRequest) (*characterv1.DeleteCharacterResponse, error) {
	if err := s.svc.Delete(ctx, req.GetUserId(), req.GetId()); err != nil {
		s.log.WarnContext(ctx, "DeleteCharacter failed", "id", req.GetId(), "err", err)
		return nil, mapErr(err)
	}
	return &characterv1.DeleteCharacterResponse{}, nil
}

func toProto(c model.Character) *characterv1.Character {
	return &characterv1.Character{
		Id:             c.ID,
		UserId:         c.OwnerUserID,
		DisplayName:    c.DisplayName,
		Description:    c.Description,
		Data:           c.Data,
		SchemaVersion:  c.SchemaVersion,
		Version:        c.Version,
		CreatedAtUnix:  c.CreatedAt.Unix(),
		UpdatedAtUnix:  c.UpdatedAt.Unix(),
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
