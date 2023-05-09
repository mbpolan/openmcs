package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Player_SetSkill_combatLevels_hitpoints(t *testing.T) {
	p := NewPlayer("mike")

	attack := NewSkill(SkillTypeHitpoints)
	attack.Level = 10
	p.SetSkill(attack)

	assert.Equal(t, 3, p.Appearance.CombatLevel)
}

func Test_Player_SetSkill_combatLevels_melee(t *testing.T) {
	p := NewPlayer("mike")

	attack := NewSkill(SkillTypeAttack)
	attack.Level = 4
	p.SetSkill(attack)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkill_combatLevels_magic(t *testing.T) {
	p := NewPlayer("mike")

	attack := NewSkill(SkillTypeMagic)
	attack.Level = 4
	p.SetSkill(attack)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkill_combatLevels_ranged(t *testing.T) {
	p := NewPlayer("mike")

	attack := NewSkill(SkillTypeRanged)
	attack.Level = 4
	p.SetSkill(attack)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkill_combatLevels_prayer(t *testing.T) {
	p := NewPlayer("mike")

	attack := NewSkill(SkillTypePrayer)
	attack.Level = 10
	p.SetSkill(attack)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}
