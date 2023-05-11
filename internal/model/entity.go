package model

// NumEquipmentSlots is the number of slots available for equipping items.
const NumEquipmentSlots = 12

// EquipmentSlot enumerates what equipment slot corresponds to what body part.
type EquipmentSlot int

const (
	EquipmentSlotHead        EquipmentSlot = 0
	EquipmentSlotCape                      = 1
	EquipmentSlotNecklace                  = 2
	EquipmentSlotPrimaryHand               = 3
	EquipmentSlotBody                      = 4
	EquipmentSlotOffHand                   = 5
	EquipmentSlotFace                      = 6
	EquipmentSlotLegs                      = 7
	EquipmentSlotHands                     = 9
	EquipmentSlotFeet                      = 10
	EquipmentSlotRing                      = 12
	EquipmentSlotAmmo                      = 13
)

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
	Equipment      map[EquipmentSlot]int
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
