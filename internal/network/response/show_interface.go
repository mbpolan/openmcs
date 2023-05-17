package response

import "github.com/mbpolan/openmcs/internal/network"

const ShowInterfaceResponseHeader byte = 0x61

// ShowInterfaceResponse is sent by the server when a player's client should open an interface.
type ShowInterfaceResponse struct {
	interfaceID int
}

// NewShowInterfaceResponse creates a new response to open an interface.
func NewShowInterfaceResponse(interfaceID int) *ShowInterfaceResponse {
	return &ShowInterfaceResponse{
		interfaceID: interfaceID,
	}
}

// Write writes the contents of the message to a stream.
func (p *ShowInterfaceResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(ShowInterfaceResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the interface id
	err = w.WriteUint16(uint16(p.interfaceID))
	if err != nil {
		return err
	}

	return nil
}
