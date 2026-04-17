package gamews

import (
	"encoding/json"

	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func (h *Handler) enqueueOutbound(out chan gamekit.Envelope, env gamekit.Envelope) {
	select {
	case out <- env:
	default:
		h.telemetry.ObserveEventDropped("ws_out")
	}
}

func (h *Handler) writeReject(out chan gamekit.Envelope, reason, message, reqType, reqService string) {
	env, err := buildRejectEnvelope(reason, message, reqType, reqService)
	if err != nil {
		return
	}
	h.enqueueOutbound(out, env)
}

func (h *Handler) processIncomingEnvelope(out chan gamekit.Envelope, userID int64, data []byte) {
	var env gamekit.Envelope
	if err := json.Unmarshal(data, &env); err != nil {
		h.writeReject(out, RejectReasonInvalidJSON, "message is not valid JSON", "", "")
		return
	}

	if env.Service != gamekit.ServiceGame {
		h.writeReject(out, RejectReasonWrongService, "expected service \"game\"", env.Type, env.Service)
		return
	}
	if env.Type == "" {
		h.writeReject(out, RejectReasonMissingType, "field \"type\" is required", "", env.Service)
		return
	}

	ok := h.app.Submit(models.Action{
		PlayerID: userID,
		Type:     env.Type,
		Payload:  env.Payload,
	})
	if !ok {
		h.telemetry.ObserveActionRejected(RejectReasonQueueFull)
		h.writeReject(out, RejectReasonQueueFull, "action queue is full, try again later", env.Type, env.Service)
	}
}
