package content

import (
	"testing"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func TestItemFitsInventorySlot_emptyItemAlwaysOK(t *testing.T) {
	c := &Catalog{Items: map[string]ItemDef{"gem": {ID: "gem"}}}
	if !ItemFitsInventorySlot(c, "  ", gamekit.InvSlotHandMain) {
		t.Fatal("empty id should fit any slot")
	}
}

func TestItemFitsInventorySlot_nilCatalog(t *testing.T) {
	if ItemFitsInventorySlot(nil, "gem", "backpack_0") {
		t.Fatal("nil catalog => false")
	}
}

func TestItemFitsInventorySlot_unknownItem(t *testing.T) {
	c := &Catalog{Items: map[string]ItemDef{"gem": {ID: "gem"}}}
	if ItemFitsInventorySlot(c, "nope", "backpack_0") {
		t.Fatal("unknown id => false")
	}
}

func TestItemFitsInventorySlot_backpackStorageRejected(t *testing.T) {
	c := &Catalog{Items: map[string]ItemDef{
		"bag": {ID: "bag", IsStorage: true},
		"gem": {ID: "gem"},
	}}
	if ItemFitsInventorySlot(c, "bag", "backpack_0") {
		t.Fatal("is_storage must not fit backpack")
	}
	if !ItemFitsInventorySlot(c, "gem", "backpack_4") {
		t.Fatal("gem should fit backpack")
	}
}

func TestItemFitsInventorySlot_equipFlags(t *testing.T) {
	c := &Catalog{Items: map[string]ItemDef{
		"axe":   {ID: "axe", AllowHand: true},
		"robe":  {ID: "robe", AllowArmor: true},
		"ring":  {ID: "ring", AllowAccessory: true},
		"plain": {ID: "plain"},
	}}
	tests := []struct {
		item, slot string
		want        bool
	}{
		{"axe", gamekit.InvSlotHandMain, true},
		{"axe", gamekit.InvSlotArmor, false},
		{"robe", gamekit.InvSlotArmor, true},
		{"robe", gamekit.InvSlotHandMain, false},
		{"ring", gamekit.InvSlotAccessory1, true},
		{"ring", gamekit.InvSlotAccessory2, true},
		{"ring", gamekit.InvSlotHandOff, false},
		{"plain", gamekit.InvSlotHandMain, false},
		{"plain", "backpack_0", true},
	}
	for _, tc := range tests {
		got := ItemFitsInventorySlot(c, tc.item, tc.slot)
		if got != tc.want {
			t.Fatalf("%s in %s: got %v want %v", tc.item, tc.slot, got, tc.want)
		}
	}
}
