package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
)

const SidebarTabResponseHeader byte = 0x6A

// SidebarTabResponse is sent by the server to set the active sidebar tab on the player's client.
type SidebarTabResponse struct {
	TabID model.ClientTab
}

// Write writes the contents of the message to a stream.
func (p *SidebarTabResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SidebarTabResponseHeader)
	if err != nil {
		return err
	}

	// write 1 byte for the client tab
	err = w.WriteUint8(uint8(int(common.ClientTabIndices[p.TabID]) * -1))
	if err != nil {
		return err
	}

	return nil
}
