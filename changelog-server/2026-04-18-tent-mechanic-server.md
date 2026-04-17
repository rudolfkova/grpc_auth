# 2026-04-18: Tent Mechanic (Server)

## Added

- Server-side tent mechanic under existing WS contract (`interact`, `pickup_item`, `state/tile_updates`).
- Separate-ID resolver on server click matching:
  1. `tile.instance_args.item_def_id`
  2. fallback `tile.texture`
- New content items in `game-service/data/content/catalog.json`:
  - `tent_folded` (pickable + interact unfold script)
  - `tent_anchor` (interact fold script, not pickable)
- New scripts:
  - `game-service/data/content/scripts/tent_unfold_2x2.json`
  - `game-service/data/content/scripts/tent_fold_2x2.json`

## Behavior

### Folded -> Deployed (unfold)

- `interact` on `tent_folded` replaces folded tile with 2x2 tent:
  - anchor `(x,y)` texture `tent_1` + `instance_args.item_def_id = "tent_anchor"`
  - `(x+1,y)` -> `tent_2`
  - `(x,y+1)` -> `tent_3`
  - `(x+1,y+1)` -> `tent_4`
- All 4 deployed tent tiles are non-pickable (pickup is disabled in deployed state).
- Only anchor tile is interactive by catalog id (`tent_anchor`).

### Deployed -> Folded (fold)

- `interact` on `tent_anchor` removes all 4 tent tiles and spawns folded tile at anchor.
- Folded tile is spawned with `instance_args.item_def_id = "tent_folded"` (separate_ids preserved).
- Current content intentionally uses `texture: "tent_1"` for folded state (visual reuse). If UX needs stronger distinction, switch folded to a dedicated texture key.

## Required invariants

- Anchor tile must carry `instance_args.item_def_id = "tent_anchor"`.
- Server click resolver matches by:
  1. `tile.instance_args.item_def_id`
  2. fallback `tile.texture`.
- Deployed 2x2 tent tiles are non-pickable; pickable state exists only for folded tent (`tent_folded`).

## Payload examples

### Interact unfolded/folded

```json
{
  "service": "game",
  "type": "interact",
  "payload": {
    "item_def_id": "tent_folded",
    "click_x": 10,
    "click_y": 10,
    "click_layer": 0
  }
}
```

```json
{
  "service": "game",
  "type": "interact",
  "payload": {
    "item_def_id": "tent_anchor",
    "click_x": 10,
    "click_y": 10,
    "click_layer": 0
  }
}
```

### State payload example (fold, full valid payload)

```json
{
  "service": "game",
  "type": "state",
  "payload": {
    "players": [
      {
        "id": 1,
        "x": 9,
        "y": 10,
        "hp": 10,
        "face_dx": 1,
        "face_dy": 0,
        "stats": { "strength": 10, "dexterity": 10, "constitution": 10, "intelligence": 10, "wisdom": 10, "charisma": 10 },
        "sprite": "adventurer",
        "inventory": {
          "armor": "",
          "accessory_1": "",
          "accessory_2": "",
          "hand_main": "",
          "hand_off": "",
          "backpack": ["", "", "", "", ""]
        }
      }
    ],
    "tile_updates": [
      { "op": "remove", "x": 10, "y": 10, "layer": 0 },
      { "op": "remove", "x": 11, "y": 10, "layer": 0 },
      { "op": "remove", "x": 10, "y": 11, "layer": 0 },
      { "op": "remove", "x": 11, "y": 11, "layer": 0 },
      {
        "op": "upsert",
        "tile": {
          "x": 10,
          "y": 10,
          "layer": 0,
          "rotation": 0,
          "texture": "tent_1",
          "blocks": false,
          "instance_args": { "item_def_id": "tent_folded" }
        }
      }
    ],
    "tick_at": "2026-04-18T21:35:10Z"
  }
}
```
