package model

// SkillType is an individual skill that can be trained by a player.
type SkillType int

const (
	SkillTypeAttack SkillType = iota
	SkillTypeStrength
	SkillTypeDefense
	SkillTypeRanged
	SkillTypePrayer
	SkillTypeMagic
	SkillTypeRunecraft
	SkillTypeHitpoints
	SkillTypeAgility
	SkillTypeHerblore
	SkillTypeThieving
	SkillTypeCrafting
	SkillTypeFletching
	SkillTypeSlayer
	SkillTypeMining
	SkillTypeSmithing
	SkillTypeFishing
	SkillTypeCooking
	SkillTypeFiremaking
	SkillTypeWoodcutting
	SkillTypeFarming
)

// Skill represents progress in a single skill.
type Skill struct {
	Level      int
	Experience int
}

// NewSkill returns a Skill initialized to its base level.
func NewSkill() *Skill {
	return &Skill{
		Level:      1,
		Experience: 0,
	}
}

// SkillMap is a map of skill types to their progress. Use the EmptySkillMap() function to create a map of all skills
// initialized to their base levels.
type SkillMap map[SkillType]*Skill

// EmptySkillMap returns a SkillMap initialized with all skills to their base levels.
func EmptySkillMap() SkillMap {
	return SkillMap{
		SkillTypeAttack:      NewSkill(),
		SkillTypeStrength:    NewSkill(),
		SkillTypeDefense:     NewSkill(),
		SkillTypeRanged:      NewSkill(),
		SkillTypePrayer:      NewSkill(),
		SkillTypeMagic:       NewSkill(),
		SkillTypeRunecraft:   NewSkill(),
		SkillTypeHitpoints:   NewSkill(),
		SkillTypeAgility:     NewSkill(),
		SkillTypeHerblore:    NewSkill(),
		SkillTypeThieving:    NewSkill(),
		SkillTypeCrafting:    NewSkill(),
		SkillTypeFletching:   NewSkill(),
		SkillTypeSlayer:      NewSkill(),
		SkillTypeMining:      NewSkill(),
		SkillTypeSmithing:    NewSkill(),
		SkillTypeFishing:     NewSkill(),
		SkillTypeCooking:     NewSkill(),
		SkillTypeFiremaking:  NewSkill(),
		SkillTypeWoodcutting: NewSkill(),
		SkillTypeFarming:     NewSkill(),
	}
}