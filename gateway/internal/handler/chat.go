// Package handler ...
package handler

import (
	chatv1 "gateway/proto/chat/v1"
	"net/http"

	"google.golang.org/grpc/metadata"
)

// ChatHandler держит gRPC-клиент chat-сервиса.
type ChatHandler struct {
	client chatv1.ChatServiceClient
}

// NewChatHandler ...
func NewChatHandler(client chatv1.ChatServiceClient) *ChatHandler {
	return &ChatHandler{client: client}
}

// forwardAuth прокидывает Authorization заголовок из HTTP в gRPC метаданные.
func forwardAuth(r *http.Request) metadata.MD {
	token := r.Header.Get("Authorization")
	if token == "" {
		return metadata.MD{}
	}
	return metadata.Pairs("authorization", token)
}

// CreateChat POST /chat/create
// Body: { "name": "session chat" }
// name опционален — если пустой, создаётся DM-заготовка без участников.
func (h *ChatHandler) CreateChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name string `json:"name"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	resp, err := h.client.CreateChat(ctx, &chatv1.CreateChatRequest{Name: req.Name})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"chat_id":    resp.GetChatId(),
		"created_at": resp.GetCreatedAt().AsTime(),
	})
}

// DeleteChat DELETE /chat
// Body: { "chat_id": 1 }
func (h *ChatHandler) DeleteChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChatID int64 `json:"chat_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChatID == 0 {
		writeError(w, http.StatusBadRequest, "chat_id is required")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	if _, err := h.client.DeleteChat(ctx, &chatv1.DeleteChatRequest{ChatId: req.ChatID}); err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// AddMember POST /chat/members
// Body: { "chat_id": 1, "user_id": 2 }
func (h *ChatHandler) AddMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChatID int64 `json:"chat_id"`
		UserID int64 `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChatID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "chat_id and user_id are required")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	if _, err := h.client.AddMember(ctx, &chatv1.AddMemberRequest{
		ChatId: req.ChatID,
		UserId: req.UserID,
	}); err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// RemoveMember DELETE /chat/members
// Body: { "chat_id": 1, "user_id": 2 }
func (h *ChatHandler) RemoveMember(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChatID int64 `json:"chat_id"`
		UserID int64 `json:"user_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChatID == 0 || req.UserID == 0 {
		writeError(w, http.StatusBadRequest, "chat_id and user_id are required")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	if _, err := h.client.RemoveMember(ctx, &chatv1.RemoveMemberRequest{
		ChatId: req.ChatID,
		UserId: req.UserID,
	}); err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

// GetOrCreateChat POST /chat/get-or-create
// Body: { "initiator_id": 1, "recipient_id": 2 }
func (h *ChatHandler) GetOrCreateChat(w http.ResponseWriter, r *http.Request) {
	var req struct {
		InitiatorID int64 `json:"initiator_id"`
		RecipientID int64 `json:"recipient_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.InitiatorID == 0 || req.RecipientID == 0 {
		writeError(w, http.StatusBadRequest, "initiator_id and recipient_id are required")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	resp, err := h.client.GetOrCreateChat(ctx, &chatv1.GetOrCreateChatRequest{
		InitiatorId: req.InitiatorID,
		RecipientId: req.RecipientID,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"chat_id":    resp.GetChatId(),
		"created":    resp.GetCreated(),
		"created_at": resp.GetCreatedAt().AsTime(),
	})
}

// GetMessages GET /chat/messages?chat_id=1&limit=50&cursor=...
func (h *ChatHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	chatID, err := queryInt64(r, "chat_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "chat_id query param required")
		return
	}
	limit, _ := queryInt64(r, "limit")
	if limit == 0 {
		limit = 50
	}
	cursor := r.URL.Query().Get("cursor")

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	resp, err := h.client.GetMessages(ctx, &chatv1.GetMessagesRequest{
		ChatId: chatID,
		Limit:  limit,
		Cursor: cursor,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	type msgDTO struct {
		ID        int64  `json:"id"`
		ChatID    int64  `json:"chat_id"`
		SenderID  int64  `json:"sender_id"`
		Text      string `json:"text"`
		CreatedAt any    `json:"created_at"`
	}
	msgs := make([]msgDTO, len(resp.GetMessages()))
	for i, m := range resp.GetMessages() {
		msgs[i] = msgDTO{
			ID:        m.GetId(),
			ChatID:    m.GetChatId(),
			SenderID:  m.GetSenderId(),
			Text:      m.GetText(),
			CreatedAt: m.GetCreatedAt().AsTime(),
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"messages":    msgs,
		"next_cursor": resp.GetNextCursor(),
	})
}

// GetUserChats GET /chat/chats?user_id=1&limit=50&offset=0
func (h *ChatHandler) GetUserChats(w http.ResponseWriter, r *http.Request) {
	userID, err := queryInt64(r, "user_id")
	if err != nil {
		writeError(w, http.StatusBadRequest, "user_id query param required")
		return
	}
	limit, _ := queryInt64(r, "limit")
	if limit == 0 {
		limit = 50
	}
	offset, _ := queryInt64(r, "offset")

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	resp, err := h.client.GetUserChats(ctx, &chatv1.GetUserChatsRequest{
		UserId: userID,
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	type chatDTO struct {
		ChatID        int64  `json:"chat_id"`
		Name          string `json:"name"`
		CompanionID   int64  `json:"companion_id"`
		LastMessage   string `json:"last_message"`
		UnreadCount   int64  `json:"unread_count"`
		LastMessageAt any    `json:"last_message_at"`
	}
	chats := make([]chatDTO, len(resp.GetChats()))
	for i, c := range resp.GetChats() {
		chats[i] = chatDTO{
			ChatID:        c.GetChatId(),
			Name:          c.GetName(),
			CompanionID:   c.GetCompanionId(),
			LastMessage:   c.GetLastMessage(),
			UnreadCount:   c.GetUnreadCount(),
			LastMessageAt: c.GetLastMessageAt().AsTime(),
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"chats": chats})
}

// SendMessage POST /chat/send
// Body: { "chat_id": 1, "sender_id": 2, "text": "hello" }
func (h *ChatHandler) SendMessage(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChatID   int64  `json:"chat_id"`
		SenderID int64  `json:"sender_id"`
		Text     string `json:"text"`
	}
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ChatID == 0 || req.SenderID == 0 || req.Text == "" {
		writeError(w, http.StatusBadRequest, "chat_id, sender_id and text are required")
		return
	}

	ctx := metadata.NewOutgoingContext(r.Context(), forwardAuth(r))
	resp, err := h.client.SendMessage(ctx, &chatv1.SendMessageRequest{
		ChatId:   req.ChatID,
		SenderId: req.SenderID,
		Text:     req.Text,
	})
	if err != nil {
		writeError(w, grpcStatusToHTTP(err), grpcMessage(err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"message_id": resp.GetMessageId(),
		"created_at": resp.GetCreatedAt().AsTime(),
	})
}
