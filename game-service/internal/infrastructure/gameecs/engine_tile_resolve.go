package gameecs

import (
	"encoding/json"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// resolveCatalogTileAtClick ищет тайл в (click_x, click_y), у которого texture == item_def_id.
// Возвращает координаты клетки, слой и instance_args выбранного тайла (как для interact).
func (e *Engine) resolveCatalogTileAtClick(in gamekit.InteractIntent) (x, y, layer int, inst json.RawMessage, ok bool) {
	if in.ClickX == nil || in.ClickY == nil {
		return 0, 0, 0, nil, false
	}
	x, y = *in.ClickX, *in.ClickY
	itemID := strings.TrimSpace(in.ItemDefID)
	if itemID == "" {
		return 0, 0, 0, nil, false
	}
	type hit struct {
		z    int
		args json.RawMessage
	}
	var hits []hit
	q := e.tileFilter.Query()
	for q.Next() {
		pos, lay, _, tex, _ := q.Get()
		if pos.X != x || pos.Y != y {
			continue
		}
		if tex.Name != itemID {
			continue
		}
		var args json.RawMessage
		if len(tex.InstanceArgs) > 0 {
			args = append(json.RawMessage(nil), tex.InstanceArgs...)
		}
		hits = append(hits, hit{z: lay.Z, args: args})
	}
	q.Close()
	if len(hits) == 0 {
		return 0, 0, 0, nil, false
	}
	if in.ClickLayer != nil {
		want := *in.ClickLayer
		for _, h := range hits {
			if h.z == want {
				return x, y, h.z, h.args, true
			}
		}
		return 0, 0, 0, nil, false
	}
	best := hits[0]
	for _, h := range hits[1:] {
		if h.z > best.z {
			best = h
		}
	}
	return x, y, best.z, best.args, true
}
