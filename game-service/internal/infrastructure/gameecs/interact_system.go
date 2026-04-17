package gameecs

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"
)

// InteractSystem обрабатывает type=interact по каталогу content.
// Резолв тайла по клетке: texture тайла (TileTexture.Name / state.texture) == item_def_id,
// у предмета в каталоге должен быть interact. См. MergeInteractBase в pkg/gamekit/content.
type InteractSystem struct {
	bundle *content.Bundle
	log    *slog.Logger
	engine *Engine
}

// NewInteractSystem bundle может быть nil (система ничего не делает).
func NewInteractSystem(bundle *content.Bundle, log *slog.Logger, engine *Engine) *InteractSystem {
	return &InteractSystem{bundle: bundle, log: log, engine: engine}
}

func (s *InteractSystem) Update(ctx *TickContext) {
	if s.bundle == nil {
		return
	}
	a := ctx.CurrentAction
	if a.Type != gamekit.TypeInteract {
		return
	}
	var in gamekit.InteractIntent
	if err := json.Unmarshal(a.Payload, &in); err != nil {
		return
	}
	id := strings.TrimSpace(in.ItemDefID)
	if id == "" {
		return
	}
	it, ok := s.bundle.Catalog.Items[id]
	if !ok || it.Interact == nil {
		return
	}
	script := strings.TrimSpace(it.Interact.Script)
	sc, ok := s.bundle.Scenarios[script]
	if !ok || sc == nil {
		return
	}

	var tileInst json.RawMessage
	if in.ClickX != nil && in.ClickY != nil {
		_, _, _, raw, ok := s.engine.resolveCatalogTileAtClick(in)
		if !ok {
			return
		}
		tileInst = raw
	}

	base := content.MergeInteractBase(it.Interact.Args, tileInst)
	if in.ClickX != nil {
		base["click_x"] = *in.ClickX
	}
	if in.ClickY != nil {
		base["click_y"] = *in.ClickY
	}
	if in.ClickLayer != nil {
		base["click_layer"] = *in.ClickLayer
	}
	rcx := &content.RunContext{
		PlayerID:  a.PlayerID,
		ItemDefID: id,
		HostData:  s.engine,
	}
	if s.log != nil {
		rcx.Log = func(msg string) {
			s.log.Debug("interact", "message", msg, "item_def_id", id, "user_id", a.PlayerID)
		}
	}
	if err := s.bundle.Runner.Run(context.Background(), rcx, sc, base); err != nil && s.log != nil {
		s.log.Warn("interact scenario failed", "item_def_id", id, "err", err)
	}
}
