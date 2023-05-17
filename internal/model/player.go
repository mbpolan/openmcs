package model

import (
	"math"
	"strings"
)

type PlayerType int

// NumPlayerSkills is the number of available player skills.
const NumPlayerSkills = 21

// MaxInventorySlots is the maximum number of inventory slots.
const MaxInventorySlots = 28

// MaxStackableSize is the maximum size of a single stack of an item.
var MaxStackableSize = int64(math.Pow(2, 31) - 1)

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// InventorySlot is an item stored in a player's inventory.
type InventorySlot struct {
	ID     int
	Item   *Item
	Amount int
}

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
	Inventory    [MaxInventorySlots]*InventorySlot
	CombatStats  EntityCombatStats
	UpdateDesign bool
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
		Equipment:  map[EquipmentSlotType]*EquipmentSlot{},
		BodyColors: make([]int, NumBodyColors),
		Updated:    false,
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

// SetEquippedItem sets an item to be equipped at a slot.
func (p *Player) SetEquippedItem(item *Item, amount int, slot EquipmentSlotType) {
	p.Appearance.Equipment[slot] = &EquipmentSlot{
		SlotType: slot,
		Item:     item,
		Amount:   amount,
	}
	p.recomputeCombatStats()
}

// ClearEquippedItem removes any equipped item at a slot.
func (p *Player) ClearEquippedItem(slot EquipmentSlotType) {
	delete(p.Appearance.Equipment, slot)
	p.recomputeCombatStats()
}

// EquipmentSlot returns an EquipmentSlot for a slot. If no item is equipped, nil will be returned instead.
func (p *Player) EquipmentSlot(slot EquipmentSlotType) *EquipmentSlot {
	return p.Appearance.Equipment[slot]
}

// IsEquipmentSlotUsed returns true if an equipment slot is in use, false if it's free.
func (p *Player) IsEquipmentSlotUsed(slot EquipmentSlotType) bool {
	_, ok := p.Appearance.Equipment[slot]
	return ok
}

// SetInventoryItem puts an item in a slot of the player's inventory. This will replace any existing items at that slot.
func (p *Player) SetInventoryItem(item *Item, amount, slot int) {
	p.Inventory[slot] = &InventorySlot{
		ID:     slot,
		Item:   item,
		Amount: amount,
	}
}

// ClearInventoryItem removes an item from the player's inventory slot.
func (p *Player) ClearInventoryItem(slot int) {
	p.Inventory[slot] = nil
}

// InventorySlotWithItem returns the slot that contains an item with an ID. If no slot contains such an item, then
// nil will be returned.
func (p *Player) InventorySlotWithItem(itemID int) *InventorySlot {
	for _, slot := range p.Inventory {
		if slot != nil && slot.Item.ID == itemID {
			return slot
		}
	}

	return nil
}

// InventoryCanHoldItem determines if an item can be added to the player's inventory. This will account for both
// stackable and non-stackable items.
func (p *Player) InventoryCanHoldItem(item *Item) bool {
	if item.Stackable {
		// see if a slot is already holding this item and its stack size can accommodate another instance
		slot := p.InventorySlotWithItem(item.ID)
		if slot != nil && int64(slot.Amount+1) <= MaxStackableSize {
			return true
		}
	}

	// if the item is not stackable, or is stackable but there is no existing slot that can hold it, we need to
	// use another free slot
	return p.NextFreeInventorySlot() != -1
}

// NextFreeInventorySlot returns the ID of the next available slot in the player's inventory. If no slot is free, -1
// will be returned.
func (p *Player) NextFreeInventorySlot() int {
	for i, item := range p.Inventory {
		if item == nil {
			return i
		}
	}

	return -1
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

// recomputeCombatStats updates the player's combat statistics based on their equipment.
func (p *Player) recomputeCombatStats() {
	stats := EntityCombatStats{}

	for _, slot := range p.Appearance.Equipment {
		if slot.Item.Attributes == nil {
			continue
		}

		stats.Attack.Stab += slot.Item.Attributes.Attack.Stab
		stats.Attack.Slash += slot.Item.Attributes.Attack.Slash
		stats.Attack.Crush += slot.Item.Attributes.Attack.Crush
		stats.Attack.Magic += slot.Item.Attributes.Attack.Magic
		stats.Attack.Range += slot.Item.Attributes.Attack.Range

		stats.Defense.Stab += slot.Item.Attributes.Defense.Stab
		stats.Defense.Slash += slot.Item.Attributes.Defense.Slash
		stats.Defense.Crush += slot.Item.Attributes.Defense.Crush
		stats.Defense.Magic += slot.Item.Attributes.Defense.Magic
		stats.Defense.Range += slot.Item.Attributes.Defense.Range

		stats.Strength += slot.Item.Attributes.StrengthBonus
		stats.Prayer += slot.Item.Attributes.PrayerBonus
	}

	p.CombatStats = stats
}
