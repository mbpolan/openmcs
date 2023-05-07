package response

import "github.com/mbpolan/openmcs/internal/network"

const SetInventoryItemsResponseHeader byte = 0x22

type inventorySlot struct {
	itemID int
	amount int
}

// SetInventoryItemsResponse is sent by the server when a player's inventory should receive one or more items.
type SetInventoryItemsResponse struct {
	interfaceID int
	slots       map[int]inventorySlot
}

// NewSetInventoryItemResponse creates a new response with a player's inventory updates.
func NewSetInventoryItemResponse(interfaceID int) *SetInventoryItemsResponse {
	return &SetInventoryItemsResponse{
		interfaceID: interfaceID,
		slots:       map[int]inventorySlot{},
	}
}

// AddSlot adds an item at a slot with an amount to be sent to the player.
func (p *SetInventoryItemsResponse) AddSlot(slotID, itemID, amount int) {
	p.slots[slotID] = inventorySlot{
		itemID: itemID,
		amount: amount,
	}
}

// Write writes the contents of the message to a stream.
func (p *SetInventoryItemsResponse) Write(w *network.ProtocolWriter) error {
	// use a buffered writer since we need to compute the packet size
	bw := network.NewBufferedWriter()

	// write 2 bytes for the interface id
	err := bw.WriteUint16(uint16(p.interfaceID))
	if err != nil {
		return err
	}

	for slotID, item := range p.slots {
		// write variable bytes for the slot id
		err = bw.WriteVarByte(uint16(slotID))
		if err != nil {
			return err
		}

		// write 2 bytes for the item id. offset the id by one since the client apparently expects it that way
		err = bw.WriteUint16(uint16(item.itemID + 1))
		if err != nil {
			return err
		}

		// if the amount is greater than what we can fit into a single byte, write 4 bytes instead
		if item.amount >= 0xFF {
			err = bw.WriteUint8(0xFF)
			if err != nil {
				return err
			}

			err = bw.WriteUint32(uint32(item.amount))
			if err != nil {
				return err
			}
		} else {
			err = bw.WriteUint8(uint8(item.amount))
			if err != nil {
				return err
			}
		}
	}

	// write packet header
	err = w.WriteUint8(SetInventoryItemsResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the length of the buffered data
	buf, err := bw.Buffer()
	err = w.WriteUint16(uint16(buf.Len()))
	if err != nil {
		return err
	}

	// finally write the payload itself
	_, err = w.Write(buf.Bytes())
	if err != nil {
		return err
	}

	return nil
}
