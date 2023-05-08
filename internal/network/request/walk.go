package request

import (
	"github.com/mbpolan/openmcs/internal/model"
	"github.com/mbpolan/openmcs/internal/network"
)

const WalkRequestHeader byte = 0xA4
const WalkOnCommandRequestHeader byte = 0x62
const WalkMinimap byte = 0xF8

type WalkRequest struct {
	ControlPressed bool
	PathLength     int
	Start          model.Vector2D
	Waypoints      []model.Vector2D
}

func (p *WalkRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	header, err := r.Uint8()
	if err != nil {
		return err
	}

	// read a byte for the length of the path
	size, err := r.Uint8()
	if err != nil {
		return err
	}

	// adjust the path size based on the walk type
	var pathLength int
	if header != WalkMinimap {
		pathLength = int((size-3)/2) - 1
	} else {
		pathLength = int((size-17)/2) - 1
	}

	// read 2 bytes for the starting x coordinate
	startX, err := r.Uint16LEAlt()
	if err != nil {
		return err
	}

	// read each waypoint in the path
	waypoints := make([]model.Vector2D, pathLength)
	for i := 0; i < pathLength; i++ {
		// read one byte each for the waypoint x and y coordinates
		wx, err := r.Int8()
		if err != nil {
			return err
		}

		wy, err := r.Int8()
		if err != nil {
			return err
		}

		waypoints[i] = model.Vector2D{
			X: int(wx),
			Y: int(wy),
		}
	}

	// read 2 bytes for the starting y coordinate
	startY, err := r.Uint16LE()
	if err != nil {
		return err
	}

	// read one byte for the flag that indicates if the control key was pressed
	ctrlKeyPressed, err := r.Uint8()
	if err != nil {
		return err
	}

	// skip 14 bytes if the walk type was based on the minimap
	if header == WalkMinimap {
		err = r.Skip(14)
		if err != nil {
			return err
		}
	}

	p.ControlPressed = ctrlKeyPressed == 0xFF
	p.PathLength = pathLength
	p.Waypoints = waypoints
	p.Start = model.Vector2D{
		X: int(startX),
		Y: int(startY),
	}
	return nil
}
