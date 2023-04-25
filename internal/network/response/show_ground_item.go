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

	// write 1 byte for the relative position
	offset := byte(p.positionRelative.X)<<7 | byte(p.positionRelative.Y)
	err = w.WriteUint8(offset)
	if err != nil {
		return err
	}

	return nil
}
