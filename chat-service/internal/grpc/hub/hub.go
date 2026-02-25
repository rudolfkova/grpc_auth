// Package hub ...
package hub

import (
	chatv1 "chat/proto/chat/v1"
	"sync"
)

// Hub хранит активные Subscribe-стримы.
// Ключ - user_id, значение - список стримов (один пользователь
// может быть подключён с нескольких устройств).
type Hub struct {
	mu      sync.RWMutex
	streams map[int][]chatv1.ChatService_SubscribeServer
}

// New ...
func New() *Hub {
	return &Hub{
		streams: make(map[int][]chatv1.ChatService_SubscribeServer),
	}
}

// Subscribe регистрирует стрим для пользователя.
func (h *Hub) Subscribe(userID int, stream chatv1.ChatService_SubscribeServer) {
	h.mu.Lock()
	h.streams[userID] = append(h.streams[userID], stream)
	h.mu.Unlock()
}

// Unsubscribe удаляет стрим пользователя.
func (h *Hub) Unsubscribe(userID int, stream chatv1.ChatService_SubscribeServer) {
	h.mu.Lock()
	defer h.mu.Unlock()

	streams := h.streams[userID]
	for i, s := range streams {
		if s == stream {
			h.streams[userID] = append(streams[:i], streams[i+1:]...)
			break
		}
	}

	if len(h.streams[userID]) == 0 {
		delete(h.streams, userID)
	}
}

// Push отправляет сообщение всем стримам пользователя.
func (h *Hub) Push(userID int, msg *chatv1.MessageDTO) {
	h.mu.RLock()
	streams := h.streams[userID]
	h.mu.RUnlock()

	for _, stream := range streams {
		_ = stream.Send(msg)
	}
}
