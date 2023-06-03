package response

import "github.com/mbpolan/openmcs/internal/network"

const PlayerWeightResponseHeader byte = 0xF0

// PlayerWeightResponse is sent by the server to report a player's current weight.
type PlayerWeightResponse struct {
	Weight int
}

// Write writes the contents of the message to a stream.
func (p *PlayerWeightResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(PlayerWeightResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the weight
	err = w.WriteUint16(uint16(p.Weight))
	if err != nil {
		return err
	}

	return nil
}
