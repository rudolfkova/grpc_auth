package gamekit

import (
	"strings"
)

// BackpackSlotCount — число слотов в рюкзаке (без вложенных контейнеров).
const BackpackSlotCount = 5

// Имена слотов для inventory_move (from / to).
const (
	InvSlotArmor        = "armor"
	InvSlotAccessory1   = "accessory_1"
	InvSlotAccessory2   = "accessory_2"
	InvSlotHandMain     = "hand_main"
	InvSlotHandOff      = "hand_off"
	InvSlotBackpackPref = "backpack_" // + индекс 0..BackpackSlotCount-1
)

// DroppedItemTileLayer — слой тайла при выбросе предмета (как у игрока для сортировки; 0 — земля, 1 — лежащее на земле).
const DroppedItemTileLayer = 3

// PlayerInventory — экипировка, руки и рюкзак: в каждом слоте не более одного item_def_id (пустая строка = пусто).
type PlayerInventory struct {
	Armor        string `json:"armor,omitempty"`
	Accessory1   string `json:"accessory_1,omitempty"`
	Accessory2   string `json:"accessory_2,omitempty"`
	HandMain     string `json:"hand_main,omitempty"`
	HandOff      string `json:"hand_off,omitempty"`
	Backpack     [BackpackSlotCount]string `json:"backpack"`
}

// DefaultPlayerInventory пустой инвентарь.
func DefaultPlayerInventory() PlayerInventory {
	return PlayerInventory{}
}

// NormalizeInventory чистит пробелы и длину рюкзака (после JSON с меньшим числом элементов).
func NormalizeInventory(inv *PlayerInventory) {
	inv.Armor = strings.TrimSpace(inv.Armor)
	inv.Accessory1 = strings.TrimSpace(inv.Accessory1)
	inv.Accessory2 = strings.TrimSpace(inv.Accessory2)
	inv.HandMain = strings.TrimSpace(inv.HandMain)
	inv.HandOff = strings.TrimSpace(inv.HandOff)
	for i := range inv.Backpack {
		inv.Backpack[i] = strings.TrimSpace(inv.Backpack[i])
	}
}

// ParseInventorySlot: для рюкзака backpackIdx 0..n-1 и isBackpack=true; для остальных слотов isBackpack=false.
func ParseInventorySlot(slot string) (backpackIdx int, isBackpack bool, ok bool) {
	slot = strings.TrimSpace(slot)
	switch slot {
	case InvSlotArmor, InvSlotAccessory1, InvSlotAccessory2, InvSlotHandMain, InvSlotHandOff:
		return -1, false, true
	}
	if strings.HasPrefix(slot, InvSlotBackpackPref) {
		suf := slot[len(InvSlotBackpackPref):]
		if len(suf) != 1 {
			return -1, false, false
		}
		c := suf[0]
		if c < '0' || c > '0'+byte(BackpackSlotCount-1) {
			return -1, false, false
		}
		return int(c - '0'), true, true
	}
	return -1, false, false
}

// ItemAtSlot возвращает item_def_id в слоте (имена как InvSlot*, backpack_0..).
func (inv *PlayerInventory) ItemAtSlot(slot string) (itemDefID string, ok bool) {
	return inv.getSlot(slot)
}

// PutItemAtSlot записывает item_def_id в слот; пустая строка очищает слот.
func (inv *PlayerInventory) PutItemAtSlot(slot, itemDefID string) bool {
	return inv.setSlot(slot, itemDefID)
}

func (inv *PlayerInventory) getSlot(slot string) (string, bool) {
	slot = strings.TrimSpace(slot)
	switch slot {
	case InvSlotArmor:
		return inv.Armor, true
	case InvSlotAccessory1:
		return inv.Accessory1, true
	case InvSlotAccessory2:
		return inv.Accessory2, true
	case InvSlotHandMain:
		return inv.HandMain, true
	case InvSlotHandOff:
		return inv.HandOff, true
	}
	if idx, isBP, ok := ParseInventorySlot(slot); ok && isBP {
		return inv.Backpack[idx], true
	}
	return "", false
}

func (inv *PlayerInventory) setSlot(slot, itemDefID string) bool {
	slot = strings.TrimSpace(slot)
	itemDefID = strings.TrimSpace(itemDefID)
	switch slot {
	case InvSlotArmor:
		inv.Armor = itemDefID
		return true
	case InvSlotAccessory1:
		inv.Accessory1 = itemDefID
		return true
	case InvSlotAccessory2:
		inv.Accessory2 = itemDefID
		return true
	case InvSlotHandMain:
		inv.HandMain = itemDefID
		return true
	case InvSlotHandOff:
		inv.HandOff = itemDefID
		return true
	}
	if idx, isBP, ok := ParseInventorySlot(slot); ok && isBP {
		inv.Backpack[idx] = itemDefID
		return true
	}
	return false
}

// TrySwapInventorySlots меняет местами содержимое двух слотов (или no-op при одинаковых / невалидных).
// toBackpackReject если целевой слот — рюкзак и itemDefID нельзя класть в рюкзак (is_storage).
func TrySwapInventorySlots(inv *PlayerInventory, from, to string, canPlaceInBackpack func(itemDefID string) bool) bool {
	from = strings.TrimSpace(from)
	to = strings.TrimSpace(to)
	if from == "" || to == "" || from == to {
		return false
	}
	if _, _, ok := ParseInventorySlot(from); !ok {
		return false
	}
	if _, _, ok := ParseInventorySlot(to); !ok {
		return false
	}
	a, ok1 := inv.getSlot(from)
	b, ok2 := inv.getSlot(to)
	if !ok1 || !ok2 {
		return false
	}
	if strings.TrimSpace(a) != "" && !canPlaceInSlot(to, a, canPlaceInBackpack) {
		return false
	}
	if strings.TrimSpace(b) != "" && !canPlaceInSlot(from, b, canPlaceInBackpack) {
		return false
	}
	_ = inv.setSlot(from, b)
	_ = inv.setSlot(to, a)
	NormalizeInventory(inv)
	return true
}

// FirstEmptyBackpackSlot возвращает индекс 0..BackpackSlotCount-1 или -1, если рюкзак полон.
func FirstEmptyBackpackSlot(inv *PlayerInventory) int {
	if inv == nil {
		return -1
	}
	for i := range inv.Backpack {
		if strings.TrimSpace(inv.Backpack[i]) == "" {
			return i
		}
	}
	return -1
}

// ChebyshevDist1 возвращает true, если клетки соседние по Чебышёву (включая диагональ, включая совпадение).
func ChebyshevDist1(ax, ay, bx, by int) bool {
	dx := ax - bx
	if dx < 0 {
		dx = -dx
	}
	dy := ay - by
	if dy < 0 {
		dy = -dy
	}
	if dx > dy {
		return dx <= 1
	}
	return dy <= 1
}

func canPlaceInSlot(slot, itemDefID string, canPlaceInBackpack func(itemDefID string) bool) bool {
	if _, isBP, ok := ParseInventorySlot(slot); ok && isBP {
		return canPlaceInBackpack(itemDefID)
	}
	return true
}
