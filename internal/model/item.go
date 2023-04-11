package model

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
}
