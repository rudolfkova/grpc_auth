package gameapp

import (
	"context"
	"encoding/json"
	"time"

	"game/internal/domain/models"
	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// Service orchestrates tick loop and batching.
type Service struct {
	engine   ports.GameEngine
	tickRate time.Duration
	ingress  chan models.Action
	events   chan models.Outbound
}

func NewService(engine ports.GameEngine, tickRate time.Duration, queueSize int) *Service {
	if tickRate <= 0 {
		tickRate = 50 * time.Millisecond
	}
	if queueSize <= 0 {
		queueSize = 1024
	}
	return &Service{
		engine:   engine,
		tickRate: tickRate,
		ingress:  make(chan models.Action, queueSize),
		events:   make(chan models.Outbound, queueSize),
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
			evs := s.engine.ProcessTick(actions)
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
