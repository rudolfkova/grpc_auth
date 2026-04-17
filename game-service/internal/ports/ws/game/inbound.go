package gamews

import (
	"encoding/json"

	"game/internal/domain/models"
	"github.com/gorilla/websocket"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func (h *Handler) writeReject(conn *websocket.Conn, reason, message, reqType, reqService string) {
	env, err := buildRejectEnvelope(reason, message, reqType, reqService)
	if err != nil {
		return
	}
	_ = conn.WriteJSON(env)
}

func (h *Handler) processIncomingEnvelope(conn *websocket.Conn, userID int64, data []byte) {
	var env gamekit.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		h.writeReject(conn, RejectReasonInvalidJSON, "message is not valid JSON", "", "")
		return
	}

	if env.Service != gamekit.ServiceGame {
		h.writeReject(conn, RejectReasonWrongService, "expected service \"game\"", env.Type, env.Service)
		return
	}
	if env.Type == "" {
		h.writeReject(conn, RejectReasonMissingType, "field \"type\" is required", "", env.Service)
		return
	}

	ok := h.app.Submit(models.Action{
		PlayerID: userID,
		Type:     env.Type,
		Payload:  env.Payload,
	})
	if !ok {
		h.telemetry.ObserveActionRejected(RejectReasonQueueFull)
		h.writeReject(conn, RejectReasonQueueFull, "action queue is full, try again later", env.Type, env.Service)
	}
}
