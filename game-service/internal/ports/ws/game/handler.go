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
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

type claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

type Handler struct {
	logger    *slog.Logger
	jwtSecret string
	app       *gameapp.Service

	mu      sync.RWMutex
	clients map[*websocket.Conn]clientConn
	byUser  map[int64]map[*websocket.Conn]chan gamekit.Envelope
}

type clientConn struct {
	userID int64
	out    chan gamekit.Envelope
}

func NewHandler(logger *slog.Logger, jwtSecret string, app *gameapp.Service) *Handler {
	h := &Handler{
		logger:    logger,
		jwtSecret: jwtSecret,
		app:       app,
		clients:   make(map[*websocket.Conn]clientConn),
		byUser:    make(map[int64]map[*websocket.Conn]chan gamekit.Envelope),
	}
	go h.broadcastSnapshots()
	return h
}

func (h *Handler) broadcastSnapshots() {
	for out := range h.app.Events() {
		h.send(out)
	}
}

func (h *Handler) send(out models.Outbound) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	if out.RecipientUserID == 0 {
		// broadcast
		for _, c := range h.clients {
			select {
			case c.out <- out.Message:
			default:
			}
		}
		return
	}

	if conns, ok := h.byUser[out.RecipientUserID]; ok {
		for _, ch := range conns {
			select {
			case ch <- out.Message:
			default:
			}
		}
	}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.logger.Error("ws upgrade failed", slog.String("err", err.Error()))
		return
	}

	userID, email, err := h.claimsFromToken(r.URL.Query().Get("token"))
	if err != nil {
		_ = conn.WriteJSON(map[string]any{
			"service": "game",
			"type":    "error",
			"payload": map[string]any{"message": "invalid token"},
		})
		_ = conn.Close()
		return
	}

	out := make(chan gamekit.Envelope, 64)
	h.mu.Lock()
	h.clients[conn] = clientConn{userID: userID, out: out}
	if _, ok := h.byUser[userID]; !ok {
		h.byUser[userID] = make(map[*websocket.Conn]chan gamekit.Envelope)
	}
	h.byUser[userID][conn] = out
	h.mu.Unlock()

	h.logger.Info("player connected",
		slog.Int64("user_id", userID),
		slog.String("email", email),
	)

	defer func() {
		h.logger.Info("player disconnected",
			slog.Int64("user_id", userID),
			slog.String("email", email),
		)
		h.mu.Lock()
		if c, ok := h.clients[conn]; ok {
			delete(h.clients, conn)
			if m, ok := h.byUser[c.userID]; ok {
				delete(m, conn)
				if len(m) == 0 {
					delete(h.byUser, c.userID)
				}
			}
		}
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

	writeReject := func(reason, message, reqType, reqService string) {
		env, err := buildRejectEnvelope(reason, message, reqType, reqService)
		if err != nil {
			return
		}
		_ = conn.WriteJSON(env)
	}

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
			continue
		}

		var env gamekit.Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			writeReject(RejectReasonInvalidJSON, "message is not valid JSON", "", "")
			continue
		}

		if env.Service != gamekit.ServiceGame {
			writeReject(RejectReasonWrongService, "expected service \"game\"", env.Type, env.Service)
			continue
		}
		if env.Type == "" {
			writeReject(RejectReasonMissingType, "field \"type\" is required", "", env.Service)
			continue
		}

		ok := h.app.Submit(models.Action{
			PlayerID: userID,
			Type:     env.Type,
			Payload:  env.Payload,
		})
		if !ok {
			writeReject(RejectReasonQueueFull, "action queue is full, try again later", env.Type, env.Service)
		}
	}
}

func (h *Handler) claimsFromToken(raw string) (userID int64, email string, err error) {
	c := &claims{}
	token, err := jwt.ParseWithClaims(raw, c, func(t *jwt.Token) (any, error) {
		return []byte(h.jwtSecret), nil
	})
	if err != nil || !token.Valid || c.UserID == 0 {
		return 0, "", err
	}
	return c.UserID, c.Email, nil
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
