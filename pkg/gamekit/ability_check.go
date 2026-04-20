package gamekit

import (
	"math"
	"strings"
)

// Ability identifies one of the six D&D-style ability scores on CharacterStats.
type Ability uint8

const (
	AbilityStrength Ability = iota
	AbilityDexterity
	AbilityConstitution
	AbilityIntelligence
	AbilityWisdom
	AbilityCharisma
)

// AbilityScore returns the raw score for a from stats. Values <= 0 are treated as 10
// (unset or corrupted per server rules).
func AbilityScore(stats CharacterStats, a Ability) int {
	var v int
	switch a {
	case AbilityStrength:
		v = stats.Strength
	case AbilityDexterity:
		v = stats.Dexterity
	case AbilityConstitution:
		v = stats.Constitution
	case AbilityIntelligence:
		v = stats.Intelligence
	case AbilityWisdom:
		v = stats.Wisdom
	case AbilityCharisma:
		v = stats.Charisma
	default:
		v = 0
	}
	if v <= 0 {
		return 10
	}
	return v
}

// AbilityModifier returns the D&D 5e-style modifier: floor((score - 10) / 2).
// Scores <= 0 are treated as 10 before applying the formula.
func AbilityModifier(score int) int {
	if score <= 0 {
		score = 10
	}
	return int(math.Floor(float64(score-10) / 2.0))
}

// ParseAbility maps common names (str, strength, …) to Ability. Unknown input returns false.
func ParseAbility(s string) (Ability, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "str", "strength":
		return AbilityStrength, true
	case "dex", "dexterity":
		return AbilityDexterity, true
	case "con", "constitution":
		return AbilityConstitution, true
	case "int", "intelligence":
		return AbilityIntelligence, true
	case "wis", "wisdom":
		return AbilityWisdom, true
	case "cha", "charisma":
		return AbilityCharisma, true
	default:
		return 0, false
	}
}

// IntnRng is a minimal random source for ability checks (e.g. *rand.Rand from math/rand).
type IntnRng interface {
	Intn(n int) int
}

// AbilityCheckResult is the outcome of RollAbilityCheck (d20 + modifier vs DC).
type AbilityCheckResult struct {
	D20       int
	Modifier  int
	Total     int
	DC        int
	Success   bool
}

// RollAbilityCheck rolls 1d20 + ability modifier for the given stats and ability, then compares to dc.
// The die uses inclusive 1..20 via 1+rng.Intn(20). Critical hit/miss rules are not applied.
func RollAbilityCheck(rng IntnRng, stats CharacterStats, a Ability, dc int) AbilityCheckResult {
	score := AbilityScore(stats, a)
	mod := AbilityModifier(score)
	d20 := 1 + rng.Intn(20)
	total := d20 + mod
	return AbilityCheckResult{
		D20:      d20,
		Modifier: mod,
		Total:    total,
		DC:       dc,
		Success:  total >= dc,
	}
}
