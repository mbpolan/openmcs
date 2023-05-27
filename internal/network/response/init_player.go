package response

import "github.com/mbpolan/openmcs/internal/network"

const InitPlayerResponseHeader byte = 0xF9

// InitPlayerResponse is sent by the server to inform the player's client of their session status.
type InitPlayerResponse struct {
	Member      bool
	ServerIndex int
}

// Write writes the contents of the message to a stream.
func (p *InitPlayerResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(InitPlayerResponseHeader)
	if err != nil {
		return err
	}

	member := 0x00
	if p.Member {
		member = 0x01
	}

	// write 1 byte for the player's membership status
	err = w.WriteUint8(uint8(member + 0x80))
	if err != nil {
		return err
	}

	// write 2 bytes for the player's server index
	err = w.WriteUint16Alt2(uint16(p.ServerIndex * -1))
	if err != nil {
		return err
	}

	return nil
}
