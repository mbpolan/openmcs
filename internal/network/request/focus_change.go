package request

import "github.com/mbpolan/openmcs/internal/network"

const FocusRequestHeader byte = 0x03

// FocusChangeRequest is sent by the client when the client window loses or acquires focus.
type FocusChangeRequest struct {
	Focused bool
}

// ReadFocusRequest parses the packet from the connection stream.
func ReadFocusRequest(r *network.ProtocolReader) (*FocusChangeRequest, error) {
	// read a single byte, indicating if the client window is current focused or not
	b, err := r.Byte()
	if err != nil {
		return nil, err
	}

	return &FocusChangeRequest{
		Focused: b == 0x01,
	}, nil
}
