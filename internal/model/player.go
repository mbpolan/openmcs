package model

import (
	"math"
	"strings"
)

type PlayerType int

// NumPlayerSkills is the number of available player skills.
const NumPlayerSkills = 21

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// Player is a human player connected to the game server. This struct stores a player's persistent data, including
// various preferences, game world properties and other such attributes.
type Player struct {
	ID           int
	Username     string
	PasswordHash string
	Type         PlayerType
	Flagged      bool
	GlobalPos    Vector3D
	Appearance   *EntityAppearance
	Modes        PlayerModes
	Muted        bool
	Friends      []string
	Ignored      []string
	Skills       SkillMap
}

// PlayerModes indicates what types of chat and interactions a player wishes to receive.
type PlayerModes struct {
	PublicChat  ChatMode
	PrivateChat ChatMode
	Interaction InteractionMode
}

// NewPlayer returns a new player model.
func NewPlayer(username string) *Player {
	// define a default appearance
	appearance := &EntityAppearance{
		Animations: map[AnimationID]int{
			AnimationStand:     0x080D, // standing
			AnimationStandTurn: 0xFFFF, // turning
			AnimationWalk:      0x067C, // walk
			AnimationTurnAbout: 0xFFFF, // turn about
			AnimationTurnRight: 0xFFFF, // turn right
			AnimationTurnLeft:  0xFFFF, // turn left
			AnimationRun:       0x067D, // run
		},
		Equipment: make([]int, NumEquipmentSlots),
		Body:      make([]int, NumBodyParts),
		Updated:   false,
	}

	return &Player{
		Username:   username,
		Appearance: appearance,
		Skills:     EmptySkillMap(),
	}
}

// SetSkill sets the data for a player skill. This will recompute the player's derived levels (total and combat).
func (p *Player) SetSkill(skill *Skill) {
	p.Skills[skill.Type] = skill
	p.recomputeSkills()
}

// HasFriend determines if the given player username is on this player's friends list.
func (p *Player) HasFriend(username string) bool {
	target := strings.ToLower(username)
	for _, friend := range p.Friends {
		if strings.ToLower(friend) == target {
			return true
		}
	}

	return false
}

// IsIgnored determines if the given player username is on this player's ignore list.
func (p *Player) IsIgnored(username string) bool {
	target := strings.ToLower(username)
	for _, ignored := range p.Ignored {
		if strings.ToLower(ignored) == target {
			return true
		}
	}

	return false
}

// recomputeSkills updates the derived levels (total and combat) based on the player's current skills.
func (p *Player) recomputeSkills() {
	totalLevel := 0

	// compute the total skill level
	for _, skill := range p.Skills {
		totalLevel += skill.Level
	}

	p.Appearance.TotalLevel = totalLevel

	// compute each component of the combat level
	base := 0.25 * (float64(p.Skills[SkillTypeDefense].Level+p.Skills[SkillTypeHitpoints].Level) + math.Floor(float64(p.Skills[SkillTypePrayer].Level)*0.5))
	melee := 0.325 * (float64(p.Skills[SkillTypeAttack].Level + p.Skills[SkillTypeStrength].Level))
	ranged := 0.325 * math.Floor(float64(p.Skills[SkillTypeRanged].Level)*1.5)
	magic := 0.325 * math.Floor(float64(p.Skills[SkillTypeMagic].Level)*1.5)

	// compute the final combat level
	p.Appearance.CombatLevel = int(math.Floor(base + math.Max(math.Max(melee, ranged), magic)))
}
