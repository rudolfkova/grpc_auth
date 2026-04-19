package gamews

import (
	"encoding/json"
	"log/slog"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	gameapp "game/internal/app/game"
	"game/internal/domain/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

type wsFakeEngine struct{}

func (wsFakeEngine) ProcessTick(_ []models.Action) []models.Event { return nil }
func (wsFakeEngine) SerializeWorld() ([]byte, error)              { return []byte(`{}`), nil }
func (wsFakeEngine) EnsurePlayerJoin(_ int64, _ gamekit.CharacterPlayData) {
}
func (wsFakeEngine) PlayerCharacterData(_ int64) ([]byte, error) { return []byte(`{}`), nil }
func (wsFakeEngine) JoinStateSnapshot() gamekit.StatePayload     { return gamekit.StatePayload{} }

// joinSnapFakeEngine — для проверки немедленного state с tiles после connect.
type joinSnapFakeEngine struct{ wsFakeEngine }

func (joinSnapFakeEngine) JoinStateSnapshot() gamekit.StatePayload {
	tiles := []gamekit.Tile{{X: 7, Y: 8, Layer: 0, Texture: "join_marker"}}
	return gamekit.StatePayload{
		Players: []gamekit.Player{},
		Tiles:   &tiles,
		TickAt:  time.Now().UTC(),
	}
}

func TestWSJoinSendsImmediateFullTileState(t *testing.T) {
	const secret = "secret"
	telemetry := gameapp.NewNoopTelemetry()
	app := gameapp.NewService(slog.Default(), joinSnapFakeEngine{}, time.Second, 8, nil, nil, telemetry, 0)
	h := NewHandler(slog.Default(), secret, app, telemetry)

	srv := httptest.NewServer(h)
	defer srv.Close()

	conn := mustDialWS(t, srv.URL, makeToken(t, secret, 202))
	defer conn.Close()

	var got gamekit.Envelope
	if err := conn.ReadJSON(&got); err != nil {
		t.Fatalf("read first message: %v", err)
	}
	if got.Service != gamekit.ServiceGame || got.Type != gamekit.TypeState {
		t.Fatalf("want game/state, got service=%q type=%q", got.Service, got.Type)
	}
	var st gamekit.StatePayload
	if err := json.Unmarshal(got.Payload, &st); err != nil {
		t.Fatalf("state unmarshal: %v", err)
	}
	if st.Tiles == nil || len(*st.Tiles) != 1 {
		t.Fatalf("want one tile in join state, got %+v", st)
	}
	if (*st.Tiles)[0].Texture != "join_marker" || (*st.Tiles)[0].X != 7 || (*st.Tiles)[0].Y != 8 {
		t.Fatalf("unexpected join tile %+v", (*st.Tiles)[0])
	}
}

func TestWSRejectsWrongService(t *testing.T) {
	const secret = "secret"
	telemetry := gameapp.NewNoopTelemetry()
	app := gameapp.NewService(slog.Default(), wsFakeEngine{}, time.Second, 8, nil, nil, telemetry, 0)
	h := NewHandler(slog.Default(), secret, app, telemetry)

	srv := httptest.NewServer(h)
	defer srv.Close()

	conn := mustDialWS(t, srv.URL, makeToken(t, secret, 101))
	defer conn.Close()
	drainInitialJoinState(t, conn)

	msg := gamekit.Envelope{
		Service: "other",
		Type:    gamekit.TypeMove,
		Payload: json.RawMessage(`{"dx":1,"dy":0}`),
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var got gamekit.Envelope
	if err := conn.ReadJSON(&got); err != nil {
		t.Fatalf("read json: %v", err)
	}
	if got.Type != gamekit.TypeReject {
		t.Fatalf("want reject, got %q", got.Type)
	}
	var payload gamekit.RejectPayload
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if payload.Reason != RejectReasonWrongService {
		t.Fatalf("want reason %q, got %q", RejectReasonWrongService, payload.Reason)
	}
}

func TestWSRejectsQueueFull(t *testing.T) {
	const secret = "secret"
	telemetry := gameapp.NewNoopTelemetry()
	app := gameapp.NewService(slog.Default(), wsFakeEngine{}, time.Second, 1, nil, nil, telemetry, 0)
	if !app.Submit(models.Action{PlayerID: 777, Type: gamekit.TypeMove, Payload: json.RawMessage(`{"dx":1,"dy":0}`)}) {
		t.Fatal("failed to prefill queue")
	}
	h := NewHandler(slog.Default(), secret, app, telemetry)

	srv := httptest.NewServer(h)
	defer srv.Close()

	conn := mustDialWS(t, srv.URL, makeToken(t, secret, 101))
	defer conn.Close()
	drainInitialJoinState(t, conn)

	msg := gamekit.Envelope{
		Service: gamekit.ServiceGame,
		Type:    gamekit.TypeMove,
		Payload: json.RawMessage(`{"dx":1,"dy":0}`),
	}
	if err := conn.WriteJSON(msg); err != nil {
		t.Fatalf("write json: %v", err)
	}

	var got gamekit.Envelope
	if err := conn.ReadJSON(&got); err != nil {
		t.Fatalf("read json: %v", err)
	}
	if got.Type != gamekit.TypeReject {
		t.Fatalf("want reject, got %q", got.Type)
	}
	var payload gamekit.RejectPayload
	if err := json.Unmarshal(got.Payload, &payload); err != nil {
		t.Fatalf("payload unmarshal: %v", err)
	}
	if payload.Reason != RejectReasonQueueFull {
		t.Fatalf("want reason %q, got %q", RejectReasonQueueFull, payload.Reason)
	}
}

func makeToken(t *testing.T, secret string, userID int64) string {
	t.Helper()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id": userID,
		"email":   "test@example.com",
		"exp":     time.Now().Add(time.Hour).Unix(),
	})
	out, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("signed token: %v", err)
	}
	return out
}

func mustDialWS(t *testing.T, httpURL, token string) *websocket.Conn {
	t.Helper()
	u, err := url.Parse(httpURL)
	if err != nil {
		t.Fatalf("parse url: %v", err)
	}
	u.Scheme = strings.Replace(u.Scheme, "http", "ws", 1)
	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()
	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		t.Fatalf("dial websocket: %v", err)
	}
	return conn
}

// drainInitialJoinState читает первый персональный state после connect (полный tiles / join sync).
func drainInitialJoinState(t *testing.T, conn *websocket.Conn) {
	t.Helper()
	var got gamekit.Envelope
	if err := conn.ReadJSON(&got); err != nil {
		t.Fatalf("read join state: %v", err)
	}
	if got.Service != gamekit.ServiceGame || got.Type != gamekit.TypeState {
		t.Fatalf("want first message game/state, got service=%q type=%q", got.Service, got.Type)
	}
}
