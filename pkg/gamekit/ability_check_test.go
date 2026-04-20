package gamekit

import (
	"math/rand"
	"testing"
)

func TestAbilityModifier_DnDFloor(t *testing.T) {
	tests := []struct {
		score int
		want  int
	}{
		{1, -5},
		{7, -2},
		{8, -1},
		{9, -1},
		{10, 0},
		{11, 0},
		{12, 1},
		{14, 2},
		{18, 4},
		{20, 5},
	}
	for _, tt := range tests {
		if got := AbilityModifier(tt.score); got != tt.want {
			t.Errorf("AbilityModifier(%d) = %d, want %d", tt.score, got, tt.want)
		}
	}
}

func TestAbilityModifier_nonPositiveScoreUsesTen(t *testing.T) {
	if got := AbilityModifier(0); got != 0 {
		t.Errorf("AbilityModifier(0) = %d, want 0 (treated as 10)", got)
	}
	if got := AbilityModifier(-3); got != 0 {
		t.Errorf("AbilityModifier(-3) = %d, want 0 (treated as 10)", got)
	}
}

func TestAbilityScore_invalidAbilityTreatsAsUnset(t *testing.T) {
	s := CharacterStats{Strength: 14}
	if got := AbilityScore(s, Ability(99)); got != 10 {
		t.Errorf("AbilityScore(..., invalid) = %d, want 10", got)
	}
}

func TestAbilityScore_nonPositiveUsesTen(t *testing.T) {
	s := CharacterStats{Strength: 0, Dexterity: -1, Wisdom: 12}
	if got := AbilityScore(s, AbilityStrength); got != 10 {
		t.Errorf("Strength 0 -> %d, want 10", got)
	}
	if got := AbilityScore(s, AbilityDexterity); got != 10 {
		t.Errorf("Dexterity -1 -> %d, want 10", got)
	}
	if got := AbilityScore(s, AbilityWisdom); got != 12 {
		t.Errorf("Wisdom -> %d, want 12", got)
	}
}

type fixedRNG struct{ v int }

func (f *fixedRNG) Intn(n int) int {
	if n <= 0 {
		return 0
	}
	return f.v % n
}

func TestRollAbilityCheck_deterministic(t *testing.T) {
	rng := &fixedRNG{v: 11}
	stats := CharacterStats{Strength: 14}
	res := RollAbilityCheck(rng, stats, AbilityStrength, 14)
	if res.D20 != 12 {
		t.Errorf("D20 = %d, want 12 (1 + 11%%20)", res.D20)
	}
	if res.Modifier != 2 {
		t.Errorf("Modifier = %d, want 2", res.Modifier)
	}
	if res.Total != 14 {
		t.Errorf("Total = %d, want 14", res.Total)
	}
	if res.DC != 14 {
		t.Errorf("DC = %d", res.DC)
	}
	if !res.Success {
		t.Fatal("expected success (14 >= 14)")
	}
}

func TestRollAbilityCheck_fail(t *testing.T) {
	rng := &fixedRNG{v: 0}
	stats := CharacterStats{Strength: 10}
	res := RollAbilityCheck(rng, stats, AbilityStrength, 15)
	if res.D20 != 1 {
		t.Errorf("D20 = %d, want 1", res.D20)
	}
	if res.Success {
		t.Fatal("expected failure for total 1 vs DC 15")
	}
}

func TestParseAbility(t *testing.T) {
	if a, ok := ParseAbility("STR"); !ok || a != AbilityStrength {
		t.Fatalf("STR: ok=%v a=%v", ok, a)
	}
	if a, ok := ParseAbility("wisdom"); !ok || a != AbilityWisdom {
		t.Fatalf("wisdom: ok=%v a=%v", ok, a)
	}
	if _, ok := ParseAbility("nope"); ok {
		t.Fatal("expected false for unknown")
	}
}

func TestRollAbilityCheck_withMathRand(t *testing.T) {
	rng := rand.New(rand.NewSource(42))
	stats := DefaultCharacterStats()
	for i := 0; i < 50; i++ {
		res := RollAbilityCheck(rng, stats, AbilityDexterity, 10)
		if res.D20 < 1 || res.D20 > 20 {
			t.Fatalf("D20 out of range: %d", res.D20)
		}
		if res.Modifier != 0 {
			t.Fatalf("Modifier = %d for 10 DEX", res.Modifier)
		}
		if res.Total != res.D20 {
			t.Fatalf("Total mismatch")
		}
		if res.Success != (res.Total >= 10) {
			t.Fatalf("Success inconsistent")
		}
	}
}
