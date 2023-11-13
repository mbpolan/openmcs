package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"sort"
)

const NPCUpdateResponseHeader byte = 0x41

const (
	npcMoveNoUpdate  byte = 0xFF
	npcMoveUnchanged      = 0x00
	npcMoveWalk           = 0x01
	npcMoveRun            = 0x02
)

// NPCUpdateResponse instructs the client to update the visible NPCs on the game world.
type NPCUpdateResponse struct {
	list map[int]*trackedNPC
}

// trackedNPC is an NPC entity that is being tracked in the update payload.
type trackedNPC struct {
	definitionID   int
	observed       bool
	clearWaypoints bool
	pos            model.Vector2D
	update         *npcUpdate
	movement       *npcMovement
}

// npcMovement is movement information associated with an NPC.
type npcMovement struct {
	moveType       byte
	position       model.Vector3D
	clearWaypoints bool
	walkDirection  model.Direction
	runDirections  [2]model.Direction
}

// npcUpdate contains various attributes to inform players about an NPC.
type npcUpdate struct {
	mask uint16
}

// NewNPCUpdateResponse returns a new response for updating NPC statuses.
func NewNPCUpdateResponse() *NPCUpdateResponse {
	return &NPCUpdateResponse{
		list: map[int]*trackedNPC{},
	}
}

// AddNPCNoMovement tracks an NPC that has not moved.
func (p *NPCUpdateResponse) AddNPCNoMovement(npcID int) {
	p.list[npcID] = &trackedNPC{
		movement: &npcMovement{
			moveType: npcMoveUnchanged,
		},
		update: &npcUpdate{},
	}
}

// AddNPCToList adds an NPC to the list of newly encountered NPCs.
func (p *NPCUpdateResponse) AddNPCToList(npcID, definitionID int, posOffset model.Vector2D, observed, clearWaypoints bool) {
	p.list[npcID] = &trackedNPC{
		definitionID:   definitionID,
		observed:       observed,
		clearWaypoints: clearWaypoints,
		pos:            posOffset,
		// this requires an update to be included
		update: &npcUpdate{},
	}
}

// Write writes the contents of the message to a stream.
func (p *NPCUpdateResponse) Write(w *network.ProtocolWriter) error {
	// use a buffered writer since we need to track the packet size
	bw := network.NewBufferedWriter()
	err := p.writePayload(bw)
	if err != nil {
		return err
	}

	// write packet header
	err = w.WriteUint8(NPCUpdateResponseHeader)
	if err != nil {
		return err
	}

	// determine the length of the payload buffer and write it as two bytes
	buf, err := bw.Buffer()
	err = w.WriteUint16(uint16(buf.Len()))
	if err != nil {
		return err
	}

	// write the payload itself
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}

// writePayload writes the payload of the response to the given writer.
func (p *NPCUpdateResponse) writePayload(w *network.ProtocolWriter) error {
	bs := network.NewBitSet()

	// write npc movement details
	p.writeMovement(bs)

	var npcIDs []int
	for npcID, _ := range p.list {
		npcIDs = append(npcIDs, npcID)
	}

	// sort the npcs by the ids in descending order
	sort.Slice(npcIDs, func(i, j int) bool {
		return npcIDs[i] < npcIDs[j]
	})

	// write the npc list
	p.writeList(npcIDs, bs)

	// write bits section first representing local, other and npc list updates
	err := bs.Write(w)
	if err != nil {
		return err
	}

	// write each npc update block
	err = p.writeUpdates(npcIDs, w)
	if err != nil {
		return err
	}

	return nil
}

// writeMovement writes movement data to a bitset.
func (p *NPCUpdateResponse) writeMovement(bs *network.BitSet) {
	movements := 0
	for _, other := range p.list {
		if other.movement == nil {
			continue
		}

		movements++
	}

	// write 8 bits indicating how many npcs have movement updates
	bs.SetBits(uint32(movements), 8)

	for _, other := range p.list {
		if other.movement == nil {
			continue
		}

		// set or clear 1 bit to flag if an update is required for this npc
		moveType := other.movement.moveType
		if moveType == npcMoveNoUpdate {
			// if the npc does have an update pending, we need to send an unchanged movement type instead
			if other.update == nil {
				bs.Clear()
				continue
			}

			moveType = npcMoveUnchanged
		}

		bs.Set()

		// write 2 bits for the movement type
		bs.SetBits(uint32(moveType), 2)

		switch moveType {
		case npcMoveUnchanged:
			// nothing to do

		case npcMoveWalk:
			// TODO: write 3 bits for the direction

			// TODO: write 1 bit of a further update is required

		case npcMoveRun:
			// TODO: write 3 bits for the first direction

			// TODO: write 3 bits for the second direction

			// TODO: write 1 bit if a further update is required
		}
	}
}

// writeList writes NPC list data to a bitset.
func (p *NPCUpdateResponse) writeList(npcIDS []int, bs *network.BitSet) {
	for _, npcID := range npcIDS {
		npc := p.list[npcID]

		// don't include npcs that already had movements reported
		if npc.movement != nil {
			continue
		}

		// write 14 bits for the npc id
		bs.SetBits(uint32(npcID), 14)

		// write 5 bits for the y and x coordinate offsets
		bs.SetBits(uint32(npc.pos.Y), 5)
		bs.SetBits(uint32(npc.pos.X), 5)

		// write 1 bit if the npc should have their path waypoints cleared
		bs.SetOrClear(npc.clearWaypoints)

		// write 12 bits for the npc appearance definition id
		bs.SetBits(uint32(npc.definitionID), 12)

		// write 1 bit if a further update is required
		needsUpdate := npc.update != nil
		bs.SetOrClear(needsUpdate)
	}

	// mark the end of the npc list
	bs.SetBits(0x3FFF, 14)
}

// writeUpdates writes NPC update data to the given writer.
func (p *NPCUpdateResponse) writeUpdates(npcIDs []int, w *network.ProtocolWriter) error {
	for _, npcID := range npcIDs {
		npc := p.list[npcID]

		// skip npcs without any updates
		if npc.update == nil {
			continue
		}

		// TODO: npc update mask
		err := w.WriteUint8(uint8(0x00))
		if err != nil {
			return err
		}
	}

	return nil
}

// ensureNPC returns or creates an entry in the update list for an NPC with an ID.
func (p *NPCUpdateResponse) ensureNPC(npcID int) *trackedNPC {
	n, ok := p.list[npcID]
	if !ok {
		n = &trackedNPC{}
		p.list[npcID] = n
	}

	return n
}
