package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const RemoveGroundItemResponseHeader byte = 0x9C

// RemoveGroundItemResponse is sent by the server when a ground item on a tile should be removed.
type RemoveGroundItemResponse struct {
	itemID           int
	positionRelative model.Vector2D
}

// NewRemoveGroundItemResponse creates a new response to remove a ground item at a position relative to an origin.
func NewRemoveGroundItemResponse(itemID int, positionRelative model.Vector2D) *RemoveGroundItemResponse {
	return &RemoveGroundItemResponse{
		itemID:           itemID,
		positionRelative: positionRelative,
	}
}

// Write writes the contents of the message to a stream.
func (p *RemoveGroundItemResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(RemoveGroundItemResponseHeader)
	if err != nil {
		return err
	}

	// use 3 bits to represent the player's region x- and y-coordinates
	x := byte(p.positionRelative.X) & 0x07
	y := byte(p.positionRelative.Y) & 0x07

	// write 1 byte for the relative position, where the x-coordinate is in the high bits
	err = w.WriteUint8(x<<4 | y)
	if err != nil {
		return err
	}

	// write 2 bytes for the item id
	err = w.WriteUint16(uint16(p.itemID))
	if err != nil {
		return err
	}

	return nil
}
