package model

// ItemNature enumerates the possible uses for an item
type ItemNature int

const (
	ItemNatureNotUsable          ItemNature = 0
	ItemNatureEquipmentOneHanded ItemNature = 1 << iota
	ItemNatureEquipmentTwoHanded
)

// ItemStackable is a descriptor of sprites to use for certain item stackable thresholds.
type ItemStackable struct {
	ID     int
	Amount int
}

// ItemAttributes are additional properties for an item.
type ItemAttributes struct {
	// ItemID is the ID of the item.
	ItemID int
	// Nature describes the uses for the item.
	Nature ItemNature
	// EquipSlotType is the equipment slot where the item is equipped to.
	EquipSlotType EquipmentSlotType
	// Speed is the amount of milliseconds between item actions.
	Speed int
	// Weight is the weight of the item.
	Weight float64
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

	return i.Attributes.Nature&ItemNatureEquipmentOneHanded != 0 || i.Attributes.Nature&ItemNatureEquipmentTwoHanded != 0
}
