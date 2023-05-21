package response

import "github.com/mbpolan/openmcs/internal/network"

const SetInterfaceModelResponseHeader byte = 0xF6

// SetInterfaceModelResponse is sent by the server when an item model should be dispayed in an interface.
type SetInterfaceModelResponse struct {
	InterfaceID int
	ItemID      int
	Zoom        int
}

// Write writes the contents of the message to a stream.
func (p *SetInterfaceModelResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SetInterfaceModelResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the interface id
	err = w.WriteUint16LE(uint16(p.InterfaceID))
	if err != nil {
		return err
	}

	// write 2 bytes for the zoom factor
	err = w.WriteUint16(uint16(p.Zoom))
	if err != nil {
		return err
	}

	// write 2 bytes for the item id
	err = w.WriteUint16(uint16(p.ItemID))
	if err != nil {
		return err
	}

	return nil
}
