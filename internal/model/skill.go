package model

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
	Experience int
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
