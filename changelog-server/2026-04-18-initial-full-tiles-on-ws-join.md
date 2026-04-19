# Server: immediate full `tiles` on WebSocket join

## Problem

Game `state` normally includes a full `tiles` array only on a **global** interval (`tile_full_sync_interval`, default ~1s in engine options). Between those ticks the broadcast payload carries only `tile_updates` (or empty tile fields).

A client connecting **between** full-sync ticks could receive many `state` messages with **no** `tiles` and no pending `tile_updates` for unchanged cells, so the world stayed visually empty until the next periodic full sync (worst case ~interval, often reported as ~5s when the interval is configured larger).

This was a **server** gap: the client applies `tiles` as soon as they arrive (`worldstate.ApplyEnvelope`: `p.Tiles != nil` → full replace).

## Change

After a successful WS upgrade and `registerConnection`, the handler enqueues **one** personal `state` message on that connection’s outbound channel:

- Same wire shape as tick `state`: `service: "game"`, `type: "state"`, `payload` = `gamekit.StatePayload` with **`players`**, **`tick_at`**, and full **`tiles`** (deep-copied tile list, same as periodic full sync).
- Built via `Engine.JoinStateSnapshot()` — does **not** update `lastFullTileSyncAt`, so global periodic behavior is unchanged.
- Non-blocking send: if the connection’s `out` buffer is full, the snapshot is dropped (metric channel label `ws_join_snapshot`).

## Client notes

No protocol change. First message after connect may be this join `state` before tick-driven broadcasts. Clients should already handle `tiles != nil` as authoritative full replace (see `internal/core/worldstate/world.go` in the game client).

## Example (first message after connect)

```json
{
  "service": "game",
  "type": "state",
  "payload": {
    "players": [
      {
        "id": 42,
        "x": 0,
        "y": 0,
        "hp": 100,
        "face_dx": 0,
        "face_dy": 1,
        "stats": {},
        "sprite": "human",
        "inventory": {}
      }
    ],
    "tick_at": "2026-04-18T12:00:00.123Z",
    "tiles": [
      {
        "x": 0,
        "y": 0,
        "layer": 0,
        "rotation": 0,
        "texture": "floor",
        "blocks": false
      }
    ]
  }
}
```

(Real `players` / `tiles` length depends on the live world; `tile_updates` is omitted when `tiles` is present.)
