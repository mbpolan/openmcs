package request

import "github.com/mbpolan/openmcs/internal/network"

const UseItemRequestHeader byte = 0x7A

// UseItemRequest is sent when the player uses an item.
type UseItemRequest struct {
	ItemID      int
	InterfaceID int
	ActionID    int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *UseItemRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the interface id
	interfaceID, err := r.Uint16LEAlt()
	if err != nil {
		return err
	}

	// read 2 bytes for the action id
	secondaryActionID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	// read 2 bytes for the item id
	itemID, err := r.Uint16LE()
	if err != nil {
		return err
	}

	p.ItemID = int(itemID)
	p.InterfaceID = int(interfaceID)
	p.ActionID = int(secondaryActionID)
	return nil
}
