package gameecs

import (
	stdctx "context"
	"encoding/json"
	"fmt"
	"strings"

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
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, in, host)
		return nil
	})
	r.RegisterOp("tent_unfold_2x2", func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("tent_unfold_2x2: invalid HostData")
		}

		x, err := asInt(args["click_x"])
		if err != nil {
			return fmt.Errorf("click_x: %w", err)
		}
		y, err := asInt(args["click_y"])
		if err != nil {
			return fmt.Errorf("click_y: %w", err)
		}
		layer := gamekit.DroppedItemTileLayer
		if v, ok := args["layer"]; ok {
			layer, err = asInt(v)
			if err != nil {
				return fmt.Errorf("layer: %w", err)
			}
		}
		if v, ok := args["click_layer"]; ok {
			layer, err = asInt(v)
			if err != nil {
				return fmt.Errorf("click_layer: %w", err)
			}
		}

		anchorID := asStringDefault(args["anchor_item_def_id"], "tent_anchor")
		topLeftTexture := asStringDefault(args["top_left_texture"], "tent_1")
		topRightTexture := asStringDefault(args["top_right_texture"], "tent_2")
		bottomLeftTexture := asStringDefault(args["bottom_left_texture"], asStringDefault(args["anchor_texture"], "tent_3"))
		bottomRightTexture := asStringDefault(args["bottom_right_texture"], "tent_4")

		// Anchor is lower-left corner. Do not unfold if target footprint is occupied on item layer.
		if host.tileExistsAtLayer(x+1, y, layer) || host.tileExistsAtLayer(x, y-1, layer) || host.tileExistsAtLayer(x+1, y-1, layer) {
			return nil
		}

		host.removeTilesAtLayerRecorded(x, y, layer)

		// lower-left anchor mapping to tent.png quadrants:
		// (x,y-1)->tent_1, (x+1,y-1)->tent_2, (x,y)->tent_3, (x+1,y)->tent_4
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:       x,
			Y:       y,
			Layer:   layer,
			Texture: bottomLeftTexture,
			Blocks:  false,
			InstanceArgs: gamekit.NormalizeTileInstanceArgsJSON(mustMarshalJSON(map[string]any{
				"item_def_id": anchorID,
			})),
		}, host)
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:       x + 1,
			Y:       y,
			Layer:   layer,
			Texture: bottomRightTexture,
			Blocks:  false,
		}, host)
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:       x,
			Y:       y - 1,
			Layer:   layer,
			Texture: topLeftTexture,
			Blocks:  false,
		}, host)
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:       x + 1,
			Y:       y - 1,
			Layer:   layer,
			Texture: topRightTexture,
			Blocks:  false,
		}, host)

		return nil
	})

	r.RegisterOp("tent_fold_2x2", func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("tent_fold_2x2: invalid HostData")
		}
		x, err := asInt(args["click_x"])
		if err != nil {
			return fmt.Errorf("click_x: %w", err)
		}
		y, err := asInt(args["click_y"])
		if err != nil {
			return fmt.Errorf("click_y: %w", err)
		}
		layer := gamekit.DroppedItemTileLayer
		if v, ok := args["layer"]; ok {
			layer, err = asInt(v)
			if err != nil {
				return fmt.Errorf("layer: %w", err)
			}
		}
		if v, ok := args["click_layer"]; ok {
			layer, err = asInt(v)
			if err != nil {
				return fmt.Errorf("click_layer: %w", err)
			}
		}
		foldedID := asStringDefault(args["folded_item_def_id"], "tent_folded")
		foldedTexture := asStringDefault(args["folded_texture"], "tent_folded")

		host.removeTilesAtLayerRecorded(x, y, layer)
		host.removeTilesAtLayerRecorded(x+1, y, layer)
		host.removeTilesAtLayerRecorded(x, y-1, layer)
		host.removeTilesAtLayerRecorded(x+1, y-1, layer)

		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:       x,
			Y:       y,
			Layer:   layer,
			Texture: foldedTexture,
			Blocks:  false,
			InstanceArgs: gamekit.NormalizeTileInstanceArgsJSON(mustMarshalJSON(map[string]any{
				"item_def_id": foldedID,
			})),
		}, host)
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

func asStringDefault(v any, fallback string) string {
	s, _ := v.(string)
	s = strings.TrimSpace(s)
	if s == "" {
		return fallback
	}
	return s
}

func mustMarshalJSON(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func (e *Engine) tileExistsAtLayer(x, y, layer int) bool {
	q := e.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, lay, _, _, _ := q.Get()
		if pos.X == x && pos.Y == y && lay.Z == layer {
			return true
		}
	}
	return false
}
