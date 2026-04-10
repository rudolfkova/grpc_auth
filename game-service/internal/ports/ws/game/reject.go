package gamews

import (
	"encoding/json"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
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

func buildRejectEnvelope(reason, message, reqType, reqService string) (gamekit.Envelope, error) {
	body, err := json.Marshal(rejectPayload{
		Reason:         reason,
		Message:        message,
		RequestType:    reqType,
		RequestService: reqService,
	})
	if err != nil {
		return gamekit.Envelope{}, err
	}
	return gamekit.Envelope{
		Service: gamekit.ServiceGame,
		Type:    gamekit.TypeReject,
		Payload: body,
	}, nil
}
