package gamekit

import (
	"encoding/json"
	"time"
)

// Имена сервиса и типов сообщений по wire (договорённость с game-service).
const (
	ServiceGame = "game"

	TypeMove          = "move"
	TypeHit           = "hit"
	TypeSpawnTile     = "spawn_tile"
	TypeClearTile     = "clear_tile"
	TypeSaveWorld     = "save_world"
	TypeInteract      = "interact"
	TypePickupItem    = "pickup_item"
	TypeDropItem      = "drop_item"
	TypeInventoryMove = "inventory_move"
	TypeState         = "state"
	TypeReject        = "reject"
	TypeError         = "error"
	// TypeSaveWorldResult — ответ на save_world (только инициатору).
	TypeSaveWorldResult = "save_world_result"
)

// Константы op для StatePayload.TileUpdates (дельта тайлов).
const (
	StateTileUpdateUpsert = "upsert"
	StateTileUpdateRemove = "remove"
)

// SnapshotSchemaVersion — schema_version для снимка ark-serde, который пишет game-service в world-service.
const SnapshotSchemaVersion int32 = 1

// Envelope — обёртка WebSocket JSON (клиент ↔ game-service).
type Envelope struct {
	Service string          `json:"service"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// MoveIntent — payload для TypeMove: обновляет «удерживаемое» направление (после clamp к Speed.MaxStep).
// Физический шаг по сетке выполняется не чаще одного раза за игровой тик на сервере; частые сообщения только обновляют интент.
// Отправьте dx=0, dy=0, чтобы сбросить движение.
type MoveIntent struct {
	DX int `json:"dx"`
	DY int `json:"dy"`
}

// HitIntent — payload для TypeHit.
type HitIntent struct {
	TargetID int64 `json:"target_id"`
	Damage   int   `json:"damage"`
}

// InvisibleTileTextureKey — wire-ключ 1×1 прозрачного PNG; интеракт и каталог задаются через instance_args.item_def_id.
const InvisibleTileTextureKey = "invisible"

// TileSpawnIntent — payload для TypeSpawnTile.
// Layer по умолчанию 0; Rotation — четверти оборота по часовой стрелке (любое целое нормализуется к 0..3).
type TileSpawnIntent struct {
	X        int    `json:"x"`
	Y        int    `json:"y"`
	Layer    int    `json:"layer"`
	Rotation int    `json:"rotation"`
	Texture  string `json:"texture"`
	Blocks   bool   `json:"blocks"`
	// InstanceArgs опционально: только JSON-объект; null/не объект при spawn отбрасываются (см. нормализацию на сервере).
	InstanceArgs json.RawMessage `json:"instance_args,omitempty"`
}

// TileClearIntent — payload для TypeClearTile: удалить все тайлы в клетке (x,y) на указанном слое.
type TileClearIntent struct {
	X     int `json:"x"`
	Y     int `json:"y"`
	Layer int `json:"layer"`
}

// InventoryMoveIntent — payload для TypeInventoryMove: обмен (swap) предметами между двумя слотами игрока.
// Имена слотов: armor, accessory_1, accessory_2, hand_main, hand_off, backpack_0..backpack_4.
type InventoryMoveIntent struct {
	From string `json:"from"`
	To   string `json:"to"`
}

// DropItemIntent — payload для TypeDropItem: выбросить предмет из слота под ноги (тайл на слое DroppedItemTileLayer).
type DropItemIntent struct {
	// From — имя слота (как для inventory_move): armor, accessory_1, accessory_2, hand_main, hand_off, backpack_0..4.
	From string `json:"from"`
}

// PickupIntent — payload для TypePickupItem: подобрать pickable-тайл с пола в первый свободный слот рюкзака.
// Поля как у InteractIntent: нужны item_def_id и оба click_x, click_y; click_layer — опционально (как при interact).
type PickupIntent struct {
	ItemDefID  string `json:"item_def_id"`
	ClickX     *int   `json:"click_x,omitempty"`
	ClickY     *int   `json:"click_y,omitempty"`
	ClickLayer *int   `json:"click_layer"`
}

// InteractIntent — payload для TypeInteract: запуск сценария каталога для item_def_id.
// Резолв по клетке (опционально): при обоих click_x и click_y ищется тайл с texture == item_def_id
// и interact в каталоге; при отсутствии click_layer — слой с максимальным layer среди подходящих.
// ClickLayer без omitempty: явный слой 0 сериализуется из Go как "click_layer":0; при nil в json.Marshal будет null (клиенту WS удобнее опускать ключ в JSON вручную).
type InteractIntent struct {
	ItemDefID  string `json:"item_def_id"`
	ClickX     *int   `json:"click_x,omitempty"`
	ClickY     *int   `json:"click_y,omitempty"`
	ClickLayer *int   `json:"click_layer"`
}

// SaveWorldIntent — payload для TypeSaveWorld (только разрешённый admin user_id на сервере).
type SaveWorldIntent struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// SaveWorldResultPayload — payload для TypeSaveWorldResult.
type SaveWorldResultPayload struct {
	Ok      bool   `json:"ok"`
	Code    string `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
	WorldID string `json:"world_id,omitempty"`
	Name    string `json:"name,omitempty"`
	Version int64  `json:"version,omitempty"`
}

// RejectPayload — стабильный payload для TypeReject.
// Используется при отклонении входящего envelope на уровне транспорта/очереди.
type RejectPayload struct {
	Reason         string `json:"reason"`
	Message        string `json:"message"`
	RequestType    string `json:"request_type,omitempty"`
	RequestService string `json:"request_service,omitempty"`
}

// ErrorPayload — payload для TypeError (например, invalid token при open WS).
type ErrorPayload struct {
	Message string `json:"message"`
}

// Player — проекция игрока в payload события TypeState (см. StatePayload.Players).
// Тот же тип используется в game-service при Broadcast; дублировать поля в других пакетах не нужно.
type Player struct {
	ID     int64          `json:"id"`
	X      int            `json:"x"`
	Y      int            `json:"y"`
	HP     int            `json:"hp"`
	FaceDX int            `json:"face_dx"`
	FaceDY int            `json:"face_dy"`
	Stats  CharacterStats `json:"stats"`
	// Sprite — id листа ходьбы (как CharacterPlayData.Sprite); клиент: data/anim/<sprite>/<sprite>.png.
	Sprite string `json:"sprite"`
	// Inventory — экипировка, руки, рюкзак (item_def_id на слот).
	Inventory PlayerInventory `json:"inventory"`
}

// Tile — элемент массива tiles в payload события TypeState.
type Tile struct {
	X            int             `json:"x"`
	Y            int             `json:"y"`
	Layer        int             `json:"layer"`
	Rotation     int             `json:"rotation"`
	Texture      string          `json:"texture"`
	Blocks       bool            `json:"blocks"`
	InstanceArgs json.RawMessage `json:"instance_args,omitempty"`
}

// TileUpdate — одно изменение тайла в дельте (type state, поле tile_updates).
type TileUpdate struct {
	Op string `json:"op"` // StateTileUpdateUpsert | StateTileUpdateRemove
	// Upsert: полный тайл как в tiles[].
	Tile *Tile `json:"tile,omitempty"`
	// Remove: клетка + слой (все сущности тайла на этом слое удалены).
	// Без omitempty: (0,0,0) должны сериализоваться.
	X     int `json:"x"`
	Y     int `json:"y"`
	Layer int `json:"layer"`
}

// StatePayload — полный JSON payload у TypeState (сервер шлёт это же из gamekit; клиент Unmarshal сюда).
// Тайлы: либо полный снимок (*Tiles), либо только дельта (TileUpdates), либо оба пусты за тик без изменений.
// Players и TickAt приходят каждый тик.
type StatePayload struct {
	Players     []Player     `json:"players"`
	Tiles       *[]Tile      `json:"tiles,omitempty"`
	TileUpdates []TileUpdate `json:"tile_updates,omitempty"`
	TickAt      time.Time    `json:"tick_at"`
}

// NormalizeTileRotationQuarter приводит произвольное целое к диапазону 0..3 (четверти оборота по часовой стрелке).
func NormalizeTileRotationQuarter(r int) int {
	return ((r % 4) + 4) % 4
}
