package model

// ItemNature enumerates the possible uses for an item
type ItemNature int

const (
	ItemNatureNotUsable ItemNature = iota
	ItemNatureEquippable
)

// WeaponStyle enumerates the possible attack styles of a weapon.
type WeaponStyle int

const (
	WeaponStyleNone WeaponStyle = iota
	WeaponStyle2HSword
	WeaponStyleAxe
	WeaponStyleBow
	WeaponStyleBlunt
	WeaponStyleClaw
	WeaponStyleCrossbow
	WeaponStyleGun
	WeaponStylePickaxe
	WeaponStylePoleArm
	WeaponStylePoleStaff
	WeaponStyleScythe
	WeaponStyleSlashSword
	WeaponStyleSpear
	WeaponStyleSpiked
	WeaponStyleStabSword
	WeaponStyleStaff
	WeaponStyleThrown
	WeaponStyleWhip
)

// ItemStackable is a descriptor of sprites to use for certain item stackable thresholds.
type ItemStackable struct {
	ID     int
	Amount int
}

// ItemCombatAttributes are the attack and defense bonuses granted by an item.
type ItemCombatAttributes struct {
	Stab  int
	Slash int
	Crush int
	Magic int
	Range int
}

// ItemAttributes are additional properties for an item.
type ItemAttributes struct {
	// ItemID is the ID of the item.
	ItemID int
	// Nature describes the uses for the item.
	Nature ItemNature
	// EquipSlotType is the equipment slot where the item is equipped to.
	EquipSlotType EquipmentSlotType
	// WeaponStyle is the set of attack options for a weapon item.
	WeaponStyle WeaponStyle
	// Speed is the amount of milliseconds between item actions.
	Speed int
	// Weight is the weight of the item.
	Weight float64
	// AttackBonuses are the offensive combat attributes.
	Attack ItemCombatAttributes
	// DefenseBonuses are the defensive combat attributes.
	Defense ItemCombatAttributes
	// StrengthBonus is the bonus granted to an entity's strength level.
	StrengthBonus int
	// PrayerBonus is the bonus granted to an entity's prayer level.
	PrayerBonus int
}

// Item represents a player-usable object.
type Item struct {
	ID             int
	Name           string
	Description    string
	Rotation       Vector3D
	Scale          Vector3D
	Stackable      bool
	MembersOnly    bool
	NoteID         int
	NoteTemplateID int
	TeamID         int
	GroundActions  []string
	Actions        []string
	Stackables     []ItemStackable
	Attributes     *ItemAttributes
}

// CanEquip returns true if the item can be equipped, or false if not.
func (i *Item) CanEquip() bool {
	if i.Attributes == nil {
		return false
	}

	return i.Attributes.Nature&ItemNatureEquippable != 0
}
