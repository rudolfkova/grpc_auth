package gamekit

import (
	"encoding/json"
	"testing"
)

func TestParseCharacterPlayData_LegacyWithoutStats(t *testing.T) {
	legacy := []byte(`{"x":3,"y":4,"hp":12,"face_dx":0,"face_dy":1}`)
	d := ParseCharacterPlayData(legacy)
	if d.X != 3 || d.Y != 4 || d.HP != 12 {
		t.Fatalf("position/hp: %+v", d)
	}
	if d.Stats.Strength != 10 || d.Stats.Charisma != 10 {
		t.Fatalf("expected default stats, got %+v", d.Stats)
	}
}

func TestMarshalCharacterPlayData_RoundTrip(t *testing.T) {
	d := NewDefaultCharacterPlayData()
	d.X, d.Y = 1, 2
	d.Stats.Strength = 16
	d.Stats.Dexterity = 14
	raw, err := MarshalCharacterPlayData(d)
	if err != nil {
		t.Fatal(err)
	}
	var got CharacterPlayData
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	got.Normalize()
	if got.Stats.Strength != 16 || got.Stats.Dexterity != 14 {
		t.Fatalf("stats: %+v", got.Stats)
	}
}
