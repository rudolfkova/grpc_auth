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
	r.RegisterOp("require_inventory_item", func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("require_inventory_item: invalid HostData")
		}
		keyID := strings.TrimSpace(asStringDefault(args["key_item_def_id"], ""))
		if keyID == "" {
			return fmt.Errorf("require_inventory_item: key_item_def_id required")
		}
		ent, ok := host.byUser[rcx.PlayerID]
		if !ok || !host.playerGearMapper.HasAll(ent) {
			return fmt.Errorf("require_inventory_item: player not found")
		}
		invPtr := host.playerGearMapper.Get(ent)
		if invPtr == nil || !invPtr.ContainsItemDefID(keyID) {
			return fmt.Errorf("require_inventory_item: missing item %q", keyID)
		}
		return nil
	})

	// door_apply_open_state — переключает «тело» двери на collision_layer: закрыто ↔ открыто (texture_closed+blocks / texture_open).
	// Имя op историческое; тайл триггера на click_layer не меняет.
	doorToggleCollisionBody := func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("door_toggle: invalid HostData")
		}
		x, err := asInt(args["click_x"])
		if err != nil {
			return fmt.Errorf("click_x: %w", err)
		}
		y, err := asInt(args["click_y"])
		if err != nil {
			return fmt.Errorf("click_y: %w", err)
		}
		clickLayer, err := asInt(args["click_layer"])
		if err != nil {
			return fmt.Errorf("click_layer: %w", err)
		}
		closedTex := strings.TrimSpace(asStringDefault(args["texture_closed"], ""))
		openTex := strings.TrimSpace(asStringDefault(args["texture_open"], ""))
		if closedTex == "" || openTex == "" {
			return fmt.Errorf("door_toggle: texture_closed and texture_open required")
		}
		if closedTex == openTex {
			return fmt.Errorf("door_toggle: texture_closed and texture_open must differ")
		}
		collisionLayer := clickLayer - 1
		if v, has := args["collision_layer"]; has {
			collisionLayer, err = asInt(v)
			if err != nil {
				return fmt.Errorf("collision_layer: %w", err)
			}
		}
		rot, curTex, _, blocks, found := host.readTileAtLayer(x, y, collisionLayer)
		cur := strings.TrimSpace(curTex)
		if !found {
			rot = 0
			spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
				X:        x,
				Y:        y,
				Layer:    collisionLayer,
				Rotation: rot,
				Texture:  closedTex,
				Blocks:   true,
			}, host)
			return nil
		}

		switch {
		case cur == openTex && !blocks:
			spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
				X:        x,
				Y:        y,
				Layer:    collisionLayer,
				Rotation: rot,
				Texture:  closedTex,
				Blocks:   true,
			}, host)
		case cur == closedTex && blocks:
			spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
				X:        x,
				Y:        y,
				Layer:    collisionLayer,
				Rotation: rot,
				Texture:  openTex,
				Blocks:   false,
			}, host)
		default:
			if !blocks {
				spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
					X:        x,
					Y:        y,
					Layer:    collisionLayer,
					Rotation: rot,
					Texture:  closedTex,
					Blocks:   true,
				}, host)
			} else {
				spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
					X:        x,
					Y:        y,
					Layer:    collisionLayer,
					Rotation: rot,
					Texture:  openTex,
					Blocks:   false,
				}, host)
			}
		}
		return nil
	}
	r.RegisterOp("door_apply_open_state", doorToggleCollisionBody)
	r.RegisterOp("door_toggle_collision_body", doorToggleCollisionBody)

	r.RegisterOp("replace_tile_texture_at_click", func(_ stdctx.Context, rcx *content.RunContext, args map[string]any) error {
		host, ok := rcx.HostData.(*Engine)
		if !ok || host == nil {
			return fmt.Errorf("replace_tile_texture_at_click: invalid HostData")
		}
		x, err := asInt(args["click_x"])
		if err != nil {
			return fmt.Errorf("click_x: %w", err)
		}
		y, err := asInt(args["click_y"])
		if err != nil {
			return fmt.Errorf("click_y: %w", err)
		}
		layer, err := asInt(args["click_layer"])
		if err != nil {
			return fmt.Errorf("click_layer: %w", err)
		}
		newTex := strings.TrimSpace(asStringDefault(args["texture"], ""))
		if newTex == "" {
			newTex = strings.TrimSpace(asStringDefault(args["texture_open"], ""))
		}
		if newTex == "" {
			return fmt.Errorf("replace_tile_texture_at_click: texture or texture_open required")
		}
		rot, curTex, inst, blocks, ok := host.readTileAtLayer(x, y, layer)
		if !ok {
			return fmt.Errorf("replace_tile_texture_at_click: no tile at (%d,%d) layer %d", x, y, layer)
		}
		if strings.TrimSpace(curTex) == newTex {
			return nil
		}
		spawnTileAt(host.world, host.tileMapper, host.tileFilter, gamekit.TileSpawnIntent{
			X:            x,
			Y:            y,
			Layer:        layer,
			Rotation:     rot,
			Texture:      newTex,
			Blocks:       blocks,
			InstanceArgs: inst,
		}, host)
		return nil
	})

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

func (e *Engine) readTileAtLayer(x, y, layer int) (rotation int, texture string, instanceArgs json.RawMessage, blocks bool, ok bool) {
	q := e.tileFilter.Query()
	defer q.Close()
	for q.Next() {
		pos, lay, face, tex, sol := q.Get()
		if pos.X != x || pos.Y != y || lay.Z != layer {
			continue
		}
		var raw json.RawMessage
		if len(tex.InstanceArgs) > 0 {
			raw = append(json.RawMessage(nil), tex.InstanceArgs...)
		}
		return face.RotationQuarter, strings.TrimSpace(tex.Name), raw, sol.Blocks, true
	}
	return 0, "", nil, false, false
}
