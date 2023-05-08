package request

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const InteractObjectRequestHeader byte = 0x84

// InteractObjectRequest is sent by the client when a player interacts with an object.
type InteractObjectRequest struct {
	GlobalPos model.Vector2D
	Action    int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *InteractObjectRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the object's x-coordinate
	x, err := r.Uint16LEAlt()
	if err != nil {
		return err
	}

	// read 2 bytes for the action id
	action, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the object's y-coordinate
	y, err := r.Uint16Alt()
	if err != nil {
		return err
	}

	p.Action = int(action)
	p.GlobalPos = model.Vector2D{
		X: int(x),
		Y: int(y),
	}
	return nil
}
