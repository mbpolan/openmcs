package response

import "github.com/mbpolan/openmcs/internal/network"

const ClearInventoryResponseHeader byte = 0x48

// ClearInventoryResponse is sent by the server when a player's inventory should be reset.
type ClearInventoryResponse struct {
	interfaceID int
}

// NewClearInventoryResponse creates a new response to clear a player's inventory for an interface.
func NewClearInventoryResponse(interfaceID int) *ClearInventoryResponse {
	return &ClearInventoryResponse{
		interfaceID: interfaceID,
	}
}

// Write writes the contents of the message to a stream.
func (p *ClearInventoryResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(ClearInventoryResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the interface id
	err = w.WriteUint16LE(uint16(p.interfaceID))
	if err != nil {
		return err
	}

	return nil
}
