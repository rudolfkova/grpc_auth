package gameapp

import (
	"context"
	"time"

	"game/internal/domain/models"
)

// Engine is the domain boundary used by app layer.
type Engine interface {
	ProcessTick(actions []models.Action) models.Snapshot
}

// Service orchestrates tick loop and batching.
type Service struct {
	engine   Engine
	tickRate time.Duration
	ingress  chan models.Action
	events   chan models.Snapshot
}

func NewService(engine Engine, tickRate time.Duration, queueSize int) *Service {
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
		events:   make(chan models.Snapshot, queueSize),
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

func (s *Service) Events() <-chan models.Snapshot {
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
			snap := s.engine.ProcessTick(actions)
			select {
			case s.events <- snap:
			default:
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
