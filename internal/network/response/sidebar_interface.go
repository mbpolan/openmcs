package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
	"github.com/mbpolan/openmcs/internal/network/common"
)

// noSidebarInterfaceID is an identifier used to remove an interface from a client tab.
const noSidebarInterfaceID int = 0xFFFF

const SidebarInterfaceResponseHeader byte = 0x47

// SidebarInterfaceResponse is sent by the server to tell the client to show a particular interface in its sidebar.
type SidebarInterfaceResponse struct {
	Tab       model.ClientTab
	SidebarID int
}

// NewSidebarInterfaceResponse creates a response to set an interface on a client tab.
func NewSidebarInterfaceResponse(tab model.ClientTab, sidebarID int) *SidebarInterfaceResponse {
	return &SidebarInterfaceResponse{
		SidebarID: sidebarID,
		Tab:       tab,
	}
}

// NewRemoveSidebarInterfaceResponse creates a response to remove an interface on a client tab.
func NewRemoveSidebarInterfaceResponse(tab model.ClientTab) *SidebarInterfaceResponse {
	return &SidebarInterfaceResponse{
		SidebarID: noSidebarInterfaceID,
		Tab:       tab,
	}
}

// Write writes the contents of the message to a stream.
func (p *SidebarInterfaceResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(SidebarInterfaceResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the sidebar id
	err = w.WriteUint16(uint16(p.SidebarID))
	if err != nil {
		return err
	}

	// write 1 byte for the tab id
	tabID := common.ClientTabIndices[p.Tab]
	err = w.WriteUint8(byte(tabID - 0x80))
	if err != nil {
		return err
	}

	return nil
}
