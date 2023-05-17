package response

import "github.com/mbpolan/openmcs/internal/network"

const ClearScreenResponseHeader byte = 0xDB

// ClearScreenResponse is sent by the server when the player's client should hide all open interfaces.
type ClearScreenResponse struct {
}

// Write writes the contents of the message to a stream.
func (p *ClearScreenResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(ClearScreenResponseHeader)
	if err != nil {
		return err
	}

	return nil
}
