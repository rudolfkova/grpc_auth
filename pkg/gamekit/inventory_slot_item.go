package gamekit

import (
	"fmt"
	"strings"
)

// InventorySlotHasItem reports whether inv holds wantItemDefID in the named inventory slot.
// Slot names are the same as for ItemAtSlot / inventory_move (e.g. hand_main, hand_off, backpack_0).
// wantItemDefID must be non-empty after TrimSpace.
// Returns (false, err) if the slot name is invalid or wantItemDefID is empty.
// Returns (false, nil) if inv is nil or the slot contents (after trim) do not equal wantItemDefID.
func InventorySlotHasItem(inv *PlayerInventory, slot, wantItemDefID string) (bool, error) {
	want := strings.TrimSpace(wantItemDefID)
	if want == "" {
		return false, fmt.Errorf("gamekit: item_def_id required")
	}
	slot = strings.TrimSpace(slot)
	if inv == nil {
		return false, nil
	}
	got, ok := inv.ItemAtSlot(slot)
	if !ok {
		return false, fmt.Errorf("gamekit: invalid inventory slot %q", slot)
	}
	return strings.TrimSpace(got) == want, nil
}
