package config

import (
	"os"
	"strconv"
	"time"
)

// EnvWorldID — имя переменной окружения с id мира для лобби (оркестратор задаёт на инстанс).
const EnvWorldID = "WORLD_ID"

// EnvWorldServiceAddr — gRPC адрес world-service (host:port).
const EnvWorldServiceAddr = "WORLD_SERVICE_ADDR"

// EnvWorldServiceToken — секрет для metadata x-service-token (если включён на world-service).
const EnvWorldServiceToken = "WORLD_SERVICE_TOKEN"

// EnvSaveWorldAdminUserID — user_id из JWT, которому разрешён TypeSaveWorld (пусто = не трогать TOML).
const EnvSaveWorldAdminUserID = "SAVE_WORLD_ADMIN_USER_ID"

// EnvCharacterServiceAddr — gRPC character-service (host:port).
const EnvCharacterServiceAddr = "CHARACTER_SERVICE_ADDR"

// EnvCharacterServiceToken — metadata x-service-token для character-service.
const EnvCharacterServiceToken = "CHARACTER_SERVICE_TOKEN"

// EnvContentCatalogPath — путь к catalog.json (pkg/gamekit/content); пусто — без interact.
const EnvContentCatalogPath = "CONTENT_CATALOG_PATH"

// EnvContentScriptsDir — каталог JSON-сценариев взаимодействий; может быть пустым если в каталоге нет interact.
const EnvContentScriptsDir = "CONTENT_SCRIPTS_DIR"

// Config contains game-service runtime options.
type Config struct {
	BindAddr  string        `toml:"bind_addr"`
	LogLevel  string        `toml:"log_level"`
	TickRate  time.Duration `toml:"tick_rate"`
	QueueSize int           `toml:"queue_size"`
	// MovementApplyEveryNTicks — применять смещение по move-интенту не чаще чем раз в N тиков (1 = каждый тик). Не трогает частоту state/hit.
	MovementApplyEveryNTicks int    `toml:"movement_apply_every_n_ticks"`
	JWTSecret                string `toml:"jwt_secret"`
	// WorldID — id мира в world-service; из TOML или из EnvWorldID, если переменная задана.
	WorldID string `toml:"world_id"`
	// WorldServiceAddr — host:port gRPC world-service (нужен, если WorldID непустой).
	WorldServiceAddr  string `toml:"world_service_addr"`
	WorldServiceToken string `toml:"world_service_token"`
	// SaveWorldAdminUserID — только этот user_id может вызывать save_world (по умолчанию 1).
	SaveWorldAdminUserID int64 `toml:"save_world_admin_user_id"`
	// CharacterServiceAddr — host:port gRPC character-service; пусто — без Resolve/save персонажа (старое поведение WS).
	CharacterServiceAddr  string `toml:"character_service_addr"`
	CharacterServiceToken string `toml:"character_service_token"`
	// ContentCatalogPath — catalog.json; пусто — TypeInteract игнорируется.
	ContentCatalogPath string `toml:"content_catalog_path"`
	// ContentScriptsDir — каталог со сценариями (*.json); пусто допустимо если нет interact.script.
	ContentScriptsDir string `toml:"content_scripts_dir"`
	// TileFullSyncInterval — период полного списка тайлов в state; между полными снимками — только tile_updates.
	TileFullSyncInterval time.Duration `toml:"tile_full_sync_interval"`
}

// EnvTileFullSyncInterval — duration string (time.ParseDuration), перекрывает TOML.
const EnvTileFullSyncInterval = "TILE_FULL_SYNC_INTERVAL"

// NewConfig returns defaults for local/dev.
func NewConfig() *Config {
	return &Config{
		BindAddr:                 ":50053",
		LogLevel:                 "info",
		TickRate:                 50 * time.Millisecond,
		QueueSize:                1024,
		MovementApplyEveryNTicks: 1,
		JWTSecret:                "123",
		SaveWorldAdminUserID:     1,
		TileFullSyncInterval:     time.Second,
	}
}

// ApplyEnvOverrides подставляет значения из окружения поверх уже загруженного TOML.
// Переменная WORLD_ID: если задана в окружении инстанса (в т.ч. пустая строка), перезаписывает WorldID.
func ApplyEnvOverrides(cfg *Config) {
	if v, ok := os.LookupEnv(EnvWorldID); ok {
		cfg.WorldID = v
	}
	if v, ok := os.LookupEnv(EnvWorldServiceAddr); ok {
		cfg.WorldServiceAddr = v
	}
	if v, ok := os.LookupEnv(EnvWorldServiceToken); ok {
		cfg.WorldServiceToken = v
	}
	if v, ok := os.LookupEnv(EnvSaveWorldAdminUserID); ok {
		if v == "" {
			cfg.SaveWorldAdminUserID = 0
		} else if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			cfg.SaveWorldAdminUserID = id
		}
	}
	if v, ok := os.LookupEnv(EnvCharacterServiceAddr); ok {
		cfg.CharacterServiceAddr = v
	}
	if v, ok := os.LookupEnv(EnvCharacterServiceToken); ok {
		cfg.CharacterServiceToken = v
	}
	if v, ok := os.LookupEnv(EnvContentCatalogPath); ok {
		cfg.ContentCatalogPath = v
	}
	if v, ok := os.LookupEnv(EnvContentScriptsDir); ok {
		cfg.ContentScriptsDir = v
	}
	if v, ok := os.LookupEnv(EnvTileFullSyncInterval); ok && v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			cfg.TileFullSyncInterval = d
		}
	}
}
