package handler

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"

	"github.com/gorilla/websocket"
)

// GameWSHandler is a pure WS proxy: client <-> game-service.
type GameWSHandler struct {
	logger          *slog.Logger
	gameServiceAddr string
}

func NewGameWSHandler(logger *slog.Logger, gameServiceAddr string) *GameWSHandler {
	return &GameWSHandler{
		logger:          logger,
		gameServiceAddr: gameServiceAddr,
	}
}

func (h *GameWSHandler) SubscribeGame(w http.ResponseWriter, r *http.Request) {
	clientConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("game ws upgrade failed", slog.String("err", err.Error()))
		return
	}
	defer clientConn.Close()

	token := r.URL.Query().Get("token")
	backendURL := "ws://" + strings.TrimPrefix(h.gameServiceAddr, "http://")
	backendURL = strings.TrimPrefix(backendURL, "https://")
	backendURL = strings.TrimPrefix(backendURL, "ws://")
	backendURL = "ws://" + backendURL + "/ws/game?token=" + url.QueryEscape(token)

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
			mt, msg, readErr := src.ReadMessage()
			if readErr != nil {
				return
			}
			if writeErr := dst.WriteMessage(mt, msg); writeErr != nil {
				return
			}
		}
	}

	go proxy(clientConn, gameConn)
	go proxy(gameConn, clientConn)
	<-done
}
