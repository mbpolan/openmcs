package responses

import "github.com/mbpolan/openmcs/internal/network"

const PlayerUpdateResponseHeader byte = 0x51

// PlayerUpdateResponse contains a game state update.
type PlayerUpdateResponse struct {
}

// NewPlayerUpdateResponse creates a new game state update response.
func NewPlayerUpdateResponse() *PlayerUpdateResponse {
	return &PlayerUpdateResponse{}
}

// Write writes the contents of the message to a stream.
func (p *PlayerUpdateResponse) Write(w *network.ProtocolWriter) error {
	// local player movement
	bs := network.NewBitSet()

	// first bit is a flag if there is an update for the local player
	bs.Set()

	// two bits represent the local player update type
	bs.SetBits(0, 2)

	// 8 bits for the number of other players to update
	bs.SetBits(0, 8)

	// TODO: updates for player list

	// add local player
	bs.SetBits(0x7FF, 11)

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
