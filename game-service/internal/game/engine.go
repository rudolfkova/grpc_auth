package game

import (
	"context"
	"log/slog"
	"time"
)

// Action is an intent from client.
type Action struct {
	PlayerID int64
	Type     string
	DX       int
	DY       int
}

// Position is the current player position.
type Position struct {
	X int `json:"x"`
	Y int `json:"y"`
}

// Snapshot is a full state snapshot for broadcast.
type Snapshot struct {
	Type    string              `json:"type"`
	Players map[int64]Position  `json:"players"`
	TickAt  time.Time           `json:"tick_at"`
}

// Engine contains game loop + in-memory state.
type Engine struct {
	logger   *slog.Logger
	tickRate time.Duration
	actions  chan Action
	events   chan Snapshot
	state    map[int64]Position
}

// NewEngine creates a game engine with in-memory state.
func NewEngine(logger *slog.Logger, tickRate time.Duration, queueSize int) *Engine {
	if queueSize <= 0 {
		queueSize = 1024
	}
	if tickRate <= 0 {
		tickRate = 50 * time.Millisecond
	}

	return &Engine{
		logger:   logger,
		tickRate: tickRate,
		actions:  make(chan Action, queueSize),
		events:   make(chan Snapshot, queueSize),
		state:    make(map[int64]Position),
	}
}

// EnqueueAction queues an action; false means queue is full.
func (e *Engine) EnqueueAction(a Action) bool {
	select {
	case e.actions <- a:
		return true
	default:
		return false
	}
}

// Events returns read-only snapshots stream.
func (e *Engine) Events() <-chan Snapshot {
	return e.events
}

// Run starts tick loop. It is blocking.
func (e *Engine) Run(ctx context.Context) {
	ticker := time.NewTicker(e.tickRate)
	defer ticker.Stop()

	e.logger.Info("game loop started", slog.Duration("tick_rate", e.tickRate))

	for {
		select {
		case <-ctx.Done():
			e.logger.Info("game loop stopped")
			return
		case <-ticker.C:
			e.update()
		}
	}
}

func (e *Engine) update() {
	// Drain action queue for current tick.
	for {
		select {
		case a := <-e.actions:
			if a.Type != "move" || a.PlayerID == 0 {
				continue
			}
			pos := e.state[a.PlayerID]
			pos.X += a.DX
			pos.Y += a.DY
			e.state[a.PlayerID] = pos
		default:
			goto snapshot
		}
	}

snapshot:
	s := Snapshot{
		Type:    "state",
		Players: make(map[int64]Position, len(e.state)),
		TickAt:  time.Now().UTC(),
	}
	for id, pos := range e.state {
		s.Players[id] = pos
	}

	select {
	case e.events <- s:
	default:
		// Drop if consumer is slow; loop must stay real-time.
	}
}
