package model

// NumEquipmentSlots is the number of slots available for equipping items.
const NumEquipmentSlots = 12

// NumBodyParts is the number of customizable character body parts.
const NumBodyParts = 5

// EntityGender enumerates valid genders for players or NPCs.
type EntityGender int

const (
	EntityMale EntityGender = iota
	EntityFemale
)

type AnimationID int

const (
	AnimationStand AnimationID = iota
	AnimationStandTurn
	AnimationWalk
	AnimationTurnAbout
	AnimationTurnRight
	AnimationTurnLeft
	AnimationRun
)

// EntityAppearance describes the properties of an entity such as a player or NPC.
type EntityAppearance struct {
	NPCAppearance  int
	Equipment      []int
	Body           []int
	Animations     map[AnimationID]int
	Gender         EntityGender
	OverheadIconID int
	CombatLevel    int
	TotalLevel     int
	Updated        bool
}

// IsNPCAppearance returns if the appearance should take that of a predefined NPC.
func (a *EntityAppearance) IsNPCAppearance() bool {
	return a.Equipment[0] == 0xFFFF
}

// SetNPCAppearance sets the ID of an NPC to use for the appearance.
func (a *EntityAppearance) SetNPCAppearance(id int) {
	// clear out the first slot and set the npc appearance id
	a.Equipment[0] = 0xFFFF
	a.NPCAppearance = id
}
