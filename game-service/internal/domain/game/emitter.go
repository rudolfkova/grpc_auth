package game

import "game/internal/domain/models"

// emitter collects domain events during one tick.
// It is intentionally not concurrency-safe: tick processing is single-threaded under Engine lock.
type emitter struct {
	events []models.Event
}

func newEmitter(capHint int) *emitter {
	if capHint <= 0 {
		capHint = 8
	}
	return &emitter{events: make([]models.Event, 0, capHint)}
}

func (e *emitter) Broadcast(typ string, payload any) {
	e.events = append(e.events, models.Event{
		RecipientUserID: 0,
		Type:            typ,
		Payload:         payload,
	})
}

func (e *emitter) ToUser(userID int64, typ string, payload any) {
	if userID == 0 {
		e.Broadcast(typ, payload)
		return
	}
	e.events = append(e.events, models.Event{
		RecipientUserID: userID,
		Type:            typ,
		Payload:         payload,
	})
}

func (e *emitter) Events() []models.Event {
	return e.events
}

