# 2026-04-18: Tent Mechanic Hotfix (Server)

## Fixed

- Fixed folded tent duplication after unfold:
  - unfold/fold now operates on item layer (`layer=3`, `gamekit.DroppedItemTileLayer`);
  - folded source tile is correctly removed from layer 3 before deployed 2x2 spawn.
- Folded visual texture on fold is now explicit:
  - server spawns folded as `texture: "tent_folded"` (single asset icon),
  - while preserving `instance_args.item_def_id = "tent_folded"`.
- Tent footprint is now anchored at **lower-left** corner:
  - deployed tiles: `(x,y)`, `(x+1,y)`, `(x,y-1)`, `(x+1,y-1)`;
  - fold returns folded tent to lower-left anchor `(x,y)`.
- Added spawn guard for unfold:
  - if any target footprint tile on layer 3 is already occupied (except anchor source),
    unfold action is ignored (no destructive replacement).

## Behavior Notes

- Deployed tent remains non-pickable (anchor + visuals).
- Folded state remains pickable (`tent_folded`).
- Separate-ID contract unchanged:
  - server resolver checks `instance_args.item_def_id` first, then `texture`.

## Payload impact

- No new WS message types.
- Uses existing:
  - incoming: `interact`
  - outgoing: `state` with `tile_updates` upsert/remove
