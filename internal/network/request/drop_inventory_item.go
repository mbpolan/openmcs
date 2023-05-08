package request

import (
	"github.com/mbpolan/openmcs/internal/network"
)

const DropInventoryItemRequestHeader byte = 0x57

// DropInventoryItemRequest is sent by the client when the player drops an inventory item.
type DropInventoryItemRequest struct {
	ItemID            int
	InterfaceID       int
	SecondaryActionID int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *DropInventoryItemRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the item id to drop
	itemID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	// read 2 bytes for the interface id
	interfaceID, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the secondary action id
	secondaryActionID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	p.ItemID = int(itemID)
	p.InterfaceID = int(interfaceID)
	p.SecondaryActionID = int(secondaryActionID)
	return nil
}
