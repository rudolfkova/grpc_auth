// Package handler ...
package handler

import (
	chatv1 "gateway/proto/chat/v1"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc/metadata"
)

var upgrader = websocket.Upgrader{
	// В продакшене здесь нужно проверять Origin
	CheckOrigin: func(r *http.Request) bool { return true },
}

// WSHandler держит gRPC-клиент chat-сервиса.
type WSHandler struct {
	client chatv1.ChatServiceClient
	logger *slog.Logger
}

// NewWSHandler ...
func NewWSHandler(client chatv1.ChatServiceClient, logger *slog.Logger) *WSHandler {
	return &WSHandler{client: client, logger: logger}
}

// Subscribe GET /ws/subscribe
// Апгрейдит HTTP соединение до WebSocket, открывает gRPC стрим
// к chat-service и пушит входящие сообщения клиенту.
func (h *WSHandler) Subscribe(w http.ResponseWriter, r *http.Request) {
	// 1. Апгрейд до WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", slog.String("err", err.Error()))
		return
	}
	defer conn.Close()

	// 2. Открываем gRPC стрим к chat-service, прокидываем JWT
	token := r.URL.Query().Get("token")
	md := metadata.Pairs("authorization", "Bearer "+token)
	ctx := metadata.NewOutgoingContext(r.Context(), md)
	stream, err := h.client.Subscribe(ctx, &chatv1.SubscribeRequest{})
	if err != nil {
		h.logger.Error("grpc subscribe failed", slog.String("err", err.Error()))
		return
	}

	// 3. Читаем сообщения из gRPC стрима и пушим в WebSocket
	for {
		msg, err := stream.Recv()
		if err != nil {
			// Клиент отключился или chat-service упал — выходим
			h.logger.Debug("grpc stream closed", slog.String("err", err.Error()))
			return
		}

		payload := map[string]any{
			"id":         msg.GetId(),
			"chat_id":    msg.GetChatId(),
			"sender_id":  msg.GetSenderId(),
			"text":       msg.GetText(),
			"created_at": msg.GetCreatedAt().AsTime(),
		}

		if err := conn.WriteJSON(payload); err != nil {
			h.logger.Debug("ws write failed", slog.String("err", err.Error()))
			return
		}
	}
}
