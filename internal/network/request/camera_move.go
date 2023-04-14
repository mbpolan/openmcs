package request

import "github.com/mbpolan/openmcs/internal/network"

const CameraModeRequestHeader byte = 0x56

// CameraModeRequest is sent when the player has moved the client's camera.
type CameraModeRequest struct {
	Vertical   int
	Horizontal int
}

func ReadCameraModeRequest(r *network.ProtocolReader) (*CameraModeRequest, error) {
	// read 2 bytes for the vertical position
	vertical, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	// read 2 bytes for the horizontal position
	horizontal, err := r.Uint16()
	if err != nil {
		return nil, err
	}

	return &CameraModeRequest{
		Vertical:   int(vertical),
		Horizontal: int(horizontal),
	}, nil
}
