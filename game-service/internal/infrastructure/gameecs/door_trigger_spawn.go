package gameecs

import (
	"encoding/json"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"

	"github.com/mlange-42/ark/ecs"
)

// doorTriggerPairFromInstanceArgs: при spawn триггера с instance_args.texture_closed и texture_open
// дополнительно ставит «тело» на collision_layer (по умолчанию triggerLayer-1) с коллизией.
func doorTriggerPairFromInstanceArgs(raw json.RawMessage, triggerLayer int) (collisionLayer int, texClosed string, ok bool) {
	if len(raw) == 0 {
		return 0, "", false
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil || m == nil {
		return 0, "", false
	}
	tc := strings.TrimSpace(asStringAny(m["texture_closed"]))
	to := strings.TrimSpace(asStringAny(m["texture_open"]))
	if tc == "" || to == "" {
		return 0, "", false
	}
	cl := triggerLayer - 1
	if v, has := m["collision_layer"]; has {
		n, err := asInt(v)
		if err == nil {
			cl = n
		}
	}
	return cl, tc, true
}

func asStringAny(v any) string {
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

// spawnDoorTriggerClosedBodyIfNeeded вызывается после spawn триггера; не трогает сам триггер.
func spawnDoorTriggerClosedBodyIfNeeded(
	w *ecs.World,
	tiles *ecs.Map5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	filter *ecs.Filter5[gamekit.GridPos, gamekit.TileLayer, gamekit.TileFacing, gamekit.TileTexture, gamekit.TileSolid],
	in gamekit.TileSpawnIntent,
	rec *Engine,
) {
	cl, texClosed, ok := doorTriggerPairFromInstanceArgs(in.InstanceArgs, in.Layer)
	if !ok {
		return
	}
	spawnTileAt(w, tiles, filter, gamekit.TileSpawnIntent{
		X:         in.X,
		Y:         in.Y,
		Layer:     cl,
		Rotation:  in.Rotation,
		Texture:   texClosed,
		Blocks:    true,
		InstanceArgs: nil,
	}, rec)
}
