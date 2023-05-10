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

// ItemAttributes are additional properties for an item.
type ItemAttributes struct {
	// ItemID is the ID of the item.
	ItemID int
	// Nature describes the uses for the item.
	Nature ItemNature
	// EquipSlotID is the equipment slot where the item is equipped to.
	EquipSlotID int
	// Speed is the amount of milliseconds between item actions.
	Speed int
	// Weight is the weight of the item.
	Weight float64
}
