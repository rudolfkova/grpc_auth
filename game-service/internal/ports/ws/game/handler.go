package gamews

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	gameapp "game/internal/app/game"
	"game/internal/domain/ports"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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
	telemetry gameapp.Telemetry

	mu      sync.RWMutex
	clients map[*websocket.Conn]clientConn
	byUser  map[int64]map[*websocket.Conn]chan gamekit.Envelope
}

type clientConn struct {
	userID int64
	out    chan gamekit.Envelope
}

func NewHandler(logger *slog.Logger, jwtSecret string, app *gameapp.Service, telemetry gameapp.Telemetry) *Handler {
	if telemetry == nil {
		telemetry = gameapp.NewNoopTelemetry()
	}
	h := &Handler{
		logger:    logger,
		jwtSecret: jwtSecret,
		app:       app,
		telemetry: telemetry,
		clients:   make(map[*websocket.Conn]clientConn),
		byUser:    make(map[int64]map[*websocket.Conn]chan gamekit.Envelope),
	}
	go h.broadcastSnapshots()
	return h
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		h.telemetry.ObserveWSUpgradeFailure()
		h.logger.Error("ws upgrade failed", slog.String("err", err.Error()))
		return
	}

	userID, email, err := h.claimsFromToken(r.URL.Query().Get("token"))
	if err != nil {
		h.sendTokenError(conn, "invalid token")
		_ = conn.Close()
		return
	}

	h.logger.Info("player connected",
		slog.Int64("user_id", userID),
		slog.String("email", email),
	)
	h.telemetry.ObserveWSConnectionOpened()

	var (
		characterPersisted     bool
		characterPlay          ports.CharacterIdentity
		characterSessionActive bool
	)
	if h.app.CharacterSessionsEnabled() {
		rawID := strings.TrimSpace(r.URL.Query().Get("character_id"))
		if rawID == "" {
			h.sendTokenError(conn, "character_id query parameter is required when character service is configured")
			_ = conn.Close()
			return
		}
		if _, err := uuid.Parse(rawID); err != nil {
			h.sendTokenError(conn, "character_id must be a valid UUID")
			_ = conn.Close()
			return
		}
		resolveCtx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		resp, err := h.app.ResolveCharacterJoin(resolveCtx, userID, rawID)
		cancel()
		if err != nil {
			h.logger.Warn("ResolvePlayCharacter failed", slog.Int64("user_id", userID), slog.String("err", err.Error()))
			h.sendTokenError(conn, "character resolve failed: "+err.Error())
			_ = conn.Close()
			return
		}
		if resp == nil || strings.TrimSpace(resp.Character.ID) == "" {
			h.sendTokenError(conn, "character resolve returned empty character")
			_ = conn.Close()
			return
		}
		characterPersisted = resp.Persisted
		characterPlay = resp.Character
		characterSessionActive = true
		h.app.PrepareCharacterJoin(userID, resp.Data)
	}

	out := h.registerConnection(conn, userID)

	defer func() {
		h.unregisterConnection(conn, userID, out, characterSessionActive, characterPersisted, characterPlay)
	}()

	go func() {
		for msg := range out {
			if err := conn.WriteJSON(msg); err != nil {
				return
			}
		}
	}()

	h.enqueueJoinFullTileState(out, userID)

	for {
		mt, data, err := conn.ReadMessage()
		if err != nil {
			return
		}
		if mt != websocket.TextMessage && mt != websocket.BinaryMessage {
			continue
		}
		h.processIncomingEnvelope(out, userID, data)
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
