package gamews

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"
	"time"

	gameapp "game/internal/app/game"
	"game/internal/domain/models"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type claims struct {
	UserID int64 `json:"user_id"`
	jwt.RegisteredClaims
}

type envelope struct {
	Service string          `json:"service"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type statePayload struct {
	Players []models.Player `json:"players"`
}

type Handler struct {
	logger    *slog.Logger
	jwtSecret string
	app       *gameapp.Service

	mu      sync.RWMutex
	clients map[*websocket.Conn]chan envelope
}

func NewHandler(logger *slog.Logger, jwtSecret string, app *gameapp.Service) *Handler {
	h := &Handler{
		logger:    logger,
		jwtSecret: jwtSecret,
		app:       app,
		clients:   make(map[*websocket.Conn]chan envelope),
	}
	go h.broadcastSnapshots()
	return h
}

func (h *Handler) broadcastSnapshots() {
	for snap := range h.app.Events() {
		players := make([]models.Player, 0, len(snap.Players))
		for id, pos := range snap.Players {
			players = append(players, models.Player{ID: id, X: pos.X, Y: pos.Y})
		}
		body, err := json.Marshal(statePayload{Players: players})
		if err != nil {
			continue
		}
		h.broadcast(envelope{Service: "game", Type: "state", Payload: body})
	}
}

func (h *Handler) broadcast(msg envelope) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for _, out := range h.clients {
		select {
		case out <- msg:
		default:
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", slog.String("err", err.Error()))
		return
	}

	userID, err := h.userIDFromToken(r.URL.Query().Get("token"))
	if err != nil {
		_ = conn.WriteJSON(map[string]any{
			"service": "game",
			"type":    "error",
			"payload": map[string]any{"message": "invalid token"},
		})
		_ = conn.Close()
		return
	}

	out := make(chan envelope, 64)
	h.mu.Lock()
	h.clients[conn] = out
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, conn)
		h.mu.Unlock()
		close(out)
		_ = conn.Close()
	}()

	go func() {
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}()

	for {
		var env envelope
		if err := conn.ReadJSON(&env); err != nil {
			return
		}
		if env.Service != "game" || env.Type != "move" {
			continue
		}
		var mv models.MoveIntent
		if err := json.Unmarshal(env.Payload, &mv); err != nil {
			continue
		}
		if mv.DX < -1 || mv.DX > 1 || mv.DY < -1 || mv.DY > 1 {
			continue
		}
		_ = h.app.Submit(models.Action{
			PlayerID: userID,
			Type:     "move",
			DX:       mv.DX,
			DY:       mv.DY,
		})
	}
}

func (h *Handler) userIDFromToken(raw string) (int64, error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid || c.UserID == 0 {
		return 0, err
	}
	return c.UserID, nil
}

func Serve(ctx context.Context, logger *slog.Logger, addr string, handler http.Handler) error {
	srv := &http.Server{
		Addr:              addr,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		_ = srv.Shutdown(context.Background())
	}()

	logger.Info("game ws server started", slog.String("addr", addr))
	err := srv.ListenAndServe()
	if err == http.ErrServerClosed {
		return nil
	}
	return err
}
