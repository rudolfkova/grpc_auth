package gameapp

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"game/internal/domain/models"
	"game/internal/domain/ports"
	"game/internal/infrastructure/worldclient"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// Коды payload.save_world_result.code (стабильные для клиента).
const (
	SaveWorldCodeForbidden       = "save_world_forbidden"
	SaveWorldCodeDisabled        = "save_world_disabled"
	SaveWorldCodeInvalidName     = "save_world_invalid_name"
	SaveWorldCodeUnconfigured    = "save_world_unconfigured"
	SaveWorldCodeSerializeFailed = "save_world_serialize_failed"
	SaveWorldCodeUpstreamFailed  = "save_world_upstream_failed"
)

// Service orchestrates tick loop and batching.
type Service struct {
	logger *slog.Logger

	engine   ports.GameEngine
	tickRate time.Duration
	ingress  chan models.Action
	events   chan models.Outbound

	worldServiceAddr  string
	worldServiceToken string
	saveWorldAdminID  int64
}

func NewService(
	logger *slog.Logger,
	engine ports.GameEngine,
	tickRate time.Duration,
	queueSize int,
	worldServiceAddr, worldServiceToken string,
	saveWorldAdminUserID int64,
) *Service {
	if logger == nil {
		logger = slog.Default()
	}
	if tickRate <= 0 {
		tickRate = 50 * time.Millisecond
	}
	if queueSize <= 0 {
		queueSize = 1024
	}
	return &Service{
		logger:            logger,
		engine:            engine,
		tickRate:          tickRate,
		ingress:           make(chan models.Action, queueSize),
		events:            make(chan models.Outbound, queueSize),
		worldServiceAddr:  strings.TrimSpace(worldServiceAddr),
		worldServiceToken: worldServiceToken,
		saveWorldAdminID:  saveWorldAdminUserID,
	}
}

func (s *Service) Submit(a models.Action) bool {
	select {
	case s.ingress <- a:
		return true
	default:
		return false
	}
}

func (s *Service) Events() <-chan models.Outbound {
	return s.events
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			actions := s.collectActions()
			var forTick []models.Action
			for _, a := range actions {
				if a.Type == gamekit.TypeSaveWorld {
					s.handleSaveWorld(ctx, a)
					continue
				}
				forTick = append(forTick, a)
			}
			evs := s.engine.ProcessTick(forTick)
			for _, ev := range evs {
				body, err := json.Marshal(ev.Payload)
				if err != nil {
					continue
				}
				out := models.Outbound{
					RecipientUserID: ev.RecipientUserID,
					Message: gamekit.Envelope{
						Service: gamekit.ServiceGame,
						Type:    ev.Type,
						Payload: body,
					},
				}
				select {
				case s.events <- out:
				default:
				}
			}
		}
	}
}

func (s *Service) handleSaveWorld(ctx context.Context, a models.Action) {
	payload := gamekit.SaveWorldResultPayload{Ok: false}

	send := func() {
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}
		out := models.Outbound{
			RecipientUserID: a.PlayerID,
			Message: gamekit.Envelope{
				Service: gamekit.ServiceGame,
				Type:    gamekit.TypeSaveWorldResult,
				Payload: body,
			},
		}
		select {
		case s.events <- out:
		default:
		}
	}
	defer send()

	if s.saveWorldAdminID == 0 {
		payload.Code = SaveWorldCodeDisabled
		payload.Message = "save_world is disabled (save_world_admin_user_id is 0)"
		return
	}
	if a.PlayerID != s.saveWorldAdminID {
		payload.Code = SaveWorldCodeForbidden
		payload.Message = "only configured admin user_id may save the world"
		return
	}

	var in gamekit.SaveWorldIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		payload.Code = SaveWorldCodeInvalidName
		payload.Message = "invalid save_world payload"
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		payload.Code = SaveWorldCodeInvalidName
		payload.Message = "name is required"
		return
	}

	if s.worldServiceAddr == "" {
		payload.Code = SaveWorldCodeUnconfigured
		payload.Message = "world_service_addr is empty"
		return
	}

	snap, err := s.engine.SerializeWorld()
	if err != nil {
		s.logger.Warn("save_world serialize failed", "err", err)
		payload.Code = SaveWorldCodeSerializeFailed
		payload.Message = "failed to serialize world"
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	res, err := worldclient.SaveWorldByName(callCtx, s.worldServiceAddr, s.worldServiceToken,
		name, strings.TrimSpace(in.Description), snap, gamekit.SnapshotSchemaVersion)
	if err != nil {
		s.logger.Warn("save_world upstream failed", "name", name, "err", err)
		payload.Code = SaveWorldCodeUpstreamFailed
		payload.Message = err.Error()
		return
	}

	payload.Ok = true
	payload.WorldID = res.WorldID
	payload.Name = name
	payload.Version = res.Version
}

func (s *Service) collectActions() []models.Action {
	actions := make([]models.Action, 0, 256)
	for {
		select {
		case a := <-s.ingress:
			actions = append(actions, a)
		default:
			return actions
		}
	}
}
