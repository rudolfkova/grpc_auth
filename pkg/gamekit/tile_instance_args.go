package gamekit

import (
	"encoding/json"
)

// MaxTileInstanceArgsJSONBytes — верхняя граница размера JSON instance_args (spawn_tile / тайл в ECS / world_spawn_tile).
const MaxTileInstanceArgsJSONBytes = 64 << 10

// NormalizeTileInstanceArgsJSON оставляет только непустой JSON-объект не больше MaxTileInstanceArgsJSONBytes; иначе nil.
func NormalizeTileInstanceArgsJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 {
		return nil
	}
	if len(raw) > MaxTileInstanceArgsJSONBytes {
		return nil
	}
	var probe any
	if err := json.Unmarshal(raw, &probe); err != nil {
		return nil
	}
	m, ok := probe.(map[string]any)
	if !ok {
		return nil
	}
	if len(m) == 0 {
		return nil
	}
	return raw
}
