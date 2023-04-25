package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const ShowGroundItemResponseHeader byte = 0x2C

// ShowGroundItemResponse is sent by the server when a ground item should be displayed at a location.
type ShowGroundItemResponse struct {
	itemID           int
	stackSize        int
	positionRelative model.Vector2D
}

// NewShowGroundItemResponse creates a new response to show a ground item at a position relative to an origin.
func NewShowGroundItemResponse(itemID, stackSize int, positionRelative model.Vector2D) *ShowGroundItemResponse {
	return &ShowGroundItemResponse{
		itemID:           itemID,
		stackSize:        stackSize,
		positionRelative: positionRelative,
	}
}

// Write writes the contents of the message to a stream.
func (p *ShowGroundItemResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(ShowGroundItemResponseHeader)
	if err != nil {
		return err
	}

	// write 2 bytes for the item id
	err = w.WriteUint16Alt2(uint16(p.itemID))
	if err != nil {
		return err
	}

	// write 2 bytes for the stack size
	err = w.WriteUint16(uint16(p.stackSize))
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

	return nil
}
