package model

// NumEquipmentSlots is the number of slots available for equipping items.
const NumEquipmentSlots = 12

// EquipmentSlotType enumerates what equipment slot corresponds to what body part.
type EquipmentSlotType int

const (
	EquipmentSlotTypeHead     EquipmentSlotType = 0
	EquipmentSlotTypeCape                       = 1
	EquipmentSlotTypeNecklace                   = 2
	EquipmentSlotTypeWeapon                     = 3
	EquipmentSlotTypeBody                       = 4
	EquipmentSlotTypeShield                     = 5
	EquipmentSlotTypeFace                       = 6
	EquipmentSlotTypeLegs                       = 7
	EquipmentSlotTypeHands                      = 9
	EquipmentSlotTypeFeet                       = 10
	EquipmentSlotTypeRing                       = 12
	EquipmentSlotTypeAmmo                       = 13
)

// EquipmentSlotTypes is a slice of all EquipmentSlotType enums sorted according to their slot IDs in ascending order.
var EquipmentSlotTypes = []EquipmentSlotType{
	EquipmentSlotTypeHead,
	EquipmentSlotTypeCape,
	EquipmentSlotTypeNecklace,
	EquipmentSlotTypeWeapon,
	EquipmentSlotTypeBody,
	EquipmentSlotTypeShield,
	EquipmentSlotTypeFace,
	EquipmentSlotTypeLegs,
	EquipmentSlotTypeHands,
	EquipmentSlotTypeFeet,
	EquipmentSlotTypeRing,
	EquipmentSlotTypeAmmo,
}

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

// EquipmentSlot is an item equipped in an entity's equipment.
type EquipmentSlot struct {
	Type   EquipmentSlotType
	Item   *Item
	Amount int
}

// EntityAppearance describes the properties of an entity such as a player or NPC.
type EntityAppearance struct {
	NPCAppearance  int
	Equipment      map[EquipmentSlotType]*EquipmentSlot
	Body           []int
	Animations     map[AnimationID]int
	Gender         EntityGender
	OverheadIconID int
	CombatLevel    int
	TotalLevel     int
	Updated        bool
}

// IsNPCAppearance returns true if the appearance should take that of a predefined NPC, false if not.
func (a *EntityAppearance) IsNPCAppearance() bool {
	return a.NPCAppearance > 0
}

// SetNPCAppearance sets the ID of an NPC to use for the appearance.
func (a *EntityAppearance) SetNPCAppearance(id int) {
	a.NPCAppearance = id
}
