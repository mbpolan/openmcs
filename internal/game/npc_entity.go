package game

import "github.com/mbpolan/openmcs/internal/model"

// npcEntity is an instance of an NPC in the game world.
type npcEntity struct {
	npc *model.NPC
}

// newNPCEntity returns a new npcEntity instance.
func newNPCEntity(npc *model.NPC) *npcEntity {
	return &npcEntity{npc: npc}
}
