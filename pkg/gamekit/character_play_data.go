package gamekit

import (
	"bytes"
	"encoding/json"
)

// CharacterDataSchemaVersion — версия JSON в character-service.data (поднимать при несовместимых изменениях).
const CharacterDataSchemaVersion int32 = 2

// CharacterStats — классические характеристики (контракт с character.data и событием state).
// Значения по умолчанию для нового персонажа: 10 по каждой характеристике (как типичный модификатор +0 в D&D-подобных системах).
type CharacterStats struct {
	Strength     int `json:"str"`
	Dexterity    int `json:"dex"`
	Constitution int `json:"con"`
	Intelligence int `json:"int"`
	Wisdom       int `json:"wis"`
	Charisma     int `json:"cha"`
}

// DefaultCharacterStats возвращает «стандартный» набор для редактора и спавна.
func DefaultCharacterStats() CharacterStats {
	return CharacterStats{
		Strength: 10, Dexterity: 10, Constitution: 10,
		Intelligence: 10, Wisdom: 10, Charisma: 10,
	}
}

// IsUnset считает, что stats не заданы (все нули — например после разбора старого JSON без поля stats).
func (s CharacterStats) IsUnset() bool {
	return s.Strength == 0 && s.Dexterity == 0 && s.Constitution == 0 &&
		s.Intelligence == 0 && s.Wisdom == 0 && s.Charisma == 0
}

// CharacterPlayData — JSON в character-service.data (opaque bytes) + зеркало для ECS при join/save.
// Поле schema_version в JSON опционально для обратной совместимости со старыми сохранениями (v1 без stats).
type CharacterPlayData struct {
	SchemaVersion int            `json:"schema_version,omitempty"`
	X             int            `json:"x"`
	Y             int            `json:"y"`
	HP            int            `json:"hp"`
	FaceDX        int            `json:"face_dx"`
	FaceDY        int            `json:"face_dy"`
	Stats         CharacterStats `json:"stats"`
}

// NewDefaultCharacterPlayData — шаблон для пустого character.data.
func NewDefaultCharacterPlayData() CharacterPlayData {
	return CharacterPlayData{
		SchemaVersion: int(CharacterDataSchemaVersion),
		HP:            DefaultPlayerHP,
		FaceDX:        DefaultPlayerFaceDX,
		FaceDY:        DefaultPlayerFaceDY,
		Stats:         DefaultCharacterStats(),
	}
}

// Normalize заполняет пропуски после json.Unmarshal (legacy v1, пустые stats).
func (d *CharacterPlayData) Normalize() {
	if d.HP <= 0 {
		d.HP = DefaultPlayerHP
	}
	if d.FaceDX == 0 && d.FaceDY == 0 {
		d.FaceDX, d.FaceDY = DefaultPlayerFaceDX, DefaultPlayerFaceDY
	}
	if d.Stats.IsUnset() {
		d.Stats = DefaultCharacterStats()
	}
	if d.SchemaVersion < int(CharacterDataSchemaVersion) {
		d.SchemaVersion = int(CharacterDataSchemaVersion)
	}
}

// ParseCharacterPlayData разбирает bytes из character-service; пустой слайс → дефолты.
func ParseCharacterPlayData(data []byte) CharacterPlayData {
	d := NewDefaultCharacterPlayData()
	if len(bytes.TrimSpace(data)) == 0 {
		return d
	}
	_ = json.Unmarshal(data, &d)
	d.Normalize()
	return d
}

// MarshalCharacterPlayData сериализует для ReplaceCharacterData / тела при создании персонажа вручную.
func MarshalCharacterPlayData(d CharacterPlayData) ([]byte, error) {
	d.Normalize()
	d.SchemaVersion = int(CharacterDataSchemaVersion)
	return json.Marshal(d)
}
