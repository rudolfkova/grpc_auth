package content

import (
	"fmt"
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// TryPutItemInFirstBackpack places itemDefID into the first empty backpack slot.
// If cat is non-nil, ItemFitsInventorySlot must allow the item in that backpack slot.
// If cat is nil, only fullness is checked (matches permissive inventory when no catalog is loaded).
func TryPutItemInFirstBackpack(inv *gamekit.PlayerInventory, cat *Catalog, itemDefID string) error {
	id := strings.TrimSpace(itemDefID)
	if id == "" {
		return fmt.Errorf("empty item_def_id")
	}
	if inv == nil {
		return fmt.Errorf("nil inventory")
	}
	idx := gamekit.FirstEmptyBackpackSlot(inv)
	if idx < 0 {
		return fmt.Errorf("backpack full")
	}
	slot := fmt.Sprintf("%s%d", gamekit.InvSlotBackpackPref, idx)
	if cat != nil && !ItemFitsInventorySlot(cat, id, slot) {
		return fmt.Errorf("item %q cannot be placed in backpack", id)
	}
	inv.Backpack[idx] = id
	gamekit.NormalizeInventory(inv)
	return nil
}
