package content

import (
	"strings"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

// ItemFitsInventorySlot reports whether itemDefID may occupy slot according to cat.
// Empty itemDefID is always allowed (clearing a slot during swap).
// Unknown itemDefID or nil cat returns false.
func ItemFitsInventorySlot(cat *Catalog, itemDefID, slot string) bool {
	if cat == nil {
		return false
	}
	id := strings.TrimSpace(itemDefID)
	if id == "" {
		return true
	}
	slot = strings.TrimSpace(slot)
	def, ok := cat.Items[id]
	if !ok {
		return false
	}
	_, isBP, okSlot := gamekit.ParseInventorySlot(slot)
	if !okSlot {
		return false
	}
	if isBP {
		return !def.IsStorage
	}
	switch slot {
	case gamekit.InvSlotArmor:
		return def.AllowArmor
	case gamekit.InvSlotAccessory1, gamekit.InvSlotAccessory2:
		return def.AllowAccessory
	case gamekit.InvSlotHandMain, gamekit.InvSlotHandOff:
		return def.AllowHand
	default:
		return false
	}
}
