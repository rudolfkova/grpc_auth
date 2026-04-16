package content

// CatalogSchemaVersion — версия формата catalog.json.
const CatalogSchemaVersion = 1

// Catalog — статический каталог предметов (редактор/клиент могут читать тот же JSON).
type Catalog struct {
	SchemaVersion int                `json:"schema_version"`
	Items         map[string]ItemDef `json:"items"`
}

// ItemDef описание предмета в каталоге.
type ItemDef struct {
	ID       string        `json:"id"`
	Name     string        `json:"name"`
	Pickable bool          `json:"pickable"`
	// IsStorage — предмет-«контейнер»; в рюкзак класть нельзя (нет сумки в сумке).
	IsStorage bool          `json:"is_storage,omitempty"`
	Interact  *InteractSpec `json:"interact,omitempty"`
}

// InteractSpec ссылка на файл сценария и аргументы по умолчанию (мержатся в каждый шаг).
type InteractSpec struct {
	Script string `json:"script"`
	Args   map[string]any `json:"args,omitempty"`
	// EditorInstanceArgsExample только для тулзов/редактора; рантайм сценариев не читает.
	EditorInstanceArgsExample map[string]any `json:"editor_instance_args_example,omitempty"`
}

// ScenarioSchemaVersion — версия формата сценария.
const ScenarioSchemaVersion = 1

// Scenario — декларативный сценарий (файл в scripts/).
type Scenario struct {
	SchemaVersion int    `json:"schema_version"`
	Steps         []Step `json:"steps"`
}

// Step один шаг: имя op и аргументы (поверх args из каталога).
type Step struct {
	Op   string         `json:"op"`
	Args map[string]any `json:"args,omitempty"`
}
