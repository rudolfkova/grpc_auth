package gamews

import (
	"game/internal/domain/models"
)

func (h *Handler) broadcastSnapshots() {
	for out := range h.app.Events() {
		h.send(out)
	}
}

func (h *Handler) send(out models.Outbound) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if out.RecipientUserID == 0 {
		for _, c := range h.clients {
			select {
			case c.out <- out.Message:
			default:
				h.telemetry.ObserveEventDropped("ws_out")
			}
		}
		return
	}

	if conns, ok := h.byUser[out.RecipientUserID]; ok {
		for _, ch := range conns {
			select {
			case ch <- out.Message:
			default:
				h.telemetry.ObserveEventDropped("ws_out")
			}
		}
	}
}
