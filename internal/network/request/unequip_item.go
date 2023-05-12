package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const UnequipItemRequestHeader byte = 0x91

// UnequipItemRequest is sent by the client when the player unequips an item.
type UnequipItemRequest struct {
	ItemID      int
	InterfaceID int
	SlotType    model.EquipmentSlotType
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *UnequipItemRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the interface id
	interfaceID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	// read 2 bytes for the equipment slot id
	slotID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	// read 2 bytes for the item id
	itemID, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	found := false
	var slotType model.EquipmentSlotType
	for _, st := range model.EquipmentSlotTypes {
		if int(slotID) == int(st) {
			slotType = st
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("unknown equipment slot type: %d", slotType)
	}

	p.ItemID = int(itemID)
	p.InterfaceID = int(interfaceID)
	p.SlotType = slotType
	return nil
}
