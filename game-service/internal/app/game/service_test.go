package gameapp

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"game/internal/domain/models"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

type fakeEngine struct {
	gotTicks [][]models.Action
	events   []models.Event
}

func (f *fakeEngine) ProcessTick(actions []models.Action) []models.Event {
	cp := append([]models.Action(nil), actions...)
	f.gotTicks = append(f.gotTicks, cp)
	return append([]models.Event(nil), f.events...)
}

func (f *fakeEngine) SerializeWorld() ([]byte, error) {
	return []byte(`{}`), nil
}

func (f *fakeEngine) EnsurePlayerJoin(_ int64, _ gamekit.CharacterPlayData) {}

func (f *fakeEngine) PlayerCharacterData(_ int64) ([]byte, error) {
	return []byte(`{}`), nil
}

func (f *fakeEngine) JoinStateSnapshot() gamekit.StatePayload {
	return gamekit.StatePayload{}
}

func TestCollectActionsDrainsIngressInOrder(t *testing.T) {
	svc := NewService(nil, &fakeEngine{}, time.Second, 16, nil, nil, nil, 0)
	for i := 0; i < 3; i++ {
		if !svc.Submit(models.Action{PlayerID: 1, Type: "x", Payload: []byte{byte(i)}}) {
			t.Fatalf("submit %d failed", i)
		}
	}

	actions := svc.collectActions()
	if len(actions) != 3 {
		t.Fatalf("want 3 actions, got %d", len(actions))
	}
	for i := 0; i < 3; i++ {
		if got := actions[i].Payload[0]; got != byte(i) {
			t.Fatalf("order changed at %d: got %d want %d", i, got, i)
		}
	}
}

func TestRunFiltersSaveWorldBeforeEngine(t *testing.T) {
	engine := &fakeEngine{}
	svc := NewService(nil, engine, 5*time.Millisecond, 16, nil, nil, nil, 0)

	movePayload, _ := json.Marshal(gamekit.MoveIntent{DX: 1, DY: 0})
	savePayload, _ := json.Marshal(gamekit.SaveWorldIntent{Name: "x"})
	if !svc.Submit(models.Action{PlayerID: 1, Type: gamekit.TypeSaveWorld, Payload: savePayload}) {
		t.Fatal("submit save_world failed")
	}
	if !svc.Submit(models.Action{PlayerID: 1, Type: gamekit.TypeMove, Payload: movePayload}) {
		t.Fatal("submit move failed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() {
		svc.Run(ctx)
		close(done)
	}()

	deadline := time.After(300 * time.Millisecond)
	for len(engine.gotTicks) == 0 {
		select {
		case <-deadline:
			cancel()
			<-done
			t.Fatal("engine did not receive tick")
		default:
			time.Sleep(2 * time.Millisecond)
		}
	}
	cancel()
	<-done

	got := engine.gotTicks[0]
	if len(got) != 1 {
		t.Fatalf("want 1 action in ProcessTick, got %d", len(got))
	}
	if got[0].Type != gamekit.TypeMove {
		t.Fatalf("want only move action in ProcessTick, got %q", got[0].Type)
	}
}

func TestRunMarshalsEventsToOutboundEnvelope(t *testing.T) {
	engine := &fakeEngine{
		events: []models.Event{
			{
				Type: "state",
				Payload: map[string]any{
					"ok": true,
				},
			},
		},
	}
	svc := NewService(nil, engine, 5*time.Millisecond, 16, nil, nil, nil, 0)
	if !svc.Submit(models.Action{PlayerID: 1, Type: gamekit.TypeMove, Payload: json.RawMessage(`{"dx":1,"dy":0}`)}) {
		t.Fatal("submit failed")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	done := make(chan struct{})
	go func() {
		svc.Run(ctx)
		close(done)
	}()

	select {
	case out := <-svc.Events():
		if out.Message.Service != gamekit.ServiceGame {
			t.Fatalf("service mismatch: %q", out.Message.Service)
		}
		if out.Message.Type != "state" {
			t.Fatalf("event type mismatch: %q", out.Message.Type)
		}
		var payload map[string]any
		if err := json.Unmarshal(out.Message.Payload, &payload); err != nil {
			t.Fatalf("payload unmarshal: %v", err)
		}
		if payload["ok"] != true {
			t.Fatalf("payload mismatch: %#v", payload)
		}
	case <-time.After(300 * time.Millisecond):
		t.Fatal("did not receive outbound event")
	}
}
