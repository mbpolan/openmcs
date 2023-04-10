package requests

import "github.com/mbpolan/openmcs/internal/network"

const FocusRequestHeader byte = 0x03

// FocusRequest is sent by the client when the client window loses or acquires focus.
type FocusRequest struct {
	Focused bool
}

// ReadFocusRequest parses the packet from the connection stream.
func ReadFocusRequest(r *network.ProtocolReader) (*FocusRequest, error) {
	// read a single byte, indicating if the client window is current focused or not
	b, err := r.Byte()
	if err != nil {
		return nil, err
	}

	return &FocusRequest{
		Focused: b == 0x01,
	}, nil
}
