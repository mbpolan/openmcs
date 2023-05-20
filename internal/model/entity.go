package model

// EquipmentSlotType enumerates the different slots items may be equipped to.
type EquipmentSlotType int

const (
	EquipmentSlotTypeHead     EquipmentSlotType = 0
	EquipmentSlotTypeCape                       = 1
	EquipmentSlotTypeNecklace                   = 2
	EquipmentSlotTypeWeapon                     = 3
	EquipmentSlotTypeBody                       = 4
	EquipmentSlotTypeShield                     = 5
	EquipmentSlotTypeLegs                       = 7
	EquipmentSlotTypeHands                      = 9
	EquipmentSlotTypeFeet                       = 10
	EquipmentSlotTypeRing                       = 12
	EquipmentSlotTypeAmmo                       = 13
)

// Visible returns true if the equipment slot affects an entity's appearance, false if not.
func (s EquipmentSlotType) Visible() bool {
	return s != EquipmentSlotTypeRing && s != EquipmentSlotTypeAmmo
}

// EquipmentSlotTypes is a slice of all EquipmentSlotType enums sorted according to their slot IDs in ascending order.
var EquipmentSlotTypes = []EquipmentSlotType{
	EquipmentSlotTypeHead,
	EquipmentSlotTypeCape,
	EquipmentSlotTypeNecklace,
	EquipmentSlotTypeWeapon,
	EquipmentSlotTypeBody,
	EquipmentSlotTypeShield,
	EquipmentSlotTypeLegs,
	EquipmentSlotTypeHands,
	EquipmentSlotTypeFeet,
	EquipmentSlotTypeRing,
	EquipmentSlotTypeAmmo,
}

// NumBodyColors is the number of customizable character body parts.
const NumBodyColors = 5

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
	SlotType EquipmentSlotType
	Item     *Item
	Amount   int
}

// EntityBase describes the entity's base model. Each entity model has a fixed amount of slots that can be assigned
// model IDs, which comprise the entity if it has no items equipped.
type EntityBase struct {
	Head  int
	Face  int
	Body  int
	Arms  int
	Hands int
	Legs  int
	Feet  int
}

// EntityCombatAttributes are the combat modifiers for an entity based on their equipment.
type EntityCombatAttributes struct {
	Stab  int
	Slash int
	Crush int
	Magic int
	Range int
}

// EntityCombatStats are the effective combat stats for an entity.
type EntityCombatStats struct {
	Attack   EntityCombatAttributes
	Defense  EntityCombatAttributes
	Strength int
	Prayer   int
}

// EntityAppearance describes the properties of an entity such as a player or NPC.
type EntityAppearance struct {
	Base           EntityBase
	NPCAppearance  int
	Equipment      map[EquipmentSlotType]*EquipmentSlot
	BodyColors     []int
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
