package model

// NPC is a non-player controlled character in the game world.
type NPC struct {
	*Entity
	// DefinitionID is the identifier for the NPC's appearance.
	DefinitionID int
}

// NewNPC returns a new NPC model.
func NewNPC() *NPC {
	return &NPC{
		Entity: &Entity{},
	}
}
