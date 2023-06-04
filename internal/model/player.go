package model

import (
	"math"
	"strings"
)

// MaxInventorySlots is the maximum number of inventory slots.
const MaxInventorySlots = 28

// MaxStackableSize is the maximum size of a single stack of an item.
var MaxStackableSize = int64(math.Pow(2, 31) - 1)

// PlayerType enumerates the possible player access levels.
type PlayerType int

const (
	PlayerNormal PlayerType = iota
	PlayerModerator
	PlayerAdmin
)

// InventorySlot is an item stored in a player's inventory.
type InventorySlot struct {
	// ID is the identifier of the inventory slot.
	ID int
	// Item is a pointer to a model.Item in this slot.
	Item *Item
	// Amount is the stack size of the item in this slot.
	Amount int
}

// Player is a human player connected to the game server. This struct stores a player's persistent data, including
// various preferences, game world properties and other such attributes.
type Player struct {
	// ID is the player's globally unique identifier.
	ID int
	// Username is the player's display name.
	Username string
	// PasswordHash is the player's hashed password.
	PasswordHash string
	// Type determines the level of access this player is granted.
	Type PlayerType
	// Flagged is true when the player is suspected of cheating, false if not.
	Flagged bool
	// GlobalPos is the player's position on the world map, in global coordinates.
	GlobalPos Vector3D
	// Appearance is the player's model appearance.
	Appearance EntityAppearance
	// AutoRetaliate controls if the player automatically responds to combat.
	AutoRetaliate bool
	// Modes determine what level of chat and trade interaction the player has configured.
	Modes PlayerModes
	// Muted is true when the player is not able to chat, false if not.
	Muted bool
	// Friends is a slice of usernames of players added as friends.
	Friends []string
	// Ignored is a slice of usernames of players that are ignored.
	Ignored []string
	// Skills is a map of this player's skill levels and experience.
	Skills SkillMap
	// Member is true when the player has an active subscription, false if not.
	Member bool
	// MemberDays is the number of days of subscription remaining.
	MemberDays int
	// Inventory is the player's current inventory of items.
	Inventory [MaxInventorySlots]*InventorySlot
	// CombatStats are the player's combat statistics.
	CombatStats EntityCombatStats
	// AttackStyles if a map of the player's preferred attack styles for a given weapon style.
	AttackStyles map[WeaponStyle]AttackStyle
	// GameOptions is a map of client/game option IDs to their values.
	GameOptions map[int]string
	// RunEnergy is the player's run energy.
	RunEnergy float64
	// QuestStatus is a map of quest IDs to their status.
	QuestStatuses map[int]QuestStatus
	// QuestFlags is a map of quest IDs to maps of their flag IDs to values.
	QuestFlags map[int]map[int]int
	// UpdateDesign is true when the player should be shown the character design interface, false if not.
	UpdateDesign bool
}

// PlayerModes indicates what types of chat and interactions a player wishes to receive.
type PlayerModes struct {
	// PublicChat is the player's public chat interaction setting.
	PublicChat ChatMode
	// PrivateChat is the player's public chat interaction setting.
	PrivateChat ChatMode
	// Interaction is the player's trade/duel request interaction setting.
	Interaction InteractionMode
}

// defaultAnimations returns a map of default player animations.
// TODO: these should not be hardcoded since the base animations can changed depending on the player's equipment or
// buff/defbuff state
func defaultAnimations() map[AnimationID]int {
	return map[AnimationID]int{
		AnimationStand:     0x080D, // standing
		AnimationStandTurn: 0xFFFF, // turning
		AnimationWalk:      0x067C, // walk
		AnimationTurnAbout: 0xFFFF, // turn about
		AnimationTurnRight: 0xFFFF, // turn right
		AnimationTurnLeft:  0xFFFF, // turn left
		AnimationRun:       0x067D, // run
	}
}

// NewPlayer returns a new player model.
func NewPlayer(username string) *Player {
	// define a default appearance
	appearance := EntityAppearance{
		Animations: defaultAnimations(),
		Equipment:  map[EquipmentSlotType]*EquipmentSlot{},
		BodyColors: make([]int, NumBodyColors),
		GraphicID:  -1,
	}

	return &Player{
		Username:      username,
		Appearance:    appearance,
		AttackStyles:  InitAttackStyleMap(),
		Skills:        EmptySkillMap(),
		GameOptions:   map[int]string{},
		RunEnergy:     100.0,
		QuestStatuses: map[int]QuestStatus{},
		QuestFlags:    map[int]map[int]int{},
	}
}

// AttackStyle returns the active attack style for a particular weapon style.
func (p *Player) AttackStyle(weaponStyle WeaponStyle) AttackStyle {
	return p.AttackStyles[weaponStyle]
}

// SetAttackStyle sets the active attack style for a weapon style, overwriting any previously applied style.
func (p *Player) SetAttackStyle(weaponStyle WeaponStyle, attackStyle AttackStyle) {
	p.AttackStyles[weaponStyle] = attackStyle
}

// Weight returns the total weight of all player inventory and equipment.
func (p *Player) Weight() float64 {
	weight := 0.0

	// factor in any inventory items that have a weight value
	for _, slot := range p.Inventory {
		if slot == nil || slot.Item.Attributes == nil {
			continue
		}

		weight += slot.Item.Attributes.Weight * float64(slot.Amount)
	}

	// and equipped items with weights
	for _, slot := range p.Appearance.Equipment {
		if slot == nil || slot.Item.Attributes == nil {
			continue
		}

		weight += slot.Item.Attributes.Weight * float64(slot.Amount)
	}

	return weight
}

// SetSkillExperience sets the experience points for a player skill. The skill level and combat levels will be
// recomputed after the fact.
func (p *Player) SetSkillExperience(skillType SkillType, experience float64) {
	p.Skills[skillType].Experience = experience
	p.Skills[skillType].Level = p.recomputeSkillLevel(experience)
	p.recomputeCombatSkills()
}

// SkillExperience returns the player's current experience points for a skill.
func (p *Player) SkillExperience(skillType SkillType) float64 {
	return p.Skills[skillType].Experience
}

// EquippedWeaponStyle returns the attack style of the player's equipped weapon, if any.
func (p *Player) EquippedWeaponStyle() WeaponStyle {
	slot, ok := p.Appearance.Equipment[EquipmentSlotTypeWeapon]
	if !ok || slot.Item.Attributes == nil {
		return WeaponStyleUnarmed
	}

	return slot.Item.Attributes.WeaponStyle
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

// GameOption returns the player's preference value for a game option. If no value is set for the option, an empty
// string is returned instead.
func (p *Player) GameOption(optionID int) string {
	value, ok := p.GameOptions[optionID]
	if ok {
		return value
	}

	return ""
}

// SetGameOption sets a value for a game option, overwriting any previous value.
func (p *Player) SetGameOption(optionID int, optionValue string) {
	p.GameOptions[optionID] = optionValue
}

// QuestStatus returns the status of a quest.
func (p *Player) QuestStatus(questID int) QuestStatus {
	return p.QuestStatuses[questID]
}

// SetQuestStatus sets the status of a quest, overwriting any previous status.
func (p *Player) SetQuestStatus(questID int, status QuestStatus) {
	p.QuestStatuses[questID] = status
}

// QuestFlag returns the value of a quest flag. If a flag does not have a value, -1 will be returned.
func (p *Player) QuestFlag(questID, flagID int) int {
	flags, ok := p.QuestFlags[questID]
	if !ok {
		return -1
	}

	flag, ok := flags[flagID]
	if !ok {
		return -1
	}

	return flag
}

// SetQuestFlag sets the value of a flag for a quest, overwriting any previously set value.
func (p *Player) SetQuestFlag(questID, flagID, value int) {
	_, ok := p.QuestFlags[questID]
	if !ok {
		p.QuestFlags[questID] = map[int]int{}
	}

	p.QuestFlags[questID][flagID] = value
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

// recomputeSkillLevel returns the level for a skill based on the total amount of experience points.
func (p *Player) recomputeSkillLevel(experience float64) int {
	for i := 1; i <= 99; i++ {
		if SkillExperienceLevels[i] > experience {
			return i - 1
		}
	}

	return 99
}

// recomputeCombatSkills updates the derived levels (total and combat) based on the player's current skills.
func (p *Player) recomputeCombatSkills() {
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
