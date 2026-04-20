# 2026-04-19: Inventory slot flags and new catalog items (server)

## Added

- Optional `ItemDef` fields in `pkg/gamekit/content/types.go` (additive JSON):
  - `allow_hand` — may occupy `hand_main` / `hand_off`
  - `allow_armor` — may occupy `armor`
  - `allow_accessory` — may occupy `accessory_1` / `accessory_2`
  - Omitted fields deserialize as `false` (strict placement).
- `content.ItemFitsInventorySlot(cat, itemDefID, slot)` in `pkg/gamekit/content/inventory_slot.go`:
  - empty `itemDefID` → allowed (swap semantics)
  - `backpack_*` → item must exist in catalog and not be `is_storage`
  - equip slots → item must exist and have the matching allow-flag
  - unknown `itemDefID` or nil `cat` → false
- Unit tests: `pkg/gamekit/content/inventory_slot_test.go`
- New catalog ids in `game-service/data/content/catalog.json`: `axe`, `flint`, `log`, `campfire` (`campfire` has no `allow_hand`; others have `allow_hand: true` where noted in catalog).

## Changed

- `gamekit.TrySwapInventorySlots` now takes `canPlaceItem func(slot, itemDefID string) bool` instead of a backpack-only callback (`pkg/gamekit/inventory.go`).
- `applyInventoryMoveLocked` uses `ItemFitsInventorySlot` when `inventoryCatalog` is loaded; if catalog is nil (dev), placement stays permissive (`game-service/internal/infrastructure/gameecs/engine_inventory.go`).
- Existing pickables that should remain movable to hands were given `allow_hand: true` in `game-service/data/content/catalog.json` (`floor_gem`, `key`, `tent_folded`).
- `game-service/internal/infrastructure/gameecs/testdata/content/catalog.json`: same flag rules for tests; added `sword` and `axe` for hand-swap coverage.

## Tests

- `go test ./...` in `pkg/gamekit`
- `go test ./...` in `game-service` (includes `TestEngine_inventoryMoveRejectsGemToArmor`, `TestEngine_inventoryMoveAllowsAxeFromBackpackToHand`, and `TestEngine_inventoryMoveSwapHands` with loaded catalog).

## Contract

- `inventory_move` swap is rejected server-side when either moved item would violate `ItemFitsInventorySlot` for its destination slot; next `state` leaves inventory unchanged for that swap.
