package response

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const UpdateGroundItemResponseHeader byte = 0x54

// UpdateGroundItemResponse is sent by the server when a ground item's stack amount has changed.
type UpdateGroundItemResponse struct {
	itemID           int
	oldStackSize     int
	newStackSize     int
	positionRelative model.Vector2D
}

// NewUpdateGroundItemResponse creates a new ground item stack update response.
func NewUpdateGroundItemResponse(itemID, oldStackSize, newStackSize int, positionRelative model.Vector2D) *UpdateGroundItemResponse {
	return &UpdateGroundItemResponse{
		itemID:           itemID,
		oldStackSize:     oldStackSize,
		newStackSize:     newStackSize,
		positionRelative: positionRelative,
	}
}

// Write writes the contents of the message to a stream.
func (p *UpdateGroundItemResponse) Write(w *network.ProtocolWriter) error {
	// write packet header
	err := w.WriteUint8(UpdateGroundItemResponseHeader)
	if err != nil {
		return err
	}

	// use 3 bits to represent the item's region x- and y-coordinates
	x := byte(p.positionRelative.X) & 0x07
	y := byte(p.positionRelative.Y) & 0x07

	// write 1 byte for the relative position, where the x-coordinate is in the high bits
	err = w.WriteUint8(x<<4 | y)
	if err != nil {
		return err
	}

	// write 2 bytes for the target item id
	err = w.WriteUint16LE(uint16(p.itemID))
	if err != nil {
		return err
	}

	// write 2 bytes for the old stack size
	err = w.WriteUint16LE(uint16(p.oldStackSize))
	if err != nil {
		return err
	}

	// write 2 bytes for the new stack size
	err = w.WriteUint16LE(uint16(p.newStackSize))
	if err != nil {
		return err
	}

	return nil
}
