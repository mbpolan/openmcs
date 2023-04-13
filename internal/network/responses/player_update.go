package responses

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const PlayerUpdateResponseHeader byte = 0x51

const (
	localMoveNoUpdate  byte = 0xFF
	localMoveUnchanged      = 0x00
	localMoveWalk           = 0x01
	localMoveRun            = 0x02
	localMovePosition       = 0x03
)

// PlayerUpdateResponse contains a game state update.
type PlayerUpdateResponse struct {
	localMoveType       byte
	localPosition       model.Vector3D
	localClearWaypoints bool
	localNeedsUpdate    bool
}

// NewPlayerUpdateResponse creates a new game state update response.
func NewPlayerUpdateResponse() *PlayerUpdateResponse {
	return &PlayerUpdateResponse{
		localMoveType: localMoveNoUpdate,
	}
}

// SetLocalPlayerNoMovement reports that the local player's state has not changed.
func (p *PlayerUpdateResponse) SetLocalPlayerNoMovement() {
	p.localMoveType = localMoveUnchanged
}

// SetLocalPlayerPosition reports the local player's position in region local coordinates and update status. The
// clearWaypoints flag indicates if the player's current path should be cancelled, such as in the case of the player
// being teleported to a location.
func (p *PlayerUpdateResponse) SetLocalPlayerPosition(pos model.Vector3D, clearWaypoints bool, needsUpdate bool) {
	p.localMoveType = localMovePosition
	p.localPosition = pos
	p.localClearWaypoints = clearWaypoints
	p.localNeedsUpdate = needsUpdate
}

// Write writes the contents of the message to a stream.
func (p *PlayerUpdateResponse) Write(w *network.ProtocolWriter) error {
	// local player movement
	bs := network.NewBitSet()

	// write local player movement details
	p.writeLocalPlayer(bs)

	// 8 bits for the number of other players to update
	bs.SetBits(0, 8)

	// TODO: updates for player list

	// add local player as the last one in the update list
	if p.localMoveType != localMoveNoUpdate {
		bs.SetBits(0x7FF, 11)
	}

	// write packet header
	err := w.WriteByte(PlayerUpdateResponseHeader)
	if err != nil {
		return err
	}

	// write packet size
	err = w.WriteUint16(uint16(bs.Size() + 1))
	if err != nil {
		return err
	}

	// write bits section
	err = bs.Write(w)
	if err != nil {
		return err
	}

	// TODO: individual player updates
	err = w.WriteByte(0x00)
	if err != nil {
		return err
	}

	return w.Flush()
}

func (p *PlayerUpdateResponse) writeLocalPlayer(bs *network.BitSet) {
	// first bit is a flag if there is an update for the local player
	if p.localMoveType == localMoveNoUpdate {
		// clear the first bit and bail out since there is nothing else to do
		bs.Clear()
		return
	}

	// set the first bit to indicate we have an update
	bs.Set()

	// two bits represent the local player update type
	bs.SetBits(uint32(p.localMoveType), 2)

	switch p.localMoveType {
	case localMoveUnchanged:
		// nothing to do

	case localMoveWalk:
		// TODO
		panic("not implemented")

	case localMoveRun:
		// TODO
		panic("not implemented")

	case localMovePosition:
		// write 2 bits for the z coordinate
		bs.SetBits(uint32(p.localPosition.Z), 2)

		// write 1 bit each for the clear waypoints and update needed flags
		bs.SetOrClear(p.localClearWaypoints)
		bs.SetOrClear(p.localNeedsUpdate)

		// write 7 bits each for the x and y coordinates
		bs.SetBits(uint32(p.localPosition.X), 7)
		bs.SetBits(uint32(p.localPosition.Y), 7)
	}
}
