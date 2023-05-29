package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_Player_SetSkillExperience_combatLevels_hitpoints(t *testing.T) {
	p := NewPlayer("mike")

	// level 10 attack
	p.SetSkillExperience(SkillTypeHitpoints, 1160)

	assert.Equal(t, 3, p.Appearance.CombatLevel)
}

func Test_Player_SetSkillExperience_combatLevels_melee(t *testing.T) {
	p := NewPlayer("mike")

	// level 4 attack
	p.SetSkillExperience(SkillTypeAttack, 277)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkillExperience_combatLevels_magic(t *testing.T) {
	p := NewPlayer("mike")

	// level 4 magic
	p.SetSkillExperience(SkillTypeMagic, 277)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkillExperience_combatLevels_ranged(t *testing.T) {
	p := NewPlayer("mike")

	// level 4 ranged
	p.SetSkillExperience(SkillTypeRanged, 277)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkillExperience_combatLevels_prayer(t *testing.T) {
	p := NewPlayer("mike")

	// level 10 prayer
	p.SetSkillExperience(SkillTypePrayer, 1160)

	assert.Equal(t, 2, p.Appearance.CombatLevel)
}

func Test_Player_SetSkillExperience_level(t *testing.T) {
	p := NewPlayer("mike")

	// level 2
	p.SetSkillExperience(SkillTypeFletching, 165)
	// level 10
	p.SetSkillExperience(SkillTypePrayer, 1160)

	assert.Equal(t, 2, p.Skills[SkillTypeFletching].Level)
	assert.Equal(t, 10, p.Skills[SkillTypePrayer].Level)
}
