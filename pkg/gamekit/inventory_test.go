package gamekit

import "testing"

func TestTrySwapInventorySlots_hands(t *testing.T) {
	var inv PlayerInventory
	inv.HandMain = "sword"
	inv.HandOff = "shield"
	if !TrySwapInventorySlots(&inv, InvSlotHandMain, InvSlotHandOff, func(string) bool { return true }) {
		t.Fatal("expected swap ok")
	}
	if inv.HandMain != "shield" || inv.HandOff != "sword" {
		t.Fatalf("after swap: main=%q off=%q", inv.HandMain, inv.HandOff)
	}
}

func TestTrySwapInventorySlots_rejectStorageInBackpack(t *testing.T) {
	var inv PlayerInventory
	inv.Backpack[0] = "gem"
	inv.HandMain = "big_bag"
	ok := TrySwapInventorySlots(&inv, InvSlotHandMain, "backpack_0", func(id string) bool {
		return id != "big_bag"
	})
	if ok {
		t.Fatal("expected reject putting big_bag into backpack")
	}
	if inv.HandMain != "big_bag" || inv.Backpack[0] != "gem" {
		t.Fatalf("state should be unchanged: %+v", inv)
	}
}

func TestParseInventorySlot(t *testing.T) {
	if _, _, ok := ParseInventorySlot("backpack_9"); ok {
		t.Fatal("expected invalid backpack index")
	}
	idx, isBP, ok := ParseInventorySlot("backpack_4")
	if !ok || !isBP || idx != 4 {
		t.Fatalf("backpack_4: idx=%d isBP=%v ok=%v", idx, isBP, ok)
	}
}
