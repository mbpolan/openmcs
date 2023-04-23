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

func ReadInteractObjectRequest(r *network.ProtocolReader) (*InteractObjectRequest, error) {
	// read 2 bytes for the object's x-coordinate
	x, err := r.Uint16LEAlt()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the action id
	action, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the object's y-coordinate
	y, err := r.Uint16Alt()
	if err != nil {
		return nil, err
	}

	return &InteractObjectRequest{
		GlobalPos: model.Vector2D{
			X: int(x),
			Y: int(y),
		},
		Action: int(action),
	}, nil
}
