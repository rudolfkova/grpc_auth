package gamews

import (
	"encoding/json"

	"game/internal/domain/models"
)

// RejectReason — машинно читаемая причина отказа (для клиентского логирования).
const (
	RejectReasonQueueFull     = "queue_full"
	RejectReasonWrongService  = "wrong_service"
	RejectReasonMissingType   = "missing_type"
	RejectReasonInvalidJSON   = "invalid_json"
)

type rejectPayload struct {
	Reason          string `json:"reason"`
	Message         string `json:"message"`
	RequestType     string `json:"request_type,omitempty"`
	RequestService  string `json:"request_service,omitempty"`
}

func buildRejectEnvelope(reason, message, reqType, reqService string) (models.Envelope, error) {
	body, err := json.Marshal(rejectPayload{
		Reason:         reason,
		Message:        message,
		RequestType:    reqType,
		RequestService: reqService,
	})
	if err != nil {
		return models.Envelope{}, err
	}
	return models.Envelope{
		Service: "game",
		Type:    "reject",
		Payload: body,
	}, nil
}
