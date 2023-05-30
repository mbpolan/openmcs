package model

import (
	"math"
)

// SkillType is an individual skill that can be trained by a player. Each skill has a well-known identifier that is
// shared between the client and server.
type SkillType int

const (
	SkillTypeAttack      SkillType = 0
	SkillTypeDefense     SkillType = 1
	SkillTypeStrength    SkillType = 2
	SkillTypeHitpoints   SkillType = 3
	SkillTypeRanged      SkillType = 4
	SkillTypePrayer      SkillType = 5
	SkillTypeMagic       SkillType = 6
	SkillTypeCooking     SkillType = 7
	SkillTypeWoodcutting SkillType = 8
	SkillTypeFletching   SkillType = 9
	SkillTypeFishing     SkillType = 10
	SkillTypeFiremaking  SkillType = 11
	SkillTypeCrafting    SkillType = 12
	SkillTypeSmithing    SkillType = 13
	SkillTypeMining      SkillType = 14
	SkillTypeHerblore    SkillType = 15
	SkillTypeAgility     SkillType = 16
	SkillTypeThieving    SkillType = 17
	SkillTypeSlayer      SkillType = 18
	SkillTypeFarming     SkillType = 19
	SkillTypeRunecraft   SkillType = 20
)

// Skill represents progress in a single skill.
type Skill struct {
	Type       SkillType
	Level      int
	Experience float64
}

// NewSkill returns a Skill initialized to its base level.
func NewSkill(skillType SkillType) *Skill {
	return &Skill{
		Type:       skillType,
		Level:      1,
		Experience: 0,
	}
}

// SkillMap is a map of skill types to their progress. Use the EmptySkillMap() function to create a map of all skills
// initialized to their base levels.
type SkillMap map[SkillType]*Skill

// SkillExperienceLevels is a slice of experience points corresponding to each skill level.
var SkillExperienceLevels = func() []float64 {
	rawExp := make([]float64, 100)
	rawExp[1] = 0

	// compute the experience points for each level based on the difference from the prior level
	for i := 2; i <= 99; i++ {
		e := 0.25 * math.Floor(float64(i-1)+300.0*math.Pow(2, float64(i-1)/7.0))
		rawExp[i] = rawExp[i-1] + e
	}

	// normalize the raw experience points into integer values
	exp := make([]float64, 100)
	for i := 1; i <= 99; i++ {
		exp[i] = math.Floor(rawExp[i])
	}
	return exp
}()

// EmptySkillMap returns a SkillMap initialized with all skills to their base levels.
func EmptySkillMap() SkillMap {
	return SkillMap{
		SkillTypeAttack:      NewSkill(SkillTypeAttack),
		SkillTypeStrength:    NewSkill(SkillTypeStrength),
		SkillTypeDefense:     NewSkill(SkillTypeDefense),
		SkillTypeRanged:      NewSkill(SkillTypeRanged),
		SkillTypePrayer:      NewSkill(SkillTypePrayer),
		SkillTypeMagic:       NewSkill(SkillTypeMagic),
		SkillTypeRunecraft:   NewSkill(SkillTypeRunecraft),
		SkillTypeHitpoints:   NewSkill(SkillTypeHitpoints),
		SkillTypeAgility:     NewSkill(SkillTypeAgility),
		SkillTypeHerblore:    NewSkill(SkillTypeHerblore),
		SkillTypeThieving:    NewSkill(SkillTypeThieving),
		SkillTypeCrafting:    NewSkill(SkillTypeCrafting),
		SkillTypeFletching:   NewSkill(SkillTypeFletching),
		SkillTypeSlayer:      NewSkill(SkillTypeSlayer),
		SkillTypeMining:      NewSkill(SkillTypeMining),
		SkillTypeSmithing:    NewSkill(SkillTypeSmithing),
		SkillTypeFishing:     NewSkill(SkillTypeFishing),
		SkillTypeCooking:     NewSkill(SkillTypeCooking),
		SkillTypeFiremaking:  NewSkill(SkillTypeFiremaking),
		SkillTypeWoodcutting: NewSkill(SkillTypeWoodcutting),
		SkillTypeFarming:     NewSkill(SkillTypeFarming),
	}
}
