package handler

import (
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gorilla/websocket"
)

// GameWSHandler is a pure WS proxy: client <-> game-service.
type GameWSHandler struct {
	logger          *slog.Logger
	gameServiceAddr string
}

// NewGameWSHandler ...
func NewGameWSHandler(logger *slog.Logger, gameServiceAddr string) *GameWSHandler {
	return &GameWSHandler{logger: logger, gameServiceAddr: gameServiceAddr}
}

func (h *GameWSHandler) SubscribeGame(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("game ws upgrade failed", slog.String("err", err.Error()))
		return
	}
	defer clientConn.Close()

	q := url.Values{}
	q.Set("token", r.URL.Query().Get("token"))
	q.Set("session_id", r.URL.Query().Get("session_id"))

	backendURL := "ws://" + h.gameServiceAddr + "/ws/game?" + q.Encode()

	gameConn, _, err := websocket.DefaultDialer.Dial(backendURL, nil)
	if err != nil {
		h.logger.Error("game ws dial failed", slog.String("err", err.Error()))
		_ = clientConn.WriteJSON(map[string]any{
			"service": "game",
			"type":    "error",
			"payload": map[string]any{"message": "game service unavailable"},
		})
		return
	}
	defer gameConn.Close()

	done := make(chan struct{}, 2)
	proxy := func(src, dst *websocket.Conn) {
		defer func() { done <- struct{}{} }()
		for {
			mt, msg, err := src.ReadMessage()
			if err != nil {
				return
			}
			if err := dst.WriteMessage(mt, msg); err != nil {
				return
			}
		}
	}

	go proxy(clientConn, gameConn)
	go proxy(gameConn, clientConn)
	<-done
}
