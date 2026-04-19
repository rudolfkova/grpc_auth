package gameapp

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"game/internal/domain/models"
	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"google.golang.org/grpc/status"
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

	engine    ports.GameEngine
	telemetry Telemetry
	tickRate  time.Duration
	ingress   chan models.Action
	events    chan models.Outbound

	worldStore        ports.WorldStore
	characterSessions ports.CharacterSessions
	saveWorldAdminID  int64
}

func NewService(
	logger *slog.Logger,
	engine ports.GameEngine,
	tickRate time.Duration,
	queueSize int,
	worldStore ports.WorldStore,
	characterSessions ports.CharacterSessions,
	telemetry Telemetry,
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
	if telemetry == nil {
		telemetry = noopTelemetry{}
	}
	return &Service{
		logger:            logger,
		engine:            engine,
		telemetry:         telemetry,
		tickRate:          tickRate,
		ingress:           make(chan models.Action, queueSize),
		events:            make(chan models.Outbound, queueSize),
		worldStore:        worldStore,
		characterSessions: characterSessions,
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

// JoinStateSnapshot делегирует движку: полный state с tiles для первого сообщения после подключения.
func (s *Service) JoinStateSnapshot() gamekit.StatePayload {
	return s.engine.JoinStateSnapshot()
}

func (s *Service) Run(ctx context.Context) {
	ticker := time.NewTicker(s.tickRate)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tickStarted := time.Now()
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
					s.telemetry.ObserveEventDropped("events")
				}
			}
			s.telemetry.ObserveTick(time.Since(tickStarted), len(forTick))
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
			s.telemetry.ObserveEventDropped("events")
		}
	}
	defer send()

	if s.saveWorldAdminID == 0 {
		payload.Code = SaveWorldCodeDisabled
		payload.Message = "save_world is disabled (save_world_admin_user_id is 0)"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}
	if a.PlayerID != s.saveWorldAdminID {
		payload.Code = SaveWorldCodeForbidden
		payload.Message = "only configured admin user_id may save the world"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}

	var in gamekit.SaveWorldIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		payload.Code = SaveWorldCodeInvalidName
		payload.Message = "invalid save_world payload"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}
	name := strings.TrimSpace(in.Name)
	if name == "" {
		payload.Code = SaveWorldCodeInvalidName
		payload.Message = "name is required"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}

	if s.worldStore == nil {
		payload.Code = SaveWorldCodeUnconfigured
		payload.Message = "world store is not configured"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}

	snap, err := s.engine.SerializeWorld()
	if err != nil {
		s.logger.Warn("save_world serialize failed", "err", err)
		payload.Code = SaveWorldCodeSerializeFailed
		payload.Message = "failed to serialize world"
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}

	callCtx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()
	callStarted := time.Now()
	res, err := s.worldStore.SaveWorldByName(callCtx, ports.SaveWorldRequest{
		Name:          name,
		Description:   strings.TrimSpace(in.Description),
		Snapshot:      snap,
		SchemaVersion: gamekit.SnapshotSchemaVersion,
	})
	if err != nil {
		s.telemetry.ObserveGRPCClientCall("world", "save_world_by_name", status.Code(err).String(), time.Since(callStarted))
	} else {
		s.telemetry.ObserveGRPCClientCall("world", "save_world_by_name", "OK", time.Since(callStarted))
	}
	if err != nil {
		s.logger.Warn("save_world upstream failed", "name", name, "err", err)
		payload.Code = SaveWorldCodeUpstreamFailed
		payload.Message = err.Error()
		s.telemetry.ObserveSaveWorldResult(false, payload.Code)
		return
	}

	payload.Ok = true
	payload.WorldID = res.WorldID
	payload.Name = name
	payload.Version = res.Version
	s.telemetry.ObserveSaveWorldResult(true, "ok")
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
