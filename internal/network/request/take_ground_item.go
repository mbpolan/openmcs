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

func ReadTakeGroundItemRequest(r *network.ProtocolReader) (*TakeGroundItemRequest, error) {
	// read 2 bytes for the y-coordinate
	y, err := r.Uint16LE()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the target item
	itemID, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the x-coordinate
	x, err := r.Uint16LE()
	if err != nil {
		return nil, err
	}

	return &TakeGroundItemRequest{
		GlobalPos: model.Vector2D{
			X: int(x),
			Y: int(y),
		},
		ItemID: int(itemID),
	}, nil
}
