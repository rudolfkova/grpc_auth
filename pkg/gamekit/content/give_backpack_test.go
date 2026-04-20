package content

import (
	"testing"

	"github.com/rudolfkova/grpc_auth/pkg/gamekit"
)

func TestTryPutItemInFirstBackpack_ok(t *testing.T) {
	cat := &Catalog{Items: map[string]ItemDef{
		"log": {ID: "log", Pickable: true},
	}}
	var inv gamekit.PlayerInventory
	if err := TryPutItemInFirstBackpack(&inv, cat, "log"); err != nil {
		t.Fatal(err)
	}
	if inv.Backpack[0] != "log" {
		t.Fatalf("got %+v", inv.Backpack)
	}
}

func TestTryPutItemInFirstBackpack_full(t *testing.T) {
	cat := &Catalog{Items: map[string]ItemDef{
		"log": {ID: "log", Pickable: true},
	}}
	var inv gamekit.PlayerInventory
	for i := range inv.Backpack {
		inv.Backpack[i] = "log"
	}
	if err := TryPutItemInFirstBackpack(&inv, cat, "log"); err == nil {
		t.Fatal("expected error when full")
	}
}

func TestTryPutItemInFirstBackpack_storageRejected(t *testing.T) {
	cat := &Catalog{Items: map[string]ItemDef{
		"bag": {ID: "bag", Pickable: true, IsStorage: true},
	}}
	var inv gamekit.PlayerInventory
	if err := TryPutItemInFirstBackpack(&inv, cat, "bag"); err == nil {
		t.Fatal("expected error for storage in backpack")
	}
}

func TestTryPutItemInFirstBackpack_nilCatalogAllows(t *testing.T) {
	var inv gamekit.PlayerInventory
	if err := TryPutItemInFirstBackpack(&inv, nil, "anything"); err != nil {
		t.Fatal(err)
	}
	if inv.Backpack[0] != "anything" {
		t.Fatalf("got %q", inv.Backpack[0])
	}
}
