package gameplay

import "game/internal/domain/models"

// Emitter собирает доменные события за один тик.
// Не потокобезопасен: вызывается только из одного потока под блокировкой движка.
type Emitter struct {
	events []models.Event
}

// NewEmitter создаёт коллектор событий с начальной ёмкостью.
func NewEmitter(capHint int) *Emitter {
	if capHint <= 0 {
		capHint = 8
	}
	return &Emitter{events: make([]models.Event, 0, capHint)}
}

func (e *Emitter) Broadcast(typ string, payload any) {
	e.events = append(e.events, models.Event{
		RecipientUserID: 0,
		Type:            typ,
		Payload:         payload,
	})
}

func (e *Emitter) ToUser(userID int64, typ string, payload any) {
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

// Events возвращает накопленные события (срез тот же, что внутри Emitter).
func (e *Emitter) Events() []models.Event {
	return e.events
}
