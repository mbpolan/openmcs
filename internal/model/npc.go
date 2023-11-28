package model

// NPC is a non-player controlled character in the game world.
type NPC struct {
	*Entity
	// DefinitionID is the identifier for the NPC's appearance.
	DefinitionID int
	// ScriptSlug is the slug for the game script to execute for this NPC.
	ScriptSlug string
}

// NewNPC returns a new NPC model.
func NewNPC(definitionID int, scriptSlug string) *NPC {
	return &NPC{
		Entity:       &Entity{},
		DefinitionID: definitionID,
		ScriptSlug:   scriptSlug,
	}
}
