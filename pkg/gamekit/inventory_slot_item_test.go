package gamekit

import "testing"

func TestInventorySlotHasItem_match(t *testing.T) {
	var inv PlayerInventory
	inv.HandMain = "axe"
	ok, err := InventorySlotHasItem(&inv, InvSlotHandMain, "axe")
	if err != nil || !ok {
		t.Fatalf("got ok=%v err=%v", ok, err)
	}
}

func TestInventorySlotHasItem_wrongSlot(t *testing.T) {
	var inv PlayerInventory
	inv.HandOff = "axe"
	ok, err := InventorySlotHasItem(&inv, InvSlotHandMain, "axe")
	if err != nil || ok {
		t.Fatalf("want mismatch: ok=%v err=%v", ok, err)
	}
}

func TestInventorySlotHasItem_inBackpackNotHand(t *testing.T) {
	var inv PlayerInventory
	inv.Backpack[0] = "axe"
	ok, err := InventorySlotHasItem(&inv, InvSlotHandMain, "axe")
	if err != nil || ok {
		t.Fatalf("axe only in backpack should not match hand: ok=%v err=%v", ok, err)
	}
}

func TestInventorySlotHasItem_nilInv(t *testing.T) {
	ok, err := InventorySlotHasItem(nil, InvSlotHandMain, "axe")
	if err != nil || ok {
		t.Fatalf("nil inv: ok=%v err=%v", ok, err)
	}
}

func TestInventorySlotHasItem_emptyWant(t *testing.T) {
	var inv PlayerInventory
	inv.HandMain = "axe"
	_, err := InventorySlotHasItem(&inv, InvSlotHandMain, "  ")
	if err == nil {
		t.Fatal("expected error for empty item_def_id")
	}
}

func TestInventorySlotHasItem_invalidSlot(t *testing.T) {
	var inv PlayerInventory
	_, err := InventorySlotHasItem(&inv, "pocket", "axe")
	if err == nil {
		t.Fatal("expected error for invalid slot")
	}
}

func TestInventorySlotHasItem_backpackSlot(t *testing.T) {
	var inv PlayerInventory
	inv.Backpack[2] = "gem"
	ok, err := InventorySlotHasItem(&inv, "backpack_2", "gem")
	if err != nil || !ok {
		t.Fatalf("got ok=%v err=%v", ok, err)
	}
}
