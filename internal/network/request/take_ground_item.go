package request

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const TakeGroundItemRequestHeader byte = 0xEC

// TakeGroundItemRequest is sent by the client when a player attempts to pick up a ground item.
type TakeGroundItemRequest struct {
	GlobalPos model.Vector2D
	ItemID    int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *TakeGroundItemRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the y-coordinate
	y, err := r.Uint16LE()
	if err != nil {
		return err
	}

	// read 2 bytes for the target item
	itemID, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the x-coordinate
	x, err := r.Uint16LE()
	if err != nil {
		return err
	}

	p.GlobalPos = model.Vector2D{
		X: int(x),
		Y: int(y),
	}
	p.ItemID = int(itemID)
	return nil
}
