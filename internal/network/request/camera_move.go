package request

import "github.com/mbpolan/openmcs/internal/network"

const CameraModeRequestHeader byte = 0x56

// CameraModeRequest is sent when the player has moved the client's camera.
type CameraModeRequest struct {
	Vertical   int
	Horizontal int
}

// Read parses the content of the request from a stream. If the data cannot be read, an error will be returned.
func (p *CameraModeRequest) Read(r *network.ProtocolReader) error {
	// read 1 byte for the header
	_, err := r.Uint8()
	if err != nil {
		return err
	}

	// read 2 bytes for the vertical position
	vertical, err := r.Uint16()
	if err != nil {
		return err
	}

	// read 2 bytes for the horizontal position
	horizontal, err := r.Uint16()
	if err != nil {
		return err
	}

	p.Vertical = int(vertical)
	p.Horizontal = int(horizontal)
	return nil
}
