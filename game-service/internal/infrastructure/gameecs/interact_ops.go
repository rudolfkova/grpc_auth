package gameecs

import (
	stdctx "context"
	"encoding/json"
	"fmt"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
	"github.com/rudolfkova/grpc_auth/pkg/gamekit/content"
)

func registerGameContentOps(r *content.Runner, e *Engine) {
	r.RegisterOp("world_spawn_tile", func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("world_spawn_tile: invalid HostData")
		}
		in, err := tileIntentFromInteractArgs(args)
		if err != nil {
			return err
		}
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, in)
		return nil
	})
}

func tileIntentFromInteractArgs(args map[string]any) (gamekit.TileSpawnIntent, error) {
	var in gamekit.TileSpawnIntent
	var err error
	if in.X, err = asInt(args["x"]); err != nil {
		return in, fmt.Errorf("x: %w", err)
	}
	if in.Y, err = asInt(args["y"]); err != nil {
		return in, fmt.Errorf("y: %w", err)
	}
	if v, ok := args["layer"]; ok {
		if in.Layer, err = asInt(v); err != nil {
			return in, fmt.Errorf("layer: %w", err)
		}
	}
	if v, ok := args["rotation"]; ok {
		if in.Rotation, err = asInt(v); err != nil {
			return in, fmt.Errorf("rotation: %w", err)
		}
	}
	tex, _ := args["texture"].(string)
	if tex == "" {
		return in, fmt.Errorf("texture is required")
	}
	in.Texture = tex
	if v, ok := args["blocks"]; ok {
		switch b := v.(type) {
		case bool:
			in.Blocks = b
		case float64:
			in.Blocks = b != 0
		}
	}
	if v, ok := args["instance_args"]; ok && v != nil {
		raw, err := json.Marshal(v)
		if err == nil {
			in.InstanceArgs = gamekit.NormalizeTileInstanceArgsJSON(raw)
		}
	}
	return in, nil
}

func asInt(v any) (int, error) {
	if v == nil {
		return 0, fmt.Errorf("nil")
	}
	switch t := v.(type) {
	case float64:
		return int(t), nil
	case int:
		return t, nil
	case int64:
		return int(t), nil
	default:
		return 0, fmt.Errorf("expected number, got %T", v)
	}
}
