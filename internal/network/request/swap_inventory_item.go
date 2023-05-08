package request

import "github.com/mbpolan/openmcs/internal/network"

const SwapInventoryItemRequestHeader byte = 0xD6

// SwapInventoryItemRequest is sent by the client when the player rearranges an item in their inventory.
type SwapInventoryItemRequest struct {
	InterfaceID int
	InsertMode  bool
	FromSlot    int
	ToSlot      int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *SwapInventoryItemRequest) Read(r *network.ProtocolReader) error {
	// read 2 bytes for the inventory interface id
	interfaceID, err := r.Uint16LEAlt()
	if err != nil {
		return err
	}

	// read 1 byte for the mode (swap or insert)
	mode, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the original slot
	from, err := r.Uint16LEAlt()
	if err != nil {
		return err
	}

	// read 2 bytes for the target slot
	to, err := r.Uint16LE()
	if err != nil {
		return err
	}

	p.InterfaceID = int(interfaceID)
	p.InsertMode = mode != 0
	p.FromSlot = int(from)
	p.ToSlot = int(to)
	return nil
}
