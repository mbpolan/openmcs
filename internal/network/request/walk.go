package request

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const WalkRequestHeader byte = 0xA4
const WalkOnCommandRequestHeader byte = 0x62

type WalkRequest struct {
	ControlPressed bool
	PathLength     int
	Start          model.Vector2D
	Waypoints      []model.Vector2D
}

func ReadWalkRequest(r *network.ProtocolReader) (*WalkRequest, error) {
	// read a byte for the length of the path
	size, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	// fix the size since the client mangles it for some reason
	pathLength := int((size-3)/2) - 1

	// read 2 bytes for the starting x coordinate
	startX, err := r.Uint16LEAlt()
	if err != nil {
		return nil, err
	}

	// read each waypoint in the path
	waypoints := make([]model.Vector2D, pathLength)
	for i := 0; i < pathLength; i++ {
		// read one byte each for the waypoint x and y coordinates
		wx, err := r.Int8()
		if err != nil {
			return nil, err
		}

		wy, err := r.Int8()
		if err != nil {
			return nil, err
		}

		waypoints[i] = model.Vector2D{
			X: int(wx),
			Y: int(wy),
		}
	}

	// read 2 bytes for the starting y coordinate
	startY, err := r.Uint16LE()
	if err != nil {
		return nil, err
	}

	// read one byte for the flag that indicates if the control key was pressed
	ctrlKeyPressed, err := r.Uint8()
	if err != nil {
		return nil, err
	}

	return &WalkRequest{
		ControlPressed: ctrlKeyPressed == 0xFF,
		PathLength:     pathLength,
		Start: model.Vector2D{
			X: int(startX),
			Y: int(startY),
		},
		Waypoints: waypoints,
	}, nil
}
