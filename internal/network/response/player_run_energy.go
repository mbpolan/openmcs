package response

import "github.com/mbpolan/openmcs/internal/network"

const PlayerRunEnergyResponseHeader byte = 0x6E

// PlayerRunEnergyResponse is sent by the server to report a player's current run energy.
type PlayerRunEnergyResponse struct {
	RunEnergy int
}

// Write writes the contents of the message to a stream.
func (p *PlayerRunEnergyResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(PlayerRunEnergyResponseHeader)
	if err != nil {
		return err
	}

	// write 1 byte for the run energy
	err = w.WriteUint8(uint8(p.RunEnergy))
	if err != nil {
		return err
	}

	return nil
}
