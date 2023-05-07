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

func ReadDropInventoryItemRequest(r *network.ProtocolReader) (*DropInventoryItemRequest, error) {
	// read 2 bytes for the item id to drop
	itemID, err := r.Uint16Alt()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the interface id
	interfaceID, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the secondary action id
	secondaryActionID, err := r.Uint16Alt()
	if err != nil {
		return nil, err
	}

	return &DropInventoryItemRequest{
		ItemID:            int(itemID),
		InterfaceID:       int(interfaceID),
		SecondaryActionID: int(secondaryActionID),
	}, nil
}
