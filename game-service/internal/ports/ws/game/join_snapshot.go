package gamews

import (
	"encoding/json"
	"log/slog"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// enqueueJoinFullTileState отправляет персональный state с полным списком tiles в очередь исходящих сообщений соединения.
// Вызывается сразу после registerConnection, до цикла чтения; не трогает глобальный broadcast и lastFullTileSyncAt.
func (h *Handler) enqueueJoinFullTileState(out chan<- gamekit.Envelope, userID int64) {
	pl := h.app.JoinStateSnapshot()
	body, err := json.Marshal(pl)
	if err != nil {
		h.logger.Warn("join state marshal failed", slog.Int64("user_id", userID), slog.String("err", err.Error()))
		return
	}
	msg := gamekit.Envelope{
		Service: gamekit.ServiceGame,
		Type:    gamekit.TypeState,
		Payload: body,
	}
	select {
	case out <- msg:
	default:
		h.telemetry.ObserveEventDropped("ws_join_snapshot")
	}
}
