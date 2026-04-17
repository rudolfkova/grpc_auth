package gamews

import (
	"context"
	"encoding/json"
	"time"

	"game/internal/domain/ports"
	"github.com/gorilla/websocket"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func (h *Handler) registerConnection(conn *websocket.Conn, userID int64) chan gamekit.Envelope {
	out := make(chan gamekit.Envelope, 64)
	h.mu.Lock()
	h.clients[conn] = clientConn{userID: userID, out: out}
	if _, ok := h.byUser[userID]; !ok {
		h.byUser[userID] = make(map[*websocket.Conn]chan gamekit.Envelope)
	}
	h.byUser[userID][conn] = out
	h.mu.Unlock()
	return out
}

func (h *Handler) unregisterConnection(conn *websocket.Conn, userID int64, out chan gamekit.Envelope, characterSessionActive bool, characterPersisted bool, characterPlay ports.CharacterIdentity) {
	h.logger.Info("player disconnected", "user_id", userID)
	h.telemetry.ObserveWSConnectionClosed()

	var lastForUser bool
	h.mu.Lock()
	delete(h.clients, conn)
	if m, ok := h.byUser[userID]; ok {
		delete(m, conn)
		if len(m) == 0 {
			delete(h.byUser, userID)
			lastForUser = true
		}
	}
	h.mu.Unlock()

	if characterSessionActive && lastForUser {
		saveCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err := h.app.PersistCharacterPlaySession(saveCtx, userID, characterPersisted, characterPlay)
		cancel()
		if err != nil {
			h.logger.Warn("PersistCharacterPlaySession failed", "user_id", userID, "err", err)
		}
	}

	close(out)
	_ = conn.Close()
}

func (h *Handler) sendTokenError(conn *websocket.Conn, message string) {
	payload, _ := json.Marshal(gamekit.ErrorPayload{Message: message})
	_ = conn.WriteJSON(gamekit.Envelope{
		Service: gamekit.ServiceGame,
		Type:    gamekit.TypeError,
		Payload: payload,
	})
}
