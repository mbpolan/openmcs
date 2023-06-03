package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const SetInterfaceColorResponseHeader byte = 0x7A

// SetInterfaceColorResponse is sent by the server to change the color of an interface.
type SetInterfaceColorResponse struct {
	InterfaceID int
	Color       model.Color
}

// Write writes the contents of the message to a stream.
func (p *SetInterfaceColorResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SetInterfaceColorResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the interface id
	err = w.WriteUint16LEAlt(uint16(p.InterfaceID))
	if err != nil {
		return err
	}

	// encode the color value and write 2 bytes
	color := (p.Color.Red&0x1F)<<10 | (p.Color.Green&0x1F)<<5 | (p.Color.Blue & 0x1F)
	err = w.WriteUint16LEAlt(uint16(color))
	if err != nil {
		return err
	}

	return nil
}
