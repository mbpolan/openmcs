package request

import (
	"fmt"
	"github.com/mbpolan/openmcs/internal/network"
)

const CastSpellOnItemRequestHeader byte = 0xED

// CastSpellOnItemRequest is sent by the client when the player casts a spell on an inventory item.
type CastSpellOnItemRequest struct {
	// SlotID is the inventory slot ID of the target item.
	SlotID int
	// ItemID is the ID of the target item.
	ItemID int
	// InventoryInterfaceID is the ID of the inventory interface.
	InventoryInterfaceID int
	// SpellInterfaceID is the ID of the spell interface that was cast.
	SpellInterfaceID int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *CastSpellOnItemRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	b, err := r.Uint8()
	if err != nil {
		return err
	}

	if b != CastSpellOnItemRequestHeader {
		return fmt.Errorf("unexpected packet header")
	}

	// read 2 bytes for the inventory slot id
	slotID, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the target item id
	itemID, err := r.Uint16Alt2()
	if err != nil {
		return err
	}

	// read 2 bytes for the inventory interface id
	invInterfaceID, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the source (spell) interface id
	spellInterfaceID, err := r.Uint16Alt2()
	if err != nil {
		return err
	}

	p.SlotID = int(slotID)
	p.ItemID = int(itemID)
	p.InventoryInterfaceID = int(invInterfaceID)
	p.SpellInterfaceID = int(spellInterfaceID)
	return nil
}
